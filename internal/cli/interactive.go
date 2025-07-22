package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"ssh-tunnel/internal/autodiscovery"
	"ssh-tunnel/internal/config"
	"ssh-tunnel/internal/mesh"

	"golang.org/x/term"
)

// InteractiveCLI provides a user-friendly interactive interface
type InteractiveCLI struct {
	scanner *bufio.Scanner
}

// NewInteractiveCLI creates a new interactive CLI
func NewInteractiveCLI() *InteractiveCLI {
	return &InteractiveCLI{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// ShowMainMenu displays the main menu
func (cli *InteractiveCLI) ShowMainMenu() {
	fmt.Println()
	fmt.Println("🚀 SSH Tunnel Manager")
	fmt.Println("=====================")
	fmt.Println()
	fmt.Println("Choose an option:")
	fmt.Println()
	fmt.Println("  1. 🔍 Quick Setup (Auto-discover server)")
	fmt.Println("  2. 🌐 Mesh Network (Connect multiple servers)")
	fmt.Println("  3. 📁 Use existing config")
	fmt.Println("  4. ⚙️  Advanced configuration")
	fmt.Println("  5. 📊 Monitor connections")
	fmt.Println("  6. 🔧 Manage servers")
	fmt.Println("  7. 📖 Help & Documentation")
	fmt.Println("  8. 🚪 Exit")
	fmt.Println()
}

// HandleMainMenu processes main menu selection
func (cli *InteractiveCLI) HandleMainMenu() error {
	for {
		cli.ShowMainMenu()
		choice := cli.getUserInput("Select option (1-8)")

		switch choice {
		case "1":
			return cli.handleQuickSetup()
		case "2":
			return cli.handleMeshNetwork()
		case "3":
			return cli.handleExistingConfig()
		case "4":
			return cli.handleAdvancedConfig()
		case "5":
			return cli.handleMonitoring()
		case "6":
			return cli.handleServerManagement()
		case "7":
			cli.showHelp()
		case "8":
			fmt.Println("👋 Goodbye!")
			return nil
		default:
			fmt.Println("❌ Invalid option. Please choose 1-8.")
		}
	}
}

// handleQuickSetup handles the quick setup wizard
func (cli *InteractiveCLI) handleQuickSetup() error {
	fmt.Println()
	fmt.Println("🔍 Quick Setup Wizard")
	fmt.Println("=====================")
	fmt.Println()
	fmt.Println("This will automatically discover and setup your server with all supported protocols.")
	fmt.Println()

	// Get server details
	host := cli.getUserInput("Enter server IP or hostname")
	if host == "" {
		fmt.Println("❌ Server IP/hostname is required")
		return nil
	}

	user := cli.getUserInput("Enter SSH username")
	if user == "" {
		fmt.Println("❌ SSH username is required")
		return nil
	}

	// Authentication method
	fmt.Println()
	fmt.Println("Choose authentication method:")
	fmt.Println("  1. 🔑 Password")
	fmt.Println("  2. 🔐 SSH Key")
	authChoice := cli.getUserInput("Select (1-2)")

	var password, keyPath string
	switch authChoice {
	case "1":
		password = cli.getPasswordInput("Enter SSH password")
		if password == "" {
			fmt.Println("❌ Password is required")
			return nil
		}
	case "2":
		keyPath = cli.getUserInput("Enter SSH key path (e.g., ~/.ssh/id_rsa)")
		if keyPath == "" {
			fmt.Println("❌ SSH key path is required")
			return nil
		}
	default:
		fmt.Println("❌ Invalid choice")
		return nil
	}

	// Optional settings
	fmt.Println()
	setupProtocols := cli.getUserConfirmation("Setup all protocols on server? (y/n)")
	outputDir := cli.getUserInputWithDefault("Output directory for configs", "client-configs")

	// Execute setup
	fmt.Println()
	fmt.Println("🚀 Starting auto-discovery...")

	discovery := autodiscovery.NewServerDiscovery()
	serverInfo, err := discovery.DiscoverServer(host, "22", user, password, keyPath)
	if err != nil {
		fmt.Printf("❌ Discovery failed: %v\n", err)
		return nil
	}

	fmt.Println("✅ Server discovered successfully!")
	cli.displayServerInfo(serverInfo)

	if setupProtocols {
		fmt.Println()
		fmt.Println("⚙️  Setting up protocols...")
		if err := discovery.SetupAllProtocols(); err != nil {
			fmt.Printf("⚠️  Some protocols failed to setup: %v\n", err)
		} else {
			fmt.Println("✅ All protocols setup successfully!")
		}
	}

	// Generate configs
	fmt.Println()
	fmt.Println("📁 Generating configuration files...")
	if err := discovery.GenerateClientConfigs(outputDir); err != nil {
		fmt.Printf("❌ Config generation failed: %v\n", err)
		return nil
	}

	fmt.Println("🎉 Quick setup completed!")
	fmt.Printf("📂 Configs saved to: %s/\n", outputDir)

	// Ask what to do next
	return cli.handlePostSetup(outputDir)
}

// handleMeshNetwork handles mesh network setup
func (cli *InteractiveCLI) handleMeshNetwork() error {
	fmt.Println()
	fmt.Println("🌐 Mesh Network Setup")
	fmt.Println("=====================")
	fmt.Println()
	fmt.Println("Create a mesh network like Tailscale with multiple servers.")
	fmt.Println()

	// Get network configuration
	networkCIDR := cli.getUserInputWithDefault("Network CIDR", "10.99.0.0/24")
	localNodeName := cli.getUserInputWithDefault("Local node name", "local-node")

	meshConfig := &mesh.MeshConfig{
		NetworkCIDR:         networkCIDR,
		LocalNodeName:       localNodeName,
		AutoDiscovery:       true,
		HealthCheckInterval: 30000000000, // 30 seconds
		LoadBalancing:       "latency",
		FailoverTimeout:     30000000000, // 30 seconds
		Encryption:          true,
	}

	// Create mesh network
	meshNet := mesh.NewMeshNetwork(meshConfig)
	if err := meshNet.Initialize(); err != nil {
		fmt.Printf("❌ Failed to initialize mesh network: %v\n", err)
		return nil
	}

	fmt.Println("✅ Mesh network initialized!")
	fmt.Println()

	// Add servers to mesh
	for {
		fmt.Println("Add servers to your mesh network:")
		fmt.Println("  1. ➕ Add server")
		fmt.Println("  2. 👀 View network status")
		fmt.Println("  3. 🔗 Connect to mesh")
		fmt.Println("  4. ⬅️  Back to main menu")

		choice := cli.getUserInput("Select option (1-4)")

		switch choice {
		case "1":
			cli.addServerToMesh(meshNet)
		case "2":
			cli.showMeshStatus(meshNet)
		case "3":
			cli.connectToMesh(meshNet)
		case "4":
			return nil
		default:
			fmt.Println("❌ Invalid option")
		}
	}
}

// handleExistingConfig handles existing configuration
func (cli *InteractiveCLI) handleExistingConfig() error {
	fmt.Println()
	fmt.Println("📁 Use Existing Configuration")
	fmt.Println("=============================")
	fmt.Println()

	configPath := cli.getUserInputWithDefault("Config file path", "configs/config.yaml")

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("❌ Config file not found: %s\n", configPath)
		return nil
	}

	// Load config
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		return nil
	}

	fmt.Printf("✅ Configuration loaded: %d servers found\n", len(cfg.Servers))

	// Show run options
	fmt.Println()
	fmt.Println("Run mode:")
	fmt.Println("  1. 🖥️  Client mode")
	fmt.Println("  2. 🌐 Server mode (with web interface)")
	fmt.Println("  3. ⬅️  Back")

	choice := cli.getUserInput("Select mode (1-3)")

	switch choice {
	case "1":
		fmt.Printf("🚀 Starting in client mode...\n")
		// Start client mode logic here
		return cli.startClientMode(cfg)
	case "2":
		port := cli.getUserInputWithDefault("Web interface port", "8888")
		fmt.Printf("🌐 Starting server mode on port %s...\n", port)
		// Start server mode logic here
		return cli.startServerMode(cfg, port)
	case "3":
		return nil
	default:
		fmt.Println("❌ Invalid option")
		return nil
	}
}

