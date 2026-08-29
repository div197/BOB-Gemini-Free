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

Historical boundary: this Phase II report records the alternate desktop
wrapper as retained because that snapshot lacked Git history. A later Git-backed
Phase III follow-up compared it with Wails, smoke-tested it, and removed it from
the active tree while preserving recovery through Git history.

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
nonstandard ports, and cross-host redirects. The exported fetch seam now
requires a guardable HTTP client whose direct transport re-resolves the host
immediately before dialing and connects only to an approved literal public IP;
configured proxies are not used for remote-image fetches. Live egress policy
and network topology remain separate acceptance gates. Query-string API keys
remain a compatibility risk and should not be used for sensitive deployments.

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

The earlier Phase-II baseline package coverage was not uniformly 80%:
`internal/gemini` 53.2%, `internal/server` 36.6%, `internal/updater` 28.4%,
and `internal/service` 0.0% were examples. This historical measurement is kept
for provenance; the 2026-08-29 addendum records the later coverage run.

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
5. Remote image SSRF controls now include a guarded final DNS lookup and
   literal-IP dial in the production HTTP-client path. A release deployment
   should additionally constrain egress at the network layer and test the
   deployed DNS/proxy topology.
6. Header-based credentials are the default. Legacy gateway query-key
   authentication is disabled unless explicitly opted in; URLs can still leak
   credentials when an operator enables that compatibility mode, so it remains
   unsuitable for exposed deployments.
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

## Historical Phase II success criterion

BOB Gemini Free is now safer to evolve: fragile behavior is executable as
deterministic tests, security and supply-chain boundaries are explicit, and
unsupported performance/compatibility claims are identified. The remaining
work is intentionally live/provider and release-operation work, not speculative
architecture rewriting.

## 7. 2026-08-29 addendum — student-owned Gemini Developer API route

The dated report above remains the historical Phase II snapshot. The current
working branch adds an explicit, opt-in Gemini Developer API path without
changing the default web-session route:

- Web Studio students can open **Config → Gemini Developer API**, use the
  official Google AI Studio key page, review the current model, rate-limit, and
  pricing pages, and choose **Use for this session**.
- The provider key is held in page memory, sent only on generation requests,
  and never stored in `localStorage`, `config.json`, logs, metrics, update
  checks, or release assets. There is no key pool, rotation, or silent
  cross-provider fallback.
- The isolated route supports OpenAI-shaped chat plus native Google
  `generateContent`, streaming, and `countTokens` forwarding. Provider-shaped
  `gemini-*` IDs are passed to Google for acceptance; the local model list is
  not presented as a live entitlement catalogue.
- Explicit provider-key use is rejected on adapters that have not been
  translated to the Developer API contract. This is intentional fail-closed
  behavior, not missing fallback logic.
- Google controls free-tier availability, rate limits, billing, model access,
  and policy. The current student procedure and official links are maintained
  in [`GEMINI-API-ROUTING.md`](GEMINI-API-ROUTING.md); no fixed RPM/RPD or
  unlimited-access promise is made.

The current local validation also passed `make web`, JavaScript syntax
validation, `go test -count=1 ./...`, `go test -race -count=1 ./...`,
`go vet ./...`, `go build ./...`, `go mod verify`, `git diff --check`, and a
repository secret-pattern scan. No live Google key or account was used, so
provider acceptance and real quota behavior remain unverified external gates.

The continuation also tightened the student bootstrap and provider-stream
boundaries: installers now verify a fixed Ed25519-signed manifest and exact
asset digest with no unsigned default path; release operators can record a
non-secret local receipt tying the asset set to a clean source commit and
toolchain; generation POST retries are limited to known pre-connection
transport failures; direct Developer API tool calls are assembled and emitted
only after a successful bounded stream; unknown candidate/finish ambiguity and
tool-result correlation errors fail closed; and web-RPC/direct API stream
failures are sent as structured errors rather than fabricated assistant text.
These are local protections, not substitutes for clean Windows/macOS package
acceptance, Apple/Windows publisher trust, public release reconciliation, or
live Google behavior.

## 8. 2026-08-29 addendum — 100-path reliability continuation

The follow-up audit converted the remaining concerns into the numbered
[`FAILURE-REGISTER-100.md`](FAILURE-REGISTER-100.md). It does not relabel open
risks as complete. The local implementation continuation added:

- local `/manifest.json` and `/sw.js` serving with versioned, API-excluding
  service-worker behavior;
- bounded StreamFlight subscribers/history with explicit slow-subscriber and
  cancellation errors, plus session-safe request coalescing rules;
