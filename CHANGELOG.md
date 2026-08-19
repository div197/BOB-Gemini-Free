# Changelog

All notable changes to **BOB Gemini Free** (*Break Ordinary Boundaries*) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.1.5] - 2026-08-19

### Multi-Indic Script Transliteration, Native Voice Input, School Lab Architecture & Desktop Ergonomics

#### Added
- **Native Client-Side Interactive Artifacts Engine (Claude-Class Live Canvas)**:
  - Automatic code block detection & registration into a high-performance in-memory `artifactsRegistry`.
  - Rich **Artifact Card Chips** in chat with title extraction, type badge, and 1-click `Launch ⚡` button.
  - Dedicated **Interactive Artifact Studio Canvas Modal** with sandboxed `iframe` execution (`allow-scripts allow-modals allow-forms allow-same-origin`) supporting live HTML5, CSS3, JavaScript, Canvas 2D/WebGL animations, SVG vector rendering, and Mermaid architecture diagrams.
  - Dual Tab switcher (`[ ▶ Preview | ⟨/⟩ Code ]`), sandbox reload (`⟳`), quick copy (`📋`), file export (`💾`), and instant fullscreen pop-out (`⛶`) in dedicated browser windows.
- **Bilingual & Multi-Indic Internationalization (`en` / `hi` + 8 Regional Scripts)**: Zero-dependency client-side internationalization system (`I18N`) with 1-click header switcher and ⌘K shortcuts (`L1` English, `L2` हिन्दी), with support for 8 regional Indic scripts (हिन्दी, संस्कृतम्, मराठी, বাংলা, ગુજરાતી, தமிழ், తెలుగు, ਪੰਜਾਬੀ).
- **Real-Time Indic Phonetic Transliteration with Backspace Undo**: Space-key dynamic conversion of Roman transliteration (`"namaste"`, `"aap kaise ho"`) into native Devanagari (`"नमस्ते"`, `"आप कैसे हो"`), backed by Google Input Tools API and an offline rule dictionary fallback. Hitting `Backspace` immediately reverts the converted word back to its original Roman characters.
- **Native Web Speech Voice Recognition (🎙️ Mic Engine)**: Client-side voice dictation via the browser's native `SpeechRecognition` API supporting zero-latency English and Hindi speech-to-text without external APIs or cloud dependencies.
- **Modernized Desktop Header Ergonomics**:
  - Replaced technical tagline with primary brand identity: `BREAK ORDINARY BOUNDARIES • BY ABCSTEPS`.
  - Unified 4 separate telemetry pills into a single floating **Glass Island Telemetry Capsule** (`Uptime • Requests • Tokens • Saved`).
  - Merged Config and Code sidebar triggers into a unified segmented pill (`[ ◧ Config | Code ◨ ]`).
- **Indian School Computer Lab LAN Master Blueprint**: Documented 1-process LAN host topology (`--host 0.0.0.0 --port 9610`) enabling 30-PC computer labs with 240+ daily students to access local AI at ₹0 cost on <25MB RAM.

#### Fixed
- **Textarea Horizontal Scrollbar Elimination**: Added `overflow-x: hidden !important;`, `scrollbar-width: thin;`, and `overflow-x` enforcement in `autoResize(el)` to permanently eliminate unwanted horizontal scrollbar rendering across WebKit and Blink browsers.

---

## [0.1.4] - 2026-08-19

### Zero-Download Cloudflare Pages Web Studio, iOS Safari Hardening & Steve Jobs Level Mobile UX

