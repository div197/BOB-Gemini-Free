#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# BOB Gemini Free - Universal Installer for macOS and Linux
# Break Ordinary Boundaries | Powered by ABCsteps (https://abcsteps.com)
# Author: Divyanshu Singh Chouhan (@div197)
# ==============================================================================

BOLD='\033[1m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

UPDATE_PUBLIC_KEY_HEX="ba7854781bca2a14da4f1ec5e931ff45f458ac9377c42ac127c349e5ecad2dff"
UPDATE_PUBLIC_KEY_SPKI_B64="MCowBQYDK2VwAyEAunhUeBvKKhTaTx7F6TH/RfRYrJN3xCrBJ8NJ5eytLf8="
RELEASE_DOWNLOAD_BASE="https://github.com/div197/BOB-Gemini-Free/releases/latest/download"
MAX_MANIFEST_BYTES=$((1024 * 1024))
MAX_SIGNATURE_BYTES=4096
MAX_RELEASE_ASSET_BYTES=$((512 * 1024 * 1024))

die() {
    echo -e "${RED}[!] $*${NC}" >&2
    exit 1
}

download_release_file() {
    local url="$1"
    local destination="$2"
    if command -v curl >/dev/null 2>&1; then
        curl --fail --silent --show-error --location \
            --proto '=https' --proto-redir '=https' \
            --max-time 120 --retry 2 --retry-delay 1 \
            --output "$destination" "$url"
    elif command -v wget >/dev/null 2>&1; then
        wget --https-only --timeout=120 --tries=3 \
            --output-document="$destination" "$url"
    else
        die "curl or wget is required to download the signed release"
    fi
}

file_size() {
    wc -c < "$1" | tr -d '[:space:]'
}

sha256_file() {
    if command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$1" | awk '{print $1}'
    elif command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$1" | awk '{print $1}'
    else
        die "shasum or sha256sum is required to verify the signed release"
    fi
}

verify_release_asset() {
    local manifest_path="$1"
    local signature_path="$2"
    local asset_path="$3"
    local asset_name="$4"
    local public_key_pem="$TMP_DIR/update-public-key.pem"
    local decoded_signature="$TMP_DIR/SHA256SUMS.sig.decoded"

    command -v openssl >/dev/null 2>&1 || die "an OpenSSL build with Ed25519 support is required to verify the release; no binary was installed"

    if [ "$(file_size "$manifest_path")" -gt "$MAX_MANIFEST_BYTES" ]; then
        die "signed release manifest is larger than the safe limit"
    fi
    if [ "$(file_size "$signature_path")" -gt "$MAX_SIGNATURE_BYTES" ]; then
        die "signed release signature is larger than the safe limit"
    fi

    printf '%s\n%s\n%s\n' \
        '-----BEGIN PUBLIC KEY-----' "$UPDATE_PUBLIC_KEY_SPKI_B64" '-----END PUBLIC KEY-----' > "$public_key_pem"

    tr -d '[:space:]' < "$signature_path" | openssl base64 -d -A > "$decoded_signature" \
        || die "release signature is not valid base64"
    if [ "$(file_size "$decoded_signature")" -ne 64 ]; then
        die "release signature has an invalid Ed25519 length"
    fi
    if ! openssl pkeyutl -verify -rawin -pubin -inkey "$public_key_pem" \
        -in "$manifest_path" -sigfile "$decoded_signature" >/dev/null 2>&1; then
        die "release manifest signature verification failed; no binary was installed"
    fi

    local expected_digest
    expected_digest="$(awk -v asset="$asset_name" '
        /^[[:space:]]*$/ { next }
        {
            if (NF != 2 || length($1) != 64 || $1 !~ /^[0-9a-fA-F]+$/) bad = 1
            if ($2 == asset) {
                matches++
                expected = tolower($1)
            }
        }
        END {
            if (bad || matches != 1) exit 1
            print expected
        }
    ' "$manifest_path")" || die "signed release manifest has no unique valid entry for $asset_name"

    local actual_digest
    actual_digest="$(sha256_file "$asset_path")"
    actual_digest="$(printf '%s' "$actual_digest" | tr '[:upper:]' '[:lower:]')"
    expected_digest="$(printf '%s' "$expected_digest" | tr '[:upper:]' '[:lower:]')"
    if [ "$actual_digest" != "$expected_digest" ]; then
        die "downloaded $asset_name does not match its signed SHA-256 digest"
    fi
}

install_verified_binary() {
    local destination="$1"
    local destination_dir
    destination_dir="$(dirname "$destination")"
    mkdir -p "$destination_dir"
    [ -L "$destination" ] && die "refusing to replace a symlink at $destination"

    # Copy into the destination directory first, then rename within that
    # directory. This keeps an interrupted install from leaving a truncated
    # executable at the user-facing path, even when /tmp is another volume.
    local staged_destination
    staged_destination="$(mktemp "$destination_dir/.bob-gemini-free.XXXXXX")" \
        || die "could not create an atomic install staging file"
    if ! cp "$TMP_FILE" "$staged_destination"; then
        rm -f "$staged_destination"
        die "could not copy the verified release into the install directory"
    fi
    chmod 0755 "$staged_destination"
    if ! mv -f "$staged_destination" "$destination"; then
        rm -f "$staged_destination"
        die "could not atomically install the verified release"
    fi
}

