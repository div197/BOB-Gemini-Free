# v0.2.0-preview.8 — Candidate Verification

**Date:** 2026-09-01 (Asia/Kolkata; receipt refreshed from the final clean tip)
**Status:** locally packaged, signed, and verified; **not published**
**Source snapshot:** `240b57b72c8575b0f14204b052f40e15385a277c` (clean reviewed
checkpoint; the package was built from this commit)
**Public baseline:** `v0.2.0-preview.6` remains the current downloadable
macOS preview. The earlier local Preview 7 candidate was never published and
is superseded by this source.

This receipt separates a locally verified candidate from a downloadable public
release. Preview 8 must not be described as downloadable until the exact
package, manifest, signature, and public-byte reconciliation are completed.

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
| `go test -count=1 ./...` | PASS |
| `go test -race -count=1 ./...` | PASS |
| `go vet ./...` | PASS |
| `go build ./...` | PASS |
| `go mod verify` | PASS; all modules verified |
| `scripts/verify-release-source.sh v0.2.0-preview.8` | PASS |
| `git diff --check` | PASS |
| Browser settings copy on current source | PASS; endpoint, BOB, and Google key boundaries render distinctly |
| Browser responsive smoke | PASS at 1440×900, 1024×768, and 390×844; no horizontal overflow |
| Open drawer toolbar hit test | PASS at 390×844; the New action resolves to `newChat()` while the drawer is open |
| Provider/session credentials | Not used; no Google key, cookie, or live generation request |
| GitHub Actions | Not used; no repository-authored workflow |

## Package and signing gates

The candidate was built from the clean source snapshot above with the existing
local release workflow:

```text
BOB_RELEASE_VERSION=v0.2.0-preview.8
BOB_RELEASE_CHANNEL=preview
BOB_WAILS_PLATFORM=darwin/universal
```

| Gate | Result |
|---|---|
| Wails packaging | PASS; Wails `v2.15.0`, macOS universal |
| App bundle signature | PASS; ad-hoc signature valid on disk |
| Bundle metadata | PASS; `com.abcsteps.bob-gemini-free`, numeric `0.2.0` |
| Binary architecture | PASS; `x86_64` and `arm64` |
| macOS DMG layout | PASS; app plus `/Applications` shortcut |
| Keychain-backed Ed25519 manifest signing | PASS; private value stayed in the local Keychain path |
| `scripts/verify-release-assets.sh` | PASS; exact signed directory |
| Local packaged startup | PASS; incompatible process on `127.0.0.1:8081` was not reused; candidate selected `127.0.0.1:49238` |
| Local packaged health | PASS; HTTP 200, `X-Bob-Version: v0.2.0-preview.8`, `X-Bob-Auth-Required: false` |
| Local packaged static routes | PASS; `/playground`, `/manifest.json`, `/sw.js`, and `/favicon.ico` returned 200 |
| Installed Preview 6 → Preview 8 replacement | OPEN; requires a writable device |
| Public upload, fresh download, signature verification, and byte comparison | OPEN; requires explicit publication |

## Signed asset hashes

The exact local publication directory was
`/tmp/bob-gemini-free-preview-20260901`:

| Asset | SHA-256 |
|---|---|
| `RELEASE-NOTICE.txt` | `dee95f9374865479c343d76e210710b060b150dd8892556e1561fae068c075db` |
| `bob-gemini-free-macos-universal.dmg` | `5981e07f827fc7598a2da0a13304fbdc097003be93f67e1cd6d44bf2784324f5` |
| `bob-gemini-free-macos-universal.zip` | `bb3123b115c23a98d17e3865fd5232a38aa40d2294370b294333f25259664194` |
| `SHA256SUMS` | `7bc40807e5e509269e620ee30dc2a76c773a41dd55eaa5d78f7d6eddc5f1ad04` |
| `SHA256SUMS.sig` | `dda86e12b17ef9b70ffa1f4bbea04366e8617c37693af3fe496d8274192ad928` |

The manifest contains one entry for each of the three release payloads and was
verified against the checked-in public trust anchor. The private signing
value was not displayed, exported, copied, committed, or included in the
package.

## Public and installed-base boundary

GitHub's current downloadable macOS preview remains `v0.2.0-preview.6`.
Preview 8 has not been tagged, uploaded, or published, so no installed device
can discover it yet. Publishing requires a new unique tag, the exact verified
five-file release directory, and a fresh download/signature/byte reconciliation
after upload. A local package or a passing updater matrix is not a fleet update.

## Release boundary

Preview 8 is an ad-hoc-signed, non-notarized macOS preview candidate. It does
not establish Apple Developer ID trust, Gatekeeper trust, Windows/Linux
support, live Google availability, quotas, rollback on a clean device, or
20–30-device rollout readiness. The updater remains explicit and
user-consented; it never silently pushes an update to student computers.

The private release-signing value remains outside Git, shell output, the
clipboard, and this receipt. Only the project public trust anchor may be
checked into the repository.
