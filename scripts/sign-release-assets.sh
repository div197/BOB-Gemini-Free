#!/usr/bin/env bash
set -euo pipefail

# Sign an already-built release directory locally. This is deliberately a
# separate operator step: build/signing credentials never belong in source or
# GitHub Actions. An existing unsigned SHA256SUMS is regenerated from the
# exact directory contents; an existing signed manifest is never overwritten.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RELEASE_DIR="${1:-}"
PUBLIC_KEY="${BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY:-}"

if [[ -z "$RELEASE_DIR" || ! -d "$RELEASE_DIR" ]]; then
	echo "usage: BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY=... BOB_GEMINI_FREE_UPDATE_PRIVATE_KEY=... $0 RELEASE_DIR" >&2
	exit 1
fi
if [[ -z "$PUBLIC_KEY" || -z "${BOB_GEMINI_FREE_UPDATE_PRIVATE_KEY:-}" ]]; then
	echo "both update public and private keys are required in the local release environment" >&2
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
BOB_GEMINI_FREE_UPDATE_PRIVATE_KEY="$BOB_GEMINI_FREE_UPDATE_PRIVATE_KEY" \
	go run ./cmd/release-manifest -dir "$RELEASE_DIR" -public-key "$PUBLIC_KEY"

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

echo "signed release assets are ready: $RELEASE_DIR"
