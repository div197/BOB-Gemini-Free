#!/usr/bin/env bash
set -euo pipefail

# Sign an already-built release directory locally. This is deliberately a
# separate operator step: build/signing credentials never belong in source or
# GitHub Actions. On macOS, when the private-key environment variable is not
# set, the key is streamed from the owner-controlled Keychain directly into
# the signer over stdin. An existing unsigned SHA256SUMS is regenerated from
# the exact directory contents; an existing signed manifest is never overwritten.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RELEASE_DIR="${1:-}"
PUBLIC_KEY="${BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY:-}"
KEYCHAIN_SERVICE="${BOB_GEMINI_FREE_UPDATE_KEYCHAIN_SERVICE:-BOB-Gemini-Free-Release-Ed25519}"
KEYCHAIN_ACCOUNT="${BOB_GEMINI_FREE_UPDATE_KEYCHAIN_ACCOUNT:-$(id -un)}"

if [[ -z "$RELEASE_DIR" || ! -d "$RELEASE_DIR" ]]; then
	echo "usage: BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY=... $0 RELEASE_DIR" >&2
	exit 1
fi
if [[ -z "$PUBLIC_KEY" ]]; then
	echo "BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY is required" >&2
	exit 1
fi
if [[ -e "$RELEASE_DIR/SHA256SUMS.sig" ]]; then
	echo "release directory already contains SHA256SUMS.sig; refusing to overwrite a signed manifest" >&2
	exit 1
fi
if [[ -e "$RELEASE_DIR/SHA256SUMS" ]]; then
	echo "regenerating the unsigned SHA256SUMS from the inspected release directory" >&2
fi

cd "$ROOT_DIR"
if [[ -n "${BOB_GEMINI_FREE_UPDATE_PRIVATE_KEY:-}" ]]; then
	BOB_GEMINI_FREE_UPDATE_PRIVATE_KEY="$BOB_GEMINI_FREE_UPDATE_PRIVATE_KEY" \
		go run ./cmd/release-manifest -dir "$RELEASE_DIR" -public-key "$PUBLIC_KEY"
elif [[ "$(uname -s)" == "Darwin" ]] && command -v security >/dev/null; then
	security find-generic-password -s "$KEYCHAIN_SERVICE" -a "$KEYCHAIN_ACCOUNT" -w |
		go run ./cmd/release-manifest -private-key-stdin -dir "$RELEASE_DIR" -public-key "$PUBLIC_KEY"
else
	echo "private key is unavailable: set it through a local secret manager or use the macOS Keychain path" >&2
	exit 1
fi

if command -v shasum >/dev/null; then
	(
		cd "$RELEASE_DIR"
		shasum -a 256 -c SHA256SUMS
	)
elif command -v sha256sum >/dev/null; then
	(
		cd "$RELEASE_DIR"
		sha256sum -c SHA256SUMS
	)
else
	echo "neither shasum nor sha256sum is available for local verification" >&2
	exit 1
fi

bash "$ROOT_DIR/scripts/verify-release-assets.sh" "$RELEASE_DIR"

echo "signed release assets are ready: $RELEASE_DIR"
