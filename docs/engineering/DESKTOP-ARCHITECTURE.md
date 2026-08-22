# Desktop Architecture Decision — Single Native Path

**Decision status:** `cmd/desktop` is the sole supported native desktop
architecture.
The former alternate desktop wrapper was archived by deletion from the active
tree after the capability comparison; Git history preserves its provenance.

## Decision

`cmd/desktop` is the canonical and only active desktop product path. It embeds
the Go gateway in the desktop process, so a packaged user does not need a
separate Go runtime, local server, SQLite database, memory service, Node
installation, or Rust installation.

The desktop gateway boundary is deliberately explicit:

- probe the requested loopback port;
- reuse only a compatible identified BOB gateway;
- select a safe loopback port when the requested port is occupied by another
  process;
- expose the actual endpoint to the frontend;
- surface startup errors in the native window; and
- close only a gateway owned by the app during shutdown.

The deleted alternate wrapper provided no capability required by the canonical
desktop path. It used
a fixed port, a sidecar process, and no compatible-gateway endpoint handoff.
Keeping it active would create a second release surface without adding product
value. Its deletion is therefore a product simplification and a security
boundary reduction, not a protocol refactor.

## Release and acceptance boundary

`make desktop` is a developer build and requires the Wails CLI, CGO, and the
host platform toolchain. For a macOS checkout hosted by File Provider, use
[`scripts/build-wails-local.sh`](../../scripts/build-wails-local.sh); it stages
the source in `/tmp`, builds with the pinned Wails module version, and verifies
an ad-hoc signed bundle.

A successful Go package build is not native GUI acceptance. Every supported
release still needs a device smoke test for window startup, dynamic endpoint
handoff, chat, shutdown, and occupied-port fallback. Published releases also
need Developer ID signing/notarization and a clean-machine install/open check.

## Recovery record

The removed wrapper and its build metadata remain available through Git history:

```text
git log --all -- desktop
git show <pre-archival-commit>:desktop/
```

No active Make target, release script, icon generator, server origin allow-list,
or current product document depends on that archived path.
