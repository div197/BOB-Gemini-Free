#!/usr/bin/env bash
set -euo pipefail

# Build a free Linux preview package on a native Linux host. Wails v2 does not
# support producing a tested Linux GUI binary from macOS, so this script keeps
# the host/toolchain boundary explicit.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="${1:-/tmp/bob-gemini-free-wails-linux-preview}"
PLATFORM="${BOB_WAILS_PLATFORM:-linux/amd64}"
APP_NAME="bob-gemini-free-wails"
STAGE_DIR="$(mktemp -d "${TMPDIR:-/tmp}/bob-gemini-free-wails-linux-stage.XXXXXX")"
trap 'rm -rf "$STAGE_DIR"' EXIT

if [[ "$(uname -s)" != "Linux" ]]; then
	echo "Linux Wails preview packaging must run on a native Linux host" >&2
	exit 1
fi
if [[ "$PLATFORM" != linux/amd64 && "$PLATFORM" != linux/arm64 ]]; then
	echo "unsupported Linux Wails platform: $PLATFORM" >&2
	exit 1
fi
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

cd "$ROOT_DIR/cmd/desktop"
"${WAILS[@]}" build -clean -platform "$PLATFORM" -tags webkit2_41

SOURCE_BINARY="$ROOT_DIR/cmd/desktop/build/bin/${APP_NAME}-${PLATFORM//\//-}"
if [[ ! -f "$SOURCE_BINARY" ]]; then
	echo "Wails did not produce the expected Linux binary: $SOURCE_BINARY" >&2
	exit 1
fi

PACKAGE_DIR="$STAGE_DIR/$APP_NAME"
mkdir -p "$PACKAGE_DIR"
cp "$SOURCE_BINARY" "$PACKAGE_DIR/$APP_NAME"
chmod 0755 "$PACKAGE_DIR/$APP_NAME"
cat > "$PACKAGE_DIR/RELEASE-NOTICE.txt" <<'NOTICE'
BOB Gemini Free Linux preview package

This build is open-source software for controlled evaluation.
It is a tar.gz package, not an AppImage, and it requires GTK3 and WebKit2GTK
4.1 runtime libraries from the host distribution.
Do not use this artifact as proof of production Linux distribution readiness.
No Google session, cookie, API key, or private release key is included.
NOTICE

ARCHIVE="$OUTPUT_DIR/${APP_NAME}-linux-${PLATFORM##*/}.tar.gz"
tar -czf "$ARCHIVE" -C "$STAGE_DIR" "$APP_NAME"
cp "$PACKAGE_DIR/RELEASE-NOTICE.txt" "$OUTPUT_DIR/RELEASE-NOTICE.txt"
(
	cd "$OUTPUT_DIR"
	shasum -a 256 "$(basename "$ARCHIVE")" RELEASE-NOTICE.txt > SHA256SUMS
)

echo "Linux preview package ready in: $OUTPUT_DIR"
ls -lh "$OUTPUT_DIR"
