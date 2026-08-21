# CLAUDE.md — Claude Code Instructions for BOB Gemini Free

This repository contains **BOB Gemini Free** (*Break Ordinary Boundaries*) by **ABCsteps** ([abcsteps.com](https://abcsteps.com/)) and **Divyanshu Singh Chouhan** ([@div197](https://github.com/div197)).

---

## 1. Quick Build & Test Commands

```bash
# Build host binary
make build

# Cross-compile for all operating systems (macOS, Linux, Windows)
make dist

# Build Native Desktop App (Requires Wails CLI & CGO)
make desktop

# Run all tests across all packages
go test -count=1 ./...

# Run diagnostic integration test kit against local server
./bob-gemini-free --test --test-url http://127.0.0.1:9610

# Run concurrency benchmark
./bob-gemini-free --bench --bench-concurrency 3 --bench-requests 6

# Format Go code
gofmt -s -w .
```

---

## 2. Architecture & Design Principles

- **Pure Go Gateway**: A Go single-binary gateway with bundled external Go dependencies; runtime size and RSS are environment/build dependent and must be measured.
- **Multi-protocol Gateway** (implemented adapters; compatibility remains endpoint- and upstream-dependent):
  - `POST /v1/chat/completions`, `POST /v1/responses`, `POST /v1/images/generations`, `GET /v1/models` (OpenAI format)
  - `POST /v1/messages` (Anthropic format with full SSE lifecycle: `message_start`, `content_block_delta`, `message_delta`, `message_stop`)
  - `GET /v1beta/models`, `POST /v1beta/models/{target}` (Google Gemini format)
  - `pkg/gateway` (Embeddable Go library)
- **Zero-Config Cookie Auto-Discovery**: Automatically checks `./cookie.txt` and `~/.config/bob-gemini-free/cookie.txt` on startup.
- **Multimodal Scotty Engine**: Automatic image compression (`internal/multimodal/compress.go`) and Google Scotty resumable upload.
- **Observability Headers**: Injects standard OpenAI headers (`x-request-id`, `openai-processing-ms`, `openai-version`, `x-ratelimit-*`).

---

## 3. How Claude Code Uses BOB Gemini Free

When using Claude Code with this gateway:

```bash
export ANTHROPIC_BASE_URL=http://127.0.0.1:9610
export ANTHROPIC_API_KEY=none
claude
```

---

## 4. Cookie Configuration & Pro Routing

- **1-Click Interactive Sign-In**: Run `./bob-gemini-free --login` to open a standalone window, sign in once, and auto-capture session cookies without DevTools.
- **Automatic Discovery**: BOB Gemini Free automatically scans `./cookie.txt` and `~/.config/bob-gemini-free/cookie.txt` on startup.
- **Manual Extraction**: In Chrome DevTools Network tab, copy `Cookie:` from the `app?eom=1...` document request or `StreamGenerate`.
- **Multi-Account**: Set `"auth_user": "1"` in `config.json` if using a secondary account (`/u/1/app`).
- **Permissions**: Cookie files are created with mode `0600` for security.

---

# Round 2 Deep Static Codebase Analysis Report

Date: 2026-08-20

Mode: static analysis only. No build, test, benchmark, gateway startup, live provider request, browser automation, or external rate-limit experiment was executed during this onboarding pass. `cookie.txt` was identified as a sensitive local session file and was not reproduced in this report.

## 1. Analysis Scope And Method

This Round 2 analysis was performed as an additive onboarding report for the current workspace at `/Users/suru/Documents/BOB-Gemini-Free`. No `shree-ganesh.md` file was present in the workspace, so the prior durable agent-facing report source was the existing `CLAUDE.md`. The earlier `CLAUDE.md` content has been preserved above and this section extends it instead of deleting or replacing it.

The workspace is a Go module named `github.com/div197/bob-gemini-free`. The repository contains source code, tests, documentation, static web distribution files, Cloudflare Pages edge-function prototypes, installers, service templates, example SDK clients, visual assets, local generated binaries under `dist/`, and a local `cookie.txt` credential file. Standard ignore files identify `dist/`, binary outputs, `config.json`, `cookie.txt`, cookie variants, coverage output, OS metadata, and IDE/editor files as ignored or sensitive. The current working tree still includes some ignored or generated artifacts, so the static inventory separates meaningful source and documentation from binary or credential material.

The analysis used static file traversal and content reading. Text and source files were inspected for their module responsibilities, data structures, imports, request/response shapes, routing, conversion logic, tests, scripts, and documented claims. Binary assets and generated executable files were recorded by file type and dimensions or executable format rather than decoded as source. The local `cookie.txt` was treated as secret material. It confirms that authenticated-session workflows are represented in the checkout, but its raw content must remain out of reports, logs, and versioned documentation.

## 2. High-Level Project Purpose

BOB Gemini Free is a local-first AI gateway written primarily in Go. Its stated purpose is to expose several familiar AI API surfaces to local developer tools while translating those requests into Google's Gemini web RPC flow. The core value proposition in the documentation is that a user can run a single local binary and point OpenAI-compatible, Anthropic-compatible, Google Gemini-compatible, and embedded Go clients at it.

The repository describes itself as a universal gateway for:

- OpenAI-style endpoints such as `/v1/chat/completions`, `/v1/responses`, `/v1/models`, `/v1/models/{model}`, `/v1/tokens/count`, and `/v1/images/generations`.
- Anthropic Messages endpoint `/v1/messages`, including streaming Server-Sent Events, Claude Code style thinking blocks, and prompt-cache counter fields.
- Google Gemini style endpoints under `/v1beta/models` and `/v1beta/models/{target}`, including `generateContent`, `streamGenerateContent`, and `countTokens`.
- Embedded Go usage through `pkg/gateway`, where another Go program can instantiate a handler or direct inference engine without launching a separate process.

The software domain is protocol translation, local AI tooling integration, browser-session authentication, multimodal upload support, streaming response parsing, and developer-facing diagnostics. It is not a traditional SaaS backend with a database. The local binary acts as a stateless gateway plus in-memory telemetry counters. Its durable state is configuration files and Google session cookie files, not an application database.

## 3. Technology Stack

The main application stack is Go. The current module declares `go 1.26.6` in
`go.mod`; older Go compatibility has not been verified. The code uses the
standard library heavily: `net/http`, `http.ServeMux`, `encoding/json`,
`context`, `sync`, `sync/atomic`, `os`, `path/filepath`, `time`, `regexp`,
`crypto`, `image`, and platform process utilities. External Go dependencies
are intentionally narrow but non-zero: `github.com/bogdanfinn/fhttp`,
`github.com/bogdanfinn/tls-client`, `github.com/bogdanfinn/websocket`, and
`golang.org/x/image`, plus desktop dependencies. Transitive dependencies
include compression, QUIC/uTLS support, websocket support, crypto, net, sys,
and text packages.

There is also a static web studio in `internal/server/playground.html` and a synchronized copy in `web/index.html`. The web studio is plain HTML, CSS, and JavaScript. It loads browser-side libraries from CDNs: Marked for Markdown, DOMPurify for sanitization, KaTeX for math rendering, Prism for code highlighting, and Mermaid for diagrams. The `web/sw.js` service worker implements a cache-first PWA strategy for static assets and pass-through behavior for API/local gateway requests.

The repository includes JavaScript edge-function prototypes in both `functions/` and `web/functions/`. These implement a minimal Cloudflare Pages style health endpoint, model catalog endpoint, and OpenAI chat-completions proxy. They are separate from the Go binary and mirror only a subset of the gateway behavior.

Build and deployment assets include `Makefile`, `Dockerfile`, `docker-compose.yml`, `install.sh`, `install.ps1`, `test-kit.sh`, `test-kit.ps1`, systemd unit, launchd plist, and Windows batch runner. Examples exist for Python, Node.js, Go, and cURL.

## 4. Architecture

The dominant architecture is a single-process local gateway monolith with internal package boundaries. It uses a layered flow:

1. CLI and process orchestration in `main.go`.
2. Configuration loading and auto-discovery in `internal/config`.
3. HTTP routing, middleware, handlers, telemetry, and SSE output in `internal/server`.
4. Request schema models in `internal/models`.
5. Protocol conversion and parsing in `internal/format`.
6. Gemini web RPC request building, auth header construction, cookie pooling, transport, and stream parsing in `internal/gemini`.
7. Image compression, token discovery, and Google Scotty upload in `internal/multimodal`.
8. Native browser login automation in `internal/browser`.
9. OS service management in `internal/service`.
10. Update checking and self-update mechanics in `internal/updater`.
11. Diagnostics and benchmarks in `internal/diag`.
12. Embeddable library facade in `pkg/gateway`.

The gateway is not event-driven in a distributed-system sense. It is request/response and stream-oriented over HTTP. It uses concurrency through Go's HTTP server, atomic counters, a cookie-pool cursor, background cookie-pool reload, and streaming callbacks. Long-running streaming endpoints keep the HTTP connection open and emit protocol-specific SSE chunks.

The runtime dependency chain is intentionally direct. `server.App` owns configuration, a Gemini client, a multimodal token cache, an HTTP client, telemetry counters, and an image-cache map. `App.Handler()` creates the route table and wraps the mux with auth/logging and CORS middleware. Each endpoint handler validates JSON, resolves the model alias, converts the client request into a Gemini prompt and optional image references, invokes `gemini.Client`, and converts the returned text or stream deltas into the expected client protocol.

## 5. Documentation Inventory

The root `README.md` is the primary English user-facing document. It describes BOB Gemini Free as an API-less local gateway, emphasizes zero cloud billing and local-first privacy, lists supported clients, explains model tiers, provides quick starts for binary, Docker, and SDK usage, documents endpoints, and presents a broad feature set including reasoning, multimodal vision, Imagen/Nano Banana image generation, model aliases, diagnostic tooling, service installation, auto-update, and a static web playground.

`README.hi.md` is a Hindi-language counterpart. It covers the same broad product story and usage pathways for Hindi readers.

`CHANGELOG.md` is the version-history and feature-evolution document. It records the move toward universal protocol support, diagnostic and benchmark tooling, background service operations, embedded Go library support, Responses API support, image generation, multimodal vision, 1-click login, cookie pooling, dynamic token extraction, documentation suite additions, and web studio distribution.

`docs/README.md` indexes the documentation suite into five chapters: getting started, authentication and routing, AI client integrations, API reference, and embedded Go SDK. The getting-started docs cover quickstart, standalone binaries without a separately managed runtime, and Docker/OrbStack. The authentication docs cover 1-click login, manual DevTools cookie extraction, cookie pools, and `auth_user` routing. The client integration docs cover Cursor/Windsurf/Continue, Claude Code CLI, OpenAI Codex CLI, and broader agent frameworks. The API reference docs define OpenAI-shaped, Anthropic-shaped, Google v1beta-shaped, image-generation, health, diagnostics, and benchmark routes. The embedded SDK guide shows `gateway.NewHandler` and `gateway.NewEngine`.

`examples/README.md` describes the examples directory: Python OpenAI, Python Anthropic, Node OpenAI, Node Anthropic, embedded Go, and shell/cURL requests.

`web/README.md` describes the static web studio as a zero-server frontend that communicates with the user's local gateway. It also documents PWA/offline behavior and potential static hosting targets such as Cloudflare Pages, GitHub Pages, Vercel, Netlify, or S3.

`AGENTS.md` is the agent-facing architecture and workflow guide. It defines this project as a high-performance local AI gateway, lists the directory structure, standard build/test/service commands, cookie authentication guidance, and cross-platform/security conventions. The pasted root instructions also requested an exhaustive Round 2 static analysis and additive report behavior, which this document implements.

## 6. Root Entrypoint And CLI Flow

`main.go` is the root executable entrypoint. It declares `Version = "dev"` and uses `resolveVersion()` to prefer an injected ldflag version or module build info. It implements several CLI modes before normal server startup:

- `--login` calls `handleBrowserLogin`, launches a Chromium-based browser through `internal/browser`, extracts Google Gemini cookies through CDP, and saves cookies to both local `./cookie.txt` and the user's config directory with secure permissions.
- `--setup-cookie` and `--cookie-string` call `handleCookieSetup`, parse user-provided raw cookie material, validate token extraction, and write the session cookie into the user config path.
- `--test` calls `diag.RunDiagnosticsWithProgress` against a target URL.
- `--bench` calls `diag.RunBenchmark`.
- `--status` performs a GET request against the health endpoint and prints telemetry.
- `--update` calls the updater package.
- `service` subcommands call the service package for install, uninstall, start, stop, and status.
- `--version` prints the resolved version.

After CLI-mode checks, normal server startup resolves config from CLI flag, environment variable, discovered config path, and config defaults. CLI flags override loaded config fields for host, port, cookie file, proxy, and impersonation profile. It then constructs `server.New(cfg, currentVersion)`, prints startup metadata, creates an `http.Server`, starts `ListenAndServe()` in a goroutine, waits for interrupt or SIGTERM, and performs graceful shutdown with a 10-second timeout.

## 7. Configuration Layer

`internal/config/config.go` defines the `Config` struct. Fields include `Port`, `Host`, `CookieFile`, `CookiePool`, `CookiePoolDir`, `GeminiBL`, `XSRFToken`, `DefaultModel`, `AuthUser`, `APIKeys`, `Proxy`, `LogRequests`, `RetryAttempts`, `RetryDelaySec`, `RequestTimeoutSec`, and `Impersonate`.

`Default()` establishes the local-first default posture: host `127.0.0.1`, port `9610`, default model `gemini-3.6-flash`, retry attempts `3`, retry delay `2`, request timeout `180`, and a Gemini build label. `Load(path)` starts from defaults, loads JSON config if a path is supplied, then applies environment variable overrides. It also auto-discovers a cookie file and cookie pool if no explicit values are set. Environment variables use the `BOB_GEMINI_FREE_` prefix for host, port, cookie file, cookie pool, cookie pool dir, default model, auth user, API keys, proxy, log requests, retry attempts, retry delay, request timeout, and impersonation profile.

`Find()` looks for a local `config.json` and then `~/.config/bob-gemini-free/config.json`. `FindCookie()` looks for `./cookie.txt` and `~/.config/bob-gemini-free/cookie.txt`. `FindCookiePool()` looks in `./cookies` and `~/.config/bob-gemini-free/cookies` for `.txt` files. Tests cover default config, config-file loading, cookie discovery, environment overrides, retry settings, timeout settings, and cookie pool directory handling.

## 8. Model Registry

`internal/models` holds protocol schemas and model resolution. `models.go` defines `Model{Mode, Think, Desc, Extra}` and a large `MODELS` map. It includes Gemini models, developer convenience aliases, OpenAI aliases, Anthropic aliases, and image-generation aliases. The key runtime result is `Resolved{Name, Mode, Think, Extra}`.

`Resolve(modelName, defaultName)` trims a leading `models/` prefix, supports model suffix overrides of the form `@think=N`, looks up the model in the registry, and falls back to the configured default model if unknown. It returns an error only for invalid `@think=` parsing. The tests verify direct Gemini aliases, thinking model aliases, `models/` prefix handling, unknown fallback behavior, OpenAI aliases such as `gpt-5.6`, reasoning aliases such as `o3`, Anthropic aliases such as `claude-code`, invalid thinking suffix errors, and enhanced Pro extra-index fields.

`openai.go`, `anthropic.go`, and `google.go` define JSON request and response structs for the gateway's supported public protocols. These files are schema-focused and contain no transport behavior.

## 9. Gemini Upstream Client

`internal/gemini` is the most protocol-specific package. `auth.go` manages cookie parsing and token-derived authentication. `CookieCache` stores file path, modtime, and loaded cookie info. It can load raw cookie strings or JSON objects, extract key/value cookie pairs, identify `SAPISID` or `__Secure-3PAPISID`, and construct a normalized `Cookie` header string. `GetAtToken` fetches `https://gemini.google.com{accountPrefix}/app`, sends cookie and optional `SAPISIDHASH` authorization headers, and extracts dynamic `SNlM0e` or `thykhd` AT tokens plus `cfb2h` build labels from page HTML. `SAPISIDHash` constructs the Google-style timestamped SHA-1 authorization header over `timestamp SAPISID origin`. `SaveCookieFile` creates parent directories and writes cookie files with `0600` permissions.

`payload.go` builds the Gemini web RPC body. `BuildBodyWithAt` creates a sparse JSON array with 102 positions and fills indices used by the Gemini frontend request, including prompt/image refs at index 0, language, thinking mode at index 17, UUID at index 59, target model mode at index 79, and any extra model-specific indices. It wraps the sparse array in the outer `f.req` form payload and optionally attaches the `at` token. `BuildURL` constructs the `StreamGenerate` URL under `https://gemini.google.com{accountPrefix}/_/BardChatUi/...` with build label, language, request id, and `rt=c` streaming format.

`client.go` constructs a transport client, optionally using `tls-client` impersonation when configured without a proxy. It sets high connection pool limits and accepts proxy configuration or environment proxy fallback. `buildHeaders` sets the form content type, origin, referer, `X-Same-Domain`, user agent, optional `X-Goog-AuthUser`, cookie, and SAPISID authorization headers. `triageStatus` maps redirects, 429s, and non-200 upstream responses into `UpstreamError`. `GenerateContext` posts a non-streaming upstream request with retries, cookie-pool session selection, failure marking, and response text extraction. `GenerateStreamContext` does the same for streaming and uses `StreamParser` to emit incremental deltas.

`parse.go` and `stream.go` decode Google BoQ/batchexecute-style response lines. `ExtractTextsFromLine` looks for `"wrb.fr"` records, parses the outer JSON, parses the embedded inner JSON, and extracts candidate text from `inner[4]` style structures. `CleanText` strips code-execution artifacts and card-content URLs. `IsBardError` recognizes `BardErrorInfo` codes. `FormatBardError` maps notable codes such as `1003` for image/auth rejection and `1024`/`42901` for rate limits into gateway-facing messages. `StreamParser` stores `prevText` and a line buffer, extracts only new deltas from cumulatively growing upstream text, and preserves previous text across stream retries to reduce duplicate emission.

`pool.go` implements a multi-account cookie pool. It can add sessions, load from explicit files, load from directories, reload from registered sources, start a 30-second auto-reload goroutine, count total/healthy sessions, select the next healthy session with an atomic round-robin cursor, mark failures with a 60-second cooldown, and reset failure counts on success.

`fingerprint.go` adapts `tls-client`/`fhttp` to the standard `http.Client`-like `Requester` interface. It supports named Chrome, Firefox, and Safari profiles and caches constructed clients by profile and timeout.

## 10. Protocol Formatting Layer

`internal/format` turns public API request objects into normalized prompts and turns Gemini text back into protocol-specific constructs.

`openai.go` builds OpenAI chat prompts. It supports role rendering for system, developer, user, assistant, and tool messages. It handles string content, multi-part text, and image URL parts. It extracts data URLs and remote image URLs as `format.Image` values for later upload. It injects tool definitions into the prompt through a fenced `tool_call` instruction format, honors tool-choice hints such as `none`, `required`, function name choices, and JSON response-format hints. `ParseToolCalls` extracts fenced JSON tool calls and returns clean text plus `OpenAIToolCall` records. `RandHex` generates IDs.

`anthropic.go` converts Anthropic Messages requests into OpenAI chat requests. It handles string, map, and list system prompts; user/assistant text blocks; image blocks converted to OpenAI `image_url` data URLs; assistant `tool_use` blocks converted to OpenAI tool calls; and `tool_result` blocks converted to tool messages. It maps Anthropic tools to OpenAI functions and maps `thinking` requests to reasoning effort and, for known Claude aliases, thinking Gemini routing. It also converts text, thinking, and OpenAI tool calls back into Anthropic content blocks.

`google.go` converts Google `contents`, `systemInstruction`, `tools`, and `toolConfig` into a prompt plus image list. It supports inline base64 image decoding, function-call parts, function-response parts, tool prompt injection, and Google tool-choice instructions for `NONE` and `ANY`. It parses fenced and unfenced function-call JSON back into Google function-call structs.

`responses.go` converts OpenAI Responses API `input` and `instructions` into chat-message maps. It supports string input, list input, function-call output items, assistant message items, output text, function-call content parts, and user message content arrays. `BuildResponseOutput` constructs Responses API output items for function calls and output text messages.

`thinking.go` implements a stream splitter that recognizes leading fenced `thought` or `thinking` blocks and separates reasoning deltas from final content. This splitter is used by both OpenAI and Anthropic streaming endpoints so reasoning can be surfaced as `reasoning_content` or Anthropic `thinking_delta`. `ExtractThinking` performs a non-streaming equivalent for complete responses.

`tokens.go` implements local token estimation. It counts word-like runs, CJK characters, symbols/emojis, punctuation, whitespace boundaries, message framing overhead, tool declaration overhead, reasoning content, and a fixed image token cost of 258 per standard image.

`citations.go` extracts Markdown-style non-image links from generated text into structured citation records. This is a simple link parser, not a full grounding metadata parser.

## 11. HTTP Server Layer

`internal/server/server.go` defines `App` and route registration. The app includes config, Gemini client, token cache, HTTP client, log function, version, atomic request/token counters, start time, and image cache. Routes include:

- `GET /` health and telemetry.
- `GET /playground` and `GET /ui`.
- `GET /v1/models`.
- `GET /v1/models/{model}`.
- `POST /v1/chat/completions`.
- `POST /v1/responses`.
- `POST /v1/messages`.
- `POST /v1/images/generations`.
- `POST /v1/tokens/count`.
- `GET /v1/update/check`.
- `GET /v1beta/models`.
- `POST /v1beta/models/{target}`.

`middleware.go` provides API-key authorization, origin-aware CORS, body-size limiting, logging, observability headers, and OpenAI-like rate-limit headers. Authorization accepts configured keys through `Authorization: Bearer`, `x-api-key`, `x-goog-api-key`, or `?key=`. If no API keys are configured, non-browser/native requests remain allowed; browser origins are restricted to loopback defaults or explicit `allowed_origins`. The body limit is 32 MB.

`helpers.go` maps `gemini.UpstreamError` statuses to HTTP status codes, writes JSON, starts SSE, writes SSE data/events, writes `[DONE]`, and uploads images. `uploadImages` fetches or uses image bytes, hashes them for the `ImageCache`, compresses where needed, uploads through `multimodal.UploadImage`, and returns Gemini file refs.

`handlers.go` serves the unauthenticated local `/healthz` probe, aggregate `/v1/metrics`, the richer `/` health view, model catalogs, single model lookup, token-count endpoints, Google model catalogs, and update-check results. Aggregate metrics never include cookies, auth headers, prompts, or image contents.

`chat.go` handles OpenAI Chat Completions. It validates JSON, resolves model and reasoning effort, defaults tool choice, converts messages to prompt/images, uploads images, increments request counters, and chooses streaming or non-streaming behavior. Streaming without tools directly passes Gemini stream deltas through a thinking splitter and emits OpenAI chat completion chunks. Non-streaming and tool paths buffer full output, parse tool calls, extract thinking, calculate usage, and return OpenAI-compatible responses.

`anthropic.go` handles Anthropic Messages. It validates JSON, resolves model and thinking budget, converts to OpenAI chat request, converts to prompt/images, uploads images, and emits Anthropic SSE lifecycle events or a non-streaming `message` response. Streaming logic starts and stops content blocks for thinking and text and sends final usage.

`google.go` handles Google-shaped Gemini routes. It parses the `{target}` path value into model and action. `countTokens` returns local token estimates. `generateContent` and `streamGenerateContent` convert Google request content to prompt/images, upload images, resolve tools, and emit Google-style candidates. Streaming without tools sends SSE candidate chunks and a final usage chunk. Tool-enabled streaming is buffered and replayed.

`responses.go` handles the OpenAI Responses API. It supports `instructions`, `input`, `stream`, tools, and model resolution. Streaming without tools emits a Responses API event lifecycle: `response.created`, `response.output_item.added`, `response.content_part.added`, text deltas, content done, output text done, output item done, and `response.completed`. Tool-call streaming buffers first and then emits function-call or output-text done events. Non-streaming returns `output`, `output_text`, and usage.

`images.go` handles OpenAI image-generation requests. It validates prompt, resolves image model aliases, asks Gemini for image output, extracts generated image URLs from response text, optionally fetches and base64-encodes images for `b64_json`, and returns OpenAI image response objects.

`playground.go` embeds `playground.html` with `go:embed` and serves it as HTML.

## 12. Multimodal Layer

`internal/multimodal/compress.go` compresses image bytes when they exceed configured limits or lack reliable MIME handling. It decodes PNG/GIF/WebP through registered image decoders, scales large images down to max dimension 1024 using Catmull-Rom sampling, and encodes JPEG at quality 75. `CompressIfNeeded` performs base64 decode/compress/re-encode for large base64 images.

`tokens.go` in the multimodal package manages page-token discovery for Google Scotty upload. Defaults are `DefaultPushID` and `DefaultPctx`, with a 600-second token cache TTL. It fetches Gemini app HTML using cookie and SAPISID authorization when available, extracts Push ID, Pctx, AT token, and BL using regexes, and avoids concurrent dog-pile refresh by updating the timestamp before the network fetch.

`upload.go` implements Google Scotty resumable upload in two steps. It starts the upload with push id, tenant id, client pctx, content length, content type, origin, referer, user agent, cookie, authorization, and optional auth-user headers. It then uploads and finalizes the image bytes to the returned upload URL. A valid response must be a file ref starting with `/`. `FetchImageBytes` only allows HTTP and HTTPS image URLs, rejecting schemes such as `file://`.

Tests cover image compression behavior, GIF re-encoding, base64 compression, fetch scheme rejection, token cache construction, and a live image upload test that is skipped when `cookie.txt` is unavailable.

## 13. Browser Login Layer

`internal/browser/browser.go` implements the 1-click login workflow. `FindBrowser` searches platform-specific Chromium browser locations and executable names for macOS, Linux, and Windows. `getFreePort` allocates a localhost debug port. `LaunchLoginSession` creates a temporary browser profile, launches the browser in app mode against Gemini with remote debugging enabled, polls the CDP version endpoint, connects to the WebSocket, retrieves cookies, watches for required Google session tokens, and returns an extracted cookie. The login is intentionally isolated from the user's normal browser profile through a temporary user-data directory.

Tests cover browser detection enough to tolerate missing browsers, free-port allocation, and CDP cookie parsing behavior.

## 14. Service And Update Layers

`internal/service/service.go` implements OS service installation and control. It has templates for macOS LaunchAgent plist, Linux systemd user service, and Windows Startup batch file. `Install` resolves the current executable path, creates platform-appropriate service definitions, loads or enables them where possible, and starts the gateway on the selected port. `Uninstall`, `Status`, `Start`, and `Stop` perform platform-specific service operations. Tests validate template rendering for macOS, Linux, and Windows.

`internal/updater/updater.go` checks GitHub releases, compares semantic versions, finds the matching binary asset for the current OS and architecture, verifies a signed `SHA256SUMS` manifest using the configured Ed25519 public key, checks binary magic bytes for Linux ELF, macOS Mach-O, or Windows PE, and atomically replaces the running executable on Unix. Tests cover manifest signatures, tampering, checksum mismatch, version comparison, and asset matching. Updates fail closed when the signed verification material is absent or invalid.

## 15. Diagnostics And Benchmarking

`internal/diag/diag.go` defines a 13-check live diagnostic suite. It tests health, OpenAI model catalog, single model lookup, fast chat completion, thinking chat completion, streaming SSE, developer role plus JSON-format behavior, Google generateContent, Responses API, Anthropic Messages, function calling, image generation, and token counting. This suite is designed to run against an already running gateway and can involve upstream/provider calls.

`bench.go` implements a concurrent benchmark against `/v1/chat/completions`. It uses worker goroutines, records latencies, calculates average/P50/P90/P99, tracks success/failure counts, estimates or reads token counts, and calculates requests/sec and tokens/sec. Tests use local `httptest` servers and do not require live upstream access.

`agent_test.go` contains broader integration-style tests for agentic multi-turn tool workflows, Codex Responses API shape, model retrieval, Claude Code Anthropic workflow, and image generation workflow. Several paths accept either successful local response or `502 Bad Gateway` because the unit test environment may not have live upstream connectivity.

## 16. Embedded Go Library

`pkg/gateway/gateway.go` exposes the gateway as an importable package. Functional options mirror config fields: host, port, cookie file, default model, API keys, proxy, cookie pool, auth user, impersonation profile, log requests, retry settings, timeout, cookie pool dir, and version override. `NewEngine` applies options over defaults, creates `server.App`, and returns an `Engine`. `Handler` exposes the HTTP handler. `Generate`, `GenerateWithMedia`, `GenerateStream`, and `GenerateStreamWithMedia` call into the underlying Gemini client after resolving the model. `NewHandler` is a convenience constructor.

Tests validate that `NewHandler` returns a handler, that options affect endpoint behavior, that API-key enforcement works in embedded mode, that an engine can be constructed with expected config, and that direct generate methods exist even though upstream behavior depends on live connectivity.

## 17. Static Web Studio

`internal/server/playground.html` and `web/index.html` are large, mostly matching HTML/CSS/JavaScript application files. Their role is the user-facing local gateway studio. The page includes theming, command palette, gateway status modal, about modal, custom instruction modal, artifact preview sandbox, file/image attachment input, chat history persistence, Markdown rendering, DOMPurify sanitization, KaTeX math rendering, Prism code highlighting, Mermaid diagrams, transliteration helpers, speech/TTS controls, message editing, retry controls, export/print/copy operations, and streamed chat interaction with the local gateway.

The JavaScript stores chat history and preferences in local storage. The default gateway endpoint is `http://127.0.0.1:9610`, and the UI can save or reset this endpoint. The chat send flow builds OpenAI-compatible `/v1/chat/completions` requests, includes attached images/documents, consumes SSE chunks, separates `reasoning_content` from final assistant content, throttles expensive Markdown rendering during streaming, and writes final conversation state back to local storage. The page registers `sw.js` when service workers are available.

The static copy in `web/` is intended for deployment. `Makefile` target `web` copies `internal/server/playground.html` to `web/index.html` and copies the `functions` directory into `web/`. The generated distribution therefore keeps the embedded server playground and static site aligned when the target is run.

`web/manifest.json` identifies the PWA as "BOB Gemini Free - Universal AI Gateway Studio", sets standalone display, background/theme colors, developer/productivity categories, and favicon metadata. `web/CNAME` points the static site to `bob-gemini-free.abcsteps.com`.

## 18. Cloudflare Pages Edge Functions

The `functions/` and `web/functions/` directories currently contain matching JavaScript edge functions. `health.js` returns a stateless health payload with version `v0.1.4`, an approximate deploy epoch, zeroed counters, and CORS headers. `v1/models.js` returns a small static OpenAI-style model list. `v1/chat/completions.js` is a serverless edge proxy that builds a Gemini web RPC body, calls the Gemini StreamGenerate endpoint from the edge runtime, parses upstream text chunks, and returns OpenAI-compatible streaming or non-streaming chat responses.

These files are not equivalent to the full Go gateway. They omit most Go features: local cookie file handling, multimodal Scotty upload, Anthropic protocol, Responses protocol, rich model registry, diagnostics, local service management, embedded Go engine, and full middleware behavior. Their role appears to be static-site/serverless demonstration or fallback, not the primary production runtime described by the Go packages.

## 19. Build, Installation, And Deployment Files

`Makefile` defines `build`, `web`, `run`, `test`, `test-cover`, `dist`, and `clean`. `build` depends on `web`, compiles the local binary with `CGO_ENABLED=0`, and injects version `v0.1.5` through ldflags. `dist` cross-compiles Darwin arm64/amd64, Linux amd64/arm64, and Windows amd64 outputs under `dist/`.

`Dockerfile` is a multi-stage build. It uses `golang:alpine` as builder, downloads modules, builds a static Linux binary, and copies it into `alpine:3.21` with CA certificates and timezone data. It creates a non-root app user, exposes port 9610, defaults host to `0.0.0.0` inside the container, and probes the local unauthenticated `/healthz` endpoint for its healthcheck.

`docker-compose.yml` builds the local Dockerfile, maps port 9610, mounts `config.json` and `cookie.txt` read-only, sets host/port environment variables, probes `/healthz`, and restarts unless stopped.

`install.sh` and `install.ps1` implement cross-platform setup. They create user config directories, copy `config.example.json` where available, prefer existing local binaries, build from source when Go exists, build Docker when Docker exists, or download release binaries as a no-build fallback; the Go binary itself still includes its declared module dependencies.

`test-kit.sh` and `test-kit.ps1` are wrappers around the binary's `--test` mode. If no binary exists they compile first. `scripts/bench.sh` wraps `--bench`. The service templates under `scripts/` provide sample Linux systemd, macOS launchd, and Windows background runner definitions.

## 20. Tests And Coverage Shape

The test suite is broad for a compact gateway. It covers config defaults/discovery/env overrides, model resolution, cookie parsing and file permissions, payload construction, stream parsing, text cleaning, TLS profile selection, cookie pool selection/failover, protocol formatting, tool-call parsing, thinking splitting, token estimation, multimodal compression/fetch/token cache behavior, server routing/middleware/headers/token counts/playground, service templates, updater helpers, diagnostics runner, benchmark runner, and agentic API workflows.

The tests are a mix of pure unit tests, local `httptest` tests, and live-capability tests that skip or accept gateway errors when upstream access is unavailable. This means the repository has a strong static and local behavior harness, but live Gemini web compatibility, Google rate-limit behavior, real image generation, and authenticated Pro routing require a separately authorized runtime verification phase.

## 21. Assets And Generated Files

`assets/` contains visual branding and UI preview files:

- `bob-gemini-free-universal-gateway.png`: PNG, 1024 x 1536.
- `theme-bob-builder.png`: PNG, 1440 x 900.
- `bob-gemini-free-architecture.jpg`: JPEG, 1376 x 768.
- `theme-spotify.png`: PNG, 1440 x 900.
- `resized-preview.png`: PNG, 460 x 820.
- `mobile-preview.png`: PNG, 780 x 1688.
- `theme-apple.png`: PNG, 1440 x 900.
- `theme-vodafone.png`: PNG, 1440 x 900.
- `bob-gemini-free-banner.jpg`: JPEG, 1376 x 768.
- `bob-gemini-free-logo.jpg`: JPEG, 1024 x 1024.
- `bob-gemini-free-playground.png`: PNG, 1440 x 900.
- `theme-quantum.png`: PNG, 1440 x 900.

`dist/` contains generated native executables for Darwin ARM64, Darwin AMD64, Linux ARM64, Linux AMD64, and Windows AMD64. These are build outputs, and `.gitignore` excludes `dist/`. They were recorded by binary format rather than decoded.

`cookie.txt` exists locally and is ignored by `.gitignore`. It is a sensitive Google session credential file. Its existence confirms local authenticated-session state, but no secret value is included here.

## 22. Conceptual Data Flow

For an OpenAI chat request, a client sends JSON to `POST /v1/chat/completions`. Middleware optionally checks API keys, injects CORS and observability headers, and limits the request body. `handleChat` decodes `OpenAIChatRequest`, resolves the model alias to Gemini mode/think settings, converts messages and tools into a prompt and images, uploads images if present, calls `gemini.Client.GenerateContext` or `GenerateStreamContext`, splits thinking from content, parses tool calls where appropriate, estimates token usage, increments counters, and writes OpenAI-compatible JSON or SSE.

For an Anthropic request, `handleAnthropicMessages` decodes Anthropic schema, applies thinking-budget routing, converts the request to an OpenAI-like chat request, reuses the prompt/image conversion path, invokes Gemini, and maps output into Anthropic content blocks and SSE lifecycle events.

For a Google request, `handleGoogleGenerate` parses native `contents`, optional inline images, system instruction, tools, and tool config. It then converts them to a prompt and image references, invokes Gemini, and maps text or function-call output into Google candidate structures.

For image generation, the OpenAI image-generation endpoint treats the prompt as a request to Gemini image-capable models, extracts image URLs from the generated text, and optionally fetches those URLs to return base64 JSON.

For multimodal vision, image inputs from OpenAI/Anthropic/Google request shapes become `format.Image` objects. The server hashes and caches images, compresses them if needed, obtains upload tokens, sends bytes to Google Scotty, receives file refs, and attaches those refs at index 0 of the Gemini web RPC payload.

For authentication, local cookie files are loaded through `CookieCache` and optionally pooled through `CookiePool`. Requests use cookie headers plus SAPISID-derived authorization when available. Dynamic AT and build tokens are refreshed from Gemini app HTML.

## 23. Static Proof Boundary

Confirmed by file content:

- The repository is a Go single-binary gateway with standard-library HTTP routing.
- It implements route handlers for OpenAI Chat Completions, OpenAI Responses, OpenAI Images, OpenAI token counting, OpenAI model listing, Anthropic Messages, Google v1beta model listing, Google generate/stream/count, health, update check, and playground.
- It has a model alias registry with Gemini, OpenAI, Anthropic, and image-generation aliases.
- It has cookie parsing, SAPISIDHASH generation, AT/BL token extraction, cookie auto-discovery, cookie-pool loading, cooldown, and auto-reload logic.
- It has Gemini web RPC payload construction using a sparse array and `f.req` form encoding.
- It has streaming parsers for cumulative Gemini BoQ text.
- It has thinking/reasoning separation for OpenAI and Anthropic compatible streams.
- It has image compression and Scotty resumable upload logic.
- It has a static PWA/chat studio and Cloudflare/static edge-function files.
- It has unit and integration-style tests across major packages.
- It has installers, Docker files, service templates, diagnostics, benchmarks, and examples.

Not proven in this static pass:

- Current live Gemini upstream compatibility.
- Current live Google rate limits or direct API limits.
- Actual model availability for the named future/frontier aliases.
- Real authenticated Pro routing behavior with the local `cookie.txt`.
- Real Imagen/Nano Banana generation behavior.
- Real browser login success on this machine.
- Actual `go test`, `make build`, `make dist`, Docker build, diagnostics, or benchmark results on 2026-08-20.
- Public deployment status for `bob-gemini-free.abcsteps.com`.

## 24. High-Level Understanding Summary

BOB Gemini Free is a local AI gateway that presents OpenAI, Anthropic, Google Gemini, and embedded Go interfaces while routing actual generation through Google's Gemini web RPC protocol. It is built as a Go monolith with clean internal packages around configuration, model aliases, HTTP handling, protocol formatting, upstream Gemini transport, cookie/session management, multimodal upload, browser login, diagnostics, services, and updating. Its user experience is rounded out by static documentation, installers, examples, generated binaries, Docker support, and a local-first web studio.

The most critical modules are:

- `internal/server`: public HTTP API surface, middleware, telemetry, and protocol-specific handlers.
- `internal/gemini`: upstream Google web RPC client, cookie auth, payload builder, stream parser, cookie pool, TLS impersonation adapter.
- `internal/format`: translation between OpenAI/Anthropic/Google request schemas and Gemini prompt/text/tool/thinking representations.
- `internal/multimodal`: image compression, token scraping, Scotty upload, and remote image fetch guardrails.
- `pkg/gateway`: embeddable Go facade for in-process usage.

The build/run mechanism is straightforward: `make build` synchronizes the static web bundle and compiles `bob-gemini-free`; `./bob-gemini-free` starts the local server; `go test -count=1 ./...` is the documented local test command; `make dist` cross-compiles release binaries; Docker and OS service files provide operational packaging.

This completes Round 2 static comprehension. Any next step that claims runtime readiness, rate-limit behavior, provider compatibility, or live authenticated model access should be treated as a separate verification phase requiring explicit permission to run code and make network/provider calls.

## 25. Local Branch Implementation Addendum - codex/public-readiness-audit

After the static analysis phase, a local-only implementation branch was created:

- Branch: `codex/public-readiness-audit`
- Push status: not pushed
- Verification performed: `go test -count=1 ./...`, `go vet ./...`, `make build`, `make dist`, and `git diff --check` passed locally

This addendum records the current post-analysis code state so the earlier report remains historically preserved while the live branch state is unambiguous.

Implemented public-readiness fixes:

- OpenAI `image_url` content now preserves `http://` and `https://` remote image URLs as `format.Image.URL` values instead of silently ignoring them.
- Multimodal upload now fetches supported remote image URLs through the guarded multimodal fetch path, detects MIME from fetched bytes, and returns explicit errors on fetch/upload failure.
- Image-only OpenAI, Anthropic, and Google requests are no longer rejected as empty prompt requests when valid image attachments are present; a neutral default analysis prompt is supplied for upstream Gemini.
- Image upload errors now return protocol-compatible `502` API errors instead of silently degrading the request to text-only behavior.
- OpenAI Responses API multipart `input_image` / `image_url` inputs are now preserved, uploaded, and passed to Gemini through the same file-ref path as Chat Completions.
- OpenAI image generation no longer fabricates a placeholder success URL when upstream Gemini text contains no extractable generated image URL; it now returns a `502` API error with upstream text in `details`.
- Remote image fetch now rejects unsupported URL schemes, non-2xx HTTP responses, oversized responses, non-image response bodies, and spoofed `image/*` headers whose body is not image-detected before upload.
- `RetryAttempts` is normalized to a minimum of `1` after file/env config load and at server construction so a user-supplied zero cannot skip the upstream request loop and produce empty success-like behavior.
- Public hardcoded version markers in `Dockerfile`, `functions/health.js`, `web/functions/health.js`, and `web/sw.js` are normalized to `v0.1.5`.
- `web/index.html` is synced from `internal/server/playground.html`, preserving the recent upstream-rate-limit versus gateway-connection UI distinction in the static web bundle.
- A dedicated `docs/1-getting-started/classroom-lan-guide.md` now explains when to use the Cloudflare Pages studio and when to use the local LAN gateway for concentrated classroom traffic, including pre-class verification commands and proof boundaries.

Regression tests added:

- `internal/format/openai_test.go`: remote OpenAI image URL preservation.
- `internal/format/openai_test.go`: OpenAI Responses `input_image` preservation into the shared image extraction path.
- `internal/config/config_test.go`: retry-attempt clamping from env and config file.
- `internal/multimodal/multimodal_test.go`: non-2xx, non-image, and spoofed-header remote fetch rejection.
- `internal/server/server_test.go`: unsupported remote image URL rejection, Responses API unsupported image rejection, and image-generation no-placeholder error behavior.
- `pkg/gateway/gateway_test.go`: embedded gateway retry-attempt clamping.
- Documentation links added from `README.md`, `docs/README.md`, and `docs/1-getting-started/quickstart.md` to the classroom LAN guide.
- `docs/6-operations/live-verification-runbook.md` added to separate public deployment checks, local gateway checks, upstream generation checks, diagnostics, benchmarks, and release-build proof levels.

Remaining proof boundary after this implementation:

- The local Go test suite, vet, host build, and cross-compile release build pass, but live Gemini upstream behavior, authenticated image upload, real Imagen generation, direct API rate limits, Docker runtime, and public deployment behavior remain unproven in this branch.
- `CHANGELOG.md` historical references to earlier version transitions and updater tests that compare `v0.1.5` against `v0.1.4` remain intentional historical/test fixtures, not stale runtime version markers.

## 26. Live Surface Check - 2026-08-20

Non-invasive public live checks were run against `https://bob-gemini-free.abcsteps.com`:

- `GET /` returned HTTP `200` with a non-empty HTML body.
- `GET /health` returned HTTP `200`, but reported `"version":"v0.1.4"`.
- `GET /v1/models` returned HTTP `200` with an OpenAI-compatible model list.
- `GET /sw.js` reported `CACHE_NAME = 'bob-gemini-studio-v0.1.4'`.
- The live HTML contains visible `v0.1.5` UI labels, but did not contain the current branch's `API Rate Limit / Upstream Error` copy.

Conclusion: the public site is live, but the public deployment is not fully aligned with this local branch. The branch has `functions/health.js`, `web/functions/health.js`, `web/sw.js`, and `web/index.html` aligned to the newer local state, but the live Cloudflare deployment still appears stale or partially rebuilt.

Local non-provider smoke checks were also run against the built binary on `127.0.0.1:19610`:

- `GET /` returned local health JSON with `"version":"v0.1.5"`.
- `GET /v1/models` returned a non-empty OpenAI-compatible model list.
- `GET /v1beta/models` returned a non-empty Google-compatible model list.

These local checks prove the built binary starts and serves local endpoints. They do not prove live Gemini generation, authenticated multimodal upload, Imagen, rate-limit behavior, or classroom LAN throughput.
