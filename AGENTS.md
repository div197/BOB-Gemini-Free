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
├── go.mod / go.sum             # Go module definition (Go 1.22+)
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
```

### Testing & Verification
```bash
# Run all unit and integration tests across all packages
go test -count=1 ./...

# Run built-in diagnostic test suite against running gateway
./bob-gemini-free --test --test-url http://127.0.0.1:8081

# Query live status, uptime, requests, and financial savings
./bob-gemini-free --status --test-url http://127.0.0.1:8081

# Run concurrency benchmark
./bob-gemini-free --bench --bench-concurrency 3 --bench-requests 6
```

### Code Formatting
```bash
# Format all Go source files according to standard Go conventions
gofmt -s -w .
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
- **Base URL**: `http://127.0.0.1:8081/v1`
- **API Key**: `none` (or your configured `api_keys`)
- **Model**: `gemini-3.7-flash` or `gemini-3.7-flash-thinking`

### Claude Code CLI (Anthropic Protocol)
```bash
export ANTHROPIC_BASE_URL=http://127.0.0.1:8081
export ANTHROPIC_API_KEY=none
claude
```

### OpenAI Codex CLI (Responses API)
```bash
export OPENAI_BASE_URL=http://127.0.0.1:8081/v1
export OPENAI_API_KEY=none
codex
```

---

## 6. Cookie Authentication & Multi-Account Guidelines

For AI agents managing or configuring authenticated Pro routing (`gemini-3.1-pro` / `gemini-pro`) and Imagen 3 synthesis:

### Session Extraction
1. **1-Click Interactive Sign-In Window (Recommended)**:
   ```bash
   ./bob-gemini-free --login
   ```
   Opens a standalone window $\rightarrow$ user logs in once $\rightarrow$ automatically extracts cookies via CDP and saves `./cookie.txt`.
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
  -p 8081:8081 \
  -v $(pwd)/cookie.txt:/app/cookie.txt:ro \
  -e BOB_GEMINI_FREE_COOKIE_FILE=/app/cookie.txt \
  bob-gemini-free:local
```