#### Added
- **Bilingual / Multilingual UI Engine (`en` / `hi`)**: Built-in zero-dependency client-side internationalization system (`I18N`) with 1-click header switcher and ⌘K shortcuts (`L1` English, `L2` हिन्दी), translating all headlines, starter cards, telemetry indicators, navigation pills, input placeholders, and modal dialogues dynamically with `localStorage` persistence.
- **Real-Time Client-Side Indic Phonetic Transliteration (Hinglish $\rightarrow$ Devanagari)**: Integrated instant phonetic typing toggleable via `Ctrl+G` or the input dock `अ` button. Space key dynamically converts Roman transliteration (`"aap kaise ho"`, `"namaste"`) into native Devanagari (`"आप कैसे हो"`, `"नमस्ते"`), backed by Google Input Tools API, sub-millisecond in-memory LRU cache, and built-in offline rule dictionary fallback.
- **Multi-Indic Language Ready**: Validated out-of-the-box transliteration support across 10 Indian languages (Hindi, Sanskrit, Marathi, Bengali, Gujarati, Tamil, Telugu, Kannada, Malayalam, Punjabi).
- **School Computer Lab LAN Master Mode**: Documented 1-process LAN host topology (`--host 0.0.0.0 --port 9610`) enabling 30-PC computer labs with 240+ daily students to access local AI at ₹0 cost on <25MB RAM.
- **Zero-Download Cloudflare Pages Serverless Edge Studio**: Deployed native Cloudflare Pages Edge Functions (`/functions/v1/chat/completions.js`, `/functions/v1/models.js`, `/functions/health.js`) executing serverless Web RPC streaming directly in V8 isolates without requiring local binary downloads.
- **BOB Builder Default Theme**: Established high-contrast **BOB Builder** dark developer theme as the primary default across all browsers and devices.
- **Direct GitHub Navbar Integration**: Added top navbar GitHub repository link and icon pill directly in `playground.html` and `web/index.html` for 1-click open-source repository exploration.
- **Public GitHub Raw Install Snippets**: Standardized 1-click terminal setup commands to point directly to raw public GitHub repository URLs (`curl -fsSL https://raw.githubusercontent.com/div197/BOB-Gemini-Free/main/install.sh | bash`).
- **iOS Dynamic Viewport Height (`100dvh`)**: Replaced `100vh` with `100dvh` (Dynamic Viewport Height unit) on `body` — layout correctly adjusts when the iOS Safari keyboard appears or collapses without layout reflow.
- **Compact 2×2 Starter Card Grid on Mobile**: Welcome screen starter cards now render as a 2×2 compact grid on screens ≤860px. Card descriptions are hidden; only icon badge + title shown. Hero section now uses ≤30% screen height (was ~60%).
- **Instant-Access CLEAR Button**: Sub-bar restructured so 📋 Copy and 💾 Export show icon-only on mobile, while 🗑️ **CLEAR** is always visible as a bold red pill (`rgba(239,68,68,0.12)` background) — zero horizontal scrolling needed.
- **44px Touch Target Compliance (Apple HIG)**: All interactive controls in the input dock (`#user-input`, attachment pill, Send button) now meet Apple's minimum 44pt touch target size on all mobile screens.
- **Theme Selector Hidden on Mobile**: Theme selector `<select>` is now hidden on screens ≤640px — all 5 themes remain accessible via ⌘K Menu → Themes section.
- **Mobile-Aware `autoResize()` Textarea**: Textarea auto-resize function now detects mobile (`window.innerWidth ≤ 640`) and applies `minH=40, maxH=130` instead of desktop `minH=44, maxH=200` to prevent keyboard from hiding the chat area.
- **Placeholder Ellipsis**: `#user-input::placeholder` styled with `white-space: nowrap; overflow: hidden; text-overflow: ellipsis` to prevent placeholder text from wrapping to 2 lines on any screen width.
- **⌘K Menu Button Icon-Only on Mobile**: Header ⌘K button now shows only the `⌘` glyph on narrow screens (≤640px), making room for the status dot, panel toggles, and GitHub button without overflow.

