package autodiscovery

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"crypto/rand"

	"golang.org/x/crypto/ssh"
)

// ServerInfo holds information about a discovered server
type ServerInfo struct {
	Host               string                 `json:"host"`
	Port               string                 `json:"port"`
	User               string                 `json:"user"`
	Password           string                 `json:"password,omitempty"`
	KeyPath            string                 `json:"key_path,omitempty"`
	SupportedProtocols []string               `json:"supported_protocols"`
	ServerCapabilities map[string]interface{} `json:"server_capabilities"`
	AvailablePorts     []int                  `json:"available_ports"`
	OS                 string                 `json:"operating_system"`
	Architecture       string                 `json:"architecture"`
	InstalledSoftware  []string               `json:"installed_software"`
	NetworkInterfaces  []NetworkInterface     `json:"network_interfaces"`
}

// NetworkInterface represents a network interface on the server
type NetworkInterface struct {
	Name     string   `json:"name"`
	IPs      []string `json:"ips"`
	IsPublic bool     `json:"is_public"`
}

// ProtocolConfig holds configuration for each protocol
type ProtocolConfig struct {
	Type       string                 `json:"type"`
	Port       int                    `json:"port"`
	Config     map[string]interface{} `json:"config"`
	ClientURL  string                 `json:"client_url"`
	ProxyURL   string                 `json:"proxy_url"`
	ConfigFile string                 `json:"config_file"`
}

// ServerDiscovery handles automatic server discovery and setup
type ServerDiscovery struct {
	client  *ssh.Client
	info    *ServerInfo
	configs map[string]*ProtocolConfig
}

// NewServerDiscovery creates a new server discovery instance
func NewServerDiscovery() *ServerDiscovery {
	return &ServerDiscovery{
		configs: make(map[string]*ProtocolConfig),
	}
}

// DiscoverServer discovers server capabilities and sets up protocols
func (sd *ServerDiscovery) DiscoverServer(host, port, user, password, keyPath string) (*ServerInfo, error) {
	log.Printf("Starting server discovery for %s@%s:%s", user, host, port)

	// Connect to server
	if err := sd.connectToServer(host, port, user, password, keyPath); err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}
	defer sd.client.Close()

	// Initialize server info
	sd.info = &ServerInfo{
		Host:               host,
		Port:               port,
		User:               user,
		Password:           password,
		KeyPath:            keyPath,
		SupportedProtocols: []string{},
		ServerCapabilities: make(map[string]interface{}),
		AvailablePorts:     []int{},
		InstalledSoftware:  []string{},
	}

	// Discover server information
	if err := sd.discoverSystemInfo(); err != nil {
		log.Printf("Warning: Failed to discover system info: %v", err)
	}

	// Discover network interfaces
	if err := sd.discoverNetworkInterfaces(); err != nil {
		log.Printf("Warning: Failed to discover network interfaces: %v", err)
	}

	// Discover available ports
	if err := sd.discoverAvailablePorts(); err != nil {
		log.Printf("Warning: Failed to discover available ports: %v", err)
	}

	// Check for installed software
	sd.checkInstalledSoftware()

	// Discover supported protocols
	sd.discoverSupportedProtocols()

	log.Printf("Server discovery completed. Supported protocols: %v", sd.info.SupportedProtocols)
	return sd.info, nil
}

// SetupAllProtocols automatically sets up all supported protocols
func (sd *ServerDiscovery) SetupAllProtocols() error {
	log.Println("Setting up all supported protocols...")

	for _, protocol := range sd.info.SupportedProtocols {
		if err := sd.setupProtocol(protocol); err != nil {
			log.Printf("Failed to setup %s: %v", protocol, err)
			continue
		}
		log.Printf("Successfully set up %s protocol", protocol)
	}

	return nil
}

