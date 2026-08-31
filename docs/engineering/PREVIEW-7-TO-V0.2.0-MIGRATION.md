# Preview 7 to v0.2.0 Migration Runbook

**Audience:** the owner operating the existing public `v0.1.7-preview.7`
macOS pilot fleet.

**Purpose:** move existing Preview 7 installations onto the current updater
path without manually replacing the application on every Mac.

This is a staged, user-consented migration. It is not a silent fleet push and
it does not require GitHub Actions, an Apple Developer account, a shared
Google cookie, or a student-facing private key.

## Before touching the fleet

- Confirm the application is copied to `/Applications` or another writable
  application directory. Do not run it directly from a mounted DMG.
- Confirm the app is the public `v0.1.7-preview.7` build from the official
  GitHub release page.
- Keep the current app available until the replacement has launched
  successfully.
- Use the same student account and ordinary school network used for normal
  operation. Do not rotate proxies, spoof TLS fingerprints, or share a
  teacher's Google session.

## Phase 1: Preview 7 to migration bridge

The signed same-key `v0.2.0-preview.4` macOS preview is now published. The
immutable `v0.2.0-preview.1` bridge remains available if a device already
selected that intermediate release:

1. Open BOB Gemini Free.
2. Select **Help → Check for Updates**.
3. Confirm that the dialog names `v0.2.0-preview.4` (or the immutable bridge)
   and a macOS universal package.
4. Select **Install Update** and allow BOB to restart.
5. Reopen BOB and confirm that the displayed version is
   `v0.2.0-preview.4` (or the bridge version selected by the updater).
6. Send one small, ordinary test prompt and confirm that the local gateway
   still starts and the response state reaches a terminal result.

If the check reports no update, do not repeatedly click it. Record the Mac's
macOS version, architecture, current version, and exact error. Check the
official release page and use the documented manual DMG recovery path if the
device is on a read-only or translocated application path.

The legacy Preview 7 updater can select the published Preview 4 directly when
its signed manifest and compatible macOS asset are present. The current-source
updater selects stable first and otherwise the highest valid preview. The exact
published version must always be confirmed in the dialog; never assume that a
source candidate is already downloadable.

## Phase 2: Bridge to stable

Only after the bridge passes the one-Mac and pilot gates, publish the signed
stable `v0.2.0` package. On each bridge device, repeat:

1. **Help → Check for Updates**.
2. Confirm the candidate is stable `v0.2.0`, not a preview downgrade.
3. Select **Install Update** and allow the restart.
4. Confirm the stable version, local gateway health, and one ordinary prompt.

The bridge is necessary because the already-published Preview 7 binary only
looks at the preview release channel. A current-source preview looks at the
stable channel first and can then migrate to stable.

## Rollout order for 30 Macs

Use this order:

1. One owner-controlled clean Mac.
2. Two or three pilot Macs.
3. Waves of 5–10 Macs, waiting for the previous wave's version and health
   checks before starting the next.

Each Mac requires an explicit user action. Do not ask all 30 users to check at
the same instant: GitHub's unauthenticated release API is shared and can
return a rate-limit or timeout error. A failed check leaves the existing app
unchanged.

Record only:

```text
device label | macOS version | arm64/Intel | current version | bridge/stable result | health result | error text
```

Never record cookies, SAPISID values, full prompts, response bodies, or the
release private key.

## Expected macOS warning

The free preview package is ad-hoc signed and not Apple Developer ID signed or
notarized. A first-launch **Open Anyway** approval may therefore remain
expected. The project-signed update manifest verifies release authenticity but
does not remove Apple's Gatekeeper warning.

## Manual recovery

If an update cannot be staged, quit BOB and install the exact official DMG into
`/Applications`. Do not delete user data to recover from an updater error. If
the downloaded release has no valid signed manifest, the updater must refuse
it; use the official release page and report the failure instead of disabling
verification.
