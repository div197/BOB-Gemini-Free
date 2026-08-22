# Release Process

The supported release path is the Go CLI plus the Wails desktop artifacts.
Wails is the only active native desktop implementation. The former alternate
wrapper was removed from the working tree after the Phase III comparison; its
source remains recoverable through Git history, but it is not part of releases.

## Free open-source preview path

The project can build a macOS Wails preview without an Apple Developer
membership:

```bash
make desktop-preview-mac
```

This path creates an ad-hoc-signed `.app`, `.zip`, and `.dmg` plus a release
notice and local checksums. It is explicitly not Developer ID signed,
notarized, or production-ready. It is intended for controlled evaluation only.
It does not replace the signed release process below and must not be uploaded
as a trusted student release without the warning notice.

The Wails Help menu's update action is also intentionally user-initiated. It
checks the official GitHub release metadata and opens the official release page
when a matching package exists; it does not silently replace a running native
bundle.

## Native automatic-update status

The current Wails preview does **not** auto-update. It performs an explicit
stable-release metadata check and lets the user open the official release page.
Preview tags are deliberately excluded from that check.

A safe automatic native updater is a separate release-system project. It needs
platform-signed artifacts, trusted signed update metadata, a user-consent
boundary, version/channel policy, atomic replacement, rollback, and clean
device verification. macOS Developer ID/notarization and Windows publisher
signing are prerequisites for treating that path as a professional student
distribution mechanism. Until those gates are complete, silently replacing a
running preview would weaken the trust model rather than improve it.

The implementation sequence is:

1. publish a stable and preview channel contract for native packages;
2. extend the signed-manifest verifier to the exact `.app` archive and
   Windows installer/executable asset for each platform;
3. add a platform-specific post-exit helper with atomic replacement, rollback,
   permission failure reporting, and restart behavior;
4. add mocked download/tamper/rollback tests and clean-device acceptance;
5. enable background checks only after explicit consent, while keeping a
   visible “Check for Updates” action and a manual release fallback.

This removes repeated delete/download/install work only after the release and
device gates above are complete. It is not enabled by the current preview.

## Preconditions

Set these values only in the local release environment:

| Name | Purpose |
|---|---|
| `BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY` | Ed25519 public key embedded in CLI binaries and checked at signing time |
| `BOB_GEMINI_FREE_UPDATE_PRIVATE_KEY` | Ed25519 private key used only by `cmd/release-manifest` |

The two values must represent the same Ed25519 key pair. The private value is
accepted as a raw 32-byte seed or 64-byte private key encoded as base64 or
hexadecimal. Keep the private key in an offline/key-management process and
the local environment; never commit it or print it in a terminal transcript.

## Local tag flow

1. Run the default validation loop locally: tests, race tests, vet, host build,
   and `git diff --check`.
2. Confirm the release version in the tag, `CHANGELOG.md`, `Makefile`, and
   release notes. Do not use a tag as evidence that Google upstream behavior was
   live-tested.
3. Run `scripts/release-local.sh` from the checkout. It runs six CLI
   cross-builds and signs the resulting directory. Use `make desktop` for a
   normal Wails developer build, or `scripts/build-wails-local.sh` on macOS
   when the checkout is under File Provider and in-place codesigning is
   affected by resource-fork metadata.
4. Upload the resulting files through the operator's chosen release channel,
   then publish the binaries plus `SHA256SUMS` and `SHA256SUMS.sig`.
5. Verify the published asset names and manifest entries from a clean machine
   before announcing the release. Never replace the developer's running
   executable as a test.

The script refuses an existing output directory so stale artifacts cannot be
silently included in the signed manifest. It does not delete files.

## Required external gates

The local script and source code do not prove release-host configuration.
Before treating a release as operationally ready, confirm that:

- the public/private values are present and are the same key pair;
- tag creation and release publication are restricted to authorized maintainers;
- the chosen release channel preserves the exact binary and manifest bytes;
- at least one Wails artifact opens and reports the actual gateway endpoint on
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
