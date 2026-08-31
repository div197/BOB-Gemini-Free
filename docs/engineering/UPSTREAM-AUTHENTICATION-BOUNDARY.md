# Upstream Authentication and Shared-Network Boundary

**Status:** Current v0.2 source / public Preview 8 engineering truth

The downloadable macOS package is the public prerelease `v0.2.0-preview.8`,
built from reviewed runtime source target `bfa68ff`. Earlier Preview 6,
Preview 5, and Preview 7 references below describe historical/current-fleet
boundaries and are not the current package identity.

This document defines what BOB Gemini Free does and does not authenticate.
It exists because “API-free” is easy to misunderstand: BOB does not require a
BOB cloud account or a Google AI Studio API key, but it still talks to an
upstream Google web service whose access policy is outside the gateway.

## The actual request path

```text
student's desktop/browser
        |
        v
local BOB gateway (127.0.0.1, or an explicitly configured LAN endpoint)
        |
        v
Google Gemini web RPC (undocumented StreamGenerate/batchexecute surface)
```

The gateway does not call the official Google Gemini API endpoint with a
project API key. It builds the reverse-engineered web-RPC payload and sends it
to `gemini.google.com`. That upstream boundary is deliberately treated as
volatile protocol code; local protocol compatibility is not a provider
entitlement.

## What “no API key” means

In the default configuration:

- BOB creates no BOB account, server-side conversation session, or cloud
  memory database.
- No Google AI Studio project key or BOB billing account is required.
- If no cookie file or cookie pool is configured, BOB does not load a Google
  cookie from disk and does not attach SAPISID authentication from disk.
- The request can still be accepted or rejected by Google's anonymous web
  experience and its current anti-abuse, traffic, network, and product rules.

Therefore “no cookie configured” is not a promise of unlimited anonymous
access, and “no API key” is not the same as “no upstream session or policy.”
The provider may apply state and limits that are not visible to the local
gateway. Those facts must be reported as upstream-dependent rather than
invented as a local quota.

## Authenticated mode

When a user explicitly configures their own `cookie.txt` or cookie pool, the
gateway may send the corresponding Google web cookie, SAPISIDHASH authorization,
and optional `auth_user` profile selection. This is per-installation local
state. It is not a BOB account and it does not turn the web RPC into the
official Google API.

Cookie pools provide bounded local routing and failure cooldown for explicitly
supplied sessions. They do **not**:

- create a higher quota;
- make a session permanent;
- guarantee model identity, vision, Imagen, context, or concurrency;
- bypass Google's rate limits or account policy; or
- make it safe to copy one person's cookie to student devices.

Each user must use only a session they are authorized to use. No release
artifact contains a teacher cookie, and no release procedure should collect
cookies, prompts, or raw authorization headers.

## Shared school networks and 20–30 devices

If multiple Macs use one school router, Google may see the same public egress
IP. That is normal network topology, not evidence that BOB is broken. The
correct response is not IP disguise, fingerprint spoofing, proxy rotation, or
cookie sharing. Those approaches would obscure accountability and can violate
provider rules.

The safe rollout controls are:

1. install and health-check devices in small waves rather than sending a
   simultaneous generation burst;
2. use one authorized local/provider session per student where authentication
   is required;
3. stop automatic retries for explicit provider policy failures such as
   401/403/429 and Bard rate/auth rejection;
4. retain retry-on-transport and stream deduplication only where a network
   failure makes a retry reasonable; and
5. record only version, OS/architecture, local health, request outcome class,
   and timestamp—not cookies, prompts, or image data.

No local change can guarantee that Google will treat shared-network traffic as
independent. The release must say that provider availability remains
account/session/network dependent.

## Failure interpretation

| Observation | What it proves | What it does not prove |
|---|---|---|
| `GET /healthz` returns `{"status":"ok"}` | The local gateway process answers | Google accepts a request or a session is valid |
| Local port is reachable | Desktop/gateway wiring is working | Upstream quota, model identity, or response generation |
| HTTP 401/403/3xx from Google | Provider/auth/network policy rejected the request | That another model, cookie, or IP should be rotated to bypass it |
| HTTP 429 or Bard 1024/42901 | The provider asked the client to slow down | A local retry will restore quota |
| Cookie pool has healthy entries | Local cookie files parsed and are not in local cooldown | The accounts have independent quota or permission |
| A completed local update | The app replacement path succeeded | The updated app can generate through Google's current web service |

For an explicit HTTP 401/403, the local client clears only its cached dynamic
`/app` page token and build identifier. The configured cookie file is retained,
the rejected POST is not replayed, and the next request performs a fresh
bootstrap. This can recover from stale page-token material; it cannot
reauthenticate a revoked cookie or override Google's access policy.

## Student-facing product wording

The honest description for the preview is:

> BOB Gemini Free is a local desktop gateway and studio. It does not require a
> BOB cloud account or Google AI Studio API key. Text, image, model, quota, and
> session behavior depends on the current Google web experience and the user's
> own authorized session. Local health and app updates do not certify upstream
> availability.

This boundary is a release requirement. Do not use “unlimited,” “guaranteed
free,” “full native Gemini API,” “quota bypass,” or “one shared student
session” as product claims.
