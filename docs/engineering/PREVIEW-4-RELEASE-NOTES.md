# BOB Gemini Free v0.2.0-preview.4

This is the next controlled macOS universal preview, built from the reviewed
public `main` source after Preview 3. It is intended for owner and pilot
validation before any stable or broad student rollout.

Published manually as a GitHub prerelease from public `main` commit
`abfeebaaaaabc740ea29602b602591a0b707fbc2`:

[`v0.2.0-preview.4`](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.4)

The five public assets were downloaded again after publication, passed the
detached Ed25519 signature/checksum verifier, and matched the locally signed
inputs byte-for-byte. No GitHub Actions workflow was used.

## What changed since v0.2.0-preview.3

- Explicit Google HTTP 401/403 responses now invalidate only the cached
  dynamic `/app` page token and build identifier. The configured cookie file
  and guest cookie remain intact, and the rejected generation is not replayed.
- An invalidation generation prevents an older in-flight bootstrap from
  restoring dynamic tokens after the rejection.
- Buffered and streaming authentication rejection, forced refresh, configured
  credential preservation, non-invalidation of quota responses, and the
  in-flight ordering race are covered by deterministic tests.
- The preview packager defaults to the new immutable `v0.2.0-preview.4`
  identity so an unqualified packaging command cannot recreate Preview 3.

## Trust and installation boundary

- The package is ad-hoc signed for bundle integrity, not Apple Developer ID
  signed, notarized, or stapled. macOS may require **Open Anyway** on first
  launch.
- `SHA256SUMS` and `SHA256SUMS.sig` authenticate the project release assets;
  they do not create Apple platform trust.
- The private release key stays outside GitHub and outside the application.
  The app contains only the public verification key.
- The updater asks for explicit user consent and retains rollback evidence. It
  is not a silent fleet-push mechanism.

## Provider boundary

The app can start locally without a Go, Node, Rust, SQLite, or separate server
installation. Successful local startup does not guarantee that Google will
accept a request: anonymous access, cookies, API keys, model availability,
quota, rate limits, and network reputation remain upstream-dependent. Never
share a student's cookie or API key.

## Pilot instructions

1. Download the exact macOS `.dmg` or `.zip` asset from the GitHub release.
2. Copy **BOB Gemini Free.app** to `/Applications` or another writable app
   directory, then launch it and approve the normal macOS warning if shown.
3. Confirm the displayed version and actual loopback gateway endpoint, then
   send one bounded smoke prompt using the student's own authorized provider
   route.
4. Use **Help → Check for Updates** only after the app is outside a mounted
   disk image. Confirm the displayed target before choosing **Install Update**.

This preview is not stable, notarized, or evidence of a completed 20–30-device
rollout. Record clean-device update, restart, rollback, and configuration
preservation evidence before announcing a broader deployment.
