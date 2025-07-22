# SSH Tunnel Manager Makefile

# Variables
APP_NAME = ssh-tunnel-manager
VERSION = 1.0.0
BUILD_DIR = bin
CMD_DIR = cmd
MAIN_FILE = $(CMD_DIR)/main.go

# Go build flags
LDFLAGS = -ldflags "-s -w -X main.version=$(VERSION)"
BUILD_FLAGS = $(LDFLAGS)

# Platforms for cross-compilation
PLATFORMS = \
	linux/amd64 \
	linux/arm64 \
	windows/amd64 \
	darwin/amd64 \
	darwin/arm64

.PHONY: all build clean test lint fmt vet deps build-all install-service generate-certs help

# Default target
all: clean deps test build

# Build for current platform
build:
	@echo "Building $(APP_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
	go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./$(MAIN_FILE)
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

# Build for all platforms
build-all: clean deps
	@echo "Building $(APP_NAME) for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		output_name=$(BUILD_DIR)/$(APP_NAME)-$$GOOS-$$GOARCH; \
		if [ $$GOOS = "windows" ]; then output_name="$$output_name.exe"; fi; \
		echo "Building for $$GOOS/$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build $(BUILD_FLAGS) -o $$output_name ./$(MAIN_FILE); \
	done
	@echo "Cross-compilation complete!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf logs
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -cover ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	gofmt -w .
	go mod tidy

# Vet code
vet:
	@echo "Vetting code..."
	go vet ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Development build with debug info
dev-build:
	@echo "Building development version..."
	@mkdir -p $(BUILD_DIR)
	go build -gcflags="all=-N -l" -o $(BUILD_DIR)/$(APP_NAME)-dev ./$(MAIN_FILE)

# Run in development mode
dev-run: dev-build
	@echo "Running in development mode..."
	./$(BUILD_DIR)/$(APP_NAME)-dev -config configs/config-minimal.yaml

# Run with full configuration
run: build
	@echo "Running with full configuration..."
	./$(BUILD_DIR)/$(APP_NAME) -config configs/config.yaml

# Run in server mode
run-server: build
	@echo "Running in server mode..."
	./$(BUILD_DIR)/$(APP_NAME) -config configs/config.yaml -server -port 8888

# Install as systemd service (Linux only)
install-service: build
	@echo "Installing systemd service..."
	@sudo cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/
	@sudo cp scripts/ssh-tunnel-manager.service /etc/systemd/system/
	@sudo systemctl daemon-reload
	@sudo systemctl enable ssh-tunnel-manager
	@echo "Service installed. Start with: sudo systemctl start ssh-tunnel-manager"

# Generate TLS certificates for development
generate-certs:
	@echo "Generating TLS certificates..."
	@mkdir -p certs
	openssl req -new -newkey rsa:4096 -days 365 -nodes -x509 \
		-subj "/C=US/ST=State/L=City/O=Organization/CN=localhost" \
		-keyout certs/server.key -out certs/server.crt
	@echo "Certificates generated in certs/ directory"

# Package for distribution
package: build-all
	@echo "Creating distribution packages..."
	@mkdir -p dist
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		archive_name=$(APP_NAME)-$(VERSION)-$$GOOS-$$GOARCH; \
		if [ $$GOOS = "windows" ]; then \
			cd $(BUILD_DIR) && zip -r ../dist/$$archive_name.zip $(APP_NAME)-$$GOOS-$$GOARCH.exe ../configs ../README.md; \
		else \
			cd $(BUILD_DIR) && tar -czf ../dist/$$archive_name.tar.gz $(APP_NAME)-$$GOOS-$$GOARCH ../configs ../README.md; \
		fi; \
	done
	@echo "Distribution packages created in dist/ directory"

# Setup development environment
setup-dev:
	@echo "Setting up development environment..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/cosmtrek/air@latest
	@echo "Development tools installed"

# Update dependencies
update-deps:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

# Docker run
docker-run:
	@echo "Running Docker container..."
	docker run -d --name $(APP_NAME) \
		-p 8080:8080 -p 8888:8888 \
		-v $(PWD)/configs:/app/configs \
		$(APP_NAME):latest

# Show available targets
help:
	@echo "Available targets:"
	@echo "  all           - Clean, deps, test, and build"
	@echo "  build         - Build for current platform"
	@echo "  build-all     - Build for all platforms"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Lint code"
	@echo "  fmt           - Format code"
	@echo "  vet           - Vet code"
	@echo "  deps          - Install dependencies"
	@echo "  dev-build     - Build development version with debug info"
	@echo "  dev-run       - Run in development mode"
	@echo "  run           - Run with full configuration"
	@echo "  run-server    - Run in server mode"
	@echo "  install-service - Install as systemd service (Linux)"
	@echo "  generate-certs  - Generate TLS certificates"
	@echo "  package       - Create distribution packages"
	@echo "  setup-dev     - Setup development environment"
	@echo "  update-deps   - Update dependencies"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  help          - Show this help message" 