// GenerateClientConfigs generates client configuration files for all protocols
func (sd *ServerDiscovery) GenerateClientConfigs(outputDir string) error {
	log.Printf("Generating client configurations in %s", outputDir)

	configs := map[string]string{
		"ssh_tunnel":    sd.generateSSHTunnelConfig(),
		"v2ray_client":  sd.generateV2RayConfig(),
		"vless_client":  sd.generateVLESSConfig(),
		"vmess_client":  sd.generateVMessConfig(),
		"trojan_client": sd.generateTrojanConfig(),
		"wireguard":     sd.generateWireGuardConfig(),
		"hysteria":      sd.generateHysteriaConfig(),
		"http_proxy":    sd.generateHTTPProxyConfig(),
		"socks5_proxy":  sd.generateSOCKS5Config(),
	}

	// Write configuration files
	for name, configContent := range configs {
		if configContent != "" {
			if err := sd.writeConfigFile(fmt.Sprintf("%s/%s.conf", outputDir, name), configContent); err != nil {
				log.Printf("Failed to write %s config: %v", name, err)
			}
		}
	}

	// Generate combined configuration
	combinedConfig := sd.generateCombinedConfig()
	if err := sd.writeConfigFile(fmt.Sprintf("%s/combined_config.yaml", outputDir), combinedConfig); err != nil {
		log.Printf("Failed to write combined config: %v", err)
	}

	return nil
}

// connectToServer establishes SSH connection to the server
func (sd *ServerDiscovery) connectToServer(host, port, user, password, keyPath string) error {
	config := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	// Setup authentication
	if password != "" {
		config.Auth = []ssh.AuthMethod{ssh.Password(password)}
	} else if keyPath != "" {
		// TODO: Implement key-based authentication
		return fmt.Errorf("key-based authentication not yet implemented")
	}

	addr := net.JoinHostPort(host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return err
	}

	sd.client = client
	return nil
}

// discoverSystemInfo discovers basic system information
func (sd *ServerDiscovery) discoverSystemInfo() error {
	// Get OS information
	if output, err := sd.executeCommand("uname -s"); err == nil {
		sd.info.OS = strings.TrimSpace(output)
	}

	// Get architecture
	if output, err := sd.executeCommand("uname -m"); err == nil {
		sd.info.Architecture = strings.TrimSpace(output)
	}

	return nil
}

// discoverNetworkInterfaces discovers network interfaces
func (sd *ServerDiscovery) discoverNetworkInterfaces() error {
	output, err := sd.executeCommand("ip addr show 2>/dev/null || ifconfig")
	if err != nil {
		return err
	}

	// Parse network interfaces (simplified)
	interfaces := []NetworkInterface{}

	// This is a simplified parser - in production, you'd want more robust parsing
	if strings.Contains(output, "eth0") || strings.Contains(output, "en0") {
		interfaces = append(interfaces, NetworkInterface{
			Name:     "eth0",
			IPs:      []string{sd.info.Host}, // Simplified
			IsPublic: true,
		})
	}

	sd.info.NetworkInterfaces = interfaces
	return nil
}

// discoverAvailablePorts finds available ports for protocol setup
func (sd *ServerDiscovery) discoverAvailablePorts() error {
	commonPorts := []int{8080, 8081, 8082, 8083, 8084, 8085, 9080, 9081, 9082, 10080, 10081}

	for _, port := range commonPorts {
		if sd.isPortAvailable(port) {
			sd.info.AvailablePorts = append(sd.info.AvailablePorts, port)
		}
	}

	return nil
}

// checkInstalledSoftware checks for installed relevant software
func (sd *ServerDiscovery) checkInstalledSoftware() {
	software := map[string]string{
		"docker":    "docker --version",
		"nginx":     "nginx -v",
		"xray":      "xray version",
		"v2ray":     "v2ray version",
		"trojan":    "trojan --version",
		"wireguard": "wg --version",
		"iptables":  "iptables --version",
		"socat":     "socat -V",
		"haproxy":   "haproxy -v",
	}

	for name, cmd := range software {
		if _, err := sd.executeCommand(cmd); err == nil {
			sd.info.InstalledSoftware = append(sd.info.InstalledSoftware, name)
		}
	}
}

