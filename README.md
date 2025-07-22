# SSH Tunnel Manager

A powerful, secure, and feature-rich SSH tunnel manager supporting multiple protocols and advanced bypassing capabilities.

## 🚀 Features

### ⚡ **NEW: Auto-Discovery & One-Click Setup**
- **🔍 Server Auto-Discovery**: Automatically discover server capabilities with just IP, username, and password
- **🛠️ One-Click Protocol Setup**: Automatically install and configure multiple protocols (SSH, V2Ray, VLESS, VMess, Trojan, Hysteria, WireGuard, etc.)
- **📁 Multi-Client Config Generation**: Generate configuration files for all popular clients (V2rayN, Trojan, WireGuard, etc.)
- **🔧 Combined Configuration**: Single SSH Tunnel Manager config file with all discovered protocols

### Core Features
- **Multi-Protocol Support**: SSH, Hysteria, V2Ray, WireGuard, Trojan, VLESS, VMess
- **Auto Server Selection**: Latency-based, load-balanced, or random selection
- **Advanced Security**: TLS encryption, token authentication, config encryption
- **Failover & Load Balancing**: Automatic failover with health monitoring
- **REST API**: Complete management API with Echo framework
- **Real-time Monitoring**: System metrics, tunnel status, performance tracking
- **Smart Routing**: Rule-based traffic routing with GeoIP support

### Security & Bypassing
- **Fake TLS**: Reality protocol support for advanced bypassing
- **Multiple Obfuscation**: Supports various obfuscation methods
- **Encrypted Configuration**: AES-256 encrypted config files
- **Token-based Authentication**: Secure API access control
- **Certificate Management**: Automatic TLS certificate handling

## 🚀 Quick Start with Auto-Discovery

### One-Command Server Setup
Transform any SSH server into a multi-protocol tunnel server with a single command:

```bash
# Auto-discover and setup all protocols
./ssh-tunnel-manager -autodiscover -host YOUR_SERVER_IP -user root -password YOUR_PASSWORD -setup
```

### Real-world Examples

#### Basic Discovery (No Changes to Server)
```bash
./ssh-tunnel-manager -autodiscover -host 1.2.3.4 -user root -password mypassword
```

#### Full Setup with SSH Key
```bash
./ssh-tunnel-manager -autodiscover -host my-server.com -user ubuntu -key ~/.ssh/my-key.pem -setup -output my-configs
```

#### Production Setup
```bash
./ssh-tunnel-manager -autodiscover \
  -host production-server.com \
  -user root \
  -password secure-password \
  -setup \
  -output production-configs
```

### What You Get
After running auto-discovery, you'll receive:

```
📋 Generated Files:
   📂 client-configs/
   📄 Configuration Files:
      • ssh_tunnel.conf           # SSH tunnel config
      • v2ray_client.conf         # V2Ray JSON config
      • vless_client.conf         # VLESS URL for V2rayN
      • vmess_client.conf         # VMess URL for mobile apps
      • trojan_client.conf        # Trojan client config
      • wireguard.conf           # WireGuard config
      • hysteria.conf            # Hysteria client config
      • http_proxy.conf          # HTTP proxy settings
      • socks5_proxy.conf        # SOCKS5 proxy settings
      • ssh-tunnel-manager-config.yaml  # Ready-to-use config
```

## 📦 Installation

### Quick Start
```bash
# Clone the repository
git clone <your-repo-url>
cd ssh_tunnel

# Build the application
go build -o bin/ssh-tunnel-manager ./cmd/main.go

# Auto-discover and setup your server
./bin/ssh-tunnel-manager -autodiscover -host YOUR_IP -user root -password YOUR_PASS -setup
```

### Advanced Installation
```bash
# Build for multiple platforms
make build-all

# Install systemd service
sudo make install-service

# Generate TLS certificates
make generate-certs
```

## ⚙️ Usage Modes

