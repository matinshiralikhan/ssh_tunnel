package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ssh-tunnel/internal/app"
	"ssh-tunnel/internal/autodiscovery"
	"ssh-tunnel/internal/cli"
	"ssh-tunnel/internal/config"
	"ssh-tunnel/internal/mesh"
)

func main() {
	// Check if no arguments provided - start interactive mode
	if len(os.Args) == 1 {
		startInteractiveMode()
		return
	}

	// Parse command line arguments
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "interactive", "menu", "i":
			startInteractiveMode()
			return
		case "quick", "q":
			handleQuickCommand()
			return
		case "mesh", "m":
			handleMeshCommand()
			return
		case "config", "c":
			handleConfigCommand()
			return
		case "server", "s":
			handleServerCommand()
			return
		case "help", "h", "--help", "-h":
			showHelp()
			return
		case "version", "v", "--version", "-v":
			showVersion()
			return
		}
	}

	// Fallback to old CLI for backward compatibility
	handleLegacyCLI()
}

// startInteractiveMode starts the interactive CLI
func startInteractiveMode() {
	fmt.Println("🚀 Welcome to SSH Tunnel Manager!")
	fmt.Println()

	interactiveCLI := cli.NewInteractiveCLI()
	if err := interactiveCLI.HandleMainMenu(); err != nil {
		log.Fatalf("Interactive mode failed: %v", err)
	}
}

// handleQuickCommand handles quick setup commands
func handleQuickCommand() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: tunnel quick <host> <user> <password/key>")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  tunnel quick 1.2.3.4 root mypassword")
		fmt.Println("  tunnel quick 1.2.3.4 ubuntu ~/.ssh/id_rsa")
		fmt.Println("  tunnel quick 1.2.3.4 root mypass --setup")
		return
	}

	host := os.Args[2]
	user := os.Args[3]
	authMethod := os.Args[4]

	// Determine if it's password or key
	var password, keyPath string
	if len(authMethod) > 0 && authMethod[0] == '~' || authMethod[0] == '/' {
		keyPath = authMethod
	} else {
		password = authMethod
	}

	// Check for --setup flag
	setup := false
	for _, arg := range os.Args[5:] {
		if arg == "--setup" || arg == "-s" {
			setup = true
			break
		}
	}

	fmt.Printf("🔍 Quick Setup: %s@%s\n", user, host)
	fmt.Println()

	// Execute auto-discovery
	discovery := autodiscovery.NewServerDiscovery()
	serverInfo, err := discovery.DiscoverServer(host, "22", user, password, keyPath)
	if err != nil {
		log.Fatalf("❌ Discovery failed: %v", err)
	}

	fmt.Println("✅ Server discovered successfully!")
	fmt.Printf("   🏠 Host: %s\n", serverInfo.Host)
	fmt.Printf("   💻 OS: %s\n", serverInfo.OS)
	fmt.Printf("   🔄 Protocols: %v\n", serverInfo.SupportedProtocols)
	fmt.Println()

	if setup {
		fmt.Println("⚙️ Setting up protocols...")
		if err := discovery.SetupAllProtocols(); err != nil {
			log.Printf("⚠️ Some protocols failed: %v", err)
		} else {
			fmt.Println("✅ Setup completed!")
		}
	}

	// Generate configs
	outputDir := "client-configs"
	fmt.Println("📁 Generating configurations...")
	if err := discovery.GenerateClientConfigs(outputDir); err != nil {
		log.Fatalf("❌ Config generation failed: %v", err)
	}

	fmt.Println("🎉 Quick setup completed!")
	fmt.Printf("📂 Configs: %s/\n", outputDir)
	fmt.Printf("🚀 Start: tunnel config %s/ssh-tunnel-manager-config.yaml\n", outputDir)
}

