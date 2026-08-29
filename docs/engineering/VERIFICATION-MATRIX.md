# BOB Gemini Free — Mission 0 Verification Matrix

**Audit date:** 2026-08-21 (Asia/Kolkata)  
**Audit scope:** Mission 0 baseline for the supplied workspace snapshot at
`/Users/apple31/Documents/BOB-Gemini-Free`, the two supplied attachments, and
read-only local execution.
**Production-source changes during this mission:** None.

This file preserves the Mission 0 evidence boundary. The repository has since
advanced through Phase II and Phase III work; the current Git-backed state and
new desktop/release evidence are summarized in the Phase III addendum below.

## Purpose and evidence boundary

This document is the Phase II Mission 0 gate. It separates what the current
source implements from what tests exercise, what a live provider has proven,
what is only measured, and what remains a claim or inference.

The supplied first attachment is a historical architecture dossier produced by
another agent. The supplied second attachment is the analysis brief that led to
that dossier. They are useful evidence, but neither is treated as stronger
than the current source tree or a reproducible command result.

The current workspace is a source snapshot, not a Git worktree: `.git` is
absent and `git rev-parse --show-toplevel` fails. Consequently, commit history,
branch identity, remote provenance, and the relationship between the attached
dossier and this exact tree cannot be verified here.

The local endpoint claimed by the workspace instructions was also checked with
a read-only request. Nothing was listening on `127.0.0.1:9610` at audit time.
This matrix therefore contains no current served-runtime or live-Gemini proof.

## Classification meanings

- **VERIFIED_IN_SOURCE** — the behavior is directly visible in the current
  implementation, but may not be exercised end to end.
- **VERIFIED_BY_UNIT_TEST** — a deterministic test in the current tree asserts
  the behavior.
- **VERIFIED_BY_INTEGRATION_TEST** — a local component boundary or
  `httptest`-style integration test asserts the behavior.
- **VERIFIED_LIVE** — a controlled live provider or served-runtime check proved
  it. No row in this audit receives this classification.
- **MEASURED** — a reproducible measurement exists with environment and
  methodology. No published performance number receives this classification.
- **INFERRED** — a conclusion follows from source behavior but is not directly
  asserted by a test or live proof.
- **DOCUMENTED_ONLY** — the claim appears in documentation or the supplied
  dossier without sufficient current source/test proof.
- **UNKNOWN** — the current evidence cannot establish the claim.
- **STALE_OR_INCORRECT** — the current source contradicts the claim or the
  wording materially overstates what the implementation proves.

Where a row also names a protocol support level, that level uses the Phase II
compatibility vocabulary: **NATIVE**, **EMULATED_RELIABLY**,
**EMULATED_PARTIALLY**, **UNSUPPORTED**, or **UNKNOWN**.

## Reproducible audit commands

The following commands were run against the current snapshot on Go
`go1.26.5 darwin/arm64` with `CGO_ENABLED=1`:

| Command | Result | Evidence strength |
|---|---|---|
| `go test -count=1 ./...` | Passed; all Go packages reported `ok` or `[no test files]` | VERIFIED_BY_UNIT_TEST, with caveats below |
| `go test -count=1 -race ./...` | Passed | VERIFIED_BY_UNIT_TEST for exercised paths |
| `go test -count=1 -cover ./...` | Passed; package coverage ranged from 0.0% to 95.0% | MEASURED for these package-local runs |
| `go vet ./...` | Passed with no output | VERIFIED_IN_SOURCE plus tool check |
| `go build ./...` | Passed | VERIFIED_IN_SOURCE plus tool check |
| `go version` | `go1.26.5 darwin/arm64` | MEASURED environment fact |
| `curl --max-time 3 http://127.0.0.1:9610/` | Connection refused | VERIFIED_LIVE negative result: no local instance observed |
| `git rev-parse --show-toplevel` | `fatal: not a git repository` | VERIFIED_IN_SOURCE/workspace fact |

The coverage run is not a project-wide aggregate quality score. At this
initial Mission 0 snapshot, the package-local results were: root 0.6%,
`cmd/desktop` 55.7%,
`internal/browser` 6.9%, `internal/config` 69.6%, `internal/diag` 82.4%,
`internal/format` 74.2%, `internal/gemini` 53.2%, `internal/metrics` 82.4%,
`internal/models` 95.0%, `internal/multimodal` 62.4%, `internal/server`
36.6%, `internal/service` 0.0%, `internal/updater` 28.4%, and
`pkg/gateway` 88.9%.

## Matrix

### Repository, build, and runtime shape

| Claim | Classification | Current evidence | Truth boundary / remaining proof |
|---|---|---|---|
| This is a Go module named `github.com/div197/bob-gemini-free` | VERIFIED_IN_SOURCE | `go.mod:1` | Current source identity is clear; Git provenance is not. |
| The module requires Go 1.26.5 | VERIFIED_IN_SOURCE | `go.mod:3`; host reports Go 1.26.5 | This is the current module requirement, not proof that older Go versions work. |
| “Go 1.22+” is a supported toolchain range | UNKNOWN | The current module declares `go 1.26.5`; the project instruction now records the current snapshot requirement rather than a verified older range | No Go 1.22 build was run, so older-toolchain compatibility remains unverified. |
| The CLI can be built as a CGO-disabled binary | VERIFIED_IN_SOURCE | `Makefile:21-24`, `Dockerfile:7-9` | `go build ./...` passed, but the exact release target and static-link properties were not separately inspected. |
| Six cross-platform release targets are supported | DOCUMENTED_ONLY | `Makefile:36-46` declares six targets | Cross-compilation was not run in this audit; platform-specific runtime behavior is unverified. |
| “Zero dependencies” means no Go dependencies | STALE_OR_INCORRECT | `go.mod:5-50` declares direct and transitive dependencies including `fhttp`, `tls-client`, `websocket`, Wails, and `x/image` | A more accurate claim is “single binary with no separately managed runtime service”; desktop and browser assets still have dependencies. |
| The main runtime is a local HTTP gateway with protocol adapters, Gemini transport, session state, and multimodal upload | VERIFIED_IN_SOURCE | `internal/server/server.go:16-99`, `internal/gemini/client.go:31-99`, `internal/multimodal/upload.go:16-126`, `pkg/gateway/gateway.go:142-201` | The architectural boundary is supported by source. It is not a live deployment proof. |
| The default binding is loopback `127.0.0.1:9610` | VERIFIED_BY_UNIT_TEST | `internal/config/config.go:30-47`; `internal/server/server_test.go:148-153`; `main.go:493-495` | CLI flags/env/config can intentionally change the host. |
| The main CLI performs graceful HTTP shutdown | VERIFIED_IN_SOURCE | `main.go:500-507`, `main.go:541-549` | This does not cover Wails desktop lifecycle or a failed bind. |
| A running gateway is currently live on port 9610 | UNKNOWN | Read-only curl received connection refused | Requires an explicit process/runtime start and served-runtime verification. |

### HTTP surface and protocol compatibility

