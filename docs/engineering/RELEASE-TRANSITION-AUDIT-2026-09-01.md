# Desktop Release Transition Audit — 2026-09-01

> This is the current transition record for already-installed BOB Gemini Free
> desktop builds. It supersedes the operational conclusions in the historical
> 2026-08-31 transition record without rewriting that historical evidence.

## Decision

There is no semantic-versioning defect between the `0.1.x` and `0.2.x`
families. The real compatibility boundary is the combination of:

1. the channel compiled into the installed binary (`stable` or `preview`);
2. the updater implementation shipped in that binary; and
3. the Ed25519 public key compiled into that binary.

Consequently, “every old version will update automatically” is not a truthful
release claim. The supported transition is a matrix, not a single global
upgrade path. A device that falls outside the matrix receives a clear manual
migration path; the updater must not weaken verification or silently change a
stable build into a preview build.

The public release inventory checked on 2026-09-01 contains `v0.1.5` as the
latest stable release, `v0.1.7-preview.1` through `.7`, and
`v0.2.0-preview.1` through `.9`. There is no public GitHub tag or release
named `v0.1.9`; `v0.1.9` is a source/changelog milestone only.

Preview 9 was rebuilt from reviewed `main` commit `4236f65` after the
artifact-family guard, signed through the owner-controlled macOS Keychain,
published manually as a prerelease, and reconciled against a clean public
download. The five release assets match byte-for-byte and pass detached
signature/checksum verification. The prior local package from `1410bc2` is
superseded and is not a release input.

## Installed-lineage matrix

| Installed identity | What the binary actually does | Safe result |
|---|---|---|
| Public `v0.1.7-preview.1`–`.3` | Early preview desktop builds do not carry the current project updater trust anchor; their public release set also lacks the current signed-manifest contract. | Manual installation of a current signed preview is required once. Do not promise an in-app update. |
| Public `v0.1.7-preview.4`–`.6` | These packages belong to the superseded project-key lineage. A current manifest signed by the present project key cannot be verified by them. | Manual installation of a current signed preview is required once. This is a trust-lineage boundary, not a retry problem. |
| Public `v0.1.7-preview.7` | Contains the current project public key and signed package contract, but its updater is preview-only and predates the stable-first migration code. | Once a newer signed preview is published, it can discover that preview directly and request a consented update. It cannot discover a future stable release directly. |
| Public `v0.2.0-preview.1`–`.6` and `.8` | Carry the current project key. The current source lineage checks stable first, then continues to the highest published preview when stable has no newer native package for the current platform. A newer stable CLI-only release does not mask a native preview. | Can move to a newer signed preview, or directly to a newer signed stable release once that stable release exists. |
| Local/source `v0.1.9` build | No public release exists. The `4c8dbee` source milestone labels the normal Wails build as stable and does not inject the current desktop public key through its Makefile. | No reliable automatic migration can be promised. It must be manually replaced with a signed current package. A stable-only build must never be redirected to a prerelease merely to make the version appear current. |
| Public stable `v0.1.5` | This is a CLI release, not a signed native Wails desktop release. Its public assets do not satisfy the native signed-manifest package contract, and the tagged source predates the updater implementation. | It cannot prompt itself, auto-convert into the native desktop app, or update through today's code. Use a manual native installation; a future signed stable CLI release can serve updater-capable later CLI binaries. |
| Public `v0.2.0-preview.9` | Rebuilt from reviewed `main` commit `4236f65`, with the current project public key embedded; the exact five assets were signed through Keychain and reconciled against the public release. | It is the current controlled macOS beta. Preview 7 can discover it, but installed replacement, restart, rollback, and fleet acceptance remain device gates. |
| Prior local Preview 9 candidate from `1410bc2` | Locally packaged with the current key, but built before the artifact-family guard in `6a5606d`; it is superseded and not a valid release input. | Never publish or distribute those bytes. |

The early-preview rows are based on downloaded public release assets and
historical binary/source inspection. The absence of `v0.1.9` is based on the
GitHub release inventory and a direct release lookup, not on a naming
assumption.

## Channel rules

The updater intentionally has asymmetric channel movement:

```text
preview build  ── stable-first ──> newer stable native package, if available
      │
      └──────── preview list ────> highest newer native preview otherwise

stable build   ── stable endpoint ─> newer stable only
                                  └> never an automatic preview downgrade
```

This preserves an important user expectation: installing a stable build does
not opt a student into prerelease software. A local `v0.1.9` binary that was
mistakenly labelled stable therefore does not get a preview merely because
the preview number is newer. That behavior is deliberate and is now covered
by `TestStableDesktopBuildNeverFollowsPreviewChannel`.

