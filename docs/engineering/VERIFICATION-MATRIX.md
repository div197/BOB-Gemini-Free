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
| Anthropic Messages has a complete compatible SSE lifecycle | STALE_OR_INCORRECT | `internal/server/anthropic.go` emits ordered success/error lifecycle events, stops after SSE write failures, and `internal/server/server_test.go` covers success and upstream-error ordering; `internal/format/anthropic.go` rejects unsupported/malformed input blocks and preserves tool-result order | Selected lifecycle and strict input conversion are implemented; full Anthropic/Claude Code compatibility is not established. |
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
| Dynamic `at`/build tokens are extracted from Gemini HTML | VERIFIED_BY_UNIT_TEST | `internal/gemini/auth.go:19-22,118-151`; `internal/multimodal/tokens.go` bounds page responses/token values, refuses redirects/non-success responses, preserves the previous set after failed refresh, and retries a failed refresh after a bounded delay; `internal/multimodal/multimodal_test.go` covers valid refresh, oversized response, redirect behavior, and delayed recovery | Regex coverage is still narrow and provider token semantics remain upstream-dependent. |
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
| Studio's visible default route cannot be silently replaced by a process-level Developer API key | VERIFIED_BY_UNIT_TEST | `internal/server/gemini_api.go` honors `X-BOB-Gemini-Route: web`; `internal/server/gemini_api_test.go` proves the override and provider-call suppression, while `internal/server/playground_test.go` locks the Studio header | Other clients that omit the selector retain the documented process-level `BOB_GEMINI_FREE_GEMINI_API_KEY` behavior; the selector is not a credential. |
| Native desktop BOB gateway access-key field is normally unnecessary | VERIFIED_BY_UNIT_TEST | `cmd/desktop/config.go` and `cmd/desktop/config_test.go` force loopback and strip `api_keys`; the Config copy and desktop README explain the separate protected-endpoint case | This does not remove the optional page field for a Studio connected to an independently running protected gateway. |
| CORS intentionally supports remote Web Studio and Private Network Access | VERIFIED_BY_INTEGRATION_TEST | `internal/server/middleware.go`; `internal/server/security_boundary_test.go` covers loopback and explicitly configured remote origins | Browser origins are reflected only when trusted; no-origin native clients remain supported. PNA is transport permission, not authentication. |
| An arbitrary web origin can cause privileged no-key gateway requests | VERIFIED_BY_INTEGRATION_TEST | Baseline wildcard behavior was reproduced before the fix; the current security tests reject an untrusted origin, a different loopback port, and a non-loopback Host/origin pair, while allowing only the exact literal-loopback gateway origin or an explicit configured origin. `docs/engineering/BROWSER-SECURITY-VALIDATION-2026-08-31.md` records a real in-app-browser cross-port models and JSON POST/preflight probe, including an explicit PNA preflight header | Public-HTTPS/PNA deployment behavior and the capability-token design remain separate release gates. |
| Request bodies are capped at 32 MB | VERIFIED_IN_SOURCE | `internal/server/middleware.go:88-94` | This bounds request bodies but does not bound every downstream allocation or remote fetch redirect. |
| `/healthz` is a dedicated unauthenticated health endpoint | VERIFIED_BY_INTEGRATION_TEST | `internal/server/server.go`; `internal/server/handlers.go`; `internal/server/security_boundary_test.go` | It is local computation only and intentionally returns stable status JSON. The richer `/` view remains subject to normal gateway auth. |
| Docker health checks work with secured gateway configuration | VERIFIED_BY_INTEGRATION_TEST | `Dockerfile` and `docker-compose.yml` call the unauthenticated `/healthz`; `internal/server/security_boundary_test.go` proves it remains 200 when API keys protect normal routes | The probe checks only local process health and intentionally does not prove Google/session readiness. |
| No external telemetry is sent by the Go gateway | VERIFIED_IN_SOURCE | `internal/server/server.go:16-27`, `handlers.go:16-54`, and no telemetry client in `go.mod` | This is true for the Go counters as inspected. The Web Studio itself loads CDN assets and calls `inputtools.google.com` (`internal/server/playground.html:20-32,4152`); “no external network” is therefore too broad. |
| Structured internal metrics listed in the dossier exist | VERIFIED_BY_UNIT_TEST | `internal/metrics/metrics.go`, `internal/metrics/metrics_test.go`, and server metrics tests cover bounded aggregate counters/histograms, fixed-cardinality route counts, and safe output | Metrics remain process-local; no external telemetry transport or sensitive field is exposed, and per-request capability correlation is intentionally not claimed. |

### Updater, desktop, and release claims

| Claim | Classification | Current evidence | Truth boundary / remaining proof |
|---|---|---|---|
| The updater verifies release authenticity | VERIFIED_BY_UNIT_TEST | `internal/updater/updater.go`; `internal/updater/golden_test.go` verifies Ed25519-signed `SHA256SUMS`, tampering, and checksum mismatch | Release operators still need to publish the manifest/signature and configure the trusted public key. Missing or invalid verification material fails closed. |
| Updater replacement is atomic and rollback-safe on all platforms | VERIFIED_BY_UNIT_TEST | Unix uses same-filesystem candidate/target renames; the desktop helper flushes Unix directory entries after backup, activation, and rollback transitions; Windows updater metadata uses native `MoveFileExW` replace-existing/write-through semantics; mocked tests isolate the candidate and verify bytes/checksum before replacement, while `desktop_durability_test.go` injects activation-sync failure and proves restoration | Windows executable replacement retains its platform-specific rollback/copy path and directory durability remains OS-specific; recursive app-bundle durability and a real installed executable were deliberately not claimed or touched. |
| Interrupted native update state is repaired on the next startup | VERIFIED_BY_UNIT_TEST | `internal/updater/desktop_recovery.go` scans only validated updater plans for the exact install target, finalizes healthy candidates, rolls back unconfirmed candidates, flushes recovery directory transitions, and refuses ambiguous states; `desktop_recovery_test.go` covers missing-target, unconfirmed, confirmed, candidate-start, and ambiguous fixtures | Actual power-loss ordering, recursive candidate-file durability, Windows journal behavior, and clean installed-bundle recovery remain external. |
| Updater tests cover download verification and replacement | VERIFIED_BY_INTEGRATION_TEST | `internal/updater/golden_test.go` uses an `httptest` release server and isolated temp candidates | Network transport and GitHub release behavior remain unverified live. |
| Native desktop deterministically starts the gateway and reports the actual port | VERIFIED_BY_INTEGRATION_TEST | `cmd/desktop/gateway.go` and tests cover occupied-port fallback, exact-version compatible reuse, stale-version rejection, and non-BOB rejection; the desktop server publishes its safe release identity through `X-BOB-Version` on `/healthz` | The packaged coexistence smoke and each supported release host still need acceptance; this proves local selection, not live Google behavior, signed/notarized distribution, or fleet rollout. |
| `cmd/desktop` is the sole active native desktop architecture | VERIFIED_IN_SOURCE | `cmd/desktop`, `docs/engineering/DESKTOP-ARCHITECTURE.md`, and the archival commit remove the alternate wrapper from the active tree; Git history retains recovery provenance | Platform publisher trust and clean-machine acceptance remain external release gates. |
| The CLI has a duplicate browser fallback bug | VERIFIED_BY_INTEGRATION_TEST | `main.go` now routes startup through one `launchStudioOrFallback` call; `main_test.go` asserts a single fallback invocation | The fix is intentionally narrow and does not refactor unrelated CLI startup. |

### Performance, diagnostics, and test confidence

