package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// TransportType represents different tunnel transport protocols
type TransportType string

const (
	TransportSSH       TransportType = "ssh"
	TransportHysteria  TransportType = "hysteria"
	TransportV2Ray     TransportType = "v2ray"
	TransportWireGuard TransportType = "wireguard"
	TransportTrojan    TransportType = "trojan"
	TransportVLESS     TransportType = "vless"
	TransportVMess     TransportType = "vmess"
)

// ProxyType represents proxy types
type ProxyType string

const (
	ProxySOCKS5 ProxyType = "socks5"
	ProxyHTTP   ProxyType = "http"
	ProxyHTTPS  ProxyType = "https"
)

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	EnableTLS         bool     `yaml:"enable_tls" json:"enable_tls"`
	TLSCertPath       string   `yaml:"tls_cert_path,omitempty" json:"tls_cert_path,omitempty"`
	TLSKeyPath        string   `yaml:"tls_key_path,omitempty" json:"tls_key_path,omitempty"`
	EnableAuth        bool     `yaml:"enable_auth" json:"enable_auth"`
	AuthTokens        []string `yaml:"auth_tokens,omitempty" json:"auth_tokens,omitempty"`
	EncryptConfig     bool     `yaml:"encrypt_config" json:"encrypt_config"`
	MasterPassword    string   `yaml:"master_password,omitempty" json:"master_password,omitempty"`
	FakeTLS           bool     `yaml:"fake_tls" json:"fake_tls"`
	Reality           bool     `yaml:"reality" json:"reality"`
	RealityTarget     string   `yaml:"reality_target,omitempty" json:"reality_target,omitempty"`
	RealityServerName string   `yaml:"reality_server_name,omitempty" json:"reality_server_name,omitempty"`
}

// HysteriaConfig specific configuration for Hysteria protocol
type HysteriaConfig struct {
	Protocol     string `yaml:"protocol" json:"protocol"` // "udp" or "faketcp"
	AuthString   string `yaml:"auth_string" json:"auth_string"`
	Bandwidth    string `yaml:"bandwidth,omitempty" json:"bandwidth,omitempty"` // "100mbps"
	ALPN         string `yaml:"alpn,omitempty" json:"alpn,omitempty"`
	Obfs         string `yaml:"obfs,omitempty" json:"obfs,omitempty"`
	ObfsPassword string `yaml:"obfs_password,omitempty" json:"obfs_password,omitempty"`
}

// V2RayConfig for V2Ray protocol configuration
type V2RayConfig struct {
	UUID       string            `yaml:"uuid" json:"uuid"`
	AlterID    int               `yaml:"alter_id,omitempty" json:"alter_id,omitempty"`
	Security   string            `yaml:"security,omitempty" json:"security,omitempty"`
	Network    string            `yaml:"network,omitempty" json:"network,omitempty"`
	HeaderType string            `yaml:"header_type,omitempty" json:"header_type,omitempty"`
	Path       string            `yaml:"path,omitempty" json:"path,omitempty"`
	Host       string            `yaml:"host,omitempty" json:"host,omitempty"`
	TLS        string            `yaml:"tls,omitempty" json:"tls,omitempty"`
	Headers    map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
}

// WireGuardConfig for WireGuard protocol
type WireGuardConfig struct {
	PrivateKey   string   `yaml:"private_key" json:"private_key"`
	PublicKey    string   `yaml:"public_key" json:"public_key"`
	PreSharedKey string   `yaml:"pre_shared_key,omitempty" json:"pre_shared_key,omitempty"`
	AllowedIPs   []string `yaml:"allowed_ips" json:"allowed_ips"`
	DNS          []string `yaml:"dns,omitempty" json:"dns,omitempty"`
	MTU          int      `yaml:"mtu,omitempty" json:"mtu,omitempty"`
}