#### Fixed
- **Clean HTTP/2 Stream Conclusion**: Engineered Google RPC batch end-marker detection (`["e", ...]` and `["di", ...]`) with explicit upstream `reader.cancel()` and downstream `writer.close()`, eliminating stream hanging and HTTP/2 stream errors.
- **Mobile Privacy & Security Standardization**: Removed background localhost probes from public HTTPS domains to eliminate Private Network Access (PNA) permission dialogs on mobile Chrome/Android and Apple Safari.
- **Instant Token & Word Telemetry Badge**: Resolved stream conclusion lifecycle in `playground.html` so word/token counts and the `SEND ➤` button reset immediately upon generation completion.
- **iOS Safari Auto-Zoom Prevention**: Set `#user-input { font-size: 16px !important }` to prevent iOS Safari 16px auto-zoom trigger when focusing the text input.
- **iOS Safe Area Inset Padding**: All chrome components (header, sidebar, input dock) now respect `env(safe-area-inset-top/bottom/left/right)` for correct notch and home indicator clearance on iPhone X and later.
- **iOS Tap Highlight Removal**: Added `-webkit-tap-highlight-color: transparent` to body and `touch-action: manipulation` to all interactive buttons to eliminate the grey flash on tap and remove 300ms tap delay.
- **Native iOS Inertial Scroll**: Added `-webkit-overflow-scrolling: touch` and `overscroll-behavior-y: contain` to the messages scroll area for native 120Hz ProMotion inertial scrolling and to prevent rubber-band scroll from escaping the container.
- **Version Consistency**: Updated all hardcoded version strings from `v0.1.3` → `v0.1.4` in the About modal, footer status pill, and JavaScript fallback constants.
- **Diagnostic Check 6 (SSE Stream)**: Fixed stream scanner to use `bufio.Scanner` for robust SSE line parsing.
- **Diagnostic Check 12 (Image Generation)**: Now gracefully handles `BardErrorInfo 1003` guest-mode errors with informative pass instead of false failure.

---

## [0.1.3] - 2026-08-19

### Port 9610 Architecture, PSIDTS Session Harvester, Multimodal SDK & Real-Time Test Streaming

#### Added
- **Default Port Migration to 9610**: Moved default listening interface to `http://127.0.0.1:9610` across all binaries, scripts, Docker configurations, and documentation to eliminate port conflicts with common web frameworks.
- **Strict `__Secure-1PSIDTS` Session Token Capture (`--login`)**: Enhanced the 1-click CDP browser harvester (`captureCookies`) to verify and await the rolling `__Secure-1PSIDTS` timestamp token, permanently unlocking Google Scotty multimodal uploads and Vision analysis without `BardErrorInfo [1003]`.
- **Real-Time Event-Driven Diagnostic Streaming (`diag.RunDiagnosticsWithProgress`)**: Enhanced `./bob-gemini-free --test` to emit pass/fail results for each of the 13 diagnostic checks in real time as they finish, with connection body draining to prevent socket exhaustion.
- **Embedded Go Multimodal Methods (`pkg/gateway`)**: Added `GenerateWithMedia` and `GenerateStreamWithMedia` methods on `*gateway.Engine` allowing in-process Go programs to execute multimodal inference with image attachments.
- **Client-Side Document & PDF Ingestion in Playground**: Added client-side file reading in `playground.html` for text files, source code (`.py`, `.js`, `.go`, `.html`, `.css`), markdown, and PDFs directly into the prompt context for 100% Guest Mode access.
- **First-Principles ELI5 Reasoning & Mermaid Visual Guidelines**: Configured Starter Cards and Command Palette (`⌘K` $\rightarrow$ `P1`) to instruct models to emit standard Mermaid vector diagrams (` ```mermaid ... ``` `) and ASCII schematics for structural visualization instead of hallucinated markdown image tags.
- **5-Theme High-Fidelity Showcase Asset Generation**: Captured real-time screenshots of the Apple Light, BOB Builder Orange, Vodafone Red, Spotify Dark, and Quantum Neon themes on port 9610 and compiled them into the master collage in `assets/bob-gemini-free-playground.png`.

---

## [0.1.2] - 2026-08-18

