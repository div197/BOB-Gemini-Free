# BOB Gemini Free — Phase III Report

**Date:** 2026-08-21 (Asia/Kolkata)
**Workspace:** /Users/apple31/Documents/BOB-Gemini-Free
**Host:** macOS darwin/arm64, Go go1.26.5
**Branch:** phase-iii/release-desktop-docs-hardening

## Executive conclusion

Phase III moved the verified Phase II foundation toward a local, release-shaped
desktop product without changing the protected Gemini payload, authentication,
or stream protocol behavior. Wails is now the sole supported desktop path. The
former Tauri wrapper was built and opened during the comparison, then removed
from the active tree after no unique capability was found. Its source remains
recoverable through Git history.

The remaining boundary is intentional: no live Google session or provider
compatibility claim was manufactured, no GitHub Actions workflow is required,
and no release was pushed or published from this worktree.

## 1. Baseline before Phase III

The Phase II merge into main was 0e2e35e. That baseline already contained:

- deterministic Gemini payload/auth/stream/thinking/tool/adapter fixtures;
- localhost origin/API-key security controls and /healthz;
- signed-manifest updater verification and mocked replacement tests;
- Wails gateway port probing, identity-aware reuse, and local metrics;
- local-only benchmark profiles and reconciled Phase II claims.

The baseline still had these release/product gaps:

- desktop identity and gateway reuse needed stronger verification;
- the Wails bootstrap could miss its one-shot gateway-ready event;
- the Tauri sidecar used a fixed port and did not explicitly kill the child on
  normal app exit;
- Tauri icon files were empty and the Wails icon was not aligned with the BOB
  wordmark/theme;
- local release signing/building had no no-CI operator path;
- active README/UI/API wording still contained native/full/free/unlimited and
  SQLite-WASM claims that exceeded current evidence;
- hosted CI/release workflows would have consumed the user's GitHub Actions
  budget.

## 2. Changes completed

### Release and supply-chain path

- Added scripts/release-local.sh for six local CLI cross-builds plus signed
  SHA256SUMS/SHA256SUMS.sig generation.
- Kept release trust bound to the embedded public key and validated private/
  public key pairing in cmd/release-manifest.
- Added scripts/build-wails-local.sh, which stages the source outside macOS
  File Provider metadata, builds Wails with the pinned v2.15.0 module, copies
  the app bundle, and verifies an ad-hoc signature.
- Removed .github/workflows/ci.yml and .github/workflows/release.yml from the
  current branch so normal validation and release work are local and do not
  depend on GitHub Actions budget.

### Wails native desktop

- Replaced the stock Wails loading page with a branded BOB bootstrap screen.
- Added strict endpoint validation: only http: loopback endpoints are accepted
  before navigation to /playground.
- Added a durable GatewayURL fallback alongside gateway-ready so the WebView
  cannot remain stuck when startup emits the event before subscription.
- Added safe text-only startup error rendering and a retry action.
- Added generated-binding ignore coverage for frontend/wailsjs/go/.
- The packaged application embeds the Go gateway; the end user does not need a
  separate Go installation, server process, SQLite database, or memory service.

### Tauri archival decision

- The former wrapper was explicitly compared with Wails and device-smoke-tested
  before removal.
- It had a fixed port, sidecar lifecycle, and no compatible-gateway endpoint
  handoff; Wails provided the required capability set with fewer moving parts.
- The complete `desktop/` tree, Cargo/npm metadata, sidecar configuration, and
  wrapper icons were deleted from the active tree in a dedicated follow-up
  commit. Git history remains the archive and recovery mechanism.

### Brand and browser surface

- Added deterministic scripts/build-desktop-icons.sh using the canonical BOB
  wordmark in assets/bob-gemini-free-logo.jpg.
- Generated aligned Wails PNG, browser favicon, and server favicon outputs with
  BOB Builder gold/cyan treatment.
- Added GET /favicon.ico with a regression test and shared favicon links in the
  studio HTML.
- Updated the bootstrap/studio copy to use the current v0.1.7 identity and
  truthful local-first/provider-dependent wording.
- Removed the browser SQLite WASM studio, its CDN loader, SQL artifact runner,
  and related claims. The Go gateway has no SQLite or server memory database;
  browser localStorage remains only for explicit UI preference/history state.

### Documentation and evidence

- Added docs/engineering/LIVE-CONFORMANCE.md to separate hermetic checks from
  authorized Google/session acceptance.
- Added docs/engineering/RELEASE-PROCESS.md for local artifact creation,
  manual publication, signing, and clean-machine checks.
- Updated the verification matrix with current Git, desktop, icon, packaging,
  and no-database evidence while preserving the historical Mission 0 record.
- Reconciled English/Hindi README, Claude, Codex, Google, Cursor, login,
  multi-account, quickstart, health, and zero-runtime documentation.
- Reclassified native/full/unlimited/model-identity/tool/token claims as
  selected adapter, emulated, measured, or upstream-dependent where evidence
  requires it.

## 3. Native-device validation

### Wails

Commands and results:

~~~
go test -count=1 ./cmd/desktop                         PASS
go test -count=1 -race ./cmd/desktop                  PASS
go vet ./cmd/desktop                                   PASS
go build -o /tmp/bob-gemini-free-wails-device ./cmd/desktop PASS
scripts/build-wails-local.sh                          PASS
~~~

The staged bundle was opened with Computer Use. After startup it reached:

~~~
127.0.0.1:9610/playground
Color Theme: BOB Builder
Header: BOB ✦ GEMINI FREE
~~~

