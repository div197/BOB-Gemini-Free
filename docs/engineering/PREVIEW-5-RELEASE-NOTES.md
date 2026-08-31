# BOB Gemini Free v0.2.0-preview.5

## Controlled macOS beta

Preview 5 is the next immutable macOS universal preview, built from public
`main` merge commit `757898cadd45904b0c0f0f454bb5ab30096e974c` (the signed
package receipt records clean build checkout `88f2881`). It is intended for
controlled evaluation before the `v0.2.0` stable gate. The release does
not require GitHub Actions, a Go installation on the student's Mac, a database,
or a separate memory service.

## What changed since Preview 4

- Config now shows the active web-session or Google Developer API route before
  sending, including gateway-auth state, provider-key state, cookie ownership,
  and model compatibility.
- An explicitly selected Developer API route fails closed before creating a
  chat turn or sending a request when its key, endpoint trust, or model/think
  settings are incompatible.
- The two credential fields have independent clear actions, are masked by
  default, and are cleared from the retained modal DOM when Config closes.
- Narrow credential and installer rows are responsive at phone widths; the
  generated web bundle is synchronized with the embedded Studio source.
- README, routing guidance, the failure register, and the verification matrix
  now state the credential boundary and release evidence consistently.

## Distribution and trust

- The macOS package is universal for `arm64` and `x86_64`.
- The DMG includes a visible `/Applications` drag target.
- The app bundle is ad-hoc signed for bundle integrity; it is not Apple
  Developer ID signed and is not notarized. macOS may therefore show **Open
  Anyway** or a first-launch warning.
- Project release assets use the checked-in Ed25519 public key and a detached
  signature over `SHA256SUMS`. The private signing key remains outside the
  repository in the owner's local Keychain.
- The updater is user-consented. It may discover and verify a newer signed
  package, but it does not silently download, replace, or restart the app.

## Credential boundary

The **BOB Gateway Access Key**, the student's **Google Gemini Developer API
key**, and engine-owned web-session cookies are different credentials. BOB
does not persist either page-entered key in browser storage or chat history,
and the Google key is sent only on the explicitly selected Developer API route.
Google quota, billing, model availability, account policy, and network limits
remain provider-dependent. A working local app and a valid project signature
do not prove a successful Google generation.

## Acceptance status

Local source, full-test, race-sensitive, static, package, manifest, and
one-host `/healthz` startup/shutdown checks passed. Public asset
byte-reconciliation, a writable `/Applications` installed-bundle update,
rollback, clean-device Gatekeeper behavior, and staged pilot acceptance are
separate release gates. See
[`PREVIEW-5-LOCAL-VERIFICATION-2026-08-31.md`](PREVIEW-5-LOCAL-VERIFICATION-2026-08-31.md)
and [`RELEASE-AUDIT-2026-08-31.md`](RELEASE-AUDIT-2026-08-31.md).
