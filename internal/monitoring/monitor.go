package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"ssh-tunnel/internal/config"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// Metrics holds system and application metrics
type Metrics struct {
	System      SystemMetrics      `json:"system"`
	Application ApplicationMetrics `json:"application"`
	Tunnels     []TunnelMetrics    `json:"tunnels"`
	Timestamp   time.Time          `json:"timestamp"`
}

// SystemMetrics holds system-level metrics
type SystemMetrics struct {
	CPUUsage   float64   `json:"cpu_usage"`
	MemUsage   float64   `json:"memory_usage"`
	MemTotal   uint64    `json:"memory_total"`
	MemUsed    uint64    `json:"memory_used"`
	NetworkIO  NetworkIO `json:"network_io"`
	Goroutines int       `json:"goroutines"`
}

// ApplicationMetrics holds application-specific metrics
type ApplicationMetrics struct {
	Uptime            time.Duration `json:"uptime"`
	ActiveTunnels     int           `json:"active_tunnels"`
	TotalConnections  uint64        `json:"total_connections"`
	FailedConnections uint64        `json:"failed_connections"`
	BytesTransferred  uint64        `json:"bytes_transferred"`
}

// TunnelMetrics holds per-tunnel metrics
type TunnelMetrics struct {
	Name       string        `json:"name"`
	Status     string        `json:"status"`
	Latency    time.Duration `json:"latency"`
	BytesSent  uint64        `json:"bytes_sent"`
	BytesRecv  uint64        `json:"bytes_received"`
	Uptime     time.Duration `json:"uptime"`
	Reconnects int           `json:"reconnects"`
}

