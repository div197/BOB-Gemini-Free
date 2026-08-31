# Preview 6 — Local and Public Verification Receipt

**Date:** 2026-08-31 (Asia/Kolkata)
**Candidate:** `v0.2.0-preview.6` / preview channel
**Public source:** `f9b3410e74d7ccc08487dc03788b54a201e12ade`
**Merged change:** PR #83, version-aware desktop gateway coexistence; PR #84,
release-state documentation and updater-matrix reconciliation; PR #86,
fail-closed gateway access-key transport guard
**Status:** published prerelease; local and post-download public verification passed

This receipt records the exact local candidate produced from the public source
merge and the independent fresh download of the published assets. It does not
turn an ad-hoc macOS package into an Apple Developer ID/notarized release, and
it does not prove Google availability or 30-device rollout.

## Source and package gates

- `bash scripts/verify-release-source.sh v0.2.0-preview.6` — PASS
- `make web` — PASS; generated `web/index.html` remained synchronized
- `go test -count=1 ./...` — PASS
- `go test -race -count=1 ./...` — PASS
- `go vet ./...` — PASS
- `go mod verify` — PASS
- `go build -o /tmp/bob-gemini-free-v02-preview6-gate-current .` — PASS
- `go run ./cmd/benchmark-local -profiles 1,10,20,30 -requests 100` — PASS;
  all 400 mocked requests succeeded on the current source
- `TestPublishedPreviewFleetMatrixSelectsPreview6Candidate` — PASS; the mocked
  preview listing offers Preview 6 to legacy Preview 7, Preview 1, and Preview
  5 clients, and offers no update to Preview 6 itself
- `gh release create v0.2.0-preview.6` — PASS; manual publication with exactly
  five assets and no GitHub Actions

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
`/tmp/bob-gemini-free-preview6-main-f9b3410`

Public download verification directory:
`/tmp/bob-gemini-free-preview6-public.tqs8E9`

| Asset | SHA-256 |
|---|---|
| `RELEASE-NOTICE.txt` | `b8e6a1e686da5c51e9d82852c4aa91b88e996e78ac3166e3d882b11c0fa5bfef` |
| `SHA256SUMS` | `37c23c2691c706c7b90a92fca85285577afe47856b2d9f4bd6a8f7f7b749496a` |
| `SHA256SUMS.sig` | `c27a522a0585183141348268f6e6f8998b125a604dace1ae950531f5b10356ac` |
| `bob-gemini-free-macos-universal.dmg` | `917e4e42a4026678effff3c1d3e6a0dc964420e35b39f89f38e6fc6f9d3fbe72` |
| `bob-gemini-free-macos-universal.zip` | `581e5f12700efe461f21c8b81cd1307d08bd74b8f2b00d8ee7cb31079b76351b` |

## Native coexistence proof

The older installed `/Applications/BOB Gemini Free.app` process was left
running on `127.0.0.1:8081`. The exact current-main Preview 6 candidate was
opened separately. Because the old gateway had no matching `X-BOB-Version`
marker, the new app did not reuse it. It selected the safe loopback port
`127.0.0.1:53065` in this run, returned local `/healthz` JSON
`{"status":"ok"}` with `X-BOB-Version: v0.2.0-preview.6`, rendered
`v0.2.0-preview.6` in the native root page, and shut down cleanly. The old
process remained on 8081 throughout this check.

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
Google AI Studio key/limits/model links, the current-main access-key transport
guard, and no horizontal overflow at desktop/tablet/phone widths. Unsafe
non-loopback HTTP withheld the BOB access key and performed no ping; loopback
HTTP and HTTPS remained eligible. The earlier browser matrix remains the
broader responsive evidence; 200% text zoom, long live streams, CDN failure,
and assistive-technology coverage remain separate gates.

## Publication boundary

The immutable public release is
[`v0.2.0-preview.6`](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.6).
All five uploaded assets were downloaded into a fresh directory, passed the
detached-signature and exact-byte verifier, and matched the local signed input.
This proves project release authenticity and publication integrity; it does not
prove Apple Developer ID/notarization, clean-device rollback, live Google
availability, or a 20–30-device rollout. No GitHub Actions are required.