| Claim | Classification | Current evidence | Current support assessment / remaining proof |
|---|---|---|---|
| The documented route surface exists in the Go mux | VERIFIED_IN_SOURCE | `internal/server/server.go:78-99` mounts health, playground/UI, OpenAI, Anthropic, Google, token, and update routes | Route registration is proven; complete wire compatibility is not. |
| OpenAI Chat Completions request normalization exists | VERIFIED_BY_INTEGRATION_TEST | `internal/server/chat.go:15-67`, `internal/format/openai.go:45-176`; `internal/format/openai_test.go:27-66,137-211`; `internal/server/server_test.go:231-291` | The success path is source-backed, while the full handler success contract is not comprehensively fixture-tested. |
| OpenAI tool-choice inputs are validated before prompt flattening | VERIFIED_BY_UNIT_TEST | `internal/format/openai.go:47-104,191-196`; `internal/format/openai_test.go` covers normalized modes, malformed objects, and undeclared named tools | This prevents false-success fallback for malformed request choices; enforcement remains prompt-based and emulated. |
| No-tools choice is normalized consistently across server adapters | VERIFIED_BY_UNIT_TEST | `internal/format.IsToolChoiceNone`; chat, Responses, and Anthropic handlers use it before deciding whether to parse tool output; `internal/format/openai_test.go` covers case/whitespace normalization and tool-definition omission | This prevents a normalized `none` value from being treated as automatic selection locally; provider-side obedience is not guaranteed. |
| Google-shaped function-calling modes fail closed | VERIFIED_BY_UNIT_TEST | `internal/format/google.go:15-65`; `internal/server/google.go` uses the canonical mode before routing; `internal/format/google_test.go` covers case/whitespace normalization, unknown modes, undeclared names, and incompatible allowed-name modes | This protects the emulated Google adapter; official API behavior and provider-side enforcement remain unverified. |
| Google-shaped parts and roles are not silently dropped | VERIFIED_BY_UNIT_TEST | `internal/format/google.go` validates one supported value per part, supported roles, tool declaration uniqueness, function-call/response names, and bounded JSON; `internal/format/google_test.go` covers empty, ambiguous, malformed, duplicate, unknown-role, and non-text system parts | Valid provider semantics and live Google acceptance remain unverified. |
| Direct Developer API typed text parts are not silently dropped | VERIFIED_BY_UNIT_TEST | `internal/geminiapi/translate.go` rejects empty typed text instead of omitting it; `internal/geminiapi/geminiapi_test.go` covers missing, wrong-type, and empty text fields | The direct route intentionally supports only its documented image/data URL subset. |
| OpenAI message content and roles are not silently discarded | VERIFIED_BY_UNIT_TEST | `internal/format/openai.go` validates string/array/null content, typed text parts, image parts, and supported roles; `internal/format/openai_test.go` covers scalar content, missing/wrong text, unknown roles, and normalized valid roles | This protects the local adapter boundary; official client-version and provider semantic coverage remain separate. |
| Assistant tool-call prompt JSON remains valid for hostile bounded names | VERIFIED_BY_UNIT_TEST | `internal/format/openai.go` JSON-encodes assistant tool-call payloads and named-choice hints; `internal/format/openai_test.go` decodes a quote/backslash/newline fixture from the emitted fence | This prevents serializer corruption; it does not make Markdown extraction native or eliminate prompt-level tool spoofing. |
| OpenAI Chat Completions streaming is “FULL” | STALE_OR_INCORRECT | `internal/server/chat.go:74-208` streams only when no tools or `tool_choice=none`; otherwise it buffers through the non-stream path at `chat.go:211-316` | Ordinary no-tool streaming is implemented; full OpenAI behavior, tool streaming, error semantics, and upstream compatibility remain partial/unknown. |
| OpenAI Responses API is “FULL” | STALE_OR_INCORRECT | `internal/format/responses.go:10-342` rejects unsupported/malformed selected input items; `internal/server/responses.go:55-68,356-369` preserves tool-call/result fields; focused tests cover valid continuations and malformed items | The source emulates a selected Responses shape. Tool streaming is explicitly buffered/replayed at `responses.go:248-326`; broad Responses API parity is not proven. |
| OpenAI system/developer roles are natively preserved | STALE_OR_INCORRECT | `internal/format/openai.go:139-142` turns them into `[System instruction]: ...`; unit assertions are in `internal/format/openai_test.go:48-66` | Support level: EMULATED_RELIABLY for the tested prompt transformation, not native Gemini/OpenAI role semantics. |
| OpenAI structured outputs are enforced | STALE_OR_INCORRECT | `internal/format/openai.go:172-174` appends a textual JSON instruction | Support level: EMULATED_PARTIALLY. No schema validator or response conformance check is present. |
| OpenAI reasoning output is supported as `reasoning_content` | VERIFIED_BY_UNIT_TEST | `internal/format/thinking.go:32-200`, `internal/server/chat.go:79-144`, `internal/format/openai_test.go:213-235`, `internal/format/thinking_test.go:7-152` | The local splitter and serialization path are tested. Whether Google emits the expected fenced markers live is UNKNOWN. |
| Anthropic Messages has a complete compatible SSE lifecycle | STALE_OR_INCORRECT | `internal/server/anthropic.go:116-280` emits the lifecycle; `internal/format/anthropic.go:12-378` now rejects unsupported/malformed input blocks and preserves tool-result order; tests cover selected conversion cases, not the complete SSE contract | Selected lifecycle and strict input conversion are implemented; full Anthropic/Claude Code compatibility is not established. |
| Anthropic extended thinking is full native support | STALE_OR_INCORRECT | `internal/server/anthropic.go:55-67` maps budgets to Gemini think integers; `internal/format/anthropic.go:147-169` maps requests | Support level: EMULATED_PARTIALLY. Budget semantics are reduced to a small internal mode set. |
| Anthropic prompt caching counters are real | STALE_OR_INCORRECT | `internal/server/anthropic.go:132-136,330-335` returns both cache counters as `0` | Support level: EMULATED. No cache accounting exists. |
| Anthropic tool use/results are native | STALE_OR_INCORRECT | `internal/format/anthropic.go:92-120,172-222` converts blocks to OpenAI-shaped data; `internal/format/openai.go:74-80,184-232` injects/parses Markdown blocks | Support level: EMULATED_PARTIALLY; no native Google function-call request is sent. |
| Anthropic tool-choice modes are preserved without silent fallback | VERIFIED_BY_UNIT_TEST | `internal/format/anthropic.go:136-160,393-461`; `internal/format/anthropic_test.go` covers `any`, named `tool`, undeclared names, invalid flags, and unsupported types | Choice enforcement is still emulated; `disable_parallel_tool_use: true` is rejected because this gateway cannot honor it. |
| Google `generateContent`, `streamGenerateContent`, and `countTokens` are implemented | VERIFIED_IN_SOURCE | `internal/server/google.go:14-57,94-165,168-231`; `internal/server/server_test.go:208-229,380-422` | Endpoint-shaped local behavior is tested. Google API semantic parity and SDK compatibility are UNKNOWN. |
| Google streaming is native streaming compatibility | STALE_OR_INCORRECT | `internal/server/google.go:94-165` emits server-sent `data:` records generated from the gateway callback | The route is implemented; the source does not prove exact official Google GenAI stream framing or all generation fields. |
| Google `systemInstruction` is preserved natively | STALE_OR_INCORRECT | `internal/format/google.go:80-100` prepends system text into a prompt | Support level: EMULATED_RELIABLY for text flattening, not native upstream system-instruction semantics. |
| Google function calling with AUTO/NONE/ANY is native | STALE_OR_INCORRECT | `internal/format/google.go:19-61,68-100,149-208` injects textual instructions and regex-parses `function_call` fences | Support level: EMULATED_PARTIALLY. Choice constraints are prompt instructions, not upstream function declarations. |
| Image generation is a native Imagen/DALL-E-compatible generation API | STALE_OR_INCORRECT | `internal/server/images.go:42-104` sends a text prompt to `GenerateContext`, extracts image URLs from returned text, and optionally fetches them | Support level: EMULATED_PARTIALLY and upstream-dependent. No native image-generation RPC is implemented in this tree. |