| Claim | Classification | Current evidence | Truth boundary / remaining proof |
|---|---|---|---|
| The repository has a benchmark runner reporting P50/P90/P99 and throughput | VERIFIED_IN_SOURCE | `internal/diag/bench.go:15-161`; `main.go:127-157`; `internal/diag/bench_test.go:10-36` | The runner exists and works against a local fake HTTP server. It does not establish provider-independent gateway overhead. |
| The README sample values (P50 ~1.72 s, proxy overhead <1.5 ms, RAM <15 MB, 100+ streams under 50 MB) are measured release baselines | STALE_OR_INCORRECT | The old wording was removed from the active benchmark section; the dossier and historical text had no reproducible artifact | No Google/live release performance number is established. |
| The benchmark supports requested local-only profiles and RSS/allocation metrics | MEASURED | `internal/diag/local_benchmark.go`, `cmd/benchmark-local`, and `docs/engineering/LOCAL-BENCHMARK-2026-08-31.md` record 1/10/20/30 profiles with P50/P90/P95/P99, allocations, RSS, goroutines, connections, and throughput | This is a mocked local gateway benchmark; live upstream profiles remain optional and unrun. |
| The core has a deterministic golden fixture suite | VERIFIED_BY_UNIT_TEST | `docs/engineering/CORE-REGRESSION-HARNESS.md`; `internal/gemini/golden_test.go`; `internal/format/golden_test.go`; `internal/multimodal/golden_test.go` | These are synthetic protocol-shaped fixtures, not live Google captures. Live acceptance remains unknown. |
| The default Go suite is hermetic and provider-independent | VERIFIED_IN_SOURCE | Upstream-facing diagnostic workflows are tagged `live`; the default suite excludes them, while `go test -tags=live ./internal/diag` is documented separately | A live-tagged run remains provider/session-dependent and is intentionally not part of the default validation gate. |
| A green diagnostic result proves only that the advertised response contract was observed | VERIFIED_BY_UNIT_TEST | `internal/diag/diag.go` validates JSON, usable output, terminal SSE state, total stream size, and complete Anthropic lifecycle; focused tests cover malformed URLs, empty streams, incomplete lifecycle, and oversized bodies | A real no-cookie run on 2026-08-29 passed 13/15 and failed instruction-only web JSON plus provider-dependent image generation; live provider/session acceptance remains separate. |
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
| `go test -count=1 -cover ./...` | Passed; package coverage ranged from 0.0% to 89.3% | MEASURED | Current values include root 3.6%, browser 6.9%, service 0.0%, server 67.6%, updater 56.2%, Gemini 75.6%, and `pkg/gateway` 89.3%; no blanket 80% claim is valid. A weighted per-package profile over the same run was 63.6%. |
| Isolated binary smoke on `127.0.0.1:19613` | `/healthz`, `/playground`, favicon, manifest, and service worker returned 200; hostile-origin requests returned 403; exact-origin API access reached authentication; process stopped cleanly | VERIFIED_BY_INTEGRATION_TEST | This proves local HTTP serving and the tested origin gate only; it is not a provider or 30-device rollout test. |
| Inline JavaScript syntax parse | Two inline scripts parsed with Node without syntax errors | VERIFIED_IN_SOURCE | DOM behavior, CDN availability, and interactive artifact execution still require a real browser/device. |
| Release source provenance | `scripts/verify-release-source.sh` requires a Git worktree with no modified/untracked paths, validates the stable/preview version against the `Makefile` base, requires an explicit numeric `PREVIEW_VERSION`, requires every preview packager to fail closed without `BOB_RELEASE_VERSION`, compares the generated Web bundle without rewriting it, and checks the canonical Ed25519 key block against Makefile/package inputs, both standalone installers, and Docker metadata | VERIFIED_IN_SOURCE | This protects local packaging entry points and key/version coherence; the operator must still advance the candidate after each publication, match the private key, review the commit, publish immutable assets, and reconcile the public release. |
| Release directory and signed manifest reconcile one-to-one | `scripts/verify-release-assets.sh` and `updater.VerifySignedReleaseDirectory` verify the detached Ed25519 signature, exact SHA-256 bytes, and no extra/missing/duplicate/symlinked local asset | VERIFIED_BY_UNIT_TEST | This is a local pre-upload/post-download gate; GitHub upload contents and release metadata still require operator verification. |
| Browser chat history has bounded runtime and persisted state | `internal/server/playground.html` caps messages, content/reasoning, attachments, serialized storage, and oversized legacy payloads; guarded writes visibly distinguish normal, compacted, recovered, unavailable, and unsaved states; `playground_test.go` locks the guard markers | VERIFIED_IN_SOURCE | A clean-browser quota/performance run and long-session device acceptance remain separate gates. |
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
- Public-HTTPS browser/PNA deployment behavior and the minimum viable
  remote-Studio pairing/token design. The local cross-port browser boundary is
  recorded in `BROWSER-SECURITY-VALIDATION-2026-08-31.md`.
- Authenticated Scotty upload and Imagen generation.
- Release authenticity and safe updater replacement under mocked downloads.
- Final Wails-only desktop product decision (completed after the Phase III
  device comparison).
- Environment-tagged performance baselines.

## Phase III evidence update (2026-08-21; historical snapshot)

The historical Mission 0 statements above were not the branch status at that
audit boundary. The following evidence was recorded then. It is retained for
  provenance; the current Preview 6 publication and migration evidence in the
  final addendum below supersedes its release-specific rows.

| Claim at the 2026-08-21 audit boundary | Classification | Evidence | Boundary |
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
| The public release is a student-ready cross-platform native installer set | UNKNOWN | GitHub release inspection found CLI assets in `v0.1.5` but no trusted native installer set; the current public desktop preview is macOS-only and only macOS ARM64 ad-hoc native smoke testing is complete | Requires native Windows/Linux builds, clean-device tests, platform signing, macOS notarization, and manual release publication. |
| A public native desktop preview is available for controlled evaluation at the historical Preview 5 boundary | VERIFIED_LIVE | At that historical boundary, manually published prerelease `v0.2.0-preview.5` contained the branded macOS universal `.dmg`/`.zip`, `RELEASE-NOTICE.txt`, `SHA256SUMS`, and detached `SHA256SUMS.sig` assets; all five were re-downloaded, signature-verified, and byte-compared with the local signed candidate | This historical beta evidence does not replace the current Preview 6 addendum: there is still no Apple notarization, Windows publisher signature, Linux asset, clean-device matrix, silent updater, or production student-readiness claim. Installed replacement remains a device gate. |
| A free branded macOS preview package can be created without Apple membership | VERIFIED_BY_INTEGRATION_TEST | `scripts/package-wails-preview.sh` creates a branded `BOB Gemini Free.app`, `.zip`, `.dmg`, release notice, and checksums without Developer ID credentials; the bundle metadata uses the `com.abcsteps` identity | This proves local packaging only; it does not establish Gatekeeper trust, notarization, clean-device acceptance, or public student readiness. |
| The native desktop app exposes an explicit update check at the historical Preview 5 boundary | VERIFIED_BY_UNIT_TEST | At that historical boundary, `internal/updater/desktop.go` selected only official stable or `preview.N` channels, with a bounded preview listing; `internal/updater/updater_test.go` covered branded/legacy names, prerelease ordering, signed-manifest discovery, stable-first migration for newly built previews, stable-failure fail-closed behavior, and the endpoint bound | The current Preview 6 source adds the published-candidate matrix in the final addendum. Neither source evidence nor the published package silently installs, removes the macOS warning, or replaces platform publisher trust. |
| The embedded Studio update status uses the same release channel as the native desktop updater | VERIFIED_BY_UNIT_TEST | `cmd/desktop/main.go` passes its build-pinned channel to `server.NewWithUpdateChannel`; `internal/server/server_test.go` verifies preview forwarding to the channel-aware checker, while `playground_test.go` verifies native/preview-specific status guidance | The status route only discovers metadata. It does not prove public Preview 7 publication, installed-device replacement, rollback, Apple trust, or fleet rollout. |

## Preview 3 v0.2.0 release-readiness update — 2026-08-31 (historical)

