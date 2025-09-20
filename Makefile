# Vanish - Safe file removal tool MakeFile
# Build configuration
BINARY_NAME=vx
BUILD_DIR=build/bin
SOURCE_DIR=.
GO_FILES=$(shell find . -name "*.go" -type f)
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -s -w"

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build: $(BUILD_DIR)/$(BINARY_NAME)

$(BUILD_DIR)/$(BINARY_NAME): $(GO_FILES) go.mod go.sum
	@mkdir -p $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
.PHONY: build-all
build-all: build-linux build-darwin # build-windows

.PHONY: build-linux
build-linux:
	@mkdir -p $(BUILD_DIR)/linux
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/linux/$(BINARY_NAME) .
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/linux/$(BINARY_NAME)-arm64 .

.PHONY: build-darwin
build-darwin:
	@mkdir -p $(BUILD_DIR)/darwin
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/darwin/$(BINARY_NAME) .
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/darwin/$(BINARY_NAME)-arm64 .

# .PHONY: build-windows
# build-windows:
# 	@mkdir -p $(BUILD_DIR)/windows
# 	@echo "Building for Windows..."
# 	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/windows/$(BINARY_NAME).exe .

# Development targets
.PHONY: dev
dev: build
	@echo "Running $(BINARY_NAME) in development mode..."
	./$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: run
run:
	@echo "Running $(BINARY_NAME) directly..."
	go run . $(ARGS)

# Testing
# .PHONY: test
# test:
# 	@echo "Running tests..."
# 	go test -v ./...

# .PHONY: test-race
# test-race:
# 	@echo "Running tests with race detection..."
# 	go test -race -v ./...

# .PHONY: test-cover
# test-cover:
# 	@echo "Running tests with coverage..."
# 	go test -coverprofile=coverage.out ./...
# 	go tool cover -html=coverage.out -o coverage.html
# 	@echo "Coverage report generated: coverage.html"

# Code quality
.PHONY: lint
lint:
	@echo "Running linter..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Installing..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; }
	golangci-lint run

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

.PHONY: check
check: fmt vet lint test

# Dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download

.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	go mod tidy
	go mod verify

# Installation
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "$(BINARY_NAME) installed successfully!"

.PHONY: install-user
install-user: build
	@echo "Installing $(BINARY_NAME) to ~/.local/bin..."
	@mkdir -p ~/.local/bin
	cp $(BUILD_DIR)/$(BINARY_NAME) ~/.local/bin/
	chmod +x ~/.local/bin/$(BINARY_NAME)
	@echo "$(BINARY_NAME) installed to ~/.local/bin/"
	@echo "Make sure ~/.local/bin is in your shell PATH"

.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	rm -f ~/.local/bin/$(BINARY_NAME)
	@echo "$(BINARY_NAME) uninstalled successfully!"

# Release
.PHONY: release
release: clean check build-all
	@echo "Creating release archives..."
	@mkdir -p $(BUILD_DIR)/release
	@cd $(BUILD_DIR)/linux && tar -czf ../release/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)
	@cd $(BUILD_DIR)/linux && tar -czf ../release/$(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-arm64
	@cd $(BUILD_DIR)/darwin && tar -czf ../release/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)
	@cd $(BUILD_DIR)/darwin && tar -czf ../release/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-arm64
	@cd $(BUILD_DIR)/windows && zip -q ../release/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME).exe
	@echo "Release archives created in $(BUILD_DIR)/release/"

# Cleanup
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	go clean

.PHONY: clean-cache
clean-cache:
	@echo "Cleaning Go module cache..."
	go clean -modcache

# Help
.PHONY: help
help:
	@echo "Vanish (vx) - Makefile targets:"
	@echo ""
	@echo "Building:"
	@echo "  build         Build the application for current platform"
	@echo "  build-all     Build for all supported platforms"
	@echo "  build-linux   Build for Linux (amd64 and arm64)"
	@echo "  build-darwin  Build for macOS (amd64 and arm64)"
	@echo "  build-windows Build for Windows (amd64)"
	@echo ""
	@echo "Development:"
	@echo "  dev           Build and run the application"
	@echo "  run           Run without building (use ARGS='--help' for arguments)"
	@echo "  test          Run tests"
	@echo "  test-race     Run tests with race detection"
	@echo "  test-cover    Run tests with coverage report"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint          Run linter (installs golangci-lint if needed)"
	@echo "  fmt           Format code"
	@echo "  vet           Run go vet"
	@echo "  check         Run fmt, vet, lint, and test"
	@echo ""
	@echo "Dependencies:"
	@echo "  deps          Download dependencies"
	@echo "  deps-update   Update and tidy dependencies"
	@echo ""
	@echo "Installation:"
	@echo "  install       Install to /usr/local/bin (requires sudo)"
	@echo "  install-user  Install to ~/.local/bin"
	@echo "  uninstall     Uninstall from system"
	@echo ""
	@echo "Release:"
	@echo "  release       Build all platforms and create release archives"
	@echo ""
	@echo "Cleanup:"
	@echo "  clean         Remove build artifacts"
	@echo "  clean-cache   Clean Go module cache"
	@echo "Examples:"
	@echo "  make build              # Build for current platform"
	@echo "  make run ARGS='--help'  # Run with --help argument"
	@echo "  make test               # Run tests"
	@echo "  make install-user       # Install to ~/.local/bin"
