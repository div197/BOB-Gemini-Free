#!/usr/bin/env bash
set -euo pipefail

# Build a free macOS preview package. This intentionally uses ad-hoc signing
# only; it does not claim Apple Developer ID trust or notarization. The output
# is suitable for controlled testing, not a production student release.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="${1:-/tmp/bob-gemini-free-wails-preview}"
PLATFORM="${BOB_WAILS_PLATFORM:-darwin/universal}"
STAGE_DIR="$(mktemp -d /tmp/bob-gemini-free-wails-preview-source.XXXXXX)"
STAGE_ROOT="$STAGE_DIR/repo"
APP_NAME="bob-gemini-free-wails"
trap 'rm -rf "$STAGE_DIR"' EXIT

if [[ "$(uname -s)" != "Darwin" ]]; then
	echo "this preview packager requires macOS" >&2
	exit 1
fi
if [[ "$PLATFORM" != darwin/universal && "$PLATFORM" != darwin/arm64 && "$PLATFORM" != darwin/amd64 ]]; then
	echo "unsupported macOS Wails platform: $PLATFORM" >&2
	exit 1
fi
if [[ -e "$OUTPUT_DIR" ]]; then
	echo "output already exists; choose a clean path: $OUTPUT_DIR" >&2
	exit 1
fi
for command_name in rsync codesign ditto hdiutil shasum; do
	if ! command -v "$command_name" >/dev/null; then
		echo "required command not found: $command_name" >&2
		exit 1
	fi
done

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
"${WAILS[@]}" build -clean -platform "$PLATFORM"

SOURCE_APP="$STAGE_ROOT/cmd/desktop/build/bin/${APP_NAME}.app"
DEST_APP="$OUTPUT_DIR/${APP_NAME}.app"
if [[ ! -d "$SOURCE_APP" ]]; then
	echo "Wails did not produce the expected app bundle: $SOURCE_APP" >&2
	exit 1
fi

ditto --norsrc --noextattr --noqtn "$SOURCE_APP" "$DEST_APP"
xattr -cr "$DEST_APP" 2>/dev/null || true
codesign --force --deep --sign - "$DEST_APP"
codesign --verify --deep --strict --verbose=2 "$DEST_APP"

ZIP_PATH="$OUTPUT_DIR/${APP_NAME}-macos-universal.zip"
DMG_PATH="$OUTPUT_DIR/${APP_NAME}-macos-universal.dmg"
ditto -c -k --norsrc --noextattr --noqtn --keepParent "$DEST_APP" "$ZIP_PATH"
hdiutil create -volname "BOB Gemini Free" -srcfolder "$DEST_APP" -ov -format UDZO "$DMG_PATH" >/dev/null

cat > "$OUTPUT_DIR/RELEASE-NOTICE.txt" <<'NOTICE'
BOB Gemini Free macOS preview package

This build is open-source software for controlled evaluation.
It is ad-hoc signed and is NOT Apple Developer ID signed or notarized.
macOS may require the user to approve the first launch in Finder.
Do not use this artifact as proof of production Mac distribution readiness.
No Google session, cookie, API key, or private release key is included.
NOTICE

(
	cd "$OUTPUT_DIR"
	shasum -a 256 "$ZIP_PATH" "$DMG_PATH" "RELEASE-NOTICE.txt" > SHA256SUMS
)

echo "macOS preview artifacts ready in: $OUTPUT_DIR"
ls -lh "$OUTPUT_DIR"
