# BOB Gemini Free — 100-Path Failure Register

**Audit date:** 2026-08-29 (Asia/Kolkata)
**Scope:** current working tree, Git history, local served gateway, native
desktop/update code, generated web bundle, and the previously supplied Phase-II
audit handoff.
**Operating constraint:** local verification only; no GitHub Actions and no
provider or release credentials were used.

This is a risk register, not a claim that every row is fixed. `PROTECTED` means
the current source has a deterministic guard and a focused regression check.
`PARTIAL` means a guard exists but a meaningful boundary or platform proof is
still missing. `OPEN` means the failure is still possible or has not yet been
regression-locked. `EXTERNAL` means the source cannot prove it without an
operator, device, release, or Google-provider action.

## Immediate decision

The project is safer to evolve than the earlier Preview 7 snapshot, but it is
not yet a universally verified, unattended, cross-platform student release.
The largest remaining gates are clean release provenance, native artifacts for
every named platform, updater acceptance from the actual installed bundle,
bounded browser/client state, and live provider/session behavior. A successful
local build or a working one-Mac preview does not close those gates.

The GitHub credential and provider keys pasted in the conversation are treated
as compromised input. They were not copied into this register, source, Git, or
the release process. Revoke/rotate them before any publication or classroom
use; do not paste replacement secrets into chat.

## 1. Release, reproducibility, and provenance

| # | Failure path | Status | Proactive step now |
|---:|---|---|---|
| 1 | A dirty worktree produces an artifact containing unreviewed local changes. | PROTECTED | `scripts/verify-release-source.sh` fails stable CLI/desktop packaging on any modified or untracked path; the final release still requires a reviewed clean commit. |
| 2 | Untracked API, desktop, PWA, or test files are omitted from the release commit. | PROTECTED | The same verifier requires a Git worktree with no untracked files before release packaging; staged-file review remains an operator gate. |
| 3 | The working branch, `main`, and `origin/main` do not represent the same source. | OPEN | Reconcile branch ancestry and publish only from the reviewed commit; never infer state from a local branch name. |
| 4 | `internal/server/playground.html` and generated `web/index.html` diverge. | PROTECTED | `scripts/verify-release-source.sh` compares the rendered source to the checked-in bundle without rewriting either file; `make web` remains the explicit generation step. |
| 5 | Version, channel, app metadata, Docker metadata, and release assets disagree. | PARTIAL | Keep one immutable version/channel matrix and reject a package when any field differs. |
| 6 | The package embeds a different updater public key from the manifest signer. | PARTIAL | Run a local public/private fingerprint match check before signing and re-verify the embedded public key. |
| 7 | A local install requests `/manifest.json` or `/sw.js` and receives 404. | PROTECTED | Local routes, embedded assets, content types, version injection, and endpoint tests now cover the contract. |
| 8 | An old service worker serves stale HTML or caches API responses. | PROTECTED | The worker uses versioned local caches, network-first static fetches, and excludes `/v1/` and `/v1beta/`; test a real browser activation before release. |
| 9 | Preview/stable channel selection silently downgrades or hides a failure. | PROTECTED | Bounded channel tests cover stable-first migration, preview ordering, and fail-closed stable errors. |
| 10 | Existing Preview 6/7 devices cannot verify a release because their embedded key differs or predates migration logic. | EXTERNAL | Pilot the documented same-key bridge or perform one manual migration; do not promise a direct fleet jump. |
| 11 | `install.sh` downloads an unsigned CLI binary and users mistake it for a trusted native package. | PARTIAL | The default path downloads only HTTPS release assets, verifies the detached Ed25519 manifest and exact asset digest, installs atomically, and has no unsigned fallback; macOS stock LibreSSL and other hosts without an Ed25519 verifier fail closed, and clean-host proof remains. |
| 12 | `install.ps1` downloads an unsigned CLI binary or executes an unreviewed remote script. | PARTIAL | The reviewed local script uses HTTPS-only `curl.exe`, verifies the detached Ed25519 manifest and exact architecture asset, and has no unsigned/source fallback by default; clean Windows PowerShell/curl/OpenSSL/.NET acceptance remains. |
| 13 | A release names an architecture whose binary or desktop package was not built. | PARTIAL | Generate the asset matrix from actual build outputs and omit unsupported platforms from release notes. |
| 14 | A desktop release omits `SHA256SUMS` or `SHA256SUMS.sig`. | PROTECTED | The desktop updater refuses absent signed-manifest material; publication still must include the exact files. |
| 15 | README/changelog advertises “full,” “native,” “unlimited,” or “zero dependency” behavior not proven by source/tests. | PARTIAL | Keep claim reconciliation in `VERIFICATION-MATRIX.md` and review user-facing copy before every release. |
| 16 | A build is made from uncommitted source or an unrecorded toolchain. | PROTECTED | `verify-release-source.sh` blocks dirty or untracked release packaging; `record-release-evidence.sh` re-runs source and asset checks and records commit, branch, Go version, host, UTC time, manifest/signature hashes, and asset hashes outside the release tree. |
| 17 | macOS quarantine, permissions, App Translocation, or Gatekeeper blocks a package despite valid project checksums. | EXTERNAL | Test the mounted DMG and copied `/Applications` bundle on a clean Mac; do not ask users to disable Gatekeeper. |
| 18 | A binary cannot be tied to an immutable source commit and exact manifest bytes. | PROTECTED | The local release receipt records the reviewed commit/toolchain/host and exact manifest, signature, and asset hashes; publication/tag reconciliation is still an operator gate. |
| 19 | The signed manifest is syntactically valid but does not contain every published asset. | PARTIAL | `scripts/verify-release-assets.sh` now verifies the detached signature and one-to-one local directory/manifest mapping; rerun it after downloading the public release because upload state remains external. |
| 20 | Exposed GitHub/provider credentials are reused, logged, or committed. | EXTERNAL | Revoke/rotate the previously pasted credentials immediately; keep all replacement secrets outside the repository and chat. |

