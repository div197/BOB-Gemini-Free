# v0.2.0-preview.8 — Candidate Verification

**Date:** 2026-08-31 (Asia/Kolkata)
**Status:** locally packaged, signed, and verified; **not published**
**Source snapshot:** `309b512a45fc17bb10de712cd110ee9bd809329b` (clean reviewed
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
| Local packaged startup | PASS; incompatible process on `127.0.0.1:8081` was not reused; candidate selected `127.0.0.1:64778` |
| Local packaged health | PASS; HTTP 200, `X-Bob-Version: v0.2.0-preview.8`, `X-Bob-Auth-Required: false` |
| Local packaged static routes | PASS; `/playground`, `/manifest.json`, `/sw.js`, and `/favicon.ico` returned 200 |
| Installed Preview 6 → Preview 8 replacement | OPEN; requires a writable device |
| Public upload, fresh download, signature verification, and byte comparison | OPEN; requires explicit publication |

## Signed asset hashes

The exact local publication directory was
`/tmp/bob-gemini-free-preview-20260831`:

| Asset | SHA-256 |
|---|---|
| `RELEASE-NOTICE.txt` | `dee95f9374865479c343d76e210710b060b150dd8892556e1561fae068c075db` |
| `bob-gemini-free-macos-universal.dmg` | `edf77a9c27039f70a67875d3959e237a0d59f677b18b5465bc5a388da8352346` |
| `bob-gemini-free-macos-universal.zip` | `a8682b8c18b0898485f4224786315acce7110c3f7287776cd90b9b7b1c15dd9c` |
| `SHA256SUMS` | `08f41eb6b2dcbac85dcfbfff7299104e8b6d878d9a75287c5b96ab6672adb33f` |
| `SHA256SUMS.sig` | `90ef8fe1ae28a1376dc25d76e04edb16628609262d8031fbb9105a987419aa03` |

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
