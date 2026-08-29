#!/usr/bin/env bash
set -euo pipefail

# Record a local, non-secret release receipt after the exact signed release
# directory has passed verification. The receipt is deliberately written
# outside the asset directory so it cannot be added after signing without
# invalidating the one-to-one manifest check.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RELEASE_DIR_INPUT="${1:-}"
VERSION="${2:-}"
OUTPUT_INPUT="${3:-}"

if [[ -z "$RELEASE_DIR_INPUT" || -z "$VERSION" ]]; then
	echo "usage: BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY=... $0 RELEASE_DIR VERSION [EVIDENCE_FILE]" >&2
	exit 1
fi
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-preview\.[0-9]+)?$ ]]; then
	echo "unsupported release version: $VERSION" >&2
	exit 1
fi
if [[ ! -d "$RELEASE_DIR_INPUT" ]]; then
	echo "release directory does not exist: $RELEASE_DIR_INPUT" >&2
	exit 1
fi

RELEASE_DIR="$(cd "$RELEASE_DIR_INPUT" && pwd)"
if [[ -n "$OUTPUT_INPUT" ]]; then
	OUTPUT_PARENT_INPUT="$(dirname "$OUTPUT_INPUT")"
	if [[ ! -d "$OUTPUT_PARENT_INPUT" ]]; then
		echo "evidence output parent must already exist: $OUTPUT_PARENT_INPUT" >&2
		exit 1
	fi
	OUTPUT_PARENT="$(cd "$OUTPUT_PARENT_INPUT" && pwd)"
	OUTPUT_FILE="$OUTPUT_PARENT/$(basename "$OUTPUT_INPUT")"
else
	OUTPUT_FILE="${TMPDIR:-/tmp}/$(basename "$RELEASE_DIR").release-evidence.txt"
fi
case "$OUTPUT_FILE" in
	"$RELEASE_DIR"/*|"$RELEASE_DIR")
		echo "evidence must be outside the signed release directory" >&2
		exit 1
		;;
	"$ROOT_DIR"/*|"$ROOT_DIR")
		echo "evidence must be outside the Git worktree so it cannot dirty release source" >&2
		exit 1
		;;
esac
if [[ -e "$OUTPUT_FILE" ]]; then
	echo "evidence file already exists; refusing to overwrite: $OUTPUT_FILE" >&2
	exit 1
fi

PUBLIC_KEY="${BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY:-}"
if [[ -z "$PUBLIC_KEY" ]]; then
	echo "BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY is required" >&2
	exit 1
fi

# Re-run the detached signature and exact asset reconciliation immediately
# before recording the receipt. This also catches a release directory changed
# after the separate signing step.
bash "$ROOT_DIR/scripts/verify-release-assets.sh" "$RELEASE_DIR"
bash "$ROOT_DIR/scripts/verify-release-source.sh" "$VERSION"

cd "$ROOT_DIR"
HEAD="$(git rev-parse HEAD)"
BRANCH="$(git branch --show-current)"
GO_VERSION="$(go version 2>/dev/null || echo unavailable)"
HOST="$(uname -srm)"
MAKE_VERSION="$(awk -F= '$1 == "VERSION" { gsub(/[[:space:]]/, "", $2); print $2; exit }' Makefile)"

sha256_file() {
	if command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$1" | awk '{print $1}'
	elif command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | awk '{print $1}'
	else
		echo "neither shasum nor sha256sum is available" >&2
		return 1
	fi
}

sha256_stdin() {
	if command -v shasum >/dev/null 2>&1; then
		shasum -a 256 | awk '{print $1}'
	elif command -v sha256sum >/dev/null 2>&1; then
		sha256sum | awk '{print $1}'
	else
		echo "neither shasum nor sha256sum is available" >&2
		return 1
	fi
}

MANIFEST_SHA256="$(sha256_file "$RELEASE_DIR/SHA256SUMS")"
SIGNATURE_SHA256="$(sha256_file "$RELEASE_DIR/SHA256SUMS.sig")"

OUTPUT_PARENT="$(dirname "$OUTPUT_FILE")"
mkdir -p "$OUTPUT_PARENT"
TEMP_FILE="$(mktemp "$OUTPUT_PARENT/.bob-release-evidence.XXXXXX")"
trap 'rm -f "$TEMP_FILE"' EXIT

{
	echo "BOB Gemini Free release evidence"
	echo "version=$VERSION"
	echo "makefile_version=$MAKE_VERSION"
	echo "git_commit=$HEAD"
	echo "git_branch=$BRANCH"
	echo "go_version=$GO_VERSION"
	echo "host=$HOST"
	echo "operator_date_utc=$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
	echo "public_key_sha256=$(printf '%s' "$PUBLIC_KEY" | tr -d '[:space:]' | sha256_stdin)"
	echo "manifest_sha256=$MANIFEST_SHA256"
	echo "signature_sha256=$SIGNATURE_SHA256"
	echo "assets="
	for path in "$RELEASE_DIR"/*; do
		[[ -f "$path" && ! -L "$path" ]] || continue
		name="$(basename "$path")"
		[[ "$name" != "SHA256SUMS" && "$name" != "SHA256SUMS.sig" ]] || continue
		digest="$(sha256_file "$path")"
		size="$(wc -c < "$path" | tr -d '[:space:]')"
		echo "  $name|$size|$digest"
	done
} > "$TEMP_FILE"

chmod 0600 "$TEMP_FILE"
mv "$TEMP_FILE" "$OUTPUT_FILE"
trap - EXIT
echo "local release evidence recorded: $OUTPUT_FILE"
