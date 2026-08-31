# BOB Gemini Free v0.2.0 Release Readiness (historical compilation)

**Audit date:** 2026-08-31 (Asia/Kolkata)
**Base HEAD before this readiness preparation:** `59a0d228ab8602427820ae90a14efe5f36f38ccd`
**Previous public fleet release:** `v0.1.7-preview.7`
**Current packaged source baseline:** public `main` at `c28d78736eaae436cc1f1f3b4ec6e0bbcd058b89`
**Current public preview:** immutable macOS universal `v0.2.0-preview.5`; Preview
4 and Preview 3 are historical, and Preview 1 remains the migration bridge
**Decision:** **NOT READY for publication as a student-facing stable release**

> **Current audit superseding this historical compilation:** the authoritative
> release state is recorded in
> [`RELEASE-AUDIT-2026-08-31.md`](RELEASE-AUDIT-2026-08-31.md). Preview 5 is
> published and its public bytes plus one writable Preview 1 → Preview 5
> migration are verified. Apple platform trust, deliberate rollback, clean
> device/pilot acceptance, provider behavior, and fleet rollout remain open.
> The private key remains only in the owner-controlled macOS Keychain.

The sections below preserve earlier readiness snapshots for provenance. Treat
their Preview 1–4 “current” wording as historical at the time each section was
written; use the current release audit and Preview 5 publication record for
present-day operations.

## Current Preview 5 publication reconciliation — 2026-08-31

The public `v0.2.0-preview.5` macOS universal prerelease was published
manually from `main` target `c28d787` after clean-source, package, Keychain
signature, and public-byte gates. A writable `/Applications` Preview 1
installation updated through **Help → Check for Updates**, restarted healthy,
and retained visible chat state. See
[`PREVIEW-5-PUBLICATION-2026-08-31.md`](PREVIEW-5-PUBLICATION-2026-08-31.md).

This is a controlled public beta, not a stable or unattended fleet release.
The remaining gates are deliberate rollback, clean-device and pilot evidence,
Apple/Windows platform trust, and live provider/network validation.

## Current source continuation — 2026-08-31

Public `main` has since advanced to `0cc81b2029d5dd467f7c96b26a8b812bee1ab461`
through protected PRs #77–#84. Those source follow-ups are not contained in the
immutable Preview 5 package. A universal `v0.2.0-preview.6` candidate was
built, signed, package-verified, and coexistence-smoke-tested locally; it has
not been uploaded or published. The exact evidence and hashes are recorded in
[`PREVIEW-6-LOCAL-VERIFICATION-2026-08-31.md`](PREVIEW-6-LOCAL-VERIFICATION-2026-08-31.md).
The new updater regression matrix covers Preview 6 discovery from legacy
`v0.1.7-preview.7`, the `v0.2.0-preview.1` bridge, and Preview 5 through a
mocked release listing.

## Preview 4 publication refresh — 2026-08-31 (historical)

The Preview 4 package source is `main` at merge commit
`abfeebaaaaabc740ea29602b602591a0b707fbc2`. Protected PR #58 merged the
explicit Google 401/403 session-cache recovery fix, and PR #59 merged the
Preview 4 packaging/versioning correction. The current controlled macOS
universal preview is published at
[`v0.2.0-preview.4`](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.4).

After that publication, protected PR #62 merged the artifact-preview and
responsive-header fix, followed by PR #64's multiline-SSE framing fix, into
`main` at `cd44b2c`. The published Preview 4 assets were not rebuilt or
overwritten; the next package must be explicitly labelled `v0.2.0-preview.5`.

The exact five assets were signed through the owner-controlled macOS Keychain,
downloaded again from the public release into a fresh directory, verified with
the checked-in Ed25519 public key, and matched the local signed input
byte-for-byte. The local package passed universal-binary inspection, ad-hoc
code-signature verification, branded DMG-layout checks, fresh launch on the
audit Mac, loopback `/healthz`, rendered Preview 4 version, occupied-port
fallback, and clean shutdown. The local full test, race, vet, module, build,
and release-source gates also passed before publication.

