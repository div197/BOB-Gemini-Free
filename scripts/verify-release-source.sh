#!/usr/bin/env bash
set -euo pipefail

# Verify source state before a release packager copies or signs it. This
# performs no generation and no network access: a release must be reproducible
# from an already-reviewed clean commit. A preview may use the Makefile's
# stable base version with a numeric -preview.N suffix. The next preview
# candidate is a separate explicit Makefile value; packagers must never infer
# it from a previously published tag.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${1:-}"

if ! git -C "$ROOT_DIR" rev-parse --show-toplevel >/dev/null 2>&1; then
	echo "release source must be inside a Git worktree: $ROOT_DIR" >&2
	exit 1
fi

if [[ -n "$(git -C "$ROOT_DIR" status --porcelain --untracked-files=all)" ]]; then
	echo "release source is not clean; commit or remove all modified and untracked files before packaging" >&2
	git -C "$ROOT_DIR" status --short >&2
	exit 1
fi

MAKE_VERSION="$(awk -F= '$1 == "VERSION" { gsub(/[[:space:]]/, "", $2); print $2; exit }' "$ROOT_DIR/Makefile")"
if [[ -z "$MAKE_VERSION" ]]; then
	echo "Makefile VERSION is missing" >&2
	exit 1
fi
if [[ ! "$MAKE_VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
	echo "Makefile VERSION is not a stable semantic version: $MAKE_VERSION" >&2
	exit 1
fi

PREVIEW_VERSION="$(awk -F= '$1 == "PREVIEW_VERSION" { gsub(/[[:space:]]/, "", $2); print $2; exit }' "$ROOT_DIR/Makefile")"
if [[ -z "$PREVIEW_VERSION" || "$PREVIEW_VERSION" != "$MAKE_VERSION-preview."* ]]; then
	echo "Makefile PREVIEW_VERSION must use Makefile base $MAKE_VERSION: $PREVIEW_VERSION" >&2
	exit 1
fi
PREVIEW_NUMBER="${PREVIEW_VERSION#"$MAKE_VERSION-preview."}"
if [[ ! "$PREVIEW_NUMBER" =~ ^[0-9]+$ ]]; then
	echo "Makefile PREVIEW_VERSION has an invalid preview suffix: $PREVIEW_VERSION" >&2
	exit 1
fi

# The standalone installers must carry the same trust anchor as the native
# package build. Keep this check in the release-source gate so a future key
# rotation cannot update only the Go/desktop path and silently leave the
# public bootstrap scripts unable to verify releases.
PUBLIC_KEY_FILE="$ROOT_DIR/docs/engineering/UPDATE-PUBLIC-KEY.txt"
if [[ ! -f "$PUBLIC_KEY_FILE" ]]; then
	echo "update public-key file is missing: $PUBLIC_KEY_FILE" >&2
	exit 1
fi
PUBLIC_KEY_HEX="$(awk '
	/^Encoding: hexadecimal Ed25519 public key$/ { in_key=1; next }
	in_key && /^[[:space:]]*$/ { in_key=0 }
	in_key && length($0)==64 && $0 !~ /[^0-9a-fA-F]/ { count++; value=tolower($0) }
	END { if (count != 1) exit 1; print value }
' "$PUBLIC_KEY_FILE")" || {
	echo "update public-key file must contain exactly one 64-character hexadecimal key" >&2
	exit 1
}
INSTALL_PUBLIC_KEY_HEX="$(awk -F'"' '$1 == "UPDATE_PUBLIC_KEY_HEX=" { count++; value=tolower($2) } END { if (count != 1) exit 1; print value }' "$ROOT_DIR/install.sh")" || {
	echo "install.sh must contain exactly one UPDATE_PUBLIC_KEY_HEX assignment" >&2
	exit 1
}
if [[ "$INSTALL_PUBLIC_KEY_HEX" != "$PUBLIC_KEY_HEX" ]]; then
	echo "install.sh update public key does not match $PUBLIC_KEY_FILE" >&2
	exit 1
fi
INSTALL_PUBLIC_KEY_SPKI_B64="$(awk -F'"' '$1 == "UPDATE_PUBLIC_KEY_SPKI_B64=" { count++; value=$2 } END { if (count != 1) exit 1; print value }' "$ROOT_DIR/install.sh")" || {
	echo "install.sh must contain exactly one UPDATE_PUBLIC_KEY_SPKI_B64 assignment" >&2
	exit 1
}
if [[ ! "$INSTALL_PUBLIC_KEY_SPKI_B64" =~ ^[A-Za-z0-9+/]+=*$ ]]; then
	echo "install.sh UPDATE_PUBLIC_KEY_SPKI_B64 is not valid Base64" >&2
	exit 1
fi
if ! command -v base64 >/dev/null 2>&1; then
	echo "base64 is required to verify the install.sh Ed25519 SPKI" >&2
	exit 1
