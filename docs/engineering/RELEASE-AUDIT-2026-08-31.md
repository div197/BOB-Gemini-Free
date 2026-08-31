# BOB Gemini Free — Current Release Audit

**Audit date:** 2026-08-31 (Asia/Kolkata)  
**Source baseline before this audit change:** public `main` at `558e8609333e`
**Current audited public code baseline:** commit `4beb1275833c387f3bcf458d99b5743720e84311` (PR #68; credential-boundary fix)
**Operating mode:** local release engineering; no GitHub Actions, provider
credentials, cookies, or private-key export

## Decision

The repository is source-ready and now has a locally built, signed, and
verified `v0.2.0-preview.5` macOS candidate from the current public `main` at
`4beb127`.
That candidate is ready for controlled publication review, but it is not yet a
public release or a fleet-ready update. The existing `v0.2.0-preview.4`
assets remain immutable and were not overwritten.

Do not call the current state a stable student release. Apple Developer ID,
notarization, clean-device replacement/rollback, Windows/Linux acceptance,
live Google behavior, and the staged 20–30-device pilot remain separate gates.

## What is confirmed now

| Surface | Current evidence | Status |
|---|---|---|
| Public source | Public `main` contains the audited code baseline `4beb1275833c387f3bcf458d99b5743720e84311`; PRs #62–#68 are merged and the release-evidence commit is documentation-only | VERIFIED |
| Public releases | Latest desktop preview is immutable `v0.2.0-preview.4`; no public Preview 5 exists | VERIFIED |
| Preview 7 package | All five public `v0.1.7-preview.7` assets verify against the checked-in Ed25519 public key | VERIFIED |
| Key custody | Keychain service `BOB-Gemini-Free-Release-Ed25519` was used by the local signer; the private value was not displayed, exported, or copied | VERIFIED_LOCAL |
| Source gate | `scripts/verify-release-source.sh v0.2.0-preview.5` passes | VERIFIED |
| Go suite | `go test -count=1 ./...` passes on this host | VERIFIED |
| Preview 5 candidate | Universal macOS ZIP/DMG, release notice, `SHA256SUMS`, and detached signature were built from `4beb127` and passed local asset verification | VERIFIED_LOCAL |
| Preview 5 package smoke | Fresh `open -n` launch owns loopback `127.0.0.1:8081`, returns `{"status":"ok"}` from `/healthz`, serves the credential map, and shuts down cleanly | VERIFIED_LOCAL |
| Automation | `.github/workflows` is absent; no Actions budget is required by the release process | VERIFIED |
| Current public stable endpoint | GitHub `/releases/latest` resolves to historical `v0.1.5`, not `v0.2.0` | VERIFIED |

The Preview 7 asset verification proves the published package and manifest use
the current project trust anchor. The local Preview 5 evidence proves a
candidate package and one-host startup, not public-byte reconciliation,
installed-bundle replacement, rollback, Apple platform trust, provider
availability, or classroom rollout.

## Version and installed-base matrix

The updater is user-consented. “Automatic update” means a bounded metadata
check and a visible install choice; it does not mean a silent classroom-wide
push.

| Installed build | Candidate behavior once Preview 5 is published | Required condition |
|---|---|---|
| `v0.1.7-preview.7` | Legacy preview-only lookup can select `v0.2.0-preview.5` | Same current project key, macOS package/manifest published, app copied to a writable location |
| `v0.2.0-preview.1`–`preview.4` | Current preview path selects the newest preview when no newer stable exists | Same key and explicit consent; Preview 4 assets remain unchanged |
| Current-source `v0.2.0-preview.5` | No update to itself; later previews are selected by semver | Published signed successor and writable install target |
| `v0.2.0` stable build | Checks only the stable endpoint; it never downgrades into preview | A newer public stable release must exist |
| `v0.1.7-preview.6` or a build with the unrecoverable old key | Cannot verify a current-key release | One manual installation of a current-key package |
| Mounted DMG, App Translocation, or read-only app path | Update is refused before download/staging | Copy the app to `/Applications` or another writable location and relaunch |

The checked-in regression matrix now locks the first three preview cases and
the installable universal archive/manifest metadata. A real installed-bundle
replacement, helper restart, rollback, and preservation of the user's local
data remain device evidence rather than source-test claims.

## Preview 5 release gate

Complete these in order, from a clean checkout based on public `main`:

1. **COMPLETED locally:** source, full test, race, vet, module, and host-build
   checks passed.
2. **COMPLETED locally:** the host currently reports about 7.9 GiB available;
   Wails staging and temporary DMG/ZIP copies completed without cleanup of
   project data.
3. **COMPLETED locally:** built exactly `v0.2.0-preview.5` with the checked-in
   public key embedded.
4. **COMPLETED locally:** inspected the app bundle, version/channel, universal
   slices, ad-hoc signature, ZIP layout, DMG `/Applications` shortcut, and
   release notice.
5. **COMPLETED locally:** signed the exact asset directory through the
   owner-controlled Keychain. The private key was never copied to Git,
   clipboard, shell history, chat, or a student device.
6. **COMPLETED locally:** `scripts/verify-release-assets.sh` passed. The
   detailed candidate receipt is in
   [`PREVIEW-5-LOCAL-VERIFICATION-2026-08-31.md`](PREVIEW-5-LOCAL-VERIFICATION-2026-08-31.md).
7. **OPEN:** publish a new immutable GitHub prerelease manually, including the exact
   five assets and signed manifest. GitHub Actions are not part of this flow.
8. **OPEN:** download every public asset into a fresh directory and re-run verification;
   compare public bytes with the signed local input before announcing it.
9. **OPEN:** test one writable `/Applications` Mac, then two or three pilots, before
   any 20–30-device wave. Record only version, OS/architecture, health result,
   generation class, and update result.

If any gate fails, keep `v0.2.0-preview.4` as the known public release and do
not reuse the Preview 5 tag. Create a new immutable preview identity after a
corrected build.

## Remaining external gates

- Apple Developer ID signing, hardened runtime, notarization, and stapling;
- clean-device Gatekeeper and `/Applications` update/rollback acceptance;
- Windows publisher signing/SmartScreen and Linux display-library acceptance;
- live anonymous, session, model, vision, and Developer API behavior;
- provider quota/rate-limit behavior on the actual classroom network; and
- staged pilot evidence before the 30-device rollout.

These are intentionally not replaced with fake fallbacks, rotating accounts,
proxy-based quota evasion, shared cookies, or silent update behavior.

## Security and hygiene note

The GitHub PAT and provider API keys pasted into the conversation are
compromised credentials and must be revoked/rotated by the owner. They were
not used, printed, stored, or added to this repository during this audit. The
release private key remains a local secret; only the public trust anchor is
checked in.
