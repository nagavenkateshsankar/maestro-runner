#!/bin/bash
# Installation script for maestro-runner
# Installs to ~/.maestro-runner/ with bin/, cache/, drivers/ layout
# Set MAESTRO_RUNNER_HOME to override the install location.

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "Installing maestro-runner..."
echo ""

# Detect OS
OS="$(uname -s)"
ARCH="$(uname -m)"

if [ "$OS" != "Darwin" ] && [ "$OS" != "Linux" ]; then
    echo -e "${RED}Error: Unsupported operating system: $OS${NC}"
    echo "maestro-runner currently supports macOS and Linux only."
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    echo "Please install Go from https://golang.org/dl/"
    exit 1
fi

echo -e "${GREEN}✓${NC} Go found: $(go version)"
echo ""

# Build from source
echo "Building maestro-runner from source..."
go build -o maestro-runner .

if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Build failed${NC}"
    exit 1
fi

echo -e "${GREEN}✓${NC} Build successful"
echo ""

# Determine install location
INSTALL_DIR="${MAESTRO_RUNNER_HOME:-$HOME/.maestro-runner}"
BIN_DIR="$INSTALL_DIR/bin"

echo "Installing to $INSTALL_DIR..."
mkdir -p "$BIN_DIR"

# Move binary
mv maestro-runner "$BIN_DIR/maestro-runner"
chmod +x "$BIN_DIR/maestro-runner"

# Copy drivers if present
if [ -d "drivers" ]; then
    cp -r drivers "$INSTALL_DIR/"
    echo -e "${GREEN}✓${NC} Drivers copied to $INSTALL_DIR/drivers/"
fi

# macOS specific: Remove quarantine attribute
if [ "$OS" = "Darwin" ]; then
    echo "Removing macOS quarantine attribute..."
    xattr -d com.apple.quarantine "$BIN_DIR/maestro-runner" 2>/dev/null || true
    echo -e "${GREEN}✓${NC} macOS Gatekeeper bypass applied"
fi

echo -e "${GREEN}✓${NC} Installed to $BIN_DIR/maestro-runner"
echo ""

# Check if bin dir is in PATH
if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
    echo -e "${YELLOW}Warning: $BIN_DIR is not in your PATH${NC}"
    echo ""
    echo "Add this line to your shell profile (~/.bashrc, ~/.zshrc, or ~/.profile):"
    echo ""
    echo "  export PATH=\"$BIN_DIR:\$PATH\""
    echo ""
    echo "Then restart your shell or run: source ~/.zshrc"
    echo ""
else
    echo -e "${GREEN}✓${NC} Installation complete!"
    echo ""
    echo "Run: maestro-runner --help"
fi
