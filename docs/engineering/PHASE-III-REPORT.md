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
compatibility claim was manufactured, and no GitHub Actions workflow is
required. The original 2026-08-21 report snapshot predates the controlled
native preview publication documented at the end of this report.

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

## Publication follow-up — 2026-08-21

- The Wails-only archival branch was pushed to GitHub as
  `phase-iii/release-desktop-docs-hardening`.
- Pull request [#2](https://github.com/div197/BOB-Gemini-Free/pull/2) was
  promoted from draft and merged into `main` as `5ccbebe`.
- No GitHub Actions workflow was used or restored. The visible Cloudflare Pages
  check was green at merge time.
- No `v0.1.7` binary release was published. The local release script correctly
  refused to run because `BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY` and
  `BOB_GEMINI_FREE_UPDATE_PRIVATE_KEY` were not configured. Unsigned binaries
  were deliberately not uploaded because the updater is fail-closed.

## Security and public-repository follow-up — 2026-08-21

- Go's official `govulncheck` initially found one reachable `golang.org/x/image`
  vulnerability and standard-library advisories fixed by Go 1.26.6.
- Updated the module floor to Go 1.26.6, upgraded `golang.org/x/image` to
  v0.45.0, refreshed the required `x/sys` and `x/text` modules, and reran the
  scan. The result was zero reachable vulnerabilities.
- Added a public `SECURITY.md` responsible-disclosure policy and a focused
  `CONTRIBUTING.md` with the hermetic/no-Actions validation contract.
- Enabled GitHub secret scanning and push protection, and protected `main`
  against direct changes, force-pushes, and deletion while retaining the
  pull-request workflow.

## Student desktop distribution follow-up — 2026-08-22

- The Wails desktop entrypoint now loads the current user's optional config and
  cookie files while forcing loopback binding, disabling desktop API keys, and
  disabling remote origins.
- Added explicit Wails product metadata and native-host build targets for
  macOS, Windows NSIS/WebView2, and Linux WebKitGTK.
- Built and locally verified a macOS universal (`arm64` + `x86_64`) ad-hoc
  Wails bundle. Local ZIP/DMG preview packaging passed archive checks; these
  artifacts were later published only as an explicitly labelled beta, not as
  Developer ID signed or notarized production software.
- Added `docs/engineering/STUDENT-DISTRIBUTION.md` and corrected the README
  boundary: the latest stable release contains CLI binaries, while the
  separate native prerelease is a limited macOS/Windows beta.
- Windows and Linux GUI artifacts remain unbuilt and unverified on their
  native hosts. The Wails app still does not expose an in-app first-run Google
  sign-in wizard; authenticated features require each student's own session.

## Public preview publication follow-up — 2026-08-22

- PR [#8](https://github.com/div197/BOB-Gemini-Free/pull/8) merged the manual
  preview packaging and documentation contract into `main` as `312e7be`.
- Annotated tag `v0.1.7-preview.1` was pushed to that merged commit.
- The public [v0.1.7-preview.1 GitHub prerelease](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.1.7-preview.1)
  was published manually with no GitHub Actions workflow.
- The release contains the macOS universal `.dmg` and `.zip`, Windows x64
  raw `.exe`, `RELEASE-NOTICE.txt`, and `SHA256SUMS` assets. A fresh download
  of every asset from GitHub passed the combined checksum manifest and ZIP
  integrity check.
- The prerelease is deliberately not marked Latest; stable `v0.1.5` remains
  the latest stable CLI release. The native desktop updater continues to use
  the stable `/releases/latest` endpoint and does not silently install beta
  packages.
- This publication does not close the external trust gates: macOS Developer
  ID/notarization, Windows publisher signing, Linux build and acceptance,
  signed desktop manifests, clean-device testing, and complete per-user
  first-run Google sign-in remain future work.

## Preview 2 correction follow-up — 2026-08-22

- Real-device evaluation of Preview 1 reproduced a frontend-only lifecycle
  defect: generation content arrived, but the red `STOP` control remained
  active after the stream because cleanup references declared inside `try`
  were unavailable to `finally`, which raised a JavaScript `ReferenceError`.
- PR [#10](https://github.com/div197/BOB-Gemini-Free/pull/10) merged the
  surgical correction into `main` as `1a75ba4`. The patch also flushes decoder
  tail bytes, reports incomplete streams, preserves cancellation/read errors,
  and clears read timers on every failure path. It does not modify Gemini
  payload construction, authentication, cookie routing, or upstream parsing.
- `internal/server/playground_test.go` now locks the cleanup scope and stream
  lifecycle markers. The targeted server/updater suite, full normal tests,
  race tests, `go vet`, host build, `git diff --check`, and `govulncheck` passed
  on the release host.
- Annotated tag `v0.1.7-preview.2` was pushed to `1a75ba4` and published
  manually as a no-GitHub-Actions prerelease at the
  [Preview 2 release page](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.1.7-preview.2).
- The public macOS universal DMG/ZIP and Windows x64 executable were
  re-downloaded from GitHub; all four published assets passed the combined
  `SHA256SUMS` check, and the public ZIP passed archive integrity testing.
  The exact release Mac bundle passed strict ad-hoc code-signature and DMG
  verification, reached `/healthz` on its configured loopback port, and
  released its gateway cleanly after quit.
- Preview 2 remains a controlled beta. It is not Apple Developer ID signed,
  notarized, Windows publisher-signed, Linux-supported, or automatically
  updating. Automatic native updates remain a separate signed-manifest,
  platform-helper, rollback, and clean-device acceptance project.

## Native updater engineering follow-up — 2026-08-22

- Added the native source path for fixed stable-channel discovery, exact native
  asset selection, embedded Ed25519 manifest verification, positive size
  checks, safe macOS archive extraction, platform binary checks, same-volume
  staging, post-exit helper replacement, health confirmation, rollback, and
  local failure/warning records.
- Added regression coverage for signed staging, mutable environment-key
  rejection, declared-size mismatch, archive traversal, symlink/indirect
  release assets, transactional commit, and failed-candidate rollback. The
  tests use temporary targets and never replace the developer executable.
- Added `scripts/sign-release-assets.sh` for the manual no-Actions operator
  path. It refuses an existing manifest, signs the exact inspected directory,
  and verifies the resulting checksums.
- Reconciled README, Hindi README, quickstart, changelog, desktop guide,
  release process, and student-distribution wording: the current Preview 3 is
  manual-update-only; no BOB signup, cloud chat service, shared student
  cookie, or external Go-gateway telemetry is introduced.
- The source implementation is not a claim that a public release is ready
  for 30 Macs. The remaining gates are release-key custody, Developer ID/
  notarization, Windows publisher signing, native-host acceptance, clean-device
  update/rollback, and per-user Google-session onboarding.

## Preview 3 branded publication — 2026-08-22

- PR [#12](https://github.com/div197/BOB-Gemini-Free/pull/12) merged the
  verified native updater and branded package implementation into `main`;
  PR [#13](https://github.com/div197/BOB-Gemini-Free/pull/13) reconciled the
  current student-facing release documentation.
- The manually published [v0.1.7-preview.3 release](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.1.7-preview.3)
  contains `bob-gemini-free-macos-universal.dmg`,
  `bob-gemini-free-macos-universal.zip`,
  `bob-gemini-free-windows-amd64.exe`, `RELEASE-NOTICE.txt`, and
  `SHA256SUMS`.
- The exact GitHub assets were downloaded again after publication; all
  checksum entries passed, the ZIP passed archive testing, and the Windows
  asset reported as a PE32+ GUI executable.
- Preview 3's Mac bundle passed strict ad-hoc code-signature verification and
  Computer Use smoke testing. The visible app title, header, About panel,
  bundle name, executable name, and `com.abcsteps.bob-gemini-free` identifier
  are branded BOB Gemini Free; it reached its actual loopback endpoint and
  released its gateway on quit.
- Preview 3 is still a genuine beta manual release, not a signed automatic
  update channel: no Ed25519 desktop manifest, Apple Developer ID/notarization,
  Windows publisher signature, Linux artifact, or 30-device acceptance is
  claimed.
