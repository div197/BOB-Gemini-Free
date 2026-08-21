# Live Conformance Runbook

This runbook is the explicit boundary between hermetic engineering tests and
provider-facing acceptance. It must be run only with an authorized Google
session and a disposable test account or an account whose operator accepts the
provider risk. Never commit cookies, SAPISID values, raw upstream payloads
containing personal data, or API keys.

## Test classes

The default suite is intentionally safe and does not require Google network
access:

```bash
go test -count=1 ./...
go test -count=1 -race ./...
go vet ./...
go build ./...
```

The `internal/diag/agent_test.go` workflow tests are build-tagged `live`.
They exercise the real upstream path through a locally constructed gateway and
may contact Google:

```bash
go test -count=1 -tags=live ./internal/diag
```

The CLI diagnostic runner is a separate live check against an already running
gateway. It is not part of the default Go test suite:

```bash
./bob-gemini-free --test --test-url http://127.0.0.1:9610
# If the gateway has API keys:
./bob-gemini-free --test --test-url http://127.0.0.1:9610 --test-key "$BOB_TEST_API_KEY"
```

`GET /healthz` proves only that the local process is serving. It does not prove
Google session validity, model entitlement, or upstream readiness.

## Evidence matrix

| Area | Minimum observation | Evidence to retain | Verdict |
|---|---|---|---|
| Local boundary | `/healthz`, `/v1/models`, `/v1beta/models` | status, version, model IDs, timestamp | local-only |
| Text | ordinary, multiline, Unicode, Hindi/Indic, code block | sanitized request/response fixture and route | pass/fail per route |
| Thinking | complete and fragmented reasoning stream | redacted stream fixture plus final text | pass/fail per splitter invariant |
| Streaming | cumulative, duplicate, truncation, retry, malformed frame | packet/fixture IDs and emitted deltas | pass/fail per invariant |
| Authentication | anonymous, cookie session, selected `auth_user` | account label only; never cookie contents | provider/session-dependent |
| Multimodal | authenticated Scotty upload and vision response | image hash, dimensions, MIME, status; no image bytes | live-only |
| Images | Imagen/Nano Banana route, if entitled | provider response class and extracted URL shape | experimental/live-only |
| Adapters | OpenAI, Anthropic, Google equivalent prompts | normalized semantic result and SSE event sequence | endpoint-specific |
| Tools | one/multiple/nested/Unicode/result continuation | tool-call JSON and refusal/malformed behavior | emulation fidelity |
| Performance | 1/10/50/100 concurrency profiles | host, Go version, date, session class, P50/P90/P95/P99 | measured only for that run |

## Stop conditions

Stop and record a failure when the upstream returns a policy, authentication,
rate-limit, or schema error. Do not convert a provider failure into a local
pass, and do not retry aggressively against a live account. A `502` from the
gateway is an observed upstream failure unless the test specifically asserts
the gateway's error mapping.

## Reporting format

Every live report should include:

- commit, tag, host OS/architecture, Go version, and build flags;
- whether the gateway was CLI, Docker, Wails, or embedded Go;
- session class and account label without secrets;
- exact test command and timestamp/time zone;
- per-route result and sanitized fixture references;
- provider errors, HTTP status, retry count, and rate-limit observations;
- what remains unknown.

Live results are evidence for the tested account, date, route, and upstream
behavior. They do not establish universal compatibility, fixed context size,
unlimited access, model identity, or a permanent latency guarantee.