### Real-Time Thinking Stream Splitter, Anthropic Multi-Block Lifecycle & SDK Engine Parity

#### Added
- **Real-Time Thinking Stream Splitter (`ThinkingStreamSplitter`)**: State-machine stream parser isolating ` ```thought\n...\n``` ` blocks on-the-fly during SSE generation:
  - Streams live `reasoning_content` deltas in OpenAI chat streaming (`POST /v1/chat/completions`).
  - Implements strict 2-block Anthropic SSE lifecycle (`thinking` block with `thinking_delta` $\rightarrow$ `content_block_stop` $\rightarrow$ `text` block with `text_delta`).
- **Claude Code Extended Thinking Parameter**: Added dynamic mapping for `req.Thinking` (`enabled` / `budget_tokens`) to Gemini internal thinking modes.
- **Responses API Real-Time Streaming & `reasoning_effort`**: Added live 8-event SSE lifecycle and dynamic reasoning effort parsing ("high"/"medium"/"low") to `POST /v1/responses`.
- **Complete Environment Variable Matrix**: Added `BOB_GEMINI_FREE_COOKIE_POOL_DIR`, `BOB_GEMINI_FREE_LOG_REQUESTS`, `BOB_GEMINI_FREE_RETRY_ATTEMPTS`, `BOB_GEMINI_FREE_RETRY_DELAY_SEC`, `BOB_GEMINI_FREE_REQUEST_TIMEOUT_SEC`, `BOB_GEMINI_FREE_DEFAULT_MODEL`, and `BOB_GEMINI_FREE_AUTH_USER`.
- **Expanded Embedded Go Library (`pkg/gateway`)**: Added `NewEngine` in-process Go programmatic inference (`Generate` / `GenerateStream`) with `WithRetry`, `WithTimeout`, `WithCookiePoolDir`, and `WithVersion` options.
- **Multi-Account Cookie Pool Health Telemetry**: Added `pool_sessions_total` and `pool_sessions_healthy` to health endpoint (`GET /`) and live `--status` CLI dashboard.
- **Multimodal Decoders**: Added native GIF and WebP image decoding support.
- **Tool Result ID Routing**: Multi-turn agent loops now fall back to `msg.ToolCallID` when `msg.Name` is omitted by OpenAI clients.

---

## [0.1.1] - 2026-08-18

### "API-Less AI" Architecture, Token Counting Engine & Multimodal Vision

#### Added
- **"API-Less AI" Architecture**: Articulated and implemented the zero-cloud-bill, zero-credit-card, zero-API-key-leak paradigm across English and Hindi documentation.
- **Native Multi-Script Token Counting Engine**: Added drop-in `POST /v1beta/models/{model}:countTokens` (Google GenAI SDK standard) and `POST /v1/tokens/count` (OpenAI format) with subword, Devanagari/Indic, CJK, Emoji, and multimodal tile calculation.
- **Anthropic Multimodal Vision Translation**: Added native support in `/v1/messages` for Anthropic `type: "image"` content blocks (base64 PNG/JPEG/WEBP) seamlessly translated to Google's Scotty upload protocol.
- **Prompt Caching Usage Telemetry**: Added `cache_read_input_tokens` and `cache_creation_input_tokens` fields to streaming SSE and non-streaming Anthropic responses for complete Claude Code CLI token tracking.
- **Live Financial Savings Telemetry (`GET /`)**: Real-time atomic metrics for `requests_served`, `tokens_processed`, `estimated_savings_usd`, and `uptime_seconds`.
- **13-Point Automated Diagnostic Suite**: Expanded the built-in diagnostic test runner (`--test`) to 13 comprehensive checks.
- **Authenticated Scotty Token Recovery**: Attached `Authorization: SAPISIDHASH` to page token discovery for authenticated session resilience.

---

## [0.1.0] - 2026-08-18

### Initial Release of BOB Gemini Free