The studio rendered its model controls, starter cards, prompt input, endpoint
status, and BOB Builder theme. Closing the native window removed the process
and released the 9610 listener.

The direct in-place Wails package under the File Provider checkout hit macOS's
resource-fork/Finder-metadata signing failure. The clean staging script
completed the same package build and signature verification, so this is an
environment/build-path issue rather than an unverified GUI result.
Apple Developer signing/notarization remains an external distribution gate.

### Archived wrapper evidence

Before archival, the wrapper passed a device smoke build and opened the shared
BOB Builder studio. That evidence informed the product decision; it is not a
current release contract. The deleted implementation remains inspectable with
`git log --all -- desktop` and is deliberately excluded from current builds.

## 4. Protocol fidelity and security status

The following were deliberately left unchanged during this desktop/product
pass and remain protected by Phase II fixtures:

- Gemini sparse payload construction;
- SAPISIDHASH and cookie/session routing;
- streaming parser boundaries, retry deduplication, and thinking extraction;
- Scotty upload wire behavior;
- model mode mapping;
- OpenAI, Anthropic, and Google adapter wire shapes.

Current fidelity remains:

| Area | Current classification |
|---|---|
| OpenAI Chat/Responses selected routes | Implemented adapter; broad compatibility not certified |
| Anthropic Messages/SSE shape | Implemented adapter/emulation; not native Claude inference |
| Google v1beta routes | Google-shaped adapter over undocumented web RPC |
| Tool calling | EMULATED_RELIABLY for tested serialization/parsing; EMULATED_PARTIALLY for enforcement/streaming; native upstream calling unknown |
| Token counting | Local estimate, not authoritative provider tokenizer |
| Model identity/context/quota/free-unlimited behavior | Upstream/session-dependent |
| Local metrics | Aggregate, process-local, no automatic telemetry |

The Phase II localhost security boundary remains active: untrusted browser
origins are rejected, explicit origins are required for hosted Studio use, and
PNA is not treated as authentication. /healthz is local computation only.

## 5. Local validation still required before handoff

The final local gate for this branch is:

~~~
go test -count=1 ./...
go test -count=1 -race ./...
go test -count=1 -cover ./...
go vet ./...
go build ./...
go run ./cmd/benchmark-local -requests 100
git diff --check
~~~

The latest local-only benchmark rerun passed 100/100 requests at every
profile on Go go1.26.5 darwin/arm64. These are mocked gateway numbers, not
Google latency or quota measurements:

| Concurrency | P50 | P90 | P95 | P99 | Throughput | Allocs/request | RSS |
|---:|---:|---:|---:|---:|---:|---:|---:|
| 1 | 0.124 ms | 0.173 ms | 0.290 ms | 0.920 ms | 6,189 req/s | 331.49 | 19.7 MiB |
| 10 | 0.762 ms | 1.330 ms | 1.570 ms | 2.271 ms | 11,600 req/s | 336.88 | 21.6 MiB |
| 50 | 3.283 ms | 6.646 ms | 6.845 ms | 6.962 ms | 13,597 req/s | 361.95 | 24.1 MiB |
| 100 | 5.120 ms | 6.396 ms | 6.651 ms | 6.866 ms | 13,051 req/s | 388.57 | 27.5 MiB |

The live provider gate remains separate:

~~~
go test -count=1 -tags=live ./internal/diag
~~~

It requires an explicitly authorized Google session and is not represented as
passed by this report.

## 6. Remaining fragility and external gates

1. Google Web RPC schema acceptance, model identity, entitlements, context,
   quotas, rate limits, authenticated vision, Imagen, and live stream behavior
   remain unverified here.
2. Wails distribution still needs Developer ID signing/notarization and a
   clean-machine install/open check for each target OS.
3. The local release script creates signed artifacts but does not publish them;
   an authorized maintainer must manually upload the binary set, manifest,
   signature, and public trust anchor.
4. The deleted alternate desktop wrapper is preserved only in Git history and
   must not be reintroduced as a second release surface without a new decision
   record and capability evidence.
5. Remote Studio deployment still needs a real browser/PNA acceptance run with
   exact origin configuration and preferably an API key or pairing capability.
6. The browser studio still uses third-party CDN assets and browser-local
   preferences; the Go gateway itself sends no automatic telemetry.

## 7. Deliberately untouched files and systems

- internal/gemini/payload.go, internal/gemini/stream.go, and
  internal/gemini/auth.go were not changed in this Phase III desktop/brand pass.
- No SQLite package, server memory database, external telemetry, or new runtime
  service was introduced.
- No cookies, SAPISID values, API keys, release private keys, or user prompts
  were committed.
- No developer executable was replaced during updater checks.
- No GitHub Actions workflow, hosted CI job, release, tag, push, or pull request
  was created by this pass.

## 8. Recommended next phase

1. Run the authorized live conformance matrix and retain sanitized, account-
   labeled evidence only.
2. Perform signed/notarized Wails release packaging on clean macOS, Windows,
   and Linux hosts as supported by the release plan.
3. Add a small native GUI smoke harness for Wails startup, endpoint handoff,
   occupied-port fallback, and shutdown on each release host.
4. Decide whether hosted Studio needs explicit ephemeral pairing tokens in
   addition to exact origins and API keys.
5. Revisit native tool calling only after a captured accepted upstream request/
   response pair demonstrates measurable value.

## Success criterion

BOB Gemini Free is safer to evolve because its fragile protocol behavior remains
fixture-protected, its desktop runtime boundary is tested on the device, its
brand assets are deterministic and shared, and its claims now distinguish what
is implemented, emulated, measured, upstream-dependent, and still unknown.