## 2. Updater and package replacement

| # | Failure path | Status | Proactive step now |
|---:|---|---|---|
| 21 | The app runs from a read-only DMG or translocation directory and staging fails. | PROTECTED | Same-filesystem staging returns an actionable error directing the user to a writable application directory. |
| 22 | `/Applications` is not writable for the current user or managed Mac. | PARTIAL | Surface the exact staging error and document a per-user writable install path; test standard and managed permissions. |
| 23 | Two update requests stage or replace the same app concurrently. | PROTECTED | `internal/updater/desktop_helper.go` takes an atomic per-install lock around replacement, recovers only old empty locks, and refuses active/non-empty conflicts; updater tests cover serialization and stale-lock safety. |
| 24 | Power loss occurs between backup, rename, and helper restart. | PARTIAL | Keep same-filesystem backup/rollback state and test interrupted replacement with a fixture package, not the running executable. |
| 25 | A failed update leaves orphaned staging directories or sensitive metadata. | PARTIAL | Committed, matching plans older than 24 hours are removed conservatively when staging begins; pre-plan crashes remain for manual inspection and diagnostics stay mode 0600. |
| 26 | A real package fails after replacement and rollback is never confirmed on a clean device. | EXTERNAL | Run a deliberately invalid candidate through the real installed-bundle path and verify cookies/config/history survive. |
| 27 | The install target changes after validation but before replacement. | PROTECTED | Replacement takes the per-install lock and revalidates both the install target and candidate immediately before rename; symlink/non-directory targets are rejected. |
| 28 | A ZIP contains traversal, symlink-like, special-mode, duplicate, or unexpected bundle entries. | PROTECTED | Extraction rejects ambiguous path components, symlinks, special files, duplicate normalized paths, multiple roots, unexpected roots, and bounded over-expansion; focused tests cover duplicates/special entries. |
| 29 | A fixed minimum asset size rejects a legitimate small binary or accepts a misleading large response. | PROTECTED | The old fixed 5 MiB floor is removed; exact release metadata size, upper bounds, and digest verification are tested. |
| 30 | A server stalls while sending metadata or an artifact. | PROTECTED | Metadata/artifact clients have total timeouts and bounded readers; add a slow-body integration case if the helper evolves. |
| 31 | A redirect moves an update request to an attacker-controlled host. | PROTECTED | Redirect policy permits only the official repository/API/release CDN hosts and has rejection tests. |
| 32 | An ad-hoc macOS signature is mistaken for Apple Developer ID/notarization. | EXTERNAL | Keep preview wording explicit; obtain Developer ID, hardened runtime, notarization, stapling, and clean-device evidence for production. |
| 33 | An unsigned Windows binary triggers SmartScreen or is replaced without publisher trust. | EXTERNAL | Use Authenticode/publisher signing and a clean Windows acceptance run before naming Windows production-ready. |
| 34 | A Linux package depends on unavailable desktop libraries or assumes an unsupported display environment. | EXTERNAL | Build and launch on each named Linux target; record dependencies and a terminal recovery path. |
| 35 | “Auto-update” is interpreted as silent fleet push or unattended installation. | PROTECTED | Product docs define a user-consented check/download/install flow; keep that wording in UI and release notes. |
| 36 | Recovery depends on a weak, mutable environment key or an untrusted downloaded configuration. | PROTECTED | The desktop path uses the embedded public key and refuses the mutable environment fallback for installation. |
| 37 | The helper restarts the wrong executable, loses arguments, or exits before health confirmation. | PARTIAL | Exercise helper restart/rollback with a packaged fixture and record the process/endpoint transition. |
| 38 | Prerelease comparison treats `preview.10` as older than `preview.9`, or stable is downgraded. | PROTECTED | Version ordering and channel tests cover numeric previews and stable/preview directionality. |
| 39 | Legacy bundle names are not recognized, causing a false “no update” result. | PROTECTED | Branded and legacy asset-name tests cover the migration candidates. |
| 40 | Uploaded DMG, ZIP, manifest, signature, release note, and README describe different bytes or versions. | PARTIAL | The local signer, one-to-one verifier, and release receipt catch byte/manifest drift before and after download; the public GitHub asset set and release-page prose still require a fresh-directory operator reconciliation. |

