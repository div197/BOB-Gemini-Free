# Preview Rollout and 20–30 Device Validation

**Current target:** unpublished `v0.1.7-preview.7` stabilization candidate,
based on the public `v0.1.7-preview.6` release

This runbook separates three different questions that are often accidentally
combined:

1. Does the local app start and update safely?
2. Can 20–30 local clients exercise the gateway without a local bottleneck?
3. Does Google's current web service accept the requests for the target
   accounts and network?

The first two can be tested hermetically. The third is a bounded, authorized
provider check; it cannot be certified by a local benchmark.

## What the current desktop updater actually does

Preview 6 has a build-pinned preview channel and a signed manifest verifier.
The updater:

- contacts the fixed official GitHub release API only after the user selects
  **Help → Check for Updates**;
- selects the newer `preview.N` package for the current platform;
- verifies the embedded Ed25519 public key, `SHA256SUMS`, signature, asset name,
  size, and package contents;
- stages beside the installed application and restarts through the tested
  helper/rollback path; and
- requires explicit user consent.

It does **not** silently check on startup, silently replace the application, or
push a release to 30 machines. A new signed preview can therefore reduce the
per-device reinstall work, but each installed app still needs a user-initiated
check/install unless an owner-controlled fleet/MDM system is added separately.
GitHub Actions are not required for this process.

An installed app must be copied out of a mounted DMG/App Translocation location
and placed in `/Applications` (or another writable application directory)
before staging an update. The updater cannot safely replace a read-only mounted
image.

## Hermetic local 20–30 client test

Run this on the release host before any live provider test:

```bash
go run ./cmd/benchmark-local -profiles 1,10,20,30 -requests 100
```

This uses a deterministic in-process upstream fixture. It measures local
gateway parsing/formatting, HTTP concurrency, allocations, RSS where
available, goroutines, and connection counts. It does not contact Google and
must not be described as a provider quota or latency test.

Acceptance requirements:

- every profile reports `failed: 0`;
- no race detector failure in the full Go suite;
- `go vet ./...` is clean;
- P95/P99 are recorded with date, OS, architecture, Go version, and commit;
- no cookies, prompts, response bodies, or image bytes appear in the report.

Run the race and static checks separately:

```bash
go test -race -count=1 ./...
go vet ./...
```

## Native app rollout sequence

Do not open 30 live generations at once. Use this order:

### Gate A — one clean Mac

- install the exact release asset from the GitHub release page;
- move the app to `/Applications`;
- open it and record the displayed version and local gateway endpoint;
- run a local `GET /healthz` check;
- send one short text request using the student's own authorized path, if
  that capability is required;
- check Help → Check for Updates against a later signed preview in a controlled
  test release;
- verify that an intentionally failed candidate leaves the original app,
  cookies, preferences, and chat history usable.

### Gate B — three pilot Macs

Use the same classroom network and record only:

```text
device label | macOS | arm64/amd64 | app version | healthz | generation class | update result
```

`generation class` must be one of `success`, `anonymous-provider-rejected`,
`authenticated-session-rejected`, `rate-limited`, `network-error`, or
`not-tested`. Never put cookies, Google account identifiers, full prompts, or
raw upstream responses in the record.

### Gate C — 20–30 device rollout

Proceed only when Gates A and B have no unresolved installer/update defects.
Install in waves of 5–10, verify local health after each wave, and keep one
operator able to stop the rollout. A shared public egress IP is expected; do
not attempt to disguise it with fingerprint changes, rotating proxies, or
shared cookies.

The rollout is not provider-certified merely because all apps open. Google may
apply account, session, model, traffic, network, or temporary policy limits.
If several devices fail with the same provider error, pause the wave, preserve
the exact error class and timestamp, and wait for the provider/network owner
to investigate. Do not multiply the failure with repeated retries.

## Live provider sampling

Live testing is optional and must be explicitly authorized. Keep it bounded:

- one short text request per pilot device;
- no diagnostic suite or benchmark burst against Google;
- no repeated model rotation to search for a working route;
- no cookie pooling across students;
- no claim that a successful sample establishes a quota, model identity,
  context limit, or unlimited access.

This is the correct evidence boundary for a public repository: local behavior
is regression-tested, while Google behavior is recorded as a time-stamped,
account/network-specific observation.

## Release decision for the next preview

The next preview may be published only after:

1. the full local test/race/vet/build gate passes;
2. the generated `web/index.html` matches `internal/server/playground.html`;
3. the updater artifacts are signed and manually verified on a clean writable
   application copy;
4. the release notes disclose that platform notarization/publisher trust is
   still absent, if that remains true;
5. the preview is tested on one clean Mac and three pilots; and
6. the release page says explicitly that updater success and `/healthz` do not
   prove upstream Google availability.

Until those gates pass, the honest status is **stabilization candidate**, not
“ready for an unattended 30-device production rollout.”
