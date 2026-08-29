#!/usr/bin/env bash
set -euo pipefail

# Build the signed CLI release set without hosted CI. The output directory
# must not already exist so stale artifacts cannot accidentally be signed.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${BOB_RELEASE_VERSION:-v0.2.0}"
OUTPUT_DIR="${1:-release-assets}"
PUBLIC_KEY="${BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY:-}"
EXPECTED_PUBLIC_KEY="$(awk 'length($0)==64 && $0 !~ /[^0-9a-fA-F]/ { print; exit }' "$ROOT_DIR/docs/engineering/UPDATE-PUBLIC-KEY.txt")"

if [[ -z "$PUBLIC_KEY" ]]; then
  echo "BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY is required" >&2
  exit 1
fi
if [[ -z "$EXPECTED_PUBLIC_KEY" || "$PUBLIC_KEY" != "$EXPECTED_PUBLIC_KEY" ]]; then
  echo "configured update public key does not match $ROOT_DIR/docs/engineering/UPDATE-PUBLIC-KEY.txt" >&2
  exit 1
fi
if [[ "$OUTPUT_DIR" = "." || "$OUTPUT_DIR" = "/" || "$OUTPUT_DIR" = /* || "$OUTPUT_DIR" = *..* ]]; then
  echo "refusing unsafe output directory: $OUTPUT_DIR" >&2
  exit 1
fi
if [[ -e "$ROOT_DIR/$OUTPUT_DIR" ]]; then
  echo "output directory already exists; choose a clean path: $ROOT_DIR/$OUTPUT_DIR" >&2
  exit 1
fi

bash "$ROOT_DIR/scripts/verify-release-source.sh" "$VERSION"

cd "$ROOT_DIR"
mkdir -p "$OUTPUT_DIR"

LDFLAGS="-s -w -X main.Version=$VERSION -X github.com/div197/bob-gemini-free/internal/updater.BuildUpdatePublicKey=$PUBLIC_KEY"

build_cli() {
  local goos="$1"
  local goarch="$2"
  local output="$OUTPUT_DIR/bob-gemini-free-$goos-$goarch"
  if [[ "$goos" = "windows" ]]; then
    output="$output.exe"
  fi
  echo "building $goos/$goarch"
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -o "$output" .
}

build_cli darwin arm64
build_cli darwin amd64
build_cli linux amd64
build_cli linux arm64
build_cli windows amd64
build_cli windows arm64

scripts/sign-release-assets.sh "$OUTPUT_DIR"

echo "signed release assets: $ROOT_DIR/$OUTPUT_DIR"
