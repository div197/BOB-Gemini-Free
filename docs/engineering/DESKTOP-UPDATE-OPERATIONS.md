# Desktop Update Operations and Rollout Contract

**Status:** Preview 7 enables a signed, user-consented preview updater; the
public app remains ad-hoc signed and not Apple-notarized.

The immutable public migration bridge is `v0.2.0-preview.1`. The current
source defaults its next preview package to `v0.2.0-preview.2`; a local
Keychain-backed candidate has been signed and verified, but it is not tagged or
published yet. Stable `v0.2.0` remains gated on clean-device and pilot
acceptance.

This document is the operator and product boundary for the native updater. An
updater can be correct in source and still be unsafe to announce if the
release key, platform signatures, artifact list, or clean-device evidence is
missing.

## What students will eventually experience

On a signed production build, the Help menu can query the fixed stable GitHub
release channel, show the exact target version and package, download only after
the user confirms, verify a signed `SHA256SUMS` manifest, stage the package
beside the installed app, restart through a short-lived helper, and roll back
if the new app does not confirm healthy startup. New native builds from the
current source also perform a delayed startup check and then check at most once
per 24 hours; this background path only discovers and presents an update, and
never installs without the same explicit user confirmation. Existing binaries
keep the behavior compiled into their release.

Before an installable update dialog is shown, the current source performs a
local, no-network preflight of the running bundle. It rejects macOS App
Translocation and read-only or non-writable same-filesystem locations with
recovery guidance before downloading the release package. The staging step
repeats this check because permissions can change after discovery; this is an
error-prevention improvement, not a bypass of the signed-manifest boundary.

This preflight does not make the updater silent: the user still chooses
**Install Update**, and the helper still waits for a local healthy-startup
confirmation before deleting the rollback copy.

If the helper or machine is interrupted after the transaction starts, the next
native launch inspects only a validated plan belonging to that exact install.
A healthy confirmation finalizes the candidate and removes the rollback copy; a
missing confirmation restores the previous app. If the filesystem presents an
ambiguous state, BOB refuses to guess and shows a visible startup error for
manual recovery. The recovery path is fixture-tested but is not a substitute
for a clean-device interrupted-update acceptance run.

The updater's small plan, confirmation, failure, and warning records are
written to a same-directory temporary file, flushed, atomically replaced, and
followed by a Unix directory sync. The helper also flushes the install
directory after moving the old app to rollback, activating the candidate, or
restoring the previous app; startup recovery does the same after its cleanup
transitions. This reduces the window in which a power interruption can lose the
transaction decision, but it is not a recursive fsync of every file in a macOS
app bundle, and Windows does not have a portable directory-fsync contract.
`b136724` and its updater tests protect this boundary.

The public `v0.1.7-preview.7` build contains the embedded public update key and
signed `SHA256SUMS`/`SHA256SUMS.sig` manifest. Its update path is still
explicit and user-consented; it is not a hidden or silent auto-update. The
released Preview 7 binary predates the later stable-first source change and
therefore discovers only newer previews. If preserving an updater path for
those installations matters, publish a same-key bridge preview first; the
bridge can then discover a newer stable release. A direct stable install is
the simpler alternative. Preview 6 installations require a one-time manual
migration because the original Preview 6 project signing key was not
recoverable. Preview 3 also remains a manual migration path because it
predates the trust key.

The local macOS, Windows, and Linux preview packagers now fail closed when a
non-development package is built without `BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY`.
That prevents a future preview from being labelled updater-capable while its
desktop binary has no embedded trust anchor. The private signing key remains
required only by the separate local release-manifest/signing step and is never
embedded.

## Trust and custody controls

- The Ed25519 private key is a release secret. It must never enter Git,
  release notes, shell history, a screenshot, or a student machine.
- The matching public key is embedded into a release build at build time. A
  desktop process cannot replace this trust anchor with an environment
  variable or downloaded configuration.
- Each release tag is immutable in practice: never reuse a version after
  publishing an artifact or manifest. Publish a new patch/pre-release version
  for every changed byte.