### Gemini Web RPC, authentication, and streaming

| Claim | Classification | Current evidence | Truth boundary / remaining proof |
|---|---|---|---|
| Requests use a 102-element sparse Web RPC array | VERIFIED_IN_SOURCE | `internal/gemini/payload.go:34-94` | The array length and populated indices are source facts. |
| `inner[0]`, `[17]`, `[59]`, and `[79]` mean exactly what the dossier says | INFERRED | Assignments are visible at `payload.go:45-70`; comments describe intended roles | The meanings are reverse-engineering hypotheses. No captured upstream fixture or live differential test proves each semantic meaning. |
| `BuildURL` targets Google `StreamGenerate` with account prefix, build label, request ID, and `rt=c` | VERIFIED_IN_SOURCE | `internal/gemini/payload.go:101-110` | URL construction is proven; current Google acceptance is UNKNOWN. |
| Dynamic `at`/build tokens are extracted from Gemini HTML | VERIFIED_IN_SOURCE | `internal/gemini/auth.go:19-22,118-151`; `internal/multimodal/tokens.go:22-116` | Regex implementation is present. No deterministic fixture test covers all token variants. |
| SAPISIDHASH follows the tested upstream signing formula | VERIFIED_BY_UNIT_TEST | `internal/gemini/auth.go:162-181`; `internal/gemini/golden_test.go:148-161` injects timestamp `1700000000` and checks the exact digest | This verifies the local formula only. SHA-1 here is the upstream protocol convention, not a release authenticity mechanism. |
| Stream parsing handles arbitrary packet boundaries | VERIFIED_BY_UNIT_TEST | `internal/gemini/stream.go:25-88`; `internal/gemini/golden_test.go` feeds every byte boundary plus malformed and final unterminated frames | Fixture coverage proves the local parser invariant. Live upstream framing remains unverified. |
| Stream retries are duplicate-free after partial output | VERIFIED_BY_INTEGRATION_TEST | `internal/gemini/client.go:264-328`; `internal/gemini/golden_test.go` forces a first-attempt disconnect and checks cumulative-text deduplication | The fixture proves the retry state machine locally; live reconnect behavior remains upstream-dependent. |
| A truncated stream with a complete unterminated final frame is delivered | VERIFIED_BY_INTEGRATION_TEST | `internal/gemini/client.go:331-365`; `internal/gemini/stream.go:81-88`; `internal/gemini/golden_test.go:188-203` | EOF now flushes the parser buffer. A connection that terminates before a complete parseable frame remains an upstream failure/absence case. |
| Bard/Gemini error codes are detected and mapped | VERIFIED_BY_UNIT_TEST | `internal/gemini/parse.go:77-99`, `internal/gemini/stream.go:25-32`; `internal/gemini/parse_test.go:16-22`, `stream_test.go:31-43` | Current regex coverage is narrow; full malformed/nested/error-stream behavior is not proven. |
| Cookie parsing and 0600 persistence are secure | VERIFIED_BY_UNIT_TEST | `internal/gemini/auth.go:186-260`; `internal/gemini/auth_test.go:21-61,95-140` | Parsing and file mode are tested. Secret lifecycle, rotation, and accidental logging need separate review. |
| Cookie pool provides round-robin and 60-second failover | VERIFIED_BY_UNIT_TEST | `internal/gemini/pool.go:217-280`; `internal/gemini/pool_test.go:7-42` | Selection/cooldown behavior is locally tested. It is not proof of Google account rate-limit behavior or transparent end-to-end failover. |
| 1-click login unlocks Pro, vision, and Imagen | DOCUMENTED_ONLY | `internal/browser/browser.go:132-300` implements CDP capture; `main.go:44-86` saves cookies and prints activation text | Browser/provider capability and session validity require an authorized live test. The printed activation message is not proof. |
| Model names identify the actual Google models they claim to represent | UNKNOWN | `internal/models/models.go:20-170` is a static alias/mode table; `models.Resolve` is tested at `models_test.go` | The table proves routing integers and aliases, not upstream model identity or capability. |
| Text/image access is free or unlimited | UNKNOWN | Pricing/limit statements are in `README.md:485-502` and `AGENTS.md:138-147` | Limits, entitlements, and policy are external and time-varying; the gateway cannot establish them. |

### Multimodal and remote-input boundaries

| Claim | Classification | Current evidence | Truth boundary / remaining proof |
|---|---|---|---|
| Remote image fetches accept only HTTP(S), enforce a 20 MB body limit, and reject non-images | VERIFIED_BY_UNIT_TEST | `internal/multimodal/upload.go:129-190`; `internal/multimodal/multimodal_test.go` covers size/type/status and injected public-host fetches | The tested checks are real. Public-host DNS resolution and redirect policy add network-dependent behavior. |
| Remote image fetching has a minimum SSRF boundary | VERIFIED_BY_UNIT_TEST | `internal/multimodal/upload.go`; tests reject loopback/private IPs, private DNS, nonstandard ports, and cross-host redirects; the exported path requires a guardable HTTP client and the transport re-resolves before dialing a literal approved IP | This is application-level evidence; live DNS, egress policy, and operating-system network topology remain separate acceptance gates. |
| Images are compressed to at most 1024 px and the requested byte budget | VERIFIED_BY_UNIT_TEST | `internal/multimodal/compress.go` validates source bounds, adaptively lowers JPEG quality/resolution, and returns an explicit error if the encoded image cannot fit; `multimodal_test.go` asserts the byte limit | JPEG quality is adaptive after the initial quality-75 attempt; provider vision acceptance and authenticated upload remain live-dependent. |
| Scotty upload uses a two-step resumable protocol | VERIFIED_IN_SOURCE | `internal/multimodal/upload.go:41-126` | Upload URL/header parsing and file-reference shape are source-backed; no controlled authenticated live upload was run. |
| Image deduplication prevents duplicate uploads | VERIFIED_BY_UNIT_TEST | `internal/server/image_cache.go`, `internal/server/helpers.go`, and `image_cache_test.go` cover bounded LRU reuse, cookie-scope separation, concurrent single-flight, and waiter cancellation | Provider reference expiry remains unknown; configured cookie pools intentionally disable reuse. |
| Google inline image data is validated before upload | VERIFIED_BY_UNIT_TEST | `internal/format/images.go`, `internal/format/google.go`, and `google_test.go` bound/decode base64 data, validate image MIME types, and reject malformed or oversized input | Provider-specific media acceptance and authenticated vision fidelity still require a live test. |

