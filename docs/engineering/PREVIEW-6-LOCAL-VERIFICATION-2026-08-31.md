# Preview 6 Candidate — Local Verification Receipt

**Date:** 2026-08-31 (Asia/Kolkata)
**Candidate:** `v0.2.0-preview.6` / preview channel
**Public source:** `9f11eef922e09110df923205eb9aad90da35e236`
**Merged change:** PR #83, version-aware desktop gateway coexistence
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
`/tmp/bob-gemini-free-preview6-gateway-handshake`

| Asset | SHA-256 |
|---|---|
| `RELEASE-NOTICE.txt` | `b8e6a1e686da5c51e9d82852c4aa91b88e996e78ac3166e3d882b11c0fa5bfef` |
| `SHA256SUMS` | `c453a05be343d7ef9c1a650c6bc84cfe17030cdec33b65671097318645ff0e7e` |
| `SHA256SUMS.sig` | `ad1bed1b79d642f46fd6370e023e8d8119b283fa81bfc55e69d7708cc9114e24` |
| `bob-gemini-free-macos-universal.dmg` | `f02e9a7f80e2934bcedd532ec1da6eaf07ddf2b1905242e8e87ebc5a6dfe19de` |
| `bob-gemini-free-macos-universal.zip` | `f7c61ef334d5ae2975727aa824a5d44ed101ffadecd4e398eac4bd3aacc72523` |

## Native coexistence proof

The older installed `/Applications/BOB Gemini Free.app` process was left
running on `127.0.0.1:8081`. The Preview 6 candidate was opened separately.
Because the old gateway had no matching `X-BOB-Version` marker, the new app
did not reuse it. It selected `127.0.0.1:60864`, returned local `/healthz`
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
