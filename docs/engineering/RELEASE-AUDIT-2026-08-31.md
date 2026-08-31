# BOB Gemini Free — Current Release Audit

**Audit date:** 2026-08-31 (Asia/Kolkata)
**Source baseline before this audit change:** public `main` at `558e8609333e`
**Historical Preview 5 packaged-code baseline:** commit `c28d78736eaae436cc1f1f3b4ec6e0bbcd058b89` (PRs #71–#73; route clarity and Preview 5 release reconciliation)
**Current Preview 6 release source target:** commit `f9b3410e74d7ccc08487dc03788b54a201e12ade` (PRs #77–#86; browser-boundary, credential-input, telemetry, release-version, settings, desktop-coexistence, release-state, and gateway-key transport reconciliation)
**Operating mode:** local release engineering; no GitHub Actions, provider
credentials, cookies, or private-key export

## Decision

The repository is source-ready and has now published the signed, verified
`v0.2.0-preview.6` macOS prerelease from packaged source target `f9b3410`.
Its five assets were re-downloaded and reconciled byte-for-byte after manual
publication. The earlier Preview 5 installation migration remains a one-host
observation; Preview 6 is not yet a fleet-ready update. The existing Preview 4
and Preview 5 assets remain immutable and were not overwritten.

The Preview 6 package was built from the recorded release source target. The
browser-boundary, credential-input, telemetry, release-version, settings,
desktop-coexistence, release-state, and gateway-key transport follow-ups are
included in the newly published Preview 6 bytes. Current `main` continues the
release reconciliation, while the package itself remains immutable; all of it
still requires staged device acceptance before broad rollout.

Do not call the current state a stable student release. Apple Developer ID,
notarization, clean-device replacement/rollback, Windows/Linux acceptance,
live Google behavior, and the staged 20–30-device pilot remain separate gates.

## What is confirmed now

| Surface | Current evidence | Status |
|---|---|---|
| Public source | Preview 6 is packaged from `f9b3410e74d7ccc08487dc03788b54a201e12ade`; the earlier Preview 5 source baseline `c28d78736eaae436cc1f1f3b4ec6e0bbcd058b89` remains historical | VERIFIED |
| Public releases | Latest desktop preview is immutable `v0.2.0-preview.6`; Preview 5 and Preview 4 remain available as historical inputs | VERIFIED |
| Preview 7 package | All five public `v0.1.7-preview.7` assets verify against the checked-in Ed25519 public key | VERIFIED |
| Key custody | Keychain service `BOB-Gemini-Free-Release-Ed25519` was used by the local signer; the private value was not displayed, exported, or copied | VERIFIED_LOCAL |
| Source gate | `scripts/verify-release-source.sh v0.2.0-preview.6` passes on the release source target | VERIFIED |
| Go suite | `go test -count=1 ./...` passes on this host | VERIFIED |
| Historical Preview 5 candidate | Universal macOS ZIP/DMG, release notice, `SHA256SUMS`, and detached signature were freshly built from clean checkout `88f2881` whose source tree matched the then-current public release target, and passed local asset verification | VERIFIED_LOCAL_HISTORICAL |
| Historical Preview 5 package smoke | Fresh `open -n` launch owned loopback `127.0.0.1:8081`, returned `{"status":"ok"}` from `/healthz`, served the credential map, and shut down cleanly | VERIFIED_LOCAL_HISTORICAL |
| Preview 6 package | Fresh universal macOS ZIP/DMG, release notice, `SHA256SUMS`, and detached signature were built from `f9b3410`; bundle, architecture, DMG layout, signed manifest, and updater version checks passed | VERIFIED_LOCAL |
| Preview 6 public assets | All five uploaded assets were downloaded into a fresh directory, signature/checksums verified, and compared byte-for-byte with the local signed inputs | VERIFIED_LIVE |
| Preview 6 coexistence smoke | The `/tmp` candidate selected `127.0.0.1:53065` while the installed app stayed on `127.0.0.1:8081`, served `X-BOB-Version: v0.2.0-preview.6`, and shut down cleanly | VERIFIED_LOCAL |
| Historical Preview 5 public assets | All five published assets were downloaded into a fresh directory, signature/checksums verified, and compared byte-for-byte with the local signed inputs | VERIFIED_LIVE_HISTORICAL |
| Historical Preview 1 installation update to Preview 5 | **Help → Check for Updates** discovered Preview 5; explicit install closed the old app, replaced `/Applications/BOB Gemini Free.app`, restarted on `127.0.0.1:8081`, preserved the visible prior chat response, and About reported `v0.2.0-preview.5`; a second check reported no newer release. This remains the only live installed-bundle migration observation. | VERIFIED_LIVE_HISTORICAL |
| Automation | `.github/workflows` is absent; no Actions budget is required by the release process | VERIFIED |
| Current public stable endpoint | GitHub `/releases/latest` resolves to historical `v0.1.5`, not `v0.2.0` | VERIFIED |

## Preview 6 publication continuation — 2026-08-31

Public `v0.2.0-preview.6` is now the latest downloadable desktop release.
The release source at `f9b3410e74d7ccc08487dc03788b54a201e12ade` contains the
version-aware desktop gateway handshake and the fail-closed gateway access-key
transport guard. A universal Preview 6 package was rebuilt from that exact
source, signed with the owner-controlled Keychain path, verified against its
detached manifest, published manually, and re-downloaded into a fresh
directory. The local package launched while the older installed gateway owned
`127.0.0.1:8081`; it selected `127.0.0.1:53065`, exposed the exact
`X-BOB-Version: v0.2.0-preview.6` marker, rendered its own version, and shut
down cleanly. Full details and hashes are in
[`PREVIEW-6-LOCAL-VERIFICATION-2026-08-31.md`](PREVIEW-6-LOCAL-VERIFICATION-2026-08-31.md).

This is a controlled public beta, not a fleet claim. Apple trust, clean-device
rollback, live Google behavior, and pilot acceptance remain open.

The Preview 7 and Preview 5 asset verifications prove the published packages
and manifests use the current project trust anchor. The one-host update proves
the installed transaction on this Mac, not rollback after a deliberately
failed candidate, Apple platform trust, provider availability, or classroom
rollout.

## Version and installed-base matrix

The updater is user-consented. “Automatic update” means a bounded metadata
check and a visible install choice; it does not mean a silent classroom-wide
push.

| Installed build | Candidate behavior now that Preview 6 is published | Required condition |
|---|---|---|
| `v0.1.7-preview.7` | Legacy preview-only lookup selects the published `v0.2.0-preview.6` candidate in the mocked matrix | Same current project key, macOS package/manifest published, app copied to a writable location; live device transition remains open |
| `v0.2.0-preview.1`–`preview.5` | Current preview path selects the newest published preview when no newer stable exists; Preview 1 → Preview 5 was observed on one Mac | Same key and explicit consent; Preview 4 and Preview 5 assets remain unchanged; Preview 6 device transition remains open |
| Current-source `v0.2.0-preview.6` | No update to itself; later previews are selected by semver | A later signed successor and writable install target |
| `v0.2.0` stable build | Checks only the stable endpoint; it never downgrades into preview | A newer public stable release must exist |
| `v0.1.7-preview.6` or a build with the unrecoverable old key | Cannot verify a current-key release | One manual installation of a current-key package |
| Mounted DMG, App Translocation, or read-only app path | Update is refused before download/staging | Copy the app to `/Applications` or another writable location and relaunch |

The checked-in regression matrix locks the preview cases and installable
universal archive/manifest metadata. One real installed-bundle replacement,
helper restart, and preservation of visible local state are now observed; a
deliberate rollback and broader device acceptance remain device evidence rather
than source-test claims.

## Preview 6 release gate

Complete these in order, from a clean checkout based on public `main`:

1. **COMPLETED locally:** current-main source, full test, race, vet, module,
   and host-build checks passed.
2. **COMPLETED locally:** the host currently reports about 7.9 GiB available;
   Wails staging and temporary DMG/ZIP copies completed without cleanup of
   project data.
3. **COMPLETED locally:** built exactly `v0.2.0-preview.6` with the checked-in
   public key embedded.
4. **COMPLETED locally:** inspected the app bundle, version/channel, universal
   slices, ad-hoc signature, ZIP layout, DMG `/Applications` shortcut, and
   release notice.
5. **COMPLETED locally:** signed the exact asset directory through the
   owner-controlled Keychain. The private key was never copied to Git,
   clipboard, shell history, chat, or a student device.
6. **COMPLETED locally:** `scripts/verify-release-assets.sh` passed. The
   detailed current candidate receipt is in
   [`PREVIEW-6-LOCAL-VERIFICATION-2026-08-31.md`](PREVIEW-6-LOCAL-VERIFICATION-2026-08-31.md).
7. **COMPLETED:** publish the immutable GitHub prerelease manually, including
   the exact five assets and signed manifest. GitHub Actions were not used.
8. **COMPLETED:** download every public asset into a fresh directory, re-run
   verification, and compare public bytes with the signed local input.
9. **COMPLETED for coexistence on one host / OPEN for rollout:** test one writable
   `/Applications` Mac; then test two or three pilots before any 20–30-device
   wave. Record only version, OS/architecture, health result, generation
   class, and update result.

If a future gate fails, keep `v0.2.0-preview.6` and all earlier releases
immutable and do not reuse any published tag or overwrite its assets. Create a
new immutable preview identity after a corrected build.

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
