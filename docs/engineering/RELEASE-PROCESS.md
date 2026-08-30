# Release Process

The supported release path is the Go CLI plus the native desktop artifacts.
The current Go desktop shell is the only active native desktop implementation. The former alternate
wrapper was removed from the working tree after the Phase III comparison; its
source remains recoverable through Git history, but it is not part of releases.

## Free open-source preview path

The project can build a macOS native preview without an Apple Developer
membership:

```bash
make desktop-preview-mac
```

This path creates an ad-hoc-signed `.app`, `.zip`, and `.dmg` with a visible
`/Applications` drag target, plus a release notice and local checksums. It is
explicitly not Developer ID signed,
notarized, or production-ready. It is intended for controlled evaluation only.
It does not replace the signed release process below and must not be uploaded
as a trusted student release without the warning notice.

The preview packager requires `BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY` even though
the app remains ad-hoc signed. This public value is the updater trust anchor;
it is not a credential. The corresponding private value is used only by the
local manifest-signing step and must never be placed in the repository or a
student command.

All CLI and native preview/release packagers run
`scripts/verify-release-source.sh` before copying or signing source. After the
manifest-signing step, run
`BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY=... scripts/verify-release-assets.sh
/path/to/release-assets`. That read-only check verifies the detached signature,
hashes every regular asset, and rejects any extra, missing, duplicate, or
symlinked file before the directory is uploaded.

The native Help menu's install action is intentionally user-initiated. Stable
builds check the fixed official stable channel. Newly built current-key preview
builds check stable first for a one-way Preview → Stable migration, then check
the fixed official preview channel when no stable update exists. A published
desktop build may also perform one delayed startup metadata check and then one
per day while running. Both paths verify a signed manifest and offer a
consented staged replacement with rollback. The already-published Preview 7
binary predates stable-first discovery, so it needs a same-key bridge preview
or a manual stable install before updater-based migration to stable. The
updater never silently downloads, replaces, or restarts a running native
bundle.

The standalone installers use the same fixed public key and fail closed unless
they can verify the detached Ed25519 signature and exact SHA-256 entry. The
default path never compiles the current directory and never installs an
unsigned binary. The macOS system OpenSSL shipped as LibreSSL on some hosts
does not provide Ed25519 verification; those hosts must use a separately
installed verifier or the explicitly opt-in source-build path. Windows needs
`curl.exe` plus OpenSSL or a runtime exposing Ed25519. This is deliberate:
missing verification is an installation failure, not permission to continue.
The installer script itself should be downloaded, inspected, and run as a
local file; do not pipe an unpinned branch directly into a shell.

## Native automatic-update status

Preview 4 was the first public native build with a build-embedded desktop trust
key and a signed preview manifest. It performs an explicit metadata check, and
a user can approve a verified staged update with health confirmation and
rollback. Preview 6 was the previous signed preview; `v0.2.0-preview.1` is the
current published migration bridge for the existing Preview 7 fleet. Preview 3 remains manual-update-only because it has no
embedded desktop trust key. Stable builds never move into preview; current-key
preview builds may migrate into a newer stable release only after explicit
user consent.

The source now enables that user-consented path for the signed macOS preview,
and the same-key `v0.2.0-preview.1` migration bridge is publicly available
for the existing Preview 7 fleet.
Platform publisher signing and clean-device verification remain required for a
professional student distribution mechanism. macOS Developer ID/notarization
and Windows publisher signing are separate operating-system trust gates; the
project signature cannot replace them.

The implementation sequence is:

1. publish a stable and preview channel contract for native packages;
2. keep the signed-manifest verifier bound to the exact `.app` archive and
   Windows installer/executable asset for each platform;
3. add a platform-specific post-exit helper with atomic replacement, rollback,
   permission failure reporting, and restart behavior;
4. add mocked download/tamper/rollback tests and clean-device acceptance;
5. keep the background metadata check low-frequency and non-installing, while
  retaining a visible “Check for Updates” action and a manual release fallback.

This removes repeated delete/download/install work after the migration bridge
is installed and the user approves the update. For the existing public Preview
7 fleet, install the published same-key bridge first, then use its stable-first
path after stable acceptance is complete. No silent install is enabled by the
current preview.

The non-technical rollout risks and operator checklist are recorded in
[`DESKTOP-UPDATE-OPERATIONS.md`](DESKTOP-UPDATE-OPERATIONS.md).

## Preconditions

