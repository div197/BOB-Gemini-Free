# Release Transition Audit — 2026-08-31

## Decision

The immutable `v0.2.0-preview.1` tag was not reused. Controlled macOS
`v0.2.0-preview.2` is now published from public-main commit `6d3a0cfc` and is
the current preview candidate. Stable `v0.2.0` remains a later release
candidate, not a release to publish until the device and pilot gates pass.

The exact publication and byte-reconciliation evidence is recorded in
[`PREVIEW-2-PUBLICATION-2026-08-31.md`](PREVIEW-2-PUBLICATION-2026-08-31.md).

The updater is working as a staged, user-consented migration mechanism. It is
not a silent fleet push: a running app may perform a delayed metadata check,
but it does not download, replace, or restart until the user selects
**Install Update**.

## Evidence snapshot before PR #42

This audit was run on 2026-08-31 from
`/Users/apple31/Documents/BOB-Gemini-Free` without using a PAT, provider key,
cookie, or release private key.

| Surface | Observed state | Meaning |
|---|---|---|
| Public `main` | `523ceeb51724dc4892c1870a4e8dd50f08916fb0` | The release-readiness branch is not `main`. |
| Reviewed source branch | `codex/release-readiness-v0.2.0` at `cec4c8e` | 92 commits ahead of public `main`; the latest source includes post-bridge Studio fixes, deterministic preview-transition fixtures, stream-cancellation ordering, strict release-asset validation, and the isolated history-limit regression. |
| Public fleet baseline | `v0.1.7-preview.7`, tag target `a5ec476` | Its released desktop updater checks the preview channel only. |
| Public migration bridge | `v0.2.0-preview.1`, tag `e019cf8` | Its macOS universal release is published with the current project key and signed manifest. |
| Public stable endpoint | `v0.1.5` | No stable `v0.2.0` is published yet; a Preview build therefore continues to the preview channel. |
| Next source candidate | `v0.2.0-preview.2` | Defaulted by all three Wails preview packagers; not tagged or published. |

The public GitHub release API listed `v0.2.0-preview.1` and
`v0.1.7-preview.7`, but no `v0.2.0-preview.2` release. The public
`v0.2.0-preview.1` DMG, ZIP, release notice, `SHA256SUMS`, and
`SHA256SUMS.sig` were downloaded into a temporary directory and passed the
repository's read-only `scripts/verify-release-assets.sh` check with the
canonical public key from `UPDATE-PUBLIC-KEY.txt`. The downloaded SHA-256
values were:

```text
bob-gemini-free-macos-universal.dmg  c7078edf83d98f02720b5762bb8013c9528f678d8af3b9177057e5eec6fc46cb
bob-gemini-free-macos-universal.zip  2cca2551020b4fb7938d55081f9160d3993fdbda9e3c4871e22cfe894a3606dc
RELEASE-NOTICE.txt                   efa4907841a16a96f3931a9331e779fb9ae2c7964ab3b0e03ec40e5456e583ca
SHA256SUMS                           56b1d9d4bcb005caa300f5221de9d90c4a5a030003d0462eb7e27eb6cee198f8
SHA256SUMS.sig                       18b969bad7d3de108d135e34d845a6f98385ff92c00e510a0a26423565c3e79d
```

## Post-merge reconciliation — 2026-08-31 (before Preview 2 publication)

