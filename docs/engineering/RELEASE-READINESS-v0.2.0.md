# BOB Gemini Free v0.2.0 Release Readiness

**Audit date:** 2026-08-29 (Asia/Kolkata)
**Base HEAD before this readiness preparation:** `59a0d228ab8602427820ae90a14efe5f36f38ccd`
**Previous public fleet release:** `v0.1.7-preview.7`
**Current public migration bridge:** `v0.2.0-preview.1`
**Decision:** **NOT READY for publication as a student-facing stable release**

## Current repository refresh — 2026-08-31

The historical publication entries below are retained as provenance. The
current local truth is:

- `origin/main` is `523ceeb`.
- The reviewed source-hardening tip is pushed to
  `codex/release-readiness-v0.2.0` at `5eae3e2`, ahead of `origin/main`; the
  branch contains the subsequent audit documentation, coalesced-stream,
  remote-image, and updater-preflight hardening commits.
- The current branch contains the later 100-path hardening follow-ups,
  including nil-safe server and Gemini-client optional logging, accessible
  attachment/image controls, and JavaScript-URL-free gateway recovery, but those
  commits are not merged into `main`.
- No stable `v0.2.0` tag or release was created here, and no GitHub Actions
  workflow was added or run.
- This refresh does not close signed-asset publication, Apple/Windows
  platform trust, clean-device updater, live provider, browser, or 30-device
  rollout gates.
- The native updater now preflights the current install location before any
  release artifact download and explains App Translocation/read-only paths;
  the preflight is source- and fixture-tested but still needs a real
  `/Applications` installed-bundle run.
- A fresh local macOS Preview package was built from clean commit `1cc33d5`;
  its universal bundle passed ad-hoc `codesign --verify`, the DMG contained a
  visible `/Applications` shortcut, and the package carried no private key or
  signed release manifest. This is local package evidence, not a publication
  or clean-device update proof.
- The current tip also contains the isolated Studio correctness pass: native
  button semantics, drawer `aria-hidden`/`inert` state, bounded dialog
  surfaces, accessible selector names, a prompt skip link, and synchronized
  generated web output. These are source/regression results, not visual
  browser acceptance.

This is a release gate, not a claim that the source is unusable. The current
branch is a large post-Preview-7 source milestone, but a signed artifact,
platform trust, an update migration, and a 30-device acceptance run are four
different proofs.

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

These changes were reviewed through protected PR [#31](https://github.com/div197/BOB-Gemini-Free/pull/31)
and are published on `main` at merge commit `c5fa74f`. The native package
evidence below was produced from clean source commit `d318b4f`, an ancestor of
that published source; subsequent commits refreshed the evidence documents and
release wiring. The signed `v0.2.0-preview.1` migration bridge is publicly
published. Stable `v0.2.0` remains unpublished until the clean-device and pilot
gates pass.

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

The working tree is currently clean. At the time of this readiness review, the
complete post-Preview-7 branch delta was 134 files relative to the then-current
`origin/main`; the reviewed source is now published on `main` as PR #31. This
remains a release-candidate audit of a large feature milestone, not a routine
patch release. Repository-wide weighted statement coverage measured 63.6% in
the current local run; the project must not claim blanket 80% coverage. The
current local-only 1/10/20/30-concurrency baseline is recorded in
[`LOCAL-BENCHMARK-2026-08-29.md`](LOCAL-BENCHMARK-2026-08-29.md); it is not a
Google capacity or latency result.

The current clean commit `d318b4f` also passed a fresh `make desktop-preview-mac`
package run on this Mac. The resulting universal ad-hoc-signed app, ZIP, DMG,
Applications shortcut, checksum file, bundle metadata, local PWA routes, and
native GUI quit/shutdown path were verified. This remains a local unsigned-
manifest candidate: the current source still requires a signed release
manifest, public-upload reconciliation, clean-device updater run, and pilot
before publication. `spctl` rejection remains expected for a package without
Apple Developer ID trust. An intentionally missing Keychain item also caused
the manifest signer to fail closed; the real private key was not read.

The public GitHub state was also checked:

- the newest preview release is `v0.2.0-preview.1`, published as the same-key
  migration bridge;
- its macOS universal DMG/ZIP, release notice, checksum manifest, and detached
  signature are present and were re-downloaded and verified;
- the reviewed source hardening was published to `main` through PR #31 as
  merge commit `c5fa74f`;
- the release-source coherence, installer trust-anchor, and session-only
  gateway-auth follow-ups were subsequently merged through PRs #33, #36, and
  #38; that historical snapshot recorded `origin/main` as `f3a0a8c`; the
  current authoritative `origin/main` is `523ceeb` (PR #41, which contains
  those earlier merge ancestors);
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
| Exact artifact integrity | Updater verifies the signed `SHA256SUMS` entry, size, package type, and platform magic | Must be regenerated and verified after the final upload bytes are known |
| macOS platform trust | No Apple Developer ID certificate, hardened-runtime notarization, or stapled ticket is available in this workflow | The result must remain clearly labelled project-signed/ad-hoc and may require first-launch approval |
| Windows publisher trust | No Windows publisher-signed installer has been accepted in this audit | Windows cannot be called production-ready from this branch |
| Release channel | Signed preview bridge `v0.2.0-preview.1` is public; stable `v0.2.0` is not | Existing Preview 7 devices can update to the bridge; stable remains gated |
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
| Already-installed public `v0.1.7-preview.7` binary, current Preview-7 key, app copied to a writable directory | It can now update to the published bridge; it still cannot jump directly to stable | Use **Help → Check for Updates**, install `v0.2.0-preview.1`, then use the bridge's stable-first updater after stable is published; alternatively install stable manually |
| A newly built current-source preview bridge, current key, writable directory | Yes, after a newer signed stable package is published | User performs the explicit bridge update, then selects Check for Updates again for Preview → Stable |
| Preview 4–6 or another build with the old/unrecoverable project key | No, not cryptographically | One manual install of a package carrying the current public key, then later updates can be verified |
| Preview 3 or older without an embedded updater key | No | Manual installation; do not use an environment variable as a production trust substitute |
| App still running from a mounted DMG or App Translocation path | No | Copy it to `/Applications` or another writable application directory and relaunch |
| A package with no signed manifest or wrong platform asset | No | Updater must refuse it; use the official release page for recovery |

Therefore, if all 30 students truly have the public Preview 7 binary, the
published bridge now provides the first updater step, but not a direct
one-step Preview 7 → stable migration. Pilot the two explicit update steps,
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
acceptance. The current authoritative remote main ref is `523ceeb`; use the
repository refresh at the top of this document for present-state decisions.
