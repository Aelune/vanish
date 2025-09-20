#!/bin/bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Repository information
OWNER="Aelune"
REPO="vanish"
BINARY_NAME="vx"
DEFAULT_TAG="v0.9.0"

# Use provided version or default
VERSION="${1:-$DEFAULT_TAG}"

# Remove 'v' prefix if present
VERSION="${VERSION#v}"

echo -e "${BLUE}🗑️  Vanish Installer${NC}"
echo -e "${BLUE}================${NC}"

# Detect operating system
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
    linux*) OS="linux" ;;
    darwin*) OS="darwin" ;;
    mingw*|msys*|cygwin*) OS="windows" ;;
    *)
        echo -e "${RED}❌ Unsupported operating system: $OS${NC}"
        exit 1
        ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    armv7l) ARCH="arm" ;;
    i386|i686) ARCH="386" ;;
    *)
        echo -e "${RED}❌ Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

# Set binary extension for Windows
BINARY_EXT=""


# Construct download URL
BINARY_FILE="${BINARY_NAME}-${OS}-${ARCH}${BINARY_EXT}"
DOWNLOAD_URL="https://github.com/$OWNER/$REPO/releases/download/v${VERSION}/${BINARY_FILE}"

echo -e "${YELLOW}📋 System Information:${NC}"
echo "   OS: $OS"
echo "   Architecture: $ARCH"
echo "   Version: v$VERSION"
echo "   Binary: $BINARY_FILE"
echo

# Determine installation directory
if [ "$OS" = "windows" ]; then
    echo -e "${red} doesn't support windows for now "
else
    # Try different installation directories in order of preference
    if [ -w "/usr/local/bin" ] 2>/dev/null; then
        INSTALL_DIR="/usr/local/bin"
    elif [ -d "$HOME/.local/bin" ]; then
        INSTALL_DIR="$HOME/.local/bin"
    else
        INSTALL_DIR="$HOME/bin"
        mkdir -p "$INSTALL_DIR"
    fi
    TARGET_PATH="$INSTALL_DIR/$BINARY_NAME"
fi

# Create installation directory if it doesn't exist
mkdir -p "$INSTALL_DIR"

echo -e "${YELLOW}⬇ Downloading from: $DOWNLOAD_URL${NC}"

# Download the binary
if command -v curl >/dev/null 2>&1; then
    if ! curl -fsSL "$DOWNLOAD_URL" -o "$TARGET_PATH"; then
        echo -e "${RED}❌ Failed to download $BINARY_NAME${NC}"
        echo -e "${RED}   URL: $DOWNLOAD_URL${NC}"
        exit 1
    fi
elif command -v wget >/dev/null 2>&1; then
    if ! wget -q "$DOWNLOAD_URL" -O "$TARGET_PATH"; then
        echo -e "${RED}❌ Failed to download $BINARY_NAME${NC}"
        echo -e "${RED}   URL: $DOWNLOAD_URL${NC}"
        exit 1
    fi
else
    echo -e "${RED}❌ Neither curl nor wget found. Please install one of them.${NC}"
    exit 1
fi

# Make executable (not needed on Windows)
if [ "$OS" != "windows" ]; then
    chmod +x "$TARGET_PATH"
fi

# Verify installation
if [ -f "$TARGET_PATH" ]; then
    echo -e "${GREEN}✅ Successfully installed $BINARY_NAME v$VERSION${NC}"
    echo -e "${GREEN}   Location: $TARGET_PATH${NC}"

    # Check if directory is in PATH
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo
        echo -e "${YELLOW}⚠️  Installation directory not in PATH${NC}"
        echo -e "${YELLOW}   Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):${NC}"
        echo -e "${BLUE}   export PATH=\"$INSTALL_DIR:\$PATH\"${NC}"
        echo
        echo -e "${YELLOW}   Or run: echo 'export PATH=\"$INSTALL_DIR:\$PATH\"' >> ~/.bashrc${NC}"
    fi

    echo
    echo -e "${GREEN}🚀 Quick Start:${NC}"
    echo "   $BINARY_NAME --help    # Show help"
    echo "   $BINARY_NAME -t        # Choose theme"
    echo "   $BINARY_NAME file.txt  # Delete safely"
    echo "   $BINARY_NAME -l        # List cached files"
    echo
    echo -e "${BLUE}📚 Documentation: https://github.com/$OWNER/$REPO${NC}"

else
    echo -e "${RED}❌ Installation failed - binary not found at $TARGET_PATH${NC}"
    exit 1
fi