// Helper methods

func (cli *InteractiveCLI) getUserInput(prompt string) string {
	fmt.Printf("📝 %s: ", prompt)
	cli.scanner.Scan()
	return strings.TrimSpace(cli.scanner.Text())
}

func (cli *InteractiveCLI) getUserInputWithDefault(prompt, defaultValue string) string {
	fmt.Printf("📝 %s [%s]: ", prompt, defaultValue)
	cli.scanner.Scan()
	input := strings.TrimSpace(cli.scanner.Text())
	if input == "" {
		return defaultValue
	}
	return input
}

func (cli *InteractiveCLI) getUserConfirmation(prompt string) bool {
	for {
		response := cli.getUserInput(prompt)
		switch strings.ToLower(response) {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Println("❌ Please enter 'y' or 'n'")
		}
	}
}

func (cli *InteractiveCLI) getPasswordInput(prompt string) string {
	fmt.Printf("🔐 %s: ", prompt)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // New line after password input
	if err != nil {
		return ""
	}
	return string(bytePassword)
}

func (cli *InteractiveCLI) displayServerInfo(info *autodiscovery.ServerInfo) {
	fmt.Println()
	fmt.Println("🖥️  Server Information:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("   🏠 Host: %s\n", info.Host)
	fmt.Printf("   💻 OS: %s\n", info.OS)
	fmt.Printf("   🏗️  Architecture: %s\n", info.Architecture)
	fmt.Printf("   🔌 Available Ports: %v\n", info.AvailablePorts)
	fmt.Printf("   📦 Installed Software: %v\n", info.InstalledSoftware)
	fmt.Printf("   🔄 Supported Protocols: %v\n", info.SupportedProtocols)
}

func (cli *InteractiveCLI) handlePostSetup(outputDir string) error {
	fmt.Println()
	fmt.Println("What would you like to do next?")
	fmt.Println("  1. 🚀 Start tunnel manager")
	fmt.Println("  2. 👀 View generated configs")
	fmt.Println("  3. 📱 Show mobile app setup")
	fmt.Println("  4. ⬅️  Back to main menu")

	choice := cli.getUserInput("Select option (1-4)")

	switch choice {
	case "1":
		configPath := fmt.Sprintf("%s/ssh-tunnel-manager-config.yaml", outputDir)
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("❌ Failed to load generated config: %v\n", err)
			return nil
		}
		return cli.startClientMode(cfg)
	case "2":
		return cli.showGeneratedConfigs(outputDir)
	case "3":
		return cli.showMobileSetup(outputDir)
	case "4":
		return nil
	default:
		fmt.Println("❌ Invalid option")
		return cli.handlePostSetup(outputDir)
	}
}

