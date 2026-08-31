# Preview 5 Local Candidate Verification — 2026-08-31

## Scope and decision

This record covers the locally produced macOS `v0.2.0-preview.5` candidate
from public `main` at commit `4beb1275833c387f3bcf458d99b5743720e84311`.
It is a local package and startup receipt, not a GitHub publication receipt.

The candidate passed the local source, package, signature, and one-host startup
checks listed below. It has **not** been published, byte-reconciled from
GitHub, or accepted as an installed-bundle update from Preview 4 or Preview 7.
The public downloadable desktop release therefore remains the immutable
`v0.2.0-preview.4` package.

## Operator boundary

- Work was performed on macOS 26.2, arm64, with Go 1.26.6.
- No GitHub Actions workflow was present or used.
- No provider API key, Google cookie, GitHub credential, or release private key
  was placed in the repository, package, logs, or this record.
- The release signer used the owner-controlled macOS Keychain; only the
  checked-in public updater key is part of the build.
- Existing Preview 4 assets were not overwritten and no public tag was reused.

## Local source and package checks

| Check | Result | Evidence |
|---|---|---|
| Public source identity | PASS | Candidate built from the clean public code baseline `4beb1275833c387f3bcf458d99b5743720e84311`; release-evidence documentation is separate and does not alter packaged code |
| Release-source preflight | PASS | `scripts/verify-release-source.sh v0.2.0-preview.5` |
| Full Go suite | PASS | `go test -count=1 ./...` |
| Race-sensitive suite | PASS | `go test -race -count=1 ./internal/server ./internal/updater ./internal/gemini ./internal/geminiapi ./internal/config ./cmd/desktop` |
| Static checks | PASS | `go vet ./...`, `go mod verify`, `git diff --check` |
| Web bundle parity | PASS | Generated `web/index.html` matches the embedded Studio source after version substitution; inline JavaScript parses |
| Package assets | PASS | Branded universal `.zip` and `.dmg`, `RELEASE-NOTICE.txt`, `SHA256SUMS`, and `SHA256SUMS.sig` produced in a fresh output directory |
| Manifest signature | PASS | Detached Ed25519 signature verified by `scripts/verify-release-assets.sh` with the checked-in public key |

## Manifest entries

These are the exact SHA-256 entries in the locally signed manifest:

```text
d565000fc152deec198714192e14ca7aff6a0b368d1e9079ff3e7b099999e6cb  RELEASE-NOTICE.txt
556023933cded7f4eaf61fda0ce6c3113b0cc28b3c7b94ec4a0d80ec1c561fa3  bob-gemini-free-macos-universal.dmg
a5126068bb35d9bbd73f1adb423f9ad806f6e802fcceb8ffd5db0e44572ccaea  bob-gemini-free-macos-universal.zip
```

The manifest intentionally covers the release notice and the two installable
artifacts. The detached signature authenticates the manifest itself.

## Bundle inspection

- Bundle name: `BOB Gemini Free.app`.
- Bundle identifier: `com.abcsteps.bob-gemini-free`.
- Base macOS bundle version: `0.2.0`; the full preview identity is injected
  into the desktop binary and release notice as `v0.2.0-preview.5`.
- Minimum macOS version: `10.13.0`.
- Architectures: `x86_64` and `arm64`.
- Code signature: ad-hoc and internally verifiable; no Apple Developer ID
  identity or notarization ticket is claimed.
- DMG layout: branded app plus a visible `/Applications` alias.
- Release notice: states the beta, Apple trust limitation, consented updater,
  missing-provider-credential boundary, and provider/session limits.

## Fresh packaged-app smoke test

The exact candidate app was launched with macOS `open -n` from its fresh local
package. After bootstrap it:

1. owned the selected loopback listener at `127.0.0.1:8081`;
2. returned stable JSON `{"status":"ok"}` from `GET /healthz` without a
   Google, GitHub, or credential operation;
3. served `/playground` containing the separated **BOB Gateway Access Key**,
   **Google Gemini Developer API**, and credential-map markers; and
4. shut down cleanly after a targeted termination of the candidate process.

This proves the candidate can bootstrap on this Mac. It does not prove Google
generation, a clean-device Gatekeeper experience, update replacement,
rollback, or classroom-scale operation.

## Publication and installed-base gates still open

Before calling Preview 5 a release, the operator must still:

1. publish a new immutable GitHub prerelease manually with these exact assets;
2. download the public assets into a fresh directory and verify the signature,
   hashes, and byte equality with this local candidate;
3. test **Help → Check for Updates** from a writable `/Applications` install
   of the same-key Preview 4/Preview 7 baseline;
4. observe successful replacement, restart, health confirmation, and a
   deliberately failed-candidate rollback without losing local state; and
5. use one or two pilot Macs before any 20–30-device rollout.

Apple Developer ID signing/notarization, Windows publisher signing, Linux
acceptance, live provider limits, and anonymous/session behavior remain
separate platform and upstream gates. See the
[`current release audit`](RELEASE-AUDIT-2026-08-31.md) and
[`desktop update operations`](DESKTOP-UPDATE-OPERATIONS.md).