The earlier release-readiness snapshot referenced source commit `e019cf8`.
The reviewed hardening work was merged through protected PR [#31](https://github.com/div197/BOB-Gemini-Free/pull/31)
and is preserved in the current `main` history. The release-coherence
follow-up was merged through protected PR [#33](https://github.com/div197/BOB-Gemini-Free/pull/33);
the installer trust-anchor follow-up was merged through protected PR [#36](https://github.com/div197/BOB-Gemini-Free/pull/36),
and the session-only gateway-auth follow-up through protected PR [#38](https://github.com/div197/BOB-Gemini-Free/pull/38).
The current public `main` tip at the pre-PR #42 audit checkpoint was merge
commit `523ceeb`; protected PR #42 subsequently merged the reviewed source at
`ba1b562`.
The historical Preview 2 native package evidence in this section was produced
from public-main commit `6d3a0cfc`; its release directory was signed and verified
through the local Keychain signer, and a fresh download of every public asset
was byte-compared with the local publication input. The signed
`v0.2.0-preview.1` migration bridge, superseded `v0.2.0-preview.2`, and
current controlled macOS `v0.2.0-preview.3` preview are published; stable
`v0.2.0` has not been tagged or published. The
separate
[`RELEASE-READINESS-v0.2.0.md`](RELEASE-READINESS-v0.2.0.md) is the authoritative
publication gate for this milestone.

| Claim at the Preview 3 readiness boundary | Classification | Evidence | Boundary |
|---|---|---|---|
| The stable build path embeds the repository's updater public key | VERIFIED_BY_INTEGRATION_TEST | `make build`, `make dist`, `make desktop-key-check`, and binary string inspection passed for the six CLI targets and the macOS Wails candidate | This proves the embedded trust anchor, not a signed release manifest or Apple/Windows publisher trust. |
| The release source gate rejects updater-key encoding and version/channel drift before packaging | VERIFIED_BY_INTEGRATION_TEST | `scripts/verify-release-source.sh` passed for stable and preview fixtures and failed closed when a standalone installer key or its SPKI encoding was changed; PR [#33](https://github.com/div197/BOB-Gemini-Free/pull/33) merged the source gate and protected PR [#36](https://github.com/div197/BOB-Gemini-Free/pull/36) merged the SPKI consistency check on `main` | This validates source/package inputs only; it does not sign, upload, or verify the public release assets. |
| The macOS v0.2.0-preview.2 package is package-valid and publicly reconciled | VERIFIED_BY_INTEGRATION_TEST | Fresh package from public-main `6d3a0cfc` produced a Wails universal app, ZIP, DMG, Applications shortcut, ad-hoc code signature, signed SHA-256 manifest, bundle metadata, local PWA routes, and native launch/health/shutdown proof; all five GitHub assets were re-downloaded, signature-verified, and byte-compared with local inputs | `spctl` rejection is expected without Apple notarization; package/public-byte verification proves neither clean-device update or student rollout. |
| Update metadata URLs are pinned to the official repository or GitHub release CDN | VERIFIED_BY_UNIT_TEST | `internal/updater/desktop.go` and `internal/updater/updater_test.go` reject other GitHub owners/repositories, non-HTTPS URLs, lookalike hosts, and unexpected ports | GitHub release metadata is still external state; the signed manifest remains the artifact authenticity boundary. |
| CLI update asset selection requires the canonical platform filename | VERIFIED_BY_UNIT_TEST | `internal/updater/updater.go:280-294` and `internal/updater/updater_test.go` reject suffix-only lookalikes and select the exact `bob-gemini-free-{platform}-{arch}` name | This protects local selection logic; the public release asset matrix and signed-upload reconciliation remain external gates. |
| Existing public Preview 7 installations can update directly to v0.2.0 stable | STALE_OR_INCORRECT | The published Preview 7 binary predates stable-first discovery; the published same-key `v0.2.0-preview.1` bridge is now available, while current source tests prove stable-first only for newly built packages | Install the bridge first, then bridge → stable after stable acceptance, or perform one manual stable installation. Preview 6 and older need manual current-key migration. |
| The source is ready for a student-facing signed stable publication | UNKNOWN | Local source, package, and mocked updater gates are green, but no exact signed/uploaded v0.2.0 assets, clean-device acceptance, platform publisher trust, or 30-device migration run exists | Do not publish or announce stable until the owner completes the release sequence and Gates A–C in `RELEASE-READINESS-v0.2.0.md`. |

## Current Gemini Developer API routing update (2026-08-29)

The 2026-08-31 settings review added an explicit credential map to the Studio
Config modal. This improves operator/student comprehension without changing the
wire contract: BOB access authentication, the optional Developer API key, and
the web-session cookie state remain separate boundaries.

This section records the explicit student-owned-key route added after the
earlier web-RPC audit. It does not change the default cookie/guest path and it
does not turn provider documentation into a product guarantee.

| Current claim | Classification | Evidence | Boundary |
|---|---|---|---|
| A student can opt into a separate Gemini Developer API route from the Web Studio | VERIFIED_IN_SOURCE | `internal/server/playground.html` links to `https://aistudio.google.com/app/apikey`, keeps the key in page memory, and sends it only as `X-BOB-Gemini-API-Key` when the session toggle is enabled | This is a local UI/source claim; the student still owns the Google project/key and must follow Google's account, billing, and data-use rules. |
| The provider key is translated to Google's documented header and never placed in the URL or request body | VERIFIED_BY_UNIT_TEST | `internal/geminiapi/client.go`, `internal/geminiapi/geminiapi_test.go`, and `internal/server/gemini_api_test.go` assert `x-goog-api-key`, no query key, no body key, and redacted provider errors | Runtime proxies or provider-side logging remain outside BOB's control. |
| Direct Developer API typed streams reject empty semantic events and classify provider error events | VERIFIED_BY_UNIT_TEST | `internal/geminiapi/client.go` rejects `{}` and `candidates:[]` events unless prompt feedback or usage metadata is present, recognizes numeric/string provider status codes from HTTP-200 error events, and redacts the API key; `internal/geminiapi/geminiapi_test.go` covers unknown SSE fields, ordered multiline data, quota error events, comments/data/`[DONE]` framing | Full provider event vocabulary and Web RPC framing remain unverified. |
| Direct Developer API tool-choice names are checked against declared tools | VERIFIED_BY_UNIT_TEST | `internal/geminiapi/translate.go:23-25`; shared `internal/format.ValidateToolChoice`; `internal/geminiapi/geminiapi_test.go` covers undeclared named choices | Gemini `AUTO`/`NONE`/`ANY` mapping is source-backed; provider acceptance and exact model semantics remain live-dependent. |
| Direct Developer API tool-call arguments obey the target object shape | VERIFIED_BY_UNIT_TEST | `internal/geminiapi/translate.go` uses one bounded decoder for tool arguments; `internal/geminiapi/geminiapi_test.go` covers malformed JSON plus scalar, array, and `null` values | The public Gemini `FunctionCall.args` field is documented as a JSON object; the local web-RPC prompt emulation remains a separate partial capability. |
| Direct Developer API typed text parts cannot disappear on conversion | VERIFIED_BY_UNIT_TEST | `internal/geminiapi/translate.go:196-218`; `internal/geminiapi/geminiapi_test.go` covers missing `type` and non-string `text` | This is a request-validation guarantee, not evidence that every provider content part is supported. |
| Explicit Developer API chat uses native public REST/SSE rather than the web-RPC cookie path | VERIFIED_BY_UNIT_TEST | `internal/server/gemini_api.go` and `internal/server/gemini_api_test.go` prove the explicit route calls the isolated client and the web client is not called | Live Google acceptance, exact model availability, quota, and semantic parity remain unverified. |
| Direct provider routing supports future Gemini model IDs without inventing BOB aliases | VERIFIED_IN_SOURCE | `directGeminiModel` maps only local convenience aliases and forwards provider-shaped Gemini IDs; Google remains the authority for whether a model exists | The Web Studio dropdown and BOB web-RPC model catalog are separate; a provider model must still be selected through a supported client surface and accepted by Google. |
| The UI or gateway promises a fixed free tier, universal RPM/RPD, or unlimited access | STALE_OR_INCORRECT | Current README, Hindi README, routing guide, and glossary link to Google's pricing/rate-limit/billing pages and explicitly reject fixed quota promises | Google can change limits by model, project, tier, region, account, and date; re-check official pages before every release. |
| The student's Developer API key is silently rotated or used as a fallback after a web-route failure | VERIFIED_BY_UNIT_TEST | Single-key resolver and explicit route tests reject repeated keys and preserve provider selection; UI/docs state no rotation and no cross-route replay | A caller can make a new explicit request after an error; BOB does not hide that provider change. |
| Browser storage retains a BOB gateway-auth token after a student leaves a shared computer | VERIFIED_BY_UNIT_TEST | `internal/server/playground.html` keeps `gatewayApiKey` in page memory, purges the legacy `bob_api_key` key once at startup, and uses the memory value for health/model/generation requests; `playground_test.go` rejects persistent reads/writes | A compromised same-origin script can still read any credential while the page is open; page-memory scope reduces persistence but is not a browser sandbox. |
| Gateway API-key auth accepts standard case-insensitive Bearer schemes without treating a provider key as local gateway auth | VERIFIED_BY_UNIT_TEST | `internal/server/middleware.go`, `internal/server/server_test.go`, and `internal/server/gemini_api_test.go` cover lower-case `bearer`, provider-only rejection, local-plus-provider separation, configured provider routing, and rejection on unsupported adapters | This proves request-boundary semantics only; it does not validate a provider key, quota, or live Google access. |
| A Developer API key is supported on every BOB adapter and image path | STALE_OR_INCORRECT | Explicit keys are currently rejected on `/v1/messages`, `/v1/responses`, and `/v1/images/generations`; only chat and native Google generation/counting are wired | Additional adapters need their own translation fixtures and release evidence before being enabled. |
| Current free-tier and model facts are maintained as a dated release gate rather than hardcoded forever | VERIFIED_IN_SOURCE | `docs/engineering/GEMINI-API-ROUTING.md` defines the official-source review and the Unreleased changelog records the policy | This is a process control, not live quota telemetry; students must inspect the current AI Studio project view. |
| A configured TLS fingerprint name silently selects a different browser profile | VERIFIED_BY_UNIT_TEST | `internal/gemini/fingerprint.go` maps documented Chrome/Firefox/Safari names to the matching `tls-client` profile and rejects unknown names; `internal/gemini/stream_test.go` covers `chrome_133` and invalid-profile behavior | Optional fingerprinting is not a provider-availability or quota workaround; standard transport remains the explicit behavior when an invalid optional profile cannot be constructed. |

## Current 100-path failure-register update (2026-08-31)

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
| Oversized or malformed upstream bodies can be read without a bound | VERIFIED_BY_UNIT_TEST | `internal/gemini/client.go` bounds normal bodies and stream lines and tests oversized response/line and nil-response cases; `internal/multimodal/tokens.go` bounds page-token responses and token values and rejects non-success/redirect responses | Limits are application-level; the provider and network can still return errors or slow responses. |
| Unsafe Scotty upload URLs, upload statuses, response bodies, or image dimensions are accepted | VERIFIED_BY_UNIT_TEST | `internal/multimodal/upload.go`, `compress.go`, and golden tests enforce host/scheme/path, status, body, byte, dimension, pixel, and base64 limits; `CompressImageBytesIfNeeded` normalizes oversized dimensions even when encoded bytes already fit the budget; remote-image DNS is revalidated at the final guarded dial; a vision-upload failure is not silently replayed as OCR-only text | OCR CPU/memory pressure and provider session capability remain open/external rows; live egress topology is not proven. |
| A cookie file change with unchanged mtime/size remains invisible, or a broad-permission/symlink file is accepted | VERIFIED_BY_UNIT_TEST | `internal/gemini/auth.go`, `auth_test.go`, and pool tests use content hashing plus secure-file validation and cover deletion, permissions, and deduplication | The operating system and user account still control file ownership and keychain/session validity. |
| The native updater follows an arbitrary redirect or accepts an unbounded/mis-sized release body | VERIFIED_BY_UNIT_TEST | `internal/updater/desktop.go`, `updater.go`, `desktop_stage.go`, and updater tests constrain official hosts, redirects, metadata, signatures, artifact size, and exact bytes; read-only metadata retries at most once for a transient transport/timeout error and tests prove cancellation, non-network, and HTTP status failures are not retried | A real update from `/Applications`, rollback after interruption, platform publisher trust, and public upload reconciliation remain external gates. |
| The published Preview 2 asset upload preserves the locally signed bytes | VERIFIED_LIVE | [`PREVIEW-2-PUBLICATION-2026-08-31.md`](PREVIEW-2-PUBLICATION-2026-08-31.md) records all five GitHub assets being re-downloaded, detached-signature-verified, and byte-compared with the local publication inputs | Repeat this fresh-directory reconciliation for every future release; it does not prove installed-bundle replacement. |
| Native updater replacement races or target swaps can silently replace the wrong install | VERIFIED_BY_UNIT_TEST | `internal/updater/desktop_helper.go` uses an atomic per-install lock with conservative stale-lock recovery and revalidates target/candidate after locking; updater tests cover conflict, stale non-empty refusal, and replacement paths | Filesystem/OS interruption and clean-device rollback remain external gates. |
| Native background update checks avoid a synchronized classroom metadata burst | VERIFIED_BY_UNIT_TEST | `cmd/desktop/updates.go` adds a 30-second startup delay plus a cryptographically random jitter capped at five minutes; `cmd/desktop/main_test.go` covers clamping and preserves the once-daily/non-installing loop | This reduces release-metadata synchronization only; Google generation capacity and a live 20–30-device classroom run remain external. |
| Stale updater staging directories can accumulate without bounded cleanup | VERIFIED_BY_UNIT_TEST | `internal/updater/desktop_stage.go` removes only old committed plans that match the exact target and pass plan validation; focused cleanup tests keep fresh/unrelated directories | Crashes before plan commit remain for manual inspection by design. |
| ZIP extraction can accept ambiguous, duplicate, or special entries | VERIFIED_BY_UNIT_TEST | `extractMacOSApp` rejects traversal components, duplicate normalized paths, symlinks, special files, unexpected roots, and expansion beyond the archive bound; focused tests cover duplicate/special entries | Real signed package contents still require release-artifact review. |
| The artifact editor can remain empty after an interactive preview has a source, chat scroll controls can remain hidden after layout, or repeated streaming renders can grow artifact state without bound | VERIFIED_IN_SOURCE | `internal/server/playground.html` normalizes Marked token objects, recovers an empty editor from the registered source, recalculates directional controls after layout/resize, uses stable render-scoped artifact IDs, caps each source at 2 million characters and the registry at 128 items/8 million characters, reports Mermaid/Pyodide dependency failures to the host, and `playground_test.go` locks the markers | The in-app browser environment did not load CDN libraries during this run, so full interactive artifact execution, dependency-failure rendering, and clean-browser memory acceptance still require a device check. |
| Local history can persist a deliberately truncated image data URL or unbounded attachment text | VERIFIED_IN_SOURCE | Client-side attachment size, image data URL, extracted-text, and attachment-count limits are defined in `playground.html`; extraction work is capped, storage tests reject the old truncation pattern, guarded persistence reports compacted or unsaved outcomes to the user, and attachment reads/worker OCR are cancellable when the browser APIs support it | Browser localStorage quota, main-thread parsing cost, legacy fallback cancellation, and large parsing/OCR CPU remain environment-dependent; bounded history retention is still open. |
| Scotty image-reference cache can grow without bound or cross a configured account pool | VERIFIED_BY_UNIT_TEST | `internal/server/image_cache.go` bounds references to 256 with LRU-like eviction, expires local entries after a conservative 15-minute age, and shares in-flight loads; `helpers.go` scopes reuse to a hashed single-cookie session and disables it for configured pools; cache/scope tests cover eviction, expiry, refresh, cookie rotation, waiter cancellation, and concurrent loads | Scotty's actual provider-side expiry remains undocumented and requires live authenticated evidence; local expiry deliberately trades some cache hit rate for stale-reference safety. |
| Browser transport, protocol, provider-stream, or deep-refiner errors are stored as successful state | VERIFIED_IN_SOURCE | `internal/server/playground.html` treats `finish_reason: error`, structured SSE errors, interrupted reads, incomplete EOF, and failed deep-refiner responses as terminal failures; it stores `status: error` or `stopped`, excludes non-complete assistant turns from later prompts, and leaves the original prompt unchanged when refinement fails; `internal/server/playground_test.go` locks these lifecycle markers | Full behavior still needs a clean browser/device execution because this run's isolated in-app browser did not load external CDN libraries. |
| Public endpoint errors can echo signed URLs, provider transport details, generated output, or update-server metadata | VERIFIED_BY_UNIT_TEST | `internal/server/helpers.go` supplies credential-safe attachment, upstream, and update-check summaries; image generation no longer returns raw generated text as `details`; server tests cover source URL and Web RPC token non-disclosure, and Gemini retry logs use bounded kind/status summaries | Deployment logging, reverse proxies, browser history, and third-party provider logs remain outside BOB's control. |
| Native updater dialogs expose raw transport, GitHub metadata, filesystem paths, or OS error text, or make a read-only/staging failure look like an app change | VERIFIED_BY_UNIT_TEST | `cmd/desktop/updates.go` maps check, staging, and start failures to bounded actionable messages that state no app change occurred; `cmd/desktop/main_test.go` rejects timeout, HTTP status, low-level transport, local-path, and read-only error-text leakage | Native dialog rendering and the actual GitHub/filesystem environment remain device/external evidence surfaces. |
| Tool schema complexity can bypass limits through typed Go values or cyclic structures | VERIFIED_BY_UNIT_TEST | `internal/format/schema.go` uses reflection-safe structural accounting for maps, slices, arrays, pointers, interfaces, typed enum/property collections, node count, and depth; `internal/format/schema_test.go` covers typed enum and cyclic-map fixtures | JSON-provider semantic support remains partial and streamed tool-call aggregation is a separate open concern. |
| JSON handlers can allocate an unbounded request body when invoked outside the normal middleware chain | VERIFIED_BY_UNIT_TEST | `internal/server/helpers.go` exposes one bounded `readRequestBody` seam used by chat, responses, Anthropic, Google, image, token, refine, and direct handler paths; `internal/server/server_test.go` proves an oversized body is rejected before JSON conversion | The bound protects local memory; request semantics and provider limits remain separate concerns. |
| Any loopback browser origin is trusted by default | VERIFIED_BY_INTEGRATION_TEST | `internal/server/middleware.go` compares the browser origin to the exact request host/scheme only when the request host is a literal loopback host; `security_boundary_test.go` rejects a different loopback port and non-loopback Host/origin confusion while preserving explicit remote-origin configuration; `BROWSER-SECURITY-VALIDATION-2026-08-31.md` records the corresponding real in-app-browser cross-port probe | Public-HTTPS/PNA deployment behavior and explicit remote-origin trust remain separate gates. |
| Shared stream or direct Developer API callbacks can panic when an embedding passes nil functions or context | VERIFIED_BY_UNIT_TEST | `internal/gemini/flight.go`, `internal/geminiapi/client.go`, and `internal/server/gemini_api.go` normalize nil contexts and return explicit callback errors; focused tests cover each seam | This protects local API misuse; it does not make a cancelled provider transport resumable. |
| Generation POST retries are limited to failures known to precede provider request delivery | VERIFIED_BY_UNIT_TEST | `internal/gemini/client.go` retries only `net.OpError` dial/connect/lookup failures; tests cover safe pre-request retry and ambiguous transport no-retry behavior | HTTP 5xx, response-read, parser, and partial-stream failures remain non-replayed because the web-RPC POST has no idempotency contract. |
| Direct Developer API stream tool calls are emitted only after valid bounded assembly | VERIFIED_BY_UNIT_TEST | `internal/server/gemini_api.go` keeps ID-ordered cumulative snapshots, validates arguments, rejects changed IDs/names, and emits finalized calls once; stream-error tests prove no assistant Markdown fallback | This protects the direct public REST/SSE route only; the reverse-engineered web-RPC tool path remains emulated. |
| Direct Developer API empty or metadata-only generation responses/streams fail closed | VERIFIED_BY_UNIT_TEST | `internal/geminiapi/client.go` validates typed generation responses, rejects an empty body, `[DONE]` without a JSON event, invalid JSON, and streams containing only usage/feedback/finish metadata; `internal/geminiapi/geminiapi_test.go` covers these fixtures. `internal/server/gemini_api.go` applies the same semantic check to the native Google-shaped raw route, with native-route fixtures in `internal/server/gemini_api_test.go`. | This prevents a fabricated successful stop; provider stream semantics beyond the tested SSE grammar remain unverified. |
| Web-RPC generation streams with no usable model text fail closed | VERIFIED_BY_UNIT_TEST | `internal/gemini/stream.go` tracks usable parsed text; `internal/gemini/client.go` returns a protocol error for an empty/metadata-only stream; `internal/gemini/client_test.go` covers empty and unparseable bodies | This prevents the adapter from fabricating a normal stop; Google Web RPC framing and provider-specific metadata semantics remain unverified. |
| Tool-result continuations cannot silently bind to an unknown, mismatched, or ambiguous call | VERIFIED_BY_UNIT_TEST | `internal/format/openai.go` validates IDs/names before prompt flattening and direct translation; valid, unknown, mismatch, and ambiguous fixtures are covered | Assistant calls without IDs remain accepted for legacy compatibility; callers should provide stable IDs for parallel tools. |
| Direct Developer API candidate and finish-reason ambiguity is hidden as success | VERIFIED_BY_UNIT_TEST | Direct translation rejects multiple candidates and fails closed on unknown finish reasons while mapping known stop/length/filter/tool outcomes | The raw web-RPC route has separate provider-dependent semantics and is not claimed equivalent. |
| A successful non-JSON Developer API response is passed through as if it were a valid provider response | VERIFIED_BY_UNIT_TEST | `internal/geminiapi/client.go` validates bounded successful JSON before raw handlers return it; `internal/geminiapi/geminiapi_test.go` covers an HTML success body | JSON validity does not establish provider schema validity; typed handlers still validate the decoded response before presenting model output. |
| UTF-8 error truncation corrupts a provider/API error or violates its output bound | VERIFIED_BY_UNIT_TEST | `internal/geminiapi/sanitizeMessage` truncates by rune after removing control characters; `internal/server/publicErrorText` applies the final URL/credential-marker boundary before public responses or optional logs; tests cover 500 Devanagari runes and credential-shaped provider text | Error text is intentionally shortened and is not a diagnostic transcript. |
| An artifact pop-out removes the iframe sandbox or leaves a generated object URL alive forever | VERIFIED_BY_UNIT_TEST | `popOutArtifact` creates a sandboxed nested iframe, uses `noopener,noreferrer`, reports blocked windows, and defers URL revocation; `internal/server/playground_test.go` locks these markers | Browser popup policy and actual generated-artifact behavior still need device/browser acceptance. |
| A refiner stage returns empty or oversized content, multiplying memory into later prompts | VERIFIED_BY_UNIT_TEST | `internal/refiner/refiner.go` bounds user prompts and each of its three stage outputs, checks cancellation between stages, and `internal/refiner/refiner_test.go` covers empty/oversized results | The refiner performs three inference calls, not a pure-local four-stage guarantee; provider latency and quota remain upstream-dependent. |
| An embedded or mobile caller silently gets zero-valued/default model routing after an invalid explicit model | VERIFIED_BY_UNIT_TEST | `pkg/gateway` and `pkg/mobile` use strict model resolution and nil/lifecycle guards; `internal/models/models_test.go`, `pkg/gateway/gateway_test.go`, and `pkg/mobile/mobile_test.go` cover failure paths | The public adapter aliases intentionally retain their legacy fallback behavior; provider model identity remains unverified. |
| An embedded guest handler starts an unnecessary session-reload worker or has no close path for configured session state | VERIFIED_BY_UNIT_TEST | `internal/server/server.go` skips idle reload for guest-only configs; `pkg/gateway` exposes `Engine.Close` and `CloseableHandler`, and tests close configured fixtures | HTTP server shutdown must still call the exported close path; `NewHandler` cannot infer the embedding server's lifecycle automatically. |
| Health, metrics, image upload, or refinement called on a partially initialized embedded app panics | VERIFIED_BY_UNIT_TEST | `internal/server/handlers.go`, `helpers.go`, and `refine.go` guard optional components; `internal/server/server_test.go` covers partial health/metrics/upload/refine requests | These are defensive embedding guards, not permission to bypass normal `server.New` construction. |
| Standalone installer accepts an unsigned or unverified release binary | VERIFIED_IN_SOURCE | `install.sh` and `install.ps1` use HTTPS-only downloads, fixed public-key verification, exact signed SHA-256 entries, size bounds, atomic local installation, and no default source/unsigned fallback | Shell/PowerShell clean-host execution and availability of an Ed25519-capable verifier remain external; macOS stock LibreSSL may fail closed. |
| Release evidence omits the source commit, toolchain, or exact signed asset hashes | VERIFIED_IN_SOURCE | `scripts/record-release-evidence.sh` re-runs source/asset verification and records commit, branch, Go version, host, time, manifest/signature hashes, and asset hashes in a 0600 receipt outside the worktree | The operator must retain and reconcile the receipt with the public GitHub release; it is not a hosted attestation. |
| The whole 100-path register is closed | STALE_OR_INCORRECT | The register contains 100 numbered paths with explicit PROTECTED/PARTIAL/OPEN/EXTERNAL statuses | Do not use this matrix or the register as a production-readiness certificate; close each release/device/provider gate with its own evidence. |

## Current version-identity hardening — 2026-08-31

| Current claim | Classification | Evidence | Boundary |
|---|---|---|---|
| An unflagged source build cannot present a Go pseudo-version as a published updateable release | VERIFIED_BY_INTEGRATION_TEST | `main.go` now accepts only the explicit `Version` build variable; `main_test.go` covers `dev`, an injected release version, and an empty injection. A direct `go build .` reproduction changed from `v0.2.0-preview.1.0.20260831062407-8ce3483234a4` to `dev`, while `make build` still reports its explicit `v0.2.0` build value. | This protects the root CLI identity; release/packaging commands must still inject the intended version and remain subject to the clean-source/public-asset gates. |
| The updater accepts only canonical published stable or `preview.N` versions | VERIFIED_BY_UNIT_TEST | `internal/updater/updater.go` rejects Go pseudo-versions, build metadata, incomplete cores, and non-preview prereleases; updater tests cover those shapes. Stable metadata is also rejected when its tag is not canonical. | A valid version and signature do not prove that a release was uploaded, Apple-notarized, or accepted on a clean device. |

## Latest control-plane and false-success hardening

| Claim | Classification | Evidence | Boundary |
|---|---|---|---|
| The browser Studio must not send a BOB Gateway Access Key over non-loopback cleartext HTTP | VERIFIED_IN_SOURCE | `internal/server/playground.html` centralizes gateway request headers behind a loopback-HTTP/HTTPS transport check; `playground_test.go` locks the helper, all four request paths, and the visible blocked state. A 2026-08-31 Playwright probe against the freshly served local bundle observed zero `Authorization` headers on the mocked `http://192.168.1.100:9610` telemetry requests, no request from the blocked Test Ping, visible `HTTPS REQUIRED`, and preserved authorization for loopback HTTP and HTTPS fixtures. | HTTPS does not establish remote operator identity; exact-origin CORS, explicit endpoint choice, and deployment-level trust remain separate gates. |
| A malformed discovered config silently falls back to defaults | VERIFIED_BY_UNIT_TEST | `internal/config/config.go`, `cmd/desktop/config.go`, `main.go`, and config/desktop tests bound the config file and return a visible startup error for invalid JSON or oversized content | A user must still repair or remove the invalid local file; the error does not infer the intended configuration. |
| A status or desktop gateway health probe can allocate an unbounded JSON body | VERIFIED_IN_SOURCE | `main.go` and `cmd/desktop/gateway.go` decode through bounded readers; the probe still treats status/identity headers as a local compatibility signal, not authentication | A local process can spoof compatibility headers; reuse remains a convenience decision and is not a security credential. |
| A CDP version endpoint can return an unbounded body or leak into repeated response handling | VERIFIED_IN_SOURCE | `internal/browser/browser.go` bounds `/json/version` to 64 KiB and closes each response exactly once before retrying | CDP is still a local browser trust boundary and requires a local browser/device check. |
| A corrupted or oversized updater plan/confirmation file can drive unbounded reads or a path outside the transaction layout | VERIFIED_BY_UNIT_TEST | `internal/updater/desktop_helper.go`, `desktop_recovery.go`, and `desktop_stage.go` use bounded reads and enforce updater-owned staging, plan, rollback, confirmation, and helper names; updater tests cover oversized metadata | A same-user local attacker and filesystem race are outside this application-level contract; clean-device interruption testing remains external. |
| The CLI reports a browser as opened when both app-mode and system-browser fallback fail | VERIFIED_BY_UNIT_TEST | `main.go` returns fallback errors and `main_test.go` covers the failed fallback path; the caller logs failure rather than success | The OS browser command and desktop session remain external dependencies. |
| Wails can remain on an indefinite bootstrap spinner after its gateway binding/configuration fails | VERIFIED_IN_SOURCE | `cmd/desktop/main.go` supplies a visible startup-error window and `cmd/desktop/frontend/index.html` displays a retryable bootstrap error when the bound gateway method rejects | Actual native WebView rendering still needs a clean-device GUI check. |
| The direct Developer API embedded client can panic or silently proceed when its receiver is nil | VERIFIED_BY_UNIT_TEST | `internal/geminiapi/client.go` rejects a nil client before endpoint construction; `geminiapi_test.go` covers `GenerateRaw` and `Stream` | This guards embedding misuse; it does not validate Google credentials or provider availability. |
| Rate-limit headers describe an enforced local quota when no limiter exists | VERIFIED_IN_SOURCE | `internal/server/middleware.go` now emits only request/processing/version metadata; server tests assert synthetic `x-ratelimit-limit-requests` is absent | Google/provider limits remain external and are shown only through provider errors or current official documentation. |
| A live benchmark invents token throughput when usage metadata is absent | VERIFIED_BY_UNIT_TEST | `internal/diag/bench.go` counts only positive provider-reported `total_tokens` values and exposes `TokenCountsMeasured`; `internal/diag/bench_test.go` proves missing usage produces no fabricated token total or throughput, and `main.go` labels unavailable token throughput explicitly | Request latency/throughput remains a local measurement; token estimates and provider quotas remain separate evidence surfaces. |

## Current local continuation evidence — 2026-08-31 (historical snapshot)

This continuation snapshot supersedes older commit labels in the historical
sections above, but is itself retained for provenance. The current Preview 6
publication addendum at the end is authoritative for the present release.
The reviewed code tip at the start of this continuation was `cec4c8e`; it contains the
deterministic stream-regression, session-bound image-reference, local
history-persistence, Developer API stream-error, session/quota-error,
Anthropic lifecycle, attachment-cancellation, fail-closed preference-storage,
root-CDN-integrity, queued-attachment-cancellation, dedicated service
health-probe, dynamic-artifact-CDN, response-status-logging, literal-loopback
CORS, page-token-retry, and Studio SSE-field-compatibility follow-ups
after the evidence documents, including the Preview 1 → Preview 2
version-transition fixtures, strict release-asset validation, and a
deterministic history-limit stream regression. The signed public
`v0.2.0-preview.1` bridge is immutable; controlled macOS `v0.2.0-preview.3`
is now published from public-main commit `284b7d1a`.
The local branch is
`codex/release-readiness-v0.2.0`; its reviewed source is now included in
public `main` through protected PRs #42–#48 and #53–#55. The Preview 3 publication baseline
was rechecked against public `main` before publication. No stable release was tagged, no GitHub Actions
workflow was added or invoked, and no provider secret was used. The release
private key was used only inside the owner-controlled local Keychain signing
operation and was never exposed to source, Git, chat, or a student package.

The updater durability continuation is now part of the public source history:
`b136724` flushes and synchronizes Unix transaction state around swaps and
recovery, while `fd279aa` uses Windows `MoveFileExW` replace-existing and
write-through semantics for updater metadata. Their documentation commits are
also merged through PRs #44–#46. Controlled macOS `v0.2.0-preview.3` is
published and its public bytes are reconciled; stable `v0.2.0` is not
published.

| Claim at the continuation audit boundary | Classification | Evidence | Boundary |
|---|---|---|---|
| A slow coalesced stream subscriber is never silently dropped | VERIFIED_BY_UNIT_TEST | `internal/gemini/flight.go` returns `ErrStreamSubscriberTooSlow` from a bounded queue; the overflow test synchronizes both subscribers and paces the burst on healthy-leader consumption, then passes repeatedly under `go test -race` | This protects the local multiplexer; live Google stream behavior and browser disconnects remain external. |
| Cancelling a leader or follower does not cancel an independent healthy subscriber, while the last subscriber cancels an abandoned flight | VERIFIED_BY_UNIT_TEST | `internal/gemini/flight_test.go` covers follower cancellation, leader cancellation, leader deadline, and last-subscriber cleanup under the race detector | The upstream client's own total timeout remains the final bound; this is not a provider retry guarantee. |
| A fetched remote image cannot bypass downstream decode dimensions through high compression | VERIFIED_BY_UNIT_TEST | `internal/multimodal/upload.go` validates fetched bytes with `inspectImageData`; `TestFetchImageBytesRejectsImagesOutsideDecodeBudget` covers a highly-compressible image over the source-dimension budget | OCR/browser CPU pressure and live remote-image egress remain external. |
| The native updater explains an unwriteable or translocated install before downloading a package | VERIFIED_BY_UNIT_TEST | `CheckDesktopInstallTarget` performs a no-network same-filesystem preflight; tests cover a writable bundle, App Translocation, and unsupported OS, while `cmd/desktop/updates.go` defers background prompts and surfaces manual guidance | A real mounted-DMG, `/Applications`, Gatekeeper, helper restart, rollback, and 30-device run remain external. |
| The current source can prove Preview 7 → bridge → stable selection without pretending it performed a live install | VERIFIED_BY_UNIT_TEST | `internal/updater/updater_test.go` covers the legacy preview-only lookup, same-key bridge discovery, stable-first migration, preview continuation, and stable-check failure | The published Preview 3 bytes are reconciled separately; private release-key custody and clean-device replacement still require operator evidence. |
| Removing an attachment can leave a FileReader, active PDF.js task, or supported OCR worker running and later mutate removed UI state | VERIFIED_BY_UNIT_TEST | `internal/server/playground.html` binds each attachment to an `AbortController`, removes cancelled queued parses, aborts `FileReader` reads on removal, destroys active PDF.js loading/document tasks, suppresses stale fallback results, and terminates a Tesseract.js v5 worker when available; `internal/server/playground_test.go` locks the cancellation markers | Main-thread DOCX/XLSX work, older browser fallbacks, OCR CPU cost, and rendered device behavior remain partial. |
| Denied browser storage can break preference initialization or click handlers | VERIFIED_BY_UNIT_TEST | `internal/server/playground.html` routes language, transliteration, panel, theme, endpoint, speech, reading-zoom, and custom-instruction preferences through fail-closed helpers; `internal/server/playground_test.go` rejects direct preference-storage bypasses | Preferences intentionally remain session/default-only when storage is unavailable; chat-history quota and long-session CPU remain separate browser gates. |
| A changed root CDN script or stylesheet can execute without an integrity check | VERIFIED_BY_UNIT_TEST | The playground head pins every external script/stylesheet with SHA-384 SRI and anonymous CORS, and pins Tesseract.js to `5.1.1`; `playground_test.go` rejects root dependencies without integrity attributes or a floating Tesseract major URL | Dynamic artifact `srcdoc` libraries, PDF worker/language assets, CDN availability, CSP behavior, and offline acceptance remain separate browser/runtime gates. |
| A dynamic artifact preview can execute a mutable Mermaid or Pyodide bootstrap | VERIFIED_BY_UNIT_TEST | Mermaid is pinned to `10.9.0` and Pyodide to `0.26.2` with exact SHA-384 SRI and anonymous CORS inside their `srcdoc` bootstraps; `TestDynamicArtifactCDNBootstrapsArePinned` rejects the former floating/missing-integrity URLs | Pyodide package/worker subresources, PDF worker/language assets, CDN availability, CSP behavior, and browser execution remain separate gates. |
| The optional service-status command reports an unrelated process as the BOB gateway | VERIFIED_BY_UNIT_TEST | `internal/service/service.go` probes the unauthenticated `/healthz` route and requires the BOB gateway identity and protocol headers; `internal/service/service_test.go` covers valid, wrong-status, missing-identity, wrong-identity, wrong-protocol, and unrelated-process responses | The headers are a compatibility signal, not an authentication credential; a real OS service manager and device process remain external. |
| Request logs report a status different from the status committed to the client after repeated headers or streaming flushes | VERIFIED_BY_UNIT_TEST | `internal/server/middleware.go` records only the first committed status and marks implicit `Write`/supported `Flush` commits; `internal/server/middleware_test.go` covers repeated `WriteHeader`, implicit `Write`, and `Flush` followed by a later header | This protects local observability only; reverse-proxy logs and external log collectors remain outside BOB's control. |
| A transient page-token refresh failure suppresses recovery for the full cache TTL or causes an immediate retry storm | VERIFIED_BY_UNIT_TEST | `internal/multimodal/tokens.go` preserves the last-known-good token set, reserves one in-flight refresh, and schedules a 15-second retry backoff after failure; `TestTokenCacheRetriesFailedRefreshAfterBoundedDelay` covers immediate suppression, backoff expiry, and recovery | The delay is a local availability guard; Google token lifetime and authenticated image capability remain provider-dependent. |
| The Studio drops a valid SSE frame when the `data:` field omits the optional space or uses CRLF | VERIFIED_BY_UNIT_TEST | `internal/server/playground.html` extracts the payload after the SSE field name and trims only the optional separator; `TestStudioSSEParserAcceptsStandardDataFieldForms` rejects the old space-dependent branch and checks `[DONE]` handling | Browser rendering and full multi-line event assembly remain unverified; the parser still deliberately treats the gateway's one-frame JSON events as the supported shape. |
| The local Studio shell is browser-verified at desktop, tablet, and phone widths | VERIFIED_LIVE | A fresh in-app browser loaded `http://127.0.0.1:19613/playground` and was checked at 1440x900, 1024x768, and 390x844; each viewport retained `document.documentElement.scrollWidth === innerWidth`, the shell rendered without console warnings/errors, and responsive drawer open/close behavior was exercised | This is live evidence for the local cold shell and drawer interaction only. Long streamed responses, provider errors, artifact execution, 200% zoom, and a clean-device/native package walkthrough remain separate gates. |
| A responsive drawer can strand keyboard focus outside its visible surface | VERIFIED_LIVE | On the phone viewport, opening both configuration and integration drawers focused the first drawer control, Shift+Tab/Tab wrapped within the drawer, Escape closed it, and focus returned to `btn-toggle-left`/`btn-toggle-right`; source regression markers cover the lifecycle | This proves the current local browser build, not every embedded WebView or assistive-technology combination. |

## Preview 3 publication addendum — 2026-08-31 (historical)

The current public release is the controlled macOS universal prerelease
[`v0.2.0-preview.3`](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.3),
targeting public `main` commit `284b7d1a`. Its five public assets were built
from the clean public tip, signed through the owner-controlled macOS Keychain,
freshly downloaded, detached-signature-verified, and byte-reconciled. The
complete evidence is in
[`PREVIEW-3-PUBLICATION-2026-08-31.md`](PREVIEW-3-PUBLICATION-2026-08-31.md).

| Claim at the Preview 3 publication boundary | Classification | Evidence | Boundary |
|---|---|---|---|
| The current public macOS preview is a complete, branded, signed-project artifact set | VERIFIED_LIVE | Preview 3 includes the branded universal `.dmg`/`.zip`, `RELEASE-NOTICE.txt`, `SHA256SUMS`, and detached `SHA256SUMS.sig`; the exact public assets match the local signed inputs byte-for-byte | The project signature authenticates release bytes, not Apple Developer ID identity; the package remains ad-hoc signed, non-notarized, and macOS-only. |
| Preview 3 is built from the current public source and carries the current updater trust anchor | VERIFIED_LIVE | Release-source verification passed at public-main `284b7d1`; the Keychain-backed signer accepted the private/public pair and the Wails build embedded the checked-in public key | Keychain presence and manifest signing do not prove future key custody, Apple trust, or a live installed-bundle update. |
| The exact Preview 3 artifact launches locally and owns a healthy loopback gateway | VERIFIED_LIVE | Fresh artifact launch bound to `127.0.0.1:8081` when 9610 was occupied, returned `{"status":"ok"}` from `/healthz`, and shut down cleanly | This is one host's package smoke test; it does not prove Google acceptance, provider quota, clean-device replacement, rollback, or fleet rollout. |
| Existing Preview 7/Preview 2 users will silently update to Preview 3 | STALE_OR_INCORRECT | The updater is explicit and user-consented; public metadata and source selection are verified, but installed replacement has not been observed on a clean student device | Use **Help → Check for Updates**, verify the exact target, install from a writable app location, and record restart/rollback evidence before a 20–30-device rollout. |

## Current session rejection recovery addendum — 2026-08-31

The Gemini client now treats an explicit upstream HTTP 401/403 as a local
session-cache invalidation event. It does not retry the rejected generation,
rotate identities, erase the configured cookie file, or claim that the next
bootstrap will be accepted by Google.

| Current claim | Classification | Evidence | Boundary |
|---|---|---|---|
| A rejected dynamic Gemini session is not reused indefinitely by the local client | VERIFIED_BY_UNIT_TEST | `internal/gemini/auth.go` clears only cached `at`/`bl` values and increments a generation; `internal/gemini/client.go` invokes it for buffered and streaming 401/403 responses; `auth_test.go` covers forced refresh, configured-cookie preservation, and an in-flight bootstrap race, while `client_test.go` covers both response paths and confirms 429 does not invalidate | Google may reject the refreshed session again; cookie reauthentication, expiry detection, account entitlement, and live provider behavior remain external. The previously rejected request is intentionally not replayed. |

## Current Preview 4 publication addendum — 2026-08-31 (historical)

The current public release is the controlled macOS universal prerelease
[`v0.2.0-preview.4`](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.4),
targeting public `main` commit `abfeebaaaaabc740ea29602b602591a0b707fbc2`. Its
five public assets were built from the clean public tip, signed through the
owner-controlled macOS Keychain, freshly downloaded, detached-signature-
verified, and byte-reconciled. The complete evidence is in
[`PREVIEW-4-PUBLICATION-2026-08-31.md`](PREVIEW-4-PUBLICATION-2026-08-31.md).

| Claim at the Preview 4 publication boundary | Classification | Evidence | Boundary |
|---|---|---|---|
| The current public macOS preview is a complete, branded, signed-project artifact set | VERIFIED_LIVE | Preview 4 includes the branded universal `.dmg`/`.zip`, `RELEASE-NOTICE.txt`, `SHA256SUMS`, and detached `SHA256SUMS.sig`; all five public assets passed fresh signature verification and matched the local signed inputs byte-for-byte | The project signature authenticates release bytes, not Apple Developer ID identity; the package remains ad-hoc signed, non-notarized, and macOS-only. |
| Preview 4 is built from the current public source and carries the current updater trust anchor | VERIFIED_LIVE | Release-source verification passed at public-main `abfeeba`; the Keychain-backed signer accepted the private/public pair and the Wails build embedded the checked-in public key | Keychain presence and manifest signing do not prove future key custody, Apple trust, or a live installed-bundle update. |
| The exact Preview 4 artifact launches locally and owns a healthy loopback gateway | VERIFIED_LIVE | Fresh artifact launch bound to `127.0.0.1:8081` when 9610 was occupied, returned 200 from `/healthz`, rendered `v0.2.0-preview.4` in `/playground`, and shut down cleanly | This is one host's package smoke test; it does not prove Google acceptance, provider quota, clean-device replacement, rollback, or fleet rollout. |
| Existing Preview 7 users silently update to Preview 4 | STALE_OR_INCORRECT | The updater is explicit and user-consented; public metadata and project-signature selection are verified, but installed replacement has not been observed on a clean student device | Use **Help → Check for Updates**, verify the exact target, install from a writable app location, and record restart/rollback evidence before a 20–30-device rollout. |

## Post-Preview 4 source state — 2026-08-31 (historical)

Protected PR #62 merged the artifact-preview lifecycle and responsive-header
fix, generated-bundle update, focused tests, and browser evidence; protected PR
#64 then merged the Studio multiline-SSE framing fix and regression coverage
into public `main` at merge commit `cd44b2c`. The latest public package remains
immutable Preview 4, built from `abfeeba`; it contains neither of the
post-publication source fixes. The next source package candidate is explicitly
`v0.2.0-preview.5`, and it has not been tagged or published. No current release
claim should imply that Preview 4 already contains either post-release source
change.

## Preview 5 publication and credential addendum — historical boundary

This section records the Preview 5 boundary and supersedes no later release
state. The immutable Preview 5 package was built from the merged
commit `c28d78736eaae436cc1f1f3b4ec6e0bbcd058b89`; the route-clarity patch and
Preview 5 release reconciliation are public source. After publication,
protected PR #77 and PR #78 advanced public `main` to the intermediate
`ade691db0be9d478cc2feab42e20eb0848f1460e` checkpoint with browser-boundary
evidence and credential-input probe hygiene. PRs #80–#83 subsequently advanced
public `main` to `9f11eef922e09110df923205eb9aad90da35e236`, including
telemetry, release-version, settings, and desktop-coexistence hardening. PR #84
then moved public `main` to `0cc81b2029d5dd467f7c96b26a8b812bee1ab461` with
the release-state documentation and updater-matrix reconciliation. PR #86 then
moved it to `49e0d3b29cffe54642fc9f2d43fc3b9d3aba511d` with the fail-closed
gateway access-key transport guard. The Preview 6 release source target is
`f9b3410e74d7ccc08487dc03788b54a201e12ade`; its package contains these
post-publication follow-ups. Preview 4 and Preview 5 remain immutable
historical inputs, while Preview 6 is the current public macOS prerelease
package. The complete package receipt is in
[`PREVIEW-6-LOCAL-VERIFICATION-2026-08-31.md`](PREVIEW-6-LOCAL-VERIFICATION-2026-08-31.md).

| Claim at the Preview 5 boundary | Classification | Evidence | Boundary |
|---|---|---|---|
| Studio users can distinguish the BOB access credential, the Google Developer API key, the web-session cookies, and the gateway endpoint | VERIFIED_IN_SOURCE | `internal/server/playground.html` presents a credential map, separate labels, password masking, scoped help, and explicit session/provider-route wording; `TestPlaygroundSeparatesGatewayProviderAndWebSessionCredentials` locks the boundary in English and Hindi | The UI cannot tell whether a remote endpoint is trustworthy or whether a student has permission to use a key; the user/operator must make that decision. |
| A rejected Google Developer API key is mislabeled as a missing BOB gateway access key | VERIFIED_BY_UNIT_TEST | The Studio preserves provider-route HTTP 401 messages, scopes gateway-auth classification away from the provider route, and names **BOB Gateway Access Key** in the gateway error; `TestChatDeveloperAPIAuthFailureNamesProviderNotGateway`, `TestPlaygroundBoundsManualRetriesAndLocksRequestControls`, and `TestPlaygroundEducatesAboutExplicitGeminiDeveloperAPIRoute` lock the backend and source behavior | A configured process-level provider key is not visible to the Studio's route toggle; the operator must still identify which route the process is using. |
| A Google Developer API key cannot silently become a BOB gateway access key or silently fall back to the web-RPC route | VERIFIED_BY_UNIT_TEST | `internal/server/gemini_api_test.go` covers provider-only rejection at an API-key-protected gateway, explicit provider routing, configured provider routing, unsupported endpoints, duplicate-key rejection, and credential redaction | Provider project quotas, billing, model availability, and live 401/403/429 behavior remain Google-dependent. |
| The Studio makes the active credential route and its pre-send constraints visible | VERIFIED_BY_UNIT_TEST | `internal/server/playground.html` renders a route-status card with web-session/provider, gateway-auth, cookie-ownership, and model-guard states; `TestPlaygroundSeparatesGatewayProviderAndWebSessionCredentials` and `TestPlaygroundBlocksIncompatibleDeveloperRouteBeforeSend` lock the source contract in the generated UI | A source test does not replace desktop/tablet/phone browser evidence, nor does it establish that a remote endpoint is trustworthy. |
| An incompatible explicit Developer API selection creates a chat turn or sends a request before the user can correct it | VERIFIED_BY_UNIT_TEST | `developerRouteSelectionIssue()` checks key presence, endpoint transport trust, and provider-model/default-think constraints before `isGenerating`, chat-history mutation, or `fetch`; clear actions remove the two secret values from the retained modal DOM | Google may still reject a syntactically valid key, model, quota, or request; those are provider outcomes, not silently converted routes. |
| The public Preview 5 package is a signed, structurally valid universal macOS package from the packaged source baseline | VERIFIED_LOCAL | Fresh candidate from clean checkout `88f2881` produced a branded universal `.dmg`/`.zip`, `x86_64`/`arm64` slices, expected identifier/name, the `/Applications` DMG alias, ad-hoc signature, and a verified detached manifest signature | This is not Apple Developer ID signing/notarization or a clean-device/pilot/rollback acceptance claim. |
| The exact public Preview 5 package bootstraps a healthy local gateway | VERIFIED_LIVE | Fresh `open -n` launch owned `127.0.0.1:8081`, returned `{"status":"ok"}` from `/healthz`, served the credential-map markers, and shut down cleanly | One-host bootstrap does not prove Google generation, Gatekeeper acceptance, rollback, Windows/Linux behavior, or 20–30-device rollout. |
| Preview 5 public assets match the signed local inputs | VERIFIED_LIVE | The five assets from the immutable Preview 5 release were downloaded into a fresh directory, signature/checksums verified, and compared byte-for-byte | This authenticates release bytes with the project key; it does not provide Apple platform trust. |
| An existing writable Preview 1 installation can replace itself with Preview 5 | VERIFIED_LIVE | Native Help → Check for Updates discovered Preview 5; explicit install replaced `/Applications/BOB Gemini Free.app`, restarted on `127.0.0.1:8081`, preserved the visible prior chat response, and About reported `v0.2.0-preview.5`; a second check found no newer release | This is one successful device path; Preview 4/Preview 7 baselines, deliberate rollback, and 20–30-device rollout remain open. |
| Public `main` contains post-Preview 5 source follow-ups | VERIFIED_LIVE | Protected PRs #77–#83 merged the in-app-browser localhost-origin, credential-input, telemetry, release-version, settings, and desktop-coexistence follow-ups; documentation/test reconciliation PR #84 moved public `main` to `0cc81b2029d5dd467f7c96b26a8b812bee1ab461`, and PR #86 moved it to `49e0d3b29cffe54642fc9f2d43fc3b9d3aba511d` with the secure gateway-key transport guard | These changes are not in immutable Preview 5 bytes. A new signed preview is required before students receive them. |
| Studio telemetry can distinguish a live secured gateway from protected aggregate stats without repeated unauthorized polling | VERIFIED_BY_UNIT_TEST | `internal/server/playground.html` checks public `/healthz` and its exposed `X-BOB-Auth-Required` marker before requesting `/`; `TestTelemetryUsesHealthzBeforeProtectedStats` and `TestTrustedBrowserCanReadHealthzAuthMarker` protect the ordering and cross-origin header contract, and a current-main browser smoke observed repeated `/healthz` requests with no `/` 401 loop | Stats remain unavailable without the BOB access key, and remote browser/PNA deployment acceptance remains open. |
| The published `v0.2.0-preview.6` package is signed and structurally valid | VERIFIED_LIVE | A universal macOS package rebuilt from release source target `f9b3410` passed package, ad-hoc bundle, detached-manifest, DMG-layout, full test, race, vet, module, build, and local 1/10/20/30 benchmark gates; the five public assets were then downloaded and verified against the signed input; receipt: `PREVIEW-6-LOCAL-VERIFICATION-2026-08-31.md` | This is project-byte authenticity, not Apple Developer ID signing/notarization or a clean-device/pilot/rollback acceptance claim. |
| A newly built native desktop app cannot silently attach to an older BOB gateway on the configured port | VERIFIED_LOCAL | The current-main signed Preview 6 package launched beside the older installed process on `127.0.0.1:8081`, selected a safe loopback port (`127.0.0.1:53065` in the run), returned `X-BOB-Version: v0.2.0-preview.6`, rendered its own version, and shut down cleanly; `cmd/desktop/gateway_test.go` covers exact reuse and stale-version fallback | This is one macOS host and a local package, not a clean-device, cross-platform, or fleet acceptance result. |
| The credential settings surface keeps BOB access, Google Developer API, web-session cookies, and endpoint identity distinct | VERIFIED_LIVE | Current-source browser smoke showed the credential map, separate masked fields, explicit provider-route toggle, page-memory wording, engine-owned cookie boundary, and Google AI Studio/limits links; source tests cover English/Hindi localization and pre-send route guards | Provider quota/availability, endpoint trust decisions, and student-owned key validity remain external. |

## Current Preview 6 publication addendum — 2026-08-31

This is the current release-state boundary for the repository. Historical
Preview 1–5 sections above remain evidence of earlier states and must not be
read as the current downloadable package.

The published public macOS prerelease is
[`v0.2.0-preview.6`](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.6),
packaged from source target `f9b3410`. Its five public assets were manually
published, re-downloaded, signature-verified, and compared byte-for-byte with
the local signed inputs. The current source advances `PREVIEW_VERSION` to the
unused `v0.2.0-preview.8` so the project does not reuse Preview 6 or the
superseded local Preview 7 candidate for changed bytes. GitHub currently
reports `immutable: false`; write-once release identity
is therefore an operator/project discipline, not a GitHub-enforced lock.

| Current claim | Classification | Evidence | Boundary |
|---|---|---|---|
| Preview 6 is the current public macOS beta | VERIFIED_LIVE | GitHub release `v0.2.0-preview.6` is published as a prerelease with the universal DMG, ZIP, notice, checksum manifest, and detached signature; the five public files were freshly verified | It remains ad-hoc signed and non-notarized; stable, Windows, Linux, clean-device, rollback, provider, and fleet claims remain open. |
| Preview 6 carries the reviewed v0.2 source follow-ups | VERIFIED_LIVE | The local package was built from release source target `f9b3410`, whose browser-boundary, credential-route, telemetry, release-version, settings, gateway-coexistence, and transport-guard changes are in the package receipt | Later `main` documentation/test-only reconciliation is not retroactively part of the published release. |
| The updater can select Preview 6 for legacy Preview 7 and Preview 5 clients | VERIFIED_BY_UNIT_TEST | `TestPublishedPreviewFleetMatrixSelectsPreview6Candidate` uses a mocked official preview listing and asserts Preview 6 selection, manifest availability, and no self-update | The mock does not prove an installed-bundle replacement; one Preview 7/Preview 5 → Preview 6 pilot is still required. |
| The settings surface explains the four credential boundaries | VERIFIED_BY_UNIT_TEST | `TestPlaygroundSeparatesGatewayProviderAndWebSessionCredentials`, the localized dictionary test, and pre-send route guards cover BOB access, Google Developer API, engine-owned cookies, and endpoint identity | Student-owned key validity, Google quota, provider availability, and the safety of a remote endpoint remain external decisions. |
| The current source Preview 8 candidate is locally packaged and signed without being represented as public | VERIFIED_LOCAL | Reviewed checkpoint `309b512a45fc17bb10de712cd110ee9bd809329b` passed the clean-source gate, Wails universal build, Keychain-backed manifest signer, exact-asset verifier, DMG-layout check, isolated `/healthz` smoke, updater transition tests, and current source tests; receipt: `PREVIEW-8-CANDIDATE-VERIFICATION-2026-08-31.md` | `v0.2.0-preview.8` is not published or downloadable; installed Preview 6 → Preview 8 replacement, public-byte reconciliation, rollback, clean-device, Apple trust, provider, and pilot gates remain open. |
| The Gateway settings dialog keeps its interaction controls usable at responsive widths | VERIFIED_LIVE | Current-source browser checks at 390×844, 1024×768, and 1440×900 found no Gateway-modal button or non-checkbox text/password input below 44px, no page-level horizontal overflow, bounded dialog scrolling, and focus return; `TestGatewayModalControlsMeetTouchTargetContract` protects the CSS contract | This is current-bundle browser evidence, not native-WebView, assistive-technology, provider, clean-device, or fleet acceptance. |
| A known protected gateway can leave the Developer API route enabled without its separate BOB access credential | VERIFIED_BY_UNIT_TEST | After `Test Ping` observes the endpoint's `401`, `gatewayAccessSelectionIssue()` feeds the route-status card and blocks the Developer API toggle until the BOB Gateway Access Key is present; `TestCredentialRouteBlocksKnownGatewayAuthRequirement` protects the guard and localized requirement copy | The guard depends on an explicit connection check; it cannot infer remote ownership or trust before the operator tests and saves an endpoint. |
| A stable 0.2.0 student release is ready | UNKNOWN | Local source/package/public-byte gates are green, but no Apple platform trust, clean-device rollback, live provider, Windows/Linux, or staged pilot evidence exists | Do not announce stable or perform a 30-device wave until the remaining acceptance gates are recorded. |
