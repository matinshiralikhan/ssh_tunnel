package protocols

import (
	"context"
	"fmt"
	"time"

	"ssh-tunnel/internal/config"
)

// V2RayTunnel implements the Tunnel interface for V2Ray protocol
type V2RayTunnel struct {
	server config.Server
	status *TunnelStatus
}

// NewV2RayTunnel creates a new V2Ray tunnel
func NewV2RayTunnel(server config.Server) *V2RayTunnel {
	return &V2RayTunnel{
		server: server,
		status: &TunnelStatus{
			ServerName: server.Name,
			Status:     "disconnected",
		},
	}
}

// Start starts the V2Ray tunnel
func (t *V2RayTunnel) Start(ctx context.Context) error {
	// TODO: Implement V2Ray protocol
	return fmt.Errorf("V2Ray protocol not yet implemented")
}

// Stop stops the V2Ray tunnel
func (t *V2RayTunnel) Stop() error {
	t.status.Status = "disconnected"
	return nil
}

// GetStatus returns the current status
func (t *V2RayTunnel) GetStatus() *TunnelStatus {
	statusCopy := *t.status
	return &statusCopy
}

// GetName returns the tunnel name
func (t *V2RayTunnel) GetName() string {
	return t.server.Name
}

// Test tests the connection
func (t *V2RayTunnel) Test() (time.Duration, error) {
	return 0, fmt.Errorf("V2Ray test not yet implemented")
}

// WireGuardTunnel implements the Tunnel interface for WireGuard protocol
type WireGuardTunnel struct {
	server config.Server
	status *TunnelStatus
}

// NewWireGuardTunnel creates a new WireGuard tunnel
func NewWireGuardTunnel(server config.Server) *WireGuardTunnel {
	return &WireGuardTunnel{
		server: server,
		status: &TunnelStatus{
			ServerName: server.Name,
			Status:     "disconnected",
		},
	}
}

// Start starts the WireGuard tunnel
func (t *WireGuardTunnel) Start(ctx context.Context) error {
	// TODO: Implement WireGuard protocol
	return fmt.Errorf("WireGuard protocol not yet implemented")
}

// Stop stops the WireGuard tunnel
func (t *WireGuardTunnel) Stop() error {
	t.status.Status = "disconnected"
	return nil
}

// GetStatus returns the current status
func (t *WireGuardTunnel) GetStatus() *TunnelStatus {
	statusCopy := *t.status
	return &statusCopy
}

// GetName returns the tunnel name
func (t *WireGuardTunnel) GetName() string {
	return t.server.Name
}

// Test tests the connection
func (t *WireGuardTunnel) Test() (time.Duration, error) {
	return 0, fmt.Errorf("WireGuard test not yet implemented")
}

// TrojanTunnel implements the Tunnel interface for Trojan protocol
type TrojanTunnel struct {
	server config.Server
	status *TunnelStatus
}

// NewTrojanTunnel creates a new Trojan tunnel
func NewTrojanTunnel(server config.Server) *TrojanTunnel {
	return &TrojanTunnel{
		server: server,
		status: &TunnelStatus{
			ServerName: server.Name,
			Status:     "disconnected",
		},
	}
}

// Start starts the Trojan tunnel
func (t *TrojanTunnel) Start(ctx context.Context) error {
	// TODO: Implement Trojan protocol
	return fmt.Errorf("Trojan protocol not yet implemented")
}

// Stop stops the Trojan tunnel
func (t *TrojanTunnel) Stop() error {
	t.status.Status = "disconnected"
	return nil
}

// GetStatus returns the current status
func (t *TrojanTunnel) GetStatus() *TunnelStatus {
	statusCopy := *t.status
	return &statusCopy
}

// GetName returns the tunnel name
func (t *TrojanTunnel) GetName() string {
	return t.server.Name
}

// Test tests the connection
func (t *TrojanTunnel) Test() (time.Duration, error) {
	return 0, fmt.Errorf("Trojan test not yet implemented")
}