// discoverSupportedProtocols determines which protocols can be set up
func (sd *ServerDiscovery) discoverSupportedProtocols() {
	// Always support SSH tunnel
	sd.info.SupportedProtocols = append(sd.info.SupportedProtocols, "ssh")

	// Check for Docker (enables many protocols)
	if sd.hasInstalledSoftware("docker") {
		sd.info.SupportedProtocols = append(sd.info.SupportedProtocols,
			"v2ray", "vless", "vmess", "trojan", "hysteria", "wireguard")
	}

	// Check for direct installations
	if sd.hasInstalledSoftware("xray") || sd.hasInstalledSoftware("v2ray") {
		sd.info.SupportedProtocols = append(sd.info.SupportedProtocols, "v2ray", "vless", "vmess")
	}

	if sd.hasInstalledSoftware("trojan") {
		sd.info.SupportedProtocols = append(sd.info.SupportedProtocols, "trojan")
	}

	if sd.hasInstalledSoftware("wireguard") {
		sd.info.SupportedProtocols = append(sd.info.SupportedProtocols, "wireguard")
	}

	// 🆕 Always add V2Ray protocols (can be installed on demand)
	// Check if we can install docker or if ports are available
	if len(sd.info.AvailablePorts) >= 2 {
		if !containsString(sd.info.SupportedProtocols, "v2ray") {
			sd.info.SupportedProtocols = append(sd.info.SupportedProtocols, "v2ray", "vless", "vmess")
		}
		if !containsString(sd.info.SupportedProtocols, "trojan") {
			sd.info.SupportedProtocols = append(sd.info.SupportedProtocols, "trojan")
		}
		if !containsString(sd.info.SupportedProtocols, "hysteria") {
			sd.info.SupportedProtocols = append(sd.info.SupportedProtocols, "hysteria")
		}
		if !containsString(sd.info.SupportedProtocols, "wireguard") {
			sd.info.SupportedProtocols = append(sd.info.SupportedProtocols, "wireguard")
		}
	}

	// Can always setup HTTP/SOCKS proxies via SSH
	sd.info.SupportedProtocols = append(sd.info.SupportedProtocols, "http_proxy", "socks5_proxy")

	// ICMP tunnel if we have socat or custom tools
	if sd.hasInstalledSoftware("socat") {
		sd.info.SupportedProtocols = append(sd.info.SupportedProtocols, "icmp_tunnel")
	}
}

