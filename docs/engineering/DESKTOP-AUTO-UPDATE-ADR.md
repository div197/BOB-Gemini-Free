# ADR: Secure Native Desktop Updates

**Status:** Accepted; source implementation and mocked security tests added

**Date:** 2026-08-22

## Context

The native application is distributed as a macOS app bundle and a Windows
executable. The existing public Preview 7 is an ad-hoc,
project-manifest-signed preview whose released binary performs an explicit
preview-channel metadata check. The public `v0.2.0-preview.2` package is the
current same-key controlled macOS preview and contains the later stable-first
migration path for newly built previews. It must not silently replace itself.

The repository already has a fail-closed Ed25519 verifier for CLI update
assets. Native updates need a separate package-aware path because replacing a
running macOS bundle or Windows executable has different filesystem and
publisher-trust rules. A SHA-256 value by itself is only an integrity check;
the manifest signature and embedded public key are the authenticity boundary.

## Decision

Build a user-consented, transaction-based native updater in these stages:

1. **Discover:** query only the fixed official GitHub repository and select a
   platform-matching release from the build's channel policy. Stable builds use
   the stable channel only. Preview builds check stable first to permit an
   intentional one-way migration into a newer stable release, then check the
   bounded preview channel only when no stable update exists; they never treat a
   prerelease as stable or silently switch channels.
2. **Authorize:** show the release, version, channel, and package type to the
   user. No package download or replacement occurs until the user confirms.
3. **Verify:** download the exact desktop archive/executable, `SHA256SUMS`,
   and `SHA256SUMS.sig`; verify the Ed25519 signature with the public key
   embedded at build time; verify the exact asset digest, size limit, URL host,
   and platform format. Missing trust material fails closed.
4. **Stage:** write the verified candidate beside the installed application
   on the same filesystem. For macOS, safely extract one expected `.app`
   bundle from the ZIP and verify its code signature. Never execute staged
   content before verification.
5. **Replace:** launch a short-lived helper copied from the current executable,
   quit the running app, atomically move the old install to a rollback backup,
   flush the relevant Unix directory entry, move the candidate into place,
   flush again, and relaunch the candidate. The helper keeps the backup until
   the new app sends a local health confirmation. Windows retains its
   platform-specific rollback boundary; updater metadata uses native
   `MoveFileExW` replace-existing/write-through semantics because there is no
   portable directory fsync contract.
6. **Recover:** if replacement, launch, or confirmation fails, restore the
   backup and leave a human-readable failure record. The manual GitHub release
   path remains available at every stage.

The native background metadata check waits 30 seconds after startup plus a
per-process random jitter capped at five minutes, and then runs at most once
per day. This spreads release metadata traffic from a classroom restart while
keeping the check non-installing and user-consented.

## Trust model

- The GitHub API URL, release URL, and asset hosts are fixed in source.
- The Ed25519 public key is a build trust anchor; it is never fetched from
  GitHub or mutable user configuration.
- The signed manifest authenticates release bytes. It does not replace Apple
  Developer ID/notarization or Windows publisher signing.
- macOS production packages must pass Developer ID signature and notarization
  checks. Windows production packages must use publisher-signed installer or
  executable artifacts. Those platform gates remain external until the owner
  enables them.
- User config, cookies, API keys, chat history, and gateway data are never
  part of the update candidate or rollback operation.

## Why this design

This gives the student workflow the useful part of modern update systems—one
consented update action, background-safe staging, restart, and recovery—while
keeping the release trust chain explicit. It avoids a large third-party update
framework, keeps the native runtime dependency-free, and reuses the existing
manifest verifier and manual GitHub release process.

## Deliberate non-goals

- No silent updates in the preview or stable channel.
- A signed preview may perform a delayed, low-frequency metadata check, but it
  must never download, replace, or restart without the same explicit consent.
- No download or replacement during ordinary app startup before the signed
  production channel is proven.
- No update from an arbitrary URL, custom JSON endpoint, or user-supplied key.
- No replacement test against the developer's running executable.
- No claim that manifest signing alone removes Gatekeeper or SmartScreen
  warnings.

## Release gate

The updater code may be developed and tested with mocked signed fixtures now.
It becomes eligible for a student-facing production channel only after the
  owner has configured the authoritative signing key, published signed desktop
  manifests, completed Apple/Windows publisher signing, and passed clean-device
  rollback tests. Preview-to-stable migration remains explicit and
  user-consented. Existing public Preview 7 installations use the published
  same-key Preview 2 bridge or a manual stable install because the released
  binary predates stable-first discovery; builds without the current trust anchor
  still need a manual migration. See
`DESKTOP-UPDATE-OPERATIONS.md` for the operator and classroom rollout gate.
