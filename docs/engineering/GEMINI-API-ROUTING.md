# Gemini Developer API Routing and Student Limits

**Status:** implemented in the v0.2 milestone; current source is public `main`
at `1b3472f` before the route-clarity patch in this review, and the public
desktop package remains Preview 4 until the
Preview 5 publication gates pass
**Audit date:** 2026-08-31 (Asia/Kolkata)

This document records the second provider path added to BOB Gemini Free. It is
not a promise that Google accepts every model alias, account, key, image, or
request shape. The default web-session path remains the primary BOB behavior;
the Developer API path is explicit and opt-in.

## Two different upstream paths

```text
                 ┌────────────────────────────────────────────┐
                 │ BOB Web Studio / OpenAI-shaped client       │
                 └───────────────────┬────────────────────────┘
                                     │
                 no provider key     │     explicit provider key
                                     │
          ┌──────────────────────────┴─────────────────────────┐
          │                                                    │
          v                                                    v
  Google Gemini web RPC                              Google Gemini Developer API
  gemini.google.com                                  generativelanguage.googleapis.com
  cookie/guest/session policy                        user's Google Cloud project/key
  undocumented protocol                              documented REST + SSE surface
```

The two paths must not be conflated:

| Question | Default web-session route | Explicit Developer API route |
|---|---|---|
| BOB account required? | No | No |
| Google AI Studio key required? | No, though web access remains provider-controlled | Yes, the user supplies one key or configures one process environment variable |
| Upstream contract | Undocumented web RPC, session/cookie/guest dependent | Official Gemini API REST/SSE contract |
| Quota owner | The Google web session/account/network context | The Google Cloud project associated with the key |
| Automatic key rotation? | No | No |
| Automatic fallback between paths? | No after a request is selected or a stream starts | No |
| BOB stores the provider key? | Not applicable | The Web Studio keeps it in memory only; it is not in `localStorage`, config JSON, logs, metrics, or release assets |
| Student guarantee | No fixed free, anonymous, model, or unlimited guarantee | Google free-tier/paid-tier, model, project, billing, and policy limits apply |

## Supported explicit route

The current source supports the Developer API key on:

- `POST /v1/chat/completions` for text, data-URL images, streaming, generation
  parameters, JSON-object output, and native function declarations/tool choice
  translation;
- `POST /v1beta/models/{model}:generateContent` and
  `:streamGenerateContent`, forwarding the validated native Google request; and
- `POST /v1beta/models/{model}:countTokens`, forwarding a native count request.

The direct chat route maps the documented local UI aliases
`gemini-3.5-flash` and `gemini-flash` to `gemini-3.6-flash`. Other
provider-shaped `gemini-*` IDs are forwarded unchanged, including future IDs,
so BOB does not have to invent a catalog entry for every provider release.
Google remains the authority for whether an ID exists or is available to the
student's project. BOB's larger web-RPC alias catalog is not evidence that
vendor-shaped aliases (`gpt-*`, `claude-*`, and similar names) exist in the
Developer API; those fail with a clear 400 response instead of being silently
reinterpreted.

The explicit key is rejected on the current `/v1/messages`, `/v1/responses`,
and `/v1/images/generations` handlers because those adapters have not been
translated to the Developer API contract. With no explicit key, those routes
continue to use their existing web-RPC behavior. This prevents a user from
believing an API-key request was billed to or served by the provider path when
it was actually sent through a cookie/session route.

`POST /v1/tokens/count` remains a local estimate. It is not Google's
authoritative tokenizer, even when a Developer API key is present.

The local `GET /v1/models` and `GET /v1beta/models` catalogues remain BOB's
web-RPC/adapter catalogues; they are not a live inventory of every Developer
API model or project entitlement. Native clients can provide a current
provider-shaped `gemini-*` ID directly, and Google decides whether that model
is available. BOB intentionally does not turn a static model list into a
promise about future provider releases.

## How students use it