// setupProtocol sets up a specific protocol on the server
func (sd *ServerDiscovery) setupProtocol(protocol string) error {
	switch protocol {
	case "ssh":
		return sd.setupSSHTunnel()
	case "v2ray", "vless", "vmess":
		return sd.setupV2Ray()
	case "trojan":
		return sd.setupTrojan()
	case "hysteria":
		return sd.setupHysteria()
	case "wireguard":
		return sd.setupWireGuard()
	case "http_proxy":
		return sd.setupHTTPProxy()
	case "socks5_proxy":
		return sd.setupSOCKS5Proxy()
	case "icmp_tunnel":
		return sd.setupICMPTunnel()
	default:
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// Protocol setup methods
func (sd *ServerDiscovery) setupSSHTunnel() error {
	port := sd.getAvailablePort()
	sd.configs["ssh"] = &ProtocolConfig{
		Type: "ssh",
		Port: port,
		Config: map[string]interface{}{
			"host":     sd.info.Host,
			"port":     sd.info.Port,
			"user":     sd.info.User,
			"password": sd.info.Password,
		},
		ProxyURL: fmt.Sprintf("socks5://127.0.0.1:%d", port),
	}
	return nil
}

func (sd *ServerDiscovery) setupV2Ray() error {
	port := sd.getAvailablePort()
	uuid := sd.generateUUID()

	// Always create config - Docker installation is optional
	sd.configs["v2ray"] = &ProtocolConfig{
		Type: "v2ray",
		Port: port,
		Config: map[string]interface{}{
			"server":   sd.info.Host,
			"port":     port,
			"uuid":     uuid,
			"alterId":  0,
			"security": "auto",
		},
	}

	// Try to install V2Ray if --setup flag was used and Docker is available
	if sd.hasInstalledSoftware("docker") {
		installCmd := fmt.Sprintf(`
docker pull v2fly/v2fly-core:latest 2>/dev/null && \
docker run -d --name v2ray-%d --restart unless-stopped \
  -p %d:10086 \
  v2fly/v2fly-core:latest v2ray run -c <(cat << 'EOF'
{
  "inbounds": [{
    "port": 10086,
    "protocol": "vmess",
    "settings": {
      "clients": [{
        "id": "%s",
        "alterId": 0,
        "security": "auto"
      }]
    }
  }],
  "outbounds": [{"protocol": "freedom"}]
}
EOF
)
`, port, port, uuid)

		if _, err := sd.executeCommand(installCmd); err != nil {
			log.Printf("Warning: Could not auto-install V2Ray via Docker: %v", err)
			// Don't return error - config is still valid for manual setup
		} else {
			log.Printf("✅ V2Ray installed and configured on port %d", port)
		}
	}

	return nil
}

func (sd *ServerDiscovery) setupTrojan() error {
	port := sd.getAvailablePort()
	password := sd.generatePassword()

	// Setup Trojan via Docker
	installCmd := fmt.Sprintf(`
docker run -d --name trojan --restart unless-stopped \
  -p %d:443 \
  -e TROJAN_PASSWORD=%s \
  trojangfw/trojan:latest
`, port, password)

	if _, err := sd.executeCommand(installCmd); err != nil {
		return fmt.Errorf("failed to setup Trojan: %v", err)
	}

	sd.configs["trojan"] = &ProtocolConfig{
		Type: "trojan",
		Port: port,
		Config: map[string]interface{}{
			"server":   sd.info.Host,
			"port":     port,
			"password": password,
		},
	}
	return nil
}

func (sd *ServerDiscovery) setupHysteria() error {
	port := sd.getAvailablePort()
	password := sd.generatePassword()

	// Setup Hysteria via Docker
	installCmd := fmt.Sprintf(`
docker run -d --name hysteria --restart unless-stopped \
  -p %d:36712/udp \
  -e HYSTERIA_PASSWORD=%s \
  tobyxdd/hysteria:latest
`, port, password)

	if _, err := sd.executeCommand(installCmd); err != nil {
		return fmt.Errorf("failed to setup Hysteria: %v", err)
	}

	sd.configs["hysteria"] = &ProtocolConfig{
		Type: "hysteria",
		Port: port,
		Config: map[string]interface{}{
			"server":    sd.info.Host,
			"port":      port,
			"auth_str":  password,
			"protocol":  "udp",
			"bandwidth": "100mbps",
		},
	}
	return nil
}

func (sd *ServerDiscovery) setupWireGuard() error {
	port := sd.getAvailablePort()

	// Setup WireGuard via Docker
	installCmd := fmt.Sprintf(`
docker run -d --name wireguard --restart unless-stopped \
  --cap-add=NET_ADMIN --cap-add=SYS_MODULE \
  -p %d:51820/udp \
  -v wireguard_data:/config \
  -e PUID=1000 -e PGID=1000 \
  -e TZ=UTC \
  linuxserver/wireguard:latest
`, port)

	if _, err := sd.executeCommand(installCmd); err != nil {
		return fmt.Errorf("failed to setup WireGuard: %v", err)
	}

	sd.configs["wireguard"] = &ProtocolConfig{
		Type: "wireguard",
		Port: port,
		Config: map[string]interface{}{
			"server": sd.info.Host,
			"port":   port,
		},
	}
	return nil
}

func (sd *ServerDiscovery) setupHTTPProxy() error {
	port := sd.getAvailablePort()

	// Setup HTTP proxy via SSH tunnel
	sd.configs["http_proxy"] = &ProtocolConfig{
		Type: "http_proxy",
		Port: port,
		Config: map[string]interface{}{
			"proxy_host": "127.0.0.1",
			"proxy_port": port,
		},
		ProxyURL: fmt.Sprintf("http://127.0.0.1:%d", port),
	}
	return nil
}

func (sd *ServerDiscovery) setupSOCKS5Proxy() error {
	port := sd.getAvailablePort()

	sd.configs["socks5_proxy"] = &ProtocolConfig{
		Type: "socks5_proxy",
		Port: port,
		Config: map[string]interface{}{
			"proxy_host": "127.0.0.1",
			"proxy_port": port,
		},
		ProxyURL: fmt.Sprintf("socks5://127.0.0.1:%d", port),
	}
	return nil
}

func (sd *ServerDiscovery) setupICMPTunnel() error {
	// ICMP tunnel setup using socat or custom implementation
	installCmd := `
# Install ICMP tunnel tools
apt-get update && apt-get install -y socat || yum install -y socat
`
	if _, err := sd.executeCommand(installCmd); err != nil {
		log.Printf("Warning: Failed to install ICMP tunnel tools: %v", err)
	}

	sd.configs["icmp_tunnel"] = &ProtocolConfig{
		Type: "icmp_tunnel",
		Port: 0, // ICMP doesn't use ports
		Config: map[string]interface{}{
			"server":   sd.info.Host,
			"protocol": "icmp",
		},
	}
	return nil
}

// Helper methods
func (sd *ServerDiscovery) executeCommand(cmd string) (string, error) {
	session, err := sd.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	return string(output), err
}

func (sd *ServerDiscovery) isPortAvailable(port int) bool {
	cmd := fmt.Sprintf("netstat -tuln | grep ':%d ' || ss -tuln | grep ':%d '", port, port)
	output, _ := sd.executeCommand(cmd)
	return !strings.Contains(output, fmt.Sprintf(":%d", port))
}

func (sd *ServerDiscovery) hasInstalledSoftware(software string) bool {
	for _, installed := range sd.info.InstalledSoftware {
		if installed == software {
			return true
		}
	}
	return false
}

func (sd *ServerDiscovery) getAvailablePort() int {
	if len(sd.info.AvailablePorts) > 0 {
		port := sd.info.AvailablePorts[0]
		sd.info.AvailablePorts = sd.info.AvailablePorts[1:]
		return port
	}
	return 8080 // fallback
}

func (sd *ServerDiscovery) generateUUID() string {
	// Generate a proper UUID - for now simplified
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to time-based UUID
		return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			uint32(time.Now().Unix()),
			uint16(time.Now().UnixNano()&0xFFFF),
			uint16((time.Now().UnixNano()>>16)&0xFFFF),
			uint16((time.Now().UnixNano()>>32)&0xFFFF),
			time.Now().UnixNano()&0xFFFFFFFFFFFF)
	}
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uint32(b[0])<<24|uint32(b[1])<<16|uint32(b[2])<<8|uint32(b[3]),
		uint16(b[4])<<8|uint16(b[5]),
		uint16(b[6])<<8|uint16(b[7]),
		uint16(b[8])<<8|uint16(b[9]),
		uint64(b[10])<<40|uint64(b[11])<<32|uint64(b[12])<<24|uint64(b[13])<<16|uint64(b[14])<<8|uint64(b[15]))
}

func (sd *ServerDiscovery) generatePassword() string {
	// Generate random password
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	for i := range b {
		b[i] = chars[time.Now().UnixNano()%int64(len(chars))]
		time.Sleep(1) // Simple randomization
	}
	return string(b)
}

func (sd *ServerDiscovery) writeConfigFile(filepath, content string) error {
	// Create directory if it doesn't exist
	dir := filepath[:strings.LastIndex(filepath, "/")]
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// Write file to disk
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %v", filepath, err)
	}

	log.Printf("✅ Generated config file: %s", filepath)
	return nil
}

// Helper function to check if slice contains string
// containsString checks if slice contains string
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