- Keep a private, offline backup of the release key. Key loss prevents future
  signed updates for already-installed builds; key compromise requires a
  documented rotation and a migration release.
- GitHub Actions are not required. The operator must retain the exact local
  command output, manifest, signature, public-key fingerprint, and acceptance
  record for every published release.

## Required release evidence

Before a release is called updater-capable, the operator must have all of the
following:

1. A native artifact for every platform named in the release notes.
2. A signed `SHA256SUMS` and detached `SHA256SUMS.sig` covering the exact
   uploaded bytes.
3. The public key embedded in the build and a verified public/private key
   match before publication.
4. macOS Developer ID signing, hardened runtime, notarization, stapling, and
   a clean-device Gatekeeper check for a professional Mac release.
5. Windows publisher signing and clean-device SmartScreen/install checks for
   a professional Windows release.
6. A clean-device update from the previous signed build, followed by a
   deliberately failed candidate test proving rollback without touching
   cookies, config, chat history, or gateway data.
7. A manual release-page recovery path in the release notes.

An Ed25519 manifest authenticates the bytes published by the project. It does
not itself create Apple Developer ID trust or Windows publisher reputation.

## Human and product risks

### Release-key failure

If the private key is lost, the application cannot safely accept a new release
under the old trust anchor. If it is exposed, a malicious release can pass the
project updater even if GitHub account access is later recovered. Keep the key
under owner-controlled custody and record a rotation procedure before a
30-device rollout.

### Provider/session failure

An updated app can start correctly while Google rejects an expired or absent
student session. The updater must never claim that a successful installation
means that text, vision, Pro, Imagen, quota, or anonymous access is available.
Each student uses their own authorized session; no teacher cookie is bundled.

### Platform trust failure

Unsigned or ad-hoc packages can still trigger Gatekeeper or SmartScreen even
when the project manifest is valid. Do not use the updater as a workaround for
platform trust warnings, and do not ask students to disable operating-system
security controls.

### Update interruption

Power loss, disk-full conditions, permissions, antivirus/WebView locks, and a
still-running process can interrupt replacement. The helper stages on the
same filesystem, takes an atomic per-install lock before replacement,
revalidates the install target and candidate after locking, keeps a rollback
path until healthy confirmation, and writes only a local, mode-0600
failure/warning record. Old committed staging plans are cleaned conservatively
on a later staging attempt; interrupted committed plans are repaired on native
startup; incomplete pre-plan directories are left for manual inspection. It
sends no telemetry.

### No-Actions operating model

Manual release engineering saves Actions budget but moves responsibility to
the maintainer. Use a two-person review of the version, asset list, manifest,
signature, checksums, platform publisher evidence, and release notes. A green
local test run is not proof that the GitHub asset upload preserved the bytes.

## Rollout gates for the 30 Macs

1. One clean Mac: install, first-run gateway, anonymous request behavior,
   per-user sign-in path, bridge-preview installation/update if testing the
   existing Preview 7 fleet path, stable update, rollback, and uninstall.
   Preview 6 devices require the documented one-time manual migration before
   the signed updater can be used.
2. Two or three pilot Macs: repeat with ordinary student accounts and the
   real classroom network; record version, OS, architecture, and provider
   session result without recording cookies or prompts.
3. Full rollout: only after the pilot has no unresolved updater or startup
   issue and the teacher has a one-page recovery instruction.

The updater reduces repeated manual replacement work; it does not remove the
need for a support owner, a release key, per-user authentication instructions,
or a manual recovery artifact.

For the exact existing-fleet sequence, use the
[`Preview 7 to v0.2.0 Migration Runbook`](PREVIEW-7-TO-V0.2.0-MIGRATION.md).

## Current decision

The code path is appropriate for the signed Preview 7 pilot, but the
repository must not label an ad-hoc/unsigned package as a production
auto-updating student release. The remaining gates are Apple/Windows platform
trust and clean-device acceptance, not a missing fake fallback.