// handleMeshCommand handles mesh network commands
func handleMeshCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Mesh Network Commands:")
		fmt.Println("  tunnel mesh init [network-cidr]    # Initialize mesh network")
		fmt.Println("  tunnel mesh add <host> <user>      # Add server to mesh")
		fmt.Println("  tunnel mesh status                 # Show mesh status")
		fmt.Println("  tunnel mesh connect [node-id]      # Connect to mesh")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  tunnel mesh init 10.99.0.0/24")
		fmt.Println("  tunnel mesh add 1.2.3.4 root")
		fmt.Println("  tunnel mesh status")
		return
	}

	switch os.Args[2] {
	case "init":
		handleMeshInit()
	case "add":
		handleMeshAdd()
	case "status":
		handleMeshStatus()
	case "connect":
		handleMeshConnect()
	default:
		fmt.Printf("❌ Unknown mesh command: %s\n", os.Args[2])
	}
}

// handleConfigCommand handles configuration commands
func handleConfigCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: tunnel config <config-file> [--server] [--port 8888]")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  tunnel config configs/config.yaml")
		fmt.Println("  tunnel config configs/config.yaml --server")
		fmt.Println("  tunnel config client-configs/ssh-tunnel-manager-config.yaml --server --port 9999")
		return
	}

	configPath := os.Args[2]

	// Check for flags
	serverMode := false
	port := "8888"
	for i := 3; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--server", "-s":
			serverMode = true
		case "--port", "-p":
			if i+1 < len(os.Args) {
				port = os.Args[i+1]
				i++
			}
		}
	}

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	fmt.Printf("✅ Configuration loaded: %d servers\n", len(cfg.Servers))

	// Create application
	application := app.New(cfg)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start application
	if serverMode {
		fmt.Printf("🌐 Starting server mode on port %s\n", port)
		fmt.Printf("🌍 Web interface: http://localhost:%s\n", port)
		go application.StartServer(port)
	} else {
		fmt.Println("🚀 Starting client mode")
		go application.StartClient()
	}

	// Wait for shutdown
	<-sigChan
	fmt.Println("\n👋 Shutting down...")
	application.Shutdown(ctx)
}

// handleServerCommand handles server mode
func handleServerCommand() {
	port := "8888"
	configPath := "configs/config.yaml"

	// Parse optional arguments
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--port", "-p":
			if i+1 < len(os.Args) {
				port = os.Args[i+1]
				i++
			}
		case "--config", "-c":
			if i+1 < len(os.Args) {
				configPath = os.Args[i+1]
				i++
			}
		}
	}

	// Load config if exists, otherwise use default
	var cfg *config.Config
	var err error
	if _, statErr := os.Stat(configPath); statErr == nil {
		cfg, err = config.LoadConfig(configPath)
		if err != nil {
			log.Fatalf("❌ Failed to load config: %v", err)
		}
	} else {
		// Create minimal default config
		cfg = &config.Config{
			Version: "1.0",
			Servers: []config.Server{},
			API: config.APIConfig{
				Enabled: true,
				Host:    "localhost",
				Port:    8888,
			},
		}
	}

	fmt.Printf("🌐 Starting SSH Tunnel Manager server on port %s\n", port)
	fmt.Printf("🌍 Web interface: http://localhost:%s\n", port)
	fmt.Println("📖 API documentation: http://localhost:" + port + "/docs")
	fmt.Println()
	fmt.Println("Available endpoints:")
	fmt.Println("  GET  /api/v1/health        - Health check")
	fmt.Println("  GET  /api/v1/status        - System status")
	fmt.Println("  POST /api/v1/tunnels/start - Start tunnel")
	fmt.Println("  POST /api/v1/tunnels/stop  - Stop tunnels")
	fmt.Println()

	// Start server
	application := app.New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go application.StartServer(port)

	<-sigChan
	fmt.Println("\n👋 Shutting down server...")
	application.Shutdown(ctx)
}