## 3. Google provider, sessions, and quota boundaries

| # | Failure path | Status | Proactive step now |
|---:|---|---|---|
| 41 | Anonymous/guest web-RPC access is blocked or changed by Google. | EXTERNAL | Treat guest generation as upstream-dependent; expose an actionable per-user session/API-key route without claiming guaranteed access. |
| 42 | A cookie expires, is revoked, or loses required tokens during a long session. | PARTIAL | Secure file reload and explicit upstream error mapping exist; add capability-aware reauthentication guidance and live expiry evidence. |
| 43 | Cookie/session bootstrap follows an unbounded HTML/token response. | PROTECTED | Bootstrap reads have bounded size and status checks; retain a fixture for oversized and malformed pages. |
| 44 | Same-size, same-mtime cookie changes are missed by an mtime-only cache. | PROTECTED | The cache includes file size and content hash and clears unreadable authenticated state. |
| 45 | Cookie files are readable by other local users or are symlinks to another file. | PROTECTED | Secure reads reject broad POSIX permissions, symlinks, nonregular files, and oversized content; writes use 0700/0600. |
| 46 | A deleted pool file remains active because reload preserves stale sessions. | PROTECTED | Reload replaces the configured snapshot and drops deleted/invalid sources; deletion tests cover the behavior. |
| 47 | The same cookie appears through multiple pool sources and receives duplicate scheduling. | PROTECTED | Session IDs are hashed and source/file inputs are deduplicated during load/reload. |
| 48 | All sessions are cooling down, but selection bypasses cooldown and hammers Google again. | PROTECTED | Healthy selection returns no session while all are in cooldown; a request test proves no upstream call occurs. |
| 49 | A session is healthy for text but invalid for vision, upload, or a particular model. | PARTIAL | Generic session cooldown and explicit route separation prevent silent provider changes; capability-specific vision/upload/model health is intentionally not guessed and requires consented live probes. |
| 50 | Gemini Developer API limits are treated as per-key or universal instead of per-project/model/tier. | EXTERNAL | Link dated Google rate-limit/pricing sources and tell students to inspect their own AI Studio project limits. |
| 51 | “Free” or “unlimited” is treated as a permanent entitlement. | EXTERNAL | Keep quota/capacity claims upstream-dependent; never encode a fixed promise in routing or UI. |
| 52 | Thirty devices create a simultaneous release-check or generation burst. | PARTIAL | Stagger classroom rollout and add local backoff/visibility; never disguise the school egress with rotating proxies. |
| 53 | Anonymous, cookie, and Developer API paths mix state or silently change provider mid-request. | PARTIAL | Current explicit route separation is tested; add a per-request route/capability record that never contains secrets. |
| 54 | Student API keys persist in localStorage, chat history, logs, or crash output. | PROTECTED | The key is page-memory only, excluded from config JSON, and sent only in a header when explicitly enabled; re-audit before release. |
| 55 | A provider key leaks into a URL, query string, or request body. | PROTECTED | Direct-client tests assert the documented `x-goog-api-key` header and absence from URL/body. |
| 56 | Gateway API-key auth and Google provider-key routing are confused by the UI or header names. | PARTIAL | Keep separate labels, headers, docs, and tests; add a negative matrix for each credential type on each endpoint. |
| 57 | Model catalog entries imply the upstream model identity or availability is guaranteed. | PARTIAL | Call entries aliases/routing targets and defer current availability to Google; add live model verification only when available. |
| 58 | A BOB alias is advertised as an OpenAI/Anthropic model rather than a Google route. | PARTIAL | Keep adapter docs explicit that aliases are local mappings and not provider identity assertions. |
| 59 | Typed/multipart content is silently dropped or flattened incorrectly by an adapter. | PROTECTED | Anthropic, OpenAI, and direct Developer API adapters now fail closed on unsupported blocks, wrong field types, missing text, invalid base64, and malformed tool results; provider semantics remain a separate evidence gate. |
| 60 | A data URL exceeds decoding, image-dimension, or request limits and causes memory or upstream failure. | PROTECTED | Gateway and direct-API paths bound base64, decoded bytes, dimensions, and pixels; keep client-side size checks aligned. |

