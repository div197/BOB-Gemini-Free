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
are disabled by default and remain only as an explicit legacy compatibility
option; header credentials should be used for sensitive deployments.

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
Authenticated `GET /v1/metrics` exposes safe aggregate JSON, including a
fixed-cardinality route breakdown for web-RPC versus the explicit Gemini
Developer API path; `/` includes the same aggregate view; `/healthz` remains
minimal. Metrics reset on restart and are never sent externally. Per-request
capability correlation is deliberately not exposed.

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
- dimension normalization for highly compressible images that are small on disk
  but exceed the downstream working-size limit;
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
- bounded browser attachment extraction with a two-job queue, capped text and
  PDF work, cooperative cancellation after removal, and explicit error-state
  styling; OCR worker termination and device CPU evidence remain open.
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
- the direct Gemini Developer API SSE parser now rejects an empty,
  `[DONE]`-only, or metadata-only stream instead of allowing the server to
  fabricate a normal stop, and rejects empty semantic events; standard
  comments and multi-line data fixtures remain supported.
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
- empty web-RPC generation streams and empty Google-shaped non-stream responses
  now fail explicitly instead of fabricating an assistant apology or normal
  stop; Responses and Anthropic stream errors use the same credential-safe
  summaries.
- page-token refresh is context-cancelable, redirect-resistant, bounded to an
  8 MiB response and 4 KiB token values, and preserves the last known-good
  token set after a failed refresh; focused tests cover valid refresh,
  oversized response, and redirect behavior.
- credential-safe image/update-check error envelopes and retry logs; signed
  image URLs, Web RPC tokens, generated diagnostic output, and update-server
  transport details are no longer echoed to API clients or optional retry logs.
- bounded artifact rendering: source is capped at 2 million characters before syntax
  highlighting or iframe construction, render-scoped IDs prevent repeated
  streaming renders from accumulating duplicate registry entries, and the
  in-memory registry is capped at 128 entries/8 million characters with interactive execution
  disabled when capacity cannot be established.
- Mermaid and Pyodide artifact dependency failures now report visibly to the
  host preview, and a failed deep-refiner request leaves the user's original
  prompt unchanged instead of silently applying a different local transformation.
- A provider vision-upload failure no longer launches an automatic OCR-only
  replay. OCR remains available as an explicit user action, preserving the
  requested multimodal semantics and avoiding an unannounced second quota
  request.
- successful non-JSON Developer API bodies are rejected before raw passthrough,
  public error truncation is UTF-8 safe, and metadata-only provider streams
  cannot be reported as normal completions;
- generated artifact pop-outs retain a sandboxed nested iframe, report blocked
  browser popups, and defer object-URL revocation long enough for the browser to
  consume the download;
- the three-stage refiner now bounds the user prompt and every intermediate or
  final stage, checks cancellation between calls, and fails closed on empty
  stages; the UI now labels the actual three-stage pipeline rather than the old
  four-stage description;
- the model catalog is read through lock-protected snapshots in deterministic
  order, while embedded and experimental mobile generation rejects invalid
  explicit models instead of silently substituting a zero-valued route;
- embedded engine lifecycle is explicit: guest-only handlers do not start an
  idle cookie-reload worker, `Engine.Close` is idempotent, and `NewHandler`
  returns a closeable handler for configured session state; partial health,
  metrics, image-upload, and refinement construction paths fail with explicit
  responses instead of panics.
- documented TLS fingerprint names now resolve to their matching profiles;
  unknown names no longer silently turn into a random Safari profile. This is
  a correctness/error-reporting control only, not a method for avoiding Google
  rate limits, WAF decisions, or shared-egress detection.
- unenforced synthetic rate-limit headers were removed; malformed discovered
  configuration now stops with a visible error instead of silently reverting to
  defaults; status, CDP, and updater transaction metadata reads are bounded;
  and updater plan paths are constrained to updater-owned names.
- browser fallback and Wails bootstrap failures now return or display an error
  instead of logging a successful open when no browser or gateway is available;
- all three preview desktop Make targets now forward the checked-in public
  updater key to their packagers, and each preview packager rejects a stable
  version or non-preview channel before invoking Wails;
- the release-source preflight now validates the explicit canonical updater-key
  block against both standalone installers and the native/package entry points,
  and rejects Docker or preview-version drift before packaging;
- the 15-check diagnostic runner now fails closed on malformed/empty response
  objects, incomplete Anthropic lifecycles, oversized streams, and unavailable
  provider-dependent image generation rather than reporting a transport-only
  success. On the current no-cookie audit gateway, a real run passed 13 checks
  and failed strict web-route JSON output plus image generation (HTTP 502); this
  is evidence of the capability boundary, not a reason to weaken the check.
  nil embedded Developer API clients fail closed.

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

## Latest local hardening addendum — 2026-08-29

