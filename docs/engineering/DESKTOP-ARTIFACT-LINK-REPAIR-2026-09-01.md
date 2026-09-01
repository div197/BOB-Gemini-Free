# Desktop Artifact and External-Link Repair — 2026-09-01

## Evidence-led diagnosis

The reported native-app screenshots showed three different failures:

1. The generated Three.js solar-system artifact started briefly and then
   reported only `Script error.`.
2. Artifact **Pop Out** displayed `The artifact window was blocked by the
   browser`.
3. The GitHub control did not open the device's default browser.

The third failure was architectural, not a GitHub outage. The Wails bootstrap
used `window.location.replace()` to navigate the native window to the
loopback gateway. Wails injects its `window.runtime` bridge into the Wails
asset page, not into the subsequently navigated loopback document. The
loopback Studio therefore reached a missing native `BrowserOpenURL` bridge and
its WebView `window.open()` fallback was not reliable.

The second failure was a native-WebView capability mismatch. The artifact
implementation attempted to create a `blob:` popup. That is a useful browser
path, but it is not a dependable second-window API in the embedded Wails
WebView. It also cannot safely open the generated document unsandboxed.

The first failure was intentionally not “fixed” by weakening the sandbox or
rewriting arbitrary generated code. The artifact source in the screenshot is
runtime output, not a checked-in fixture, and the previous listener discarded
the filename, line, column, and failed-resource URL. The host could therefore
not distinguish a generated Three.js exception from a failed CDN/module load.

## Implemented boundary

- The Wails page now retains its native shell and loads the loopback Studio in
  a frame using `desktop_shell=1`.
- Only that explicit query path may be framed, and only by the Wails asset
  origins used by macOS/Linux and Windows. Ordinary `/playground` responses
  retain `X-Frame-Options: SAMEORIGIN`.
- The loopback Studio validates external protocols and sends a
  `BOB_OPEN_EXTERNAL_URL` message to the parent. The Wails shell accepts it
  only from the exact gateway frame and exact gateway origin before calling
  its native default-browser bridge.
- Native artifact full-screen now expands the existing sandboxed Studio in
  place. The state is reversible and exposes `aria-pressed`; no
  `allow-same-origin` permission was added.
- Browser pop-out remains sandboxed and keeps its existing JSON/script
  escaping. If the browser blocks it, the action falls back to the in-place
  full-screen state rather than reporting a dead end.
- Artifact runtime errors now include failed resource URLs or source
  filename/line/column when the WebView supplies them, plus CSP-blocked
  resource information.

## Verification

The following source-level gates pass on the repair branch:

- `go test -count=1 ./internal/server`
- `go test -count=1 ./cmd/desktop`
- focused artifact, external-link, desktop-shell, and playground tests
- inline playground JavaScript syntax parse
- desktop bootstrap JavaScript body syntax parse
- `make web` generated-bundle synchronization

A fresh desktop package still must be built and launched before this repair is
available to users. The published Preview 8 assets are immutable and do not
silently acquire source changes.

## Remaining evidence gate

After launching a fresh local candidate, retry the exact 3D Solar System
prompt. If it still fails, the new overlay should identify either the failed
Three.js/OrbitControls resource or the generated source location. That is the
minimum evidence required for a code-specific artifact repair; the repository
must not claim that every arbitrary CDN-dependent generated program is
universally executable without that fixture and live WebView result.