## 4. Streaming, retry, and protocol state

| # | Failure path | Status | Proactive step now |
|---:|---|---|---|
| 61 | A connection sends headers but never completes the response. | PROTECTED | HTTP total timeout, response-size limits, and stream line limits are enforced and tested. |
| 62 | A POST is retried after ambiguous partial upstream processing, duplicating a request. | PROTECTED | Generation retries are limited to transport failures identified as pre-connection `dial`/`connect`/`lookup`; HTTP 5xx, response-read failures, parser failures, and any partial stream are surfaced without replay. Tests cover pre-request retry and ambiguous-failure no-retry behavior. |
| 63 | Fixed retry bursts amplify a Google 429/5xx incident. | PROTECTED | HTTP 429, redirects, and Bard policy frames are not retried; transient transport/5xx retries use bounded config values, capped exponential jitter, and a bounded `Retry-After` delay. |
| 64 | Large JSON request bodies exhaust memory before format conversion. | PROTECTED | `readRequestBody` applies a second 32 MiB limit at every JSON handler seam in addition to the HTTP middleware; direct helper coverage proves oversized input fails before conversion. |
| 65 | A Bard error arrives after useful stream text and is lost or misreported. | PROTECTED | Fixture tests cover midstream errors and stream cleanup; live provider error envelopes remain upstream-dependent. |
| 66 | Cumulative upstream text resets or candidate changes cause duplicated or missing output. | PROTECTED | Golden stream tests cover cumulative/repeated chunks, reconnects, partial state, and deduplication. |
| 67 | A cancelled follower cancels the shared upstream request for other callers. | PROTECTED | Subscriber contexts are independent; follower cancellation tests prove the leader stream continues. |
| 68 | A slow follower silently loses chunks and receives a corrupted transcript. | PROTECTED | Subscriber buffers are bounded and return `ErrStreamSubscriberTooSlow` instead of dropping data silently. |
| 69 | Chat history grows without a count/byte/turn bound and eventually freezes the UI or local storage. | PARTIAL | The studio now caps runtime history at 200 messages, clips persisted content/reasoning, enforces a 4-million-character serialized budget, and rejects oversized legacy storage; real browser quota/performance acceptance remains. |
| 70 | Stream coalescing shares a response across different cookies, API keys, provider modes, or capabilities. | PARTIAL | Configured session-bound clients disable coalescing; anonymous scope is explicit, but a future multi-provider key must remain in the scope. |
| 71 | Cancelling the leader leaves followers waiting forever or returns a false successful stream. | PROTECTED | Context-aware leader emission propagates cancellation into the shared stream terminal error; the leader/follower cancellation regression asserts `context.Canceled` reaches both callers. |
| 72 | A transport/protocol error is rendered as an assistant Markdown answer and stored as successful history. | PROTECTED | Web-RPC and direct Developer API streams emit a top-level structured SSE error plus `[DONE]`, never assistant Markdown; the browser recognizes protocol/transport interruption and incomplete EOF, stores `status: error` or `stopped`, and excludes those turns from future prompt history. |
| 73 | Unknown SSE event names or fields crash a client or silently change ordering. | PARTIAL | The Developer API parser ignores standard comments/fields, preserves multi-line `data:` events, and now fails closed on an empty or `[DONE]`-only stream; complete unknown-event diagnostics and Web RPC framing evidence remain open. |
| 74 | UTF-8 sequences split at a byte boundary are truncated or replaced. | PROTECTED | The thinking splitter is exercised one byte at a time with Hindi and emoji, and the upstream parser is exercised across arbitrary byte boundaries while preserving the Unicode response fixture. |
| 75 | Lack of keepalive causes a proxy/browser to close a slow reasoning stream. | PROTECTED | Local SSE keepalive exists and has lifecycle tests; measure real proxy behavior separately. |
| 76 | Prompt/Markdown tool extraction is presented as native Google function calling. | PARTIAL | Keep the classification as emulated/partial and do not advertise native tool fidelity. |
| 77 | Ordinary Markdown accidentally matches a tool-call fence and is executed/extracted. | PARTIAL | Existing parser tests cover accidental Markdown; expand adversarial fence/content cases before broad tool claims. |
| 78 | Nested schemas, enums, arrays, nullable fields, or large arguments are flattened or rejected unexpectedly. | PARTIAL | Golden OpenAI and direct-Developer-API fixtures preserve nested arrays, enums, nullable values, Unicode, and large valid arguments; malformed content is now rejected instead of dropped, while cross-adapter semantic fidelity and provider acceptance remain open. |
| 79 | `none`, `auto`, `required`, or a named tool choice changes behavior silently. | PARTIAL | Shared validation rejects unknown modes, malformed choices, and undeclared names; prompt/direct-API paths normalize case/whitespace and Anthropic maps `any`/`tool`; model obedience and parallel-tool semantics remain emulated/partial. |
| 80 | Streaming tool calls are split across deltas and emitted as invalid partial JSON. | PARTIAL | The direct Developer API stream uses a bounded ID-ordered assembler, retains cumulative snapshots, rejects changed IDs/names and invalid arguments, and emits calls only after successful completion; the reverse-engineered web-RPC tool path remains emulated/buffered and is not relabelled native. |

