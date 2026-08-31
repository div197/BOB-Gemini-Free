# Health, Diagnostics & Concurrency Benchmarking

BOB Gemini Free includes built-in automated test and diagnostic tooling to verify upstream connectivity, token parsing, token counting, and concurrency throughput.

---

## 1. Engine Health & Live Telemetry Endpoint (`GET /`)

```bash
curl http://127.0.0.1:9610/
```

### Response
```json
{
  "status": "ok",
  "version": "v0.1.7",
  "models": [
    "gemini-3.7-flash",
    "gemini-3.7-flash-thinking",
    "gemini-3.1-pro",
    "gemini-nano-banana-2",
    "imagen-3"
  ],
  "requests_served": 142,
  "tokens_processed": 185420,
  "estimated_savings_usd": "$0.70",
  "uptime_seconds": 3600,
  "pool_sessions_total": 3,
  "pool_sessions_healthy": 3
}
```

---

## Container Health Probe (`GET /healthz`)

Use the dedicated unauthenticated probe for Docker/Kubernetes health checks:

```bash
curl http://127.0.0.1:9610/healthz
```

It performs local computation only and returns stable `{"status":"ok"}` JSON.
The telemetry route `/` may require an API key when authentication is configured.

## 2. Live CLI Telemetry Dashboard (`--status`)

Query live metrics directly from your terminal:

```bash
./bob-gemini-free --status --test-url http://127.0.0.1:9610
```

```
==================================================================
    BOB Gemini Free - Live Gateway Telemetry & Status             
    Break Ordinary Boundaries | ABCsteps (https://abcsteps.com)   
==================================================================
  • Gateway Status:        ok (Version v0.1.7)
  • Target Gateway URL:    http://127.0.0.1:9610
  • Server Uptime:         3600 seconds (60.0 minutes)
  • Requests Served:       142 requests
  • Tokens Processed:      185420 tokens
  • Estimated USD Savings: $0.70 (vs commercial cloud APIs)
  • Active Models Loaded:  64 models
  • Cookie Pool Sessions:  3 total, 3 healthy
==================================================================
```

---

## 3. Automated 15-Point Diagnostic Suite (`--test`)

Run the 15-point diagnostic test against any local or remote BOB gateway:

```bash
./bob-gemini-free --test --test-url http://127.0.0.1:9610
```

### Verification Checks Performed:
1. `Gateway Engine Health (GET /)`
2. `OpenAI Models Registry (GET /v1/models)`
3. `Single Model Lookup (GET /v1/models/gemini-3.7-flash)`
4. `Gemini 3.7 Flash Fast Completion`
5. `Gemini 3.7 Flash Deep Reasoning`
6. `Real-time SSE Delta Stream & Usage`
7. `Developer Role & JSON Output Enforcement`
8. `Google-shaped Gemini Adapter Format`
9. `OpenAI Codex CLI Responses API Format`
10. `Anthropic Messages API Protocol (POST /v1/messages)`
11. `OpenAI Function Calling & Tool Invocation`
12. `Image Generation & Gemini Nano Banana Pipeline`
13. `Token Counting Engine (Google :countTokens & OpenAI /v1/tokens/count)`
14. `Claude Code SSE Streaming Tool Execution Protocol`
15. `StreamFlight Concurrency Multiplexing (5 Parallel Coalesced Requests)`

The suite is fail-closed: malformed JSON, empty or `[DONE]`-only streams,
metadata-only responses, and unavailable provider-dependent image generation
are reported as failures. A green local run proves the responses observed from
that gateway; it does not prove Google quota, session longevity, model
identity, or clean-device release behavior.

---

## 4. Concurrency & Stress Benchmark (`--bench`)

Run a load test with concurrent workers:

```bash
./bob-gemini-free --bench --bench-concurrency 5 --bench-requests 10
```

### Metrics Output:
- **Completed & Failed Requests**
- **Total Elapsed Time**
- **Average Latency**
- **Median Latency (P50)**
- **90th Percentile Latency (P90)**
- **99th Percentile Latency (P99)**
- **Request Throughput (req/sec)**
- **Token Generation Throughput (tokens/sec)**

The live benchmark caps the requested workload at 128 workers and 10,000
requests. Invalid, empty, oversized, or malformed JSON responses count as
failures. Token throughput is shown only when every successful response
includes positive provider-reported usage; otherwise it is reported as
unavailable.

---

## 5. Release Update Check Endpoint (`GET /v1/update/check`)

Check for newer GitHub releases programmatically. The response follows the
running gateway build's release channel: the default CLI/embedded constructor
checks stable releases, while a native preview build checks stable first for a
one-way migration and then the published preview channel when no stable update
exists. This endpoint only reads release metadata; it never downloads,
replaces, or restarts the application.

```bash
curl http://127.0.0.1:9610/v1/update/check
```

### Response
```json
{
  "current_version": "v0.2.0-preview.6",
  "latest_version": "v0.2.0-preview.7",
  "has_update": true,
  "channel": "preview",
  "asset_available": true,
  "manifest_available": true
}
```

The native Wails app uses **Help → Check for Updates** for installation. The
web Studio status badge is informational: on a native build it points users to
the native Help action, on a preview CLI/browser route it points to the
official Releases page, and on a stable CLI route it retains the `--update`
command guidance. An unavailable metadata check returns HTTP 200 with
`has_update: false` and a bounded `error` message so the status probe cannot
interrupt the Studio.