### Security, trust, and operational boundaries

| Claim | Classification | Current evidence | Truth boundary / remaining proof |
|---|---|---|---|
| Optional API keys use constant-time comparison | VERIFIED_BY_UNIT_TEST | `internal/server/middleware.go`; `internal/server/server_test.go` covers Bearer, `x-api-key`, and `x-goog-api-key`, while query compatibility is separately tested | Query `?key=` authentication is disabled by default and requires explicit `allow_query_api_key`/`BOB_GEMINI_FREE_ALLOW_QUERY_API_KEY=true`; header credentials remain preferred. |
| CORS intentionally supports remote Web Studio and Private Network Access | VERIFIED_BY_INTEGRATION_TEST | `internal/server/middleware.go`; `internal/server/security_boundary_test.go` covers loopback and explicitly configured remote origins | Browser origins are reflected only when trusted; no-origin native clients remain supported. PNA is transport permission, not authentication. |
| An arbitrary web origin can cause privileged no-key gateway requests | VERIFIED_BY_INTEGRATION_TEST | Baseline wildcard behavior was reproduced before the fix; the current security tests reject an untrusted origin and a different loopback port, while allowing only the exact gateway origin or an explicit configured origin | Browser exploitability still depends on browser/PNA behavior; a real browser-origin exploit test and capability-token design remain release gates. |
| Request bodies are capped at 32 MB | VERIFIED_IN_SOURCE | `internal/server/middleware.go:88-94` | This bounds request bodies but does not bound every downstream allocation or remote fetch redirect. |
| `/healthz` is a dedicated unauthenticated health endpoint | VERIFIED_BY_INTEGRATION_TEST | `internal/server/server.go`; `internal/server/handlers.go`; `internal/server/security_boundary_test.go` | It is local computation only and intentionally returns stable status JSON. The richer `/` view remains subject to normal gateway auth. |
| Docker health checks work with secured gateway configuration | VERIFIED_BY_INTEGRATION_TEST | `Dockerfile` and `docker-compose.yml` call the unauthenticated `/healthz`; `internal/server/security_boundary_test.go` proves it remains 200 when API keys protect normal routes | The probe checks only local process health and intentionally does not prove Google/session readiness. |
| No external telemetry is sent by the Go gateway | VERIFIED_IN_SOURCE | `internal/server/server.go:16-27`, `handlers.go:16-54`, and no telemetry client in `go.mod` | This is true for the Go counters as inspected. The Web Studio itself loads CDN assets and calls `inputtools.google.com` (`internal/server/playground.html:20-32,4152`); “no external network” is therefore too broad. |
| Structured internal metrics listed in the dossier exist | VERIFIED_BY_UNIT_TEST | `internal/metrics/metrics.go`, `internal/metrics/metrics_test.go`, and server metrics tests cover bounded aggregate counters/histograms and safe output | Metrics remain process-local; no external telemetry transport or sensitive field is exposed. |

### Updater, desktop, and release claims

| Claim | Classification | Current evidence | Truth boundary / remaining proof |
|---|---|---|---|
| The updater verifies release authenticity | VERIFIED_BY_UNIT_TEST | `internal/updater/updater.go`; `internal/updater/golden_test.go` verifies Ed25519-signed `SHA256SUMS`, tampering, and checksum mismatch | Release operators still need to publish the manifest/signature and configure the trusted public key. Missing or invalid verification material fails closed. |
| Updater replacement is atomic and rollback-safe on all platforms | VERIFIED_BY_UNIT_TEST | Unix uses same-filesystem temp download plus rename; mocked tests isolate the candidate binary and verify bytes/checksum before replacement | Windows retains a platform-specific rollback/copy path; a real installed executable was deliberately not touched. |
| Interrupted native update state is repaired on the next startup | VERIFIED_BY_UNIT_TEST | `internal/updater/desktop_recovery.go` scans only validated updater plans for the exact install target, finalizes healthy candidates, rolls back unconfirmed candidates, and refuses ambiguous states; `desktop_recovery_test.go` covers missing-target, unconfirmed, confirmed, candidate-start, and ambiguous fixtures | Actual power-loss filesystem durability and clean installed-bundle recovery remain external. |
| Updater tests cover download verification and replacement | VERIFIED_BY_INTEGRATION_TEST | `internal/updater/golden_test.go` uses an `httptest` release server and isolated temp candidates | Network transport and GitHub release behavior remain unverified live. |
| Native desktop deterministically starts the gateway and reports the actual port | VERIFIED_BY_INTEGRATION_TEST | `cmd/desktop/gateway.go` and tests cover occupied-port fallback, compatible reuse, and non-BOB rejection; the branded bundle was built and smoke-tested on this Mac, reaching `/playground` with the actual loopback endpoint | Live Google behavior, signed/notarized distribution, and occupied-port GUI fallback still need release-device acceptance. |
| `cmd/desktop` is the sole active native desktop architecture | VERIFIED_IN_SOURCE | `cmd/desktop`, `docs/engineering/DESKTOP-ARCHITECTURE.md`, and the archival commit remove the alternate wrapper from the active tree; Git history retains recovery provenance | Platform publisher trust and clean-machine acceptance remain external release gates. |
| The CLI has a duplicate browser fallback bug | VERIFIED_BY_INTEGRATION_TEST | `main.go` now routes startup through one `launchStudioOrFallback` call; `main_test.go` asserts a single fallback invocation | The fix is intentionally narrow and does not refactor unrelated CLI startup. |

### Performance, diagnostics, and test confidence

| Claim | Classification | Current evidence | Truth boundary / remaining proof |
|---|---|---|---|
| The repository has a benchmark runner reporting P50/P90/P99 and throughput | VERIFIED_IN_SOURCE | `internal/diag/bench.go:15-161`; `main.go:127-157`; `internal/diag/bench_test.go:10-36` | The runner exists and works against a local fake HTTP server. It does not establish provider-independent gateway overhead. |
| The README sample values (P50 ~1.72 s, proxy overhead <1.5 ms, RAM <15 MB, 100+ streams under 50 MB) are measured release baselines | STALE_OR_INCORRECT | The old wording was removed from the active benchmark section; the dossier and historical text had no reproducible artifact | No Google/live release performance number is established. |
| The benchmark supports requested local-only profiles and RSS/allocation metrics | MEASURED | `internal/diag/local_benchmark.go`, `cmd/benchmark-local`, and `docs/engineering/LOCAL-BENCHMARK-2026-08-29.md` record 1/10/20/30 profiles with P50/P90/P95/P99, allocations, RSS, goroutines, connections, and throughput | This is a mocked local gateway benchmark; live upstream profiles remain optional and unrun. |
| The core has a deterministic golden fixture suite | VERIFIED_BY_UNIT_TEST | `docs/engineering/CORE-REGRESSION-HARNESS.md`; `internal/gemini/golden_test.go`; `internal/format/golden_test.go`; `internal/multimodal/golden_test.go` | These are synthetic protocol-shaped fixtures, not live Google captures. Live acceptance remains unknown. |
| The default Go suite is hermetic and provider-independent | VERIFIED_IN_SOURCE | Upstream-facing diagnostic workflows are tagged `live`; the default suite excludes them, while `go test -tags=live ./internal/diag` is documented separately | A live-tagged run remains provider/session-dependent and is intentionally not part of the default validation gate. |
| Current package coverage is at least 80% everywhere | STALE_OR_INCORRECT | `go test -cover ./...` measured 6.9% browser, 34.2% gemini, 46.0% multimodal, 33.4% server, 0.0% service, and 17.6% updater among others | The local package numbers directly contradict a blanket 80% claim. Coverage is not the only quality metric, but the claim is not current. |
| The existing suite proves live Google compatibility | DOCUMENTED_ONLY | The supplied dossier says tests and vet passed; current tests pass locally, but no controlled live acceptance run was performed and no live gateway was listening | Requires separately authorized provider/session verification. |

