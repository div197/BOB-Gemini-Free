#!/usr/bin/env bash
set -euo pipefail

# Build a locally signed BOB Gemini Free macOS bundle from a clean /tmp staging
# tree. macOS File Provider checkouts can add resource-fork metadata that makes
# in-place codesign fail; staging removes that environmental variable without
# changing repository files.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="${1:-/tmp/bob-gemini-free-release}"
STAGE_DIR="$(mktemp -d /tmp/bob-gemini-free-source.XXXXXX)"
STAGE_ROOT="$STAGE_DIR/repo"
PLATFORM="${BOB_WAILS_PLATFORM:-darwin/$(go env GOARCH)}"
VERSION="${BOB_RELEASE_VERSION:-dev}"
trap 'rm -rf "$STAGE_DIR"' EXIT

if [[ "$(uname -s)" != "Darwin" ]]; then
	echo "this script builds the macOS native bundle and requires macOS" >&2
	exit 1
fi
if [[ "$PLATFORM" != darwin/* ]]; then
	echo "this script only builds darwin targets; got: $PLATFORM" >&2
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

WAILS_LDFLAGS=(-ldflags "-X main.desktopVersion=${VERSION}")
if [[ -n "${BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY:-}" ]]; then
  WAILS_LDFLAGS=(-ldflags "-X main.desktopVersion=${VERSION} -X github.com/div197/bob-gemini-free/internal/updater.BuildUpdatePublicKey=${BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY}")
fi

cd "$STAGE_ROOT/cmd/desktop"
"${WAILS[@]}" build -clean -platform "$PLATFORM" "${WAILS_LDFLAGS[@]}"

SOURCE_APP="$STAGE_ROOT/cmd/desktop/build/bin/bob-gemini-free.app"
DEST_APP="$OUTPUT_DIR/BOB Gemini Free.app"
if [[ ! -d "$SOURCE_APP" ]]; then
  echo "desktop build did not produce the expected app bundle: $SOURCE_APP" >&2
  exit 1
fi

ditto --norsrc --noextattr --noqtn "$SOURCE_APP" "$DEST_APP"
xattr -cr "$DEST_APP" 2>/dev/null || true
codesign --force --deep --sign - "$DEST_APP"
codesign --verify --deep --strict --verbose=2 "$DEST_APP"

echo "BOB Gemini Free macOS $PLATFORM app ready: $DEST_APP"