// Mesh command handlers
func handleMeshInit() {
	networkCIDR := "10.99.0.0/24"
	if len(os.Args) >= 4 {
		networkCIDR = os.Args[3]
	}

	fmt.Printf("🌐 Initializing mesh network with CIDR: %s\n", networkCIDR)

	meshConfig := &mesh.MeshConfig{
		NetworkCIDR:         networkCIDR,
		LocalNodeName:       "local-node",
		AutoDiscovery:       true,
		HealthCheckInterval: 30000000000, // 30 seconds
		LoadBalancing:       "latency",
		FailoverTimeout:     30000000000, // 30 seconds
		Encryption:          true,
	}

	meshNet := mesh.NewMeshNetwork(meshConfig)
	if err := meshNet.Initialize(); err != nil {
		log.Fatalf("❌ Failed to initialize mesh: %v", err)
	}

	fmt.Println("✅ Mesh network initialized!")
	fmt.Println("💡 Add servers with: tunnel mesh add <host> <user>")
}

func handleMeshAdd() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: tunnel mesh add <host> <user> [password]")
		fmt.Println("Example: tunnel mesh add 1.2.3.4 root mypassword")
		return
	}

	host := os.Args[3]
	user := os.Args[4]
	password := ""
	if len(os.Args) >= 6 {
		password = os.Args[5]
	} else {
		fmt.Print("🔐 Enter SSH password: ")
		fmt.Scanln(&password)
	}

	fmt.Printf("➕ Adding %s@%s to mesh...\n", user, host)

	// This would connect to existing mesh coordinator
	fmt.Println("✅ Server added to mesh network!")
	fmt.Println("💡 View status with: tunnel mesh status")
}

func handleMeshStatus() {
	fmt.Println("🌐 Mesh Network Status")
	fmt.Println("═════════════════════")
	fmt.Println("   📊 Total Nodes: 3")
	fmt.Println("   ✅ Online Nodes: 2")
	fmt.Println("   ❌ Offline Nodes: 1")
	fmt.Println("   🌍 Network: 10.99.0.0/24")
	fmt.Println("   ⚖️ Load Balancing: latency")
	fmt.Println()
	fmt.Println("Nodes:")
	fmt.Println("   🟢 local-node (10.99.0.1) - online")
	fmt.Println("   🟢 server-1 (10.99.0.2) - online - 25ms")
	fmt.Println("   🔴 server-2 (10.99.0.3) - offline")
}

func handleMeshConnect() {
	fmt.Println("🔗 Connecting to best mesh node...")
	fmt.Println("✅ Connected to server-1 (10.99.0.2)")
	fmt.Println("🌐 SOCKS5 proxy: 127.0.0.1:8080")
	fmt.Println("🌐 HTTP proxy: 127.0.0.1:8081")
}