## 5. Adapter and counting fidelity

| # | Failure path | Status | Proactive step now |
|---:|---|---|---|
| 81 | OpenAI Responses items unsupported by the translator disappear without a client-visible error. | PROTECTED | `ResponsesInputToMessages` rejects unsupported item/content types, wrong field types, missing tool IDs/arguments/results, and invalid JSON arguments; focused fixtures cover supported tool continuations and malformed items. The selected Responses surface remains partial overall. |
| 82 | Anthropic content blocks or SSE lifecycle events are dropped/reordered. | PARTIAL | Anthropic input conversion now rejects unsupported/malformed blocks and preserves tool-result order; focused fixtures cover image, tool-use, tool-result, and error cases. Complete Anthropic SSE lifecycle parity remains unproven. |
| 83 | Thinking fences split at arbitrary boundaries or remain unterminated. | PROTECTED | Golden formatter/parser tests cover fragmented open/close fences, missing closes, and following content. |
| 84 | `RandHex` receives an invalid/oversized request and panics or allocates excessively. | PROTECTED | `internal/format/openai.go` clamps invalid lengths to a bounded result and `openai_test.go` covers negative, zero, oversized, and normal lengths. |
| 85 | Estimated token counts are interpreted as authoritative billing/context counts. | PARTIAL | Label all estimates, expose the method/limits, and compare against provider counting only in explicit live experiments. |
| 86 | Tool arguments are required to be an object even when the declared schema permits another JSON value. | PARTIAL | Preserve JSON value types where the target protocol supports them; add scalar/array/null fixtures. |
| 87 | Tool schema depth, property count, enum size, or argument bytes are unbounded. | PROTECTED | Shared validation bounds tool count, declaration bytes, names/descriptions, argument bytes, schema depth/nodes/properties/enums, and provider-returned tool calls across prompt, Google, and direct-Developer-API paths; typed/cyclic schema fixtures are covered. Streaming aggregation is tracked separately in row 80. |
| 88 | Tool-result IDs/names are mismatched, so a continuation is attached to the wrong call. | PROTECTED | `ValidateToolResultReferences` rejects unknown IDs, name mismatches, and ambiguous name-only continuations before prompt flattening or direct translation; the Responses adapter preserves `call_id`, `name`, and `tool_calls` through server conversion, with valid and invalid-correlation fixtures. |
| 89 | Multiple upstream candidates are collapsed without documenting selection semantics. | PARTIAL | The direct Developer API adapter rejects more than one candidate instead of silently selecting one; the reverse-engineered web-RPC parser remains separately constrained and live candidate semantics are not claimed. |
| 90 | Unknown finish reasons are mapped to normal `stop`, hiding truncation or safety failure. | PARTIAL | Direct Developer API translation maps only known stop/length/filter/tool reasons and fails closed on unknown or malformed reasons; the web-RPC path remains provider-dependent and is not claimed equivalent. |