// NetworkIO holds network I/O statistics
type NetworkIO struct {
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Component string                 `json:"component"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// Monitor handles monitoring and metrics collection
type Monitor struct {
	config    config.MonitoringConfig
	metrics   *Metrics
	logs      []LogEntry
	startTime time.Time
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewMonitor creates a new monitoring instance
func NewMonitor(cfg config.MonitoringConfig) *Monitor {
	return &Monitor{
		config:    cfg,
		logs:      make([]LogEntry, 0, 1000), // Keep last 1000 log entries
		startTime: time.Now(),
	}
}

// Start begins monitoring
func (m *Monitor) Start(ctx context.Context) error {
	m.mu.Lock()
	m.ctx, m.cancel = context.WithCancel(ctx)
	m.mu.Unlock()

	log.Println("Starting monitoring system...")

	// Start metrics collection
	go m.collectMetrics()

	// Start log rotation if configured
	if m.config.LogFile != "" {
		go m.rotateLogFiles()
	}

	return nil
}

// Stop stops monitoring
func (m *Monitor) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cancel != nil {
		m.cancel()
	}

	log.Println("Monitoring system stopped")
	return nil
}

// GetMetrics returns current metrics
func (m *Monitor) GetMetrics() *Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.metrics == nil {
		return &Metrics{
			Timestamp: time.Now(),
		}
	}

	// Return a copy to avoid race conditions
	metricsCopy := *m.metrics
	return &metricsCopy
}

// GetLogs returns recent log entries
func (m *Monitor) GetLogs() []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy of logs
	logsCopy := make([]LogEntry, len(m.logs))
	copy(logsCopy, m.logs)
	return logsCopy
}

// LogEvent adds a log entry
func (m *Monitor) LogEvent(level, component, message string, details map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Component: component,
		Message:   message,
		Details:   details,
	}

	// Add to logs
	m.logs = append(m.logs, entry)

	// Keep only the latest entries
	if len(m.logs) > 1000 {
		m.logs = m.logs[len(m.logs)-1000:]
	}

	// Log to stdout as well
	detailsJSON := ""
	if details != nil {
		if data, err := json.Marshal(details); err == nil {
			detailsJSON = string(data)
		}
	}

	log.Printf("[%s] %s: %s %s", level, component, message, detailsJSON)
}

// collectMetrics periodically collects system and application metrics
func (m *Monitor) collectMetrics() {
	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.updateMetrics()
		}
	}
}

// updateMetrics updates the current metrics
func (m *Monitor) updateMetrics() {
	metrics := &Metrics{
		Timestamp: time.Now(),
	}

	// Collect system metrics
	metrics.System = m.collectSystemMetrics()

	// Collect application metrics
	metrics.Application = m.collectApplicationMetrics()

	// Update stored metrics
	m.mu.Lock()
	m.metrics = metrics
	m.mu.Unlock()
}

// collectSystemMetrics collects system-level metrics
func (m *Monitor) collectSystemMetrics() SystemMetrics {
	sysMetrics := SystemMetrics{
		Goroutines: runtime.NumGoroutine(),
	}

	// CPU usage
	if cpuPercent, err := cpu.Percent(time.Second, false); err == nil && len(cpuPercent) > 0 {
		sysMetrics.CPUUsage = cpuPercent[0]
	}

	// Memory usage
	if vmStat, err := mem.VirtualMemory(); err == nil {
		sysMetrics.MemUsage = vmStat.UsedPercent
		sysMetrics.MemTotal = vmStat.Total
		sysMetrics.MemUsed = vmStat.Used
	}

	// Network I/O
	if netStat, err := net.IOCounters(false); err == nil && len(netStat) > 0 {
		sysMetrics.NetworkIO = NetworkIO{
			BytesSent:   netStat[0].BytesSent,
			BytesRecv:   netStat[0].BytesRecv,
			PacketsSent: netStat[0].PacketsSent,
			PacketsRecv: netStat[0].PacketsRecv,
		}
	}

	return sysMetrics
}

// collectApplicationMetrics collects application-specific metrics
func (m *Monitor) collectApplicationMetrics() ApplicationMetrics {
	return ApplicationMetrics{
		Uptime: time.Since(m.startTime),
		// Other metrics would be updated by the tunnel manager
		ActiveTunnels:     0, // Placeholder
		TotalConnections:  0, // Placeholder
		FailedConnections: 0, // Placeholder
		BytesTransferred:  0, // Placeholder
	}
}

// UpdateTunnelMetrics updates metrics for a specific tunnel
func (m *Monitor) UpdateTunnelMetrics(name, status string, latency time.Duration, bytesSent, bytesRecv uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.metrics == nil {
		return
	}

	// Find existing tunnel metrics or create new
	var tunnelMetrics *TunnelMetrics
	for i := range m.metrics.Tunnels {
		if m.metrics.Tunnels[i].Name == name {
			tunnelMetrics = &m.metrics.Tunnels[i]
			break
		}
	}

	if tunnelMetrics == nil {
		m.metrics.Tunnels = append(m.metrics.Tunnels, TunnelMetrics{
			Name: name,
		})
		tunnelMetrics = &m.metrics.Tunnels[len(m.metrics.Tunnels)-1]
	}

	// Update metrics
	tunnelMetrics.Status = status
	tunnelMetrics.Latency = latency
	tunnelMetrics.BytesSent = bytesSent
	tunnelMetrics.BytesRecv = bytesRecv
}

// rotateLogFiles handles log file rotation
func (m *Monitor) rotateLogFiles() {
	// Simple log rotation implementation
	ticker := time.NewTicker(24 * time.Hour) // Rotate daily
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			// TODO: Implement actual log file rotation
			m.LogEvent("info", "monitor", "Log rotation triggered", nil)
		}
	}
}

// GetHealthStatus returns the health status of the system
func (m *Monitor) GetHealthStatus() map[string]interface{} {
	metrics := m.GetMetrics()

	status := map[string]interface{}{
		"status": "healthy",
		"checks": map[string]interface{}{
			"cpu": map[string]interface{}{
				"status": "ok",
				"usage":  fmt.Sprintf("%.2f%%", metrics.System.CPUUsage),
			},
			"memory": map[string]interface{}{
				"status": "ok",
				"usage":  fmt.Sprintf("%.2f%%", metrics.System.MemUsage),
			},
			"tunnels": map[string]interface{}{
				"status": "ok",
				"active": metrics.Application.ActiveTunnels,
			},
		},
		"uptime": metrics.Application.Uptime.String(),
	}

	// Check for unhealthy conditions
	if metrics.System.CPUUsage > 90 {
		status["checks"].(map[string]interface{})["cpu"].(map[string]interface{})["status"] = "warning"
		status["status"] = "degraded"
	}

	if metrics.System.MemUsage > 90 {
		status["checks"].(map[string]interface{})["memory"].(map[string]interface{})["status"] = "warning"
		status["status"] = "degraded"
	}

	return status
}
