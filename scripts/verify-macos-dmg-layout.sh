#!/usr/bin/env bash
set -euo pipefail

# Verify the user-facing root of a macOS preview DMG. This intentionally checks
# the mounted image rather than only the source staging directory: hdiutil is
# responsible for preserving the Applications shortcut in the final artifact.

DMG_PATH="${1:?usage: $0 /path/to/bob-gemini-free-macos-universal.dmg}"
MOUNT_DIR="$(mktemp -d /tmp/bob-gemini-free-dmg-verify.XXXXXX)"
ATTACHED=0

cleanup() {
	if [[ "$ATTACHED" == "1" ]]; then
		hdiutil detach "$MOUNT_DIR" >/dev/null 2>&1 || hdiutil detach -force "$MOUNT_DIR" >/dev/null 2>&1 || true
	fi
	rmdir "$MOUNT_DIR" 2>/dev/null || true
}
trap cleanup EXIT

if [[ ! -f "$DMG_PATH" ]]; then
	echo "DMG does not exist: $DMG_PATH" >&2
	exit 1
fi
if ! command -v hdiutil >/dev/null; then
	echo "required command not found: hdiutil" >&2
	exit 1
fi

hdiutil attach -nobrowse -readonly -mountpoint "$MOUNT_DIR" "$DMG_PATH" >/dev/null
ATTACHED=1

APP_PATH="$MOUNT_DIR/BOB Gemini Free.app"
APPLICATIONS_LINK="$MOUNT_DIR/Applications"
if [[ ! -d "$APP_PATH" ]]; then
	echo "DMG is missing the BOB Gemini Free app bundle" >&2
	exit 1
fi
if [[ ! -L "$APPLICATIONS_LINK" || "$(readlink "$APPLICATIONS_LINK")" != "/Applications" ]]; then
	echo "DMG is missing the /Applications drag target" >&2
	exit 1
fi

hdiutil detach "$MOUNT_DIR" >/dev/null
ATTACHED=0
echo "macOS DMG layout verified: BOB Gemini Free.app + Applications -> /Applications"
