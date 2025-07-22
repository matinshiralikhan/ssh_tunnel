package protocols

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"ssh-tunnel/internal/config"
)

// TunnelStatus represents the status of a tunnel
type TunnelStatus struct {
	ServerName string        `json:"server_name"`
	Status     string        `json:"status"` // "connected", "connecting", "disconnected", "error"
	StartTime  time.Time     `json:"start_time"`
	LastError  string        `json:"last_error,omitempty"`
	BytesSent  uint64        `json:"bytes_sent"`
	BytesRecv  uint64        `json:"bytes_recv"`
	Latency    time.Duration `json:"latency"`
}

// TunnelManager manages multiple tunnel connections
type TunnelManager struct {
	config  *config.Config
	tunnels map[string]Tunnel
	status  map[string]*TunnelStatus
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
}

// Tunnel interface for different protocol implementations
type Tunnel interface {
	Start(ctx context.Context) error
	Stop() error
	GetStatus() *TunnelStatus
	GetName() string
	Test() (time.Duration, error)
}

// NewTunnelManager creates a new tunnel manager
func NewTunnelManager(cfg *config.Config) *TunnelManager {
	return &TunnelManager{
		config:  cfg,
		tunnels: make(map[string]Tunnel),
		status:  make(map[string]*TunnelStatus),
	}
}

// Start starts the tunnel manager
func (tm *TunnelManager) Start(ctx context.Context) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.ctx, tm.cancel = context.WithCancel(ctx)

	// Initialize tunnels for all enabled servers
	for _, server := range tm.config.Servers {
		if !server.Enabled {
			continue
		}

		tunnel, err := tm.createTunnel(server)
		if err != nil {
			log.Printf("Failed to create tunnel for %s: %v", server.Name, err)
			continue
		}

		tm.tunnels[server.Name] = tunnel
		tm.status[server.Name] = &TunnelStatus{
			ServerName: server.Name,
			Status:     "disconnected",
		}
	}

	// Start auto-selection if enabled
	if tm.config.AutoSelect {
		return tm.startAutoSelected()
	}

	return nil
}

// Stop stops all tunnels
func (tm *TunnelManager) Stop() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.cancel != nil {
		tm.cancel()
	}

	var errors []error
	for name, tunnel := range tm.tunnels {
		if err := tunnel.Stop(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop tunnel %s: %v", name, err))
		}

		if status, ok := tm.status[name]; ok {
			status.Status = "disconnected"
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors stopping tunnels: %v", errors)
	}

	return nil
}

// StartTunnel starts a specific tunnel
func (tm *TunnelManager) StartTunnel(serverName string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tunnel, exists := tm.tunnels[serverName]
	if !exists {
		return fmt.Errorf("tunnel %s not found", serverName)
	}

	status := tm.status[serverName]
	status.Status = "connecting"
	status.StartTime = time.Now()

	go func() {
		if err := tunnel.Start(tm.ctx); err != nil {
			tm.mu.Lock()
			status.Status = "error"
			status.LastError = err.Error()
			tm.mu.Unlock()
			log.Printf("Tunnel %s failed: %v", serverName, err)
		} else {
			tm.mu.Lock()
			status.Status = "connected"
			tm.mu.Unlock()
		}
	}()

	return nil
}

// StopAllTunnels stops all running tunnels
func (tm *TunnelManager) StopAllTunnels() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	var errors []error
	for name, tunnel := range tm.tunnels {
		if err := tunnel.Stop(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop tunnel %s: %v", name, err))
		}

		if status, ok := tm.status[name]; ok {
			status.Status = "disconnected"
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors stopping tunnels: %v", errors)
	}

	return nil
}

// RestartTunnels restarts all tunnels
func (tm *TunnelManager) RestartTunnels() error {
	if err := tm.StopAllTunnels(); err != nil {
		return err
	}

	time.Sleep(time.Second) // Brief pause between stop and start

	return tm.Start(tm.ctx)
}

// GetStatus returns the status of all tunnels
func (tm *TunnelManager) GetStatus() map[string]*TunnelStatus {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	result := make(map[string]*TunnelStatus)
	for name, status := range tm.status {
		// Create a copy to avoid race conditions
		statusCopy := *status
		result[name] = &statusCopy
	}

	return result
}

// GetTunnels returns all tunnel configurations
func (tm *TunnelManager) GetTunnels() []config.Server {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return tm.config.Servers
}

// TestServer tests connectivity to a specific server
func (tm *TunnelManager) TestServer(serverName string) interface{} {
	tm.mu.RLock()
	tunnel, exists := tm.tunnels[serverName]
	tm.mu.RUnlock()

	if !exists {
		return map[string]interface{}{
			"server": serverName,
			"error":  "Server not found",
		}
	}

	latency, err := tunnel.Test()
	if err != nil {
		return map[string]interface{}{
			"server": serverName,
			"error":  err.Error(),
		}
	}

	return map[string]interface{}{
		"server":  serverName,
		"latency": latency.String(),
		"status":  "ok",
	}
}

// UpdateConfig updates the configuration
func (tm *TunnelManager) UpdateConfig(cfg *config.Config) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.config = cfg

	// TODO: Implement configuration update logic
	// This would involve stopping current tunnels and recreating them
	// with the new configuration

	return nil
}

// startAutoSelected starts the best available server based on selection method
func (tm *TunnelManager) startAutoSelected() error {
	switch tm.config.SelectionMethod {
	case "latency":
		return tm.startBestLatency()
	case "random":
		return tm.startRandom()
	case "load":
		return tm.startLeastLoad()
	default:
		return tm.startBestLatency()
	}
}

// startBestLatency starts the server with the best latency
func (tm *TunnelManager) startBestLatency() error {
	var bestServer string
	var bestLatency time.Duration = time.Hour // Initialize with a high value

	for name, tunnel := range tm.tunnels {
		latency, err := tunnel.Test()
		if err != nil {
			log.Printf("Failed to test server %s: %v", name, err)
			continue
		}

		if latency < bestLatency {
			bestLatency = latency
			bestServer = name
		}
	}

	if bestServer == "" {
		return fmt.Errorf("no available servers found")
	}

	log.Printf("Auto-selected server %s with latency %v", bestServer, bestLatency)
	return tm.StartTunnel(bestServer)
}

// startRandom starts a random available server
func (tm *TunnelManager) startRandom() error {
	// Simple implementation - just pick the first available
	for name := range tm.tunnels {
		return tm.StartTunnel(name)
	}
	return fmt.Errorf("no available servers found")
}

// startLeastLoad starts the server with the least load (placeholder)
func (tm *TunnelManager) startLeastLoad() error {
	// For now, fallback to latency-based selection
	return tm.startBestLatency()
}

// createTunnel creates a tunnel instance based on the server configuration
func (tm *TunnelManager) createTunnel(server config.Server) (Tunnel, error) {
	switch server.Transport {
	case config.TransportSSH:
		return NewSSHTunnel(server), nil
	case config.TransportHysteria:
		return NewHysteriaTunnel(server), nil
	case config.TransportV2Ray, config.TransportVMess, config.TransportVLESS:
		return NewV2RayTunnel(server), nil
	case config.TransportWireGuard:
		return NewWireGuardTunnel(server), nil
	case config.TransportTrojan:
		return NewTrojanTunnel(server), nil
	default:
		return nil, fmt.Errorf("unsupported transport type: %s", server.Transport)
	}
}
