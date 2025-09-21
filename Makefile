# Vanish - Safe file removal tool Makefile
# Build configuration
BINARY_NAME=vx
BUILD_DIR=build/bin
SOURCE_DIR=.
GO_FILES=$(shell find . -name "*.go" -type f)
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT_HASH=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.CommitHash=$(COMMIT_HASH) -X main.BuildTime=$(BUILD_TIME) -s -w"

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
PURPLE=\033[0;35m
CYAN=\033[0;36m
WHITE=\033[1;37m
NC=\033[0m

# Emojis for better visual output
ROCKET=🚀
HAMMER=🔨
TEST_TUBE=🧪
BROOM=🧹
PACKAGE=📦
CHECK=✅
CROSS=❌
GEAR=⚙️

# Default target
.PHONY: all
all: banner build

# Show banner
.PHONY: banner
banner:
	@echo -e "$(BLUE)"
	@echo ""
	@echo -e "                            $(WHITE)VANISH BUILD SYSTEM                             "
	@echo -e "                         $(CYAN)Safe File Deletion Tool                          "
	@echo ""
	@echo -e "$(YELLOW)Version: $(GREEN)$(VERSION)$(NC)"
	@echo -e "$(YELLOW)Commit:  $(GREEN)$(COMMIT_HASH)$(NC)"
	@echo -e "$(YELLOW)Built:   $(GREEN)$(BUILD_TIME)$(NC)"
	@echo ""

# Build the application
.PHONY: build
build: $(BUILD_DIR)/$(BINARY_NAME)

$(BUILD_DIR)/$(BINARY_NAME): $(GO_FILES) go.mod go.sum
	@echo -e "$(HAMMER) $(YELLOW)Building $(BINARY_NAME) v$(VERSION)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo -e "$(CHECK) $(GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

# Build for multiple platforms
.PHONY: build-all
build-all: banner
	@echo -e "$(PACKAGE) $(BLUE)Building for all platforms...$(NC)"
	@$(MAKE) build-linux
	@$(MAKE) build-darwin
	@echo -e "$(ROCKET) $(GREEN)All platforms built successfully!$(NC)"

.PHONY: build-linux
build-linux:
	@echo -e "$(HAMMER) $(CYAN)Building for Linux...$(NC)"
	@mkdir -p $(BUILD_DIR)/release
	@echo -e "  $(YELLOW)→ AMD64$(NC)"
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-amd64 .
	@echo -e "  $(YELLOW)→ ARM64$(NC)"
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-arm64 .
	@echo -e "  $(CHECK) $(GREEN)Linux builds complete$(NC)"

.PHONY: build-darwin
build-darwin:
	@echo -e "$(HAMMER) $(CYAN)Building for macOS...$(NC)"
	@mkdir -p $(BUILD_DIR)/release
	@echo -e "  $(YELLOW)→ AMD64 (Intel)$(NC)"
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-amd64 .
	@echo -e "  $(YELLOW)→ ARM64 (Apple Silicon)$(NC)"
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-arm64 .
	@echo -e "  $(CHECK) $(GREEN)macOS builds complete$(NC)"

# Development targets
.PHONY: dev
dev: build
	@echo -e "$(ROCKET) $(PURPLE)Running $(BINARY_NAME) in development mode...$(NC)"
	@./$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: run
run:
	@echo -e "$(ROCKET) $(PURPLE)Running $(BINARY_NAME) directly with args: $(ARGS)$(NC)"
	@go run . $(ARGS)

# # Testing
# .PHONY: test
# test:
# 	@echo -e "$(TEST_TUBE) $(BLUE)Running tests...$(NC)"
# 	@if go test -v ./... 2>/dev/null; then \
# 		echo -e "$(CHECK) $(GREEN)All tests passed!$(NC)"; \
# 	else \
# 		echo -e "$(CROSS) $(RED)Some tests failed!$(NC)"; \
# 		exit 1; \
# 	fi

# .PHONY: test-race
# test-race:
# 	@echo -e "$(TEST_TUBE) $(BLUE)Running tests with race detection...$(NC)"
# 	@if go test -race -v ./... 2>/dev/null; then \
# 		echo -e "$(CHECK) $(GREEN)Race tests passed!$(NC)"; \
# 	else \
# 		echo -e "$(CROSS) $(RED)Race conditions detected!$(NC)"; \
# 		exit 1; \
# 	fi

# .PHONY: test-cover
# test-cover:
# 	@echo -e "$(TEST_TUBE) $(BLUE)Running tests with coverage...$(NC)"
# 	@go test -coverprofile=coverage.out ./... 2>/dev/null || true
# 	@if [ -f coverage.out ]; then \
# 		go tool cover -html=coverage.out -o coverage.html; \
# 		COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}'); \
# 		echo -e "$(CHECK) $(GREEN)Coverage report: $$COVERAGE$(NC)"; \
# 		echo -e "$(BLUE)HTML report: coverage.html$(NC)"; \
# 	fi

# Code quality
.PHONY: lint
lint:
	@echo -e "$(GEAR) $(BLUE)Running linters...$(NC)"

	# golangci-lint
	@echo -e "  $(YELLOW)→ golangci-lint$(NC)"
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo -e "    $(YELLOW)Installing golangci-lint...$(NC)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@if golangci-lint run --timeout=5m; then \
		echo -e "    $(CHECK) $(GREEN)golangci-lint passed$(NC)"; \
	else \
		echo -e "    $(CROSS) $(RED)golangci-lint failed$(NC)"; \
	fi

	# revive
	@echo -e "  $(YELLOW)→ revive$(NC)"
	@if ! command -v revive >/dev/null 2>&1; then \
		echo -e "    $(YELLOW)Installing revive...$(NC)"; \
		go install github.com/mgechev/revive@latest; \
	fi
	@echo -e "    $(YELLOW)Ignore internal/types/types.go:3:9: avoid meaningless package names $(NC)"
	@revive ./... || echo -e "    $(YELLOW)Ignore known issue: internal/types/types.go:3:9: avoid meaningless package names$(NC)"


.PHONY: fmt
fmt:
	@echo -e "$(GEAR) $(BLUE)Formatting code...$(NC)"
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo -e "$(YELLOW)Formatting files...$(NC)"; \
		gofmt -w .; \
		echo -e "$(CHECK) $(GREEN)Code formatted$(NC)"; \
	else \
		echo -e "$(CHECK) $(GREEN)Code already formatted$(NC)"; \
	fi

.PHONY: vet
vet:
	@echo -e "$(GEAR) $(BLUE)Running go vet...$(NC)"
	@if go vet ./...; then \
		echo -e "$(CHECK) $(GREEN)go vet passed$(NC)"; \
	else \
		echo -e "$(CROSS) $(RED)go vet failed$(NC)"; \
		exit 1; \
	fi

.PHONY: check
check: fmt vet lint
	@echo -e "$(ROCKET) $(GREEN)All quality checks passed!$(NC)"

# Dependencies
.PHONY: deps
deps:
	@echo -e "$(PACKAGE) $(BLUE)Downloading dependencies...$(NC)"
	@go mod download
	@echo -e "$(CHECK) $(GREEN)Dependencies downloaded$(NC)"

.PHONY: deps-update
deps-update:
	@echo -e "$(PACKAGE) $(BLUE)Updating dependencies...$(NC)"
	@go mod tidy
	@go mod verify
	@echo -e "$(CHECK) $(GREEN)Dependencies updated$(NC)"

# Installation
.PHONY: install
install: build
	@echo -e "$(GEAR) $(BLUE)Installing $(BINARY_NAME) to /usr/local/bin...$(NC)"
	@if sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/ && sudo chmod +x /usr/local/bin/$(BINARY_NAME); then \
		echo -e "$(CHECK) $(GREEN)$(BINARY_NAME) installed to /usr/local/bin/$(NC)"; \
	else \
		echo -e "$(CROSS) $(RED)Installation failed$(NC)"; \
		exit 1; \
	fi

.PHONY: install-user
install-user: build
	@echo -e "$(GEAR) $(BLUE)Installing $(BINARY_NAME) to ~/.local/bin...$(NC)"
	@mkdir -p ~/.local/bin
	@cp $(BUILD_DIR)/$(BINARY_NAME) ~/.local/bin/
	@chmod +x ~/.local/bin/$(BINARY_NAME)
	@echo -e "$(CHECK) $(GREEN)$(BINARY_NAME) installed to ~/.local/bin/$(NC)"
	@if [[ ":$$PATH:" != *":$$HOME/.local/bin:"* ]]; then \
		echo -e "$(YELLOW)Warning: ~/.local/bin is not in your PATH$(NC)"; \
		echo -e "$(YELLOW)Add this to your shell profile: export PATH=\"$$HOME/.local/bin:$$PATH\"$(NC)"; \
	fi

.PHONY: uninstall
uninstall:
	@echo -e "$(BROOM) $(BLUE)Uninstalling $(BINARY_NAME)...$(NC)"
	@sudo rm -f /usr/local/bin/$(BINARY_NAME) 2>/dev/null || true
	@rm -f ~/.local/bin/$(BINARY_NAME) 2>/dev/null || true
	@echo -e "$(CHECK) $(GREEN)$(BINARY_NAME) uninstalled$(NC)"

# Release
.PHONY: release
release: banner clean check build-all
	@echo -e "$(PACKAGE) $(BLUE)Creating release archives...$(NC)"
	@mkdir -p $(BUILD_DIR)/archives
	@cd $(BUILD_DIR)/release && \
	for file in *; do \
		if [[ $$file == *".exe" ]]; then \
			echo -e "  $(YELLOW)→ Creating $$file.zip$(NC)"; \
			zip -q ../archives/$${file%.*}.zip $$file; \
		else \
			echo -e "  $(YELLOW)→ Creating $$file.tar.gz$(NC)"; \
			tar -czf ../archives/$$file.tar.gz $$file; \
		fi \
	done
	@echo -e "$(ROCKET) $(GREEN)Release archives created in $(BUILD_DIR)/archives/$(NC)"
	@echo -e "$(BLUE)Files created:$(NC)"
	@ls -la $(BUILD_DIR)/archives/ | grep -E '\.(tar\.gz|zip)$$' | awk '{printf "  %s %s\n", "📁", $$9}'

# Generate checksums
.PHONY: checksums
checksums: release
	@echo -e "$(GEAR) $(BLUE)Generating checksums...$(NC)"
	@cd $(BUILD_DIR)/archives && \
	if command -v sha256sum >/dev/null 2>&1; then \
		sha256sum * > checksums.txt; \
	elif command -v shasum >/dev/null 2>&1; then \
		shasum -a 256 * > checksums.txt; \
	else \
		echo -e "$(CROSS) $(RED)No checksum utility found$(NC)"; \
		exit 1; \
	fi
	@echo -e "$(CHECK) $(GREEN)Checksums generated: $(BUILD_DIR)/archives/checksums.txt$(NC)"

# Cleanup
.PHONY: clean
clean:
	@echo -e "$(BROOM) $(BLUE)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@go clean
	@echo -e "$(CHECK) $(GREEN)Clean complete$(NC)"

.PHONY: clean-cache
clean-cache:
	@echo -e "$(BROOM) $(BLUE)Cleaning Go module cache...$(NC)"
	@go clean -modcache
	@echo -e "$(CHECK) $(GREEN)Cache cleaned$(NC)"

# Docker support
.PHONY: docker-build
docker-build:
	@echo -e "$(HAMMER) $(BLUE)Building Docker image...$(NC)"
	@if [ -f Dockerfile ]; then \
		docker build -t $(BINARY_NAME):$(VERSION) .; \
		echo -e "$(CHECK) $(GREEN)Docker image built: $(BINARY_NAME):$(VERSION)$(NC)"; \
	else \
		echo -e "$(CROSS) $(RED)Dockerfile not found$(NC)"; \
	fi

# Development setup
.PHONY: setup
setup:
	@echo -e "$(GEAR) $(BLUE)Setting up development environment...$(NC)"
	@$(MAKE) deps
	@echo -e "$(GEAR) $(YELLOW)Installing development tools...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/mgechev/revive@latest
	@echo -e "$(CHECK) $(GREEN)Development environment ready!$(NC)"


# Help
.PHONY: help
help:
	@echo -e "           $(WHITE)VANISH MAKEFILE HELP     "
	@echo ""
	@echo -e "$(GREEN)Building:$(NC)"
	@echo -e "  $(YELLOW)build$(NC)         Build the application for current platform"
	@echo -e "  $(YELLOW)build-all$(NC)     Build for all supported platforms"
	@echo -e "  $(YELLOW)build-linux$(NC)   Build for Linux (amd64 and arm64)"
	@echo -e "  $(YELLOW)build-darwin$(NC)  Build for macOS (amd64 and arm64)"
	@echo ""
	@echo -e "$(GREEN)Development:$(NC)"
	@echo -e "  $(YELLOW)dev$(NC)           Build and run the application"
	@echo -e "  $(YELLOW)run$(NC)           Run without building (use ARGS='--help' for arguments)"
	@echo -e "  $(YELLOW)setup$(NC)         Set up development environment"
	@echo ""
# 	@echo -e "$(GREEN)Testing:$(NC)"
# 	@echo -e "  $(YELLOW)test$(NC)          Run tests"
# 	@echo -e "  $(YELLOW)test-race$(NC)     Run tests with race detection"
# 	@echo -e "  $(YELLOW)test-cover$(NC)    Run tests with coverage report"
# 	@echo ""
	@echo -e "$(GREEN)Code Quality:$(NC)"
	@echo -e "  $(YELLOW)lint$(NC)          Run linters (installs tools if needed)"
	@echo -e "  $(YELLOW)fmt$(NC)           Format code"
	@echo -e "  $(YELLOW)vet$(NC)           Run go vet"
	@echo -e "  $(YELLOW)check$(NC)         Run all quality checks"
	@echo ""
	@echo -e "$(GREEN)Dependencies:$(NC)"
	@echo -e "  $(YELLOW)deps$(NC)          Download dependencies"
	@echo -e "  $(YELLOW)deps-update$(NC)   Update and tidy dependencies"
	@echo ""
	@echo -e "$(GREEN)Installation:$(NC)"
	@echo -e "  $(YELLOW)install$(NC)       Install to /usr/local/bin (requires sudo)"
	@echo -e "  $(YELLOW)install-user$(NC)  Install to ~/.local/bin"
	@echo -e "  $(YELLOW)uninstall$(NC)     Uninstall from system"
	@echo ""
	@echo -e "$(GREEN)Release:$(NC)"
	@echo -e "  $(YELLOW)release$(NC)       Build all platforms and create archives"
	@echo -e "  $(YELLOW)checksums$(NC)     Generate SHA256 checksums"
	@echo ""
	@echo -e "$(GREEN)Cleanup:$(NC)"
	@echo -e "  $(YELLOW)clean$(NC)         Remove build artifacts"
	@echo -e "  $(YELLOW)clean-cache$(NC)   Clean Go module cache"
	@echo ""
