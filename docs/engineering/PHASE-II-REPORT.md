# BOB Gemini Free — Phase II Report

**Date:** 2026-08-21 (Asia/Kolkata)  
**Workspace:** `/Users/apple31/Documents/BOB-Gemini-Free`  
**Host:** `go1.26.5 darwin/arm64`, `CGO_ENABLED=1`

## Executive conclusion

Phase II converted the repository from a documentation-heavy, partially
verified gateway into a safer engineering baseline. The fragile Gemini Web
RPC seams now have deterministic fixture coverage; the localhost browser trust
boundary is explicit; updater authenticity is verified with a signed manifest;
desktop port selection and health checks are deterministic; aggregate metrics
are local-only; and performance claims are replaced by a reproducible mocked
baseline.

This is not live Google compatibility certification. At audit time, no process
was listening on `127.0.0.1:9610`, no authenticated Google session was used,
and the workspace has no `.git` metadata. Provider behavior, model identity,
entitlements, rate limits, and release publication remain external gates.

## 1. Verified architecture corrections

- The repository is a Go `1.26.5` module with non-zero declared dependencies.
  “Zero dependencies” is now documented as “no separately managed runtime”
  for distributed binaries, not as “no Go dependencies.”
- OpenAI, Anthropic, and Google surfaces are implemented adapter routes. Their
  compatibility is endpoint-specific; tool calling is prompt/Markdown
  emulation in the current implementation, not native Google function calling.
- Token counting is an estimate. Model aliases and context limits do not prove
  the identity or capability of the upstream Google model.
- Image generation is a text-routing/extraction path. Native Imagen/Nano
  Banana RPC fidelity is not established.
- Go gateway metrics are aggregate and process-local. They do not include
  prompts, tool arguments, image bytes, cookies, SAPISID values, or auth
  headers, and they are not transmitted automatically.
- The Wails desktop path is canonical. The Tauri directory is documented as
  legacy and retained because this snapshot has no Git history for a safe,
  reviewable deletion.

The full claim-by-claim evidence record is in
[`VERIFICATION-MATRIX.md`](VERIFICATION-MATRIX.md).

## 2. Mission results

### Mission 0 — Truth and confidence audit

Completed before production changes. The audit reconciled the two supplied
attachments against source and reproducible local commands. It marked “FULL,”
“native,” unlimited, fixed-memory, latency, RAM, context-window, model,
zero-dependency, and token-accuracy claims according to evidence strength.

Baseline commands passed on the audit host: unit tests, race tests, `go vet`,
and `go build`. The endpoint probe to `127.0.0.1:9610` returned connection
refused, so no served-runtime or live-Google claim was promoted to
`VERIFIED_LIVE`.

### Mission 1 — Golden core regression harness

Added deterministic fixtures in:

- `internal/gemini/golden_test.go`
- `internal/format/golden_test.go`
- `internal/multimodal/golden_test.go`
- `docs/engineering/CORE-REGRESSION-HARNESS.md`

Coverage includes sparse payload positions, Unicode/Hindi/multiline/code and
artifact/citation text, thinking fence fragmentation and missing close fences,
cumulative stream deltas, arbitrary byte boundaries, malformed nested JSON,
Bard errors, truncated final frames, retry deduplication after partial output,
deterministic SAPISIDHASH, nested/Unicode/invalid/multiple tool calls, adapter
semantic equivalence, and resumable upload sequencing.

The protected core was changed only after those tests existed:

- `internal/gemini/auth.go` adds `SAPISIDHashAt` so signing is deterministic;
  the existing `SAPISIDHash` wrapper retains wall-clock behavior.
- `internal/gemini/stream.go` adds `Flush` for a valid unterminated final
  frame.
- `internal/gemini/client.go` flushes the parser at EOF and preserves the
  existing retry/dedup state machine.

These are narrow invariant fixes, not cleanup or an upstream wire redesign.

### Mission 2 — Localhost security boundary

The pre-change wildcard-CORS/no-key behavior was reproduced with a local
`httptest` boundary test. The minimum compatible control is now implemented:

- no-origin native requests remain supported;
- loopback browser origins are allowed by default;
- remote browser origins require exact `allowed_origins` configuration or
  `BOB_GEMINI_FREE_ALLOWED_ORIGINS`;
