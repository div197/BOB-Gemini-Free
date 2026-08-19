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
  "version": "v0.1.3",
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
  • Gateway Status:        ok (Version v0.1.3)
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

## 3. Automated 13-Point Diagnostic Suite (`--test`)

Run the 13-point diagnostic test against any local or remote BOB gateway:

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
8. `Google Native Gemini API Format`
9. `OpenAI Codex CLI Responses API Format`
10. `Anthropic Messages API Protocol (POST /v1/messages)`
11. `OpenAI Function Calling & Tool Invocation`
12. `Image Generation & Gemini Nano Banana Pipeline`
13. `Token Counting Engine (Google :countTokens & OpenAI /v1/tokens/count)`

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
