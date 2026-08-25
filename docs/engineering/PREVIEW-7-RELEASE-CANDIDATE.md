# BOB Gemini Free v0.1.7-preview.7 Release Candidate

**Status:** local, unpublished stabilization candidate as of 2026-08-25.
The public release remains `v0.1.7-preview.6` until the release gates below
are closed. A new owner-controlled Ed25519 key pair has been generated because
the original Preview 6 private key was not recoverable; its public half is
recorded in `UPDATE-PUBLIC-KEY.txt`, while its private half remains only in
macOS Keychain.

## Candidate scope

This candidate is a surgical reliability pass over the Preview 6 baseline:

- provider policy and rejection responses no longer receive automatic retry
  amplification; only transport and server failures remain retryable;
- cumulative stream retry deduplication remains intact;
- a successful cookie-pool session clears its failure cooldown;
- manual UI retries are bounded and paced, and retrying an earlier response
  targets the corresponding user turn;
- generation locks model and request controls and turns terminal SSE provider
  errors into visible failures instead of completed assistant messages;
- responsive drawers stay below the New/model toolbar;
- documentation now distinguishes anonymous web access, optional per-user
  Google sessions, shared school egress, local health, and provider limits;
- non-development native packagers fail closed without an embedded updater
  public key.

## Validation completed in this checkout

- `go test -count=1 ./...`
- `go test -race -count=1 ./...`
- `go vet ./...`
- `go build ./...`
- all inline Web Studio scripts parse and `web/index.html` matches the source
  template;
- deterministic local benchmark profiles at concurrency 1, 10, 20, and 30
  completed with 100/100 requests and zero failures;
- the installed local gateway returned `GET /healthz` status `ok`;
- package scripts reject non-development builds when the update public key is
  absent;
- signed updater tests use temporary fixtures and never replace the developer
  executable.

These are local and deterministic checks. They do not certify Google
availability, quota, account state, shared-IP capacity, or a 30-device live
provider burst.

## Release blockers

Do not publish this candidate until all of these are evidenced:

1. the Keychain-held Ed25519 private key is retrieved only by the local
   signing process and matches the documented public key;
2. macOS artifacts are built with the matching embedded public key and the
   exact artifacts are signed with `SHA256SUMS` and `SHA256SUMS.sig`;
3. a clean writable `/Applications` installation updates from Preview 6 to
   Preview 7 and a deliberately failed candidate rolls back safely;
4. the release assets, tag, notes, and manifest are manually inspected before
   publication;
5. one clean Mac and two or three pilot Macs pass the rollout runbook;
6. release notes continue to state that the package is ad-hoc signed and not
   Apple Developer ID signed or notarized.

The updater remains explicit and user-consented. Publishing Preview 7 will not
silently push it to every installed Preview 6 device and will not change the
Google upstream authentication or shared-network boundary.
