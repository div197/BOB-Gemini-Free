#!/usr/bin/env bash
set -euo pipefail

# Build a free BOB Gemini Free Linux beta package on a native Linux host. The
# native toolkit does not support producing a tested Linux GUI binary from
# macOS, so this script keeps the host/toolchain boundary explicit.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="${1:-/tmp/bob-gemini-free-linux-preview}"
PLATFORM="${BOB_WAILS_PLATFORM:-linux/amd64}"
INTERNAL_APP_NAME="bob-gemini-free"
PUBLIC_APP_NAME="bob-gemini-free"
VERSION="${BOB_RELEASE_VERSION:-v0.2.0-preview.5}"
CHANNEL="${BOB_RELEASE_CHANNEL:-preview}"
EXPECTED_PUBLIC_KEY="$(awk '
	/^Encoding: hexadecimal Ed25519 public key$/ { in_key=1; next }
	in_key && /^[[:space:]]*$/ { in_key=0 }
	in_key && length($0)==64 && $0 !~ /[^0-9a-fA-F]/ { print; exit }
' "$ROOT_DIR/docs/engineering/UPDATE-PUBLIC-KEY.txt")"
STAGE_DIR="$(mktemp -d "${TMPDIR:-/tmp}/bob-gemini-free-linux-stage.XXXXXX")"
trap 'rm -rf "$STAGE_DIR"' EXIT

if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+-preview\.[0-9]+$ ]]; then
	echo "preview packages require a semantic -preview.N version: $VERSION" >&2
	exit 1
fi
if [[ "$CHANNEL" != "preview" ]]; then
	echo "preview packages require the preview update channel: $CHANNEL" >&2
	exit 1
fi
if [[ -z "${BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY:-}" ]]; then
	echo "BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY is required for updater-capable preview packages" >&2
	exit 1
fi
if [[ -z "$EXPECTED_PUBLIC_KEY" || "${BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY}" != "$EXPECTED_PUBLIC_KEY" ]]; then
	echo "configured update public key does not match $ROOT_DIR/docs/engineering/UPDATE-PUBLIC-KEY.txt" >&2
	exit 1
fi

if [[ "$(uname -s)" != "Linux" ]]; then
	echo "Linux desktop preview packaging must run on a native Linux host" >&2
	exit 1
fi
if [[ "$PLATFORM" != linux/amd64 && "$PLATFORM" != linux/arm64 ]]; then
	echo "unsupported Linux desktop platform: $PLATFORM" >&2
	exit 1
fi

bash "$ROOT_DIR/scripts/verify-release-source.sh" "$VERSION"

if [[ -e "$OUTPUT_DIR" ]]; then
	echo "output already exists; choose a clean path: $OUTPUT_DIR" >&2
	exit 1
fi
for command_name in go pkg-config tar shasum; do
	if ! command -v "$command_name" >/dev/null; then
		echo "required command not found: $command_name" >&2
		exit 1
	fi
done
if ! pkg-config --exists gtk+-3.0 webkit2gtk-4.1; then
	echo "requires GTK3 and WebKit2GTK 4.1 development packages" >&2
	exit 1
fi

mkdir -p "$OUTPUT_DIR"
if [[ -n "${WAILS_BIN:-}" ]]; then
	WAILS=("$WAILS_BIN")
else
	WAILS=(go run github.com/wailsapp/wails/v2/cmd/wails@v2.15.0)
fi

WAILS_LDFLAGS=(-ldflags "-X main.desktopVersion=${VERSION} -X main.desktopChannel=${CHANNEL}")
if [[ -n "${BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY:-}" ]]; then
	WAILS_LDFLAGS=(-ldflags "-X main.desktopVersion=${VERSION} -X main.desktopChannel=${CHANNEL} -X github.com/div197/bob-gemini-free/internal/updater.BuildUpdatePublicKey=${BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY}")
fi

cd "$ROOT_DIR/cmd/desktop"
"${WAILS[@]}" build -clean -platform "$PLATFORM" -tags webkit2_41 "${WAILS_LDFLAGS[@]}"

SOURCE_BINARY="$ROOT_DIR/cmd/desktop/build/bin/${INTERNAL_APP_NAME}-${PLATFORM//\//-}"
if [[ ! -f "$SOURCE_BINARY" ]]; then
	echo "desktop build did not produce the expected Linux binary: $SOURCE_BINARY" >&2
	exit 1
fi

PACKAGE_DIR="$STAGE_DIR/$PUBLIC_APP_NAME"
mkdir -p "$PACKAGE_DIR"
cp "$SOURCE_BINARY" "$PACKAGE_DIR/$PUBLIC_APP_NAME"
chmod 0755 "$PACKAGE_DIR/$PUBLIC_APP_NAME"
cat > "$PACKAGE_DIR/RELEASE-NOTICE.txt" <<'NOTICE'
BOB Gemini Free Linux open-source beta package

This is the complete BOB Gemini Free desktop application for open-source
evaluation. The package is a beta and should be tested before a broad student
rollout.
It is a tar.gz package, not an AppImage, and it requires GTK3 and WebKit2GTK
4.1 runtime libraries from the host distribution.
No Google session, cookie, API key, or private release key is included.
NOTICE

ARCHIVE="$OUTPUT_DIR/bob-gemini-free-linux-${PLATFORM##*/}.tar.gz"
tar -czf "$ARCHIVE" -C "$STAGE_DIR" "$PUBLIC_APP_NAME"
cp "$PACKAGE_DIR/RELEASE-NOTICE.txt" "$OUTPUT_DIR/RELEASE-NOTICE.txt"
(
	cd "$OUTPUT_DIR"
	shasum -a 256 "$(basename "$ARCHIVE")" RELEASE-NOTICE.txt > SHA256SUMS
)

echo "BOB Gemini Free Linux preview package ready in: $OUTPUT_DIR"
ls -lh "$OUTPUT_DIR"
