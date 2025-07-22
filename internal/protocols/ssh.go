package protocols

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"time"

	"ssh-tunnel/internal/config"

	"golang.org/x/crypto/ssh"
)

// SSHTunnel implements the Tunnel interface for SSH connections
type SSHTunnel struct {
	server   config.Server
	client   *ssh.Client
	listener net.Listener
	status   *TunnelStatus
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewSSHTunnel creates a new SSH tunnel
func NewSSHTunnel(server config.Server) *SSHTunnel {
	return &SSHTunnel{
		server: server,
		status: &TunnelStatus{
			ServerName: server.Name,
			Status:     "disconnected",
		},
	}
}

// Start starts the SSH tunnel
func (t *SSHTunnel) Start(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.ctx, t.cancel = context.WithCancel(ctx)
	t.status.Status = "connecting"
	t.status.StartTime = time.Now()

	// Create SSH client configuration
	config := &ssh.ClientConfig{
		User:            t.server.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // In production, use proper host key verification
		Timeout:         t.server.Timeout,
	}

	// Add authentication method
	if t.server.Password != "" {
		config.Auth = []ssh.AuthMethod{
			ssh.Password(t.server.Password),
		}
	} else if t.server.KeyPath != "" {
		// TODO: Implement key-based authentication
		return fmt.Errorf("key-based authentication not yet implemented")
	} else {
		return fmt.Errorf("no authentication method provided")
	}

	// Connect to SSH server
	addr := fmt.Sprintf("%s:%s", t.server.Host, t.server.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		t.status.Status = "error"
		t.status.LastError = err.Error()
		return fmt.Errorf("failed to connect to SSH server: %v", err)
	}

	t.client = client
	t.status.Status = "connected"

	// Start the appropriate proxy type
	switch t.server.Proxy {
	case "socks5":
		return t.startSOCKS5()
	case "http":
		return t.startHTTP()
	default:
		return fmt.Errorf("unsupported proxy type: %s", t.server.Proxy)
	}
}

// Stop stops the SSH tunnel
func (t *SSHTunnel) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cancel != nil {
		t.cancel()
	}

	if t.listener != nil {
		t.listener.Close()
	}

	if t.client != nil {
		t.client.Close()
		t.client = nil
	}

	t.status.Status = "disconnected"
	return nil
}

// GetStatus returns the current status
func (t *SSHTunnel) GetStatus() *TunnelStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Return a copy
	statusCopy := *t.status
	return &statusCopy
}

// GetName returns the tunnel name
func (t *SSHTunnel) GetName() string {
	return t.server.Name
}

// Test tests the connection and measures latency
func (t *SSHTunnel) Test() (time.Duration, error) {
	return t.pingTest()
}

// startSOCKS5 starts a SOCKS5 proxy
func (t *SSHTunnel) startSOCKS5() error {
	// Create local listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", t.server.LocalPort))
	if err != nil {
		return fmt.Errorf("failed to create local listener: %v", err)
	}

	t.listener = listener
	log.Printf("SOCKS5 proxy started on port %d for %s", t.server.LocalPort, t.server.Name)

	// Accept connections
	go t.acceptConnections()

	return nil
}

// startHTTP starts an HTTP proxy
func (t *SSHTunnel) startHTTP() error {
	// Create local listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", t.server.LocalPort))
	if err != nil {
		return fmt.Errorf("failed to create local listener: %v", err)
	}

	t.listener = listener
	log.Printf("HTTP proxy started on port %d for %s", t.server.LocalPort, t.server.Name)

	// Accept connections
	go t.acceptConnections()

	return nil
}

// acceptConnections accepts and handles incoming connections
func (t *SSHTunnel) acceptConnections() {
	defer t.listener.Close()

	for {
		select {
		case <-t.ctx.Done():
			return
		default:
			conn, err := t.listener.Accept()
			if err != nil {
				if t.ctx.Err() != nil {
					return // Context cancelled
				}
				log.Printf("Error accepting connection: %v", err)
				continue
			}

			go t.handleConnection(conn)
		}
	}
}

// handleConnection handles a single connection
func (t *SSHTunnel) handleConnection(localConn net.Conn) {
	defer localConn.Close()

	// This is a simplified implementation
	// In a full implementation, you would parse SOCKS5/HTTP requests
	// and establish remote connections through the SSH tunnel

	log.Printf("Handling connection for %s", t.server.Name)

	// For now, just close the connection
	// TODO: Implement full SOCKS5/HTTP proxy logic
}

// pingTest performs a ping test to measure latency
func (t *SSHTunnel) pingTest() (time.Duration, error) {
	var cmd *exec.Cmd
	var re *regexp.Regexp

	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "-n", "1", t.server.Host)
		re = regexp.MustCompile(`time[=<]?(\d+)ms`)
	} else {
		cmd = exec.Command("ping", "-c", "1", t.server.Host)
		re = regexp.MustCompile(`time[=<]([\d.]+) ms`)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("ping failed: %v", err)
	}

	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		// Fallback to simple connection test
		return t.connectionTest()
	}

	latencyFloat, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return t.connectionTest()
	}

	return time.Duration(latencyFloat * float64(time.Millisecond)), nil
}

// connectionTest performs a simple connection test
func (t *SSHTunnel) connectionTest() (time.Duration, error) {
	start := time.Now()

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", t.server.Host, t.server.Port), 5*time.Second)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	return time.Since(start), nil
}
