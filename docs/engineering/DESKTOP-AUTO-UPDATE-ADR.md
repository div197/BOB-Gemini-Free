# ADR: Secure Native Desktop Updates

**Status:** Accepted; source implementation and mocked security tests added

**Date:** 2026-08-22

## Context

The native application is distributed as a macOS app bundle and a Windows
executable. The existing public Preview 7 is an ad-hoc,
project-manifest-signed preview whose released binary performs an explicit
preview-channel metadata check. The public `v0.2.0-preview.3` package is the
current same-key controlled macOS preview and contains the later stable-first
migration path for newly built previews. It must not silently replace itself.

The repository already has a fail-closed Ed25519 verifier for CLI update
assets. Native updates need a separate package-aware path because replacing a
running macOS bundle or Windows executable has different filesystem and
publisher-trust rules. A SHA-256 value by itself is only an integrity check;
the manifest signature and embedded public key are the authenticity boundary.

## Decision

Build a user-consented, transaction-based native updater in these stages:

1. **Discover:** read the fixed, detached-signed desktop feed first and select
   a platform-matching release from the build's channel policy. Stable builds
   use the stable channel only. Preview builds check stable first to permit an
   intentional one-way migration into a newer stable native package, then
   check the bounded preview channel when no newer native package exists for
   the current platform. A stable CLI-only release must not mask a native
   preview; the updater never treats a prerelease as stable or silently
   switches channels. If the feed is unavailable, expired, or fails signature
   validation, the existing fixed GitHub API path is used as a discovery
   fallback.
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
keeping the check non-installing and user-consented. An explicit Help → Check
for Updates action bypasses the feed and requests fresh official release
metadata, so the visible manual check does not rely on a still-valid snapshot.

## Discovery availability decision — 2026-09-05

Thirty devices can otherwise create a burst of unauthenticated GitHub REST
requests: a preview check may query stable and preview metadata separately.
The source now supports one small `updates/desktop-feed.json` document and a
detached `updates/desktop-feed.json.sig` signature, fetched as two small
resources from the official raw GitHub content host. The feed contains release
metadata only; the selected package, `SHA256SUMS`, and `SHA256SUMS.sig` are
still fetched from the official release and verified independently before
staging.

The feed is pinned to exact URLs, bounded in size, signed with the same
project Ed25519 key, and rejected when its validity window expires or is
malformed. It is therefore an availability optimization, not a new trust
anchor. A bad or unreachable feed cannot authorize an install. The API
fallback preserves compatibility with already-published binaries and gives a
manual check a recovery path when the feed commit or raw CDN is unavailable.

Every future published desktop release must refresh the feed after the public
release assets have been reconciled, sign it locally with
`scripts/sign-update-feed.sh`, and commit the feed plus detached signature to
`main`. The release process must never advertise a version that has not been
published. Binaries already in the field cannot gain this discovery code
retroactively; they continue using the updater path compiled into their
release.

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
  same-key current preview or a manual stable install because the released
  binary predates stable-first discovery; builds without the current trust anchor
  still need a manual migration. See
`DESKTOP-UPDATE-OPERATIONS.md` for the operator and classroom rollout gate.
