# Health, Diagnostics & Concurrency Benchmarking

BOB Gemini Free includes built-in automated test and diagnostic tooling to verify upstream connectivity, token parsing, and concurrency throughput.

---

## 1. Engine Health Endpoint (`GET /`)

```bash
curl http://127.0.0.1:8081/
```

### Response
```json
{
  "models": [
    "gemini-3.7-flash",
    "gemini-3.7-flash-thinking",
    "gemini-3.1-pro",
    "gemini-nano-banana-2",
    "imagen-3"
  ],
  "status": "ok",
  "version": "v0.1.0"
}
```

---

## 2. Automated 12-Point Diagnostic Suite (`--test`)

Run the 12-point diagnostic test against any local or remote BOB gateway:

```bash
./bob-gemini-free --test --test-url http://127.0.0.1:8081
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

---

## 3. Concurrency & Stress Benchmark (`--bench`)

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
