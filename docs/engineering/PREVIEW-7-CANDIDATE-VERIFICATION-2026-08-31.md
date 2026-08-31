# v0.2.0-preview.7 — Local Candidate Verification

**Date:** 2026-08-31 (Asia/Kolkata)
**Status:** locally packaged, signed, and verified; **not published**
**Source:** `main` / `2d42d4443818f594bc8d27f906977b30acce7495`
**Public baseline:** `v0.2.0-preview.6` remains the current published macOS
preview.

This receipt separates a release candidate from a downloadable public release.
The candidate was built only after the worktree was clean and synchronized with
`origin/main`. No GitHub Actions, provider credential, Google session, cookie,
or private release-key export was used.

## Source and signing gates

The local candidate was produced with:

```text
BOB_RELEASE_VERSION=v0.2.0-preview.7
BOB_RELEASE_CHANNEL=preview
BOB_WAILS_PLATFORM=darwin/universal
```

The following checks passed:

| Check | Result |
|---|---|
| `scripts/verify-release-source.sh v0.2.0-preview.7` | PASS |
| Wails packaging | PASS; Wails `v2.15.0`, macOS universal |
| App bundle signature | PASS; ad-hoc signature valid on disk |
| Ed25519 release-manifest signing | PASS; owner-controlled macOS Keychain signer |
| `scripts/verify-release-assets.sh` | PASS; exact signed directory |
| macOS DMG layout | PASS; app plus `/Applications` shortcut |
| `lipo -info` | PASS; `x86_64` and `arm64` |
| `go test` updater transition matrix | PASS |

The private signing value was not displayed, exported, copied, committed, or
placed in the package. The signed asset receipt was written outside the
worktree at audit time; its manifest SHA-256 was
`bbb2a27bc73c7c7b745eddb75d5d855a843d114b0e3ddb42bc3a2db19e0a4219`.

## Bundle identity and runtime proof

The candidate has two intentionally different version surfaces:

- macOS `CFBundleShortVersionString` and `CFBundleVersion`: numeric base
  version `0.2.0`, suitable for the platform bundle metadata;
- injected desktop/updater/UI identity: `v0.2.0-preview.7`, served through
  `X-BOB-Version`, the About surface, and the local HTML bundle.

The updater must use the injected channel-aware identity, not Finder's numeric
bundle field, when deciding whether a preview update exists.

The bundled executable was started directly from the candidate app on an
isolated loopback port (`127.0.0.1:18082`) while the installed app was left
untouched. The candidate:

- returned HTTP 200 from `/healthz` with `X-BOB-Version:
  v0.2.0-preview.7`;
- served `/playground`, `/manifest.json`, `/sw.js`, and `/favicon.ico`;
- rendered the injected `v0.2.0-preview.7` identity in the playground; and
- shut down cleanly after the smoke test.

No generation request was made and no Google cookie, web session, or Developer
API key was entered. This proves package startup and local serving, not live
Google availability.

## Hermetic Studio artifact lifecycle smoke

The current local Studio shell was exercised with a loopback-only synthetic
SSE server; no Google request or credential was involved. The fixture streamed
an HTML code fence, the Studio registered one artifact, opened its sandboxed
preview, and then switched to **Code**. The editor contained the complete
source (`2 lines / 58 chars`) and the preview iframe received the source. This
reproduces the user-visible lifecycle without treating the older empty-editor
screenshot as current behavior; the source regression remains locked by
`TestArtifactEditorPreservesGeneratedSource`.

## Installed-base updater evidence

Hermetic updater tests passed for:

- legacy `v0.1.7-preview.7` discovering a same-key `v0.2.0` preview;
- earlier `v0.2.0-preview.1` through Preview 5 clients selecting the newest
  signed preview when stable has no newer release;
- current-preview stable-first migration; and
- staged replacement, lock serialization, rollback, confirmation, and
  read-only installation failures.

These tests prove selection and transaction rules, not an installed-device
replacement. Since Preview 7 is not published, existing devices currently see
the published Preview 6 candidate, subject to their compiled updater behavior,
writable install location, signature key, network, and explicit consent.

## Public release truth

GitHub currently reports `v0.2.0-preview.6` as a published prerelease at source
target `f9b3410e74d7ccc08487dc03788b54a201e12ade`, with five uploaded assets.
GitHub's release metadata currently reports `immutable: false`; the project
therefore treats published release identities and their assets as operationally
write-once, but does not claim that GitHub has applied an immutable-release
lock. A future publication must use the new Preview 7 tag and exact signed
bytes, followed by fresh public download and verification.

There is no repository-authored `.github/workflows` directory. The GitHub API
may list its platform-managed `pages-build-deployment` entry; it is not part of
the BOB release process and no Actions workflow was used for this candidate.

## Gates that remain open

- public Preview 7 publication and post-download byte reconciliation;
- one installed Preview 6 → Preview 7 replacement on a writable `/Applications`
  Mac;
- deliberate failed-candidate rollback and clean-device acceptance;
- Apple Developer ID, notarization, and Gatekeeper trust;
- Windows/Linux native builds and platform trust;
- live Google/session/Developer-API behavior and provider limits; and
- staged two- or three-device pilot before any 20–30-device rollout.

Until those gates are observed, this candidate must not be described as a
production or unattended fleet release.
