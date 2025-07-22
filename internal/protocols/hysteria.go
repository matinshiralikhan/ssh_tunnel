package protocols

import (
	"context"
	"fmt"
	"time"

	"ssh-tunnel/internal/config"
)

// HysteriaTunnel implements the Tunnel interface for Hysteria protocol
type HysteriaTunnel struct {
	server config.Server
	status *TunnelStatus
}

// NewHysteriaTunnel creates a new Hysteria tunnel
func NewHysteriaTunnel(server config.Server) *HysteriaTunnel {
	return &HysteriaTunnel{
		server: server,
		status: &TunnelStatus{
			ServerName: server.Name,
			Status:     "disconnected",
		},
	}
}

// Start starts the Hysteria tunnel
func (t *HysteriaTunnel) Start(ctx context.Context) error {
	// TODO: Implement Hysteria protocol
	return fmt.Errorf("Hysteria protocol not yet implemented")
}

// Stop stops the Hysteria tunnel
func (t *HysteriaTunnel) Stop() error {
	t.status.Status = "disconnected"
	return nil
}

// GetStatus returns the current status
func (t *HysteriaTunnel) GetStatus() *TunnelStatus {
	statusCopy := *t.status
	return &statusCopy
}

// GetName returns the tunnel name
func (t *HysteriaTunnel) GetName() string {
	return t.server.Name
}

// Test tests the connection
func (t *HysteriaTunnel) Test() (time.Duration, error) {
	return 0, fmt.Errorf("Hysteria test not yet implemented")
}
