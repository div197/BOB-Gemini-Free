#!/usr/bin/env bash
set -euo pipefail

# Rebuild the checked-in Wails/PWA icon set from the canonical BOB wordmark
# source. ImageMagick creates the browser/server ICO; sips provides the
# bootstrap PNG consumed by the Wails frontend.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE="$ROOT_DIR/assets/bob-gemini-free-logo.jpg"
SOURCE_PNG=""
FRONTEND_DIR="$ROOT_DIR/cmd/desktop/frontend/assets"
TEMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TEMP_DIR"' EXIT

if [[ ! -f "$SOURCE" ]]; then
  echo "missing canonical icon: $SOURCE" >&2
  exit 1
fi
if ! command -v sips >/dev/null || ! command -v magick >/dev/null; then
  echo "requires macOS sips and ImageMagick magick" >&2
  exit 1
fi

mkdir -p "$FRONTEND_DIR"
SOURCE_PNG="$TEMP_DIR/bob-gemini-free-icon.png"
magick "$SOURCE" -resize 1024x1024 \
  -stroke '#f6c453' -strokewidth 16 -fill none \
  -draw 'roundrectangle 28,28 996,996 190,190' \
  -stroke '#2bd4ee' -strokewidth 4 \
  -draw 'roundrectangle 52,52 972,972 174,174' \
  "$SOURCE_PNG"
cp "$SOURCE_PNG" "$ROOT_DIR/cmd/desktop/build/appicon.png"

resize() {
  local size="$1"
  local output="$2"
  sips -s format png -z "$size" "$size" "$SOURCE_PNG" --out "$output" >/dev/null
  xattr -c "$output" 2>/dev/null || true
}

resize 256 "$FRONTEND_DIR/bob-gemini-free-icon.png"
magick "$SOURCE_PNG" -define icon:auto-resize=256,128,64,48,32,16 "$ROOT_DIR/web/favicon.ico"
cp "$ROOT_DIR/web/favicon.ico" "$ROOT_DIR/internal/server/favicon.ico"
xattr -c "$ROOT_DIR/web/favicon.ico" "$ROOT_DIR/internal/server/favicon.ico" 2>/dev/null || true

echo "desktop icons rebuilt from $SOURCE"