Protected PR [#42](https://github.com/div197/BOB-Gemini-Free/pull/42) merged
the reviewed source and release-readiness documentation into public `main` at
`ba1b56228f999bcded0fc6539ddb8ccca1935a11`. The reviewed code tip `cec4c8e`
and the final documentation-only reconciliation commits are therefore part of
the public source history. At that point this merge did not create a tag or
GitHub Release: public `v0.2.0-preview.1` remained immutable, Preview 2 was a
locally verified but unpublished candidate, and stable `v0.2.0` remained
unpublished.
No GitHub Actions workflow was added or run.

## Follow-up updater durability hardening — 2026-08-31

The isolated follow-up commit `b136724` keeps the release trust and migration
design unchanged while strengthening the local filesystem transaction. Plan,
confirmation, failure, and warning records now use flushed private temporary
files followed by same-directory replacement. Backup, candidate activation,
rollback, and startup-recovery directory transitions flush on Unix; a focused
fault-injection test proves candidate activation is rolled back when that flush
fails. This is local source/test evidence, not a physical power-loss or clean
installed-bundle acceptance result, and it does not alter the no-Actions or
user-consent requirements. The follow-up Windows-specific metadata path uses
native `MoveFileExW` replace-existing/write-through semantics in `fd279aa`.

Protected PR [#45](https://github.com/div197/BOB-Gemini-Free/pull/45) and PR
[#46](https://github.com/div197/BOB-Gemini-Free/pull/46) subsequently placed
the provenance reconciliation and Windows updater hardening on public
`main`. The local branch was rechecked against the public file tree before
this document refresh. No new tag or release was created by those merges.

## Preview 2 publication reconciliation — 2026-08-31

The owner-controlled local release operation built the universal macOS
package from public-main commit `6d3a0cfc0a7a0bf05a3c136baf96a48f503b45ef`,
signed its exact manifest through the macOS Keychain, and published
`v0.2.0-preview.2` as a prerelease without GitHub Actions. GitHub's five
uploaded assets were downloaded again; the detached signature verified and
the downloaded files matched the local inputs byte-for-byte. The release is
macOS-only, ad-hoc signed, and not notarized.

This closes publication and public-byte reconciliation for Preview 2. It does
not close the clean-device replacement, rollback, Apple trust, provider, or
30-device rollout gates.

The public Preview 2 ZIP reports `CFBundleShortVersionString` and
`CFBundleVersion` as `0.2.0`; the full preview channel/version is carried by
the injected desktop updater version and release metadata. This is expected
macOS bundle-version normalization, but the Help/About surface and release
notice must continue to show the full `v0.2.0-preview.2` identity.

## Migration matrix

| Installed build | Metadata path | Candidate it can discover | What is actually proven |
|---|---|---|---|
| Public `v0.1.7-preview.7` | Preview list only | Published `v0.2.0-preview.2` directly, with the same key and compatible macOS asset | Current regression fixture, public manifest/signature, and public-byte reconciliation are verified; installed replacement remains a device gate. |
| Public `v0.2.0-preview.1` | Stable first, then preview list | Stable `v0.2.0` when published; otherwise a newer preview such as `v0.2.0-preview.2` | Stable-first and preview-continuation selection are covered by updater tests. Clean-device replacement is still open. |
| Published `v0.2.0-preview.2` | Stable first, then preview list | Stable `v0.2.0` when published; otherwise later previews | Public-main package passed manifest/signature/checksum, bundle, launch, health, shutdown, fresh-download, and byte-reconciliation checks; clean-device replacement remains open. |
| Stable `v0.2.0` | Stable endpoint only | A newer stable release | Stable builds do not move backward into preview. |

The safe operator sequence for a device currently on Preview 7 is therefore:

```text
Preview 7 → published same-key Preview 2 (or Preview 1 bridge if already selected)
          → stable v0.2.0 after stable publication and acceptance
```

The safe sequence for a device already on Preview 1 is:

```text
Preview 1 → Preview 2 when the new preview is published
          → stable v0.2.0 when the stable release is published
```

If stable is published before Preview 2, Preview 1 can move directly to
stable because the bridge checks stable first. In every case the user must
confirm the exact candidate, and the manifest, package, path, restart, and
health/rollback gates must succeed.

## Regression locks added in this audit

- `v0.2.0-preview.1 → v0.2.0` stable-first selection is explicitly covered.
- `v0.2.0-preview.1 → v0.2.0-preview.2` preview continuation is explicitly
  covered.
- A legacy Preview 7 preview-only lookup is explicitly covered for the next
  preview candidate and for the current published bridge.
- Semantic ordering covers preview-to-preview and preview-to-stable
  transitions.
- Preview packager defaults now advance to `v0.2.0-preview.2`, preventing a
  changed source tree from being accidentally rebuilt under the published
  Preview 1 identity.

These are deterministic metadata and selection proofs. They do not claim that
the updater has replaced an installed `/Applications` bundle, survived a
power interruption, passed Gatekeeper without user approval, or completed a
30-device rollout.

## Release gate

Do not publish stable yet. Before a stable tag, the owner still needs to:

1. re-download and verify the published Preview 2 candidate from a clean
   writable Mac;
2. exercise Preview 7 → Preview 2 (or the already-published Preview 1 bridge)
   and Preview 1 → Preview 2 on real installed bundles;
3. exercise Preview 2/Preview 1 → stable with a controlled stable candidate;
4. deliberately invalidate/interruption-test rollback without losing config,
   cookies, preferences, or history;
5. repeat on two or three pilot Macs before any 20–30 device wave; and
6. keep Apple Developer ID/notarization and Windows publisher signing as
   separate platform-trust gates.

The project signing key authenticates release bytes. It does not remove the
macOS first-launch warning, create Apple publisher trust, or prove Google
availability, quota, model identity, or unlimited use.
