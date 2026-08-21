# Desktop Architecture Decision — Mission 9

## Comparison

| Capability | Wails (`cmd/desktop`) | Tauri (`desktop`) |
|---|---|---|
| Gateway packaging | Embeds the Go gateway in the desktop process | Spawns the CLI as a sidecar |
| Port behavior | Probes/reuses compatible `9610` or selects a free loopback port | Hardcodes `--port 9610` |
| Actual endpoint handoff | Wails `gateway-ready` event and `GatewayURL` binding | Frontend remains tied to the static web bundle/default port |
| Shutdown | Wails `OnShutdown` calls the owned HTTP server shutdown | Tauri sidecar lifecycle kills the child, but does not own gateway port selection |
| Required product capability | Covers the current native studio and embedded engine | No currently required capability absent from Wails |
| Current verification | Go tests cover collision, reuse, and shutdown behavior | No equivalent deterministic port/lifecycle tests |

## Decision

Wails is the canonical desktop architecture. The Tauri workspace is legacy and
must not be presented as a second supported desktop path for new releases.
Root build/documentation references already point to `make desktop` and
`cmd/desktop`; the Tauri files have no dependency from the Go build graph.

The Tauri directory is retained in this source snapshot because the workspace
has no `.git` history and therefore cannot preserve or prove its historical
provenance while deleting it. A future Git-backed cleanup can archive or
remove it in a reviewable commit after downstream consumers are checked.

This decision does not claim that the old Tauri sidecar is safe for current
production use; its fixed-port behavior is intentionally outside the
canonical release path until it is removed or separately repaired.
