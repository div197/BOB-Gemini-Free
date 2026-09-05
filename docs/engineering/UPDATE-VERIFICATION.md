# Release Update Verification — Phase III

## Current flaw

The updater previously calculated a SHA-256 digest for a downloaded binary and
logged it, but did not compare that digest with a trusted manifest or verify a
signature over the manifest. A digest calculated after download detects local
I/O corruption; it does not authenticate the download source.

## Release contract

Future GitHub releases must publish these assets alongside each platform
binary:

```text
SHA256SUMS
SHA256SUMS.sig
```

`SHA256SUMS` uses the conventional format:

```text
<64 lowercase hex SHA-256>  <exact binary asset name>
```

`SHA256SUMS.sig` is a base64-encoded Ed25519 signature over the exact bytes of
`SHA256SUMS`. The verifier accepts hexadecimal signatures as a convenience
for controlled release tooling, but base64 is the documented format.

The trusted Ed25519 public key is embedded into official CLI binaries at build
time through `internal/updater.BuildUpdatePublicKey` and `-ldflags -X`. An
unflagged development build remains version `dev` and does not identify itself
as a published update client; an operator may explicitly inject a release
version and use `BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY` as a local CLI fallback
for a controlled verification run. An update check is rejected before network
access when the current version is not a canonical release identity. Both
forms accept base64 or hexadecimal.
The key is deliberately
not fetched from GitHub. If the key, manifest, signature, asset entry, or
signature verification is missing or invalid, `--update` fails closed before
the binary can replace the installed executable.

The native desktop path is stricter: it accepts only the build-embedded public
key and never treats `BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY` as a desktop trust
anchor. Its stable-channel metadata must identify the exact package, positive
declared size, signed manifest, and official GitHub URLs. The public
`v0.2.0-preview.9` updater carries the current key; its preview-channel
transition remains subject to a live installed-device test. Historical
`v0.1.7-preview.3` has no embedded desktop key and therefore cannot install a
native update; the current `v0.2.0-preview.9` package embeds the current
project public key.

The local release command requires `BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY` and
injects it into every CLI matrix binary. On macOS,
`scripts/sign-release-assets.sh` reads the private signing value from the
owner-controlled Keychain and streams it to `cmd/release-manifest` over stdin;
other operators may use a protected local secret-manager environment. The
signer verifies that the derived public key matches the embedded public key. A
pair mismatch fails packaging before publication. No hosted CI service is
required for this process.

The updater rejects a GitHub asset whose declared size exceeds
`MaxUpdateArtifactBytes` (512 MiB), and the streaming download has the same
bound even when a server omits or lies about `Content-Length`.

For a native preview, a newer stable release is actionable only when the
release contains a matching native package for the current platform. A
stable CLI-only release is intentionally skipped in favor of the newer
native preview list; this keeps the CLI and Wails artifact families from being
treated as interchangeable.

## Embedded Studio status-check boundary

The `GET /v1/update/check` status route is metadata-only and now carries the
owning build's channel into `CheckLatestDesktopForChannelContext`. `server.New`
keeps the stable channel for CLI and embedded-library callers; the native
desktop constructor uses `NewWithUpdateChannel` with its build-pinned Wails
channel. A preview build therefore observes the same stable-first/preview
continuation policy as **Help → Check for Updates**: a newer stable release is
a migration target only when it contains a native package for the current
platform, so a stable CLI-only release cannot hide a newer native preview. A
development build is rejected before any GitHub request. The endpoint's JSON
remains additive and includes `channel`, package availability, and
signed-manifest availability.

This is discovery only. The browser status badge never downloads or installs a
package, and a native Wails user must still explicitly choose installation in
the Help dialog. `internal/server/server_test.go` locks channel forwarding and
`internal/server/playground_test.go` locks the corresponding native/preview
status copy. The same settings surface now blocks the optional Developer API
route when a prior connection check has positively identified `api_keys`
protection but the separate BOB Gateway Access Key is absent; this is a
pre-send error-prevention state, not an alternate authentication mechanism.
The probe state is endpoint-scoped, and Send stays blocked while an explicit
endpoint check is still pending so a fast click cannot create a known 401
failure. Stale telemetry from an older endpoint is ignored.
`TestCredentialRouteBlocksKnownGatewayAuthRequirement` protects the source
boundary. When the Developer API toggle is off, the Studio also sends an
explicit `X-BOB-Gemini-Route: web` selector so a process-level provider key
cannot silently contradict the route shown in Config; clients that omit that
selector retain the existing process-level behavior.

## Replacement boundary

The candidate is downloaded into a temporary file in the executable's own
directory, verified against the signed manifest, checked for platform binary
magic, and only then considered for replacement. Unix replacement uses an
atomic same-filesystem rename. The updater flushes the temporary metadata file
before its commit and flushes the Unix containing directory after transaction
renames and rollback; Windows retains the existing rollback path because a
running executable cannot be renamed over in place. Windows metadata commits
use native `MoveFileExW` replace-existing/write-through semantics; directory
flushing still has no portable Go contract.