## 6. Multimodal, browser security, artifacts, and UI

| # | Failure path | Status | Proactive step now |
|---:|---|---|---|
| 91 | A remote image URL reaches loopback, private IP space, or an unsafe scheme. | PROTECTED | URL scheme/host/port validation, private-address rejection, redirect checks, and fetch-size limits are tested. |
| 92 | Scotty upload accepts a non-2xx response, oversized body, malformed ref, or untrusted upload URL. | PROTECTED | Upload start/finish status, bounded response bodies, exact Google upload-host validation, and malformed-ref tests exist. |
| 93 | DNS changes between remote-image validation and connection, bypassing the SSRF check. | PROTECTED | The exported fetch path requires a guardable HTTP client; its direct transport re-resolves immediately before dialing, rejects private answers, dials the approved literal public IP, and has a rotating-resolver regression test. Live network/egress behavior remains an external acceptance gate. |
| 94 | Image decode, resize, OCR, or base64 expansion exhausts memory. | PARTIAL | Byte, dimension, pixel, and client file limits exist; audit OCR/browser memory and concurrent-image budgets. |
| 95 | The SHA-256 → Scotty reference cache grows forever or retains a ref beyond its session validity. | PARTIAL | The cache now stores only 256 bounded, LRU-like references, shares concurrent misses per hashed single-cookie scope, and disables reuse for configured multi-account pools. Scotty reference TTL/expiry remains provider-dependent and open. |
| 96 | `FileReader`, PDF/DOCX/XLSX parsing, or OCR reads an oversized local file and freezes the WebView. | PROTECTED | Client-side attachment size is bounded at 32 MiB and storage text is capped; add worker/cancellation coverage for parse CPU. |
| 97 | A malicious webpage uses wildcard CORS/PNA to drive a privileged localhost gateway. | PARTIAL | Default browser access now requires the exact gateway origin; remote/alternate origins require explicit allow-listing. A real browser/PNA exploit test and capability-token decision remain. |
| 98 | API keys or sensitive values leak through query strings, referrers, logs, error text, or browser history. | PARTIAL | Provider keys are header-only and errors redact them; gateway `?key=` authentication is disabled by default and only an explicit legacy opt-in enables it. URL/log/referrer behavior outside BOB still requires deployment-level controls. |
| 99 | CDN libraries disappear, CSP/iframe policy changes, or unbounded artifact HTML grows into a browser/desktop denial of service. | PARTIAL | Artifact sandboxing and runtime-error reporting exist; pin/verify critical assets or provide a documented offline behavior and size budget. |
| 100 | Artifact preview, code hydration, scroll controls, external links, resize handling, or modal lifecycle diverge across native and hosted builds. | PROTECTED | Source tests now cover token normalization, empty-editor recovery, geometry-based scroll controls, local/public bundle sync, browser-opening bridge, sandbox isolation, and resize hooks; real-browser CDN and clean-device checks remain. |

