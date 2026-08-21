#!/usr/bin/env bash
set -euo pipefail

# Build a locally signed macOS Wails bundle from a clean /tmp staging tree.
# macOS File Provider checkouts can add resource-fork metadata that makes
# Wails' in-place codesign fail; staging removes that environmental variable
# without changing repository files.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="${1:-/tmp/bob-gemini-free-wails-release}"
STAGE_DIR="$(mktemp -d /tmp/bob-gemini-free-wails-source.XXXXXX)"
STAGE_ROOT="$STAGE_DIR/repo"
trap 'rm -rf "$STAGE_DIR"' EXIT

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "this script builds the macOS Wails bundle and requires macOS" >&2
  exit 1
fi
if [[ -e "$OUTPUT_DIR" ]]; then
  echo "output already exists; choose a clean path: $OUTPUT_DIR" >&2
  exit 1
fi
if ! command -v rsync >/dev/null || ! command -v codesign >/dev/null; then
  echo "requires rsync and codesign" >&2
  exit 1
fi

mkdir -p "$OUTPUT_DIR"
rsync -a \
  --exclude '.git/' \
  --exclude 'cmd/desktop/build/bin/' \
  --exclude 'release-assets/' \
  "$ROOT_DIR/" "$STAGE_ROOT/"
xattr -cr "$STAGE_ROOT" 2>/dev/null || true

if [[ -n "${WAILS_BIN:-}" ]]; then
  WAILS=("$WAILS_BIN")
else
  WAILS=(go run github.com/wailsapp/wails/v2/cmd/wails@v2.15.0)
fi

cd "$STAGE_ROOT/cmd/desktop"
"${WAILS[@]}" build -clean

SOURCE_APP="$STAGE_ROOT/cmd/desktop/build/bin/bob-gemini-free-wails.app"
DEST_APP="$OUTPUT_DIR/bob-gemini-free-wails.app"
if [[ ! -d "$SOURCE_APP" ]]; then
  echo "Wails did not produce the expected app bundle: $SOURCE_APP" >&2
  exit 1
fi

ditto --norsrc --noextattr --noqtn "$SOURCE_APP" "$DEST_APP"
xattr -cr "$DEST_APP" 2>/dev/null || true
codesign --force --deep --sign - "$DEST_APP"
codesign --verify --deep --strict --verbose=2 "$DEST_APP"

echo "Wails macOS app ready: $DEST_APP"
