#!/usr/bin/env bash
set -euo pipefail

# Verify a complete local release directory after signing or after downloading
# the exact public release assets. This checks the control files, the signed
# manifest, and the one-to-one relationship between manifest entries and
# regular files. It never changes the directory and never accesses GitHub.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RELEASE_DIR="${1:-}"
PUBLIC_KEY="${BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY:-}"

if [[ -z "$RELEASE_DIR" || ! -d "$RELEASE_DIR" ]]; then
	echo "usage: BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY=... $0 RELEASE_DIR" >&2
	exit 1
fi
if [[ -z "$PUBLIC_KEY" ]]; then
	echo "BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY is required" >&2
	exit 1
fi

MANIFEST="$RELEASE_DIR/SHA256SUMS"
SIGNATURE="$RELEASE_DIR/SHA256SUMS.sig"
if [[ ! -f "$MANIFEST" || ! -f "$SIGNATURE" ]]; then
	echo "release directory must contain SHA256SUMS and SHA256SUMS.sig" >&2
	exit 1
fi

for path in "$MANIFEST" "$SIGNATURE"; do
	if [[ -L "$path" || ! -f "$path" ]]; then
		echo "release control file is not a regular file: $path" >&2
		exit 1
	fi
done

cd "$ROOT_DIR"
if ! go run ./cmd/release-verify \
	-dir "$RELEASE_DIR" \
	-public-key "$PUBLIC_KEY"; then
	echo "release manifest verification failed" >&2
	exit 1
fi

echo "release assets verified: $RELEASE_DIR"
