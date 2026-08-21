#!/usr/bin/env bash
set -euo pipefail

# Rebuild the checked-in desktop/PWA icon set from the canonical BOB wordmark
# source. macOS tools are used for ICNS generation; ImageMagick creates ICO.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE="$ROOT_DIR/assets/bob-gemini-free-logo.jpg"
SOURCE_PNG=""
TAURI_DIR="$ROOT_DIR/desktop/src-tauri/icons"
FRONTEND_DIR="$ROOT_DIR/cmd/desktop/frontend/assets"
TEMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TEMP_DIR"' EXIT

if [[ ! -f "$SOURCE" ]]; then
  echo "missing canonical icon: $SOURCE" >&2
  exit 1
fi
if ! command -v sips >/dev/null || ! command -v iconutil >/dev/null || ! command -v magick >/dev/null; then
  echo "requires macOS sips/iconutil and ImageMagick magick" >&2
  exit 1
fi

mkdir -p "$TAURI_DIR" "$FRONTEND_DIR" "$TEMP_DIR/icon.iconset"
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

resize 32 "$TAURI_DIR/32x32.png"
resize 128 "$TAURI_DIR/128x128.png"
resize 256 "$TAURI_DIR/128x128@2x.png"
cp "$TAURI_DIR/128x128@2x.png" "$FRONTEND_DIR/bob-gemini-free-icon.png"

resize 16 "$TEMP_DIR/icon.iconset/icon_16x16.png"
resize 32 "$TEMP_DIR/icon.iconset/icon_16x16@2x.png"
resize 32 "$TEMP_DIR/icon.iconset/icon_32x32.png"
resize 64 "$TEMP_DIR/icon.iconset/icon_32x32@2x.png"
resize 128 "$TEMP_DIR/icon.iconset/icon_128x128.png"
resize 256 "$TEMP_DIR/icon.iconset/icon_128x128@2x.png"
resize 256 "$TEMP_DIR/icon.iconset/icon_256x256.png"
resize 512 "$TEMP_DIR/icon.iconset/icon_256x256@2x.png"
resize 512 "$TEMP_DIR/icon.iconset/icon_512x512.png"
resize 1024 "$TEMP_DIR/icon.iconset/icon_512x512@2x.png"
iconutil -c icns "$TEMP_DIR/icon.iconset" -o "$TAURI_DIR/icon.icns"
magick "$SOURCE_PNG" -define icon:auto-resize=256,128,64,48,32,16 "$TAURI_DIR/icon.ico"
cp "$TAURI_DIR/icon.ico" "$ROOT_DIR/web/favicon.ico"
cp "$TAURI_DIR/icon.ico" "$ROOT_DIR/internal/server/favicon.ico"
xattr -c "$TAURI_DIR/icon.icns" "$TAURI_DIR/icon.ico" "$ROOT_DIR/web/favicon.ico" "$ROOT_DIR/internal/server/favicon.ico" 2>/dev/null || true

echo "desktop icons rebuilt from $SOURCE"
