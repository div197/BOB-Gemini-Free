# Preview 3 Publication Record — 2026-08-31

## Published identity

- Release: [`v0.2.0-preview.3`](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.3)
- Tag target: `284b7d1a9a2e7c45402318f29f08f0c1dba36d43` (public `main`)
- Channel: `preview`
- Scope: macOS universal desktop package only
- GitHub Actions: not used; the package was built and signed locally
- Trust status: project Ed25519 manifest signature plus ad-hoc macOS bundle
  signature; no Apple Developer ID signature or notarization

## Source and key custody

The release source gate passed for `v0.2.0-preview.3` on the clean public-main
tip. The local Keychain-backed signer derived the expected public key from the
private release key and signed the exact release manifest. The private key was
never printed, copied into the repository, embedded in the application, added
to release notes, or included in the uploaded assets.

The public source changes preceding publication were merged through protected
PRs [#53](https://github.com/div197/BOB-Gemini-Free/pull/53),
[#54](https://github.com/div197/BOB-Gemini-Free/pull/54), and
[#55](https://github.com/div197/BOB-Gemini-Free/pull/55). The final source
tip is public `main` commit `284b7d1`; no source-only change remains on the
local release branch.

## Local acceptance evidence

- `go test -count=1 ./...` passed.
- `go vet ./...` passed.
- `bash scripts/verify-release-source.sh v0.2.0-preview.3` passed.
- The universal Wails app passed strict ad-hoc `codesign --verify --deep
  --strict` verification.
- The ZIP passed archive integrity testing and contains the branded
  `BOB Gemini Free.app` root; the DMG contains the branded app and visible
  `/Applications` shortcut.
- A fresh launch of the exact artifact bound to loopback `127.0.0.1:8081`
  because port 9610 was occupied, returned `{"status":"ok"}` from
  `/healthz`, and shut down cleanly.
- The package release directory passed `scripts/verify-release-assets.sh`
  before publication.

## Public asset reconciliation

All five public assets were downloaded into a fresh temporary directory after
publication. The detached signature verified, and every downloaded file
matched the locally signed input byte-for-byte.

| Asset | Public size | Public SHA-256 |
|---|---:|---|
| `bob-gemini-free-macos-universal.dmg` | 20,532,549 bytes | `dde9466255f76a0026cc1958aadf17e6cc9442f0632c98597bbbbf1a35e98a81` |
| `bob-gemini-free-macos-universal.zip` | 18,991,248 bytes | `2b268c191930ead119a7e26302f40c80449101d1d43c1aeb8f36db7080d1c38a` |
| `RELEASE-NOTICE.txt` | 1,257 bytes | `3824c7d36a30ca36795b4419bae4a0656b14a85830db614453d27c80a27f39c6` |
| `SHA256SUMS` | 289 bytes | `07af99ca76050cf9bf56b7709031b59277189bd752d3bbfbabbc71d6b59c85ce` |
| `SHA256SUMS.sig` | 89 bytes | `17744607aeb7ac4effba29d22c918576ade9f9da465829a7fa0111c20b2f76d2` |

The `SHA256SUMS` manifest covers the release notice, DMG, and ZIP. Its own
signature is a detached Ed25519 signature over the exact manifest bytes.

## Installed-base meaning

The public metadata now exposes `v0.2.0-preview.3` as the latest compatible
same-key macOS preview candidate. A current-source preview can discover a
newer stable release first, or a newer preview when stable has no update. The
legacy `v0.1.7-preview.7` updater still uses its preview-only lookup and can
only be treated as having a public candidate available; actual installed-bundle
replacement, restart, rollback, and local-data preservation remain device
acceptance gates.

The updater is not a silent fleet-push mechanism. Each device must be in a
writable application directory, the user must confirm the exact candidate,
and macOS may require the normal first-launch **Open Anyway** approval because
the package is not notarized.

## Still deliberately gated

- clean-device update from the existing Preview 7/Preview 2 installations;
- deliberate interrupted-candidate rollback with config, cookies, preferences,
  and history preserved;
- Apple Developer ID signing, hardened runtime, notarization, and stapling;
- Windows publisher signing and Linux package acceptance;
- live Google session/API-key, quota, model, and upstream-IP behavior; and
- a two- or three-device pilot followed by any 20–30-device rollout.

The project signing key authenticates project release bytes. It does not
remove platform warnings, create Apple publisher trust, or guarantee Google
availability, free-tier capacity, or unlimited use.