// Server represents a tunnel server configuration
type Server struct {
	Name       string        `yaml:"name" json:"name"`
	Host       string        `yaml:"host" json:"host"`
	Port       string        `yaml:"port" json:"port"`
	User       string        `yaml:"user,omitempty" json:"user,omitempty"`
	Password   string        `yaml:"password,omitempty" json:"password,omitempty"`
	KeyPath    string        `yaml:"key_path,omitempty" json:"key_path,omitempty"`
	Transport  TransportType `yaml:"transport" json:"transport"`
	Proxy      ProxyType     `yaml:"proxy" json:"proxy"`
	LocalPort  int           `yaml:"local_port" json:"local_port"`
	Priority   int           `yaml:"priority,omitempty" json:"priority,omitempty"`
	MaxRetries int           `yaml:"max_retries,omitempty" json:"max_retries,omitempty"`
	Timeout    time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Enabled    bool          `yaml:"enabled" json:"enabled"`

	// Protocol-specific configurations
	Hysteria  *HysteriaConfig  `yaml:"hysteria,omitempty" json:"hysteria,omitempty"`
	V2Ray     *V2RayConfig     `yaml:"v2ray,omitempty" json:"v2ray,omitempty"`
	WireGuard *WireGuardConfig `yaml:"wireguard,omitempty" json:"wireguard,omitempty"`

	// Additional metadata
	Region string   `yaml:"region,omitempty" json:"region,omitempty"`
	Tags   []string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// RoutingRule defines routing rules for traffic
type RoutingRule struct {
	Type    string   `yaml:"type" json:"type"` // "domain", "ip", "geoip"
	Pattern string   `yaml:"pattern" json:"pattern"`
	Server  string   `yaml:"server,omitempty" json:"server,omitempty"`
	Action  string   `yaml:"action" json:"action"` // "proxy", "direct", "block"
	Domains []string `yaml:"domains,omitempty" json:"domains,omitempty"`
	IPs     []string `yaml:"ips,omitempty" json:"ips,omitempty"`
	GeoIP   []string `yaml:"geoip,omitempty" json:"geoip,omitempty"`
}

// MonitoringConfig for health monitoring
type MonitoringConfig struct {
	Enabled         bool          `yaml:"enabled" json:"enabled"`
	CheckInterval   time.Duration `yaml:"check_interval,omitempty" json:"check_interval,omitempty"`
	HealthEndpoint  string        `yaml:"health_endpoint,omitempty" json:"health_endpoint,omitempty"`
	MetricsEndpoint string        `yaml:"metrics_endpoint,omitempty" json:"metrics_endpoint,omitempty"`
	LogLevel        string        `yaml:"log_level,omitempty" json:"log_level,omitempty"`
	LogFile         string        `yaml:"log_file,omitempty" json:"log_file,omitempty"`
	MaxLogSize      string        `yaml:"max_log_size,omitempty" json:"max_log_size,omitempty"`
}

// APIConfig for REST API server
type APIConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled"`
	Host       string `yaml:"host" json:"host"`
	Port       int    `yaml:"port" json:"port"`
	EnableCORS bool   `yaml:"enable_cors" json:"enable_cors"`
	RateLimit  int    `yaml:"rate_limit,omitempty" json:"rate_limit,omitempty"`
}

// Config represents the main configuration structure
type Config struct {
	Version    string           `yaml:"version" json:"version"`
	Servers    []Server         `yaml:"servers" json:"servers"`
	Security   SecurityConfig   `yaml:"security" json:"security"`
	Routing    []RoutingRule    `yaml:"routing,omitempty" json:"routing,omitempty"`
	Monitoring MonitoringConfig `yaml:"monitoring" json:"monitoring"`
	API        APIConfig        `yaml:"api" json:"api"`

	// Auto-selection settings
	AutoSelect      bool          `yaml:"auto_select" json:"auto_select"`
	SelectionMethod string        `yaml:"selection_method,omitempty" json:"selection_method,omitempty"` // "latency", "load", "random"
	LatencyTimeout  time.Duration `yaml:"latency_timeout,omitempty" json:"latency_timeout,omitempty"`

	// Failover settings
	EnableFailover  bool          `yaml:"enable_failover" json:"enable_failover"`
	FailoverTimeout time.Duration `yaml:"failover_timeout,omitempty" json:"failover_timeout,omitempty"`
}