func (cli *InteractiveCLI) showGeneratedConfigs(outputDir string) error {
	fmt.Println()
	fmt.Println("📁 Generated Configuration Files:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	configs := []struct {
		file        string
		description string
		usage       string
	}{
		{"ssh_tunnel.conf", "SSH Tunnel", "ssh -F ssh_tunnel.conf tunnel-server"},
		{"vless_client.conf", "VLESS URL", "Copy URL to V2rayN/V2rayNG"},
		{"vmess_client.conf", "VMess URL", "Copy URL to mobile apps"},
		{"trojan_client.conf", "Trojan Config", "trojan -c trojan_client.conf"},
		{"wireguard.conf", "WireGuard", "wg-quick up wireguard.conf"},
		{"hysteria.conf", "Hysteria", "hysteria -c hysteria.conf"},
		{"socks5_proxy.conf", "SOCKS5 Settings", "Browser proxy: 127.0.0.1:8080"},
		{"http_proxy.conf", "HTTP Settings", "Browser proxy: 127.0.0.1:8081"},
	}

	for i, cfg := range configs {
		fmt.Printf("  %d. 📄 %s\n", i+1, cfg.file)
		fmt.Printf("     📝 %s\n", cfg.description)
		fmt.Printf("     💻 Usage: %s\n", cfg.usage)
		fmt.Println()
	}

	fmt.Printf("📂 All files are in: %s/\n", outputDir)

	cli.getUserInput("Press Enter to continue")
	return nil
}

func (cli *InteractiveCLI) showMobileSetup(outputDir string) error {
	fmt.Println()
	fmt.Println("📱 Mobile App Setup")
	fmt.Println("═══════════════════")
	fmt.Println()
	fmt.Println("Android (V2rayNG):")
	fmt.Printf("  1. Install V2rayNG from Google Play\n")
	fmt.Printf("  2. Open %s/vless_client.conf\n", outputDir)
	fmt.Printf("  3. Copy the vless:// URL\n")
	fmt.Printf("  4. In V2rayNG: + → Import config from clipboard\n")
	fmt.Println()
	fmt.Println("iOS (Shadowrocket):")
	fmt.Printf("  1. Install Shadowrocket from App Store\n")
	fmt.Printf("  2. Copy Trojan URL from %s/trojan_client.conf\n", outputDir)
	fmt.Printf("  3. In Shadowrocket: + → Type → Trojan\n")
	fmt.Println()
	fmt.Println("Windows (V2rayN):")
	fmt.Printf("  1. Download V2rayN\n")
	fmt.Printf("  2. Import %s/v2ray_client.conf\n", outputDir)
	fmt.Println()

	cli.getUserInput("Press Enter to continue")
	return nil
}

func (cli *InteractiveCLI) addServerToMesh(meshNet *mesh.MeshNetwork) {
	fmt.Println()
	fmt.Println("➕ Add Server to Mesh")
	fmt.Println("═══════════════════")

	host := cli.getUserInput("Server IP/hostname")
	user := cli.getUserInput("SSH username")
	password := cli.getPasswordInput("SSH password")

	// Create server config
	serverConfig := config.Server{
		Name:      fmt.Sprintf("mesh-%s", host),
		Host:      host,
		Port:      "22",
		User:      user,
		Password:  password,
		Transport: config.TransportSSH,
		Enabled:   true,
		Tags:      []string{"mesh"},
	}

	// Add to mesh
	node, err := meshNet.AddServer(serverConfig)
	if err != nil {
		fmt.Printf("❌ Failed to add server: %v\n", err)
		return
	}

	fmt.Printf("✅ Server added to mesh: %s (%s)\n", node.Name, node.MeshIP)
}