## Current local verification evidence (2026-08-29)

The later implementation work was validated separately from the historical
Mission 0 snapshot above. The current host is macOS `darwin/arm64` with
Go `go1.26.6`; no live Google credential was used.

| Command/evidence | Result | Classification | Boundary |
|---|---|---|---|
| `go test -count=1 ./...` | Passed for all packages | VERIFIED_BY_UNIT_TEST | Tests prove the exercised local contracts, not live provider behavior. |
| `go test -race -count=1 ./...` | Passed for all packages | VERIFIED_BY_UNIT_TEST | Race coverage is limited to code paths exercised by the suite. |
| `go vet ./...`, `go build ./...`, `go mod verify` | Passed; all modules verified | VERIFIED_IN_SOURCE | This is compile/static/module-integrity evidence, not package acceptance. |
| `make build`, `make desktop-key-check`, `make web` | Passed; public updater key embedded and generated bundle synchronized | VERIFIED_BY_INTEGRATION_TEST | Wails native packaging and platform trust remain separate gates. |
| `go test -count=1 -cover ./...` | Passed; package coverage ranged from 0.0% to 93.1% | MEASURED | Current values include root 0.6%, browser 6.9%, service 0.0%, server 59.1%, updater 54.4%, Gemini 73.8%, and `pkg/gateway` 88.9%; no blanket 80% claim is valid. |
| Isolated binary smoke on `127.0.0.1:19613` | `/healthz`, `/playground`, favicon, manifest, and service worker returned 200; hostile-origin requests returned 403; exact-origin API access reached authentication; process stopped cleanly | VERIFIED_BY_INTEGRATION_TEST | This proves local HTTP serving and the tested origin gate only; it is not a provider or 30-device rollout test. |
| Inline JavaScript syntax parse | Two inline scripts parsed with Node without syntax errors | VERIFIED_IN_SOURCE | DOM behavior, CDN availability, and interactive artifact execution still require a real browser/device. |
| Release source provenance | `scripts/verify-release-source.sh` requires a Git worktree with no modified/untracked paths, aligns a stable or `-preview.N` version with the `Makefile` base version, and compares the generated Web bundle without rewriting it | VERIFIED_IN_SOURCE | This protects local packaging entry points; commit review, tag publication, and final public-asset reconciliation remain operator gates. |
| Release directory and signed manifest reconcile one-to-one | `scripts/verify-release-assets.sh` and `updater.VerifySignedReleaseDirectory` verify the detached Ed25519 signature, exact SHA-256 bytes, and no extra/missing/duplicate/symlinked local asset | VERIFIED_BY_UNIT_TEST | This is a local pre-upload/post-download gate; GitHub upload contents and release metadata still require operator verification. |
| Browser chat history has bounded runtime and persisted state | `internal/server/playground.html` caps messages, content/reasoning, attachments, serialized storage, and oversized legacy payloads; `playground_test.go` locks the guard markers | VERIFIED_IN_SOURCE | A clean-browser quota/performance run and long-session device acceptance remain separate gates. |
| In-app browser smoke against rebuilt binary | `http://127.0.0.1:19613/playground` rendered the branded studio, exposed one scroll-to-top control in the empty state, reported expected non-scrollable geometry, and produced zero console errors | VERIFIED_BY_INTEGRATION_TEST | This does not exercise a provider generation, CDN-backed artifact, long history, or clean-device update. |

## Mission 0 decision

Mission 0 is complete as an audit artifact. The matrix establishes a safe gate
for implementation:

1. The source/build baseline is currently green on this host.
2. The fragile protocol core is only partially regression-locked.
3. The strongest compatibility claims must be rewritten as selected,
   source-backed emulation until fixtures and live checks prove more.
4. The CORS/API-key/health boundary, updater authenticity gap, desktop port
   handling, duplicate browser fallback, and missing benchmark evidence are
   real Phase II work items.
5. No production source file was modified during Mission 0.

### Protected-file gate for Mission 1

The protected files exist and were not changed:

```text
internal/gemini/payload.go
internal/gemini/stream.go
internal/gemini/auth.go
```

Mission 1 first added deterministic tests for exact payload positions,
injectable SAPISIDHASH, arbitrary stream boundaries, malformed lines/nested
JSON, EOF handling, Bard errors, and retry deduplication. Only then were the
two justified protected-core changes described in
`docs/engineering/CORE-REGRESSION-HARNESS.md` applied.

## Explicitly unresolved after Mission 0

- Current Git commit/branch/remote provenance.
- Actual live process/deployment state.
- Google Web RPC acceptance and undocumented field semantics.
- Real model identity, context windows, entitlements, rate limits, and
  “free/unlimited” behavior.
- Browser-level malicious-origin/PNA exploitability and the minimum viable
  pairing/token design.
- Authenticated Scotty upload and Imagen generation.
- Release authenticity and safe updater replacement under mocked downloads.
- Final Wails-only desktop product decision (completed after the Phase III
  device comparison).
- Environment-tagged performance baselines.

## Phase III evidence update (2026-08-21)

The historical Mission 0 statements above are not the current branch status.
The following later evidence is now available:

