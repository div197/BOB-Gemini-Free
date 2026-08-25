#!/usr/bin/env bash
set -euo pipefail

# Build a free BOB Gemini Free macOS beta package. This intentionally uses
# ad-hoc signing only; it does not claim Apple Developer ID trust or
# notarization. The output is suitable for controlled testing, not a
# production student release.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="${1:-/tmp/bob-gemini-free-preview}"
PLATFORM="${BOB_WAILS_PLATFORM:-darwin/universal}"
STAGE_DIR="$(mktemp -d /tmp/bob-gemini-free-preview-source.XXXXXX)"
STAGE_ROOT="$STAGE_DIR/repo"
INTERNAL_APP_NAME="bob-gemini-free"
PUBLIC_APP_NAME="BOB Gemini Free"
VERSION="${BOB_RELEASE_VERSION:-v0.1.7}"
trap 'rm -rf "$STAGE_DIR"' EXIT

if [[ "$(uname -s)" != "Darwin" ]]; then
	echo "this preview packager requires macOS" >&2
	exit 1
fi
if [[ "$PLATFORM" != darwin/universal && "$PLATFORM" != darwin/arm64 && "$PLATFORM" != darwin/amd64 ]]; then
	echo "unsupported macOS desktop platform: $PLATFORM" >&2
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

WAILS_LDFLAGS=(-ldflags "-X main.desktopVersion=${VERSION}")
if [[ -n "${BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY:-}" ]]; then
	WAILS_LDFLAGS=(-ldflags "-X main.desktopVersion=${VERSION} -X github.com/div197/bob-gemini-free/internal/updater.BuildUpdatePublicKey=${BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY}")
fi

cd "$STAGE_ROOT/cmd/desktop"
"${WAILS[@]}" build -clean -platform "$PLATFORM" "${WAILS_LDFLAGS[@]}"

SOURCE_APP="$STAGE_ROOT/cmd/desktop/build/bin/${INTERNAL_APP_NAME}.app"
DEST_APP="$OUTPUT_DIR/${PUBLIC_APP_NAME}.app"
if [[ ! -d "$SOURCE_APP" ]]; then
	echo "desktop build did not produce the expected app bundle: $SOURCE_APP" >&2
	exit 1
fi

ditto --norsrc --noextattr --noqtn "$SOURCE_APP" "$DEST_APP"
xattr -cr "$DEST_APP" 2>/dev/null || true
codesign --force --deep --sign - "$DEST_APP"
codesign --verify --deep --strict --verbose=2 "$DEST_APP"

ZIP_PATH="$OUTPUT_DIR/bob-gemini-free-macos-universal.zip"
DMG_PATH="$OUTPUT_DIR/bob-gemini-free-macos-universal.dmg"
ditto -c -k --norsrc --noextattr --noqtn --keepParent "$DEST_APP" "$ZIP_PATH"

# Build a conventional drag-to-install DMG root. A DMG containing only the
# application bundle is technically mountable, but it gives users no visible
# installation target and makes the preview feel unfinished. Keep the alias
# outside the signed app bundle so it cannot affect bundle verification.
DMG_ROOT="$STAGE_DIR/dmg-root"
mkdir -p "$DMG_ROOT"
ditto --norsrc --noextattr --noqtn "$DEST_APP" "$DMG_ROOT/$PUBLIC_APP_NAME.app"
ln -s /Applications "$DMG_ROOT/Applications"
if [[ ! -L "$DMG_ROOT/Applications" || "$(readlink "$DMG_ROOT/Applications")" != "/Applications" ]]; then
	echo "failed to create the DMG Applications shortcut" >&2
	exit 1
fi
hdiutil create -volname "BOB Gemini Free" -srcfolder "$DMG_ROOT" -ov -format UDZO "$DMG_PATH" >/dev/null
bash "$ROOT_DIR/scripts/verify-macos-dmg-layout.sh" "$DMG_PATH"

cat > "$OUTPUT_DIR/RELEASE-NOTICE.txt" <<'NOTICE'
BOB Gemini Free macOS open-source beta package

This is the complete BOB Gemini Free desktop application for open-source
evaluation. It is ad-hoc signed for bundle integrity, but it is not signed by
Apple Developer ID or notarized; macOS may require first-launch approval.
That platform-trust limitation does not change the product identity or the
fact that this is a genuine build from the public source repository.
The package is a beta and should be tested before a broad student rollout.
No Google session, cookie, API key, or private release key is included.
NOTICE

(
	cd "$OUTPUT_DIR"
	shasum -a 256 "$ZIP_PATH" "$DMG_PATH" "RELEASE-NOTICE.txt" > SHA256SUMS
)

echo "BOB Gemini Free macOS preview artifacts ready in: $OUTPUT_DIR"
ls -lh "$OUTPUT_DIR"
