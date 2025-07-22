package mesh

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"ssh-tunnel/internal/config"
)

// MeshNode represents a node in the mesh network
type MeshNode struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	PublicIP     string          `json:"public_ip"`
	PrivateIP    string          `json:"private_ip"`
	MeshIP       string          `json:"mesh_ip"`
	Port         int             `json:"port"`
	PublicKey    string          `json:"public_key"`
	PrivateKey   string          `json:"private_key"`
	Status       string          `json:"status"` // online, offline, connecting
	LastSeen     time.Time       `json:"last_seen"`
	Protocols    []string        `json:"protocols"`
	LoadScore    float64         `json:"load_score"`
	Latency      time.Duration   `json:"latency"`
	Tags         []string        `json:"tags"`
	Region       string          `json:"region"`
	Capabilities map[string]bool `json:"capabilities"`
}

// MeshNetwork manages the entire mesh network
type MeshNetwork struct {
	nodes           map[string]*MeshNode
	coordinatorNode *MeshNode
	localNode       *MeshNode
	routes          map[string]*Route
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	config          *MeshConfig
}

// MeshConfig holds mesh network configuration
type MeshConfig struct {
	NetworkCIDR         string        `yaml:"network_cidr" json:"network_cidr"`
	CoordinatorURL      string        `yaml:"coordinator_url" json:"coordinator_url"`
	LocalNodeName       string        `yaml:"local_node_name" json:"local_node_name"`
	AutoDiscovery       bool          `yaml:"auto_discovery" json:"auto_discovery"`
	HealthCheckInterval time.Duration `yaml:"health_check_interval" json:"health_check_interval"`
	LoadBalancing       string        `yaml:"load_balancing" json:"load_balancing"` // round_robin, least_connections, latency
	FailoverTimeout     time.Duration `yaml:"failover_timeout" json:"failover_timeout"`
	Encryption          bool          `yaml:"encryption" json:"encryption"`
	Tags                []string      `yaml:"tags" json:"tags"`
	Regions             []string      `yaml:"regions" json:"regions"`
}

// Route represents a route in the mesh network
type Route struct {
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Interface   string `json:"interface"`
	Metric      int    `json:"metric"`
	Protocol    string `json:"protocol"`
}

