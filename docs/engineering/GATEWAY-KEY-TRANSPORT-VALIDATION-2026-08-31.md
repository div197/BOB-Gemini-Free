# Gateway Access-Key Transport Validation — 2026-08-31

## Scope

This receipt records the focused browser security correction for the optional
**BOB Gateway Access Key**. It is separate from the Google Gemini Developer
API key and from engine-owned web-session cookies.

The test used a non-secret test sentinel only. No provider key, GitHub token,
cookie, or release-signing material was used or recorded.

## Reproduced baseline

Before the patch, the local Web Studio was configured with a non-loopback
cleartext endpoint (`http://192.168.1.100:9610`) and a page-session gateway
access key. The browser emitted an `Authorization: Bearer …` header to that
endpoint during the explicit ping path. This was a confidentiality boundary
failure even though the endpoint was user-configured.

## Corrected contract

| Endpoint | BOB access-key behavior |
|---|---|
| Loopback HTTP | Allowed for local/native use |
| HTTPS | Allowed, including explicitly configured remote endpoints |
| Non-loopback HTTP | Key withheld; settings show `HTTPS REQUIRED`; ping does not probe while the key is present |
| Any endpoint without a key | No BOB authorization header is created |

The shared `gatewayRequestHeaders()` constructor is used by ping,
`refreshTelemetry`, `syncBackendModels`, and `sendMessage`. The browser still
uses the existing exact-origin/CORS policy; this change does not make a remote
endpoint trusted merely because it uses HTTPS.

## Fresh browser result

Command shape: a local Playwright Chromium probe against a freshly started
source gateway on `127.0.0.1:19621`, with the remote endpoint intercepted by a
mock route so no external request was made.

- 1440×900, 1024×768, and 390×844 all reported zero document and body
  horizontal overflow.
- Unsafe remote HTTP: route state `blocked`, access state `Access key
  withheld — use HTTPS for a non-local endpoint`, modal pill `HTTPS REQUIRED`,
  and zero authorization headers on two mocked telemetry requests.
- Unsafe remote HTTP Test Ping: zero remote ping requests.
- Loopback HTTP: the guarded header constructor produced authorization and the
  real local request capture observed it. The local server returned 401 because
  the test key was not configured; that negative response was expected.
- HTTPS fixture: the guarded header constructor produced authorization for the
  explicit HTTPS endpoint.
- Baseline page load/geometry phase: no console or page errors.

## Regression coverage

- `go test ./internal/server -run
  'TestGatewayAccessKeyRequiresSecureTransport|TestDeveloperAPIRouteRequiresSafeGatewayTransport|TestGatewayAuthKeyIsSessionOnly|TestTelemetryUsesHealthzBeforeProtectedStats' -count=1`
- `internal/server/playground_test.go`
- Generated bundle synchronized by `make web`.

This receipt proves the browser-side transport rule in the tested bundle. It
does not prove remote TLS certificate trust, gateway operator identity, Google
provider acceptance, or a clean-device classroom rollout.