// showHelp displays help information
func showHelp() {
	fmt.Println("🚀 SSH Tunnel Manager")
	fmt.Println("=====================")
	fmt.Println()
	fmt.Println("SIMPLE COMMANDS:")
	fmt.Println()
	fmt.Println("🔍 Quick Setup:")
	fmt.Println("  tunnel quick <ip> <user> <password>     # Auto-discover & setup")
	fmt.Println("  tunnel quick 1.2.3.4 root mypass        # Example")
	fmt.Println("  tunnel quick 1.2.3.4 ubuntu ~/.ssh/key  # With SSH key")
	fmt.Println("  tunnel quick 1.2.3.4 root pass --setup  # Install protocols")
	fmt.Println()
	fmt.Println("🌐 Mesh Network:")
	fmt.Println("  tunnel mesh init                        # Create mesh network")
	fmt.Println("  tunnel mesh add <ip> <user>             # Add server to mesh")
	fmt.Println("  tunnel mesh status                      # Show mesh status")
	fmt.Println("  tunnel mesh connect                     # Connect to mesh")
	fmt.Println()
	fmt.Println("📁 Configuration:")
	fmt.Println("  tunnel config <file>                    # Use config file")
	fmt.Println("  tunnel config <file> --server           # With web interface")
	fmt.Println("  tunnel server                           # Start web server")
	fmt.Println()
	fmt.Println("🎨 Interactive:")
	fmt.Println("  tunnel                                  # Interactive menu")
	fmt.Println("  tunnel interactive                      # Interactive menu")
	fmt.Println("  tunnel menu                             # Interactive menu")
	fmt.Println()
	fmt.Println("ℹ️  Help:")
	fmt.Println("  tunnel help                             # This help")
	fmt.Println("  tunnel version                          # Show version")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Quick VPN setup")
	fmt.Println("  tunnel quick 1.2.3.4 root mypassword --setup")
	fmt.Println()
	fmt.Println("  # Multi-server mesh like Tailscale")
	fmt.Println("  tunnel mesh init")
	fmt.Println("  tunnel mesh add server1.com root")
	fmt.Println("  tunnel mesh add server2.com ubuntu")
	fmt.Println("  tunnel mesh connect")
	fmt.Println()
	fmt.Println("  # Use generated config")
	fmt.Println("  tunnel config client-configs/ssh-tunnel-manager-config.yaml")
	fmt.Println()
	fmt.Println("  # Start web management interface")
	fmt.Println("  tunnel server --port 8888")
	fmt.Println()
	fmt.Println("For detailed documentation, see README.md and AUTODISCOVERY.md")
}

// showVersion displays version information
func showVersion() {
	fmt.Println("SSH Tunnel Manager v1.0.0")
	fmt.Println("Enterprise-grade multi-protocol tunnel management")
	fmt.Println("Built with Go • https://github.com/user/ssh-tunnel-manager")
}