// NewMeshNetwork creates a new mesh network manager
func NewMeshNetwork(cfg *MeshConfig) *MeshNetwork {
	ctx, cancel := context.WithCancel(context.Background())

	return &MeshNetwork{
		nodes:  make(map[string]*MeshNode),
		routes: make(map[string]*Route),
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Initialize initializes the mesh network
func (mn *MeshNetwork) Initialize() error {
	log.Println("🌐 Initializing Mesh Network...")

	// Create local node
	localNode, err := mn.createLocalNode()
	if err != nil {
		return fmt.Errorf("failed to create local node: %v", err)
	}
	mn.localNode = localNode
	mn.nodes[localNode.ID] = localNode

	// Start services
	go mn.startHealthChecker()
	go mn.startLoadBalancer()
	go mn.startRouteManager()

	if mn.config.AutoDiscovery {
		go mn.startAutoDiscovery()
	}

	log.Printf("✅ Mesh Network initialized. Local node: %s (%s)", localNode.Name, localNode.MeshIP)
	return nil
}

// AddServer adds a server to the mesh network
func (mn *MeshNetwork) AddServer(serverConfig config.Server) (*MeshNode, error) {
	mn.mu.Lock()
	defer mn.mu.Unlock()

	// Create mesh node from server config
	node := &MeshNode{
		ID:           generateNodeID(),
		Name:         serverConfig.Name,
		PublicIP:     serverConfig.Host,
		Port:         parsePort(serverConfig.Port),
		Status:       "connecting",
		Protocols:    []string{string(serverConfig.Transport)},
		Tags:         serverConfig.Tags,
		Region:       serverConfig.Region,
		Capabilities: make(map[string]bool),
	}

	// Assign mesh IP
	meshIP, err := mn.assignMeshIP()
	if err != nil {
		return nil, fmt.Errorf("failed to assign mesh IP: %v", err)
	}
	node.MeshIP = meshIP

	// Test connection and get node info
	if err := mn.probeNode(node); err != nil {
		log.Printf("Warning: Failed to probe node %s: %v", node.Name, err)
		node.Status = "offline"
	} else {
		node.Status = "online"
		node.LastSeen = time.Now()
	}

	// Add to network
	mn.nodes[node.ID] = node

	// Setup routing
	if err := mn.setupNodeRouting(node); err != nil {
		log.Printf("Warning: Failed to setup routing for node %s: %v", node.Name, err)
	}

	log.Printf("✅ Added node to mesh: %s (%s) - %s", node.Name, node.MeshIP, node.Status)
	return node, nil
}

// GetBestNode returns the best node for a given criteria
func (mn *MeshNetwork) GetBestNode(criteria string) (*MeshNode, error) {
	mn.mu.RLock()
	defer mn.mu.RUnlock()

	var bestNode *MeshNode
	var bestScore float64

	for _, node := range mn.nodes {
		if node.Status != "online" || node == mn.localNode {
			continue
		}

		score := mn.calculateNodeScore(node, criteria)
		if bestNode == nil || score > bestScore {
			bestNode = node
			bestScore = score
		}
	}

	if bestNode == nil {
		return nil, fmt.Errorf("no available nodes found")
	}

	return bestNode, nil
}

// GetNodesByRegion returns nodes in a specific region
func (mn *MeshNetwork) GetNodesByRegion(region string) []*MeshNode {
	mn.mu.RLock()
	defer mn.mu.RUnlock()

	var nodes []*MeshNode
	for _, node := range mn.nodes {
		if node.Region == region && node.Status == "online" {
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// GetNodesByTag returns nodes with a specific tag
func (mn *MeshNetwork) GetNodesByTag(tag string) []*MeshNode {
	mn.mu.RLock()
	defer mn.mu.RUnlock()

	var nodes []*MeshNode
	for _, node := range mn.nodes {
		if node.Status == "online" && containsString(node.Tags, tag) {
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// ConnectToNode establishes connection to a specific node
func (mn *MeshNetwork) ConnectToNode(nodeID string, protocol string) error {
	mn.mu.RLock()
	node, exists := mn.nodes[nodeID]
	mn.mu.RUnlock()

	if !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	if node.Status != "online" {
		return fmt.Errorf("node %s is not online", node.Name)
	}

	// Establish connection based on protocol
	switch protocol {
	case "wireguard":
		return mn.connectViaWireGuard(node)
	case "ssh":
		return mn.connectViaSSH(node)
	case "v2ray":
		return mn.connectViaV2Ray(node)
	default:
		return mn.connectViaBestProtocol(node)
	}
}

// LoadBalance distributes traffic across multiple nodes
func (mn *MeshNetwork) LoadBalance(target string) (*MeshNode, error) {
	nodes := mn.getHealthyNodes()
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no healthy nodes available")
	}

	switch mn.config.LoadBalancing {
	case "round_robin":
		return mn.roundRobinSelect(nodes), nil
	case "least_connections":
		return mn.leastConnectionsSelect(nodes), nil
	case "latency":
		return mn.latencyBasedSelect(nodes), nil
	default:
		return mn.latencyBasedSelect(nodes), nil
	}
}

// GetNetworkStatus returns the current network status
func (mn *MeshNetwork) GetNetworkStatus() map[string]interface{} {
	mn.mu.RLock()
	defer mn.mu.RUnlock()

	totalNodes := len(mn.nodes)
	onlineNodes := 0
	offlineNodes := 0

	for _, node := range mn.nodes {
		if node.Status == "online" {
			onlineNodes++
		} else {
			offlineNodes++
		}
	}

	return map[string]interface{}{
		"total_nodes":      totalNodes,
		"online_nodes":     onlineNodes,
		"offline_nodes":    offlineNodes,
		"local_node":       mn.localNode,
		"coordinator_node": mn.coordinatorNode,
		"network_cidr":     mn.config.NetworkCIDR,
		"load_balancing":   mn.config.LoadBalancing,
		"auto_discovery":   mn.config.AutoDiscovery,
	}
}

// Private methods

func (mn *MeshNetwork) createLocalNode() (*MeshNode, error) {
	nodeID := generateNodeID()

	// Get local IP
	localIP, err := getLocalIP()
	if err != nil {
		return nil, err
	}

	// Assign mesh IP
	meshIP, err := mn.assignMeshIP()
	if err != nil {
		return nil, err
	}

	node := &MeshNode{
		ID:        nodeID,
		Name:      mn.config.LocalNodeName,
		PublicIP:  localIP,
		MeshIP:    meshIP,
		Status:    "online",
		LastSeen:  time.Now(),
		Protocols: []string{"ssh", "wireguard"},
		Tags:      mn.config.Tags,
		Capabilities: map[string]bool{
			"coordinator":  true,
			"routing":      true,
			"loadbalancer": true,
		},
	}

	// Generate WireGuard keys
	privateKey, publicKey, err := generateWireGuardKeys()
	if err != nil {
		return nil, err
	}

	node.PrivateKey = privateKey
	node.PublicKey = publicKey

	return node, nil
}

func (mn *MeshNetwork) startHealthChecker() {
	ticker := time.NewTicker(mn.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-mn.ctx.Done():
			return
		case <-ticker.C:
			mn.performHealthCheck()
		}
	}
}

func (mn *MeshNetwork) performHealthCheck() {
	mn.mu.Lock()
	defer mn.mu.Unlock()

	for _, node := range mn.nodes {
		if node == mn.localNode {
			continue
		}

		// Check node health
		latency, err := mn.pingNode(node)
		if err != nil {
			if node.Status == "online" {
				log.Printf("⚠️  Node %s went offline: %v", node.Name, err)
				node.Status = "offline"
			}
		} else {
			if node.Status != "online" {
				log.Printf("✅ Node %s is back online", node.Name)
				node.Status = "online"
			}
			node.LastSeen = time.Now()
			node.Latency = latency
		}
	}
}

func (mn *MeshNetwork) startLoadBalancer() {
	// Load balancer logic
	for {
		select {
		case <-mn.ctx.Done():
			return
		default:
			// Update load scores
			mn.updateLoadScores()
			time.Sleep(30 * time.Second)
		}
	}
}

func (mn *MeshNetwork) startRouteManager() {
	// Route management logic
	for {
		select {
		case <-mn.ctx.Done():
			return
		default:
			// Update routes
			mn.updateRoutes()
			time.Sleep(60 * time.Second)
		}
	}
}

func (mn *MeshNetwork) startAutoDiscovery() {
	// Auto-discovery logic
	for {
		select {
		case <-mn.ctx.Done():
			return
		default:
			// Discover new nodes
			mn.discoverNewNodes()
			time.Sleep(5 * time.Minute)
		}
	}
}

// Helper methods
func (mn *MeshNetwork) assignMeshIP() (string, error) {
	// Parse network CIDR
	_, network, err := net.ParseCIDR(mn.config.NetworkCIDR)
	if err != nil {
		return "", err
	}

	// Find next available IP
	ip := network.IP
	for {
		ip = nextIP(ip)
		if !mn.isIPUsed(ip.String()) {
			return ip.String(), nil
		}
	}
}

func (mn *MeshNetwork) isIPUsed(ip string) bool {
	for _, node := range mn.nodes {
		if node.MeshIP == ip {
			return true
		}
	}
	return false
}

func (mn *MeshNetwork) probeNode(node *MeshNode) error {
	// Test SSH connectivity
	// Test other protocols
	// Get node capabilities
	return nil // Simplified
}

func (mn *MeshNetwork) setupNodeRouting(node *MeshNode) error {
	// Setup routing rules for the node
	return nil // Simplified
}

func (mn *MeshNetwork) calculateNodeScore(node *MeshNode, criteria string) float64 {
	score := 0.0

	// Latency score (lower is better)
	latencyScore := 1000.0 / float64(node.Latency.Milliseconds()+1)
	score += latencyScore * 0.4

	// Load score (lower is better)
	loadScore := 1.0 - node.LoadScore
	score += loadScore * 0.3

	// Status bonus
	if node.Status == "online" {
		score += 100.0
	}

	// Region preference
	if criteria != "" && node.Region == criteria {
		score += 50.0
	}

	return score
}

func (mn *MeshNetwork) getHealthyNodes() []*MeshNode {
	var nodes []*MeshNode
	for _, node := range mn.nodes {
		if node.Status == "online" && node != mn.localNode {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (mn *MeshNetwork) roundRobinSelect(nodes []*MeshNode) *MeshNode {
	// Simple round-robin implementation
	if len(nodes) == 0 {
		return nil
	}
	return nodes[time.Now().Second()%len(nodes)]
}

func (mn *MeshNetwork) leastConnectionsSelect(nodes []*MeshNode) *MeshNode {
	// Select node with least load
	var bestNode *MeshNode
	var bestLoad float64 = 1.0

	for _, node := range nodes {
		if node.LoadScore < bestLoad {
			bestLoad = node.LoadScore
			bestNode = node
		}
	}

	return bestNode
}

func (mn *MeshNetwork) latencyBasedSelect(nodes []*MeshNode) *MeshNode {
	// Select node with best latency
	var bestNode *MeshNode
	var bestLatency time.Duration = time.Hour

	for _, node := range nodes {
		if node.Latency < bestLatency {
			bestLatency = node.Latency
			bestNode = node
		}
	}

	return bestNode
}

func (mn *MeshNetwork) connectViaWireGuard(node *MeshNode) error {
	// WireGuard connection logic
	log.Printf("🔗 Connecting to %s via WireGuard", node.Name)
	return nil
}

func (mn *MeshNetwork) connectViaSSH(node *MeshNode) error {
	// SSH connection logic
	log.Printf("🔗 Connecting to %s via SSH", node.Name)
	return nil
}

func (mn *MeshNetwork) connectViaV2Ray(node *MeshNode) error {
	// V2Ray connection logic
	log.Printf("🔗 Connecting to %s via V2Ray", node.Name)
	return nil
}

func (mn *MeshNetwork) connectViaBestProtocol(node *MeshNode) error {
	// Auto-select best protocol
	if containsString(node.Protocols, "wireguard") {
		return mn.connectViaWireGuard(node)
	} else if containsString(node.Protocols, "ssh") {
		return mn.connectViaSSH(node)
	} else if containsString(node.Protocols, "v2ray") {
		return mn.connectViaV2Ray(node)
	}
	return fmt.Errorf("no suitable protocol found")
}

func (mn *MeshNetwork) pingNode(node *MeshNode) (time.Duration, error) {
	// Ping implementation
	start := time.Now()
	// ... ping logic ...
	return time.Since(start), nil
}

func (mn *MeshNetwork) updateLoadScores() {
	// Update load scores for all nodes
}

func (mn *MeshNetwork) updateRoutes() {
	// Update routing table
}

func (mn *MeshNetwork) discoverNewNodes() {
	// Auto-discovery logic
}

// Utility functions
func generateNodeID() string {
	return fmt.Sprintf("node-%d", time.Now().UnixNano())
}

func parsePort(portStr string) int {
	// Parse port string to int
	return 22 // Simplified
}

func getLocalIP() (string, error) {
	// Get local IP address
	return "127.0.0.1", nil // Simplified
}

func generateWireGuardKeys() (privateKey, publicKey string, err error) {
	// Generate WireGuard key pair
	return "private-key", "public-key", nil // Simplified
}

func nextIP(ip net.IP) net.IP {
	// Get next IP address
	next := make(net.IP, len(ip))
	copy(next, ip)
	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] > 0 {
			break
		}
	}
	return next
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
