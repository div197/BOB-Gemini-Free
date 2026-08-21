# Desktop Architecture Decision — Phase III

**Decision status:** Wails is canonical for supported desktop releases. Tauri
is legacy source retained for a deliberate, reviewable follow-up removal.

## Comparison

| Capability | Wails (`cmd/desktop`) | Tauri (`desktop`) |
|---|---|---|
| Gateway packaging | Embeds the Go HTTP handler in the desktop process | Spawns the CLI as a sidecar |
| Port behavior | Probes the requested loopback port, reuses only an identified compatible BOB gateway, or selects a free loopback port | Passes a fixed `--port 9610` to the sidecar |
| Actual endpoint handoff | Emits `gateway-ready` with the actual endpoint and exposes `GatewayURL` to the frontend | Uses the static web bundle/default endpoint; no dynamic endpoint contract is implemented |
| Reuse safety | Requires BOB identity, health protocol version, and no API-key requirement before reuse | No compatible-gateway probe |
| Startup failure | Opens the desktop error state and emits `gateway-error` | Uses `expect(...)`, so startup failure is not surfaced through a tested product state |
| Shutdown | `OnShutdown` closes only a gateway owned by this app; reused gateways are left running | Explicit `ExitRequested`/`Exit` handling kills the managed sidecar; the wrapper still uses a fixed port |
| Tooling | Go tests cover collision, reuse, identity rejection, and lifecycle paths; the Wails bundle was also built and smoke-tested on this Mac | Rust/Cargo build and a device smoke path pass locally, but no deterministic desktop regression suite is present |
| Release path | Built locally with `make desktop`; local release packaging is documented separately | Excluded from the root Makefile and local release packaging |

## Decision

Wails is the canonical desktop architecture. The Tauri workspace must not be
presented as a second supported desktop path for new releases. Root build and
release references point to `make desktop` and `cmd/desktop`; a source scan
found no required Tauri-only capability or Go build dependency.

The repository now has Git history, so retaining Tauri is a conscious
reversibility decision rather than a provenance limitation. Its history can be
removed in a dedicated commit after checking downstream references and release
consumers. Until then, `desktop/README.md` labels it legacy and the local
release procedure excludes it.

This decision does not claim that the old Tauri sidecar is a second supported
release path. Its fixed-port behavior and lack of a dynamic endpoint contract
remain outside the canonical release path, although its packaged macOS smoke
test now proves startup, BOB Builder rendering, and sidecar shutdown on this
device.

## Release and acceptance boundary

`make desktop` is a developer build and requires the Wails CLI, CGO, and the
host platform toolchain. For a macOS checkout hosted by File Provider, use
`scripts/build-wails-local.sh`; it stages the source in `/tmp`, builds with the
pinned Wails module version, and verifies an ad-hoc signed bundle. A successful
Go package build is not a native GUI acceptance test; each release still needs
a device smoke test for window startup, dynamic endpoint handoff, chat,
shutdown, and the occupied-port fallback.
