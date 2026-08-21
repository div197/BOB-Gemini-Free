# BOB Gemini Free — Wails Desktop

This is the supported native desktop entrypoint. The packaged Wails
application contains the Go gateway and opens the shared BOB Builder studio;
end users do not need a separate Go installation, local server, SQLite
database, or memory service.

The bootstrap page in `frontend/index.html` validates the gateway endpoint
before handing off to `/playground`. It also has a durable `GatewayURL` fallback
for the one-shot startup event, so a fast local gateway cannot strand the
window on its loading screen.

## Development prerequisites

Building from source requires Go, CGO, the host platform toolchain, and the
Wails CLI. Those are build-time prerequisites only; they are not runtime
dependencies of a packaged application.

## Build and smoke test

From the repository root:

```bash
make web
cd cmd/desktop
wails build -clean
```

The app starts an embedded loopback gateway, reuses only an identified
compatible existing BOB gateway, or selects a safe free port. The actual
endpoint is passed to the frontend and the owned gateway is closed with the
desktop lifecycle.

For the complete local release and clean-machine acceptance boundary, see
`docs/engineering/RELEASE-PROCESS.md` and
`docs/engineering/DESKTOP-ARCHITECTURE.md`.
