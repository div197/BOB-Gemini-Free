#!/usr/bin/env bash
set -e

# ==============================================================================
# BOB Gemini Free - Universal Installer for macOS and Linux
# Break Ordinary Boundaries | Powered by ABCsteps (https://abcsteps.com)
# Author: Divyanshu Singh Chouhan (@div197)
# ==============================================================================

BOLD='\033[1m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}${BOLD}================================================================${NC}"
echo -e "${GREEN}${BOLD}    BOB Gemini Free - Break Ordinary Boundaries                ${NC}"
echo -e "${BLUE}    Powered by ABCsteps.com | Divyanshu Singh Chouhan (@div197) ${NC}"
echo -e "${BLUE}${BOLD}================================================================${NC}"
echo ""

APP_NAME="bob-gemini-free"
CONFIG_DIR="$HOME/.config/bob-gemini-free"
INSTALL_DIR="/usr/local/bin"

# 1. Ensure config directory exists
mkdir -p "$CONFIG_DIR"
if [ ! -f "$CONFIG_DIR/config.json" ]; then
    if [ -f "config.example.json" ]; then
        cp config.example.json "$CONFIG_DIR/config.json"
        echo -e "${GREEN}[+] Created default config at $CONFIG_DIR/config.json${NC}"
    fi
fi

# 2. Check if local binary already exists
if [ -f "./$APP_NAME" ]; then
    echo -e "${GREEN}[✔] Existing $APP_NAME binary found locally.${NC}"
# 3. Check if Go is installed to compile from source
elif command -v go >/dev/null 2>&1; then
    echo -e "${BLUE}[*] Go detected ($(go version)). Compiling from source...${NC}"
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$APP_NAME" .
    echo -e "${GREEN}[+] Successfully built $APP_NAME binary!${NC}"
# 4. Check if Docker is installed
elif command -v docker >/dev/null 2>&1; then
    echo -e "${YELLOW}[!] Go is not installed, but Docker was detected.${NC}"
    echo -e "${BLUE}[*] Building local Docker image: $APP_NAME...${NC}"
    docker build -t "$APP_NAME" .
    echo -e "${GREEN}[+] Docker container built successfully!${NC}"
    echo -e "${GREEN}[*] Run anytime with:${NC} docker run -d --name $APP_NAME -p 8081:8081 $APP_NAME"
    exit 0
# 5. Zero-dependency fallback: Auto-download pre-compiled binary for OS & Architecture
else
    echo -e "${BLUE}[*] No Go or Docker detected. Fetching pre-compiled standalone binary...${NC}"
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *) echo -e "${YELLOW}[!] Unsupported architecture: $ARCH. Please build from source.${NC}"; exit 1 ;;
    esac

    DOWNLOAD_URL="https://github.com/div197/bob-gemini-free/releases/latest/download/bob-gemini-free-${OS}-${ARCH}"
    echo -e "${BLUE}[*] Downloading from: $DOWNLOAD_URL${NC}"
    if curl -fsSL "$DOWNLOAD_URL" -o "$APP_NAME" 2>/dev/null || wget -qO "$APP_NAME" "$DOWNLOAD_URL" 2>/dev/null; then
        chmod +x "$APP_NAME"
        echo -e "${GREEN}[+] Standalone binary installed successfully!${NC}"
    else
        echo -e "${YELLOW}[!] Pre-compiled binary not yet available on GitHub Releases.${NC}"
        echo -e "${BLUE}[*] Please install Go (https://go.dev/dl/) to build locally, or download a release binary.${NC}"
        exit 1
    fi
fi

echo ""
echo -e "${GREEN}${BOLD}[✔] Setup Complete!${NC}"
echo -e "Start the server by running:"
echo -e "  ${BOLD}./$APP_NAME --port 8081${NC}"
echo ""
echo -e "Base URL: ${BOLD}http://127.0.0.1:8081/v1${NC}"
echo -e "Visit ABCsteps: ${BLUE}https://abcsteps.com/${NC}"
echo ""