This closes Preview 4 source/package/public-byte publication integrity. It does
not close Apple Developer ID/notarization, clean-device replacement, rollback
after interruption, live Google acceptance, or 20–30-device rollout gates.
The private signing key remains outside GitHub and the repository; only its
matching public key is embedded in the application. No GitHub Actions workflow
was added or used.

## Preview 3 publication refresh — 2026-08-31 (historical)

The current public-main tip is merge commit `284b7d1a9a2e7c45402318f29f08f0c1dba36d43`.
Protected PRs [#53](https://github.com/div197/BOB-Gemini-Free/pull/53),
[#54](https://github.com/div197/BOB-Gemini-Free/pull/54), and
[#55](https://github.com/div197/BOB-Gemini-Free/pull/55) placed the responsive
drawer focus fix, Preview 3 versioning, and branded ZIP packaging on public
`main`. The current macOS universal preview is publicly published at
[`v0.2.0-preview.3`](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.3).

The exact five public assets were re-downloaded into a fresh directory,
verified with the checked-in public Ed25519 key, and byte-compared with the
local Keychain-signed inputs. The universal app passed ad-hoc code-signature,
DMG-layout, launch, loopback `/healthz`, and clean-shutdown checks. This
closes Preview 3 source/package/public-byte publication integrity, not Apple
platform trust, clean-device updater replacement, rollback, Google provider
availability, or 20–30-device rollout acceptance. The matching private key
remains outside the repository in the owner-controlled local secret store.

## Repository refresh before Preview 4 publication — 2026-08-31 (historical)

The historical publication entries below are retained as provenance. The
current local truth is:

- The pre-merge audit baseline was `523ceeb`; protected PR [#42](https://github.com/div197/BOB-Gemini-Free/pull/42)
  merged the reviewed source, PR [#43](https://github.com/div197/BOB-Gemini-Free/pull/43)
  reconciled the post-merge documentation, and PR [#44](https://github.com/div197/BOB-Gemini-Free/pull/44)
  merged the first updater durability follow-up into public `main`. Protected
  PR [#45](https://github.com/div197/BOB-Gemini-Free/pull/45) reconciled the
  public-main provenance note, and protected PR
  [#46](https://github.com/div197/BOB-Gemini-Free/pull/46) merged the native
  Windows metadata replacement path.
- The Preview 3 release source baseline is public-main merge commit
  `284b7d1a`; protected PR #53 merged the responsive drawer focus fix, PR #54
  established the Preview 3 versioning candidate, and PR #55 corrected the
  branded ZIP bundle name. No runtime source remains only on the preserved
  `codex/release-readiness-v0.2.0` branch, and the exact public merge commit
  was rechecked before this documentation refresh.
- The immutable public `v0.2.0-preview.1` bridge and superseded Preview 2
  were not rebuilt or overwritten. Controlled macOS Preview 3 is now
  published as `v0.2.0-preview.3` from public-main commit `284b7d1a`; its
  exact asset and signature reconciliation is recorded in
  [`PREVIEW-3-PUBLICATION-2026-08-31.md`](PREVIEW-3-PUBLICATION-2026-08-31.md).
- The merged source contains the later 100-path hardening follow-ups,
  including nil-safe server and Gemini-client optional logging, accessible
  attachment/image controls, and JavaScript-URL-free gateway recovery.
- No stable `v0.2.0` tag or release was created here, and no GitHub Actions
  workflow was added or run.
- This refresh closes signed-asset publication for the controlled macOS
  Preview 3; it does not close Apple/Windows platform trust, clean-device
  updater, live provider, browser, or 30-device rollout gates.
- The native updater now preflights the current install location before any
  release artifact download and explains App Translocation/read-only paths;
  the preflight is source- and fixture-tested but still needs a real
  `/Applications` installed-bundle run.
- The public macOS Preview 3 package was rebuilt from clean public-main commit
  `284b7d1a`; its universal bundle passed ad-hoc `codesign --verify`, the DMG
  contained a visible `/Applications` shortcut, and the exact uploaded assets
  were downloaded and reverified through the local Keychain-backed Ed25519
  release signer. A 0600 local evidence receipt was recorded outside the
  worktree. This is package/public-byte evidence, not a clean-device update
  or platform-trust proof.
- The current tip also contains the isolated Studio correctness pass: native
  button semantics, drawer `aria-hidden`/`inert` state, bounded dialog
  surfaces, accessible selector names, a prompt skip link, and synchronized
  generated web output. These are source/regression results, not visual
  browser acceptance.
- The current source also invalidates cached dynamic Gemini `/app` tokens after
  an explicit upstream HTTP 401/403. Buffered and streaming rejection,
  configured-cookie preservation, fresh-bootstrap recovery, and an in-flight
  bootstrap ordering race are regression-tested. This is local stale-token
  recovery only; live expiry, reauthentication, and provider acceptance remain
  external gates.
- Follow-up commit `b136724` hardens updater transaction durability with
  flushed metadata commits, Unix directory synchronization around swaps and
  recovery, and a fault-injected rollback regression. Recursive app-bundle
  durability, Windows directory-flush/journal evidence, physical power
  interruption, and clean-device acceptance remain open gates; `fd279aa` adds
  the native Windows metadata replacement path.

This is a release gate, not a claim that the source is unusable. The current
branch is a large post-Preview-7 source milestone, but a signed artifact,
platform trust, an update migration, and a 30-device acceptance run are four
different proofs.

## Historical readiness snapshot (before Preview 5 publication)

The remaining sections preserve the decision records and acceptance tables that
were written before Preview 5 was published. They are retained for audit
provenance and are not the current rollout instruction.

## Executive decision

The project-level release-signing mechanism is available: the matching Ed25519
private key is held outside the repository in the owner's local macOS Keychain,
and the public key is recorded in
[`UPDATE-PUBLIC-KEY.txt`](UPDATE-PUBLIC-KEY.txt). The private key was not read,
copied, printed, or added to this repository during this audit.

That is not the same as Apple Developer ID signing or notarization. A macOS
package can therefore be project-authenticated for the BOB updater and still
show the normal first-launch Gatekeeper approval on an unnotarized Mac.

The source preparation in this branch closes the main release-engineering
gaps found after Preview 7:

- stable Wails build targets require and embed the public updater key;
- Wails product metadata, Docker version metadata, and the local CLI release
  default are aligned to `v0.2.0`;
- a dedicated macOS release packager distinguishes stable and preview builds;
- newly built Preview builds first check the fixed official stable channel,
  allowing a deliberate one-way Preview → Stable migration, and only then
  check the fixed preview channel when no stable update exists;
- updater metadata and asset URLs are pinned to the exact official repository
  path or GitHub's release CDN, with regression tests for lookalike and
  unrelated GitHub URLs;
- regression tests prove stable migration, preview continuation, and the
  legacy Preview 7 bridge/no-direct-stable boundary.

The following paragraph is retained as historical provenance for the earlier
PR-based publication sequence. The current source and package truth is in the
refresh above and in
[`RELEASE-TRANSITION-AUDIT-2026-08-31.md`](RELEASE-TRANSITION-AUDIT-2026-08-31.md).
The signed `v0.2.0-preview.1` migration bridge, superseded `v0.2.0-preview.2`,
and current controlled macOS `v0.2.0-preview.3` are publicly published. Stable `v0.2.0` remains unpublished
until the clean-device and pilot gates pass.

## Evidence already available

The following checks passed on the audit host (macOS 26.2, arm64, Go 1.26.6):

```text
go test -count=1 ./...
go test -race -count=1 ./...
go test -count=1 -cover ./...
go vet ./...
go build ./...
go mod verify
make desktop-key-check
git diff --check
```

The same local readiness run also passed `make build`, the release-script
syntax checks (with `bash -n`), and `go test -count=1 -cover ./...`. A weighted
per-package profile over that run measured 63.6% statement coverage; the
package-local results range from 0.0% to 89.3%, so no blanket 80% claim is
valid. The run also passed `go mod verify`, six equivalent CGO-disabled CLI
cross-builds in an isolated
temporary directory, and the public-key presence gate. Coverage remains a
measured limitation.

At the time of this historical readiness review, the working tree was clean and the
complete post-Preview-7 branch delta was 134 files relative to the then-current
`origin/main`; the reviewed source is now published on `main` as PR #31. This
remains a release-candidate audit of a large feature milestone, not a routine
patch release. Repository-wide weighted statement coverage measured 63.6% in
the current local run; the project must not claim blanket 80% coverage. The
current local-only 1/10/20/30-concurrency baseline is recorded in
[`LOCAL-BENCHMARK-2026-08-31.md`](LOCAL-BENCHMARK-2026-08-31.md); it is not a
Google capacity or latency result.

An earlier clean commit `d318b4f` also passed a fresh `make desktop-preview-mac`
package run on this Mac. That historical package was an unsigned-manifest
candidate. The later Preview 3 candidate was signed and verified through the
Keychain-backed manifest flow as recorded in the current refresh; clean-device
updater and pilot gates remain open.
`spctl` rejection remains expected for a package without Apple Developer ID
trust. An intentionally missing Keychain item also caused the manifest signer
to fail closed; the real private key was not read.

The public GitHub state was also checked:

- the newest preview release is `v0.2.0-preview.3`, published as a controlled
  macOS preview from public-main commit `284b7d1a`;
- its macOS universal DMG/ZIP, release notice, checksum manifest, and detached
  signature are present, were re-downloaded, signature-verified, and matched
  the locally signed files byte-for-byte;
- the reviewed source hardening and release record were published to `main`
  through protected PRs #31, #42, #44, #46, #47, #48, #53, #54, and #55; the
  current Preview 3 publication baseline is `284b7d1a`;
- the release-source coherence, installer trust-anchor, and session-only
  gateway-auth follow-ups were subsequently merged through PRs #33, #36, and
  #38; that historical snapshot recorded `origin/main` as `f3a0a8c`; the
  current authoritative `origin/main` at that historical snapshot was
  `523ceeb` (PR #41, which contains those earlier merge ancestors); the later
  protected PRs #42–#48 are recorded in the current-local-truth section above
  and were rechecked before this refresh;
- there is no stable `v0.2.0` tag or GitHub Release yet;
- no GitHub Actions workflow is required or present in the current tree.

## Publication follow-up — 2026-08-29

Protected PR [#31](https://github.com/div197/BOB-Gemini-Free/pull/31) was merged
through the normal repository workflow at `c5fa74f`. The local checkout was
fast-forwarded to that commit and the reviewed branch was not force-pushed.
This closes source publication/provenance for the audited hardening work; it
does not close the separate signed-artifact, platform-trust, clean-device,
provider, or 30-device rollout gates below.

## Publication refresh — 2026-08-30

Protected PR [#36](https://github.com/div197/BOB-Gemini-Free/pull/36) merged the
installer trust-anchor encoding check into `main` at `627e73c`. The local
checkout was fast-forwarded to that merge and the post-merge source, test,
race, vet, module, build, and release-preflight gates passed. This is a
source-coherence result only; the stable release, platform trust, clean-device
update, provider, and 30-device rollout gates remain separate.

Protected PR [#38](https://github.com/div197/BOB-Gemini-Free/pull/38) then
merged the Web Studio credential-persistence fix into `main` at `f3a0a8c`.
Optional BOB gateway-auth tokens are now page-session-only, and the legacy
`bob_api_key` browser-storage entry is purged on load. This improves shared
classroom-device hygiene but does not change the requirement to re-enter a
protected gateway token after reload, nor does it close the separate signed
artifact, platform-trust, clean-device, provider, or fleet gates.

## Signing and trust gates

| Gate | Current truth | Release consequence |
|---|---|---|
| Project manifest authenticity | Ed25519 public key is in the repository; private key is local-only | Can be completed by the owner with a local signing operation |
| Exact artifact integrity | Updater verifies the signed `SHA256SUMS` entry, size, package type, and platform magic; Preview 3's five public assets were re-downloaded and byte-reconciled | Repeat the same regeneration, signature, and public-byte checks for every future release |
| macOS platform trust | No Apple Developer ID certificate, hardened-runtime notarization, or stapled ticket is available in this workflow | The result must remain clearly labelled project-signed/ad-hoc and may require first-launch approval |
| Windows publisher trust | No Windows publisher-signed installer has been accepted in this audit | Windows cannot be called production-ready from this branch |
| Release channel | Signed previews `v0.2.0-preview.1`, superseded `v0.2.0-preview.2`, historical Preview 3, and current controlled macOS `v0.2.0-preview.4` are public; stable `v0.2.0` is not | Existing same-key Preview 7 devices have a current Preview 4 candidate through the preview path; installed replacement and stable remain gated |
| Private-key custody | Keychain presence was checked without reading the secret | Keep it out of Git, chat, screenshots, student machines, and shell transcripts |

The stable Wails targets now fail closed when the public key file is absent.
The macOS package script only needs the public key; the private key belongs only
to the separate local manifest-signing step. No private key is needed in a
student package.

## Update behavior for existing installations

The updater is user-consented, not a fleet push and not a silent background
updater. The exact flow is:

```text
Help → Check for Updates
  → query fixed official release metadata
  → show the candidate
  → user chooses Install Update
  → download and verify signed manifest and exact asset
  → stage beside the installed app
  → restart through helper
  → keep rollback copy until health confirmation
```

The new source behavior is intentionally one-way:

- a stable build checks only the latest stable release;
- a preview build checks stable first and can migrate to a newer stable
  release;
- if stable has no newer release, a preview build checks the bounded preview
  listing for a newer `preview.N` release;
- a stable build never downgrades or moves into preview;
- a failed stable metadata check is not hidden by a preview result.

For the 30 devices, the result depends on the installed version and location:

| Existing device state | Can it update to a published signed `v0.2.0` stable package? | Required action |
|---|---|---|
| Already-installed public `v0.1.7-preview.7` binary, current Preview-7 key, app copied to a writable directory | It has a current same-key `v0.2.0-preview.4` candidate through its preview lookup; it still cannot jump directly to stable | Use **Help → Check for Updates**, confirm Preview 4 (or Preview 1 first if it is the displayed candidate), then use the current preview's stable-first updater after stable is published; alternatively install stable manually |
| A newly built current-source preview bridge, current key, writable directory | Yes, it can discover a newer signed Preview 4 or stable package | User performs the explicit update, then selects Check for Updates again for Preview → Stable |
| Preview 4–6 or another build with the old/unrecoverable project key | No, not cryptographically | One manual install of a package carrying the current public key, then later updates can be verified |
| Historical `v0.1.7-preview.3` or older without an embedded updater key | No | Manual installation; do not use an environment variable as a production trust substitute |
| App still running from a mounted DMG or App Translocation path | No | Copy it to `/Applications` or another writable application directory and relaunch |
| A package with no signed manifest or wrong platform asset | No | Updater must refuse it; use the official release page for recovery |

Therefore, if all 30 students truly have the public Preview 7 binary, the
published Preview 4 candidate provides the current updater step, but not a
direct one-step Preview 7 → stable migration. Pilot the explicit update steps,
and only then publish/announce stable. This still does **not** prove that all 30
machines will update: OS version, architecture, permissions, network access,
release-asset availability, and provider usage are independent gates. If a
device is on Preview 6 or earlier, it needs the manual one-time migration to
the current key before either updater path can be trusted.

## Required acceptance run before publication

### Gate A — one clean Mac

Use the exact final DMG/ZIP, not a developer build:

1. copy the app to a writable application directory;
2. approve the normal macOS first-launch warning without disabling Gatekeeper;
3. confirm the displayed version and actual loopback gateway endpoint;
4. run local `/healthz` and one bounded student-authentication smoke test if
   that capability is in scope;
5. install a controlled successor release signed by the same project key;
6. for an already-installed public Preview 7, verify the bridge preview update
   first and then verify bridge → Stable; for a newly built current preview,
   verify the direct Preview → Stable path;
7. interrupt or deliberately invalidate a candidate and confirm rollback;
8. confirm cookies, config, preferences, and chat data are not replaced.

### Gate B — two or three pilot Macs

Record only:

```text
device label | macOS version | arm64/amd64 | installed version | healthz |
generation class | update result | rollback result
```

Do not record cookies, Google account identifiers, full prompts, response
bodies, or release private-key material.

### Gate C — 30-device rollout

Proceed only after Gates A and B pass. Roll out in waves of 5–10 devices and
verify the local app version and `/healthz` after each wave. Each user still
needs to approve the update; this updater cannot remotely push to 30 Macs.
The shared school egress IP must not be disguised with rotating proxies,
fingerprint changes, or shared cookies. The updater uses GitHub's unauthenticated
release API; the audit host currently reported a 60-request hourly limit with
57 remaining. Thirty devices checking two channels at the same moment can
consume that budget, so use staggered rollout waves and treat API rate limiting
as a recoverable update-check failure. A successful install is not proof of
Google quota, model identity, unlimited access, or provider availability.

## Feature claims that must remain bounded

The large source milestone contains useful work, but the following are not
release proofs and must not be advertised as shipped capabilities without new
artifacts or measurements:

- `pkg/mobile` is an experimental Go bridge. The repository does not contain a
  native Android project, iOS project, AAR, or XCFramework; its current code
  starts a local HTTP listener and is not a zero-socket mobile app.
- `internal/refiner` runs three inference stages through the supplied inference
  function. It is not a pure-local four-stage engine, and its token/duration
  fields are estimates/serialization values rather than authoritative model
  telemetry.
- model aliases are routing aliases to the available Google web path, not proof
  that BOB is running the named OpenAI or Anthropic model.
- token counts, latency, RAM, concurrency, savings, free use, and unlimited use
  require dated measurements or remain upstream-/environment-dependent.
- StreamFlight, cookie-pool cooldown behavior, and live Google acceptance need
  additional targeted tests before they are used as 30-device capacity claims.

## Exact no-Actions release sequence

1. Keep the version/tag immutable and confirm `CHANGELOG.md`, Makefile, Wails
   metadata, Docker metadata, and release notes all say the same version.
2. Run the full local test/race/vet/build/diff gate.
3. Build the macOS candidate with `make desktop-release-mac` on macOS. This
   creates an ad-hoc package and unsigned local `SHA256SUMS`.
4. Inspect the app, ZIP, DMG layout, Wails version, endpoint behavior, and
   release notice on a clean writable path.
5. Run the separate local `scripts/sign-release-assets.sh` operation using the
   owner-controlled key custody. Never put the private key in this repository
   or a student command.
6. Build and accept Windows/Linux artifacts on their native hosts if they are
   included in the release. Do not list a platform whose artifact was not
   tested.
7. The signed same-key bridge preview is now published. After Gates A and B,
   create the immutable stable tag and manually upload its exact signed bytes.
   GitHub Actions are neither required nor used.
8. Re-download every public asset, compare bytes and checksums, verify the
   detached signature, and rerun Gate A before announcing the release.

Until those gates are complete, the truthful label is **release candidate under
controlled validation**, not “fully ready” or “automatic rollout completed.”

## Publication refresh — 2026-08-31 (historical source record)

Protected PR [#40](https://github.com/div197/BOB-Gemini-Free/pull/40) merged the
CLI updater asset-selection hardening into `main` at `a80f08d`. The updater now
requires the canonical platform filename instead of accepting an arbitrary
lookalike filename that merely shares a platform suffix; desktop package
selection already used exact branded/legacy migration names. Focused, full,
race, vet, module, build, and release-source checks passed. This closes a
locally testable selector ambiguity; it does not prove public asset completeness,
signed stable publication, platform trust, or installed-device update
acceptance. At that historical snapshot the authoritative remote main ref was
`523ceeb`; use the repository refresh at the top of this document for
present-state decisions.