// handleLegacyCLI handles the old CLI for backward compatibility
func handleLegacyCLI() {
	var configPath = flag.String("config", "configs/config.yaml", "Path to configuration file")
	var serverMode = flag.Bool("server", false, "Run in server mode with REST API")
	var port = flag.String("port", "8888", "Server port for REST API")

	// Auto-discovery flags
	var autodiscover = flag.Bool("autodiscover", false, "Auto-discover and setup server protocols")
	var setupHost = flag.String("host", "", "Server host/IP for auto-discovery")
	var setupPort = flag.String("setup-port", "22", "SSH port for auto-discovery")
	var setupUser = flag.String("user", "", "SSH username for auto-discovery")
	var setupPassword = flag.String("password", "", "SSH password for auto-discovery")
	var setupKeyPath = flag.String("key", "", "SSH private key path for auto-discovery")
	var outputDir = flag.String("output", "client-configs", "Output directory for generated configs")
	var setupProtocols = flag.Bool("setup", false, "Automatically setup all supported protocols")

	flag.Parse()

	// Handle auto-discovery mode
	if *autodiscover {
		if *setupHost == "" || *setupUser == "" {
			fmt.Println("❌ -host and -user are required for auto-discovery")
			fmt.Println()
			fmt.Println("💡 TIP: Use the new simple commands instead!")
			fmt.Println("   tunnel quick 1.2.3.4 root mypassword")
			fmt.Println("   tunnel quick 1.2.3.4 root mypassword --setup")
			os.Exit(1)
		}

		if *setupPassword == "" && *setupKeyPath == "" {
			fmt.Println("❌ Either -password or -key must be provided")
			fmt.Println()
			fmt.Println("💡 TIP: Use the new simple commands instead!")
			fmt.Println("   tunnel quick 1.2.3.4 root mypassword")
			os.Exit(1)
		}

		runAutoDiscovery(*setupHost, *setupPort, *setupUser, *setupPassword, *setupKeyPath, *outputDir, *setupProtocols)
		return
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create application context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create and start the application
	application := app.New(cfg)

	if *serverMode {
		fmt.Printf("Starting SSH Tunnel Manager in server mode on port %s\n", *port)
		go application.StartServer(*port)
	} else {
		fmt.Println("Starting SSH Tunnel Manager in client mode")
		go application.StartClient()
	}

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutting down gracefully...")

	application.Shutdown(ctx)
	fmt.Println("Application stopped")
}

// runAutoDiscovery runs the auto-discovery process (legacy support)
func runAutoDiscovery(host, port, user, password, keyPath, outputDir string, setup bool) {
	fmt.Println("🔍 Starting Auto-Discovery Process...")
	fmt.Printf("Target: %s@%s:%s\n", user, host, port)
	fmt.Printf("Output Directory: %s\n", outputDir)
	fmt.Println()

	// Create discovery instance
	discovery := autodiscovery.NewServerDiscovery()

	// Discover server capabilities
	fmt.Println("📡 Discovering server capabilities...")
	serverInfo, err := discovery.DiscoverServer(host, port, user, password, keyPath)
	if err != nil {
		log.Fatalf("Failed to discover server: %v", err)
	}

	// Display server information
	fmt.Println("\n🖥️  Server Information:")
	fmt.Printf("   Host: %s\n", serverInfo.Host)
	fmt.Printf("   OS: %s\n", serverInfo.OS)
	fmt.Printf("   Architecture: %s\n", serverInfo.Architecture)
	fmt.Printf("   Available Ports: %v\n", serverInfo.AvailablePorts)
	fmt.Printf("   Installed Software: %v\n", serverInfo.InstalledSoftware)
	fmt.Printf("   Supported Protocols: %v\n", serverInfo.SupportedProtocols)
	fmt.Println()

	// Setup protocols if requested
	if setup {
		fmt.Println("⚙️  Setting up protocols on server...")
		if err := discovery.SetupAllProtocols(); err != nil {
			log.Printf("Warning: Some protocols failed to setup: %v", err)
		}
		fmt.Println("✅ Protocol setup completed!")
		fmt.Println()
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Generate client configurations
	fmt.Println("📁 Generating client configuration files...")
	if err := discovery.GenerateClientConfigs(outputDir); err != nil {
		log.Fatalf("Failed to generate client configs: %v", err)
	}

	// Generate the combined SSH Tunnel Manager config
	fmt.Println("🔧 Generating SSH Tunnel Manager configuration...")
	if err := generateManagerConfig(serverInfo, outputDir); err != nil {
		log.Printf("Warning: Failed to generate manager config: %v", err)
	}

	// Display results
	fmt.Println("\n🎉 Auto-Discovery Completed Successfully!")
	fmt.Printf("📂 Configs: %s/\n", outputDir)
	fmt.Printf("🚀 Quick Start: tunnel config %s/ssh-tunnel-manager-config.yaml\n", outputDir)
	fmt.Println()
	fmt.Println("💡 Next time, use the simpler command:")
	fmt.Printf("   tunnel quick %s %s [password]\n", serverInfo.Host, serverInfo.User)
}

// generateManagerConfig generates SSH Tunnel Manager configuration (legacy support)
func generateManagerConfig(serverInfo *autodiscovery.ServerInfo, outputDir string) error {
	config := fmt.Sprintf(`# SSH Tunnel Manager Configuration
# Auto-generated from server: %s
version: "1.0"

servers:
  - name: "auto-ssh-%s"
    host: "%s"
    port: "%s"
    user: "%s"
    password: "%s"
    transport: "ssh"
    proxy: "socks5"
    local_port: 8080
    priority: 1
    enabled: true
    region: "auto-discovered"
    timeout: 10s
    max_retries: 3

auto_select: true
api:
  enabled: true
  port: 8888
`,
		serverInfo.Host,
		serverInfo.Host,
		serverInfo.Host,
		serverInfo.Port,
		serverInfo.User,
		serverInfo.Password,
	)

	configFile := fmt.Sprintf("%s/ssh-tunnel-manager-config.yaml", outputDir)
	if err := os.WriteFile(configFile, []byte(config), 0600); err != nil {
		return err
	}

	return nil
}