### 1. Auto-Discovery Mode (Recommended)
Automatically discover and setup protocols:

```bash
# Discover server capabilities
./ssh-tunnel-manager -autodiscover -host 1.2.3.4 -user root -password mypass

# Setup all protocols automatically  
./ssh-tunnel-manager -autodiscover -host 1.2.3.4 -user root -password mypass -setup

# Use SSH key authentication
./ssh-tunnel-manager -autodiscover -host 1.2.3.4 -user ubuntu -key ~/.ssh/id_rsa -setup
```

### 2. Traditional Configuration Mode
Use pre-configured YAML files:

```bash
# Start in client mode
./ssh-tunnel-manager -config configs/config.yaml

# Start in server mode with REST API
./ssh-tunnel-manager -config configs/config.yaml -server -port 8888
```

### 3. Server Management Mode
Run with web interface for management:

```bash
# Start with generated config in server mode
./ssh-tunnel-manager -config client-configs/ssh-tunnel-manager-config.yaml -server
```

## 🔧 Protocol Support

### Automatically Detected & Configured:

| Protocol | Description | Use Case |
|----------|-------------|----------|
| **SSH Tunnel** | Classic SSH SOCKS5/HTTP proxy | Universal compatibility |
| **V2Ray/VMess** | Modern protocol with WebSocket | CDN compatibility |
| **VLESS** | Lightweight V2Ray variant | Better performance |
| **Trojan** | TLS-camouflaged protocol | Deep packet inspection bypass |
| **Hysteria** | UDP-based high-speed protocol | High-bandwidth scenarios |
| **WireGuard** | Modern VPN protocol | Full device VPN |
| **HTTP Proxy** | Standard HTTP proxy | Web browsing |
| **SOCKS5 Proxy** | SOCKS5 with DNS tunneling | Application proxy |
| **ICMP Tunnel** | ICMP-based tunnel | Firewall bypass |

## 🎯 Use Case Examples

### Personal VPN Server
```bash
# Setup your VPS as a personal VPN
./ssh-tunnel-manager -autodiscover -host your-vps.com -user root -password pass -setup

# Use the generated WireGuard config on your devices
sudo wg-quick up client-configs/wireguard.conf
```

### Corporate Remote Access
```bash
# Discovery existing corporate server
./ssh-tunnel-manager -autodiscover -host corp-server.com -user admin -key ~/.ssh/corp-key

# Use SSH tunnel for secure access
ssh -F client-configs/ssh_tunnel.conf tunnel-server
```

### Content Access & Streaming
```bash
# Setup server in different region
./ssh-tunnel-manager -autodiscover -host eu-server.com -user root -password pass -setup

# Use SOCKS5 proxy in browser: 127.0.0.1:8080
```

### Development & Testing
```bash
# Quick tunnel for development
./ssh-tunnel-manager -autodiscover -host dev-server.com -user developer -key ~/.ssh/dev-key

# Access internal services through generated proxy
```

## 📱 Client Configuration

### V2rayN/V2rayNG (Mobile)
1. Copy VLESS URL from `client-configs/vless_client.conf`
2. Import in V2rayN/V2rayNG
3. Connect and browse

### Trojan Clients
```bash
trojan -c client-configs/trojan_client.conf
```

### WireGuard
```bash
sudo wg-quick up client-configs/wireguard.conf
```

### Browser Proxy Settings
- **SOCKS5**: 127.0.0.1:8080
- **HTTP**: 127.0.0.1:8081

## 🔒 Security Features

### Configuration Encryption
```bash
# Encrypt your configs
export CONFIG_PASSWORD="your-secure-password"
./ssh-tunnel-manager -config encrypted-config.yaml
```

### API Authentication
All auto-generated configs include secure authentication:
- Random auth tokens
- TLS support
- Rate limiting

