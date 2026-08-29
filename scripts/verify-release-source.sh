#!/usr/bin/env bash
set -euo pipefail

# Verify source state before a release packager copies or signs it. This
# performs no generation and no network access: a release must be reproducible
# from an already-reviewed clean commit. A preview may use the Makefile's
# stable base version with a numeric -preview.N suffix.

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
printf 'release source verified: %s (%s)\n' "$MAKE_VERSION" "$HEAD"
