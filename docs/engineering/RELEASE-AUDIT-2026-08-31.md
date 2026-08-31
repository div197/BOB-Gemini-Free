# BOB Gemini Free — Current Release Audit

**Audit date:** 2026-08-31 (Asia/Kolkata)  
**Source baseline before this audit change:** public `main` at `558e8609333e`
**Current audited packaged-code baseline:** commit `c28d78736eaae436cc1f1f3b4ec6e0bbcd058b89` (PRs #71–#73; route clarity and Preview 5 release reconciliation)
**Operating mode:** local release engineering; no GitHub Actions, provider
credentials, cookies, or private-key export

## Decision

The repository is source-ready and has published the signed, verified
`v0.2.0-preview.5` macOS prerelease from the current public `main` at
`c28d787`. Its assets were re-downloaded and reconciled byte-for-byte, and one
writable `/Applications` Preview 1 installation updated and restarted into
Preview 5 with its visible chat state preserved. The release is not yet a
fleet-ready update. The existing `v0.2.0-preview.4` assets remain immutable and
were not overwritten.

Do not call the current state a stable student release. Apple Developer ID,
notarization, clean-device replacement/rollback, Windows/Linux acceptance,
live Google behavior, and the staged 20–30-device pilot remain separate gates.

## What is confirmed now

| Surface | Current evidence | Status |
|---|---|---|
| Public source | Public `main` contains the audited code baseline `c28d78736eaae436cc1f1f3b4ec6e0bbcd058b89`; the route/release reconciliation PRs #71–#73 are merged | VERIFIED |
| Public releases | Latest desktop preview is immutable `v0.2.0-preview.5`; Preview 4 remains available as historical input | VERIFIED |
| Preview 7 package | All five public `v0.1.7-preview.7` assets verify against the checked-in Ed25519 public key | VERIFIED |
| Key custody | Keychain service `BOB-Gemini-Free-Release-Ed25519` was used by the local signer; the private value was not displayed, exported, or copied | VERIFIED_LOCAL |
| Source gate | `scripts/verify-release-source.sh v0.2.0-preview.5` passes | VERIFIED |
| Go suite | `go test -count=1 ./...` passes on this host | VERIFIED |
| Preview 5 candidate | Universal macOS ZIP/DMG, release notice, `SHA256SUMS`, and detached signature were freshly built from clean checkout `88f2881` whose source tree matches the public release target, and passed local asset verification | VERIFIED_LOCAL |
| Preview 5 package smoke | Fresh `open -n` launch owns loopback `127.0.0.1:8081`, returns `{"status":"ok"}` from `/healthz`, serves the credential map, and shuts down cleanly | VERIFIED_LOCAL |
| Preview 5 public assets | All five published assets were downloaded into a fresh directory, signature/checksums verified, and compared byte-for-byte with the local signed inputs | VERIFIED_LIVE |
| Existing writable Preview 1 installation updates to Preview 5 | **Help → Check for Updates** discovered Preview 5; explicit install closed the old app, replaced `/Applications/BOB Gemini Free.app`, restarted on `127.0.0.1:8081`, preserved the visible prior chat response, and About reported `v0.2.0-preview.5`; a second check reported no newer release | VERIFIED_LIVE |
| Automation | `.github/workflows` is absent; no Actions budget is required by the release process | VERIFIED |
| Current public stable endpoint | GitHub `/releases/latest` resolves to historical `v0.1.5`, not `v0.2.0` | VERIFIED |

The Preview 7 and Preview 5 asset verifications prove the published packages
and manifests use the current project trust anchor. The one-host update proves
the installed transaction on this Mac, not rollback after a deliberately
failed candidate, Apple platform trust, provider availability, or classroom
rollout.

## Version and installed-base matrix

The updater is user-consented. “Automatic update” means a bounded metadata
check and a visible install choice; it does not mean a silent classroom-wide
push.

| Installed build | Candidate behavior now that Preview 5 is published | Required condition |
|---|---|---|
| `v0.1.7-preview.7` | Legacy preview-only lookup can select `v0.2.0-preview.5` | Same current project key, macOS package/manifest published, app copied to a writable location |
| `v0.2.0-preview.1`–`preview.4` | Current preview path selects the newest preview when no newer stable exists; Preview 1 → Preview 5 was observed on one Mac | Same key and explicit consent; Preview 4 assets remain unchanged |
| Current-source `v0.2.0-preview.5` | No update to itself; later previews are selected by semver | Published signed successor and writable install target |
| `v0.2.0` stable build | Checks only the stable endpoint; it never downgrades into preview | A newer public stable release must exist |
| `v0.1.7-preview.6` or a build with the unrecoverable old key | Cannot verify a current-key release | One manual installation of a current-key package |
| Mounted DMG, App Translocation, or read-only app path | Update is refused before download/staging | Copy the app to `/Applications` or another writable location and relaunch |

The checked-in regression matrix locks the preview cases and installable
universal archive/manifest metadata. One real installed-bundle replacement,
helper restart, and preservation of visible local state are now observed; a
deliberate rollback and broader device acceptance remain device evidence rather
than source-test claims.

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
7. **COMPLETED:** publish the immutable GitHub prerelease manually, including the
   exact five assets and signed manifest. GitHub Actions are not part of this flow.
8. **COMPLETED:** download every public asset into a fresh directory, re-run
   verification, and compare public bytes with the signed local input.
9. **COMPLETED for one host / OPEN for rollout:** test one writable
   `/Applications` Mac; then test two or three pilots before any 20–30-device
   wave. Record only version, OS/architecture, health result, generation
   class, and update result.

If a future gate fails, keep `v0.2.0-preview.5` immutable and do not reuse its
tag or overwrite its assets. Create a new immutable preview identity after a
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
