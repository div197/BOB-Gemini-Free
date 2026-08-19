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
  <img src="https://img.shields.io/badge/Release-v0.1.5-7c3aed?style=flat-square" alt="Release">
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

**BOB Gemini Free** is part of the **BOB Series** (*Break Ordinary Boundaries*) developed by [**ABCsteps.com**](https://abcsteps.com/) — an online AI engineering school founded by **Divyanshu Singh Chouhan** ([@div197](https://github.com/div197)) in Jodhpur, Rajasthan, India.

---

## 🌟 The Philosophy: Why "Break Ordinary Boundaries"?

In the modern AI landscape, learners, independent creators, and developers constantly hit **three ordinary boundaries**:

1. 💸 **The Economic Boundary (Cost Barrier)**: Frontier AI models often require expensive monthly subscriptions or high pay-per-token API credit cards. A student or solo builder experimenting with multi-agent coding can burn through hundreds of dollars in a few days.
2. 🔒 **The Ecosystem Boundary (Platform Lock-in)**: Closed ecosystems trap developers — Anthropic tools only talk to Anthropic, OpenAI tools only talk to OpenAI, and Google Gemini's web power remains trapped inside a browser tab.
3. ⚙️ **The Complexity Boundary (Friction & Setup Hell)**: Most proxy tools require compiling dependencies, installing runtimes (Python, Go, Node), or copy-pasting cryptic tokens with confusing error dialogs.

**BOB breaks all three boundaries at once**:
- ✨ **Zero Cost**: Unlocks Google's flagship **Gemini 3.7 Flash**, **Flash Thinking**, **3.1 Pro**, **Imagen 3**, and **Gemini Nano Banana** for every Google account.
- 🔓 **The "API-Less AI" Architecture**: No cloud console setup, no credit cards, no billing accounts, and zero risk of API key leaks. Your local session powers everything directly.
- 🌉 **Universal 4-in-1 Protocol**: One single local gateway translates Google's web stream simultaneously into **OpenAI Standard** (`/v1/chat/completions`, `/v1/responses`, `/v1/tokens/count`), **Anthropic Standard** (`/v1/messages` for Claude Code CLI), **Google Gemini Standard** (`/v1beta/models`, `:countTokens`), and **Embedded Go Library** (`pkg/gateway`).
- ⚡ **Zero-Friction Simplicity**: Runs as a single, self-contained native binary with **No Go, No Python, and No runtime required**. Includes a **1-Click Native Login Window (`--login`)** that sets up everything in seconds.

---

## 🚀 The "API-Less AI" Paradigm: True Freedom for Developers

Traditional developer tools force you through a maze of cloud billing setups, credit cards, and pay-per-token API fees. **BOB Gemini Free introduces the API-Less AI model**:

| Traditional Cloud API Model | The BOB API-Less Architecture |
| :--- | :--- |
| 💳 Requires credit card & billing account | **$0.00 / Zero credit cards needed** |
| 💸 Pay per million tokens & reasoning steps | **Unlimited daily coding on Flash & Thinking** |
| 🔑 Fragile API keys prone to leaks & theft | **Secure local session locked with `0600` permissions** |
| 🔒 Locked to single vendor CLI or protocol | **Universal 4-in-1 translation (OpenAI, Claude, Google, Go)** |
| 📊 Blind monthly invoices | **Live on-device token & dollar savings tracking (`GET /`)** |

---

<p align="center">
  <img src="assets/bob-gemini-free-universal-gateway.png" alt="BOB Gemini Free Universal AI Gateway Architecture & Ecosystem" width="100%">
</p>

---

## 💡 How It Works (In Plain English)

Think of **BOB Gemini Free** as a fast, private **Universal Translator** that runs quietly in the background on your laptop:

```
┌─────────────────────────────────────────────────────────────┐
│  Deep Agentic Coding Tools & Autonomous Frameworks          │
│  (Codex CLI, Claude Code CLI, Cursor, Grok Build, OpenHands)│
└──────────────────────────────┬──────────────────────────────┘
                               │ Speaks standard OpenAI / Anthropic API
                               ▼
┌─────────────────────────────────────────────────────────────┐
│  ⚡ BOB Gemini Free (Local Gateway on your machine)          │
│  Translates requests instantly with zero cloud tracking     │
└──────────────────────────────┬──────────────────────────────┘
                               │ Speaks Google Web RPC Stream
                               ▼
┌─────────────────────────────────────────────────────────────┐
│  🌐 Google Gemini Web (Flash 3.7 / Thinking / Pro / Images)  │
└─────────────────────────────┘
```

When **OpenAI Codex CLI**, **Claude Code CLI**, **Cursor**, or **Grok Build** asks for an API response, it sends a request to BOB. BOB translates that request into Gemini's format, fetches the response with full reasoning and vision support, and hands it back in milliseconds.

---

## The BOB Series (*Break Ordinary Boundaries*)

The **BOB Series** by **ABCsteps** is a developer-first collection of open-source runtimes and engines designed to remove paywalls and artificial constraints from modern software workflows:

* ⚡ [**BOB Gemini Free**](https://github.com/div197/bob-gemini-free) — Universal 4-in-1 local gateway unlocking Google Gemini Web for coding agents, IDEs, and developer tools.
* 🎥 [**BOB YouTube**](https://github.com/div197/BOB-Youtube) — Docker-first YouTube ingestion and transcription runtime for developers, products, and AI agents.

---

## Key Features

* **Free for Every Gmail User**: Out of the box, every Google account includes free Gemini access with high-speed Flash, adaptive Flash Lite, and deep Flash Thinking (up to 20,000+ characters of reasoning).
* **Universal 4-in-1 Protocol**: Drop-in compatible with OpenAI SDKs, Claude Code CLI, Anthropic SDKs, Codex CLI, Google GenAI SDKs, and in-process Go applications.
* **Native Token Counting Engine**: Accurate multi-script token counting via `POST /v1beta/models/{model}:countTokens` and `POST /v1/tokens/count`.
* **Live Telemetry & Cost Savings Tracking**: Real-time dollar savings reported in `GET /` (`estimated_savings_usd`).
* **13-Point Automated Diagnostic Suite**: Built-in verification command (`./bob-gemini-free --test`) and concurrency stress runner (`--bench`).
* **1-Click Native Login Window (`--login`)**: Standalone Google sign-in window automatically captures session tokens without Developer Tools or scary Keychain prompts.
* **Multi-Account Cookie Pool (`cookie_pool`)**: Distribute requests across multiple Google accounts with automatic 60-second backoff and transparent failover on rate limits.
* **Gemini Advanced ($20/mo) Integration**: Attach your session cookie to legitimately route to Google's flagship **Pro** model (`gemini-3.1-pro`) for deep mathematical and coding capabilities.
* **Imagen 3 & Gemini Nano Banana 2/Pro**: Standard OpenAI image generation endpoint (`/v1/images/generations`) with photorealistic and native visual rendering.
* **Claude 3.7 / 3.5 Native Thinking Support**: Accepts `thinking: { type: "enabled" }` and emits structured reasoning blocks and prompt caching counters for Claude Code CLI.
* **Full Multimodal Vision**: Send base64 images or image URLs via standard OpenAI payloads — automatically uploaded via Google's Scotty Resumable Upload protocol with automatic compression.
* **Zero Cost, Privacy First & Local Only**: Single static binary with near-zero memory footprint (<15MB RAM baseline), safe local-first binding (`127.0.0.1`), and zero external telemetry.

## Supported Tools & Ecosystem

BOB Gemini Free works out of the box with modern AI tools across deep coding, automation, and agentic workflows:

| Category | Supported Clients & Frameworks | Connection Endpoint |
| :--- | :--- | :--- |
| **Terminal Coding Agents** | OpenAI Codex CLI (`codex`), Claude Code CLI (`claude`), Aider CLI (`aider`), Gemini CLI (`gemini`) | Native Base URLs |
| **Agentic IDEs & Extensions** | Cursor (Agent Mode), Roo Code, Cline, Continue.dev, Windsurf | `http://127.0.0.1:9610/v1` |
| **Autonomous Agent Runtimes** | Grok Build, OpenHands, SWE-agent, Goose, LangChain, CrewAI, AutoGen, OpenAI Agents SDK | `http://127.0.0.1:9610/v1` |
| **Routers & Local Proxies** | LiteLLM, OneAPI, NewAPI, Portkey, OpenRouter | `http://127.0.0.1:9610/v1` |
| **Official SDKs** | OpenAI (Python/JS/Go/.NET/Java), Anthropic (Python/TypeScript), Google GenAI | Local Base URLs |

---

## Quick Start (Zero-Friction for All Users)

### Super Simple Start (No Config Required)

```bash
# 1. Start the gateway
./bob-gemini-free

# 2. Point any AI tool or script to:
# Base URL: http://127.0.0.1:9610/v1
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
docker run -d --name bob-gemini-free -p 9610:9610 bob-gemini-free
```

---

### Option 3: Automated Diagnostic Test Kit

Verify every endpoint, streaming chunk, reasoning model, and API format with the built-in diagnostic test kit:

```bash
# Run automated test kit against the default local server
./bob-gemini-free --test

# Or against a custom port / authenticated instance
./bob-gemini-free --test --test-url http://127.0.0.1:9610 --test-key your_api_key

# Or run the standalone script
./test-kit.sh
```

```text
[1/13]  [✔ PASS] Gateway Engine Health (GET /) (5ms)
[2/13]  [✔ PASS] OpenAI Models Registry (GET /v1/models) (0s)
[3/13]  [✔ PASS] Single Model Lookup (GET /v1/models/gemini-3.7-flash) (0s)
[4/13]  [✔ PASS] Gemini 3.7 Flash Fast Completion (3.0s)
[5/13]  [✔ PASS] Gemini 3.7 Flash Deep Reasoning (8.3s)
[6/13]  [✔ PASS] Real-time SSE Delta Stream & Usage (1.5s)
[7/13]  [✔ PASS] Developer Role & JSON Output Enforcement (3.9s)
[8/13]  [✔ PASS] Google Native Gemini API Format (3.5s)
[9/13]  [✔ PASS] OpenAI Codex CLI Responses API Format (3.5s)
[10/13] [✔ PASS] Anthropic Messages API Protocol (POST /v1/messages) (3.2s)
[11/13] [✔ PASS] OpenAI Function Calling & Tool Invocation (4.1s)
[12/13] [✔ PASS] Image Generation & Gemini Nano Banana Pipeline (3.8s)
[13/13] [✔ PASS] Token Counting Engine (Google :countTokens & OpenAI /v1/tokens/count) (1ms)
==================================================================
    ALL 13 DIAGNOSTIC CHECKS PASSED (100% SUCCESS)
==================================================================
```

### Live Status & Telemetry CLI (`--status`)

Query live metrics, token throughput, and estimated dollar savings from any running gateway directly in the terminal:

```bash
./bob-gemini-free --status
```

---

### Option 4: Native Background Service Daemonization (`service`)

Run BOB Gemini Free 24/7 in the background across system reboots with native OS daemons (macOS `launchd`, Linux `systemd`, Windows Startup):

```bash
# Install and enable auto-start on boot (Zero terminal windows needed)
./bob-gemini-free service install

# Check background daemon health and service definition
./bob-gemini-free service status

# Start / Stop / Uninstall daemon
./bob-gemini-free service start
./bob-gemini-free service stop
./bob-gemini-free service uninstall
```

---

### Option 5: In-Place Atomic Auto-Updater (`--update`)

Keep BOB Gemini Free updated with the latest releases directly from GitHub with 1 command:

```bash
./bob-gemini-free --update
```

---

### Option 6: Interactive Local-First Web Studio (`/playground` & `bob-gemini-free.abcsteps.com`)

Access the built-in, zero-dependency visual studio directly in your web browser:
* 🌐 **Online Cloudflare Pages Web Studio**: **[bob-gemini-free.abcsteps.com](https://bob-gemini-free.abcsteps.com/)** *(100% Unlimited Scalable Local-First PWA)*
* 🏠 **Local Server Address**: `http://127.0.0.1:9610/playground` (or `/ui`)

#### 🌟 100% Unlimited Scalability & State-of-the-Art Client Capabilities:
* 🔒 **100% On-Device Privacy**: When loaded from [bob-gemini-free.abcsteps.com](https://bob-gemini-free.abcsteps.com/), the static web studio auto-discovers and connects directly to your local BOB gateway engine at `http://127.0.0.1:9610` using Chrome Private Network Access (PNA). No prompts, chats, or thinking tokens ever touch intermediate cloud servers.
* 🐍 **Institutional-Grade In-Browser Pyodide WASM Python Sandbox**: Live client-side CPython 3.11 execution in an isolated WebAssembly sandbox with zero server-side execution risk, zero Python setup, interactive `input()` support, and automatic scientific package wheel streaming (`numpy`, `pandas`, `matplotlib`, `scipy`, `sympy`).
* 🗄️ **In-Browser SQLite WASM Database Studio (`🗄️ SQL WASM`)**: In-memory relational database powered by official SQLite WebAssembly (`sql.js`). Run queries in <1ms with `[ Run SQL ⚡ ]` chips, inspect schemas, and view interactive styled data tables with zero cloud costs.
* ⚡ **Native Interactive Artifacts Canvas Studio (Claude-Class Live Sandbox)**: Automatically detects and compiles standalone HTML5 applications, CSS3 animations, Canvas 2D/WebGL simulations, SVG vector graphics, and Mermaid diagrams with 1-click **`Launch ⚡`** chips, a sandboxed `iframe` studio modal (`[ ▶ Preview | ⟨/⟩ Code ]`), sandbox reload (`⟳`), source copy, and standalone window pop-out (`⛶`).
* 🪄 **Live AI Prompt Metaprompting Wand Engine (`🪄`)**: Background prompt optimization powered by `gemini-3.7-flash` that transforms rough thoughts into structured master specifications in ~200ms with seamless offline fallback (`⌘ + Shift + P`).
* 🔍 **Non-Breaking Reading Text Zoom Controller (`🔍 100%`)**: Targeted typography scaling (`calc(0.92rem * var(--reading-zoom))`) via sub-bar pill, `⌘+`/`⌘-`/`⌘0`, and Command Palette without breaking outer geometry, headers, or navigation.
* 🏛️ **Unified Sacred Geometry Studio Input Capsule**: Single cohesive dark-glass input capsule seamlessly housing auto-expanding textarea, vision attachments, power tools (`📎`, `अ Indic`, `🎙️ Voice`, `🪄 AI Wand`), and the golden `SEND ➤` CTA across mobile and desktop.
* 🎙️ **Natural HD Speech Studio & Floating Audio Controller Bar**: NotebookLM-class neural voice synthesis with Play/Pause (`⏸️`/`▶️`), speed cycling (`0.8x`–`1.5x`), 4-bar sound equalizer (`ılılı`), and sentence progress tracking.
* ✏️ **In-Place User Message Editing & Conversation Branching**: Inline editing with `✏️`, auto-growing editor, and conversation rewind (`Ctrl+Enter` / `Escape`).
* 🌐 **Bilingual & Multi-Indic Internationalization (`en` / `hi` + 8 Regional Scripts)**: 1-click dynamic language switcher and ⌘K shortcuts (`L1` English, `L2` हिन्दी), with support for 8 regional Indic scripts (हिन्दी, संस्कृतम्, मराठी, বাংলা, ગુજરાતી, தமிழ், తెలుగు, ਪੰਜਾਬੀ).
* ✍️ **Real-Time Indic Phonetic Transliteration with Backspace Undo**: Space-key dynamic conversion of Roman transliteration (`"namaste"`, `"aap kaise ho"`) into native Devanagari (`"नमस्ते"`, `"आप कैसे हो"`). Hitting `Backspace` instantly reverts converted words back to their original Roman characters.
* 🏛️ **Indian School Computer Lab LAN Master Architecture**: 1-process LAN host topology (`--host 0.0.0.0 --port 9610`) enabling 30-PC computer labs with 240+ daily students to access local AI at ₹0 cost on <25MB RAM.
* 👁️ **Multimodal Vision Engine**: Attach files (`📎`), drag-and-drop images onto the canvas, or paste screenshots directly from your clipboard (`⌘V` / `Ctrl+V`).
* 🧠 **Real-Time Reasoning Visualizer**: Stream step-by-step thinking tokens live inside isolated reasoning cards without distracting from the main response.
* 📐 **Synchronous Scientific Typography (KaTeX)**: Zero-flicker mathematical rendering for Dirac bra-kets ($\langle \psi | \phi \rangle$), Hilbert spaces, matrices ($\begin{pmatrix}1\\0\end{pmatrix}$), integrals, and proofs.
* ⚡ **Multi-Language Terminal Highlighting (Prism.js)**: Syntax highlighting across 200+ programming languages with one-click **`📋 Copy`** and **`💾 Save`** actions.
* 📊 **Interactive Architecture Diagrams (Mermaid.js)**: Automatically renders ````mermaid ```` blocks into live interactive SVG flowcharts and sequence diagrams.
* ⌨️ **Spotlight Command Palette (`⌘K` / `Ctrl+K`) & Keybindings**:
  * `1`–`5`: Switch between flagship models (`gemini-3.7-flash`, `thinking`, `gemini-3.1-pro`, `imagen-3`).
  * `T1`–`T5`: Switch Sacred Themes (Apple Design, BOB Builder, Vodafone Editorial, Spotify Dark, Gemini Quantum).
  * `G`: Open Gateway Engine Status & Endpoint Config • `N`: New Chat • `[` / `]`: Toggle sidebars • `E`: Export Markdown.
* 📊 **Live On-Device Telemetry & Savings**: Real-time ticker tracking uptime, requests served, token throughput, and estimated USD financial savings.

<p align="center">
  <img src="assets/bob-gemini-free-playground.png" alt="BOB Gemini Free Web Playground & Telemetry Dashboard — BOB Builder Theme Default" width="100%">
</p>

| Theme | Aesthetic Philosophy | Keybinding | Direct URL Preview |
| :--- | :--- | :---: | :--- |
| 🏗️ **BOB Builder** *(Default)* | Industrial high-contrast dark slate & energetic builder amber | `T1` | [View Snapshot](assets/theme-bob-builder.png) • `/playground?theme=bob-builder` |
| 🍏 **Apple Design** | SF Pro typography, frosted glass, parchment cards & Action Blue | `T5` | [View Snapshot](assets/theme-apple.png) • `/playground?theme=apple` |
| 📰 **Vodafone Editorial** | Clean light editorial paper, serif typography & crisp crimson | `T2` | [View Snapshot](assets/theme-vodafone.png) • `/playground?theme=vodafone` |
| 🎧 **Spotify Dark** | Pure AMOLED obsidian deep pitch & electric emerald green | `T3` | [View Snapshot](assets/theme-spotify.png) • `/playground?theme=spotify` |
| ⚛️ **Gemini Quantum** | Cyber deep indigo canvas & luminescent cyan neon glow | `T4` | [View Snapshot](assets/theme-quantum.png) • `/playground?theme=quantum` |

---

### Option 5: Build from Source with Make or Go (Go 1.22+)

```bash
# Build binary
make build

# Start the gateway
./bob-gemini-free --port 9610
```

The gateway will start listening at `http://127.0.0.1:9610/v1`.

---

## 📂 Multi-Language Examples & Client Integrations

Production-ready, copy-pasteable integration examples are located in the [`examples/`](examples/) directory:

* 🐍 **Python**:
  * [`examples/python/openai_chat.py`](examples/python/openai_chat.py) — OpenAI SDK with streaming reasoning extraction (`reasoning_content`).
  * [`examples/python/anthropic_messages.py`](examples/python/anthropic_messages.py) — Anthropic SDK Messages API with extended thinking.
* 🟨 **Node.js / TypeScript**:
  * [`examples/nodejs/openai_chat.mjs`](examples/nodejs/openai_chat.mjs) — OpenAI npm SDK stream consumer.
  * [`examples/nodejs/anthropic_messages.mjs`](examples/nodejs/anthropic_messages.mjs) — `@anthropic-ai/sdk` Messages API client.
* 🔷 **Go (Embedded Engine)**:
  * [`examples/go/embedded_sdk.go`](examples/go/embedded_sdk.go) — Direct in-process Go programmatic inference (`pkg/gateway.NewEngine()`).
* 🐚 **cURL & Shell**:
  * [`examples/curl/chat.sh`](examples/curl/chat.sh) — Standard completion.
  * [`examples/curl/stream_thinking.sh`](examples/curl/stream_thinking.sh) — Real-time reasoning stream.
  * [`examples/curl/anthropic.sh`](examples/curl/anthropic.sh) — Anthropic messages endpoint.
  * [`examples/curl/responses_codex.sh`](examples/curl/responses_codex.sh) — OpenAI Codex CLI `/v1/responses`.

---

## Client Integration

### OpenAI Python SDK

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://127.0.0.1:9610/v1",
    api_key="none"  # Or your configured api_key
)

# Text Chat Completion with Deep Thinking
response = client.chat.completions.create(
    model="gemini-3.7-flash-thinking",
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

client = OpenAI(base_url="http://127.0.0.1:9610/v1", api_key="none")

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
export ANTHROPIC_BASE_URL=http://127.0.0.1:9610
export ANTHROPIC_API_KEY=none
claude
```

#### Anthropic Python SDK

```python
from anthropic import Anthropic

client = Anthropic(
    base_url="http://127.0.0.1:9610",
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
export OPENAI_BASE_URL=http://127.0.0.1:9610/v1
export OPENAI_API_KEY=none
codex
```

### cURL (OpenAI Chat Completions)

```bash
curl http://127.0.0.1:9610/v1/chat/completions \
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
    base_url="http://127.0.0.1:9610/v1",
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
export GOOGLE_GEMINI_BASE_URL=http://127.0.0.1:9610
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

	http.ListenAndServe("127.0.0.1:9610", handler)
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
./scripts/bench.sh http://127.0.0.1:9610 3 6 your_api_key
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

## Unlocking Pro: Gemini Advanced ($20/mo) Cookies & Session Setup

Anonymous and standard free accounts have immediate access to Flash 3.7, Flash 3.6, Flash Thinking, and Flash Lite out of the box with zero cookies or setup.

If you have an active **Google AI / Gemini Advanced ($20/mo)** subscription (or 18 months free via Reliance Jio / university partnership offers) or want to unlock **Imagen 3 Image Generation**, configure your session cookie to activate authentic **Pro** model routing (`gemini-3.1-pro` / `gemini-pro`).

---

### Step 1: Extract Your Cookie in 15 Seconds (3 Visual Paths)

Open **Google Chrome**, **Arc**, **Edge**, or **Brave** and visit [**gemini.google.com**](https://gemini.google.com). Make sure you are signed in.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Chrome DevTools (F12 or Cmd+Opt+I)                                         │
├─────────────────────────────────────────────────────────────────────────────┤
│  [Network] Tab  Filter: [ app                   ] [X] Preserve log          │
│  ─────────────────────────────────────────────────────────────────────────  │
│  Name                                  Status   Type      Size              │
│  📄 app?eom=1&awwd=1&em=2&...          200      document  22.2 kB  <── CLICK│
│  ⚙️ batchexecute?rpcids=...            200      xhr       14.5 kB           │
│  ─────────────────────────────────────────────────────────────────────────  │
│  [Headers] [Payload] [Preview] [Response]                                   │
│  ▼ Request Headers                                                          │
│    :authority: gemini.google.com                                            │
│    Cookie: __Secure-BUCKET=...; SID=...; SAPISID=...;  <── RIGHT CLICK & COPY│
└─────────────────────────────────────────────────────────────────────────────┘
```

#### Path A: The Instant 1-Click Method (No Chat Needed)
1. Press **`F12`** (or **`Cmd + Option + I`** on macOS) to open Developer Tools.
2. Click the **Network** tab at the top.
3. Refresh the page (**`Cmd + R`** or **`F5`**).
4. Click on the document request named **`app?eom=1...`** (or any top **`batchexecute`** request).
5. In the right panel, ensure **Headers** is selected and scroll to **Request Headers**.
6. Right-click on the **`cookie:`** (or **`Cookie:`**) value and click **Copy value**.

#### Path B: The 1-Word Chat Method (`StreamGenerate`)
1. In DevTools **Network** tab, type **`StreamGenerate`** in the filter search box.
2. In Gemini, type any 1-word prompt (e.g. *"hi"*) and press Enter.
3. The request **`StreamGenerate`** will instantly appear in the Network list.
4. Click **`StreamGenerate`** $\rightarrow$ **Headers** $\rightarrow$ **Request Headers** $\rightarrow$ right-click **`cookie:`** $\rightarrow$ **Copy value**.

#### Path C: Application Storage Tab
1. In DevTools, click the **Application** tab (or click `>>` to find Application).
2. Expand **Storage** $\rightarrow$ **Cookies** $\rightarrow$ select `https://gemini.google.com`.
3. Highlight the cookie rows and copy.

---

### Step 2: Configure BOB Gemini Free (Zero Copy-Paste or Manual Methods)

#### 🌟 Method 0: 1-Click Interactive Sign-In Window (Zero Copy-Paste — Easiest for Everyone!)

Run the login command in your terminal:

```bash
./bob-gemini-free --login
```

```
┌─────────────────────────────────────────────────────────────┐
│  🌐 Sign in to Google Gemini (BOB Login Window)             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│                    Google                                   │
│                    Sign in with Google                      │
│                    [ your-email@gmail.com ]                 │
│                    [ Enter Password / Passkey ]             │
│                                                             │
│                    [ Next ]                                 │
│                                                             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼ (User signs in once)
┌─────────────────────────────────────────────────────────────┐
│  [✔] Verified 19 session tokens!                            │
│  [✔] Saved to ./cookie.txt and ~/.config/bob-gemini-free/   │
│  [✔] Gemini Pro model (gemini-3.1-pro) & Imagen 3 unlocked! │
└─────────────────────────────────────────────────────────────┘
```

1. A standalone Google Gemini sign-in window will open on your screen.
2. Sign in to your Google Account (supports Passkeys, 2FA, SMS).
3. As soon as login completes, BOB Gemini Free **automatically captures all 19+ session tokens via the Chrome DevTools Protocol**, saves `cookie.txt` (`mode 0600`), and closes the window.
4. **Zero DevTools, zero copy-pasting, zero Keychain password prompts!**

---

#### 📸 The Whole Truth: Multimodal Image Analysis (Vision) & Session Requirements

Google's internal web architecture strictly distinguishes between standard text reasoning and multimodal media uploads:

| Capability | Anonymous / Guest Session | Authenticated Google Session (`./cookie.txt` via `--login`) |
| :--- | :--- | :--- |
| **Text Chat & Coding** (`gemini-3.7-flash`, `gemini-3.6-flash`) | ✅ Unlocked | ✅ Full 1M+ Context & Peak Speed |
| **Deep Step-by-Step Reasoning** (`gemini-3.7-flash-thinking`) | ✅ Unlocked | ✅ Full Deep Thinking Blocks |
| **Multimodal Image Analysis (Vision)** (Diagrams, Screenshots, OCR) | ❌ **Blocked by Google** (`BardErrorInfo [1003]`) | ✅ **Fully Unlocked** (Zero Paywall) |
| **Imagen 3 Image Synthesis** (`imagen-3`) | ❌ **Blocked by Google** | ✅ **Fully Unlocked** (Photorealistic Synthesis) |
| **Pro Model Routing** (`gemini-3.1-pro` / `gemini-pro`) | ❌ Reverts to Flash | ✅ **Fully Unlocked** |

**Why does Google require a session for images?**
When you attach an image or screenshot, BOB Gemini Free compresses the payload and initiates a resumable upload to Google's Scotty storage (`content-push.googleapis.com/upload/`). Google's backend verifies that the storage tenant is cryptographically signed by an authenticated Google account (`SAPISIDHASH` + `__Secure-1PSIDTS`). If unauthenticated or expired, Google returns `BardErrorInfo [1003]`. 

Running `./bob-gemini-free --login` once authenticates your session and permanently unlocks vision analysis.

---

#### Method A: Interactive Automated Setup Helper (Paste Prompt)

```bash
./bob-gemini-free --setup-cookie
```

Paste your copied cookie string and press **Enter**. The helper automatically:
* Extracts and cryptographically verifies all 19+ session tokens (`SID`, `HSID`, `SSID`, `APISID`, `SAPISID`, `__Secure-1PSID`, `__Secure-3PSID`, `SIDCC`).
* Activates dynamic SHA-1 **`SAPISIDHASH`** generation (`SAPISIDHASH <timestamp>_<sha1>`).
* Securely writes to `~/.config/bob-gemini-free/cookie.txt` with locked POSIX `0600` permissions.
* Instantly activates Pro routing (`gemini-3.1-pro`) and Imagen 3 synthesis.

#### Method B: Zero-Config Local `cookie.txt`

Paste your cookie directly into `./cookie.txt` in the gateway folder:

```bash
cat << 'EOF' > cookie.txt
PASTE_YOUR_COOKIE_STRING_HERE
EOF
chmod 600 cookie.txt
./bob-gemini-free
```
*(BOB Gemini Free automatically discovers `./cookie.txt` on startup without requiring any CLI flags).*

#### Method C: Direct Command-Line Flag or Environment Variable

```bash
# Non-interactive CLI flag
./bob-gemini-free --cookie-string "SID=...; HSID=...; SAPISID=...; __Secure-1PSID=..."

# Or via environment variable
export BOB_GEMINI_FREE_COOKIE_FILE="/path/to/cookie.txt"
./bob-gemini-free
```

---

### Step 3: Multi-Account Profile Handling (`auth_user`)

If you are signed into multiple Google accounts in the same browser (e.g. Account 0 = Personal, Account 1 = Work):

1. Check your Gemini browser URL:
   * `https://gemini.google.com/app` $\rightarrow$ Account Index is `0` (default).
   * `https://gemini.google.com/u/1/app` $\rightarrow$ Account Index is `1`.
2. Specify your account index in `config.json` or pass via CLI:
   ```json
   {
     "auth_user": "1",
     "cookie_file": "./cookie.txt"
   }
   ```
   This ensures BOB Gemini Free dispatches `X-Goog-AuthUser: 1` and prefixes `/u/1/` to route to the correct Google profile.

---

### Step 4: Multi-Account Cookie Pool & Auto-Rotation (`cookie_pool`)

For high-concurrency teams, AI labs, or automated pipelines requiring massive throughput:

1. Create a `cookies/` folder and place multiple cookie files:
   ```
   ./cookies/
   ├── account_primary.txt
   ├── account_secondary.txt
   └── account_team.txt
   ```
2. Or configure `"cookie_pool"` in `config.json`:
   ```json
   {
     "cookie_pool": [
       "./cookies/account_primary.txt",
       "./cookies/account_secondary.txt"
     ]
   }
   ```
3. **Automated Round-Robin & Failover**: BOB Gemini Free automatically load-balances requests across all active accounts. If one account encounters a temporary burst rate-limit (HTTP 429), BOB **automatically backs off that account for 60 seconds and transparently retries on the next healthy account** without interrupting the client connection!

---

### Step 5: Running with Docker & OrbStack

To run your authenticated container in Docker or **OrbStack**:

```bash
# 1. Build local container image
docker build -t bob-gemini-free:local .

# 2. Run container with cookie mounted
docker run -d \
  --name bob-gemini-free \
  -p 9610:9610 \
  -v $(pwd)/cookie.txt:/app/cookie.txt:ro \
  -e BOB_GEMINI_FREE_COOKIE_FILE=/app/cookie.txt \
  bob-gemini-free:local
```

Verify container status:
```bash
docker logs bob-gemini-free
# Output:
#   Cookie: yes (/app/cookie.txt)
#   Listening: http://0.0.0.0:9610
```

---

## Configuration (`config.json`)

Create `config.json` or place it in `~/.config/bob-gemini-free/config.json`:

```json
{
  "port": 9610,
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
| `make run` | Build and start server immediately on port 9610 |
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
<summary><strong>4. How do I unlock Google's flagship Pro models (`gemini-3.1-pro` / `gemini-pro`) and Imagen 3?</strong></summary>

Out of the box, Free tier accounts access Flash 3.7, Flash 3.6, Flash Thinking, and Flash Lite. If you have an active Gemini Advanced ($20/mo) subscription (or 18 months free via Reliance Jio / college partnership offers):
1. Run `./bob-gemini-free --login` (Recommended 1-click native interactive sign-in).
2. Or run `./bob-gemini-free --setup-cookie` and paste your session cookie string.
3. The helper automatically extracts required tokens (`SID`, `HSID`, `SSID`, `APISID`, `SAPISID`, `__Secure-1PSID`, `__Secure-1PSIDTS`), computes dynamic `SAPISIDHASH` per request, and unlocks `gemini-3.1-pro` / `gemini-pro` and `imagen-3`.
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
<summary><strong>8. How do I use Vision and multimodal image inputs? Why is a session cookie required for images?</strong></summary>

Send standard OpenAI image payloads containing base64 data URLs (`data:image/png;base64,...`) or image files. BOB Gemini Free optimizes oversized images (downscaling to max 1024px, 75% JPEG quality, <1MB) and uploads them via Google's Scotty Resumable Upload protocol (`content-push.googleapis.com`). 

* **Session Requirement**: Google strictly binds Scotty file uploads to authenticated Google account sessions (`SAPISIDHASH` + `__Secure-1PSIDTS`). Unauthenticated requests will fail with `BardErrorInfo [1003]`.
* **Resolution**: Run `./bob-gemini-free --login` once to authenticate your session and permanently unlock vision.
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
export ANTHROPIC_BASE_URL=http://127.0.0.1:9610
export ANTHROPIC_API_KEY=none
claude
```
</details>

<details>
<summary><strong>13. How do I use BOB Gemini Free with OpenAI Codex CLI (`openai/codex`) and AI Router Proxies (LiteLLM / OpenRouter / Portkey)?</strong></summary>

* **OpenAI Codex CLI**: Supported via native `/v1/responses` and `/v1/chat/completions`. Set `OPENAI_BASE_URL=http://127.0.0.1:9610/v1` and `OPENAI_API_KEY=none`.
* **LiteLLM / OpenRouter / Portkey / OneAPI**: Configure `http://127.0.0.1:9610/v1` as your custom OpenAI upstream provider. The gateway returns standard SSE delta chunks, `reasoning_content` thinking blocks, and token usage accounting.
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

BOB Gemini Free stands on the collective wisdom and engineering breakthroughs of the global AI and open-source communities:

1. **Google Research & DeepMind**: For publishing the foundational Transformer architecture (*"Attention Is All You Need"*, Vaswani et al., 2017) and for engineering the state-of-the-art Gemini 3.7 Flash, Flash Thinking, 3.1 Pro, and Imagen 3 models with generous public web accessibility.
2. **OpenAI & Anthropic**: For establishing the open API standards, Messages schemas, reasoning block conventions, and coding agent CLI patterns that unite modern developer workflows.
3. **The Go Language Team & Chromium Engineers**: For the systems-level foundations (Go standard library concurrency, zero-dependency static compilation, and Chrome DevTools Protocol) enabling high-performance, local-first, zero-friction execution.
4. **The Global Open-Source Community**: The creators and maintainers of Cursor, Windsurf, Aider, Continue.dev, OpenWebUI, Cherry Studio, ChatBox, and the global indie hacker ecosystem pushing the frontiers of software engineering.
5. **ABCsteps Technologies (Jodhpur, Rajasthan)**: For championing truthful, first-principles AI engineering education, open learning foundations, and the **Break Ordinary Boundaries (BOB)** developer empowerment mission.

## ⚖️ Legal Disclaimer & Trademark Notice

**BOB Gemini Free** is an independent, open-source local gateway engine developed for research, interoperability, and developer educational purposes by **ABCsteps** ([abcsteps.com](https://abcsteps.com)) and **Divyanshu Singh Chouhan** ([@div197](https://github.com/div197)).

* **No Affiliation**: BOB Gemini Free is not affiliated with, endorsed by, sponsored by, or in any way officially connected with Google LLC, Alphabet Inc., OpenAI Inc., Anthropic PBC, or any of their subsidiaries.
* **Trademarks**: "Google", "Gemini", "OpenAI", "ChatGPT", "Anthropic", "Claude", and related marks are trademarks of their respective owners. Their use in this codebase and documentation is strictly nominative for compatibility and protocol interoperability description.
* **Compliance**: Users are solely responsible for complying with the applicable terms of service and acceptable use policies of any upstream web services or accounts they connect.

---

## License

MIT License. Developed with pride by [ABCsteps.com](https://abcsteps.com/) & [Divyanshu Singh Chouhan](https://github.com/div197).
