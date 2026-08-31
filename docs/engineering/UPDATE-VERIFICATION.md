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
time through `internal/updater.BuildUpdatePublicKey` and `-ldflags -X`. A
development build may use `BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY` as a local
fallback. Both forms accept base64 or hexadecimal. The key is deliberately
not fetched from GitHub. If the key, manifest, signature, asset entry, or
signature verification is missing or invalid, `--update` fails closed before
the binary can replace the installed executable.

The native desktop path is stricter: it accepts only the build-embedded public
key and never treats `BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY` as a desktop trust
anchor. Its stable-channel metadata must identify the exact package, positive
declared size, signed manifest, and official GitHub URLs. Public Preview 7 has
the current key, but its released updater predates stable-first discovery; a
same-key bridge preview or manual install is required before it can reach a
new stable release through the updater. Historical `v0.1.7-preview.3` has no
embedded desktop key and therefore cannot install a native update; the current
`v0.2.0-preview.3` package embeds the current project public key.

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

## Preview 3 publication evidence — 2026-08-31

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

## Remaining release gate

The owner-controlled macOS Keychain signer was exercised for the published
`v0.2.0-preview.3` candidate: the exact release directory received a detached
Ed25519 signature, passed `scripts/verify-release-assets.sh`, and produced a
0600 evidence receipt outside the worktree. All five public assets were then
downloaded, re-verified, and byte-compared with the local inputs. This proves
publication integrity, not a successful installed-app replacement.

Stable `v0.2.0` remains unpublished. A live successful update cannot be
claimed until the owner completes clean-device replacement, rollback, and pilot
acceptance for the published Preview 3 candidate. The private key remains
outside the repository and must stay in the local secret store.
