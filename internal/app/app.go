package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"ssh-tunnel/internal/config"
	"ssh-tunnel/internal/monitoring"
	"ssh-tunnel/internal/protocols"
)

// Application represents the main application
type Application struct {
	config    *config.Config
	tunnelMgr *protocols.TunnelManager
	monitor   *monitoring.Monitor
	server    *echo.Echo
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

// New creates a new application instance
func New(cfg *config.Config) *Application {
	ctx, cancel := context.WithCancel(context.Background())

	app := &Application{
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize tunnel manager
	app.tunnelMgr = protocols.NewTunnelManager(cfg)

	// Initialize monitoring
	if cfg.Monitoring.Enabled {
		app.monitor = monitoring.NewMonitor(cfg.Monitoring)
	}

	// Initialize Echo server
	if cfg.API.Enabled {
		app.setupServer()
	}

	return app
}

// StartClient starts the application in client mode
func (a *Application) StartClient() error {
	log.Println("Starting SSH Tunnel Manager in client mode...")

	// Start monitoring if enabled
	if a.monitor != nil {
		go a.monitor.Start(a.ctx)
	}

	// Start tunnel manager
	return a.tunnelMgr.Start(a.ctx)
}

// StartServer starts the application in server mode with REST API
func (a *Application) StartServer(port string) error {
	log.Printf("Starting SSH Tunnel Manager server on port %s...", port)

	// Start monitoring if enabled
	if a.monitor != nil {
		go a.monitor.Start(a.ctx)
	}

	// Start tunnel manager in background
	go func() {
		if err := a.tunnelMgr.Start(a.ctx); err != nil {
			log.Printf("Tunnel manager error: %v", err)
		}
	}()

	// Start HTTP server
	if a.server != nil {
		return a.server.Start(":" + port)
	}

	return fmt.Errorf("HTTP server not initialized")
}

// Shutdown gracefully shuts down the application
func (a *Application) Shutdown(ctx context.Context) error {
	log.Println("Shutting down application...")

	var errors []error

	// Stop tunnel manager
	if err := a.tunnelMgr.Stop(); err != nil {
		errors = append(errors, fmt.Errorf("tunnel manager shutdown error: %v", err))
	}

	// Stop monitoring
	if a.monitor != nil {
		if err := a.monitor.Stop(); err != nil {
			errors = append(errors, fmt.Errorf("monitor shutdown error: %v", err))
		}
	}

	// Stop HTTP server
	if a.server != nil {
		if err := a.server.Shutdown(ctx); err != nil {
			errors = append(errors, fmt.Errorf("HTTP server shutdown error: %v", err))
		}
	}

	// Cancel context
	a.cancel()

	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}

	return nil
}

// setupServer sets up the Echo HTTP server with routes and middleware
func (a *Application) setupServer() {
	a.server = echo.New()
	a.server.HideBanner = true

	// Middleware
	a.server.Use(middleware.Logger())
	a.server.Use(middleware.Recover())

	if a.config.API.EnableCORS {
		a.server.Use(middleware.CORS())
	}

	// Rate limiting if configured
	if a.config.API.RateLimit > 0 {
		a.server.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
			rate.Limit(a.config.API.RateLimit),
		)))
	}

	// Authentication middleware if enabled
	if a.config.Security.EnableAuth {
		a.server.Use(a.authMiddleware)
	}

	// API routes
	api := a.server.Group("/api/v1")

	// System routes
	api.GET("/health", a.handleHealth)
	api.GET("/status", a.handleStatus)
	api.GET("/config", a.handleGetConfig)
	api.PUT("/config", a.handleUpdateConfig)

	// Server management routes
	api.GET("/servers", a.handleGetServers)
	api.POST("/servers", a.handleAddServer)
	api.PUT("/servers/:id", a.handleUpdateServer)
	api.DELETE("/servers/:id", a.handleDeleteServer)
	api.POST("/servers/:id/test", a.handleTestServer)

	// Tunnel management routes
	api.GET("/tunnels", a.handleGetTunnels)
	api.POST("/tunnels/start", a.handleStartTunnel)
	api.POST("/tunnels/stop", a.handleStopTunnel)
	api.POST("/tunnels/restart", a.handleRestartTunnel)

	// Monitoring routes
	if a.config.Monitoring.Enabled {
		api.GET("/metrics", a.handleMetrics)
		api.GET("/logs", a.handleLogs)
	}
}