// LoadConfig loads configuration from file with decryption support
func LoadConfig(configPath string) (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", configPath)
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// Check if config is encrypted
	if isEncrypted(data) {
		password := os.Getenv("CONFIG_PASSWORD")
		if password == "" {
			return nil, fmt.Errorf("encrypted config detected but CONFIG_PASSWORD not set")
		}

		data, err = decrypt(data, password)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt config: %v", err)
		}
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	// Set default values
	setDefaults(&config)

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %v", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to file with optional encryption
func SaveConfig(config *Config, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	// Encrypt if required
	if config.Security.EncryptConfig {
		password := config.Security.MasterPassword
		if password == "" {
			password = os.Getenv("CONFIG_PASSWORD")
		}
		if password == "" {
			return fmt.Errorf("encryption requested but no password provided")
		}

		data, err = encrypt(data, password)
		if err != nil {
			return fmt.Errorf("failed to encrypt config: %v", err)
		}
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	return os.WriteFile(configPath, data, 0600)
}

// setDefaults sets default values for configuration
func setDefaults(config *Config) {
	if config.Version == "" {
		config.Version = "1.0"
	}

	if config.LatencyTimeout == 0 {
		config.LatencyTimeout = 5 * time.Second
	}

	if config.FailoverTimeout == 0 {
		config.FailoverTimeout = 30 * time.Second
	}

	if config.SelectionMethod == "" {
		config.SelectionMethod = "latency"
	}

	// Set defaults for monitoring
	if config.Monitoring.Enabled && config.Monitoring.CheckInterval == 0 {
		config.Monitoring.CheckInterval = 30 * time.Second
	}

	if config.Monitoring.LogLevel == "" {
		config.Monitoring.LogLevel = "info"
	}

	// Set defaults for API
	if config.API.Host == "" {
		config.API.Host = "localhost"
	}
	if config.API.Port == 0 {
		config.API.Port = 8888
	}

	// Set defaults for each server
	for i := range config.Servers {
		server := &config.Servers[i]

		if server.Transport == "" {
			server.Transport = TransportSSH
		}

		if server.Proxy == "" {
			server.Proxy = ProxySOCKS5
		}

		if server.LocalPort == 0 {
			server.LocalPort = 8080 + i
		}

		if server.MaxRetries == 0 {
			server.MaxRetries = 3
		}

		if server.Timeout == 0 {
			server.Timeout = 10 * time.Second
		}

		if server.Name == "" {
			server.Name = fmt.Sprintf("server-%d", i+1)
		}
	}
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	if len(config.Servers) == 0 {
		return fmt.Errorf("no servers configured")
	}

	// Validate each server
	for i, server := range config.Servers {
		if server.Host == "" {
			return fmt.Errorf("server %d: host is required", i)
		}

		if server.Port == "" {
			return fmt.Errorf("server %d: port is required", i)
		}

		// Validate transport-specific requirements
		switch server.Transport {
		case TransportSSH:
			if server.User == "" {
				return fmt.Errorf("server %d: user is required for SSH transport", i)
			}
			if server.Password == "" && server.KeyPath == "" {
				return fmt.Errorf("server %d: either password or key_path is required for SSH", i)
			}

		case TransportHysteria:
			if server.Hysteria == nil {
				return fmt.Errorf("server %d: hysteria configuration is required", i)
			}
			if server.Hysteria.AuthString == "" {
				return fmt.Errorf("server %d: hysteria auth_string is required", i)
			}

		case TransportV2Ray, TransportVMess, TransportVLESS:
			if server.V2Ray == nil {
				return fmt.Errorf("server %d: v2ray configuration is required", i)
			}
			if server.V2Ray.UUID == "" {
				return fmt.Errorf("server %d: v2ray UUID is required", i)
			}

		case TransportWireGuard:
			if server.WireGuard == nil {
				return fmt.Errorf("server %d: wireguard configuration is required", i)
			}
			if server.WireGuard.PrivateKey == "" || server.WireGuard.PublicKey == "" {
				return fmt.Errorf("server %d: wireguard private_key and public_key are required", i)
			}
		}
	}

	return nil
}

// Encryption/Decryption functions
func isEncrypted(data []byte) bool {
	return strings.HasPrefix(string(data), "ENC:")
}

func encrypt(data []byte, password string) ([]byte, error) {
	key := sha256.Sum256([]byte(password))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return []byte("ENC:" + hex.EncodeToString(ciphertext)), nil
}

func decrypt(data []byte, password string) ([]byte, error) {
	if !strings.HasPrefix(string(data), "ENC:") {
		return nil, fmt.Errorf("not encrypted data")
	}

	encryptedHex := strings.TrimPrefix(string(data), "ENC:")
	encrypted, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return nil, err
	}

	key := sha256.Sum256([]byte(password))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
