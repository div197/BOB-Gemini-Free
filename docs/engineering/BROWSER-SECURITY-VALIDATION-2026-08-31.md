# Browser Security Boundary Validation — 2026-08-31

> Historical evidence record. The browser run below used the source checkpoint
> recorded in this document; subsequent protected-main follow-ups are recorded
> in the current release audit. The observed cross-port rejection remains valid
> evidence, but this file is not a claim about the immutable Preview 5 bytes.

## Scope

This is a real in-app-browser check of the local HTTP origin boundary at the
historical source checkpoint (`726dd049fe9436bb22a89abcf74f961f86f9996c`). It
tests whether a different browser origin can read or drive the loopback
gateway. It does not claim that a public HTTPS deployment, browser Private
Network Access policy, or a future capability-token pairing design has been
validated.

- Gateway: `http://127.0.0.1:19613`
- Unrelated browser origin: `http://127.0.0.1:19614`
- Browser surface: Codex in-app browser
- Credentials: none; no Google account, cookie, provider key, or gateway key
- Upstream: no request was made; the probes stopped at the gateway boundary

## Procedure

1. Serve a temporary attacker page from `127.0.0.1:19614`.
2. From that page, attempt a cross-origin `GET /v1/models`.
3. From that page, attempt a JSON `POST /v1/chat/completions`, which requires
   a CORS preflight.
4. Repeat the preflight at the HTTP level with
   `Access-Control-Request-Private-Network: true`.
5. Load the intended same-origin Studio at
   `http://127.0.0.1:19613/playground` and inspect its rendered status and
   browser diagnostics.

## Observed result

The attacker page reported `TypeError: Failed to fetch` for both the models
read and the JSON POST attempt. The gateway-side confirmation for the models
request was:

```text
HTTP 403 Forbidden
{"error":{"message":"origin is not allowed","type":"invalid_request_error"}}
```

The response had `Vary: Origin`, no `Access-Control-Allow-Origin`, and did not
reach an upstream provider. The explicit PNA preflight was also rejected with
the same `403 origin is not allowed` response. This confirms that a different
loopback port is not treated as the gateway's trusted browser origin.

The intended Studio loaded from the gateway origin, rendered `Gateway online
at 127.0.0.1:19613`, and produced no browser error or warning logs during the
check. An exact-origin health response reflected the exact origin and returned
the stable `{"status":"ok"}` health body.

## Interpretation and remaining gate

The source-level origin allow-list and the real browser cross-port behavior
agree. Explicitly configured remote origins remain an operator trust decision;
they are not made safe merely by CORS or PNA. A public HTTPS-origin test, a
browser implementation test involving a non-loopback public origin, and a
decision on whether remote Studio needs an ephemeral capability token remain
release gates. No capability-token feature was added speculatively.