### Best Practices
```bash
# Use SSH keys instead of passwords
./ssh-tunnel-manager -autodiscover -host server.com -user user -key ~/.ssh/secure-key

# Limit config file permissions  
chmod 600 client-configs/*

# Use encrypted configurations for sensitive data
```

## 📊 Monitoring & Management

### REST API
The generated configs include a management API:

```bash
# Health check
curl http://localhost:8888/api/v1/health

# Get tunnel status
curl -H "Authorization: Bearer your-token" http://localhost:8888/api/v1/status

# Start/stop tunnels
curl -X POST -H "Authorization: Bearer token" http://localhost:8888/api/v1/tunnels/start
```

### Web Interface
Access the management interface at: `http://localhost:8888`

## 🛠️ Advanced Configuration

### Manual Configuration
For advanced users, create custom `config.yaml`:

```yaml
version: "1.0"

servers:
  - name: "my-server"
    host: "your-server.com"
    port: "22"
    user: "root"
    password: "your-password"
    transport: "ssh"
    proxy: "socks5"
    local_port: 8080
    enabled: true

auto_select: true
monitoring:
  enabled: true
api:
  enabled: true
  port: 8888
```

### Protocol-Specific Configuration

#### Hysteria
```yaml
servers:
  - name: "hysteria-server"
    transport: "hysteria"
    hysteria:
      protocol: "udp"
      auth_string: "your-password"
      bandwidth: "100mbps"
      obfs: "salamander"
```

#### V2Ray/VLESS
```yaml
servers:
  - name: "v2ray-server"
    transport: "v2ray"
    v2ray:
      uuid: "your-uuid"
      network: "ws"
      path: "/v2ray"
      tls: "tls"
```

## 🔄 Migration & Backup

### Backup Configurations
```bash
tar -czf tunnel-backup.tar.gz client-configs/
```

### Update Configurations
```bash
# Re-run discovery to update configs
./ssh-tunnel-manager -autodiscover -host server.com -user root -password newpass -output updated-configs
```

## 🚀 Performance & Optimization

### Performance Benchmarks
- **SSH Tunnels**: 100-500 Mbps
- **WireGuard**: 1+ Gbps
- **Hysteria**: 200-800 Mbps
- **V2Ray**: 100-400 Mbps

### Optimization Tips
```bash
# TCP optimization
echo 'net.core.rmem_max = 16777216' >> /etc/sysctl.conf
sysctl -p

# Monitor performance
netstat -i
iftop -i tun0
```

## 📝 Troubleshooting

### Connection Issues
```bash
# Test SSH connectivity
ssh -D 8080 root@your-server

# Check firewall
ufw status

# Monitor logs
journalctl -f
```

### Common Solutions
1. **Connection refused**: Check server firewall settings
2. **Authentication failed**: Verify SSH credentials  
3. **Port conflicts**: Use different local ports
4. **DNS issues**: Configure DNS servers properly

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new protocols
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## ⚠️ Disclaimer

This software is for educational and legitimate use only. Users are responsible for complying with all applicable laws and regulations. The developers do not condone or support any illegal activities.

## 🆘 Support & Documentation

- 📖 **Auto-Discovery Guide**: [AUTODISCOVERY.md](AUTODISCOVERY.md)
- 📋 **Feature Documentation**: [FEATURES.md](FEATURES.md)
- 🐛 **Issues**: [GitHub Issues](https://github.com/user/repo/issues)
- 💬 **Community**: [Join our Discord](https://discord.gg/example)

---

## 🎉 Quick Start Summary

1. **Download/Build**: `go build -o ssh-tunnel-manager ./cmd/main.go`
2. **Auto-Setup**: `./ssh-tunnel-manager -autodiscover -host YOUR_IP -user root -password PASS -setup`
3. **Connect**: Use generated configs in `client-configs/` directory
4. **Manage**: Access web interface at `http://localhost:8888`

Transform any server into a multi-protocol tunnel hub in under 5 minutes! 🚀 