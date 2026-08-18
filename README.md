<p align="center">
  <img src="assets/bob-gemini-free-banner.jpg" alt="BOB Gemini Free Banner" width="100%">
</p>

<h1 align="center">BOB Gemini Free</h1>

<p align="center">
  <strong>Universal 3-in-1 AI Gateway Engine</strong><br>
  <em>Drop-in OpenAI, Anthropic, and Google Gemini API for Developers & Agents</em>
</p>

<p align="center">
  <a href="https://abcsteps.com/"><img src="https://img.shields.io/badge/Powered%20by-ABCsteps.com-2563eb?style=flat-square" alt="ABCsteps"></a>
  <a href="https://github.com/div197/bob-gemini-free"><img src="https://img.shields.io/badge/Author-Divyanshu%20Singh%20Chouhan-16a34a?style=flat-square" alt="Author"></a>
  <img src="https://img.shields.io/badge/Release-v0.1.0-7c3aed?style=flat-square" alt="Release">
  <img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/Protocols-OpenAI%20%7C%20Anthropic%20%7C%20Gemini-059669?style=flat-square" alt="Protocols">
  <img src="https://img.shields.io/badge/License-MIT-f59e0b?style=flat-square" alt="License">
</p>

<p align="center">
  <a href="README.md"><strong>English Documentation</strong></a> &nbsp;•&nbsp;
  <a href="README.hi.md"><strong>हिंदी गाइड (Hindi)</strong></a> &nbsp;•&nbsp;
  <a href="CHANGELOG.md"><strong>Changelog</strong></a>
</p>

---

