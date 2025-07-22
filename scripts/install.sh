#!/bin/bash

# SSH Tunnel Manager Installation Script

set -e

# Configuration
APP_NAME="ssh-tunnel-manager"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/${APP_NAME}"
SERVICE_DIR="/etc/systemd/system"
LOG_DIR="/var/log/${APP_NAME}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

# Detect OS
detect_os() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS=$NAME
        VER=$VERSION_ID
    elif type lsb_release >/dev/null 2>&1; then
        OS=$(lsb_release -si)
        VER=$(lsb_release -sr)
    elif [[ -f /etc/redhat-release ]]; then
        OS=RHEL
        VER=$(grep -oE '[0-9]+\.[0-9]+' /etc/redhat-release)
    else
        OS=$(uname -s)
        VER=$(uname -r)
    fi
    
    print_info "Detected OS: $OS $VER"
}

# Install dependencies
install_dependencies() {
    print_info "Installing system dependencies..."
    
    if [[ "$OS" == *"Ubuntu"* ]] || [[ "$OS" == *"Debian"* ]]; then
        apt-get update
        apt-get install -y curl wget openssl systemd
    elif [[ "$OS" == *"CentOS"* ]] || [[ "$OS" == *"RHEL"* ]] || [[ "$OS" == *"Fedora"* ]]; then
        yum update -y
        yum install -y curl wget openssl systemd
    elif [[ "$OS" == *"Alpine"* ]]; then
        apk update
        apk add curl wget openssl openrc
    else
        print_warning "Unsupported OS: $OS. Please install curl, wget, and openssl manually."
    fi
}

# Create necessary directories
create_directories() {
    print_info "Creating directories..."
    
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$LOG_DIR"
    mkdir -p "/etc/${APP_NAME}/certs"
    
    print_success "Directories created"
}

# Install binary
install_binary() {
    print_info "Installing binary..."
    
    if [[ ! -f "bin/${APP_NAME}" ]]; then
        print_error "Binary not found. Please build the application first with 'make build'"
        exit 1
    fi
    
    cp "bin/${APP_NAME}" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/${APP_NAME}"
    
    print_success "Binary installed to $INSTALL_DIR/${APP_NAME}"
}

