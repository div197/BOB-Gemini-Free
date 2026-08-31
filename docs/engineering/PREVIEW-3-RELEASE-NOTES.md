# BOB Gemini Free v0.2.0-preview.3

This is a controlled macOS universal beta preview built from the public
repository. It is intended for owner and pilot testing before any stable or
broad student rollout.

## What changed since v0.2.0-preview.2

- Bounded cancellation-aware retry handling for transient GitHub release
  metadata timeouts.
- Actionable, calm native update-check, staging, and startup-recovery dialogs.
- Responsive tablet/phone drawer semantics with initial focus, keyboard Tab
  trapping, Escape close, and focus return.
- Branded `BOB Gemini Free.app` bundle naming in both the ZIP and DMG install
  surfaces.
- Local browser smoke evidence at desktop, tablet, and phone widths with no
  document horizontal overflow or browser console warnings/errors.
- A version-aware package notice so the downloaded artifact identifies its
  exact preview version and does not carry historical migration instructions.

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
3. Confirm the app reaches its loopback endpoint and send a bounded smoke
   prompt using the student's own authorized provider route.
4. Use **Help → Check for Updates** only after the app is installed outside a
   mounted disk image. Confirm the displayed target version before choosing
   **Install Update**.

This preview is not stable, notarized, or evidence of a completed 20–30-device
rollout. Record clean-device update, restart, rollback, and configuration
preservation evidence before announcing a broader deployment.
