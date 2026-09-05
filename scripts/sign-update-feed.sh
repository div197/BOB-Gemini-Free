#!/usr/bin/env bash
set -euo pipefail

# Sign the checked-in discovery feed locally. The private key is streamed from
# the owner-controlled macOS Keychain into the signer and is never written to
# Git, shell history, or command output. This is deliberately manual; no
# GitHub Actions or hosted signing secret is required.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FEED_FILE="${1:-$ROOT_DIR/updates/desktop-feed.json}"
SIGNATURE_FILE="${2:-$FEED_FILE.sig}"
PUBLIC_KEY="${BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY:-}"
KEYCHAIN_SERVICE="${BOB_GEMINI_FREE_UPDATE_KEYCHAIN_SERVICE:-BOB-Gemini-Free-Release-Ed25519}"
KEYCHAIN_ACCOUNT="${BOB_GEMINI_FREE_UPDATE_KEYCHAIN_ACCOUNT:-$(id -un)}"

if [[ ! -f "$FEED_FILE" ]]; then
	echo "update feed does not exist: $FEED_FILE" >&2
	exit 1
fi
if [[ -z "$PUBLIC_KEY" ]]; then
	echo "BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY is required" >&2
	exit 1
fi

cd "$ROOT_DIR"
if [[ -n "${BOB_GEMINI_FREE_UPDATE_PRIVATE_KEY:-}" ]]; then
	BOB_GEMINI_FREE_UPDATE_PRIVATE_KEY="$BOB_GEMINI_FREE_UPDATE_PRIVATE_KEY" \
		go run ./cmd/update-feed -file "$FEED_FILE" -signature "$SIGNATURE_FILE" -public-key "$PUBLIC_KEY"
elif [[ "$(uname -s)" == "Darwin" ]] && command -v security >/dev/null; then
	security find-generic-password -s "$KEYCHAIN_SERVICE" -a "$KEYCHAIN_ACCOUNT" -w |
		go run ./cmd/update-feed -private-key-stdin -file "$FEED_FILE" -signature "$SIGNATURE_FILE" -public-key "$PUBLIC_KEY"
else
	echo "private key is unavailable: use the owner-controlled macOS Keychain or a local secret manager" >&2
	exit 1
fi

echo "signed update feed is ready: $FEED_FILE"