| Current claim | Classification | Evidence | Boundary |
|---|---|---|---|
| The Wails-only archival change is published on `main` | VERIFIED_IN_SOURCE | GitHub PR #2 merged `phase-iii/release-desktop-docs-hardening` into `main` as merge commit `5ccbebe`; repository metadata and the merge were inspected | This proves source publication, not a signed binary release or end-user acceptance. |
| Packaged desktop users need no separate Go runtime, database, or memory service | VERIFIED_BY_INTEGRATION_TEST | `scripts/build-wails-local.sh` produced an ad-hoc signed macOS app; Computer Use opened the branded app, reached the BOB Builder studio at `127.0.0.1:9610/playground`, and closing the window released the listener | Source builds still require Go, CGO, the desktop build toolchain, and the host toolchain; macOS notarization/signing is an external release gate. |
| The active desktop/browser icon set is coherent | VERIFIED_IN_SOURCE | `scripts/build-desktop-icons.sh` rebuilds Wails PNG, browser favicon, and server favicon from `assets/bob-gemini-free-logo.jpg`; non-empty outputs and the native studio were inspected | Visual brand approval remains a product-owner judgment; no alternate desktop icon set is shipped. |
| The gateway has no SQLite or server memory database | VERIFIED_IN_SOURCE | No SQLite dependency exists in the Go module or server; the browser studio's SQLite WASM surface and claims were removed, leaving only client UI preference/history storage | Browser `localStorage` remains client-only UI state; it is not a gateway database or memory service. |
| A signed `v0.1.7` binary release is published | UNKNOWN | GitHub release inspection showed no `v0.1.7` release; the local release-key variables were unset and unsigned artifacts were not uploaded | Requires the operator Ed25519 key pair, local signed matrix build, manifest verification, manual asset upload, and clean-machine updater verification. |
| The current module/toolchain floor is Go 1.26.6 | VERIFIED_IN_SOURCE | `go.mod:3` declares `go 1.26.6`; the Go 1.26.6 toolchain was downloaded and used for the current validation run | Older Go versions remain unsupported/unverified. |
| Reachable known Go vulnerabilities are clear after the update | MEASURED | `GOTOOLCHAIN=go1.26.6 go run golang.org/x/vuln/cmd/govulncheck@latest ./...` reported zero reachable vulnerabilities after upgrading `golang.org/x/image` to v0.45.0 | The scan is time-sensitive; rerun it before each release. It does not replace review of application-specific flaws. |
| Public-repository secret scanning and push protection are enabled | VERIFIED_LIVE | GitHub repository settings were enabled and the current secret-alert query returned an empty list | GitHub settings are external state and must remain part of maintainer release review. |
| `main` requires pull requests and blocks force-push/deletion | VERIFIED_LIVE | GitHub branch protection reports enforced admins, pull-request review gate, conversation resolution, and disabled force pushes/deletions | No hosted status check is required because this project intentionally does not use GitHub Actions; local validation remains mandatory. |
| The native desktop process honors a user's local config/cookie without widening its network boundary | VERIFIED_BY_UNIT_TEST | `cmd/desktop/config_test.go` verifies cookie discovery while forcing loopback, empty API keys, and empty remote origins | First-run login UX is not implemented; authenticated features remain per-user/session-dependent. |
| The public release is a student-ready cross-platform native installer set | UNKNOWN | GitHub release inspection found CLI assets in `v0.1.5` but no trusted native installer set; only macOS ARM64 ad-hoc native smoke testing is complete | Requires native Windows/Linux builds, clean-device tests, platform signing, macOS notarization, and manual release publication. |
| A public native desktop preview is available for controlled evaluation | VERIFIED_LIVE | Manually published prerelease `v0.1.7-preview.7` contains the corrected branded macOS universal `.dmg`/`.zip`, `RELEASE-NOTICE.txt`, `SHA256SUMS`, and detached `SHA256SUMS.sig` assets; the downloaded assets were byte-compared with the local signed candidate and re-verified | This is a beta preview only: no Apple notarization, Windows publisher signature, Linux asset, clean-device matrix, silent updater, or production student-readiness claim. Existing Preview 6 installations require one manual migration because the original project signing key was not recoverable. Its released updater predates the current stable-first source change, so an existing Preview 7 fleet needs a same-key bridge preview or manual stable installation. |
| A free branded macOS preview package can be created without Apple membership | VERIFIED_BY_INTEGRATION_TEST | `scripts/package-wails-preview.sh` creates a branded `BOB Gemini Free.app`, `.zip`, `.dmg`, release notice, and checksums without Developer ID credentials; the bundle metadata uses the `com.abcsteps` identity | This proves local packaging only; it does not establish Gatekeeper trust, notarization, clean-device acceptance, or public student readiness. |
| The native desktop app exposes an explicit update check | VERIFIED_BY_UNIT_TEST | `internal/updater/desktop.go` selects only official stable or `preview.N` channels, with a bounded preview listing; `internal/updater/updater_test.go` covers branded/legacy names, prerelease ordering, signed-manifest discovery, stable-first migration for newly built previews, stable-failure fail-closed behavior, and the endpoint bound | The current source can offer a consented verified update; the already-published Preview 7 binary must first receive a same-key bridge preview to reach stable through the updater. It does not silently install, remove the macOS warning, or replace platform publisher trust. |

## Current v0.2.0 release-readiness update (2026-08-29)

The current working branch was merged into `main` at source commit `e019cf8`.
The signed `v0.2.0-preview.1` migration bridge is published; stable `v0.2.0`
has not been tagged or published. The separate
[`RELEASE-READINESS-v0.2.0.md`](RELEASE-READINESS-v0.2.0.md) is the authoritative
publication gate for this milestone.

| Current claim | Classification | Evidence | Boundary |
|---|---|---|---|
| The stable build path embeds the repository's updater public key | VERIFIED_BY_INTEGRATION_TEST | `make build`, `make dist`, `make desktop-key-check`, and binary string inspection passed for the six CLI targets and the macOS Wails candidate | This proves the embedded trust anchor, not a signed release manifest or Apple/Windows publisher trust. |
| The macOS v0.2.0 candidate is package-valid | VERIFIED_BY_INTEGRATION_TEST | Local Wails universal app, ZIP, DMG, DMG Applications shortcut, ad-hoc code signature, SHA-256 checks, bundle metadata, and `/healthz` smoke test passed; the `v0.2.0-preview.1` bridge was then signed and publicly re-downloaded with checksum/signature verification | `spctl` rejection is expected without Apple notarization; this proves the controlled preview bridge, not a stable student release. |
| Update metadata URLs are pinned to the official repository or GitHub release CDN | VERIFIED_BY_UNIT_TEST | `internal/updater/desktop.go` and `internal/updater/updater_test.go` reject other GitHub owners/repositories, non-HTTPS URLs, lookalike hosts, and unexpected ports | GitHub release metadata is still external state; the signed manifest remains the artifact authenticity boundary. |
| Existing public Preview 7 installations can update directly to v0.2.0 stable | STALE_OR_INCORRECT | The published Preview 7 binary predates stable-first discovery; the published same-key `v0.2.0-preview.1` bridge is now available, while current source tests prove stable-first only for newly built packages | Install the bridge first, then bridge → stable after stable acceptance, or perform one manual stable installation. Preview 6 and older need manual current-key migration. |
| The source is ready for a student-facing signed stable publication | UNKNOWN | Local source, package, and mocked updater gates are green, but no exact signed/uploaded v0.2.0 assets, clean-device acceptance, platform publisher trust, or 30-device migration run exists | Do not publish or announce stable until the owner completes the release sequence and Gates A–C in `RELEASE-READINESS-v0.2.0.md`. |

## Current Gemini Developer API routing update (2026-08-29)

This section records the explicit student-owned-key route added after the
earlier web-RPC audit. It does not change the default cookie/guest path and it
does not turn provider documentation into a product guarantee.

