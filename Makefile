# Vanish - Safe file removal tool
# Build configuration

# Variables
BINARY_NAME=vx
BINARY_PATH=./bin/$(BINARY_NAME)
MAIN_PATH=./main.go
GO_FILES=$(shell find . -name "*.go" -type f)
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -s -w"

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build: $(BINARY_PATH)

$(BINARY_PATH): $(GO_FILES)
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	go build $(LDFLAGS) -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_PATH)"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -rf bin/
	go clean
	@echo "Clean complete"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Install the binary to system PATH
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp $(BINARY_PATH) /usr/local/bin/$(BINARY_NAME)
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "Installation complete"

# Uninstall the binary from system PATH
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstall complete"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	golangci-lint run

# Run the application in development
.PHONY: run
run: build
	$(BINARY_PATH) $(ARGS)

# Build for multiple platforms
.PHONY: build-all
build-all: build-linux build-darwin build-windows

.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p bin
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p bin
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

# Create release packages
.PHONY: package
package: build-all
	@echo "Creating release packages..."
	@mkdir -p releases
	tar -czf releases/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz -C bin $(BINARY_NAME)-linux-amd64
	tar -czf releases/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz -C bin $(BINARY_NAME)-darwin-amd64
	tar -czf releases/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz -C bin $(BINARY_NAME)-darwin-arm64
	zip -j releases/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip bin/$(BINARY_NAME)-windows-amd64.exe
	@echo "Packages created in releases/"

# Development setup
.PHONY: dev-setup
dev-setup:
	@echo "Setting up development environment..."
	go mod tidy
	go mod download
	@echo "Development setup complete"

# Generate documentation
.PHONY: docs
docs:
	@echo "Generating documentation..."
	go doc -all . > docs/API.md
	@echo "Documentation generated"

# Check for security vulnerabilities
.PHONY: security
security:
	@echo "Checking for security vulnerabilities..."
	govulncheck ./...

# Benchmark tests
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Profile the application
.PHONY: profile
profile: build
	@echo "Running with profiling..."
	$(BINARY_PATH) -cpuprofile=cpu.prof -memprofile=mem.prof $(ARGS)

# Show help
.PHONY: help
help:
	@echo "Vanish Build System"
	@echo "==================="
	@echo ""
	@echo "Available targets:"
	@echo "  build         Build the binary"
	@echo "  build-all     Build for all platforms"
	@echo "  clean         Remove build artifacts"
	@echo "  test          Run tests"
	@echo "  test-coverage Run tests with coverage"
	@echo "  install       Install to system PATH"
	@echo "  uninstall     Remove from system PATH"
	@echo "  fmt           Format code"
	@echo "  lint          Lint code"
	@echo "  run           Run the application (use ARGS=... for arguments)"
	@echo "  package       Create release packages"
	@echo "  dev-setup     Set up development environment"
	@echo "  docs          Generate documentation"
	@echo "  security      Check for security vulnerabilities"
	@echo "  bench         Run benchmarks"
	@echo "  profile       Run with profiling"
	@echo "  help          Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make run ARGS='file1.txt file2.txt'"
	@echo "  make run ARGS='--themes'"
	@echo "  make run ARGS='--noconfirm *.log'"