fi
EXPECTED_PUBLIC_KEY_SPKI_B64="$({
	printf '%b' '\x30\x2a\x30\x05\x06\x03\x2b\x65\x70\x03\x21\x00'
	for ((offset = 0; offset < ${#PUBLIC_KEY_HEX}; offset += 2)); do
		printf '%b' "\\x${PUBLIC_KEY_HEX:offset:2}"
	done
} | base64 | tr -d '[:space:]')"
if [[ "$INSTALL_PUBLIC_KEY_SPKI_B64" != "$EXPECTED_PUBLIC_KEY_SPKI_B64" ]]; then
	echo "install.sh Ed25519 SPKI does not match $PUBLIC_KEY_FILE" >&2
	exit 1
fi
WINDOWS_PUBLIC_KEY_HEX="$(awk -F'"' '$1 ~ /^\$UpdatePublicKeyHex[[:space:]]*=[[:space:]]*$/ { count++; value=tolower($2) } END { if (count != 1) exit 1; print value }' "$ROOT_DIR/install.ps1")" || {
	echo "install.ps1 must contain exactly one UpdatePublicKeyHex assignment" >&2
	exit 1
}
if [[ "$WINDOWS_PUBLIC_KEY_HEX" != "$PUBLIC_KEY_HEX" ]]; then
	echo "install.ps1 update public key does not match $PUBLIC_KEY_FILE" >&2
	exit 1
fi

DOCKER_VERSION="$(awk -F= '$1 == "ARG VERSION" { gsub(/[[:space:]]/, "", $2); print $2; exit }' "$ROOT_DIR/Dockerfile")"
if [[ -n "$DOCKER_VERSION" && "$DOCKER_VERSION" != "$MAKE_VERSION" ]]; then
	echo "Dockerfile VERSION $DOCKER_VERSION does not match Makefile VERSION $MAKE_VERSION" >&2
	exit 1
fi

for preview_script in \
	"scripts/package-wails-preview.sh" \
	"scripts/package-wails-windows-preview.sh" \
	"scripts/package-wails-linux-preview.sh"; do
	preview_path="$ROOT_DIR/$preview_script"
	if [[ ! -f "$preview_path" ]]; then
		echo "preview packager is missing: $preview_script" >&2
		exit 1
	fi
	if ! awk 'index($0, "VERSION=\"${BOB_RELEASE_VERSION:-}\"") == 1 { found=1 } END { exit !found }' "$preview_path"; then
		echo "$preview_script must require an explicit BOB_RELEASE_VERSION" >&2
		exit 1
	fi
	if ! awk 'index($0, "BOB_RELEASE_VERSION is required") { found=1 } END { exit !found }' "$preview_path"; then
		echo "$preview_script must fail closed when BOB_RELEASE_VERSION is unset" >&2
		exit 1
	fi
	preview_channel="$(awk -F':-' '/^CHANNEL=.*BOB_RELEASE_CHANNEL/ { value=$2; sub(/}.*/, "", value); gsub(/"/, "", value); print value; exit }' "$preview_path")"
	if [[ "$preview_channel" != "preview" ]]; then
		echo "$preview_script default channel must be preview" >&2
		exit 1
	fi
done

if [[ -n "$VERSION" ]]; then
	if [[ "$VERSION" == "$MAKE_VERSION" ]]; then
		:
	elif [[ "$VERSION" == "$MAKE_VERSION-preview."* ]]; then
		PREVIEW_NUMBER="${VERSION#"$MAKE_VERSION-preview."}"
		if [[ ! "$PREVIEW_NUMBER" =~ ^[0-9]+$ ]]; then
			echo "requested release version $VERSION has an invalid preview suffix" >&2
			exit 1
		fi
	else
		echo "requested release version $VERSION does not match Makefile base version $MAKE_VERSION" >&2
		exit 1
	fi
fi

if [[ ! -f "$ROOT_DIR/internal/server/playground.html" || ! -f "$ROOT_DIR/web/index.html" ]]; then
	echo "generated Web Studio bundle is missing" >&2
	exit 1
fi
if ! diff -u \
	<(sed "s/__BOB_DESKTOP_VERSION__/$MAKE_VERSION/g" "$ROOT_DIR/internal/server/playground.html") \
	"$ROOT_DIR/web/index.html" >/dev/null; then
	echo "web/index.html is not synchronized with internal/server/playground.html; run 'make web' and review the generated diff" >&2
	exit 1
fi

HEAD="$(git -C "$ROOT_DIR" rev-parse HEAD)"
CHECKED_VERSION="${VERSION:-$MAKE_VERSION}"
printf 'release source verified: %s (Makefile base %s; commit %s)\n' "$CHECKED_VERSION" "$MAKE_VERSION" "$HEAD"