| Current claim | Classification | Evidence | Boundary |
|---|---|---|---|
| A student can opt into a separate Gemini Developer API route from the Web Studio | VERIFIED_IN_SOURCE | `internal/server/playground.html` links to `https://aistudio.google.com/app/apikey`, keeps the key in page memory, and sends it only as `X-BOB-Gemini-API-Key` when the session toggle is enabled | This is a local UI/source claim; the student still owns the Google project/key and must follow Google's account, billing, and data-use rules. |
| The provider key is translated to Google's documented header and never placed in the URL or request body | VERIFIED_BY_UNIT_TEST | `internal/geminiapi/client.go`, `internal/geminiapi/geminiapi_test.go`, and `internal/server/gemini_api_test.go` assert `x-goog-api-key`, no query key, no body key, and redacted provider errors | Runtime proxies or provider-side logging remain outside BOB's control. |
| Direct Developer API typed streams reject empty semantic events | VERIFIED_BY_UNIT_TEST | `internal/geminiapi/client.go` rejects `{}` and `candidates:[]` events unless prompt feedback or usage metadata is present; `internal/geminiapi/geminiapi_test.go` covers both forms alongside valid comments/data/`[DONE]` framing | Full provider event vocabulary and Web RPC framing remain unverified. |
| Direct Developer API tool-choice names are checked against declared tools | VERIFIED_BY_UNIT_TEST | `internal/geminiapi/translate.go:23-25`; shared `internal/format.ValidateToolChoice`; `internal/geminiapi/geminiapi_test.go` covers undeclared named choices | Gemini `AUTO`/`NONE`/`ANY` mapping is source-backed; provider acceptance and exact model semantics remain live-dependent. |
| Direct Developer API typed text parts cannot disappear on conversion | VERIFIED_BY_UNIT_TEST | `internal/geminiapi/translate.go:196-218`; `internal/geminiapi/geminiapi_test.go` covers missing `type` and non-string `text` | This is a request-validation guarantee, not evidence that every provider content part is supported. |
| Explicit Developer API chat uses native public REST/SSE rather than the web-RPC cookie path | VERIFIED_BY_UNIT_TEST | `internal/server/gemini_api.go` and `internal/server/gemini_api_test.go` prove the explicit route calls the isolated client and the web client is not called | Live Google acceptance, exact model availability, quota, and semantic parity remain unverified. |
| Direct provider routing supports future Gemini model IDs without inventing BOB aliases | VERIFIED_IN_SOURCE | `directGeminiModel` maps only local convenience aliases and forwards provider-shaped Gemini IDs; Google remains the authority for whether a model exists | The Web Studio dropdown and BOB web-RPC model catalog are separate; a provider model must still be selected through a supported client surface and accepted by Google. |
| The UI or gateway promises a fixed free tier, universal RPM/RPD, or unlimited access | STALE_OR_INCORRECT | Current README, Hindi README, routing guide, and glossary link to Google's pricing/rate-limit/billing pages and explicitly reject fixed quota promises | Google can change limits by model, project, tier, region, account, and date; re-check official pages before every release. |
| The student's Developer API key is silently rotated or used as a fallback after a web-route failure | VERIFIED_BY_UNIT_TEST | Single-key resolver and explicit route tests reject repeated keys and preserve provider selection; UI/docs state no rotation and no cross-route replay | A caller can make a new explicit request after an error; BOB does not hide that provider change. |
| A Developer API key is supported on every BOB adapter and image path | STALE_OR_INCORRECT | Explicit keys are currently rejected on `/v1/messages`, `/v1/responses`, and `/v1/images/generations`; only chat and native Google generation/counting are wired | Additional adapters need their own translation fixtures and release evidence before being enabled. |
| Current free-tier and model facts are maintained as a dated release gate rather than hardcoded forever | VERIFIED_IN_SOURCE | `docs/engineering/GEMINI-API-ROUTING.md` defines the official-source review and the Unreleased changelog records the policy | This is a process control, not live quota telemetry; students must inspect the current AI Studio project view. |

## Current 100-path failure-register update (2026-08-29)

The implementation continuation after the earlier matrix audit is tracked in
[`FAILURE-REGISTER-100.md`](FAILURE-REGISTER-100.md). The register deliberately
keeps open and external risks visible; the following claims are the new local
evidence only.