# Install configuration
install_config() {
    print_info "Installing configuration files..."
    
    # Copy configuration files
    cp configs/config.yaml "$CONFIG_DIR/config.yaml.example"
    cp configs/config-minimal.yaml "$CONFIG_DIR/config-minimal.yaml.example"
    
    # Create default config if it doesn't exist
    if [[ ! -f "$CONFIG_DIR/config.yaml" ]]; then
        cp configs/config-minimal.yaml "$CONFIG_DIR/config.yaml"
        print_info "Default configuration installed"
    else
        print_warning "Configuration file already exists, skipping..."
    fi
    
    # Set proper permissions
    chown -R root:root "$CONFIG_DIR"
    chmod 600 "$CONFIG_DIR"/*.yaml*
    
    print_success "Configuration files installed"
}

# Install systemd service
install_service() {
    print_info "Installing systemd service..."
    
    # Create service file
    cat > "$SERVICE_DIR/${APP_NAME}.service" << EOF
[Unit]
Description=SSH Tunnel Manager
Documentation=https://github.com/user/ssh-tunnel-manager
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
Group=root
ExecStart=$INSTALL_DIR/$APP_NAME -config $CONFIG_DIR/config.yaml
ExecReload=/bin/kill -HUP \$MAINPID
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=$APP_NAME

# Security settings
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=$LOG_DIR $CONFIG_DIR
PrivateTmp=yes

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable "$APP_NAME"
    
    print_success "Systemd service installed and enabled"
}

# Generate certificates
generate_certificates() {
    print_info "Generating TLS certificates..."
    
    CERT_DIR="/etc/${APP_NAME}/certs"
    
    if [[ ! -f "$CERT_DIR/server.crt" ]]; then
        openssl req -new -newkey rsa:4096 -days 365 -nodes -x509 \
            -subj "/C=US/ST=State/L=City/O=SSH-Tunnel-Manager/CN=localhost" \
            -keyout "$CERT_DIR/server.key" \
            -out "$CERT_DIR/server.crt"
        
        chmod 600 "$CERT_DIR/server.key"
        chmod 644 "$CERT_DIR/server.crt"
        
        print_success "TLS certificates generated"
    else
        print_warning "TLS certificates already exist, skipping..."
    fi
}

# Create user for service (optional)
create_user() {
    if ! id "$APP_NAME" >/dev/null 2>&1; then
        print_info "Creating service user..."
        useradd -r -s /bin/false -d /var/lib/$APP_NAME -M $APP_NAME
        mkdir -p /var/lib/$APP_NAME
        chown $APP_NAME:$APP_NAME /var/lib/$APP_NAME
        print_success "Service user created"
    fi
}

# Post-installation tasks
post_install() {
    print_info "Running post-installation tasks..."
    
    # Set up log rotation
    cat > "/etc/logrotate.d/${APP_NAME}" << EOF
$LOG_DIR/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 root root
    postrotate
        systemctl reload $APP_NAME >/dev/null 2>&1 || true
    endscript
}
EOF
    
    print_success "Log rotation configured"
}

# Show post-installation information
show_info() {
    echo ""
    print_success "Installation completed successfully!"
    echo ""
    echo "Configuration:"
    echo "  - Config file: $CONFIG_DIR/config.yaml"
    echo "  - Log directory: $LOG_DIR"
    echo "  - Binary location: $INSTALL_DIR/$APP_NAME"
    echo ""
    echo "Service management:"
    echo "  - Start: sudo systemctl start $APP_NAME"
    echo "  - Stop: sudo systemctl stop $APP_NAME"
    echo "  - Status: sudo systemctl status $APP_NAME"
    echo "  - Logs: sudo journalctl -u $APP_NAME -f"
    echo ""
    echo "Configuration:"
    echo "  - Edit: sudo nano $CONFIG_DIR/config.yaml"
    echo "  - Test: sudo $INSTALL_DIR/$APP_NAME -config $CONFIG_DIR/config.yaml"
    echo ""
    echo "Web interface (if enabled):"
    echo "  - URL: http://localhost:8888"
    echo "  - Health check: curl http://localhost:8888/api/v1/health"
    echo ""
    print_warning "Remember to configure your servers in $CONFIG_DIR/config.yaml before starting!"
}

# Uninstall function
uninstall() {
    print_info "Uninstalling SSH Tunnel Manager..."
    
    # Stop and disable service
    systemctl stop "$APP_NAME" 2>/dev/null || true
    systemctl disable "$APP_NAME" 2>/dev/null || true
    
    # Remove files
    rm -f "$INSTALL_DIR/$APP_NAME"
    rm -f "$SERVICE_DIR/${APP_NAME}.service"
    rm -f "/etc/logrotate.d/${APP_NAME}"
    rm -rf "$CONFIG_DIR"
    rm -rf "$LOG_DIR"
    
    # Remove user
    userdel "$APP_NAME" 2>/dev/null || true
    rm -rf "/var/lib/$APP_NAME"
    
    systemctl daemon-reload
    
    print_success "SSH Tunnel Manager uninstalled"
}

# Main installation function
install() {
    print_info "Starting SSH Tunnel Manager installation..."
    
    detect_os
    check_root
    install_dependencies
    create_directories
    install_binary
    install_config
    install_service
    generate_certificates
    post_install
    show_info
}

# Script options
case "${1:-}" in
    install)
        install
        ;;
    uninstall)
        uninstall
        ;;
    *)
        echo "Usage: $0 {install|uninstall}"
        echo ""
        echo "SSH Tunnel Manager Installation Script"
        echo ""
        echo "Commands:"
        echo "  install   - Install SSH Tunnel Manager"
        echo "  uninstall - Remove SSH Tunnel Manager"
        echo ""
        exit 1
        ;;
esac 