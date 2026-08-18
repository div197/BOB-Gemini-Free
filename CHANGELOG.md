# Changelog

All notable changes to **BOB Gemini Free** (*Break Ordinary Boundaries*) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
- **Native Anthropic Messages API Engine (`/v1/messages`)**: Direct drop-in support for **Claude Code CLI** (`ANTHROPIC_BASE_URL=http://127.0.0.1:8081`) and Anthropic SDKs with complete SSE event streaming (`message_start`, `content_block_delta`, `message_delta`, `message_stop`).
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