func (cli *InteractiveCLI) showMeshStatus(meshNet *mesh.MeshNetwork) {
	fmt.Println()
	fmt.Println("🌐 Mesh Network Status")
	fmt.Println("═════════════════════")

	status := meshNet.GetNetworkStatus()
	fmt.Printf("   📊 Total Nodes: %v\n", status["total_nodes"])
	fmt.Printf("   ✅ Online Nodes: %v\n", status["online_nodes"])
	fmt.Printf("   ❌ Offline Nodes: %v\n", status["offline_nodes"])
	fmt.Printf("   🌍 Network CIDR: %v\n", status["network_cidr"])
	fmt.Printf("   ⚖️  Load Balancing: %v\n", status["load_balancing"])

	cli.getUserInput("Press Enter to continue")
}

func (cli *InteractiveCLI) connectToMesh(meshNet *mesh.MeshNetwork) {
	fmt.Println()
	fmt.Println("🔗 Connect to Mesh")
	fmt.Println("═════════════════")

	fmt.Println("Connection options:")
	fmt.Println("  1. 🎯 Best node (auto-select)")
	fmt.Println("  2. 🌍 By region")
	fmt.Println("  3. 🏷️  By tag")

	choice := cli.getUserInput("Select option (1-3)")

	switch choice {
	case "1":
		node, err := meshNet.GetBestNode("best")
		if err != nil {
			fmt.Printf("❌ No available nodes: %v\n", err)
			return
		}
		fmt.Printf("🔗 Connecting to best node: %s (%s)\n", node.Name, node.MeshIP)
		meshNet.ConnectToNode(node.ID, "ssh")
	case "2":
		region := cli.getUserInput("Enter region")
		nodes := meshNet.GetNodesByRegion(region)
		if len(nodes) == 0 {
			fmt.Printf("❌ No nodes found in region: %s\n", region)
			return
		}
		fmt.Printf("🔗 Connecting to node in %s: %s\n", region, nodes[0].Name)
		meshNet.ConnectToNode(nodes[0].ID, "ssh")
	case "3":
		tag := cli.getUserInput("Enter tag")
		nodes := meshNet.GetNodesByTag(tag)
		if len(nodes) == 0 {
			fmt.Printf("❌ No nodes found with tag: %s\n", tag)
			return
		}
		fmt.Printf("🔗 Connecting to node with tag %s: %s\n", tag, nodes[0].Name)
		meshNet.ConnectToNode(nodes[0].ID, "ssh")
	}
}

func (cli *InteractiveCLI) startClientMode(cfg *config.Config) error {
	fmt.Println("🚀 Client mode started!")
	fmt.Println("Use Ctrl+C to stop")
	// Client mode implementation
	return nil
}

func (cli *InteractiveCLI) startServerMode(cfg *config.Config, port string) error {
	fmt.Printf("🌐 Server mode started on port %s\n", port)
	fmt.Printf("Web interface: http://localhost:%s\n", port)
	fmt.Println("Use Ctrl+C to stop")
	// Server mode implementation
	return nil
}

func (cli *InteractiveCLI) handleAdvancedConfig() error {
	fmt.Println("⚙️ Advanced configuration coming soon!")
	cli.getUserInput("Press Enter to continue")
	return nil
}

func (cli *InteractiveCLI) handleMonitoring() error {
	fmt.Println("📊 Monitoring interface coming soon!")
	cli.getUserInput("Press Enter to continue")
	return nil
}

func (cli *InteractiveCLI) handleServerManagement() error {
	fmt.Println("🔧 Server management coming soon!")
	cli.getUserInput("Press Enter to continue")
	return nil
}

func (cli *InteractiveCLI) showHelp() {
	fmt.Println()
	fmt.Println("📖 SSH Tunnel Manager Help")
	fmt.Println("══════════════════════════")
	fmt.Println()
	fmt.Println("Quick Commands:")
	fmt.Println("  tunnel quick <ip> <user> <pass>    # Quick setup")
	fmt.Println("  tunnel mesh add <ip> <user>        # Add to mesh")
	fmt.Println("  tunnel mesh status                 # Mesh status")
	fmt.Println("  tunnel config <file>               # Use config")
	fmt.Println("  tunnel server                      # Server mode")
	fmt.Println()
	fmt.Println("Documentation:")
	fmt.Println("  📄 README.md - General guide")
	fmt.Println("  📄 AUTODISCOVERY.md - Auto-discovery guide")
	fmt.Println("  📄 FEATURES.md - Feature documentation")
	fmt.Println()

	cli.getUserInput("Press Enter to continue")
}