| Current claim | Classification | Evidence | Boundary |
|---|---|---|---|
| The served local studio exposes its PWA manifest and service worker without a separate static server | VERIFIED_BY_INTEGRATION_TEST | `internal/server/playground.go`, `internal/server/server.go`, embedded `manifest.json`/`sw.js`, `internal/server/server_test.go`, and a local HTTP probe return 200 with the expected content types and injected version marker | A browser still needs to activate the worker; CDN-dependent libraries remain a separate offline/runtime concern. |
| The service worker caches only local static resources and never caches generation/API responses | VERIFIED_IN_SOURCE | `internal/server/sw.js` restricts cache operations to same-origin GET requests and excludes `/v1/` and `/v1beta/` | Browser implementation and OS WebView cache behavior still need clean-device acceptance. |
| Stream subscribers silently lose chunks when a follower is slow | VERIFIED_BY_UNIT_TEST | `internal/gemini/flight.go` bounds subscriber/history buffers and returns `ErrStreamSubscriberTooSlow`; focused and race tests cover slow followers, cancellation, and history limits | A direct caller of the low-level flight API must still use the bounded API correctly; live upstream behavior remains unverified. |
| Transient retries can overflow timing values or synchronize classroom bursts | VERIFIED_BY_UNIT_TEST | `internal/config/config.go` clamps attempts/delay/timeout; `internal/gemini/client.go` uses bounded exponential jitter and bounded `Retry-After` parsing while refusing 429/policy retries; config/client tests cover the limits | Google capacity and school-network rate limits remain external; no proxy rotation or quota-evasion behavior is introduced. |
| Browser-session responses are coalesced across different configured sessions | VERIFIED_BY_UNIT_TEST | `internal/gemini/client.go` disables request coalescing for configured/loaded session clients and tests anonymous-only coalescing | Future provider/session scopes must be added to the flight key before re-enabling coalescing. |
| Oversized or malformed upstream bodies can be read without a bound | VERIFIED_BY_UNIT_TEST | `internal/gemini/client.go` bounds normal bodies and stream lines and tests oversized response/line and nil-response cases | Limits are application-level; the provider and network can still return errors or slow responses. |
| Unsafe Scotty upload URLs, upload statuses, response bodies, or image dimensions are accepted | VERIFIED_BY_UNIT_TEST | `internal/multimodal/upload.go`, `compress.go`, and golden tests enforce host/scheme/path, status, body, byte, dimension, pixel, and base64 limits; remote-image DNS is revalidated at the final guarded dial | OCR CPU/memory pressure and provider session capability remain open/external rows; live egress topology is not proven. |
| A cookie file change with unchanged mtime/size remains invisible, or a broad-permission/symlink file is accepted | VERIFIED_BY_UNIT_TEST | `internal/gemini/auth.go`, `auth_test.go`, and pool tests use content hashing plus secure-file validation and cover deletion, permissions, and deduplication | The operating system and user account still control file ownership and keychain/session validity. |
| The native updater follows an arbitrary redirect or accepts an unbounded/mis-sized release body | VERIFIED_BY_UNIT_TEST | `internal/updater/desktop.go`, `updater.go`, `desktop_stage.go`, and updater tests constrain official hosts, redirects, metadata, signatures, artifact size, and exact bytes | A real update from `/Applications`, rollback after interruption, platform publisher trust, and public upload reconciliation remain external gates. |
| Native updater replacement races or target swaps can silently replace the wrong install | VERIFIED_BY_UNIT_TEST | `internal/updater/desktop_helper.go` uses an atomic per-install lock with conservative stale-lock recovery and revalidates target/candidate after locking; updater tests cover conflict, stale non-empty refusal, and replacement paths | Filesystem/OS interruption and clean-device rollback remain external gates. |
| Stale updater staging directories can accumulate without bounded cleanup | VERIFIED_BY_UNIT_TEST | `internal/updater/desktop_stage.go` removes only old committed plans that match the exact target and pass plan validation; focused cleanup tests keep fresh/unrelated directories | Crashes before plan commit remain for manual inspection by design. |
| ZIP extraction can accept ambiguous, duplicate, or special entries | VERIFIED_BY_UNIT_TEST | `extractMacOSApp` rejects traversal components, duplicate normalized paths, symlinks, special files, unexpected roots, and expansion beyond the archive bound; focused tests cover duplicate/special entries | Real signed package contents still require release-artifact review. |
| The artifact editor can remain empty after an interactive preview has a source, chat scroll controls can remain hidden after layout, or repeated streaming renders can grow artifact state without bound | VERIFIED_IN_SOURCE | `internal/server/playground.html` normalizes Marked token objects, recovers an empty editor from the registered source, recalculates directional controls after layout/resize, uses stable render-scoped artifact IDs, caps each source at 2 million characters and the registry at 128 items/8 million characters, and `playground_test.go` locks the markers | The in-app browser environment did not load CDN libraries during this run, so full interactive artifact execution, CDN failure UI, and clean-browser memory acceptance still require a device check. |
| Local history can persist a deliberately truncated image data URL or unbounded attachment text | VERIFIED_IN_SOURCE | Client-side attachment size, image data URL, text, and attachment-count limits are defined in `playground.html`; storage tests reject the old truncation pattern | Browser localStorage quota and large parsing/OCR CPU remain environment-dependent; bounded history retention is still open. |
| Scotty image-reference cache can grow without bound or cross a configured account pool | VERIFIED_BY_UNIT_TEST | `internal/server/image_cache.go` bounds references to 256 with LRU-like eviction and shares in-flight loads; `helpers.go` scopes reuse to a hashed single-cookie session and disables it for configured pools; cache/scope tests cover eviction, cookie rotation, waiter cancellation, and concurrent loads | Provider-side reference expiry remains open and requires live authenticated evidence. |
| Browser transport, protocol, and provider-stream errors are stored as successful chat turns | VERIFIED_IN_SOURCE | `internal/server/playground.html` treats `finish_reason: error`, structured SSE errors, interrupted reads, and incomplete EOF as terminal failures, stores `status: error` or `stopped`, and excludes non-complete assistant turns from later prompts; `internal/server/playground_test.go` locks these lifecycle markers | Full behavior still needs a clean browser/device execution because this run's isolated in-app browser did not load external CDN libraries. |
| Tool schema complexity can bypass limits through typed Go values or cyclic structures | VERIFIED_BY_UNIT_TEST | `internal/format/schema.go` uses reflection-safe structural accounting for maps, slices, arrays, pointers, interfaces, typed enum/property collections, node count, and depth; `internal/format/schema_test.go` covers typed enum and cyclic-map fixtures | JSON-provider semantic support remains partial and streamed tool-call aggregation is a separate open concern. |
| JSON handlers can allocate an unbounded request body when invoked outside the normal middleware chain | VERIFIED_BY_UNIT_TEST | `internal/server/helpers.go` exposes one bounded `readRequestBody` seam used by chat, responses, Anthropic, Google, image, token, refine, and direct handler paths; `internal/server/server_test.go` proves an oversized body is rejected before JSON conversion | The bound protects local memory; request semantics and provider limits remain separate concerns. |
| Any loopback browser origin is trusted by default | VERIFIED_BY_INTEGRATION_TEST | `internal/server/middleware.go` compares the browser origin to the exact request host/scheme and `security_boundary_test.go` rejects a different loopback port while preserving explicit remote-origin configuration | A real browser/PNA test is still required; explicit remote origins remain an operator trust decision. |
| Shared stream or direct Developer API callbacks can panic when an embedding passes nil functions or context | VERIFIED_BY_UNIT_TEST | `internal/gemini/flight.go`, `internal/geminiapi/client.go`, and `internal/server/gemini_api.go` normalize nil contexts and return explicit callback errors; focused tests cover each seam | This protects local API misuse; it does not make a cancelled provider transport resumable. |
| Generation POST retries are limited to failures known to precede provider request delivery | VERIFIED_BY_UNIT_TEST | `internal/gemini/client.go` retries only `net.OpError` dial/connect/lookup failures; tests cover safe pre-request retry and ambiguous transport no-retry behavior | HTTP 5xx, response-read, parser, and partial-stream failures remain non-replayed because the web-RPC POST has no idempotency contract. |
| Direct Developer API stream tool calls are emitted only after valid bounded assembly | VERIFIED_BY_UNIT_TEST | `internal/server/gemini_api.go` keeps ID-ordered cumulative snapshots, validates arguments, rejects changed IDs/names, and emits finalized calls once; stream-error tests prove no assistant Markdown fallback | This protects the direct public REST/SSE route only; the reverse-engineered web-RPC tool path remains emulated. |
| Direct Developer API empty SSE streams fail closed | VERIFIED_BY_UNIT_TEST | `internal/geminiapi/client.go:354-421` rejects an empty body or `[DONE]` without a JSON event; `internal/geminiapi/geminiapi_test.go` covers both fixtures | This prevents a fabricated successful stop; provider stream semantics beyond the tested SSE grammar remain unverified. |
| Tool-result continuations cannot silently bind to an unknown, mismatched, or ambiguous call | VERIFIED_BY_UNIT_TEST | `internal/format/openai.go` validates IDs/names before prompt flattening and direct translation; valid, unknown, mismatch, and ambiguous fixtures are covered | Assistant calls without IDs remain accepted for legacy compatibility; callers should provide stable IDs for parallel tools. |
| Direct Developer API candidate and finish-reason ambiguity is hidden as success | VERIFIED_BY_UNIT_TEST | Direct translation rejects multiple candidates and fails closed on unknown finish reasons while mapping known stop/length/filter/tool outcomes | The raw web-RPC route has separate provider-dependent semantics and is not claimed equivalent. |
| Standalone installer accepts an unsigned or unverified release binary | VERIFIED_IN_SOURCE | `install.sh` and `install.ps1` use HTTPS-only downloads, fixed public-key verification, exact signed SHA-256 entries, size bounds, atomic local installation, and no default source/unsigned fallback | Shell/PowerShell clean-host execution and availability of an Ed25519-capable verifier remain external; macOS stock LibreSSL may fail closed. |
| Release evidence omits the source commit, toolchain, or exact signed asset hashes | VERIFIED_IN_SOURCE | `scripts/record-release-evidence.sh` re-runs source/asset verification and records commit, branch, Go version, host, time, manifest/signature hashes, and asset hashes in a 0600 receipt outside the worktree | The operator must retain and reconcile the receipt with the public GitHub release; it is not a hosted attestation. |
| The whole 100-path register is closed | STALE_OR_INCORRECT | The register contains 100 numbered paths with explicit PROTECTED/PARTIAL/OPEN/EXTERNAL statuses | Do not use this matrix or the register as a production-readiness certificate; close each release/device/provider gate with its own evidence. |