The published Preview 7 binary has the older preview-only code path. Current
source cannot retroactively change that binary. Publishing a current preview
does not update the updater algorithm inside already-installed Preview 7; it
only gives that binary a compatible signed target to discover.

## What happens on a real device

### Public Preview 7 after a newer preview is published

The installed binary performs the following sequence when the user chooses
**Help → Check for Updates**:

1. Query the bounded official GitHub preview list.
2. Select the highest canonical `vX.Y.Z-preview.N` release.
3. Show the exact version and ask for **Install Update**.
4. Check that the running app is in a writable, non-translocated location.
5. Download the package, `SHA256SUMS`, and `SHA256SUMS.sig` only after consent.
6. Verify the manifest with the embedded project key, then verify the exact
   package digest, size, archive shape, binary format, and macOS bundle
   signature.
7. Stage a helper and transaction plan beside the installed app, quit the
   running Wails process, swap the app, relaunch it, and wait for the local
   healthy-startup marker.
8. Keep the previous app until confirmation succeeds; restore it if startup
   fails or the transaction is interrupted.

The actual replacement still requires a writable application location and a
real installed-device test. Metadata selection and signed fixture tests are
not proof that an `/Applications` replacement has succeeded on every Mac.

### Current Preview build after a stable release is published

The current preview updater checks the stable endpoint first. If a newer
stable native package exists, it offers that package; if the native package is
present but lacks its signed manifest, it presents the stable release as a
manual-update case. If the stable release is CLI-only for the current
platform, it continues to the preview list and offers the highest newer
native preview. A stable metadata failure is surfaced; it is not silently
hidden behind a preview result.

### A local `v0.1.9` build

Today, the public stable endpoint reports `v0.1.5`, which is older than
`v0.1.9`, so a stable-channel `v0.1.9` build reports no stable update. It does
not query previews. If a future stable `v0.2.0` is published, a properly
keyed stable `v0.1.9` build could discover it; the `4c8dbee` Wails build does
not provide that trust proof, so staging still fails closed. Neither case is
a reason to make stable builds consume prereleases.

## Stateful engineering assessment

The updater is stateful where state protects a filesystem transaction, but it
is intentionally not a mutable remote fleet controller.

### State already protected

- Build version, channel, and desktop public key are compiled into the
  release. They cannot be replaced by a downloaded setting or a student's
  environment variable.
- A verified update is represented by a private same-filesystem plan,
  candidate, rollback copy, confirmation marker, and failure/warning record.
- An inter-process lock prevents two app instances from updating the same
  install concurrently.
- Startup recovery inspects only updater-owned plans that point to the exact
  current install target. It finalizes a confirmed transaction or restores an
  unconfirmed one; ambiguous state is refused rather than guessed.
- Config, cookies, gateway data, and chat state are outside the app-bundle
  replacement target. Preservation is protected by the transaction boundary,
  but still needs clean-device acceptance evidence.
- Background checks in current builds are metadata-only and consent-gated.
  They do not silently download, quit, replace, or restart a classroom Mac.

### State deliberately not invented

- There is no remote “desired version” or fleet database.
- There is no persisted channel override that could turn a stable build into a
  preview build without a new signed binary.
- There is no trust-key download or automatic key rotation.
- There is no guarantee that a build created from an untagged source milestone
  has valid public release provenance.

Adding a mutable channel ledger or a remote migration exception would make the
system appear more automatic while weakening the exact invariants that make
the updater safe. The correct solution for a missing or obsolete trust anchor
is a one-time manual bridge, followed by same-key updates.

## Regression evidence

The following source tests protect the transition contract:

- `TestIsNewerVersion` protects preview ordering and preview-to-stable
  semantic precedence.
- `TestLegacyPreview7CanDiscoverSameKeyBridgePreview` and
  `TestLegacyPreview7CanDiscoverNextPreviewCandidate` protect Preview 7's
  preview-only behavior.
- `TestLegacyPreview7CannotDiscoverStableWithoutBridge` prevents a historical
  Preview 7 binary from being described as stable-aware.
- `TestPreviewDesktopBuildCanMigrateToNewerStableRelease` protects the
  current preview stable-first path.
- `TestPreviewDesktopBuildChecksPreviewWhenStableHasNoUpdate` protects
  preview continuation.
- `TestPreviewDesktopBuildSkipsStableCLIOnlyRelease` protects the artifact
  family boundary when stable publishes a CLI without a native desktop asset.
- `TestPreviewDesktopBuildDoesNotHideStableCheckFailure` protects truthful
  failure reporting.
- `TestStableDesktopBuildNeverFollowsPreviewChannel` protects the stable-only
  boundary for identities such as a local `v0.1.9` build.
