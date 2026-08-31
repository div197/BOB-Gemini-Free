# BOB Gemini Free v0.2.0-preview.6

## Controlled macOS beta

Preview 6 is the current immutable macOS universal beta, published manually
from source target `f9b3410`. It is intended for owner and staged pilot
validation before the `v0.2.0` stable gate. The package does not require Go,
Node, Rust, SQLite, a separate server, or GitHub Actions on a student's Mac.

## What is included

- Branded universal macOS application for Apple Silicon and Intel.
- Embedded Go gateway and BOB Builder studio with loopback-first startup and
  safe occupied-port coexistence.
- Explicit, user-consented desktop update checks with a signed project manifest,
  bounded download size, atomic staging, and rollback-aware replacement.
- Credential-route guidance that separates the BOB Gateway Access Key, the
  optional student-owned Google Gemini Developer API key, engine-owned
  web-session cookies, and the gateway endpoint URL.
- Fail-closed transport handling: BOB access keys and student provider keys are
  not sent to non-loopback cleartext HTTP endpoints.

## Distribution and trust

The public release contains the universal DMG, universal ZIP,
`RELEASE-NOTICE.txt`, `SHA256SUMS`, and `SHA256SUMS.sig`. The project Ed25519
signature and exact asset hashes were verified again after download from GitHub.
The macOS bundle is ad-hoc signed only; it is not Apple Developer ID signed,
notarized, or stapled. macOS may therefore show a first-launch warning or
require **Open Anyway**. The private release key remains outside the repository
in the owner's local Keychain; only the public verification key is shipped.

## Update boundary

The updater is explicit and user-consented. It may check release metadata on a
bounded schedule and show an update dialog, but it does not silently download,
replace, or restart a student's application. A current-source Preview 6 client
does not update to itself. The published legacy `v0.1.7-preview.7` path can
select Preview 6 through its preview-only lookup, but a real installed
Preview 7 → Preview 6 replacement remains a staged pilot gate. Legacy
`v0.1.7-preview.6` builds need a one-time manual migration because their old
project signing key was not recoverable.

## Provider and credential boundary

Leave both key fields empty for the default web-session route. That route is
still dependent on Google's guest/session policy, network reputation, quotas,
and model availability. The Developer API route is an opt-in path for a
student's own Google AI Studio project; its key is held in page memory and sent
only when that route is enabled. Google controls its quota, billing, model,
policy, and availability. BOB does not rotate keys, share cookies, or silently
fall back between provider routes.

## Acceptance status

Source tests, race tests, `go vet`, module verification, generated-bundle
verification, local benchmark profiles, package inspection, signed-manifest
verification, public-byte reconciliation, loopback health, and one-host
coexistence smoke passed. Apple platform trust, clean-device installation and
rollback, live Google behavior, Windows/Linux acceptance, and a 20–30-device
pilot remain open. Treat this release as a controlled beta, not a stable
student-wide deployment.

See [`RELEASE-AUDIT-2026-08-31.md`](RELEASE-AUDIT-2026-08-31.md),
[`PREVIEW-6-LOCAL-VERIFICATION-2026-08-31.md`](PREVIEW-6-LOCAL-VERIFICATION-2026-08-31.md),
and [`STUDENT-DISTRIBUTION.md`](STUDENT-DISTRIBUTION.md) for the evidence and
rollout contract.
