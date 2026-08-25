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

The coverage run is not a project-wide aggregate quality score. The latest
package-local results were: root 0.6%, `cmd/desktop` 55.7%,
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
| OpenAI Chat Completions streaming is “FULL” | STALE_OR_INCORRECT | `internal/server/chat.go:74-208` streams only when no tools or `tool_choice=none`; otherwise it buffers through the non-stream path at `chat.go:211-316` | Ordinary no-tool streaming is implemented; full OpenAI behavior, tool streaming, error semantics, and upstream compatibility remain partial/unknown. |
| OpenAI Responses API is “FULL” | STALE_OR_INCORRECT | `internal/server/responses.go:15-121,248-343` implements selected input and an eight-event-looking lifecycle | The source emulates a selected Responses shape. Tool streaming is explicitly buffered/replayed at `responses.go:248-326`; broad Responses API parity is not proven. |
| OpenAI system/developer roles are natively preserved | STALE_OR_INCORRECT | `internal/format/openai.go:139-142` turns them into `[System instruction]: ...`; unit assertions are in `internal/format/openai_test.go:48-66` | Support level: EMULATED_RELIABLY for the tested prompt transformation, not native Gemini/OpenAI role semantics. |
| OpenAI structured outputs are enforced | STALE_OR_INCORRECT | `internal/format/openai.go:172-174` appends a textual JSON instruction | Support level: EMULATED_PARTIALLY. No schema validator or response conformance check is present. |
| OpenAI reasoning output is supported as `reasoning_content` | VERIFIED_BY_UNIT_TEST | `internal/format/thinking.go:32-200`, `internal/server/chat.go:79-144`, `internal/format/openai_test.go:213-235`, `internal/format/thinking_test.go:7-152` | The local splitter and serialization path are tested. Whether Google emits the expected fenced markers live is UNKNOWN. |
| Anthropic Messages has a complete compatible SSE lifecycle | STALE_OR_INCORRECT | `internal/server/anthropic.go:116-280` emits the lifecycle, but tests do not assert the complete event ordering/payload contract against a fixture | Selected lifecycle support is implemented; full Anthropic/Claude Code compatibility is not established. |
| Anthropic extended thinking is full native support | STALE_OR_INCORRECT | `internal/server/anthropic.go:55-67` maps budgets to Gemini think integers; `internal/format/anthropic.go:147-169` maps requests | Support level: EMULATED_PARTIALLY. Budget semantics are reduced to a small internal mode set. |
| Anthropic prompt caching counters are real | STALE_OR_INCORRECT | `internal/server/anthropic.go:132-136,330-335` returns both cache counters as `0` | Support level: EMULATED. No cache accounting exists. |
| Anthropic tool use/results are native | STALE_OR_INCORRECT | `internal/format/anthropic.go:92-120,172-222` converts blocks to OpenAI-shaped data; `internal/format/openai.go:74-80,184-232` injects/parses Markdown blocks | Support level: EMULATED_PARTIALLY; no native Google function-call request is sent. |
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
| Remote image fetching has a minimum SSRF boundary | VERIFIED_BY_UNIT_TEST | `internal/multimodal/upload.go:129-190`; tests reject loopback/private IPs, private DNS, nonstandard ports, and cross-host redirects | This is a defense-in-depth application check, not proof against every proxy/DNS-rebinding or custom HTTP transport topology. |
| Images are compressed to 1024 px JPEG quality 75 and below 1 MB | STALE_OR_INCORRECT | `internal/multimodal/compress.go:17-58` caps dimensions and uses quality 75 when recompressing | The code does not loop or reject if the JPEG remains over the requested byte limit. The `<1MB` claim is not guaranteed. |
| Scotty upload uses a two-step resumable protocol | VERIFIED_IN_SOURCE | `internal/multimodal/upload.go:41-126` | Upload URL/header parsing and file-reference shape are source-backed; no controlled authenticated live upload was run. |
| Image deduplication prevents duplicate uploads | VERIFIED_IN_SOURCE | `internal/server/server.go:23-27`; `internal/server/helpers.go:150-170` | Sequential cache hits are implemented. Concurrent identical misses can race because there is no per-hash single-flight/lock; the stronger claim is UNKNOWN. |
| Google inline image data is faithfully accepted | VERIFIED_IN_SOURCE | `internal/format/google.go:103-120` decodes base64 and creates upload inputs | Decode errors are silently ignored and no upstream vision fixture exists. Full fidelity is UNKNOWN. |

### Security, trust, and operational boundaries

