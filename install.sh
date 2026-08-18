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

# 2. Check if Go is installed
if command -v go >/dev/null 2>&1; then
    echo -e "${BLUE}[*] Go detected ($(go version)). Compiling from source...${NC}"
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$APP_NAME" .
    echo -e "${GREEN}[+] Successfully built $APP_NAME binary!${NC}"
else
    # Check if Docker is installed
    if command -v docker >/dev/null 2>&1; then
        echo -e "${YELLOW}[!] Go is not installed, but Docker was detected.${NC}"
        echo -e "${BLUE}[*] Building local Docker image: $APP_NAME...${NC}"
        docker build -t "$APP_NAME" .
        echo -e "${GREEN}[+] Docker container built successfully!${NC}"
        echo -e "${GREEN}[*] Run anytime with:${NC} docker run -d --name $APP_NAME -p 8081:8081 $APP_NAME"
        exit 0
    else
        echo -e "${YELLOW}[!] Neither Go nor Docker is installed.${NC}"
        echo -e "${BLUE}[*] Please install Go (https://go.dev/dl/) or Docker (https://docker.com) to run BOB Gemini Free.${NC}"
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
