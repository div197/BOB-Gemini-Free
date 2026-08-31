# v0.2.0-preview.8 — Candidate Verification

**Date:** 2026-08-31 (Asia/Kolkata)
**Status:** source-reviewed; clean packaging and signing gates pending; **not
published**
**Source snapshot:** the clean reviewed commit that records this Preview 8
checkpoint (to be pinned before packaging)
**Public baseline:** `v0.2.0-preview.6` remains the current downloadable
macOS preview. The earlier local Preview 7 candidate was never published and
is superseded by this source.

This receipt is intentionally staged. A source change, a locally packaged
candidate, and a public release are three different evidence states. Preview 8
must not be described as downloadable until the exact package, manifest,
signature, and public-byte reconciliation are recorded below.

## Reviewed source changes

- `internal/gemini` request-flight keys now include sparse per-request payload
  overrides, using a length-delimited JSON representation; an unserializable
  override disables coalescing instead of risking a semantic collision.
- Responsive phone/tablet drawers are positioned below the primary New/model
  toolbar, and the generated `web/index.html` bundle is synchronized.
- Config copy explicitly distinguishes the endpoint URL, the optional BOB
  Gateway Access Key, the engine-owned web-session cookies, and the optional
  Google Gemini Developer API key. Route headers, in-memory handling, HTTPS
  guards, and no-fallback behavior are unchanged.

## Evidence completed before packaging

| Check | Result |
|---|---|
| `go test -count=1 ./internal/server ./internal/gemini` | PASS |
| `git diff --check` | PASS |
| Browser settings copy on current source | PASS; endpoint, BOB, and Google key boundaries render distinctly |
| Browser responsive smoke | PASS at 1440×900, 1024×768, and 390×844; no horizontal overflow |
| Open drawer toolbar hit test | PASS at 390×844; the New action resolves to `newChat()` while the drawer is open |
| Provider/session credentials | Not used; no Google key, cookie, or live generation request |
| GitHub Actions | Not used; no repository-authored workflow |

## Clean release gates

These are intentionally pending until this source checkpoint is committed and
packaged from a clean worktree:

| Gate | Result |
|---|---|
| `scripts/verify-release-source.sh v0.2.0-preview.8` | PENDING |
| Full Go tests, race tests, vet, build, and module verification | PENDING |
| macOS universal Wails package and local `/healthz` startup | PENDING |
| Bundle identity, architecture, DMG layout, and update-key embedding | PENDING |
| Keychain-backed manifest signing and exact asset verification | PENDING |
| Installed Preview 6 → Preview 8 replacement | OPEN; requires a writable device |
| Public upload, fresh download, signature verification, and byte comparison | OPEN; requires explicit publication |

## Release boundary

Preview 8 is an ad-hoc-signed, non-notarized macOS preview candidate. It does
not establish Apple Developer ID trust, Gatekeeper trust, Windows/Linux
support, live Google availability, quotas, rollback on a clean device, or
20–30-device rollout readiness. The updater remains explicit and
user-consented; it never silently pushes an update to student computers.

The private release-signing value remains outside Git, shell output, the
clipboard, and this receipt. Only the project public trust anchor may be
checked into the repository.