`cmd/release-manifest` creates a sorted manifest from regular release files,
excluding `SHA256SUMS` and `SHA256SUMS.sig`, and signs the exact manifest bytes.
Tests verify deterministic manifest generation, embedded-key precedence,
base64/hexadecimal decoding, signature verification, oversized downloads,
tampering, checksum mismatch, and mocked downloads into temporary candidate
paths. They never call replacement against `os.Executable()` or the
developer's running binary.

## Operator sequence

1. Generate or retrieve the Ed25519 key pair using an offline, controlled
   process. Keep the private key only in the release secret store; publish the
   public key through the repository file and the project's trusted release
   documentation.
2. Confirm that the public and private values are the same key pair before
   tagging. The local command repeats this check, so an accidental mismatch cannot
   produce a release that every updater rejects.
3. Run `scripts/release-local.sh` from the checkout. It runs six CLI
   cross-builds and signs the resulting directory. Build native artifacts on
   their supported hosts, inject the public key at build time, and use
   `scripts/sign-release-assets.sh` after inspecting the combined artifact
   directory. Use `make desktop` for a normal desktop developer build, or
   `scripts/build-wails-local.sh` on macOS when the checkout is under File
   Provider.
4. Upload the resulting files through the operator's chosen release channel,
   then publish the binaries plus `SHA256SUMS` and `SHA256SUMS.sig`.
5. Verify the published asset names and manifest entries from a clean machine
   before announcing the release. Do not test replacement against a running
   developer executable.

The private key must never be placed in the repository, shell history, command
output, issue comments, or a local fixture. Rotate both values together if the
private key is exposed.

## Preview 6 publication evidence — 2026-08-31 (historical)

At that historical boundary, the public macOS preview was the published
[`v0.2.0-preview.6`](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.6),
packaged from source target `f9b3410`. The exact universal DMG, ZIP, release
notice, `SHA256SUMS`, and `SHA256SUMS.sig` were signed locally through the
owner-controlled Keychain, published manually without GitHub Actions, then
downloaded again from GitHub. Signature verification, checksum verification,
and byte-for-byte comparison all passed.

This proves release-byte authenticity and publication integrity. It does not
prove Apple Developer ID/notarization, a clean-device replacement or rollback,
live Google availability, or a 20–30-device rollout. The earlier Preview 1 →
Preview 5 installed migration is the only live installed-update observation;
Preview 7/Preview 5 → Preview 6 remained a staged device gate.

## Superseded Preview 7 candidate — 2026-08-31

The `main` source snapshot at `0c6a6ff12c3f2ff9963b73837933b2abae676270`
was freshly packaged and signed locally as `v0.2.0-preview.7` after PR #99.
This supersedes the earlier local same-version candidate built from `6c27ac8`
(which superseded `8c35a11`, `049ca2f`, and `2d42d44`). The fresh candidate
receipt is tied to the merged source snapshot.
Its candidate receipt is in
[`PREVIEW-7-CANDIDATE-VERIFICATION-2026-08-31.md`](PREVIEW-7-CANDIDATE-VERIFICATION-2026-08-31.md).
It is not a public GitHub release yet, so it must not be presented as an
available update or used for installed-fleet instructions. GitHub currently
reports `immutable: false` for Preview 6; the project's write-once release
policy therefore requires a new tag and fresh public-byte reconciliation for a
future candidate.

## Current Preview 9 publication — 2026-09-01

The current public macOS prerelease is
[`v0.2.0-preview.9`](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.9),
built from reviewed runtime source commit
`4236f65b9e4972a581d140ce46b0c5126602df65`. It was packaged, signed, and
verified locally; the exact package hashes and local startup evidence are
recorded in
[`PREVIEW-9-CANDIDATE-VERIFICATION-2026-09-01.md`](PREVIEW-9-CANDIDATE-VERIFICATION-2026-09-01.md).
The exact five-file release set was then uploaded manually without GitHub
Actions, downloaded again into a fresh directory, signature- and
checksum-verified, and compared byte-for-byte with the local signed inputs.

This proves public release-byte integrity and metadata discovery. It does not
prove Apple Developer ID/notarization, a clean-device replacement, rollback,
live Google availability, or a 20–30-device rollout. On the audit Mac, the
exact public Preview 7 app discovered Preview 9 through **Help → Check for
Updates**; the install action was canceled, so replacement remains an open
device gate.

## Signed discovery feed — 2026-09-05

The current source adds `updates/desktop-feed.json` and its detached
`updates/desktop-feed.json.sig` as a low-volume discovery layer. The feed is
signed with the same project Ed25519 key used for release manifests, has a
bounded validity window, and is pinned to the exact raw-content paths in
`internal/updater/update_feed.go`. It contains release metadata only; the
native archive and its signed `SHA256SUMS` manifest remain the installation
trust boundary.