1. Open **Config** in the BOB Studio.
2. Open the optional **Gemini Developer API** section.
3. Follow **Create or manage a key in Google AI Studio**:
   [aistudio.google.com/app/apikey](https://aistudio.google.com/app/apikey).
4. Review the provider's [current models](https://ai.google.dev/gemini-api/docs/models),
   [rate limits](https://ai.google.dev/gemini-api/docs/rate-limits), and
   [pricing/free-tier conditions](https://ai.google.dev/gemini-api/docs/pricing)
   for the student's project.
5. Paste the student's own key, enable **Use for this session**, and choose a
   provider-supported `gemini-*` model ID.
6. Turn the toggle off to return to the default web-session route.

### Config credential map

The Studio's Config modal deliberately presents four separate concepts:

| Config item | Meaning | Stored/sent boundary |
|---|---|---|
| **Gateway Endpoint URL** | Which BOB process receives the request | May be saved as a UI preference; remote endpoints require an explicit trust decision and HTTPS for provider-key use |
| **BOB Gateway Access Key** | Optional authentication for an operator-protected BOB endpoint (`api_keys`) | Page memory only; sent as BOB request authorization; it is not a Google credential |
| **Google Gemini Developer API key** | The student's own Google AI Studio project credential | Page memory only; sent only when the Developer API toggle is enabled, through BOB's dedicated request header |
| **Web session / cookies** | The default reverse-engineered web-RPC identity | Managed by the running engine and its configured cookie state; there is no cookie input in the Studio and cookies must not be pasted into either key field |

The gateway access key and the Developer API key are not interchangeable. The
first controls entry to BOB; the second selects a different Google upstream and
assigns provider quota/billing responsibility to the student's project. If the
gateway has no `api_keys` configured, the BOB access field remains empty. If the
Developer API toggle is off, BOB uses the default web-session route and does not
silently fall back between provider paths.

### Runtime route status and guards

The Config modal shows the route that the current Studio page will use before a
prompt is sent. It reports the BOB gateway door, Developer API-key state,
engine-owned web-session state, and the model guard. The status is intentionally
descriptive rather than a health claim: a successful local ping proves that BOB
is reachable, not that Google will accept the selected session, key, model, or
quota.

When the Developer API route is selected, the Studio blocks Send before adding
the user turn to chat history or issuing a network request if any of these are
true:

- no Developer API key is present;
- the saved endpoint is not loopback HTTP or an explicitly saved HTTPS endpoint;
- the selected model is a BOB/vendor alias, Imagen route, thinking suffix, or
  non-default `@think` mode that the direct adapter does not translate.

The two key fields have independent **Clear** actions. Closing Config clears
their DOM input values while retaining neither credential in browser storage;
reopening the modal rehydrates only the current page-memory value. This reduces
accidental exposure in a shared classroom WebView but is not a security boundary
against JavaScript that already runs in the same page.

The UI key is held only in page memory and is sent only as
`X-BOB-Gemini-API-Key` on generation requests. BOB translates it to the
official upstream `x-goog-api-key` header. It is not sent on health, metrics,
model-list, update-check, or ping requests. The custom header is separate from
BOB's local gateway `Authorization`/`x-api-key` authentication.

The optional BOB gateway-auth token is held only in the current page session as
well. It is never written to browser storage and must be entered again after a
reload; this deliberately avoids leaving a reusable local-gateway credential
on a shared student computer.

For the same reason, a hosted public Studio cannot send this header until the
user explicitly saves a trusted gateway endpoint. A loopback-served Studio or
the native Wails app may use the local gateway directly. Saving a LAN or remote
endpoint is an explicit trust decision; use HTTPS and an application API key
when the endpoint is not on the same machine. Never paste a provider key into
a public demo whose gateway ownership and transport you have not verified.

For a controlled CLI process, the only supported environment form is one key:

```bash
export BOB_GEMINI_FREE_GEMINI_API_KEY='YOUR_OWN_KEY'
./bob-gemini-free
unset BOB_GEMINI_FREE_GEMINI_API_KEY
```

Do not put this value in `config.json`, source code, shell scripts committed to
Git, screenshots, issue reports, or a shared classroom key pool. Do not set a
comma-separated list. A key list/rotation would make quota ownership opaque
and could evade provider limits. Each student should use a key and project
they are authorized to use.

## Free-tier and rate-limit truth

Google's official documentation describes a free tier for selected models and
separately describes paid billing. Exact requests-per-minute, requests-per-day,
token, model, grounding, and account limits are provider-controlled and can
vary by model, project, region, account state, and date. Google directs users
to view the current limits in AI Studio; BOB therefore does not hardcode a
universal `15 RPM`, `1,500 RPD`, or “until midnight” promise.

Authoritative references for release reviews:

- [Gemini API pricing and free tier](https://ai.google.dev/gemini-api/docs/pricing)
- [Gemini API rate limits](https://ai.google.dev/gemini-api/docs/rate-limits)
- [Gemini API models](https://ai.google.dev/gemini-api/docs/models)
- [Gemini API billing](https://ai.google.dev/gemini-api/docs/billing)
- [Gemini API key usage](https://ai.google.dev/gemini-api/docs/generate-content/api-key)
- [Generate content and streaming reference](https://ai.google.dev/api/generate-content)

The free tier is not the same thing as “unlimited.” Linking a paid billing
account or using a key from a paid project can create charges. Students must
check the project and billing state before using a key. Provider terms and
Google's current data-use settings also apply; BOB cannot change them.

The web-session path has a different boundary. It may work without a cookie,
but anonymous web acceptance, session expiry, account entitlement, shared
school egress IP, model access, and anti-abuse decisions remain upstream
dependent. A local `/healthz` success proves only that BOB is running.

## Failure and retry policy

- A Developer API 401/403 is reported as a key/project rejection.
- A Developer API 429 is reported as a quota/rate-limit condition with a link
  to AI Studio guidance.
- Transport errors may be retried only by the caller's normal request policy;
  BOB does not replay a started web request through the Developer API or the
  reverse direction.
- The web-RPC cookie pool is not used by the explicit provider route.
- Errors are sanitized and never include the provider key or raw auth headers.

This is deliberate for classroom accountability: changing the provider should
be an observable user choice, not an invisible quota workaround.

## Evidence and maintenance gate

The transport, translation, native tool declaration, SSE parsing, invalid-key,
secret-redaction, and server-routing behavior are covered by deterministic
tests in:

- `internal/geminiapi/geminiapi_test.go`;
- `internal/config/config_test.go`; and
- `internal/server/gemini_api_test.go`.

Before each preview/release, maintainers should re-check the official links,
the current supported model IDs, and the UI wording. Update the matrix and
student notices with a dated source review. Never convert a provider webpage
or an old changelog number into a permanent product guarantee.
