# Release Update Verification — Mission 3

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

The trusted Ed25519 public key is supplied through
`BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY` as base64 or hexadecimal. It is deliberately
not fetched from GitHub. If the key, manifest, signature, asset entry, or
signature verification is missing or invalid, `--update` fails closed before
the binary can replace the installed executable.

## Replacement boundary

The candidate is downloaded into a temporary file in the executable's own
directory, verified against the signed manifest, checked for platform binary
magic, and only then considered for replacement. Unix replacement uses an
atomic same-filesystem rename. Windows retains the existing rollback path
because a running executable cannot be renamed over in place.

Tests verify manifest signatures and mocked downloads into a temporary
candidate path. They never call replacement against `os.Executable()` or the
developer's running binary.

## Remaining release gate

This source snapshot contains no authoritative release public key and no
signed `SHA256SUMS` assets. Therefore the implementation can enforce the
contract and safely reject unconfigured updates, but a live successful update
cannot be claimed until the release pipeline publishes the assets and the
operator configures the matching public key.
