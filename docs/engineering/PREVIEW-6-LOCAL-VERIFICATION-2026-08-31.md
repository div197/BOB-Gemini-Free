# Preview 6 Candidate — Local Verification Receipt

**Date:** 2026-08-31 (Asia/Kolkata)
**Candidate:** `v0.2.0-preview.6` / preview channel
**Public source:** `0cc81b2029d5dd467f7c96b26a8b812bee1ab461`
**Merged change:** PR #83, version-aware desktop gateway coexistence; PR #84,
release-state documentation and updater-matrix reconciliation
**Status:** locally verified candidate; not uploaded or published

This receipt records the exact local candidate produced after the public
source merge. It does not turn an ad-hoc macOS package into an Apple
Developer ID/notarized release, and it does not prove Google availability or
30-device rollout.

## Source and package gates

- `bash scripts/verify-release-source.sh v0.2.0-preview.6` — PASS
- `make web` — PASS; generated `web/index.html` remained synchronized
- `go test -count=1 ./...` — PASS
- `go test -race -count=1 ./...` — PASS
- `go vet ./...` — PASS
- `go mod verify` — PASS
- `go build -o /tmp/bob-gemini-free-v02-preview6-gate .` — PASS
- `go run ./cmd/benchmark-local -profiles 1,10,20,30 -requests 100` — PASS;
  all 400 mocked requests succeeded
- `TestPublishedPreviewFleetMatrixSelectsPreview6Candidate` — PASS; the mocked
  preview listing offers Preview 6 to legacy Preview 7, Preview 1, and Preview
  5 clients, and offers no update to Preview 6 itself

The benchmark is local-only and uses a deterministic in-process upstream. It
is not a Google quota, latency, or classroom-capacity claim.

## Package and trust checks

The candidate was built by `scripts/package-wails-preview.sh` for
`darwin/universal` with Wails `v2.15.0`, then signed through the owner-
controlled macOS Keychain using the existing project key. The private key was
not exported, printed, committed, or placed in the package.

- bundle name: `BOB Gemini Free`
- bundle identifier: `com.abcsteps.bob-gemini-free`
- executable: universal `x86_64` + `arm64`
- project signature: detached Ed25519 signature over `SHA256SUMS`
- macOS bundle signature: ad-hoc only; no Team ID and no notarization
- DMG layout: branded app plus `/Applications` alias — PASS
- release asset verifier: `scripts/verify-release-assets.sh` — PASS

Local candidate directory:
`/tmp/bob-gemini-free-preview6-main-0cc81b2`

| Asset | SHA-256 |
|---|---|
| `RELEASE-NOTICE.txt` | `b8e6a1e686da5c51e9d82852c4aa91b88e996e78ac3166e3d882b11c0fa5bfef` |
| `SHA256SUMS` | `cfd45fdd9e4c7875bd09b62d00655278164134c2562062cd6099dd5a5b752e36` |
| `SHA256SUMS.sig` | `12a46631d60b0e5eef0305a259e4efd992c6beb2c07dd9a9fdc2e4a2fcfce025` |
| `bob-gemini-free-macos-universal.dmg` | `7eaacc9cc451b45432f12bbab83eb0916ac3d89c0473f611c5848e3663e99b04` |
| `bob-gemini-free-macos-universal.zip` | `7b156b787bc20d0670f0ba42c240f822fa8834654cd2ce1e2cb6790168fb0643` |

## Native coexistence proof

The older installed `/Applications/BOB Gemini Free.app` process was left
running on `127.0.0.1:8081`. The exact merged-main Preview 6 candidate was
opened separately.
Because the old gateway had no matching `X-BOB-Version` marker, the new app
did not reuse it. It selected `127.0.0.1:63768`, returned local `/healthz`
JSON `{"status":"ok"}` with `X-BOB-Version: v0.2.0-preview.6`, rendered
`v0.2.0-preview.6` in the native footer, and passed the macOS zoom/maximize
action. The old process remained on 8081 throughout this check.

This closes the reproduced stale-gateway attachment defect for the tested
source/package path. Same-version reuse remains intentional; a different,
missing, unauthenticated, or non-BOB health identity falls back to a safe
loopback port.

## Settings and browser smoke

The current source loaded at a local loopback endpoint and showed:

- BOB Gateway Access Key: optional page-memory credential for an operator's
  gateway door;
- Google Gemini Developer API key: separate page-memory credential, sent only
  when its explicit route is enabled;
- web-session cookies: engine-owned and never entered into either field; and
- gateway endpoint URL: the process destination, not a credential.

The smoke also confirmed the first-use guidance, English route-status card,
Google AI Studio key/limits/model links, and no horizontal overflow at the
tested desktop viewport. The earlier desktop/tablet/phone browser matrix
remains the broader responsive evidence; 200% text zoom, long live streams,
CDN failure, and assistive-technology coverage remain separate gates.

## Publication boundary

The candidate is intentionally not a public release in this receipt. Before
publishing it, re-run the clean-worktree source gate, sign the exact final
asset directory, upload all five assets manually, download them into a fresh
directory, verify the detached signature and byte hashes, and update the
release page and rollout matrix. No GitHub Actions are required.