| Claim | Classification | Current evidence | Truth boundary / remaining proof |
|---|---|---|---|
| Optional API keys use constant-time comparison | VERIFIED_BY_UNIT_TEST | `internal/server/middleware.go:34-68`; `internal/server/server_test.go:71-128` | Bearer, `x-api-key`, `x-goog-api-key`, and query `?key=` are accepted. Query keys create URL/log/referrer exposure risk. |
| CORS intentionally supports remote Web Studio and Private Network Access | VERIFIED_BY_INTEGRATION_TEST | `internal/server/middleware.go`; `internal/server/security_boundary_test.go` covers loopback and explicitly configured remote origins | Browser origins are reflected only when trusted; no-origin native clients remain supported. PNA is transport permission, not authentication. |
| An arbitrary web origin can cause privileged no-key gateway requests | VERIFIED_BY_INTEGRATION_TEST | Baseline wildcard behavior was reproduced before the fix; the current security test rejects an untrusted `Origin` with 403 | Browser exploitability still depends on browser/PNA behavior, but the gateway HTTP trust decision is now explicit and test-locked. |
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
| Updater tests cover download verification and replacement | VERIFIED_BY_INTEGRATION_TEST | `internal/updater/golden_test.go` uses an `httptest` release server and isolated temp candidates | Network transport and GitHub release behavior remain unverified live. |
| Native desktop deterministically starts the gateway and reports the actual port | VERIFIED_BY_INTEGRATION_TEST | `cmd/desktop/gateway.go` and tests cover occupied-port fallback, compatible reuse, and non-BOB rejection; the branded bundle was built and smoke-tested on this Mac, reaching `/playground` with the actual loopback endpoint | Live Google behavior, signed/notarized distribution, and occupied-port GUI fallback still need release-device acceptance. |
| `cmd/desktop` is the sole active native desktop architecture | VERIFIED_IN_SOURCE | `cmd/desktop`, `docs/engineering/DESKTOP-ARCHITECTURE.md`, and the archival commit remove the alternate wrapper from the active tree; Git history retains recovery provenance | Platform publisher trust and clean-machine acceptance remain external release gates. |
| The CLI has a duplicate browser fallback bug | VERIFIED_BY_INTEGRATION_TEST | `main.go` now routes startup through one `launchStudioOrFallback` call; `main_test.go` asserts a single fallback invocation | The fix is intentionally narrow and does not refactor unrelated CLI startup. |

### Performance, diagnostics, and test confidence

| Claim | Classification | Current evidence | Truth boundary / remaining proof |
|---|---|---|---|
| The repository has a benchmark runner reporting P50/P90/P99 and throughput | VERIFIED_IN_SOURCE | `internal/diag/bench.go:15-161`; `main.go:127-157`; `internal/diag/bench_test.go:10-36` | The runner exists and works against a local fake HTTP server. It does not establish provider-independent gateway overhead. |
| The README sample values (P50 ~1.72 s, proxy overhead <1.5 ms, RAM <15 MB, 100+ streams under 50 MB) are measured release baselines | STALE_OR_INCORRECT | The old wording was removed from the active benchmark section; the dossier and historical text had no reproducible artifact | No Google/live release performance number is established. |
| The benchmark supports requested local-only profiles and RSS/allocation metrics | MEASURED | `internal/diag/local_benchmark.go`, `cmd/benchmark-local`, and `docs/engineering/LOCAL-BENCHMARK-2026-08-21.md` record 1/10/50/100 profiles with P50/P90/P95/P99, allocations, RSS, goroutines, connections, and throughput | This is a mocked local gateway benchmark; live upstream profiles remain optional and unrun. |
| The core has a deterministic golden fixture suite | VERIFIED_BY_UNIT_TEST | `docs/engineering/CORE-REGRESSION-HARNESS.md`; `internal/gemini/golden_test.go`; `internal/format/golden_test.go`; `internal/multimodal/golden_test.go` | These are synthetic protocol-shaped fixtures, not live Google captures. Live acceptance remains unknown. |
| The default Go suite is hermetic and provider-independent | VERIFIED_IN_SOURCE | Upstream-facing diagnostic workflows are tagged `live`; the default suite excludes them, while `go test -tags=live ./internal/diag` is documented separately | A live-tagged run remains provider/session-dependent and is intentionally not part of the default validation gate. |
| Current package coverage is at least 80% everywhere | STALE_OR_INCORRECT | `go test -cover ./...` measured 6.9% browser, 34.2% gemini, 46.0% multimodal, 33.4% server, 0.0% service, and 17.6% updater among others | The local package numbers directly contradict a blanket 80% claim. Coverage is not the only quality metric, but the claim is not current. |
| The existing suite proves live Google compatibility | DOCUMENTED_ONLY | The supplied dossier says tests and vet passed; current tests pass locally, but no controlled live acceptance run was performed and no live gateway was listening | Requires separately authorized provider/session verification. |

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
| A public native desktop preview is available for controlled evaluation | VERIFIED_LIVE | Manually published prerelease `v0.1.7-preview.7` contains the corrected branded macOS universal `.dmg`/`.zip`, `RELEASE-NOTICE.txt`, `SHA256SUMS`, and detached `SHA256SUMS.sig` assets; the downloaded assets were byte-compared with the local signed candidate and re-verified | This is a beta preview only: no Apple notarization, Windows publisher signature, Linux asset, clean-device matrix, silent updater, or production student-readiness claim. Existing Preview 6 installations require one manual migration because the original project signing key was not recoverable. |
| A free branded macOS preview package can be created without Apple membership | VERIFIED_BY_INTEGRATION_TEST | `scripts/package-wails-preview.sh` creates a branded `BOB Gemini Free.app`, `.zip`, `.dmg`, release notice, and checksums without Developer ID credentials; the bundle metadata uses the `com.abcsteps` identity | This proves local packaging only; it does not establish Gatekeeper trust, notarization, clean-device acceptance, or public student readiness. |
| The native desktop app exposes an explicit update check | VERIFIED_BY_UNIT_TEST | `internal/updater/desktop.go` selects only official stable or `preview.N` channels, with a bounded preview listing; `internal/updater/updater_test.go` covers branded/legacy names, prerelease ordering, signed-manifest discovery, and the endpoint bound | Preview 7 can offer a consented verified update after migration; it does not silently install, remove the macOS warning, or replace platform publisher trust. |