- Desktop staging, helper, recovery, manifest, checksum, signature, lock, and
  durability tests protect the post-discovery transaction.

These tests are deterministic. They do not replace the one installed-bundle
observation required for each release lineage.

## Release and rollout gates

Before publishing a new preview as the student-facing current target:

1. Finish the transition documentation and ensure every current-facing page
   says “candidate” until the public release actually exists.
2. Build the package from one clean merged commit with the explicit preview
   version and the current public key.
3. Sign the exact five-file release set locally through the owner-controlled
   private-key path; never export or commit the private value.
4. Upload manually, without GitHub Actions, then download all public assets
   into a fresh directory and verify signature, checksums, asset names, and
   byte equality.
5. Test one writable installed Preview 7 device and one current Preview
   device. Record discovery, consent, staging, restart, health, state
   preservation, and rollback evidence separately.
6. Only after that pilot should the release be described as a controlled
   classroom preview. Apple Gatekeeper trust remains a separate Developer ID
   and notarization gate.
7. After publication, advance `PREVIEW_VERSION` to a new unused candidate so
   the next source tree cannot be rebuilt under an immutable published tag.

For the current fleet, the honest operator instruction is:

```text
Preview 1–6 or local v0.1.9  → manual install of the current signed preview
Preview 7                    → same-key update to the published Preview 9, with consent
Current v0.2 previews       → stable-first update only after stable exists
Stable                       → stable-only updates
```

## Current Preview 9 publication — 2026-09-01

The current public macOS prerelease is
[`v0.2.0-preview.9`](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.9),
built from reviewed `main` commit
`4236f65b9e4972a581d140ce46b0c5126602df65`. The local universal package was
signed through the owner-controlled Keychain; the public DMG, ZIP, release
notice, checksum manifest, and detached signature were downloaded into a
clean directory, verified, and compared byte-for-byte with the local signed
inputs.

The exact public `v0.1.7-preview.7` app was also downloaded and run in
isolation. Its native **Help → Check for Updates** flow displayed a signed
update dialog naming `v0.2.0-preview.9`. The test selected **Cancel** before
staging, so this is live discovery evidence, not installed-bundle
replacement, restart, rollback, or 30-device acceptance evidence. The public
Preview 7 web status endpoint may still report its legacy stable metadata
(`v0.1.5`); that endpoint is separate from the native Help updater and must
not be used to infer the native preview result.

| Current claim | Classification | Evidence | Boundary |
|---|---|---|---|
| Preview 9 is the current public macOS prerelease | VERIFIED_LIVE | GitHub release `v0.2.0-preview.9` is public with the universal DMG/ZIP, release notice, `SHA256SUMS`, and `SHA256SUMS.sig`; all five public files match the local signed inputs byte-for-byte and pass verification | The package remains ad-hoc signed and non-notarized; stable, Windows, Linux, clean-device, rollback, provider, and fleet claims remain open. |
| Preview 9 identifies the reviewed source and current updater trust anchor | VERIFIED_LIVE | The release receipt records source `4236f65`, Keychain-backed signing, public-key digest, and the exact manifest/signature hashes | The private key remains local; this does not establish Apple Developer ID/notarization or future key custody. |
| Public Preview 7 can discover Preview 9 | VERIFIED_LIVE | The exact public Preview 7 app displayed the Preview 9 consent dialog through **Help → Check for Updates** on an isolated audit run | The install action was canceled; replacement, restart, rollback, clean-device, and fleet acceptance remain open. |
| Preview 9 silently updates all 30 student Macs | STALE_OR_INCORRECT | The updater requires explicit consent and a writable application location; background checks are metadata-only | Each Mac needs an approved install or manual recovery. No GitHub Actions or remote classroom push exists. |
| A stable `v0.2.0` student release is ready | UNKNOWN | Preview 9 package, source, signature, public-byte, and discovery gates are green | Apple trust, rollback, clean-device, Windows/Linux, live Google behavior, and staged pilot gates remain open. |

## Deliberate non-claims

This audit does not claim that:

- The superseded local Preview 9 package is a valid release input;
- a version number alone proves release provenance;
- every historical package carries the current trust key;
- a background check silently updates 30 devices;
- a signed project manifest removes macOS Gatekeeper warnings;
- Google availability, quotas, cookies, API keys, model identity, or upstream
  rate limits are improved by the desktop updater.

The central conclusion is simple: same-key preview lineage is updateable;
obsolete-key or unproven local lineage requires one manual migration. The
historical stable `v0.1.5` CLI is older than the updater itself, so no current
repository change can make that already-installed binary show a new prompt.
The manual bridge is not a product failure—it is the security boundary that
prevents an untrusted binary or artifact-family change from gaining update
authority.