- bounded upstream response bodies/stream lines and nil/status/timeout guards;
- secure cookie-file hashing/reload and pool deduplication/cooldown behavior;
- bounded image/upload input and strict Scotty URL/response validation;
- final remote-image DNS revalidation with a direct literal-public-IP dial and
  rejection of unguardable HTTP-client seams;
- updater redirect, metadata, asset-size, exact-byte, and staged-package
  guardrails;
- artifact token normalization and empty-editor recovery; geometry-driven
  top/bottom chat controls; generated public-bundle synchronization; and
  bounded client attachment/history storage.
- atomic per-install updater locking with conservative stale-lock recovery,
  post-lock target revalidation, and conservative cleanup of old committed
  staging plans;
- ZIP extraction rejection for ambiguous paths, duplicates, symlinks, and
  special files;
- bounded, single-cookie-scoped Scotty reference caching with explicit
  multi-account-pool disablement and concurrent upload single-flight; provider
  reference expiry remains open and upstream-dependent.
- bounded operational retry configuration plus capped exponential jitter and
  `Retry-After` handling for transient upstream failures, while retaining
  immediate failure for 429, policy, and provider rejection responses.
- explicit browser failure-state handling for `finish_reason: error`, broken
  transport, incomplete EOF, and user cancellation; failed/stopped assistant
  turns are not replayed as successful conversation history;
- shared reflection-safe tool-schema budgets for typed and JSON-decoded
  schemas, plus fail-closed provider tool-output validation and explicit nil
  callback/context guards at the stream boundaries.
- strict selected Responses input normalization: unsupported items and content
  blocks now return client-visible validation errors, malformed function calls
  and results fail closed, and `call_id`, `name`, and `tool_calls` survive the
  server conversion into the shared OpenAI message model.
- strict Anthropic input normalization: unsupported content blocks, malformed
  images, invalid tool-use fields, and missing tool-result correlation data now
  fail closed; mixed text/tool-result blocks retain their order before shared
  prompt translation.
- the direct Gemini Developer API SSE parser now rejects an empty or
  `[DONE]`-only stream instead of allowing the server to fabricate a normal
  stop, and rejects empty semantic events; standard comments and multi-line
  data fixtures remain supported.
- native updater recovery now repairs validated interrupted transactions at
  the next startup: healthy candidates are finalized, unconfirmed candidates
  are rolled back, and ambiguous states fail visibly; isolated fixtures cover
  missing-target, unconfirmed, confirmed, candidate-start, and ambiguous
  states, while real power-loss and clean-device proof remain external gates.
- a shared 32 MiB request-body reader at every JSON handler seam, so direct
  handler or embedding calls retain the same memory bound as normal HTTP
  middleware.
- a clean-source release gate on every CLI/native packager and a read-only
  signed-asset verifier that reconciles the detached manifest with every local
  release file before upload or after public download.
- bounded browser history retention: 200 runtime messages, clipped persisted
  content/reasoning, a 4-million-character serialized budget, and a safe guard
  for oversized legacy localStorage payloads.
- structured, credential-safe upstream failures across web-RPC, direct
  Developer API, and native Google-shaped streams; native Google SSE now emits
  a top-level error rather than assistant-authored Markdown, while OpenAI-style
  streams retain their `[DONE]` terminal sentinel.
- bounded artifact rendering: source is capped at 2 million characters before syntax
  highlighting or iframe construction, render-scoped IDs prevent repeated
  streaming renders from accumulating duplicate registry entries, and the
  in-memory registry is capped at 128 entries/8 million characters with interactive execution
  disabled when capacity cannot be established.

The focused server, Gemini, multimodal, updater, and desktop tests passed
after these changes, as did inline JavaScript syntax validation. The rebuilt
binary also passed local endpoint and hostile-origin checks, and the in-app
browser rendered the branded studio with zero console errors. The empty-state
browser smoke did not exercise provider generation, CDN-backed artifact
execution, long-history performance, or a clean-device updater; those remain
separate gates. The register records 100 failure paths with explicit
`PROTECTED`, `PARTIAL`, `OPEN`, and `EXTERNAL` statuses and remains the honest
boundary for future release decisions.

## Current success criterion

BOB is safer to evolve and easier for students to understand: the default
web-session route remains intact, the optional provider route is explicit and
quota-accountable, and local behavior is protected by deterministic tests.
Student-facing release readiness still requires an authorized live provider
sample, clean-device acceptance, and the separate signed desktop release gates.
