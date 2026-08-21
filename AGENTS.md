# AGENTS.md — AI Agent Guidelines for BOB Gemini Free

Welcome, AI Agent! This document provides technical architecture, build instructions, directory layouts, and execution workflows for automated agents (Cursor, Windsurf, Claude Code, Codex CLI, Aider, OpenHands, Devin, Roo Code) interacting with the **BOB Gemini Free** codebase.

---

## 1. Project Overview

**BOB Gemini Free** (*Break Ordinary Boundaries*) by **ABCsteps** ([abcsteps.com](https://abcsteps.com/)) and **Divyanshu Singh Chouhan** ([@div197](https://github.com/div197)) is a high-performance, single-binary local AI gateway written in Go.

It translates three major protocol standards into Google's internal web RPC protocol (`batchexecute` / `streamGenerateContent`):
1. **OpenAI Standard**: `POST /v1/chat/completions`, `POST /v1/responses` (with `output_text`), `POST /v1/tokens/count`, `GET /v1/models`, `GET /v1/models/{model}`, `POST /v1/images/generations`
2. **Anthropic Standard**: `POST /v1/messages` (Claude Code CLI with complete SSE stream lifecycle and prompt caching counters)
3. **Google Gemini Standard**: `GET /v1beta/models`, `POST /v1beta/models/{target}` (including `generateContent`, `streamGenerateContent`, and `:countTokens`)
4. **Embedded Go Library**: `pkg/gateway` (`import "github.com/div197/bob-gemini-free/pkg/gateway"`)

---

## 2. Directory & Package Structure

```
.
├── main.go                     # Entrypoint & CLI flag routing
├── go.mod / go.sum             # Go module definition (Go 1.26.5 in this snapshot)
├── Makefile                    # Multi-arch compilation & test commands
├── install.sh / install.ps1    # Cross-platform automated setup scripts
├── test-kit.sh / test-kit.ps1  # Automated diagnostic runners
├── pkg/
│   └── gateway/                # Exported embedded Go library (pkg/gateway)
└── internal/
    ├── browser/                # 1-click native interactive login window (CDP WebSocket)
    ├── config/                 # Configuration loader & cookie auto-discovery
    ├── diag/                   # 13-point diagnostic test kit, bench runner & e2e tests
    ├── format/                 # Protocol translation (OpenAI, Anthropic, Google, Images, Citations, Tokens)
    ├── gemini/                 # Upstream client, auth, cookie pool load-balancer, parser
    ├── models/                 # Model registry & alias catalog (Imagen 3, Nano Banana 2/Pro)
    ├── multimodal/             # Image compression, Scotty upload, token cache, dynamic refresh
    └── server/                 # HTTP mux, middleware, SSE streamer, handlers
```

---

## 3. Standard Development Commands

### Building
```bash
# Build local host binary
make build

# Cross-compile for all platforms (macOS ARM/Intel, Linux AMD64/ARM64, Windows)
make dist

# Build Native Desktop App (Requires Wails CLI & CGO)
make desktop
```

### Testing & Verification
```bash
# Run all unit and integration tests across all packages
go test -count=1 ./...

# Run built-in diagnostic test suite against running gateway
./bob-gemini-free --test --test-url http://127.0.0.1:9610

# Query live status, uptime, requests, and financial savings
./bob-gemini-free --status --test-url http://127.0.0.1:9610

# Run concurrency benchmark
./bob-gemini-free --bench --bench-concurrency 3 --bench-requests 6
```

# Format all Go source files according to standard Go conventions
gofmt -s -w .
```

### Automation & Background Service Operations
```bash
# In-place auto-update to latest GitHub Release
./bob-gemini-free --update

# Register and start 24/7 background OS daemon across reboots
./bob-gemini-free service install [--port 9610]

# Check background daemon health and service definition
./bob-gemini-free service status

# Start / Stop / Uninstall daemon
./bob-gemini-free service start
./bob-gemini-free service stop
./bob-gemini-free service uninstall
```

---

## 4. Coding & Architecture Conventions

- **Pure Standard Library & Minimal Dependencies**: Zero heavy runtime frameworks. Handlers use standard `http.Handler` and `net/http.ServeMux`.
- **Security & Privacy First**:
  - The default listening interface is `127.0.0.1`.
  - Credentials (`cookie.txt`, `config.json`) are locked with `0600` permissions and included in `.gitignore`.
- **Cross-Platform Compatibility**:
  - Always use `filepath.Join` and `os.UserHomeDir()` for file system access to maintain identical behavior across macOS, Linux, and Windows.
- **Thinking / Reasoning Handling**:
  - Step-by-step reasoning tokens must be extracted to `reasoning_content` for OpenAI clients and preserved in reasoning blocks for Anthropic clients.
- **Error Handling**:
  - Upstream Google API errors must be mapped to standard HTTP error envelopes (`type: "invalid_request_error"` or `"api_error"`).

---

## 5. Connecting AI Agents to BOB Gemini Free

### Cursor / Windsurf / Continue / Roo Code (OpenAI Protocol)
- **Base URL**: `http://127.0.0.1:9610/v1`
- **API Key**: `none` (or your configured `api_keys`)
- **Model**: `gemini-3.7-flash` or `gemini-3.7-flash-thinking`

### Claude Code CLI (Anthropic Protocol)
```bash
export ANTHROPIC_BASE_URL=http://127.0.0.1:9610
export ANTHROPIC_API_KEY=none
claude
```

### OpenAI Codex CLI (Responses API)
```bash
export OPENAI_BASE_URL=http://127.0.0.1:9610/v1
export OPENAI_API_KEY=none
codex
```

---

## 6. Cookie Authentication & Multi-Account Guidelines

For AI agents managing or configuring authenticated Pro routing (`gemini-3.1-pro` / `gemini-pro`), Imagen 3 synthesis, and **Multimodal Image Vision Analysis**:

### Session Requirements & Capability Matrix
| Capability | Anonymous / Guest Session | Authenticated Google Session (`./cookie.txt` via `--login`) |
| :--- | :--- | :--- |
| **Text Chat & Coding** (`gemini-3.7-flash`, `gemini-3.6-flash`) | Upstream/session-dependent | Upstream/session-dependent; no fixed context or speed guarantee |
| **Deep Step-by-Step Reasoning** (`gemini-3.7-flash-thinking`) | Upstream/session-dependent | Upstream/session-dependent; local reasoning splitter is fixture-tested |
| **Multimodal Image Analysis (Vision)** (Diagrams, Screenshots, OCR) | May fail without an authenticated session | Authenticated session may permit it; live capability is unverified |
| **Imagen 3 Image Synthesis** (`imagen-3`) | May fail without an authenticated session | Authenticated session may permit it; route is experimental/upstream-dependent |
| **Pro Model Routing** (`gemini-3.1-pro` / `gemini-pro`) | May fall back or fail | Authenticated session may permit it; model identity is unverified |

### Why Images Require Session Binding (`BardErrorInfo 1003`)
Image attachments are uploaded to Google Scotty storage (`content-push.googleapis.com/upload/`). Google requires the storage bucket and downstream `StreamGenerate` RPC call to be signed by an active session with valid `SAPISIDHASH` and `__Secure-1PSIDTS`. Unauthenticated attempts will fail with `BardErrorInfo [1003]`.

### Session Extraction
1. **1-Click Interactive Sign-In Window (Recommended)**:
   ```bash
   ./bob-gemini-free --login
   ```
   Opens a standalone window $\rightarrow$ user logs in once $\rightarrow$ attempts to extract cookies via CDP and saves `./cookie.txt` when the required tokens are available.
2. **Manual DevTools Copy**:
   - **Instant**: Open `gemini.google.com/app` $\rightarrow$ DevTools Network tab $\rightarrow$ click `app?eom=1...` (or `batchexecute`) $\rightarrow$ copy `Cookie:` from Request Headers.
   - **Chat**: Send `"hi"` in Gemini web $\rightarrow$ click `StreamGenerate` in Network list $\rightarrow$ copy `Cookie:`.
   - Save via `./bob-gemini-free --setup-cookie` or write directly to `./cookie.txt` (`chmod 600`).

### Multi-Account Profiles (`auth_user`)
- Primary profile (`https://gemini.google.com/app`): `auth_user: "0"` (default).
- Secondary profile (`https://gemini.google.com/u/1/app`): `auth_user: "1"`.
- Set via `"auth_user": "1"` in `config.json` or CLI flag.

### Docker & OrbStack Container Usage
```bash
docker run -d \
  --name bob-gemini-free \
  -p 9610:9610 \
  -v $(pwd)/cookie.txt:/app/cookie.txt:ro \
  -e BOB_GEMINI_FREE_COOKIE_FILE=/app/cookie.txt \
  bob-gemini-free:local
```

### Complete Environment Variable Reference

All configuration fields can be set without mounting a `config.json` file:

| Environment Variable | Type | Default | Description |
|---|---|---|---|
| `BOB_GEMINI_FREE_HOST` | string | `127.0.0.1` | Binding host interface |
| `BOB_GEMINI_FREE_PORT` | int | `9610` | Listening port |
| `BOB_GEMINI_FREE_COOKIE_FILE` | string | `./cookie.txt` | Path to primary session cookie file |
| `BOB_GEMINI_FREE_COOKIE_POOL` | string | `` | Comma-separated pool of cookie files |
| `BOB_GEMINI_FREE_COOKIE_POOL_DIR` | string | `` | Directory of `*.txt` cookie files for auto-discovery |
| `BOB_GEMINI_FREE_DEFAULT_MODEL` | string | `gemini-3.6-flash` | Default model when none is specified in request |
| `BOB_GEMINI_FREE_AUTH_USER` | string | `""` | Multi-account index (`"0"`, `"1"`, etc.) |
| `BOB_GEMINI_FREE_API_KEYS` | string | `` | Comma-separated authorized API keys |
| `BOB_GEMINI_FREE_PROXY` | string | `` | HTTP/SOCKS5 proxy URL |
| `BOB_GEMINI_FREE_IMPERSONATE` | string | `` | TLS fingerprint profile (`chrome`, `firefox`, `safari`) |
| `BOB_GEMINI_FREE_LOG_REQUESTS` | bool | `false` | Enable request lifecycle logging (`true`/`1`/`yes`) |
| `BOB_GEMINI_FREE_RETRY_ATTEMPTS` | int | `3` | Max upstream retry attempts per request |
| `BOB_GEMINI_FREE_RETRY_DELAY_SEC` | int | `2` | Seconds between retry attempts |
| `BOB_GEMINI_FREE_REQUEST_TIMEOUT_SEC` | int | `180` | Per-request upstream timeout in seconds |