// authMiddleware provides authentication for API endpoints
func (a *Application) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Authorization token required",
			})
		}

		// Remove "Bearer " prefix if present
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// Check if token is valid
		valid := false
		for _, validToken := range a.config.Security.AuthTokens {
			if token == validToken {
				valid = true
				break
			}
		}

		if !valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid authorization token",
			})
		}

		return next(c)
	}
}

// API Handlers

func (a *Application) handleHealth(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   a.config.Version,
	})
}

func (a *Application) handleStatus(c echo.Context) error {
	status := a.tunnelMgr.GetStatus()
	return c.JSON(http.StatusOK, status)
}

func (a *Application) handleGetConfig(c echo.Context) error {
	// Return config without sensitive information
	safeConfig := *a.config
	safeConfig.Security.AuthTokens = nil
	safeConfig.Security.MasterPassword = ""

	for i := range safeConfig.Servers {
		safeConfig.Servers[i].Password = ""
		safeConfig.Servers[i].KeyPath = ""
		if safeConfig.Servers[i].Hysteria != nil {
			safeConfig.Servers[i].Hysteria.AuthString = ""
		}
	}

	return c.JSON(http.StatusOK, safeConfig)
}

func (a *Application) handleUpdateConfig(c echo.Context) error {
	var newConfig config.Config
	if err := c.Bind(&newConfig); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid configuration format",
		})
	}

	// Validate new configuration
	if err := a.validateConfig(&newConfig); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("Configuration validation failed: %v", err),
		})
	}

	// Update application configuration
	a.mu.Lock()
	a.config = &newConfig
	a.mu.Unlock()

	// Restart tunnel manager with new config
	if err := a.tunnelMgr.UpdateConfig(&newConfig); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to update tunnel configuration: %v", err),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Configuration updated successfully",
	})
}

func (a *Application) handleGetServers(c echo.Context) error {
	return c.JSON(http.StatusOK, a.config.Servers)
}

func (a *Application) handleAddServer(c echo.Context) error {
	var server config.Server
	if err := c.Bind(&server); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid server configuration",
		})
	}

	a.mu.Lock()
	a.config.Servers = append(a.config.Servers, server)
	a.mu.Unlock()

	return c.JSON(http.StatusCreated, server)
}

func (a *Application) handleUpdateServer(c echo.Context) error {
	id := c.Param("id")
	// Implementation for updating specific server
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Server updated",
		"id":      id,
	})
}

func (a *Application) handleDeleteServer(c echo.Context) error {
	id := c.Param("id")
	// Implementation for deleting specific server
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Server deleted",
		"id":      id,
	})
}

func (a *Application) handleTestServer(c echo.Context) error {
	id := c.Param("id")
	result := a.tunnelMgr.TestServer(id)
	return c.JSON(http.StatusOK, result)
}

func (a *Application) handleGetTunnels(c echo.Context) error {
	tunnels := a.tunnelMgr.GetTunnels()
	return c.JSON(http.StatusOK, tunnels)
}

func (a *Application) handleStartTunnel(c echo.Context) error {
	serverID := c.QueryParam("server")
	if err := a.tunnelMgr.StartTunnel(serverID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Tunnel started",
	})
}

func (a *Application) handleStopTunnel(c echo.Context) error {
	if err := a.tunnelMgr.StopAllTunnels(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Tunnels stopped",
	})
}

func (a *Application) handleRestartTunnel(c echo.Context) error {
	if err := a.tunnelMgr.RestartTunnels(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Tunnels restarted",
	})
}

func (a *Application) handleMetrics(c echo.Context) error {
	if a.monitor == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Monitoring not enabled",
		})
	}

	metrics := a.monitor.GetMetrics()
	return c.JSON(http.StatusOK, metrics)
}

func (a *Application) handleLogs(c echo.Context) error {
	if a.monitor == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Monitoring not enabled",
		})
	}

	logs := a.monitor.GetLogs()
	return c.JSON(http.StatusOK, logs)
}

// validateConfig validates the configuration
func (a *Application) validateConfig(cfg *config.Config) error {
	// Basic validation logic here
	if len(cfg.Servers) == 0 {
		return fmt.Errorf("no servers configured")
	}
	return nil
}
