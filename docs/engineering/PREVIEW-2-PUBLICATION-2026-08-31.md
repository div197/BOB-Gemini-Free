# Preview 2 Publication Record — 2026-08-31

## Published identity

- Release: [`v0.2.0-preview.2`](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.2)
- Tag target: `6d3a0cfc0a7a0bf05a3c136baf96a48f503b45ef` (the public `main` merge commit)
- Channel: `preview`
- Scope: macOS universal desktop package only
- GitHub Actions: not used; the package was built and signed locally
- Trust status: project Ed25519 manifest signature plus ad-hoc macOS bundle signature; no Apple Developer ID signature or notarization

The package was built from the clean public-main tree after the updater
durability and documentation merges. The private release key was read only by
the owner-controlled macOS Keychain signing command; it was not copied into
the repository, package, release note, shell output, or student instructions.

## Public asset reconciliation

The release assets were downloaded again from GitHub into a fresh temporary
directory. `scripts/verify-release-assets.sh` accepted the detached signature,
and every downloaded file matched the locally signed file byte-for-byte.

| Asset | GitHub-served size | GitHub-served SHA-256 |
|---|---:|---|
| `bob-gemini-free-macos-universal.dmg` | 20,526,956 bytes | `25334d1c0a34e6f5124f868586d46fc4d239493c951b519344e573bc1d96ddfd` |
| `bob-gemini-free-macos-universal.zip` | 18,987,856 bytes | `e7ed278cd0ca0a089f3f534419e10c10960ec515de3e2080f8179d03e61ef7dd` |
| `RELEASE-NOTICE.txt` | 1,034 bytes | `84079a201f5857da8b5bc86dcce75dbb352e364586070b214f3e4e5ecf6edab8` |
| `SHA256SUMS` | 289 bytes | `ff01871ee0325abd51fc4bb805e9006630d799354d75567d4e0e220d0d776832` |
| `SHA256SUMS.sig` | 89 bytes | `1e1e50464bec721eb09c46f710244bf3ca1c582bf3fe1c889cfc169065f2f53b` |

The uploaded tag is a lightweight tag resolving directly to the public-main
commit above. The release is a prerelease and was not marked as a stable
release.

## Local acceptance evidence

The exact release candidate passed these checks before publication:

- `scripts/verify-release-source.sh v0.2.0-preview.2`
- `make desktop-key-check`
- `go test -count=1 ./internal/updater`
- Keychain-backed `scripts/sign-release-assets.sh` followed by
  `scripts/verify-release-assets.sh`
- ad-hoc `codesign --verify --deep --strict` on the universal `.app`
- DMG layout verification with a visible `Applications` shortcut
- fresh-bundle launch and clean shutdown
- local `GET /healthz`, `/manifest.json`, `/sw.js`, and `/favicon.ico` smoke
  checks, all returning the expected successful response

This is source/package evidence. It is not proof that a clean student's Mac
will pass Gatekeeper without approval, that an installed bundle will complete
replacement and health confirmation, or that Google will accept a generation
request.

## Installed-base meaning

An existing macOS `v0.1.7-preview.7` installation that contains the same
project public key and runs from a writable application directory can now
discover `v0.2.0-preview.2` through its preview-channel lookup. The user must
still choose **Help → Check for Updates**, accept the update, and approve any
macOS first-launch warning. The new bridge then provides stable-first lookup
for a later stable release.

The source and public asset path are verified; the actual replacement,
restart, rollback, and preservation of each device's local configuration are
still clean-device and pilot gates. Do not announce a 30-device rollout as
complete until those actions have been observed and recorded.

## Still deliberately gated

- stable `v0.2.0` publication;
- Apple Developer ID signing, hardened runtime, notarization, and stapling;
- Windows publisher-signed and Linux-native package acceptance;
- clean-device update and rollback runs;
- real browser/PNA and artifact dependency acceptance;
- live Google session/API-key, quota, and model behavior;
- two- or three-device pilot followed by a staged 20–30 device rollout.