echo -e "${BLUE}${BOLD}================================================================${NC}"
echo -e "${GREEN}${BOLD}    BOB Gemini Free - Break Ordinary Boundaries                ${NC}"
echo -e "${BLUE}    Powered by ABCsteps.com | Divyanshu Singh Chouhan (@div197) ${NC}"
echo -e "${BLUE}${BOLD}================================================================${NC}"
echo ""

APP_NAME="bob-gemini-free"
USER_HOME="${HOME:-}"
[ -n "$USER_HOME" ] || die "HOME is not set"
CONFIG_DIR="$USER_HOME/.config/bob-gemini-free"
LOCAL_BIN="$USER_HOME/.local/bin"

# Detect OS and Arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo -e "${RED}[!] Unsupported architecture: $ARCH. Please build from source.${NC}"; exit 1 ;;
esac

TARGET_BIN=""
TMP_DIR=""

# Source builds are deliberately opt-in. The default installer must not treat
# an arbitrary current directory with a matching go.mod as trusted source.
INSTALL_FROM_SOURCE="${BOB_GEMINI_FREE_INSTALL_FROM_SOURCE:-0}"
case "$INSTALL_FROM_SOURCE" in
    0) ;;
    1) ;;
    *) die "BOB_GEMINI_FREE_INSTALL_FROM_SOURCE must be 0 or 1" ;;
esac

if [ "$INSTALL_FROM_SOURCE" = "1" ]; then
    command -v go >/dev/null 2>&1 || die "Go is required for the explicit source-build path"
    [ -f "go.mod" ] || die "explicit source-build path requires a repository checkout"
    grep -q '^module github.com/div197/bob-gemini-free$' go.mod || die "current directory is not the BOB Gemini Free source module"
    echo -e "${BLUE}[*] Explicit source-build mode enabled; compiling the checked-out module...${NC}"
    CGO_ENABLED=0 go build \
        -ldflags="-s -w -X github.com/div197/bob-gemini-free/internal/updater.BuildUpdatePublicKey=$UPDATE_PUBLIC_KEY_HEX" \
        -o "$APP_NAME" .
    TARGET_BIN="./$APP_NAME"
    echo -e "${GREEN}[✔] Successfully built $APP_NAME binary!${NC}"
else
    # 2. Download and authenticate the pre-compiled release.
    echo -e "${BLUE}[*] Fetching and authenticating the latest signed standalone binary for ${OS}-${ARCH}...${NC}"
    TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/${APP_NAME}-install.XXXXXX")"
    trap 'rm -rf -- "$TMP_DIR"' EXIT
    ASSET_NAME="${APP_NAME}-${OS}-${ARCH}"
    MANIFEST_PATH="$TMP_DIR/SHA256SUMS"
    SIGNATURE_PATH="$TMP_DIR/SHA256SUMS.sig"
    TMP_FILE="$TMP_DIR/$ASSET_NAME"

    download_release_file "$RELEASE_DOWNLOAD_BASE/SHA256SUMS" "$MANIFEST_PATH" \
        || die "could not download the signed release manifest"
    download_release_file "$RELEASE_DOWNLOAD_BASE/SHA256SUMS.sig" "$SIGNATURE_PATH" \
        || die "could not download the release signature"
    download_release_file "$RELEASE_DOWNLOAD_BASE/$ASSET_NAME" "$TMP_FILE" \
        || die "signed release asset is not available on GitHub Releases"
    [ -f "$TMP_FILE" ] && [ ! -L "$TMP_FILE" ] || die "downloaded release asset is not a regular file"
    [ "$(file_size "$TMP_FILE")" -le "$MAX_RELEASE_ASSET_BYTES" ] || die "release asset is larger than the safe limit"
    verify_release_asset "$MANIFEST_PATH" "$SIGNATURE_PATH" "$TMP_FILE" "$ASSET_NAME"
    chmod 0755 "$TMP_FILE"

    # Try to install globally or to ~/.local/bin.
    if [ -d "/usr/local/bin" ] && [ -w "/usr/local/bin" ]; then
        install_verified_binary "/usr/local/bin/$APP_NAME"
        TARGET_BIN="/usr/local/bin/$APP_NAME"
        echo -e "${GREEN}[✔] Installed authenticated release to $TARGET_BIN${NC}"
    else
        mkdir -p "$LOCAL_BIN"
        install_verified_binary "$LOCAL_BIN/$APP_NAME"
        TARGET_BIN="$LOCAL_BIN/$APP_NAME"
        echo -e "${GREEN}[✔] Installed authenticated release to $TARGET_BIN${NC}"

        # Warn if ~/.local/bin is not in PATH
        if [[ ":$PATH:" != *":$LOCAL_BIN:"* ]]; then
            echo -e "${YELLOW}[!] Note: $LOCAL_BIN is not in your PATH.${NC}"
            echo -e "    Add it by running: ${BOLD}export PATH=\"\$HOME/.local/bin:\$PATH\"${NC}"
        fi
    fi
fi

mkdir -p "$CONFIG_DIR"
if [ ! -f "$CONFIG_DIR/config.json" ] && [ -f "config.example.json" ]; then
    cp config.example.json "$CONFIG_DIR/config.json"
    chmod 0600 "$CONFIG_DIR/config.json"
    echo -e "${GREEN}[✔] Created default config at $CONFIG_DIR/config.json${NC}"
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
