# Getting Started: Quickstart Guide

Welcome to **BOB Gemini Free** (*Break Ordinary Boundaries*) by **ABCsteps** ([abcsteps.com](https://abcsteps.com/)) and **Divyanshu Singh Chouhan** ([@div197](https://github.com/div197)).

BOB Gemini Free is a high-performance, single-binary local AI gateway written in Go that translates standard **OpenAI**, **Anthropic Messages**, and **Google Gemini** API protocols into Google's internal web RPC stream.

---

## ⚡ 30-Second Installation

### macOS & Linux (1-Line Universal Setup)
```bash
curl -fsSL https://raw.githubusercontent.com/div197/bob-gemini-free/main/install.sh | bash
```
*(Auto-detects OS and architecture, compiles from source if Go is present, or downloads the standalone zero-dependency binary directly).*

### Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/div197/bob-gemini-free/main/install.ps1 | iex
```

### Docker & OrbStack (Instant Container)
```bash
docker run -d \
  --name bob-gemini-free \
  -p 8081:8081 \
  bob-gemini-free:local
```

---

## 🚀 Starting the Gateway

```bash
# Start on default port 8081
./bob-gemini-free

# Or specify custom port and host
./bob-gemini-free --port 8081 --host 127.0.0.1
```

Once started, the gateway listens at `http://127.0.0.1:8081`.

---

## 🎯 Model Tiers

| Tier | Required Setup | Available Models | Pricing |
| :--- | :--- | :--- | :--- |
| **Anonymous Free Tier** | Zero setup (no cookies) | `gemini-3.7-flash`, `gemini-3.6-flash`, `gemini-3.5-flash-thinking`, `gemini-flash-lite` | $0.00 / Free |
| **Authenticated Pro Tier** | 1-Click Login (`--login`) | `gemini-3.1-pro`, `gemini-3.1-pro-enhanced`, `imagen-3`, `gemini-nano-banana-2`, `gemini-nano-banana-pro` | $0.00 / Free ($20/mo subscription) |

---

## 🧪 Testing Your Installation

Run the built-in 12-point automated diagnostic suite:

```bash
./bob-gemini-free --test --test-url http://127.0.0.1:8081
```

Run a live stress benchmark:

```bash
./bob-gemini-free --bench --bench-concurrency 3 --bench-requests 6
```
