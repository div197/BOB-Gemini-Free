with open('install.sh', 'w') as f:
    f.write("""#!/usr/bin/env bash
set -e

# ==============================================================================
# BOB Gemini Free - Universal Installer for macOS and Linux
# Break Ordinary Boundaries | Powered by ABCsteps (https://abcsteps.com)
# Author: Divyanshu Singh Chouhan (@div197)
# ==============================================================================

BOLD='\\033[1m'
GREEN='\\033[0;32m'
BLUE='\\033[0;34m'
YELLOW='\\033[1;33m'
RED='\\033[0;31m'
NC='\\033[0m'

echo -e "${BLUE}${BOLD}================================================================${NC}"
echo -e "${GREEN}${BOLD}    BOB Gemini Free - Break Ordinary Boundaries                ${NC}"
echo -e "${BLUE}    Powered by ABCsteps.com | Divyanshu Singh Chouhan (@div197) ${NC}"
echo -e "${BLUE}${BOLD}================================================================${NC}"
echo ""

APP_NAME="bob-gemini-free"
CONFIG_DIR="$HOME/.config/bob-gemini-free"
LOCAL_BIN="$HOME/.local/bin"

mkdir -p "$CONFIG_DIR"
if [ ! -f "$CONFIG_DIR/config.json" ] && [ -f "config.example.json" ]; then
    cp config.example.json "$CONFIG_DIR/config.json"
    echo -e "${GREEN}[✔] Created default config at $CONFIG_DIR/config.json${NC}"
fi

# Detect OS and Arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo -e "${RED}[!] Unsupported architecture: $ARCH. Please build from source.${NC}"; exit 1 ;;
esac

TARGET_BIN=""

# 1. Check if compiling from source is possible/intended
if command -v go >/dev/null 2>&1 && [ -f "go.mod" ] && grep -q "bob-gemini-free" go.mod; then
    echo -e "${BLUE}[*] Go detected in source tree. Compiling locally...${NC}"
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$APP_NAME" .
    TARGET_BIN="./$APP_NAME"
    echo -e "${GREEN}[✔] Successfully built $APP_NAME binary!${NC}"
else
    # 2. Download Pre-compiled release
    echo -e "${BLUE}[*] Fetching latest pre-compiled standalone binary for ${OS}-${ARCH}...${NC}"
    DOWNLOAD_URL="https://github.com/div197/bob-gemini-free/releases/latest/download/bob-gemini-free-${OS}-${ARCH}"
    
    TMP_FILE="/tmp/${APP_NAME}"
    if curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE" 2>/dev/null || wget -qO "$TMP_FILE" "$DOWNLOAD_URL" 2>/dev/null; then
        chmod +x "$TMP_FILE"
        
        # Try to install globally or to ~/.local/bin
        if [ -w "/usr/local/bin" ]; then
            mv "$TMP_FILE" "/usr/local/bin/$APP_NAME"
            TARGET_BIN="/usr/local/bin/$APP_NAME"
            echo -e "${GREEN}[✔] Installed globally to $TARGET_BIN${NC}"
        else
            mkdir -p "$LOCAL_BIN"
            mv "$TMP_FILE" "$LOCAL_BIN/$APP_NAME"
            TARGET_BIN="$LOCAL_BIN/$APP_NAME"
            echo -e "${GREEN}[✔] Installed to $TARGET_BIN${NC}"
            
            # Warn if ~/.local/bin is not in PATH
            if [[ ":$PATH:" != *":$LOCAL_BIN:"* ]]; then
                echo -e "${YELLOW}[!] Note: $LOCAL_BIN is not in your PATH.${NC}"
                echo -e "    Add it by running: ${BOLD}export PATH=\"\$HOME/.local/bin:\$PATH\"${NC}"
            fi
        fi
    else
        echo -e "${RED}[!] Pre-compiled binary not yet available on GitHub Releases.${NC}"
        echo -e "${BLUE}[*] Please install Go (https://go.dev/dl/) to build locally.${NC}"
        exit 1
    fi
fi

echo ""
echo -e "${GREEN}${BOLD}================================================================${NC}"
echo -e "${GREEN}${BOLD}    INSTALLATION COMPLETE! 🚀${NC}"
echo -e "${GREEN}${BOLD}================================================================${NC}"
echo ""
echo -e "To launch the gateway and open the Web Studio, run:"
if [[ "$TARGET_BIN" == "./"* ]]; then
    echo -e "  ${BOLD}${TARGET_BIN} --port 9610${NC}"
else
    echo -e "  ${BOLD}${APP_NAME} --port 9610${NC}"
fi
echo ""
echo -e "API Base URL: ${BLUE}http://127.0.0.1:9610/v1${NC}"
echo -e "UI Dashboard: ${BLUE}http://127.0.0.1:9610/playground${NC}"
echo ""
""")