**BOB Gemini Free** is part of the **BOB Series** (*Break Ordinary Boundaries*) developed by [**ABCsteps.com**](https://abcsteps.com/) — an online AI engineering school founded by **Divyanshu Singh Chouhan** ([@div197](https://github.com/div197)).

It is a high-performance, single-binary Go gateway that converts Google Gemini's web interface into a **Universal 3-in-1 AI Gateway Engine** supporting:
* **OpenAI Standard**: `/v1/chat/completions`, `/v1/models`, `/v1/models/{model}`, and `/v1/responses` (OpenAI Codex CLI & Agents).
* **Anthropic Standard**: `/v1/messages` (Claude Code CLI & Anthropic SDKs).
* **Google Gemini Standard**: `/v1beta/models` and `/v1beta/models/{target}` (Gemini CLI & Google GenAI SDKs).

---

## The BOB Series (*Break Ordinary Boundaries*)

The **BOB Series** by **ABCsteps** is a developer-first suite of high-impact runtimes, proxies, and automation engines designed to remove paywalls and artificial constraints from modern AI workflows:

* 🎥 [**BOB YouTube**](https://github.com/div197/BOB-Youtube) — Docker-first YouTube ingestion runtime for developers, products, bulk workflows, and AI agents.
* ⚡ [**BOB Gemini Free**](https://github.com/div197/bob-gemini-free) — High-performance OpenAI, Anthropic, and Gemini gateway unlocking Google Gemini Web for agents, IDEs, and developer tools.

---

## Key Features

* **Free for Every Gmail User**: Out of the box, every Google account includes free Gemini access with high-speed Flash, adaptive Flash Lite, and deep Flash Thinking (up to 20,000+ characters of reasoning).
* **Universal 3-in-1 Protocol**: Drop-in compatible with OpenAI SDKs, Claude Code CLI, Anthropic SDKs, Codex CLI, and Google GenAI SDKs.
* **Gemini Advanced ($20/mo) Integration**: Attach your session cookie to legitimately route to Google's flagship **Pro** model for deep mathematical and coding capabilities.
* **OpenAI Drop-In Replacement**: Seamlessly works with Cherry Studio, ChatBox, Codex CLI, Cursor, Claude Code, OpenAI/Anthropic Python/TypeScript SDKs, and custom AI agents.
* **Full Multimodal Vision**: Send base64 images or image URLs via standard OpenAI payloads — automatically uploaded via Google's Scotty Resumable Upload protocol with automatic compression.
* **Reasoning Control**: Tune thinking depth dynamically via `@think=N` (0 = deepest reasoning, 4 = fast concise output).
* **Zero Cost & Local First**: Single static binary with near-zero memory footprint (<15MB RAM baseline), safe local-first binding (`127.0.0.1`), and high concurrency throughput.
* **TLS Browser Impersonation**: Built-in support to mimic Chrome/Firefox/Safari TLS fingerprints via `tls-client` for network environments facing WAF restrictions.

## Supported Tools & Ecosystem

BOB Gemini Free works out of the box with modern AI tools across coding, automation, and conversational workflows:

| Category | Supported Clients & Frameworks | Connection Endpoint |
| :--- | :--- | :--- |
| **Code Editors & IDEs** | Cursor, Windsurf, VS Code (Continue, Cline, Roo Code, Aider) | `http://127.0.0.1:8081/v1` |
| **CLI Coding Engines** | Claude Code CLI (`claude`), OpenAI Codex CLI (`codex`), Gemini CLI (`gemini`) | Native Base URLs |
| **GUI Chat Apps** | Cherry Studio, ChatBox, OpenWebUI, NextChat, LibreChat | `http://127.0.0.1:8081/v1` |
| **Agent Frameworks** | LangChain, LlamaIndex, CrewAI, AutoGen, OpenAI Agents SDK | `http://127.0.0.1:8081/v1` |
| **Routers & Proxies** | LiteLLM, OneAPI, NewAPI, Portkey, OpenRouter | `http://127.0.0.1:8081/v1` |
| **Official SDKs** | OpenAI (Python/JS/Go/.NET/Java), Anthropic (Python/TypeScript), Google GenAI | Local Base URLs |

---

## Quick Start (Zero-Friction for All Users)

### Super Simple Start (No Config Required)

```bash
# 1. Start the gateway
./bob-gemini-free

# 2. Point any AI tool or script to:
# Base URL: http://127.0.0.1:8081/v1
# API Key:  none
```

---

### Option 1: Automatic Installer (No Go Required)

#### macOS & Linux
```bash
chmod +x install.sh
./install.sh
```

#### Windows (PowerShell)
```powershell
.\install.ps1
```

---

### Option 2: Docker & Docker Compose

```bash
# Using Docker Compose
docker compose up -d

# Or standard Docker
docker build -t bob-gemini-free .
docker run -d --name bob-gemini-free -p 8081:8081 bob-gemini-free
```

---

### Option 3: Automated Diagnostic Test Kit

Verify every endpoint, streaming chunk, reasoning model, and API format with the built-in diagnostic test kit:

```bash
# Run automated test kit against the default local server
./bob-gemini-free --test

# Or against a custom port / authenticated instance
./bob-gemini-free --test --test-url http://127.0.0.1:8081 --test-key your_api_key

# Or run the standalone script
./test-kit.sh
```

```text
[1/9] [✔ PASS] Gateway Engine Health (GET /) (5ms)
[2/9] [✔ PASS] OpenAI Models Registry (GET /v1/models) (0s)
[3/9] [✔ PASS] Single Model Lookup (GET /v1/models/gemini-3.7-flash) (0s)
[4/9] [✔ PASS] Gemini 3.7 Flash Fast Completion (4.0s)
[5/9] [✔ PASS] Gemini 3.7 Flash Deep Reasoning (8.3s)
[6/9] [✔ PASS] Real-time SSE Delta Stream & Usage (1.5s)
[7/9] [✔ PASS] Developer Role & JSON Output Enforcement (3.9s)
[8/9] [✔ PASS] Google Native Gemini API Format (3.5s)
[9/9] [✔ PASS] OpenAI Codex CLI Responses API Format (3.5s)
==================================================================
    ALL 9 DIAGNOSTIC CHECKS PASSED (100% SUCCESS)
==================================================================
```

---

### Option 4: Build from Source with Make or Go (Go 1.22+)

```bash
# Build binary
make build

# Start the gateway
./bob-gemini-free --port 8081
```

The gateway will start listening at `http://127.0.0.1:8081/v1`.

---

## Client Integration

### OpenAI Python SDK

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://127.0.0.1:8081/v1",
    api_key="none"  # Or your configured api_key
)

# Text Chat Completion with Deep Thinking
response = client.chat.completions.create(
    model="gemini-3.5-flash-thinking",
    messages=[
        {"role": "user", "content": "Explain quantum error correction step by step."}
    ]
)
print(response.choices[0].message.content)
```

### Multimodal Vision (OpenAI Format)

```python
import base64
from openai import OpenAI

client = OpenAI(base_url="http://127.0.0.1:8081/v1", api_key="none")

with open("diagram.png", "rb") as img_file:
    b64_img = base64.b64encode(img_file.read()).decode("utf-8")

response = client.chat.completions.create(
    model="gemini-3.6-flash",
    messages=[
        {
            "role": "user",
            "content": [
                {"type": "text", "text": "What architecture is depicted in this image?"},
                {"type": "image_url", "image_url": {"url": f"data:image/png;base64,{b64_img}"}}
            ]
        }
    ]
)
### Claude Code CLI & Anthropic SDK (Native Direct Support)

BOB Gemini Free includes native support for Anthropic's Messages API protocol (`POST /v1/messages`).

#### Claude Code CLI Setup

```bash
export ANTHROPIC_BASE_URL=http://127.0.0.1:8081
export ANTHROPIC_API_KEY=none
claude
```

#### Anthropic Python SDK

```python
from anthropic import Anthropic

client = Anthropic(
    base_url="http://127.0.0.1:8081",
    api_key="none"
)

message = client.messages.create(
    model="gemini-3.7-flash-thinking",
    max_tokens=4096,
    messages=[
        {"role": "user", "content": "Write a concurrent Go worker pool."}
    ]
)
print(message.content[0].text)
```

### OpenAI Codex CLI

```bash
export OPENAI_BASE_URL=http://127.0.0.1:8081/v1
export OPENAI_API_KEY=none
codex
```

### cURL (OpenAI Chat Completions)

```bash
curl http://127.0.0.1:8081/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-3.7-flash",
    "messages": [{"role": "user", "content": "Hello BOB Gemini Free!"}]
  }'
```

### OpenAI Image Generation (`POST /v1/images/generations`)

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://127.0.0.1:8081/v1",
    api_key="none"
)

response = client.images.generate(
    model="dall-e-3",
    prompt="A futuristic neon cybernetic logo for BOB Gemini Free, digital art, 8k",
    size="1024x1024",
    n=1
)
print("Generated Image URL:", response.data[0].url)
```

> [!NOTE]
> **Imagen Tool Authentication**: Google requires an active signed-in Google account for Imagen image generation. When running without cookies in anonymous mode, Gemini returns an informative policy response. To unlock Imagen generation, attach your session cookie via `--cookie-file cookie.txt` or `--setup-cookie`.

### Gemini CLI (Native Google API)

```bash
export GEMINI_API_KEY=none
export GOOGLE_GEMINI_BASE_URL=http://127.0.0.1:8081
gemini
```

### Embedded Go Library / Package Import

Embed BOB Gemini Free directly as an in-process module inside any Go application or agent runtime:

```go
package main

import (
	"net/http"
	"github.com/div197/bob-gemini-free/pkg/gateway"
)

func main() {
	handler := gateway.NewHandler(
		gateway.WithDefaultModel("gemini-3.7-flash"),
		gateway.WithCookieFile("cookie.txt"), // optional
	)

	http.ListenAndServe("127.0.0.1:8081", handler)
}
```

---

## Deep Architectural Comparison: Google AI Studio vs. BOB Gemini Free

### Official Limits & Feature Matrix (Free Tier Without Paid Billing)

| Dimension / Metric | Google AI Studio (Free Tier) | **BOB Gemini Free (Local Gateway Engine)** |
| :--- | :--- | :--- |
| **Flash Daily Limit (RPD)** | **1,500 Requests / Day** (Hard daily shutoff) | **Practically Unlimited Interactive Queries** |
| **Flash Rate Limits (RPM)** | **15 Requests / Minute** (`429 RESOURCE_EXHAUSTED`) | **High-Throughput Web Session (30+ RPM burst)** |
| **Flash Token Rate (TPM)** | 1,000,000 Tokens / Minute | Stream-buffered interactive throughput |
| **Pro Daily Limit (RPD)** | **50 Requests / Day** (Punishing hard cap) | **Essentially Unlimited Pro Queries** *(with Advanced cookie)* |
| **Pro Rate Limits (RPM)** | **2 Requests / Minute** (Severe throttling) | **Standard interactive web concurrency** |
| **Thinking Reasoning Depth** | Restricted / suppressed on free keys | **20,000+ characters of deep step-by-step reasoning** |
| **OpenAI Protocol Support** | ❌ None (Requires custom SDK / glue code) | **✅ 100% Native Drop-In (`/v1/chat/completions`, `/v1/responses`)** |
| **Developer Role Support** | ❌ None | **✅ Full OpenAI `developer` & `system` role parsing** |
| **Reasoning Tokens Export** | ❌ Proprietary format | **✅ Standard `reasoning_content` (renders cards in Cursor/OpenWebUI)** |
| **Data Training & Logging** | ⚠️ **Logged and reviewed by Google reviewers for training** | **🛡️ 100% Local Gateway (Bound to `127.0.0.1`, zero telemetry)** |
| **Setup Friction** | Cloud console, project creation, API key rotation | **Zero config: `./bob-gemini-free` and start coding** |
| **Total Financial Cost** | $0 (until throttled) or paid per-token bill | **100% Free forever** |

---

### Ultimate Maximum Limits & Operational Thresholds

To achieve maximum stability and zero upstream rate limiting, follow these empirical thresholds:

* **Maximum Output Token Length**: Google Flash models return up to **8,192 tokens (~32,000 characters)** per response. Thinking traces can exceed **20,000 characters** before emitting the final answer.
* **Optimal Concurrency Bandwidth**:
  * **Flash 3.7 / 3.6 / Flash Lite**: **3 to 5 simultaneous streams** per IP/instance.
  * **Deep Thinking & Pro 3.1**: **2 to 3 simultaneous streams** (to prevent payload chunk congestion).
* **Automatic Retry Strategy**: Built-in exponential backoff (`retry_attempts: 3`, `retry_delay_sec: 2`) handles momentary network hiccups without breaking client requests.
* **Maximum Image Dimension**: Automatic downscaling engine resizes large screenshots and photos to **1024px JPEG quality 75 (<1MB)** before uploading to Google Scotty storage.

---

## Live Performance & Stress Benchmark

BOB Gemini Free includes a built-in stress tester and throughput benchmark:

```bash
# Run benchmark with 3 concurrent workers and 6 batch queries
./bob-gemini-free --bench --bench-concurrency 3 --bench-requests 6

# Or benchmark against a custom URL / port
./scripts/bench.sh http://127.0.0.1:8081 3 6 your_api_key
```

```text
------------------------------------------------------------------
                    BENCHMARK RESULTS & METRICS                   
------------------------------------------------------------------
  • Completed Requests:   6 / 6 (100.0% Success)
  • Total Elapsed Time:   4.12s
  • Average Latency:      1.85s
  • Median Latency (P50): 1.72s
  • 90th Percentile (P90):2.10s
  • Request Throughput:   1.46 req/sec
  • Token Throughput:     48.5 tokens/sec
==================================================================
```

---

## Background Daemon & Auto-Start Services

Keep BOB Gemini Free running silently in the background across all major operating systems:

### Linux (Systemd Service)

```bash
# 1. Copy binary and service file
sudo cp bob-gemini-free /usr/local/bin/
sudo cp scripts/bob-gemini-free.service /etc/systemd/system/

# 2. Enable and start on boot
sudo systemctl daemon-reload
sudo systemctl enable --now bob-gemini-free
```

### macOS (Launchd Daemon)

```bash
# 1. Copy binary to path
sudo cp bob-gemini-free /usr/local/bin/

# 2. Install plist to user LaunchAgents
cp scripts/com.abcsteps.bob-gemini-free.plist ~/Library/LaunchAgents/
launchctl load ~/Library/LaunchAgents/com.abcsteps.bob-gemini-free.plist
```

### Windows (Background Auto-Start)

```cmd
REM Run the background batch runner
scripts\start-service.bat
```

---

## Model Matrix & Reasoning Controls

| Local Model Alias | Backend Mode | Default Think Depth | Output Profile | Auth Requirement |
| :--- | :---: | :---: | :--- | :--- |
| `gemini-3.7-flash` | Mode 1 | `@think=4` | **Latest flagship fast model** (~12k chars) | Free / Anonymous |
| `gemini-3.7-flash-thinking` | Mode 2 | `@think=0` | **Latest flagship deep thinking mode** (~20k+ chars) | Free / Anonymous |
| `gemini-3.6-flash` / `gemini-flash` | Mode 1 | `@think=4` | High-speed all-around model | Free / Anonymous |
| `gemini-3.5-flash-thinking` / `gemini-thinking` | Mode 2 | `@think=0` | **Deep thinking mode** (~20k+ chars reasoning) | Free / Anonymous |
| `gemini-3.5-flash-thinking-lite` | Mode 5 | `@think=0` | Adaptive thinking depth (~15k chars) | Free / Anonymous |
| `gemini-flash-lite` / `gemini-lite` | Mode 6 | `@think=4` | Ultra-low latency responses | Free / Anonymous |
| `gemini-auto` | Mode 4 | `@think=4` | Google server-side auto routing | Free / Anonymous |
| `gemini-3.1-pro` / `gemini-pro` | Mode 3 | `@think=4` | Flagship Pro reasoning & code | **Gemini Advanced Cookie** |
| `gemini-3.1-pro-enhanced` | Mode 3 | `@think=4` | Pro with enhanced output (experimental) | **Gemini Advanced Cookie** |

### Dynamic Thinking Depth Override

Append `@think=N` to any model alias in your client to control reasoning depth on the fly:

```text
gemini-3.6-flash@think=0    # Deepest step-by-step reasoning tokens
gemini-3.6-flash@think=2    # Balanced medium reasoning
gemini-3.6-flash@think=4    # Direct fast response (shallowest reasoning)
```

---

## Architecture & Data Flow

<p align="center">
  <img src="assets/bob-gemini-free-architecture.jpg" alt="BOB Gemini Free Architecture" width="100%" />
</p>

BOB Gemini Free bridges modern developer clients directly to Google Gemini's web infrastructure:
1. **Client Tier**: Receives standard OpenAI/Gemini REST calls from Cursor, Cherry Studio, ChatBox, OpenWebUI, or Python/TS SDKs.
2. **Gateway Engine**: Translates OpenAI message arrays into Google BoQ RPC payloads, compresses oversized multimodal vision uploads via Google Scotty, extracts thinking traces into `reasoning_content`, and deduplicates real-time SSE stream frames.
3. **Google Web Tier**: Dispatches requests directly over TLS with browser fingerprint impersonation and dynamic `SAPISIDHASH` authentication.

---

## Unlocking Pro: Gemini Advanced ($20/mo) Cookies

Anonymous and standard free accounts have immediate access to Flash, Thinking, and Lite models out of the box with zero cookies.

If you have an active **Google AI / Gemini Advanced ($20/mo)** subscription or want to unlock **Imagen 3 Image Generation**, configure your session cookie to activate authentic **Pro** model routing:

### Step 1: Extract Your Cookie in 15 Seconds

1. Open **Google Chrome**, **Edge**, or **Brave** and visit [**gemini.google.com**](https://gemini.google.com). Make sure you are signed in.
2. Press **`F12`** (or **`Cmd + Option + I`** on macOS) to open Developer Tools.
3. Click on the **Network** tab at the top.
4. Send any short message in Gemini (e.g. *"hello"*).
5. In the Network filter search box, type **`batchexecute`** (or click the first request in the list).
6. Under the **Headers** tab, scroll down to **Request Headers**.
7. Locate the **`Cookie:`** line, right-click on the value, and select **Copy Value**.

---

### Step 2: Configure BOB Gemini Free (3 Simple Methods)

#### Method A: Interactive Automated Setup Helper (Easiest)

Run the built-in setup command and paste the copied cookie string when prompted:

```bash
./bob-gemini-free --setup-cookie
```

The setup helper automatically:
* Extracts and verifies all essential session tokens (`SID`, `HSID`, `SSID`, `APISID`, `SAPISID`, `__Secure-1PSID`).
* Validates `SAPISID` for dynamic `SAPISIDHASH` generation.
* Saves the cookie file to `~/.config/bob-gemini-free/cookie.txt` with secure POSIX `0600` permissions.
* Activates Pro routing (`gemini-3.1-pro` / `gemini-pro`) and Imagen 3 generation.

#### Method B: Zero-Config Local `cookie.txt`

Simply paste your cookie string into a file named `cookie.txt` in the same directory as the executable:

```bash
echo "YOUR_PASTED_COOKIE_STRING" > cookie.txt
chmod 600 cookie.txt
./bob-gemini-free
```
*(BOB Gemini Free will automatically detect and load `./cookie.txt` on startup without any flags).*

#### Method C: Direct Command-Line Flag

```bash
./bob-gemini-free --cookie-string "SID=...; HSID=...; SAPISID=...; __Secure-1PSID=..."
```

---

## Configuration (`config.json`)

Create `config.json` or place it in `~/.config/bob-gemini-free/config.json`:

```json
{
  "port": 8081,
  "host": "127.0.0.1",
  "retry_attempts": 3,
  "retry_delay_sec": 2,
  "request_timeout_sec": 180,
  "default_model": "gemini-3.6-flash",
  "api_keys": ["sk-your-secret-key"],
  "cookie_file": null,
  "proxy": null,
  "impersonate": "chrome_133",
  "log_requests": true
}
```

* **`host`**: Defaults to `127.0.0.1` for maximum security.
* **`api_keys`**: When empty `[]`, authentication is disabled. When configured, requests require `Authorization: Bearer <key>`.
* **`impersonate`**: Mimic browser TLS signatures (`chrome_120`, `chrome_131`, `chrome_133`, `chrome_146`, `firefox_147`, `safari_16_0`).
* **`proxy`**: Route requests through an HTTP/HTTPS/SOCKS5 proxy (e.g. `http://127.0.0.1:7890`).

---

## Developer Commands (Makefile)

| Command | Description |
| :--- | :--- |
| `make build` | Compile static binary for host architecture into `./bob-gemini-free` |
| `make run` | Build and start server immediately on port 8081 |
| `make test` | Run complete unit test suite with verbose logging |
| `make test-cover` | Run test suite with line-by-line coverage report |
| `make dist` | Cross-compile standalone binaries for macOS (ARM64/Intel), Linux (AMD64/ARM64), and Windows |
| `make clean` | Remove local binaries and distribution directory |

---

## Frequently Asked Questions (FAQ)

<details>
<summary><strong>1. How does BOB Gemini Free provide free access without an API key?</strong></summary>

Google provides access to Gemini (Flash 3.7, Flash 3.6, Flash Lite, and Flash Thinking) to every user with a standard Google/Gmail account via the web interface. **BOB Gemini Free** acts as a high-performance local proxy that translates standard OpenAI (`/v1/chat/completions`) and Gemini (`/v1beta/models`) API calls directly into Google's internal web RPC format.
</details>

<details>
<summary><strong>2. Is my Google Account safe? What about rate limits or bans?</strong></summary>

* **In Anonymous / Free Mode**: BOB Gemini Free attaches **zero session cookies and zero account identifiers**. All requests are completely unlinked from your Google identity.
* **In Gemini Advanced ($20/mo) Mode**: BOB Gemini Free uses authentic browser TLS fingerprints (`tls-client` impersonating Chrome 133) and standard web RPC payloads. It behaves identically to a user typing in a web browser tab.
* **Operational Best Practices**: Keep concurrency between 3 to 5 simultaneous requests. Do not blast 100 queries/second from a single IP to avoid temporary upstream rate limiting.
</details>

<details>
<summary><strong>3. How does this compare to Google AI Studio Free Tier or Paid OpenAI / Anthropic APIs?</strong></summary>

* **Google AI Studio Free Tier**: Enforces a strict 15 RPM (Requests Per Minute) and a hard daily token quota. Once you hit the quota, your app stops working until midnight UTC.
* **BOB Gemini Free**: Operates on Google's interactive web backend with **essentially unlimited daily interactive queries**, $0 token cost, and up to **20,000+ characters of deep reasoning** (`gemini-3.7-flash-thinking` / `gemini-thinking`).
</details>

<details>
<summary><strong>4. How do I unlock Google's flagship Pro models (`gemini-3.1-pro` / `gemini-pro`)?</strong></summary>

Out of the box, Free tier accounts access Flash 3.7, Flash 3.6, Flash Thinking, and Flash Lite. If you have an active Gemini Advanced ($20/mo) subscription (or 18 months free via Reliance Jio / college partnership offers):
1. Run `./bob-gemini-free --setup-cookie`
2. Paste your session cookie string.
3. The helper automatically extracts required tokens (`SID`, `HSID`, `SSID`, `APISID`, `SAPISID`, `__Secure-1PSID`), computes dynamic `SAPISIDHASH` per request, and unlocks `gemini-3.1-pro` / `gemini-pro`.
</details>

<details>
<summary><strong>5. How does thinking / reasoning mode work? Where do I see thinking tokens?</strong></summary>

When querying thinking models (`gemini-3.7-flash-thinking` or any model with `@think=0`), BOB Gemini Free isolates internal reasoning traces (` ```thought ... ``` `) and populates the standard OpenAI `reasoning_content` field. In frontends like Cursor, Cherry Studio, ChatBox, or OpenWebUI, this automatically renders a collapsible "Reasoning Process" card alongside the clean final response.
</details>

<details>
<summary><strong>6. Can I run this on headless Linux VPS, Raspberry Pi, or Docker? What if datacenter IPs are challenged?</strong></summary>

Yes! BOB Gemini Free is a single lightweight static binary (<15MB RAM) with official multi-arch Docker support (`alpine:3.21`).
* To bind publicly on a VPS, set `BOB_GEMINI_FREE_HOST=0.0.0.0` and configure `BOB_GEMINI_FREE_API_KEYS`.
* If a cloud datacenter IP (AWS, Hetzner, DigitalOcean) encounters Google WAF challenges, route traffic through a residential/SOCKS5 proxy via `--proxy socks5://...` or enable TLS browser impersonation (`--impersonate chrome_133`).
</details>

<details>
<summary><strong>7. Does this support Tool / Function Calling and Structured JSON Outputs?</strong></summary>

Yes. BOB Gemini Free automatically injects tool schemas into system instructions and parses markdown ` ```tool_call ` outputs back into standard OpenAI `tool_calls` JSON objects. Passing `response_format: {"type": "json_object"}` strictly enforces structured JSON generation.
</details>

<details>
<summary><strong>8. How do I use Vision and multimodal image inputs?</strong></summary>

Send standard OpenAI image payloads containing base64 data URLs (`data:image/png;base64,...`) or base64 strings. BOB Gemini Free automatically optimizes oversized images (downscaling to max 1024px, 75% JPEG quality, <1MB) and uploads them via Google's Scotty Resumable Upload protocol to obtain authentic WIZ file references.
</details>

<details>
<summary><strong>9. Can multiple apps or teammates share one instance?</strong></summary>

Yes. Set `api_keys: ["sk-team-key-1", "sk-team-key-2"]` in `config.json` or pass `BOB_GEMINI_FREE_API_KEYS`. All incoming requests are authenticated using constant-time comparison (`crypto/subtle`) to prevent timing attacks.
</details>

<details>
<summary><strong>10. Is there any telemetry, tracking, or external logging?</strong></summary>

Zero. BOB Gemini Free is 100% open source under the MIT License, written in pure Go, with zero analytics, zero external logging, and network calls made strictly between your machine and Google's official endpoints over TLS.
</details>

<details>
<summary><strong>11. Why Go instead of Python or Node.js?</strong></summary>

* **Single Static Binary**: No Python virtual environments, pip dependencies, or heavy `node_modules` folders.
* **Instant Startup**: Cold boots in <5ms with <15MB baseline RAM usage.
* **High Concurrency**: Handles multiple concurrent SSE streaming delta connections effortlessly with zero garbage collection pauses.
</details>

<details>
<summary><strong>12. Can I use Claude Code CLI directly with BOB Gemini Free without LiteLLM or an external router?</strong></summary>

**Yes, 100% natively.** BOB Gemini Free implements the complete **Anthropic Messages API protocol (`POST /v1/messages`)** with standard SSE lifecycle events (`message_start`, `content_block_start`, `content_block_delta`, `message_delta`, `message_stop`) and bash/tool execution.

Simply export the official environment variables and launch Claude Code:
```bash
export ANTHROPIC_BASE_URL=http://127.0.0.1:8081
export ANTHROPIC_API_KEY=none
claude
```
</details>

<details>
<summary><strong>13. How do I use BOB Gemini Free with OpenAI Codex CLI (`openai/codex`) and AI Router Proxies (LiteLLM / OpenRouter / Portkey)?</strong></summary>

* **OpenAI Codex CLI**: Supported via native `/v1/responses` and `/v1/chat/completions`. Set `OPENAI_BASE_URL=http://127.0.0.1:8081/v1` and `OPENAI_API_KEY=none`.
* **LiteLLM / OpenRouter / Portkey / OneAPI**: Configure `http://127.0.0.1:8081/v1` as your custom OpenAI upstream provider. The gateway returns standard SSE delta chunks, `reasoning_content` thinking blocks, and token usage accounting.
</details>

---

## About ABCsteps

[**ABCsteps**](https://abcsteps.com/) is an online AI engineering school founded by **Divyanshu Singh Chouhan** in Jodhpur, Rajasthan, India (operated by ABC Steps Technologies Pvt Ltd).

Its mission is to provide an open, practical engineering education for builders, students, and software developers:
* 📚 [**20-Lesson AI Engineering Curriculum**](https://abcsteps.com/offerings/) — A connected, project-based foundation spanning AI copilots, Docker, APIs, SQLite/PostgreSQL, prompt engineering, and full-stack AI architectures.
* ✍️ [**Practical Engineering Blog**](https://abcsteps.com/blog/) — Deep-dive tutorials on LLM internals, coding agents, containerization, and systems design.
* 🧭 [**Curated Reading Paths & Glossary**](https://abcsteps.com/blog/paths/) — Structured reference guides and technical definitions for self-directed learners.
* 🎓 [**Founder-Led Programs**](https://abcsteps.com/enroll/compare/) — Guided tracks, live cohorts, 1:1 mentorship, architecture reviews, and institutional workshops.

Explore the complete learning platform at [**https://abcsteps.com/**](https://abcsteps.com/).

## Acknowledgements & Research Foundations

- **Google Research**: For publishing the foundational Transformer architecture (*"Attention Is All You Need"*, Vaswani et al., 2017) and for providing generous public web access to state-of-the-art Gemini intelligence.
- **Open-Source Community**: For the continuous advancements in open API tooling, agent frameworks, and cross-platform developer tools.

---

## License

MIT License. Developed with pride by [ABCsteps.com](https://abcsteps.com/).
