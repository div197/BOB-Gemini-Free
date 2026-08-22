#!/usr/bin/env bash
set -euo pipefail

# Build a free Windows preview executable. This is a raw Wails executable, not
# an NSIS installer and not publisher-signed. It is useful for controlled
# Windows testing only; WebView2 must already be available on the device.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="${1:-/tmp/bob-gemini-free-wails-windows-preview}"
PLATFORM="${BOB_WAILS_PLATFORM:-windows/amd64}"
APP_NAME="bob-gemini-free-wails"

if [[ "$PLATFORM" != windows/amd64 && "$PLATFORM" != windows/arm64 ]]; then
	echo "unsupported Windows Wails platform: $PLATFORM" >&2
	exit 1
fi
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

cd "$ROOT_DIR/cmd/desktop"
"${WAILS[@]}" build -clean -platform "$PLATFORM"

SOURCE_EXE="$ROOT_DIR/cmd/desktop/build/bin/${APP_NAME}.exe"
DEST_EXE="$OUTPUT_DIR/${APP_NAME}-windows-${PLATFORM##*/}.exe"
if [[ ! -f "$SOURCE_EXE" ]]; then
	echo "Wails did not produce the expected Windows executable: $SOURCE_EXE" >&2
	exit 1
fi
cp "$SOURCE_EXE" "$DEST_EXE"

cat > "$OUTPUT_DIR/RELEASE-NOTICE.txt" <<'NOTICE'
BOB Gemini Free Windows preview executable

This build is open-source software for controlled evaluation.
It is NOT Authenticode/publisher signed and is NOT an NSIS installer.
Microsoft WebView2 must already be available on the Windows device.
Do not use this artifact as proof of production Windows distribution readiness.
No Google session, cookie, API key, or private release key is included.
NOTICE

(
	cd "$OUTPUT_DIR"
	shasum -a 256 "$(basename "$DEST_EXE")" RELEASE-NOTICE.txt > SHA256SUMS
)

echo "Windows preview executable ready in: $OUTPUT_DIR"
ls -lh "$OUTPUT_DIR"