- untrusted origins receive a failed preflight/403 and no reflected origin;
- trusted origins receive exact reflection plus `Vary: Origin`;
- API-key authorization remains an independent control;
- `/healthz` is unauthenticated and local-only;
- `/v1/metrics` remains API-key protected.

Remote image fetching also rejects private/local IP and DNS results,
nonstandard ports, and cross-host redirects. This is defense in depth, not a
claim of complete protection against every proxy, DNS-rebinding, or custom
transport topology. Query-string API keys remain a compatibility risk and
should not be used for sensitive deployments.

### Mission 3 — Supply-chain updater hardening

The updater now requires:

```text
binary + SHA256SUMS + SHA256SUMS.sig + configured Ed25519 public key
```

`BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY` accepts base64 and hexadecimal forms.
Manifest signatures and exact asset checksums are verified before replacement;
missing or invalid verification material fails closed. Unix replacement uses
an isolated same-filesystem temporary file followed by rename. Windows keeps
its platform-specific rollback path.

`internal/updater/golden_test.go` uses an `httptest` release server and isolated
temporary candidates to test valid signatures, tampered manifests, invalid
signatures, checksum mismatches, binary validation, and replacement inputs.
No developer executable was used as a replacement target. Release operators
still need to publish the manifest/signature and configure a trusted key.

### Mission 4 — Runtime reliability

- **Desktop port:** `cmd/desktop/gateway.go` probes the requested port, reuses
  a compatible BOB `/healthz`, otherwise binds a safe loopback port, exposes
  the actual endpoint to Wails, reports startup errors, and shuts down the
  owned server during the Wails lifecycle. Tests cover compatible reuse,
  collision fallback, and non-BOB rejection.
- **Docker health:** `GET /healthz` returns stable `{"status":"ok"}` without
  Google, GitHub, or credentials. Dockerfile and Compose healthchecks use it.
- **Browser fallback:** startup now makes one fallback decision through
  `launchStudioOrFallback`; a regression test protects against the duplicate
  execution caused by the old error-scope bug.

### Mission 5 — Tool-calling fidelity lab

The torture suite covers OpenAI, Anthropic, and Google shapes; nested schemas,
enums, arrays, nullable values, Unicode, large arguments, multiple calls,
malformed output, accidental Markdown, choice modes, result continuation, and
streaming behavior.

Current classification is documented in
[`TOOL-CALLING-FIDELITY.md`](TOOL-CALLING-FIDELITY.md): declaration transport,
tool results, choice enforcement, and streaming are **EMULATED_PARTIALLY**;
tested serialization/parsing of nested schemas and valid arguments is
**EMULATED_RELIABLY** within fixture bounds; native Google function calling is
**UNKNOWN**. No “full native” claim remains justified.

### Mission 6 — Upstream protocol adaptation

[`ADR-001-UPSTREAM-PROTOCOL-BOUNDARY.md`](ADR-001-UPSTREAM-PROTOCOL-BOUNDARY.md)
defers the proposed `UpstreamProtocol` interface. There is one concrete,
volatile upstream, and the existing payload/parser/stream seams provide better
fixture visibility without a behavior-changing factory abstraction. The ADR
defines the evidence required before revisiting the boundary.

### Mission 7 — Observability

Added `internal/metrics` with bounded counters and latency histograms for
requests, in-flight work, upstream attempts/errors/429s, stream retries,
session-pool state/failovers, image uploads/cache, and estimated tokens.
Authenticated `GET /v1/metrics` exposes safe aggregate JSON; `/` includes the
same aggregate view; `/healthz` remains minimal. Metrics reset on restart and
are never sent externally.

### Mission 8 — Real benchmark baseline

Added `cmd/benchmark-local` and the in-process mocked gateway benchmark. The
latest reproducible run was:

```text
go run ./cmd/benchmark-local -requests 100
```

on `go1.26.5 darwin/arm64`, with 100/100 successful requests at every profile:

| Concurrency | P50 | P95 | P99 | Throughput | Allocs/request | RSS reported |
|---:|---:|---:|---:|---:|---:|---:|
| 1 | 0.124 ms | 0.223 ms | 0.457 ms | 6,381 req/s | 330.57 | 20.1 MiB |
| 10 | 0.785 ms | 1.885 ms | 2.093 ms | 10,954 req/s | 338.29 | 21.9 MiB |
| 50 | 3.149 ms | 6.666 ms | 7.004 ms | 13,392 req/s | 369.41 | 24.7 MiB |
| 100 | 7.872 ms | 10.120 ms | 10.147 ms | 9,662 req/s | 388.26 | 28.1 MiB |