This addendum supersedes the host/toolchain and Git-state wording of the dated
historical snapshot above for the current working tree. The current audit host
is macOS `darwin/arm64`, Go `go1.26.6`, with `CGO_ENABLED=1`. The reviewed
branch is `codex/release-readiness-v0.2.0`; its relationship to `main` and the
public release is recorded separately from source tests and must not be
inferred from the branch name.

The latest local slice added and tested:

- bounded status, desktop health-probe, CDP version, updater-plan, and
  updater-confirmation reads;
- strict updater-owned staging/plan/rollback/confirmation/helper path names;
- visible Wails gateway/configuration startup errors and truthful system-browser
  fallback errors;
- fail-closed nil `geminiapi.Client` behavior;
- removal of unenforced synthetic request-rate headers; and
- semantic validation for native Google-shaped JSON/SSE generation so
  metadata-only or finish-only provider responses cannot appear successful;
- removal of the benchmark's invented token-count fallback, with token
  throughput reported only when successful responses provide usage; and
- the generated `web/index.html` synchronization after the UI changes.

The following local commands passed after this slice: `make build`,
`make desktop-key-check`, `make web`, `go test -count=1 ./...`,
`go test -race -count=1 ./...`, `go vet ./...`, `go build ./...`,
`go mod verify`, and `git diff --check`. Six CGO-disabled CLI targets also
cross-compiled into an isolated temporary directory for Darwin arm64/amd64,
Linux arm64/amd64, and Windows arm64/amd64; their Mach-O, ELF, and PE file
types were inspected. Inline JavaScript parsing passed for the source bundle,
generated web bundle, and Wails module. Release-directory publication,
platform-native packaging, and real installed-bundle checks remain separate
gates. No GitHub Actions were added or run, and no provider/API/release
private key was used.

The subsequent clean-commit preview packaging check also passed on macOS:
`make desktop-preview-mac` forwarded the checked-in public updater key,
produced a universal arm64/x86_64 `.app`, ZIP, and drag-to-Applications DMG,
and passed checksum, bundle-signature, PWA-route, and native GUI
Quit/gateway-shutdown checks. The preview packager now rejects a stable
version or non-preview channel before invoking Wails. This is local package
evidence only; the manifest was not signed or uploaded, and no Apple
Developer ID/notarization, clean-device update, Windows, Linux, or provider
acceptance claim follows from it.

### Release-coherence follow-up — 2026-08-29

