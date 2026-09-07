# Native Desktop Bootstrap Repair — 2026-09-05

## Decision

The Wails desktop shell and the browser/PWA shell have different lifecycle
owners. A native Wails frame must not depend on a service worker created by an
earlier browser session or desktop build. Native mode now retires that local
registration, bypasses the worker for `desktop_shell=1`, and uses a bounded
parent/child readiness handshake.

## Investigation

The failure was not treated as a Google, cookie, port, or architecture issue.
The following observations isolated it to the local WebKit/PWA boundary:

1. A current-source diagnostic build reached the complete embedded Studio on
   both ARM64 and AMD64, ruling out the Go gateway and the generated Studio as
   the primary bootstrap failure.
2. The installed app opened normally on a fresh loopback origin at
   `127.0.0.1:65047`.
3. After the app was stopped and relaunched on the same port, the same bundle
   remained on `LOCAL AI GATEWAY STUDIO` indefinitely. The gateway served
   `/playground` and `/sw.js`, but the native frame never became visible.
4. The service worker was introduced for the ordinary browser/PWA path and
   controlled the loopback origin. This made its navigation/cache lifecycle a
   hidden dependency of the native Wails iframe.

## Minimum implementation

- `internal/server/playground.go` sends `Cache-Control: no-store` for native
  shell navigations.
- `internal/server/sw.js` returns without interception for
  `desktop_shell=1` requests, while retaining ordinary browser caching and
  API exclusions.
- `internal/server/playground.html` detects native shell mode before loading
  external libraries, announces readiness immediately, and unregisters any
  legacy local registrations asynchronously. Ordinary PWA registration is
  disabled only for native shell mode.
- `cmd/desktop/frontend/index.html` validates the exact iframe source and
  origin, uses a per-launch bootstrap nonce to avoid a legacy cached
  navigation, accepts `BOB_DESKTOP_SHELL_READY`, and has a 15-second
  retryable error state rather than an unbounded splash.
- `web/index.html` was regenerated from the canonical server HTML.

The immediate readiness message is deliberately not an authentication or
trust signal. It only allows the native parent to reveal its already-validated
loopback frame; the parent rejects all other message sources/origins and the
gateway remains governed by its normal route/auth rules.

## Evidence

| Check | Result |
|---|---|
| `go test -count=1 ./internal/server ./cmd/desktop ./internal/updater` | Passed |
| `go vet ./...` | Passed |
| Patched ARM64 source bundle on the previously affected port, first launch | Passed; complete Studio accessibility tree visible |
| Same patched bundle, quit and relaunch on the same port | Passed; complete Studio accessibility tree visible |
| New native `/sw.js` registration in the patched path | Not observed in the server log |
| Public Preview 9 contains this repair | No; it predates this source change |

## Release consequence

This is a source-candidate repair, not a retroactive modification of a public
release. The next candidate must use a new immutable version, currently
`v0.2.0-preview.10`, and must be rebuilt, signed, reconciled against all
public assets, and tested from an installed writable bundle. Only then can it
be offered as the next update to Preview 7 or current v0.2 preview users.

## Remaining boundary

The repair does not make every historical build updateable. Preview 1–6 and
legacy stable-labelled builds with an obsolete or missing project key still
need one manual migration. Preview 7 can discover a same-key newer preview;
current v0.2 previews use the current stable-first policy. All update installs
remain explicit and per-device; no silent classroom-wide push exists.