Full methodology and all percentiles are in
[`LOCAL-BENCHMARK-2026-08-21.md`](LOCAL-BENCHMARK-2026-08-21.md). These are
local mocked gateway numbers, not Google latency, quota, or rate-limit claims.

### Mission 9 — Desktop consolidation

The comparison found no required Tauri capability absent from Wails. Wails is
canonical; Tauri remains legacy and is retained only because the source
snapshot lacks Git history for a safe archive/delete operation. Active build
and engineering documentation points to Wails.

### Mission 10 — Documentation reconciliation

Updated the English and Hindi README claims, agent guides, API docs, Docker
health documentation, image-generation documentation, desktop documentation,
and the engineering matrix. Documentation now distinguishes implemented,
emulated, tested, measured, upstream-dependent, and experimental behavior.

## 3. Validation evidence

The final validation gate passed:

```text
go test -count=1 ./...
go test -count=1 -race ./...
go test -count=1 -cover ./...
go vet ./...
go build ./...
go run ./cmd/benchmark-local -requests 100
```

Latest package coverage is not uniformly 80%: `internal/gemini` 53.2%,
`internal/server` 36.6%, `internal/updater` 28.4%, and `internal/service`
0.0% are examples. Stronger deterministic core coverage is not being confused
with whole-package coverage.

## 4. Remaining fragility and unresolved assumptions

1. Google’s undocumented RPC schema, sparse-array semantics, model identity,
   context limits, rate limits, and entitlements still require an authorized
   live conformance run.
2. No current gateway process was reachable at `127.0.0.1:9610` during the
   audit. The report therefore contains no served-runtime or live-Google pass.
3. Live authenticated Scotty upload, vision, Imagen/Nano Banana generation,
   Pro routing, browser login, and Wails GUI acceptance remain unrun.
4. Native Google function calling has not been evidenced; prompt extraction
   remains vulnerable to model noncompliance and hostile tool-like text.
5. Remote image SSRF controls are application-level checks. A release
   deployment should additionally constrain egress at the network/transport
   layer and test DNS rebinding behavior.
6. API keys in query parameters can leak through URLs, logs, and referrers;
   header-based credentials should be required for exposed deployments.
7. Legacy diagnostic tests still contain best-effort upstream-aware paths and
   permit 502 responses. They are not a substitute for live acceptance.
8. The updater’s security contract is fail-closed only after release assets and
   the trusted public key are published and distributed correctly.
9. The workspace has no `.git` directory, so branch, commit, remote, and
   historical Tauri provenance cannot be verified.

## 5. Files deliberately left untouched

- `internal/gemini/payload.go` was left unchanged; sparse wire construction is
  protected by fixtures rather than refactored.
- No broad rewrite of `internal/gemini/parse.go`, model mappings, response
  formats, HTTP architecture, or tool extraction was attempted.
- Tauri source/configuration remains present as legacy material because this
  snapshot cannot provide a recoverable Git deletion.
- No live cookies, config secrets, release keys, or protected runtime data were
  added to the workspace.
- No real running executable was replaced during updater testing.

## 6. Recommended next engineering phase

1. Run an explicitly authorized live conformance matrix against a known
   session, recording exact upstream request/response fixtures and account
   identity separately from local adapter behavior.
2. Add a release pipeline that signs `SHA256SUMS`, publishes the public key
   through a trusted channel, and tests the updater against real release API
   responses without replacing a live executable.
3. Decide whether hosted Studio access needs a short-lived capability/pairing
   token in addition to exact origins and API keys; validate with real browser
   PNA behavior.
4. Separate the legacy upstream-aware diagnostic tests from the hermetic core
   suite and add explicit live-test markers.
5. Only revisit native tool calling or an upstream interface after capturing
   evidence of the Google schema and a measurable reduction in coupling.

## Final success criterion

BOB Gemini Free is now safer to evolve: fragile behavior is executable as
deterministic tests, security and supply-chain boundaries are explicit, and
unsupported performance/compatibility claims are identified. The remaining
work is intentionally live/provider and release-operation work, not speculative
architecture rewriting.
