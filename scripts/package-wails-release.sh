#!/usr/bin/env bash
set -euo pipefail

# Build a macOS native release candidate without signing the release manifest.
# The app is ad-hoc signed for bundle integrity; Apple Developer ID signing and
# notarization remain separate external distribution gates. The private
# Ed25519 key is intentionally not accepted or needed by this script.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${BOB_RELEASE_VERSION:-}"
CHANNEL="${BOB_RELEASE_CHANNEL:-stable}"
OUTPUT_DIR="${1:-/tmp/bob-gemini-free-release}"
PLATFORM="${BOB_WAILS_PLATFORM:-darwin/universal}"
PUBLIC_KEY="${BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY:-}"
EXPECTED_PUBLIC_KEY="$(awk '
	/^Encoding: hexadecimal Ed25519 public key$/ { in_key=1; next }
	in_key && /^[[:space:]]*$/ { in_key=0 }
	in_key && length($0)==64 && $0 !~ /[^0-9a-fA-F]/ { print; exit }
' "$ROOT_DIR/docs/engineering/UPDATE-PUBLIC-KEY.txt")"
STAGE_DIR="$(mktemp -d /tmp/bob-gemini-free-release-source.XXXXXX)"
STAGE_ROOT="$STAGE_DIR/repo"
INTERNAL_APP_NAME="bob-gemini-free"
PUBLIC_APP_NAME="BOB Gemini Free"
ARCH_LABEL="${PLATFORM#darwin/}"
trap 'rm -rf "$STAGE_DIR"' EXIT

if [[ -z "$VERSION" ]]; then
	echo "BOB_RELEASE_VERSION is required (for example v0.2.0)" >&2
	exit 1
fi
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-preview\.[0-9]+)?$ ]]; then
	echo "unsupported release version: $VERSION" >&2
	exit 1
fi
case "$CHANNEL" in
	stable)
		if [[ "$VERSION" == *-preview.* ]]; then
			echo "stable packages cannot use a preview version: $VERSION" >&2
			exit 1
		fi
		;;
	preview)
		if [[ "$VERSION" != *-preview.* ]]; then
			echo "preview packages require a -preview.N version: $VERSION" >&2
			exit 1
		fi
		;;
	*)
		echo "unsupported release channel: $CHANNEL" >&2
		exit 1
		;;
esac
if [[ -z "$PUBLIC_KEY" ]]; then
	echo "BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY is required for an updater-capable package" >&2
	exit 1
fi
if [[ -z "$EXPECTED_PUBLIC_KEY" || "$PUBLIC_KEY" != "$EXPECTED_PUBLIC_KEY" ]]; then
	echo "configured update public key does not match $ROOT_DIR/docs/engineering/UPDATE-PUBLIC-KEY.txt" >&2
	exit 1
fi
if [[ "$(uname -s)" != "Darwin" ]]; then
	echo "this macOS packager requires macOS" >&2
	exit 1
fi
if [[ "$PLATFORM" != darwin/universal && "$PLATFORM" != darwin/arm64 && "$PLATFORM" != darwin/amd64 ]]; then
	echo "unsupported macOS platform: $PLATFORM" >&2
	exit 1
fi

bash "$ROOT_DIR/scripts/verify-release-source.sh" "$VERSION"

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
"${WAILS[@]}" build -clean -platform "$PLATFORM" -ldflags "-X main.desktopVersion=${VERSION} -X main.desktopChannel=${CHANNEL} -X github.com/div197/bob-gemini-free/internal/updater.BuildUpdatePublicKey=${PUBLIC_KEY}"

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

ZIP_PATH="$OUTPUT_DIR/bob-gemini-free-macos-${ARCH_LABEL}.zip"
DMG_PATH="$OUTPUT_DIR/bob-gemini-free-macos-${ARCH_LABEL}.dmg"
ditto -c -k --norsrc --noextattr --noqtn --keepParent "$DEST_APP" "$ZIP_PATH"

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

if [[ "$CHANNEL" == "stable" ]]; then
	CHANNEL_LABEL="stable update channel candidate"
else
	CHANNEL_LABEL="public preview update channel candidate"
fi
cat > "$OUTPUT_DIR/RELEASE-NOTICE.txt" <<NOTICE
BOB Gemini Free macOS $CHANNEL_LABEL

Version: $VERSION

This package is built from the public BOB Gemini Free repository. The macOS
bundle is ad-hoc signed for bundle integrity, but it is not signed with an
Apple Developer ID certificate and is not notarized. macOS may require the
user to approve the first launch in Privacy & Security.

The project updater uses an embedded Ed25519 public key and a detached signed
SHA256SUMS manifest. That project-level authenticity check does not create
Apple platform trust. The manifest must be signed in a separate local release
operator step before publication.

No Google session, cookie, API key, or private release key is included. User
authentication, provider availability, quotas, model access, and network
behavior remain account- and upstream-dependent.

The native updater can perform a low-frequency metadata check while the app is
running, but every download, replacement, and restart remains explicit and
user-consented. Existing builds must be installed in a writable application
directory before an update can be staged.
NOTICE

(
	cd "$OUTPUT_DIR"
	shasum -a 256 "$(basename "$ZIP_PATH")" "$(basename "$DMG_PATH")" RELEASE-NOTICE.txt > SHA256SUMS
)

echo "BOB Gemini Free macOS $VERSION package candidate ready in: $OUTPUT_DIR"
ls -lh "$OUTPUT_DIR"
