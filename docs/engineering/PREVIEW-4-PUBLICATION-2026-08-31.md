# Preview 4 Publication Record — 2026-08-31

## Published identity

- Release: [`v0.2.0-preview.4`](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.4)
- Tag target: `abfeebaaaaabc740ea29602b602591a0b707fbc2` (public `main`)
- Channel: `preview`
- Scope: macOS universal desktop package only
- GitHub Actions: not used; the package was built, signed, and uploaded locally
- Trust status: project Ed25519 manifest signature plus ad-hoc macOS bundle
  signature; no Apple Developer ID signature or notarization

## Source and key custody

The release-source gate passed for `v0.2.0-preview.4` on the clean public-main
tip. The local Keychain-backed signer derived the expected public key from the
private release key and signed the exact release manifest. The private key was
never printed, copied into the repository, embedded in the application, added
to release notes, or included in the uploaded assets.

The source changes preceding publication were merged through protected PRs
[#58](https://github.com/div197/BOB-Gemini-Free/pull/58) and
[#59](https://github.com/div197/BOB-Gemini-Free/pull/59). The final source tip
is public `main` commit `abfeeba`; no source-only change remains unmerged from
the release branch.

## Local acceptance evidence

- `go test -count=1 ./...` passed.
- `go test -race -count=1 ./...` passed.
- `go vet ./...`, `go mod verify`, and `go build ./...` passed.
- `make build` passed and reported the expected base version `v0.2.0`.
- `bash scripts/verify-release-source.sh v0.2.0-preview.4` passed.
- The universal Wails app passed strict ad-hoc `codesign --verify --deep
  --strict` verification and contained arm64 and x86_64 slices.
- The ZIP passed archive inspection and contains the branded
  `BOB Gemini Free.app` root; the DMG contains the branded app and visible
  `/Applications` shortcut.
- A fresh launch of the exact artifact bound to loopback `127.0.0.1:8081`
  because port 9610 was occupied, returned 200 from `/healthz`, rendered
  `v0.2.0-preview.4` in `/playground`, and shut down cleanly.
- The exact local signed release directory passed
  `scripts/verify-release-assets.sh` before publication.

## Public asset reconciliation

All five public assets were downloaded into a fresh temporary directory after
publication. The detached signature verified, and every downloaded file
matched the locally signed input byte-for-byte.

| Asset | Public size | Public SHA-256 |
|---|---:|---|
| `bob-gemini-free-macos-universal.dmg` | 20,539,337 bytes | `1a1dcbc7d80b49f788077e8add06db9b1e2c76a84dc427604962e935b93ef4d6` |
| `bob-gemini-free-macos-universal.zip` | 18,995,018 bytes | `68b4a3a2a7a89569e925a784ea39199de89022d05a25c9082df4a312eda48320` |
| `RELEASE-NOTICE.txt` | 1,257 bytes | `897d306fa6c315101f20c86c8bcf7c9e036c7c281967a9ae051f03b7bee10625` |
| `SHA256SUMS` | 289 bytes | `1f177c7a3dc47f6a02e1651003373b8da567b1f7273cb03f38cc8c1160c13890` |
| `SHA256SUMS.sig` | 89 bytes | `6ba8ca37ddf34845251ab3c31d385cac89d2faf0462f9974ed057264f4532126` |

The `SHA256SUMS` manifest covers the release notice, DMG, and ZIP. Its own
signature is a detached Ed25519 signature over the exact manifest bytes.

## Installed-base meaning

The public metadata now exposes `v0.2.0-preview.4` as the latest compatible
same-key macOS preview candidate. An installed public `v0.1.7-preview.7`
binary can discover this newer preview through its preview-only lookup. After
that explicit update, the Preview 4 updater can check stable first when a
stable release is eventually published. A direct stable install remains the
simpler alternative during this preview phase.

The updater is not a silent fleet-push mechanism. Each device must be in a
writable application directory, the user must confirm the exact candidate,
and macOS may require the normal first-launch **Open Anyway** approval because
the package is not notarized.

## Still deliberately gated

- clean-device update from the existing Preview 7 installations;
- deliberate interrupted-candidate rollback with config, cookies, preferences,
  and history preserved;
- Apple Developer ID signing, hardened runtime, notarization, and stapling;
- Windows publisher signing and Linux package acceptance; and
- a two- or three-device pilot followed by any 20–30-device rollout.

The project signing key authenticates project release bytes. It does not
remove platform warnings, create Apple publisher trust, or guarantee Google
availability, free-tier capacity, or unlimited use.
