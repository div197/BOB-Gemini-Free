#!/usr/bin/env bash
set -euo pipefail

# Build a free Windows preview executable. This is a raw desktop executable,
# not an NSIS installer and not publisher-signed. It is useful for controlled
# Windows testing only; WebView2 must already be available on the device.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="${1:-/tmp/bob-gemini-free-windows-preview}"
PLATFORM="${BOB_WAILS_PLATFORM:-windows/amd64}"
INTERNAL_APP_NAME="bob-gemini-free"
VERSION="${BOB_RELEASE_VERSION:-}"
CHANNEL="${BOB_RELEASE_CHANNEL:-preview}"
EXPECTED_PUBLIC_KEY="$(awk '
	/^Encoding: hexadecimal Ed25519 public key$/ { in_key=1; next }
	in_key && /^[[:space:]]*$/ { in_key=0 }
	in_key && length($0)==64 && $0 !~ /[^0-9a-fA-F]/ { print; exit }
' "$ROOT_DIR/docs/engineering/UPDATE-PUBLIC-KEY.txt")"

if [[ -z "$VERSION" ]]; then
	echo "BOB_RELEASE_VERSION is required; refusing to guess or reuse an immutable preview tag" >&2
	exit 1
fi
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

if [[ "$PLATFORM" != windows/amd64 && "$PLATFORM" != windows/arm64 ]]; then
	echo "unsupported Windows desktop platform: $PLATFORM" >&2
	exit 1
fi

bash "$ROOT_DIR/scripts/verify-release-source.sh" "$VERSION"

if [[ -e "$OUTPUT_DIR" ]]; then
	echo "output already exists; choose a clean path: $OUTPUT_DIR" >&2
	exit 1
fi
if ! command -v go >/dev/null; then
	echo "requires Go" >&2
	exit 1
fi
if ! command -v shasum >/dev/null; then
	echo "requires shasum for the local artifact manifest" >&2
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
"${WAILS[@]}" build -clean -platform "$PLATFORM" "${WAILS_LDFLAGS[@]}"

SOURCE_EXE="$ROOT_DIR/cmd/desktop/build/bin/${INTERNAL_APP_NAME}.exe"
DEST_EXE="$OUTPUT_DIR/bob-gemini-free-windows-${PLATFORM##*/}.exe"
if [[ ! -f "$SOURCE_EXE" ]]; then
	echo "desktop build did not produce the expected Windows executable: $SOURCE_EXE" >&2
	exit 1
fi
cp "$SOURCE_EXE" "$DEST_EXE"

cat > "$OUTPUT_DIR/RELEASE-NOTICE.txt" <<'NOTICE'
BOB Gemini Free Windows open-source beta executable

This is the complete BOB Gemini Free desktop application for open-source
evaluation. It is not Authenticode/publisher signed and is not an installer;
Windows SmartScreen and WebView2 requirements may apply.
That platform-trust limitation does not change the product identity or the
fact that this is a genuine build from the public source repository.
The package is a beta and should be tested before a broad student rollout.
Microsoft WebView2 must already be available on the Windows device.
No Google session, cookie, API key, or private release key is included.
NOTICE

(
	cd "$OUTPUT_DIR"
	shasum -a 256 "$(basename "$DEST_EXE")" RELEASE-NOTICE.txt > SHA256SUMS
)

echo "BOB Gemini Free Windows preview executable ready in: $OUTPUT_DIR"
ls -lh "$OUTPUT_DIR"
