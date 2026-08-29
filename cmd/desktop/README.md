# BOB Gemini Free — Native Desktop

This is the supported native desktop entrypoint. The packaged application
contains the Go gateway and opens the shared BOB Builder studio;
end users do not need a separate Go installation, local server, SQLite
database, or memory service.

The bootstrap page in `frontend/index.html` validates the gateway endpoint
before handing off to `/playground`. It also has an event-independent
`GatewayURL` handoff for the one-shot startup event, so a fast local gateway
cannot strand the window on its loading screen.

The packaged app discovers the current user's config and cookie files when
they exist, while forcibly keeping the desktop gateway on loopback with API
keys and remote origins disabled. It never ships a shared Google session or
developer credential.

## Development prerequisites

Building from source requires Go, CGO, and the host platform toolchain. The
Makefile invokes the pinned Wails v2.15.0 CLI through `go run`; direct
`wails build` usage requires the Wails CLI on `PATH`. These are build-time
prerequisites only; they are not runtime dependencies of a packaged
application.

## Build and smoke test

From the repository root:

```bash
make web
cd cmd/desktop
wails build -clean
```

Platform-specific production targets are explicit:

```bash
make desktop-mac      # macOS host; universal app bundle
make desktop-windows  # Windows host; NSIS installer + WebView2 bootstrap
make desktop-linux    # Linux host; WebKitGTK 4.1 build
```

On macOS, the checked-in staging helper can build the native host target or a
universal bundle:

```bash
scripts/build-wails-local.sh /tmp/bob-gemini-free-wails-macos
BOB_WAILS_PLATFORM=darwin/universal scripts/build-wails-local.sh /tmp/bob-gemini-free-wails-universal
```

For the no-Apple-membership evaluation path, build a clearly labelled local
preview package containing an ad-hoc-signed `.app`, `.zip`, `.dmg`, checksums,
and an installation notice:

```bash
make desktop-preview-mac
```

This preview is not Developer ID signed or notarized and must not be called a
production student release. The full boundary is recorded in
[`docs/engineering/FREE-DESKTOP-PREVIEW.md`](../../docs/engineering/FREE-DESKTOP-PREVIEW.md).

These are native-host build commands, not claims that a macOS checkout can
produce a tested Windows or Linux desktop release. The student download
matrix, signing requirements, authentication boundary, and acceptance gates
are recorded in
[`docs/engineering/STUDENT-DISTRIBUTION.md`](../../docs/engineering/STUDENT-DISTRIBUTION.md).

The native Help menu provides an explicit update check against the official
GitHub release metadata. A desktop build produced from the current source also
performs a low-frequency background metadata check after startup and at most
once per day while it remains open. The check can present the same update
dialog, but it never silently downloads, replaces, or restarts the running app.
Already-published binaries keep the behavior compiled into that release. The public
Preview 7 binary checks its preview channel; newly built current-source
previews check stable first and can use a same-key bridge for existing Preview
7 installations. The source also contains a signed-manifest staging/helper
path, but it is enabled only in a build with the embedded release key and a
platform-appropriate signed manifest.

The app starts an embedded loopback gateway, reuses only an identified
compatible existing BOB gateway, or selects a safe free port. The actual
endpoint is passed to the frontend and the owned gateway is closed with the
desktop lifecycle.

For the complete local release and clean-machine acceptance boundary, see
`docs/engineering/RELEASE-PROCESS.md` and
`docs/engineering/DESKTOP-ARCHITECTURE.md`.