## Required execution order

1. Close release/provenance rows 1–6, 11–20, and 40 before another public
   release. Revoke the exposed credentials independently of code work.
2. Keep the updater-lock, leader-cancel, history-bound, cache-bound, schema,
   and stream-boundary tests (23, 25, 69, 71, 84, 87, 95) in every release
   validation run; their presence is not a substitute for the still-open
   provider and device gates.
3. Run a clean mounted-DMG and `/Applications` acceptance sequence, followed by
   a real same-key update and rollback (17, 22, 24, 26, 32–34, 37, 40).
4. Run the cross-adapter tool torture suite and publish NATIVE/
   EMULATED_PARTIALLY/UNSUPPORTED results (76–90).
5. Run the browser-origin/SSRF and CDN/offline tests (97–99); retain the DNS
   rebinding regression for every release before opening remote Web Studio
   access or handling untrusted classroom webpages.
6. Only after those gates pass, stage a two- or three-device pilot and then
   roll out to 30 devices in waves. Installation/update success must be
   recorded separately from Google generation success and quota behavior.

## Evidence surfaces to keep synchronized

- [`VERIFICATION-MATRIX.md`](VERIFICATION-MATRIX.md) — claim classification.
- [`CORE-REGRESSION-HARNESS.md`](CORE-REGRESSION-HARNESS.md) — protocol
  invariants and fixture scope.
- [`RELEASE-READINESS-v0.2.0.md`](RELEASE-READINESS-v0.2.0.md) — signed
  release and device gates.
- [`DESKTOP-UPDATE-OPERATIONS.md`](DESKTOP-UPDATE-OPERATIONS.md) — updater
  custody, rollback, and rollout contract.
- [`GEMINI-API-ROUTING.md`](GEMINI-API-ROUTING.md) — explicit student-owned
  Developer API boundary.

This register deliberately leaves unresolved rows visible. Hiding them would
make the repository less safe to evolve than acknowledging them.
