# Localhost Security Threat Model — Mission 2

**Audit date:** 2026-08-21 (Asia/Kolkata)

## Assets and trust boundary

BOB is a privileged local HTTP service. When a session is configured, it can
send Google requests authenticated with cookie material, SAPISID-derived
authorization, and Scotty upload credentials. The gateway also receives user
prompts, tool-like text, remote image URLs, and image bytes.

The relevant boundary is:

```text
web page origin
    │ browser CORS/PNA request
    ▼
127.0.0.1 gateway ── cookie/SAPISID ──> Google Gemini web RPC
```

Private Network Access (PNA) is a browser permission/preflight mechanism. It
is not an application credential and does not prove that the requesting web
origin is trusted.

## Baseline reproduced before the change

With the default empty `api_keys` configuration, the current middleware:

- returns `Access-Control-Allow-Origin: *`;
- returns `Access-Control-Allow-Private-Network: true`;
- accepts the preflight from an arbitrary origin; and
- allows that origin to read ordinary API responses if the browser completes
  the preflight.

The baseline was reproduced against `httptest` before the policy change (the
temporary baseline test was then converted into the current secure assertions
in `internal/server/security_boundary_test.go`). It is recorded here as an
implementation fact, not as a browser exploit claim. A real browser/PNA test
remains useful for release acceptance, but the HTTP policy was sufficient to
identify the trust gap.

## Threat assessment

| Threat | Current status | Consequence |
|---|---|---|
| Malicious page sends JSON API calls to localhost | Reproduced at HTTP policy level | Uses the local session and can spend upstream quota or expose responses to the page |
| PNA is treated as authentication | Incorrect assumption | A browser permission prompt does not bind the request to an approved application |
| API keys are optional | Source-verified | No-key installs have no credential check for API routes |
| Public Web Studio needs cross-origin access | Intentional product requirement | A blanket CORS disable would break the hosted Studio use case |
| Remote image fetch reaches attacker-controlled/private targets | Mitigated by application checks | `FetchImageBytes` now rejects private/local IP and DNS results, nonstandard ports, and cross-host redirects; residual DNS/proxy topology risk remains |
| Query-string API key leaks through logs/history/referrers | Source-verified | Existing compatibility path makes accidental disclosure possible |

## Design options

### Origin allow-list

Use a strict loopback-only default for browser `Origin` headers. Add an
explicit `allowed_origins`/`BOB_GEMINI_FREE_ALLOWED_ORIGINS` list for a known
hosted Studio or LAN deployment. Reflect only an exact approved origin; never
reflect arbitrary input and never use `*` for an origin-bearing request.

This is the minimum compatible change: local Web Studio and local SDKs still
work, while a random web origin fails preflight and actual requests. It is a
browser boundary, not a substitute for API keys or capability tokens.

### Ephemeral local capability token

Generate a random per-process token and require it on browser-origin API
requests. The embedded local Studio can receive it from the gateway response.
This is stronger than origin filtering, but the token must never be exposed
through a public bootstrap endpoint or sent to third-party origins.

### Desktop-only token injection

Wails can inject the capability into its own frontend. This is a strong
desktop path, but it does not solve the hosted Studio or static web bundle
without a separate pairing flow.

### CSRF-like nonce

A nonce proves that a page completed a prior pairing step, but a public
pairing endpoint that returns the nonce is itself an attacker-readable token
minting endpoint. It must be coupled to explicit user interaction or a
local-only channel.

### Explicit remote-Studio pairing

The hosted Studio should require the operator to opt in by configuring its
exact origin and, preferably, an API key in the Studio. A future pairing flow
can issue a short-lived capability after a user-visible local confirmation.

### API keys

Configured API keys remain the strongest general-purpose application
credential in this codebase. They should be sent in `Authorization` or a
dedicated header. Query `?key=` remains only for compatibility and should be
discouraged because URLs are more likely to be logged or copied.

## Chosen minimum change

Implement strict origin filtering now:

1. allow no-origin native clients as before;
2. allow loopback browser origins by default;
3. allow additional exact origins only through explicit configuration;
4. return a failed preflight/403 for an unapproved origin;
5. reflect the exact approved origin and add `Vary: Origin`;
6. keep API-key authorization independent and unchanged;
7. keep `/playground` publicly navigable, but do not grant its API calls to
   an unapproved remote origin.

This preserves local-first behavior and creates an explicit remote-Studio
trust gate. Remote image fetching is also restricted to publicly routable
hosts, rejects private/local DNS results, and does not follow cross-host
redirects. This still does not claim complete CSRF resistance or desktop
pairing; those remain separate follow-up work.
