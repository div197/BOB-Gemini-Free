# Release Process

The supported release path is the Go CLI plus the Wails desktop artifacts.
Wails is the only active native desktop implementation. The former alternate
wrapper was removed from the working tree after the Phase III comparison; its
source remains recoverable through Git history, but it is not part of releases.

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
