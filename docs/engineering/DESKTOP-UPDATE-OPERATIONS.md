# Desktop Update Operations and Rollout Contract

**Status:** Preview 6 enables a signed, user-consented preview updater; the
public app remains ad-hoc signed and not Apple-notarized.

This document is the operator and product boundary for the native updater. An
updater can be correct in source and still be unsafe to announce if the
release key, platform signatures, artifact list, or clean-device evidence is
missing.

## What students will eventually experience

On a signed production build, the Help menu can query the fixed stable GitHub
release channel, show the exact target version and package, download only after
the user confirms, verify a signed `SHA256SUMS` manifest, stage the package
beside the installed app, restart through a short-lived helper, and roll back
if the new app does not confirm healthy startup.

The public `v0.1.7-preview.6` build contains the embedded public update key
and signed `SHA256SUMS`/`SHA256SUMS.sig` manifest. Its update path is still
explicit and user-consented; it is not a hidden or silent auto-update.
Preview 3 remains a manual migration path because it predates the trust key.

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
same filesystem, keeps a rollback path until healthy confirmation, and writes
only a local, mode-0600 failure/warning record. It sends no telemetry.

### No-Actions operating model

Manual release engineering saves Actions budget but moves responsibility to
the maintainer. Use a two-person review of the version, asset list, manifest,
signature, checksums, platform publisher evidence, and release notes. A green
local test run is not proof that the GitHub asset upload preserved the bytes.

## Rollout gates for the 30 Macs

1. One clean Mac: install, first-run gateway, anonymous request behavior,
   per-user sign-in path, Preview 6 update, rollback, and uninstall.
2. Two or three pilot Macs: repeat with ordinary student accounts and the
   real classroom network; record version, OS, architecture, and provider
   session result without recording cookies or prompts.
3. Full rollout: only after the pilot has no unresolved updater or startup
   issue and the teacher has a one-page recovery instruction.

The updater reduces repeated manual replacement work; it does not remove the
need for a support owner, a release key, per-user authentication instructions,
or a manual recovery artifact.

## Current decision

The code path is appropriate for the signed Preview 6 pilot, but the
repository must not label an ad-hoc/unsigned package as a production
auto-updating student release. The remaining gates are Apple/Windows platform
trust and clean-device acceptance, not a missing fake fallback.
