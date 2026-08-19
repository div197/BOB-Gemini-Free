# CLAUDE.md — Claude Code Instructions for BOB Gemini Free

This repository contains **BOB Gemini Free** (*Break Ordinary Boundaries*) by **ABCsteps** ([abcsteps.com](https://abcsteps.com/)) and **Divyanshu Singh Chouhan** ([@div197](https://github.com/div197)).

---

## 1. Quick Build & Test Commands

```bash
# Build host binary
make build

# Cross-compile for all operating systems (macOS, Linux, Windows)
make dist

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

- **Pure Go Gateway**: Zero external runtime dependencies. Runs as a lightweight single binary consuming <15MB RAM.
- **Universal 3-in-1 Gateway**:
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