`internal/updater/update_feed_test.go` verifies valid-feed selection without a
GitHub API request, signature/tamper rejection with API fallback, explicit
fresh-check bypass, expiry/validity limits, exact URL pinning, and the
checked-in feed signature against the documented public key. The feed is not
retroactive: public Preview 7–9 binaries retain their compiled API discovery
path. A future native build must refresh and sign the feed after public asset
reconciliation; if the feed is stale or unavailable, the source falls back to
the fixed GitHub API path. This improves discovery availability and request
spreading; it does not create silent installation or fleet control.

Canonical `github.com` release metadata is also bound to the selected tag and
asset name before it can reach the update UI. A release page must identify that
exact tag, and a package or manifest URL must identify both that tag and the
metadata asset name. GitHub's official opaque release CDN hosts remain allowed
because their paths are not tag-addressable; those downloads still require the
detached signed manifest before installation. This prevents a malformed or
tampered metadata response from pairing one release's displayed identity with
another release's download path.

## Preview 2 publication evidence — 2026-08-31

The controlled macOS `v0.2.0-preview.2` release is now published from public
main commit `6d3a0cfc0a7a0bf05a3c136baf96a48f503b45ef`:

[`github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.2`](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.2)

The local Keychain-backed signer produced the detached signature for the exact
DMG, ZIP, and release notice. A fresh download of all five public assets passed
`scripts/verify-release-assets.sh`, and every downloaded file matched its local
publication input byte-for-byte. The release is macOS universal only and is a
prerelease; it is not Apple Developer ID signed or notarized.

The public release makes the Preview 7 → Preview 2 discovery path real, but it
does not prove an installed-bundle replacement. A device must still be in a
writable application location, the user must explicitly approve installation,
and a clean-device/pilot run must observe restart, confirmation, rollback
preservation, and local configuration survival.

## Preview 5 publication and installed-migration evidence — 2026-08-31

The controlled macOS `v0.2.0-preview.5` release was built from the clean
source receipt `88f2881`, published manually from public `main` target
`c28d787`, and signed through the owner-controlled Keychain. The five public
assets were downloaded into a fresh directory; the detached signature and
checksums passed, and every downloaded file matched its local signed input
byte-for-byte. On one Mac, a writable `/Applications` Preview 1 installation
discovered Preview 5 through **Help → Check for Updates**, installed it after
explicit consent, restarted healthy, preserved the visible chat response, and
reported no newer release on a second check.

This proves the project-signature/public-byte boundary and one successful
same-key migration. It does not prove Apple platform trust, deliberate
rollback, clean-device acceptance, Google provider behavior, or fleet rollout.

## Preview 4 publication evidence — historical

The controlled macOS `v0.2.0-preview.4` release was built from public `main`
commit `abfeebaaaaabc740ea29602b602591a0b707fbc2`, signed through the local
owner-controlled Keychain, and published manually without GitHub Actions:

[`github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.4`](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.4)

The five public assets were downloaded into a fresh directory. The detached
Ed25519 signature verified, and every downloaded file matched the local signed
input byte-for-byte. The release is macOS universal, ad-hoc signed, not
Developer ID signed, and not notarized. The local artifact also passed fresh
launch, loopback `/healthz`, rendered-version, occupied-port fallback, and
clean-shutdown smoke testing.

This proves source/package/public-byte integrity. It does not prove an
installed-bundle replacement, rollback after interruption, Gatekeeper trust,
Google availability, or a 20–30-device rollout.

## Preview 3 publication evidence — 2026-08-31 (historical)

The controlled macOS `v0.2.0-preview.3` release was built from public `main`
commit `284b7d1a9a2e7c45402318f29f08f0c1dba36d43`, signed through the local
owner-controlled Keychain, and published manually without GitHub Actions:

[`github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.3`](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.3)

The five public assets were downloaded into a fresh directory. The detached
Ed25519 signature verified, and every downloaded file matched the local signed
input byte-for-byte. The release is macOS universal, ad-hoc signed, not
Developer ID signed, and not notarized. The local artifact also passed fresh
launch, loopback `/healthz`, and clean shutdown smoke testing.

This proves source/package/public-byte integrity. It does not prove an
installed-bundle replacement, rollback after interruption, Gatekeeper trust,
Google availability, or a 20–30-device rollout.

## Remaining release gate after Preview 4 — historical

The owner-controlled macOS Keychain signer was exercised for the published
`v0.2.0-preview.4` candidate: the exact release directory received a detached
Ed25519 signature, passed `scripts/verify-release-assets.sh`, and produced a
0600 evidence receipt outside the worktree. All five public assets were then
downloaded, re-verified, and byte-compared with the local inputs. This proves
publication integrity, not a successful installed-app replacement.

Stable `v0.2.0` remains unpublished. A live successful update cannot be
claimed until the owner completes clean-device replacement, rollback, and pilot
acceptance for the published Preview 4 candidate. The private key remains
outside the repository and must stay in the local secret store.
