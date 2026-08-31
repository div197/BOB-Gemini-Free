# Preview 5 Public Publication and Installed Update — 2026-08-31

## Decision

`v0.2.0-preview.5` is the current immutable macOS universal prerelease:

[`github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.5`](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.5)

The release targets public `main` commit
`c28d78736eaae436cc1f1f3b4ec6e0bbcd058b89`. The signed package was built from
clean checkout `88f288151b35273bb2ba06b770c8da0050b9e8e4`; its source tree
matches the release target. The difference is release-branch/merge history and
does not change the packaged source tree.

No GitHub Actions workflow was created or invoked. The release private key was
read only through the owner-controlled macOS Keychain signer; it was not
exported, printed, committed, or placed in the assets.

## Public asset reconciliation

All five assets were downloaded into a fresh temporary directory and passed
`scripts/verify-release-assets.sh` with the checked-in Ed25519 public key. Each
downloaded file matched the exact local signed input byte-for-byte.

| Asset | Size | SHA-256 |
|---|---:|---|
| `RELEASE-NOTICE.txt` | 1,257 | `d565000fc152deec198714192e14ca7aff6a0b368d1e9079ff3e7b099999e6cb` |
| `bob-gemini-free-macos-universal.dmg` | 20,554,412 | `6922aa8fb4805c8978273544ac53a005afba13a600594c044ac22abe118739b5` |
| `bob-gemini-free-macos-universal.zip` | 19,005,119 | `8597bb1ed28148b5cdb9d96b61bd0f7c149b0460d68b93cba583226d39fc950b` |
| `SHA256SUMS` | 289 | `cd40ab335bc7cee7b952cd0b17b62d9d633902264e99419e9358dcaaecb97483` |
| `SHA256SUMS.sig` | 89 | `3cb13a05d2db93e2aafb278b9e143f2164643565810f0e80a4df35de6e8d129b` |

## One-host installed migration

On the audit Mac, the writable `/Applications/BOB Gemini Free.app` initially
contained the older `v0.2.0-preview.1` desktop build. Through the native UI:

1. **Help → Check for Updates** discovered `v0.2.0-preview.5`.
2. The verified **Install Update** action closed the old app and ran the
   updater transaction.
3. The `/Applications` bundle then contained `v0.2.0-preview.5`; the app
   restarted on `127.0.0.1:8081` and returned `{"status":"ok"}` from
   `/healthz`.
4. The previously visible chat response remained present after restart.
5. About reported `v0.2.0-preview.5`; a second update check reported that no
   newer desktop release was available.

This proves one successful same-key installed migration and local state
continuity. It does not prove rollback after an intentionally failed candidate,
Apple Developer ID/notarization trust, Google provider availability, Windows or
Linux acceptance, or a 20–30-device rollout.

## Remaining gates

- deliberate failed-candidate rollback and recovery on a disposable pilot;
- two or three pilot Macs on the target classroom network;
- only then a staged 20–30-device wave;
- Apple Developer ID, hardened runtime, notarization, and stapling if trusted
  macOS distribution is required; and
- live provider/session/quota evidence for any generation claim.