The public value is supplied to the local release command. On macOS, the
signing wrapper prefers the owner-controlled Keychain item
`BOB-Gemini-Free-Release-Ed25519` for the current account and streams the
private value directly to the signer over stdin. The private-key environment
variable remains an explicit fallback for a local secret manager on other
systems; it is not required for the macOS Keychain path.

| Name | Purpose |
|---|---|
| `BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY` | Ed25519 public key embedded in CLI binaries and checked at signing time |
| `BOB_GEMINI_FREE_UPDATE_PRIVATE_KEY` | Optional non-macOS/local-secret-manager fallback used only by `cmd/release-manifest` |
| `BOB_GEMINI_FREE_UPDATE_KEYCHAIN_SERVICE` | Optional macOS Keychain service override; default is `BOB-Gemini-Free-Release-Ed25519` |
| `BOB_GEMINI_FREE_UPDATE_KEYCHAIN_ACCOUNT` | Optional macOS Keychain account override; default is the current OS user |

The public value and the Keychain/private value must represent the same
Ed25519 key pair. The private value is accepted as a raw 32-byte seed or
64-byte private key encoded as base64 or hexadecimal. Keep it in the
owner-controlled Keychain or another offline/key-management process; never
commit it, put it in a shell command, or print it in a terminal transcript.

## Local tag flow

1. Run the default validation loop locally: tests, race tests, vet, host build,
   and `git diff --check`.
2. Run `bash scripts/verify-release-source.sh v0.2.0` (or the exact release
   version). It fails closed on a non-Git, dirty, or untracked source tree, a
   stale generated `web/index.html`, an invalid version matrix, or drift
   between the canonical updater key and the Makefile, package scripts, Docker
   metadata, or standalone installers. It also derives the Ed25519 Subject
   Public Key Info (SPKI) encoding used by the Bash installer and compares it
   with the canonical raw key.
3. Confirm the release version in the tag, `CHANGELOG.md`, `Makefile`, and
   release notes. Do not use a tag as evidence that Google upstream behavior was
   live-tested.
4. Run `scripts/release-local.sh` from the checkout. It runs six CLI
   cross-builds and signs the resulting directory. For the macOS native
   candidate, use `make desktop-release-mac`; it embeds the public updater key
   and creates a stable-channel app, ZIP, DMG, notice, and unsigned local
   checksum manifest. Use `make desktop` for a normal stable desktop build, or
   `scripts/build-wails-local.sh` for an explicitly development-only macOS
   bundle when the checkout is under File Provider and in-place codesigning is
   affected by resource-fork metadata.
5. After the exact release directory is signed and locally verified, run
   `BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY=... scripts/record-release-evidence.sh
   /path/to/release-assets v0.2.0`. Keep the resulting 0600 receipt outside
   the worktree and release directory.
6. Upload the resulting files through the operator's chosen release channel,
   then publish the binaries plus `SHA256SUMS` and `SHA256SUMS.sig`.
7. Download the exact published assets into a fresh directory and rerun
   `scripts/verify-release-assets.sh` with the embedded public key. Compare the
   release tag, app version, channel, and release notice manually.
8. Verify the published asset names and manifest entries from a clean machine
   before announcing the release. Never replace the developer's running
   executable as a test.

The script refuses an existing output directory so stale artifacts cannot be
silently included in the signed manifest. It does not delete files.

For an already-built native artifact directory on macOS, use the no-Actions
signing wrapper after inspecting its contents. It reads the default Keychain
item without requiring the private key to be copied into the shell:

```bash
BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY=... \
	scripts/sign-release-assets.sh /path/to/release-assets
```

For a non-macOS secret-manager workflow, provide
`BOB_GEMINI_FREE_UPDATE_PRIVATE_KEY` only through that manager's protected
process environment. The private value must never be pasted into a repository
file or a student-facing command.

## Required external gates

The local script and source code do not prove release-host configuration.
Before treating a release as operationally ready, confirm that:

- the public/private values are present and are the same key pair;
- tag creation and release publication are restricted to authorized maintainers;
- the chosen release channel preserves the exact binary and manifest bytes;
- at least one native artifact opens and reports the actual gateway endpoint on
  each supported desktop device;
- a clean-machine updater check succeeds without using a developer executable.

## Artifact contract

Every supported CLI binary must have an exact manifest entry. The signed
manifest and signature are release assets, not files fetched from an untrusted
mirror. Updaters reject missing trust material, invalid signatures, manifest
mismatches, non-executable magic bytes, and downloads over 512 MiB.

This process proves artifact integrity and authenticity only when the operator
has configured the key pair and completed the clean-machine check. It does not
prove provider availability or client compatibility.
