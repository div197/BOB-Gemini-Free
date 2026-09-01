# Preview Rollout and 20–30 Device Validation

**Current public target:** controlled macOS preview `v0.2.0-preview.8`, for
controlled evaluation of the existing public `v0.1.7-preview.7` fleet and
earlier `v0.2.0` previews. Preview 6, Preview 5, Preview 4, and Preview 3 are
historical; Preview 8 is the current signed public prerelease, and the earlier
Preview 7 candidate was superseded before publication. The published
`v0.2.0-preview.1` migration bridge remains
available.

This runbook separates three different questions that are often accidentally
combined:

1. Does the local app start and update safely?
2. Can 20–30 local clients exercise the gateway without a local bottleneck?
3. Does Google's current web service accept the requests for the target
   accounts and network?

The first two can be tested hermetically. The third is a bounded, authorized
provider check; it cannot be certified by a local benchmark.

## What the current desktop updater actually does

The current source has a build-pinned update policy, a signed manifest
verifier, and a low-frequency background metadata check for newly built
published desktop packages. A newly built preview first checks the fixed
official stable endpoint so it can migrate to a newer stable release; when
stable has no update, it checks the preview listing. The already-published
`v0.1.7-preview.7` binary predates that stable-first behavior and checks only
the preview listing. The updater:

- contacts the fixed official GitHub release API after the user selects
  **Help → Check for Updates**, or through one delayed and once-daily
  background metadata check while a published desktop build is running; the
  first background check waits 30 seconds plus a per-process random jitter of
  up to five minutes so a classroom restart does not synchronize every
  metadata request;
- selects the newer stable package, or the highest published `preview.N`
  package when no newer stable native package exists for the current platform;
- verifies the embedded Ed25519 public key, `SHA256SUMS`, signature, asset name,
  size, and package contents;
- stages beside the installed application and restarts through the tested
  helper/rollback path; and
- requires explicit user consent.

Legacy `v0.1.7-preview.6` installations cannot verify Preview 7 because the
original Preview 6 project signing key was not recoverable. Install Preview 7
manually once on those devices; later releases signed with the Preview 7 key
can then use this updater. This is separate from the current `v0.2.0-preview.8`
public release. A newly built current-source preview can migrate to a newer
stable release after the stable candidate is published with the same project
key. An existing public Preview 7 installation can now discover the published
same-key Preview 8 through its preview-only path, subject to the live device
transition gate, or be manually replaced with stable.

It does **not** silently download, replace, or restart the application, and it
does not push a release to 30 machines. A new signed preview can therefore
reduce the per-device reinstall work: each running app can discover the update,
but the user still approves the verified install unless an owner-controlled
fleet/MDM system is added separately.
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
- for the existing public `v0.1.7-preview.7` path, check the published
  same-key Preview 8 candidate (or the published Preview 1 bridge only if that
  intermediate step was selected), then check stable from that bridge (and
  separately verify a later preview when no newer stable native package exists);
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

## Post-publication rollout decision

The published `v0.2.0-preview.1` bridge and controlled macOS
`v0.2.0-preview.8` are public. Preview 8's exact public macOS assets have been re-downloaded,
signature-verified, and byte-reconciled with the local Keychain-backed
publication input. The earlier Preview 1 → Preview 5 installed migration is
verified on one writable host; the audit Mac discovered Preview 8 from the
installed Preview 5 app, but the install action was canceled. A Preview 5 or
Preview 7 → Preview 8 installed transition therefore remains a device gate.
The remaining
rollout gates are:

1. **COMPLETED:** the full local test/race/vet/build gate passes;
2. **COMPLETED:** the generated `web/index.html` is regenerated from
   `internal/server/playground.html` with the documented version substitution;
3. **COMPLETED for one writable `/Applications` host:** the updater artifacts
   were signed and manually verified, and Preview 1 migrated to Preview 5;
   deliberate rollback, clean-device recovery, and Preview 8 installed-base
   transition remain open;
4. **COMPLETED:** the release notes disclose that platform notarization/publisher trust is
   still absent, if that remains true;
5. **COMPLETED for one audit Mac / OPEN for pilots:** the preview is tested on
   one writable installed bundle; two or three pilot Macs remain required; and
6. **COMPLETED:** the release page says explicitly that updater success and `/healthz` do not
   prove upstream Google availability.

Until those gates pass, the honest status is **controlled public beta**, not
“ready for an unattended 30-device production rollout.”

The complete version and installed-base transition matrix is recorded in
[`RELEASE-TRANSITION-AUDIT-2026-08-31.md`](RELEASE-TRANSITION-AUDIT-2026-08-31.md).