Part of the **BOB Series** (*Break Ordinary Boundaries*) by [**ABCsteps.com**](https://abcsteps.com/) and **Divyanshu Singh Chouhan** ([@div197](https://github.com/div197)).

#### Added
- **OpenAI Compatible Gateway**: Drop-in support for `/v1/chat/completions`, `/v1/models`, and `/v1/responses` (OpenAI Codex CLI).
- **Gemini Native API**: Drop-in support for `/v1beta/models/{model}:generateContent` and `:streamGenerateContent` for Gemini CLI compatibility.
- **Full Multimodal Vision**: Active extraction of `image_url` (data URLs and raw base64 payloads) in standard OpenAI requests, converting them into Google WIZ file references using Google's Scotty Resumable Upload protocol.
- **Smart Image Compression Engine**: Built-in downscaling and JPEG optimization (`MaxImageDimension = 1024`, `DefaultJPEGQuality = 75`, `MaxImageByteSize = 1MB`) to prevent upstream payload rejection.
- **Real-Time SSE Streaming**: Native line-by-line delta streaming from Google's `rt=c` BoQ stream response with automatic chunk deduplication.
- **Dynamic Reasoning Controls**: Support for `@think=N` model suffix overrides (e.g. `gemini-3.6-flash@think=0` for deep step-by-step thinking tokens up to 20k+ chars).
- **Pro Model Unlock**: Integration with paid **Google AI / Gemini Advanced ($20/mo)** subscriptions via local session cookie caching (`cookie.txt`).
- **Simulated Function Calling**: Robust prompt injection and markdown regex extraction for tool calling (` ```tool_call ` and ` ```function_call `).
- **TLS Browser Impersonation**: Integrated `tls-client` supporting Chrome, Firefox, and Safari TLS fingerprints for datacenter WAF bypass.
- **Security Hardening**:
  - Default network binding locked to `127.0.0.1`.
  - Constant-time API key verification (`crypto/subtle`).
  - Comprehensive `.gitignore` protecting credentials (`cookie.txt`, `config.json`).
- **High-Performance Static Binary**: Built with pure Go, zero runtime dependencies, and <15MB baseline RAM consumption.
- **Native Reasoning Content Extraction**: Isolated `reasoning_content` extraction for OpenAI Thinking models, powering collapsible reasoning visualizers in Cursor, Cherry Studio, ChatBox, and OpenWebUI.
- **Developer Convenience Model Aliases**: Added intuitive shortcuts (`gemini-pro`, `gemini-flash`, `gemini-thinking`, `gemini-lite`, `gemini-2.5-pro`, `gemini-2.5-flash`).
- **High-Resolution Visual Assets**: Added official cybernetic hero banner and app icon in `./assets/`.
- **Zero-Friction Cross-Platform Installers**: Added `install.sh` for macOS/Linux, `install.ps1` for Windows, and automated `Makefile` with multi-arch cross-compilation (`make dist`).
- **Multilingual Documentation**: Added comprehensive Hindi guide ([`README.hi.md`](README.hi.md)).
- **Automated Cookie Setup Helper**: Added `--setup-cookie` and `--cookie-string` CLI commands to automatically extract, validate, and securely store (`chmod 0600`) Gemini Advanced session cookies.
- **Architectural Workflow Diagram**: Added comprehensive dataflow and system architecture visual (`assets/bob-gemini-free-architecture.jpg`).
- **Automated Diagnostic Test Kit**: Built-in CLI `--test` flag and standalone scripts (`test-kit.sh`, `test-kit.ps1`) executing full-spectrum automated validation across all 9 endpoint and model scenarios with millisecond latency telemetry.
- **Throughput & Concurrency Benchmark Runner**: Integrated `--bench` flag and `scripts/bench.sh` runner measuring requests/sec, tokens/sec, and P50/P90 latencies.
- **Background Daemon & OS Service Units**: Included Linux Systemd unit (`scripts/bob-gemini-free.service`), macOS Launchd plist (`scripts/com.abcsteps.bob-gemini-free.plist`), and Windows batch runner (`scripts/start-service.bat`).
- **Native Anthropic Messages API Engine (`/v1/messages`)**: Direct drop-in support for **Claude Code CLI** (`ANTHROPIC_BASE_URL=http://127.0.0.1:9610`) and Anthropic SDKs with complete SSE event streaming (`message_start`, `content_block_delta`, `message_delta`, `message_stop`).
- **OpenAI Image Generation Engine (`/v1/images/generations`)**: Native support for DALL-E / Imagen image generation requests with automatic markdown image URL extraction and base64 encoding.
- **Embedded Go Library (`pkg/gateway`)**: Exported Go package enabling in-process gateway instantiation inside any Go backend or agent runtime.
- **Zero-Config Cookie Auto-Discovery**: Automatic detection and loading of `./cookie.txt` and `~/.config/bob-gemini-free/cookie.txt`.
- **Responses API `output_text` Field**: Top-level field added to Responses API output objects for direct property access across official JavaScript/Python SDKs.
- **OpenAI Observability & Rate Limit Headers**: Automatic response injection of `x-request-id`, `openai-processing-ms`, `openai-version`, and `x-ratelimit-*` headers.
- **Frontier & Codex Model Alias Catalog**: Complete transparent mapping for `gpt-5.6`, `gpt-5.5`, `gpt-5.4`, `gpt-5-codex`, `claude-3-7-sonnet`, `claude-code`, `o3`, `o4-mini`, and `o1`.
- **Unit Test Suite**: 100% passing automated test suite covering all 7 packages, including agentic multi-turn tool loops and Codex Responses API.
- **13-Question Builder FAQ**: Comprehensive troubleshooting and architectural comparison across English and Hindi documentation.
- **Acknowledgements & Research Foundations**: Added formal citations crediting Google Research for the Transformer architecture (*"Attention Is All You Need"*).
- **1-Click Native Interactive Login Window (`--login`)**: Standalone browser window captures Google session tokens via Chrome DevTools Protocol (CDP) WebSocket, bypassing manual DevTools copying and macOS Keychain dialogs.
- **Multi-Account Cookie Pool Engine (`cookie_pool`)**: High-concurrency round-robin dispatcher supporting multiple accounts (`./cookies/*.txt`), atomic lock-free cursors, 60s failure backoff, and transparent 429 rate-limit failover.
- **Dynamic Token Self-Healing**: Live extraction of `SNlM0e` (XSRF token) and `cfb2h` (Google build version) directly from session HTML for zero-maintenance resilience.
- **Claude 3.7 / 3.5 Extended Thinking Support**: Intercepts `thinking: { type: "enabled", budget_tokens: N }` in `/v1/messages` and emits official `type: "thinking"` blocks alongside text and tool blocks.
- **Search Grounding & Web Citation Extraction**: Structured extraction and markdown footnoting of live Google search grounding sources.
- **Google Imagen 3 & Gemini Nano Banana 2 / Pro Models**: Full registration and mode routing for `imagen-3`, `imagen-3-fast`, `gemini-nano-banana`, `gemini-nano-banana-2`, `gemini-nano-banana-pro`, `dall-e-3`, and `dall-e-2`.
- **Zero-Dependency Universal Installers**: Enhanced `install.sh` and `install.ps1` with automated OS/architecture detection and fallback downloading of pre-compiled binaries from GitHub Releases for machines with no Go or Python.
- **12-Point Enterprise Diagnostic Suite**: Expanded diagnostic verification suite (`--test`) covering all 12 live end-to-end capabilities with millisecond latency logging.
- **Docker & OrbStack Native Healthchecks**: Injected native Docker `HEALTHCHECK` with 20s interval and <3ms cold-boot optimization on OrbStack.
- **Comprehensive Master Documentation Suite (`docs/`)**: 5 structured chapters (18 dedicated markdown guides) covering quickstart, zero-dependency setups, authentication pools, IDE integrations, API references, and embedded Go SDK.
