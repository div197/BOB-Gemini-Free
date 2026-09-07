#!/usr/bin/env bash
set -euo pipefail

# Wails emits the native executable inside the app/installer. Checking only
# the requested ldflags is not enough: stale build output or an incorrect
# artifact could otherwise be packaged and signed as the wrong release.

if [[ $# -ne 4 ]]; then
	echo "usage: $0 BINARY EXPECTED_VERSION EXPECTED_CHANNEL EXPECTED_PUBLIC_KEY" >&2
	exit 2
fi

BINARY_PATH="$1"
EXPECTED_VERSION="$2"
EXPECTED_CHANNEL="$3"
EXPECTED_PUBLIC_KEY="$4"

if [[ -z "$EXPECTED_VERSION" || -z "$EXPECTED_CHANNEL" || -z "$EXPECTED_PUBLIC_KEY" ]]; then
	echo "desktop build identity values must not be empty" >&2
	exit 2
fi
if [[ -L "$BINARY_PATH" || ! -f "$BINARY_PATH" ]]; then
	echo "desktop build identity binary is not a regular file" >&2
	exit 1
fi
if ! command -v go >/dev/null 2>&1; then
	echo "Go is required to inspect desktop build identity" >&2
	exit 1
fi

INFO_FILE="$(mktemp "${TMPDIR:-/tmp}/bob-desktop-build-identity.XXXXXX")"
trap 'rm -f "$INFO_FILE"' EXIT

if ! go version -m "$BINARY_PATH" >"$INFO_FILE" 2>/dev/null; then
	echo "compiled desktop binary has no readable Go build metadata" >&2
	exit 1
fi

if ! grep -Fq -- "-X main.desktopVersion=$EXPECTED_VERSION" "$INFO_FILE"; then
	echo "compiled desktop binary version does not match the requested release" >&2
	exit 1
fi
if ! grep -Fq -- "-X main.desktopChannel=$EXPECTED_CHANNEL" "$INFO_FILE"; then
	echo "compiled desktop binary channel does not match the requested release" >&2
	exit 1
fi
if ! grep -Fq -- "BuildUpdatePublicKey=$EXPECTED_PUBLIC_KEY" "$INFO_FILE"; then
	echo "compiled desktop binary trust key does not match the requested release" >&2
	exit 1
fi

echo "desktop build identity verified: version=$EXPECTED_VERSION channel=$EXPECTED_CHANNEL public-key=matched"