Protected PR [#33](https://github.com/div197/BOB-Gemini-Free/pull/33) merged the
release-source preflight into `main` at `2f5d498`. The gate now extracts the
canonical Ed25519 public key from its explicit document block and compares it
with the standalone installers, native package inputs, Docker version, and
preview version/channel defaults. A controlled installer-key mismatch failed
closed, while stable and preview fixtures passed. This closes source-input
drift as a locally testable failure path; it does not establish private-key
custody, a signed public release, Apple trust, or installed-device update
acceptance.

### Release trust-anchor encoding follow-up — 2026-08-30

The release-source gate now also derives the Ed25519 SPKI value used by
`install.sh` from the canonical raw public key and rejects a mismatch. This
closes a packaging-input gap where the installer’s hexadecimal key could match
while its separately embedded SPKI Base64 value—the value consumed by
OpenSSL—could drift. Stable and preview source fixtures passed, and a
controlled SPKI tamper fixture failed closed. This still proves only local
source coherence; it does not sign, publish, or verify public release assets.

### Studio credential persistence follow-up — 2026-08-30

Protected PR [#38](https://github.com/div197/BOB-Gemini-Free/pull/38) merged the
Web Studio gateway-auth persistence fix into `main` at `f3a0a8c`. The optional
BOB gateway-auth token now remains in page memory, the legacy `bob_api_key`
browser-storage entry is purged on load, and health/model/generation requests
use only the current session value. Focused, full, race, vet, build, module,
and release-source checks passed. This reduces credential retention on shared
student devices; it does not make same-page scripts trustworthy, and it does
not close provider, platform-trust, signed-artifact, clean-device, or fleet
acceptance gates.

### CLI updater asset-selection follow-up — 2026-08-31

Protected PR [#40](https://github.com/div197/BOB-Gemini-Free/pull/40) merged the
CLI updater filename hardening into `main` at `a80f08d`. The selector now
requires the canonical platform asset name and rejects a lookalike that only
matches the old suffix-based rule. Focused, full, race, vet, module, build, and
release-source checks passed. This is a defense-in-depth source guarantee; it
does not establish public release completeness or clean-device updater proof.

### Studio Markdown-link protocol follow-up — 2026-08-31

The Studio previously described its Markdown links as allow-listed but only
blocked three dangerous-looking string prefixes. That left the renderer's
security decision dependent on downstream sanitization details. A shared
protocol allow-list now accepts only `http:`, `https:`, `mailto:`, and `tel:`
(plus same-document hash links), converts unsupported or malformed targets to
`#`, and is reused by the native/hosted external-link bridge. The focused
source regression rejects the old blacklist markers and requires the shared
policy in both rendering and navigation paths. This does not change Gemini
wire behavior or make generated content trusted; artifact HTML remains
sandboxed and browser acceptance is still required.

### Studio artifact-action follow-up — 2026-08-31

The interactive artifact card had both a click-only container action and a
nested launch button. That was ambiguous for keyboard users and made the
whole card look actionable even though the real operation is the launch
control. The container is now a named group with neutral hover treatment, and
only an explicit `type="button"` launch button carries `data-action`. The
artifact registry, sandbox, editor hydration, and preview lifecycle are
unchanged; this is a focused interaction/accessibility correction protected by
`TestArtifactLaunchChipUsesOneKeyboardAction`.

### Studio Developer API route-state follow-up — 2026-08-31

The optional student-owned Gemini Developer API toggle could previously remain
checked after an empty key was submitted, or after an active key was cleared;
the next generation then failed instead of the UI reflecting the route change.
The route now validates key presence before enabling, resets the checkbox and
focuses the key field when absent, and explicitly returns to the default
web-session route when an active key is cleared. This is a client-state safety
fix only: it does not rotate keys, bypass provider limits, or change the
server-side header boundary. `TestDeveloperAPIRouteToggleFailsClosedWithoutKey`
protects the ordering and visible reset behavior.

### Studio provider-key transport follow-up — 2026-08-31

Native Wails context was previously treated as sufficient trust for sending a
student-owned Developer API key, even when the configured gateway endpoint had
been changed to a remote plain-HTTP address. The client now permits HTTP only
for loopback endpoints; remote endpoints must be explicitly saved and use
HTTPS, and endpoints containing credentials or no hostname are rejected. This
does not prevent the default local gateway or intentionally configured HTTPS
deployments, and it does not change the server-side provider-key header
contract. `TestDeveloperAPIRouteRequiresSafeGatewayTransport` protects the
decision boundary; rendered native/hosted behavior remains a browser/device
acceptance gate.

### Studio gateway-status follow-up — 2026-08-31

The generation catch path previously set the local gateway indicator to
offline for every non-abort error, including HTTP 401/502 responses and
provider or stream failures that arrived after the gateway had already
responded. The client now records response receipt, marks the gateway online
at that point, and only marks it offline when no response was received. The
visible provider/session error remains intact while local connectivity state
stays truthful. `TestReachableGatewayIsNotShownOfflineAfterHTTPOrStreamFailure`
protects this distinction; live browser/provider behavior remains external.

### Studio attachment-control follow-up — 2026-08-31

Attachment preview and removal controls previously interpolated persisted
attachment IDs into inline `onclick` JavaScript. The shelf now writes the
escaped ID into `data-file-id` and routes both actions through the existing
delegated event handler; the controls have explicit button types and
accessible names. This preserves the same preview/remove operations while
removing an avoidable local-history injection boundary.
`TestAttachmentControlsDoNotEmbedUntrustedIDsInInlineJavaScript` protects the
construction and rejects the old handlers.

### Studio attachment-image follow-up — 2026-08-31

The same history/rendering path also allowed persisted attachment icons to be
inserted as raw HTML and made image thumbnails pointer-only inline handlers.
Attachment icons are now escaped before HTML insertion. Image previews are
keyboard/touch buttons backed by a delegated action, accept only bounded
base64 raster data URLs, use `noopener,noreferrer` for the preview window, and
show an explicit unavailable state for unsupported formats. This preserves
normal PNG/JPEG/WebP-style previews while avoiding active SVG/data navigation
from crafted local history. `TestPersistedAttachmentIconsAreEscapedBeforeHistoryHTML`
and `TestAttachmentImagePreviewsUseAccessibleRasterOnlyControls` protect the
boundary.

### Studio gateway-recovery follow-up — 2026-08-31

Generation error copy used `javascript:void(0)` anchors for the Config
recovery action. Those affordances are now real named buttons routed through
the existing delegated action handler, preserving keyboard behavior and
keeping failure UI free of executable URL schemes. `TestErrorRecoveryConfigActionsAvoidJavaScriptURLs`
protects the construction.

### Optional logging boundary follow-up — 2026-08-31

The upstream Gemini client had the same partial-construction hazard that was
previously found in the server: a retry after a pre-request transport failure
called an optional logger directly. The new `Client.logf` helper makes both
buffered and streaming retry logging nil-safe. Deterministic regressions cover
the final error and a successful stream recovery with no logger. This change
does not alter retry classification, delay, payload construction, session
routing, or stream deduplication.

The current source-hardening code tip in this historical report was on the local
branch `codex/release-readiness-v0.2.0` at `cec4c8e`; the reviewed code was then
published in public `main` through PR #42 at `ba1b562`. At the time of this
report no public Preview 2 or stable release, tag, or GitHub Actions workflow
had been created by that continuation. The subsequent controlled Preview 2
publication is recorded in [`PREVIEW-2-PUBLICATION-2026-08-31.md`](PREVIEW-2-PUBLICATION-2026-08-31.md);
the browser viewport acceptance matrix and clean-device/provider gates remain
open.
