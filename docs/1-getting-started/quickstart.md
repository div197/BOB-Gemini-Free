# Getting Started: Quickstart Guide

Welcome to **BOB Gemini Free** (*Break Ordinary Boundaries*) by **ABCsteps** ([abcsteps.com](https://abcsteps.com/)) and **Divyanshu Singh Chouhan** ([@div197](https://github.com/div197)).

BOB Gemini Free is a high-performance, single-binary local AI gateway written in Go that translates standard **OpenAI**, **Anthropic Messages**, and **Google Gemini** API protocols into Google's internal web RPC stream.

---

## ⚡ 30-Second Installation

### macOS & Linux (1-Line Universal Setup)
```bash
curl -fsSL https://raw.githubusercontent.com/div197/bob-gemini-free/main/install.sh | bash
```
*(Auto-detects OS and architecture, compiles from source if Go is present, or downloads a standalone binary that needs no separately managed Go/Python/Node runtime.)*

### Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/div197/bob-gemini-free/main/install.ps1 | iex
```

### Docker & OrbStack (Instant Container)
```bash
docker run -d \
  --name bob-gemini-free \
  -p 9610:9610 \
  bob-gemini-free:local
```

---

## 🚀 Starting the Gateway

```bash
# Start on default port 9610
./bob-gemini-free

# Or specify custom port and host
./bob-gemini-free --port 9610 --host 127.0.0.1
```

Once started, the gateway listens at `http://127.0.0.1:9610`.

### Classroom / Computer Lab Mode

For a classroom where many students connect at the same time, run the gateway on one teacher machine or local lab server:

```bash
./bob-gemini-free --host 0.0.0.0 --port 9610 --cookie-pool-dir ./cookies
```

Then students open:

```text
http://TEACHER_LAN_IP:9610/playground
```

Use the [Classroom LAN Deployment Guide](./classroom-lan-guide.md) before a live class. It explains why the Cloudflare Pages demo can hit datacenter egress rate limits during concentrated student bursts, and how to verify the local LAN path before class.

---

## 🎯 Model Tiers

| Tier | Required Setup | Available Models | Pricing |
| :--- | :--- | :--- | :--- |
| **Unauthenticated route** | No cookie configured | Flash/thinking aliases may be available if the current Google web session permits them | Upstream-dependent |
| **Authenticated route** | 1-Click Login (`--login`) | Pro/image aliases may be available if the current Google account and web protocol permit them | Upstream/account-dependent |

---

## 💻 3-Protocol Verification in 10 Seconds

### 1. OpenAI Standard (cURL)
```bash
curl http://127.0.0.1:9610/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-3.7-flash",
    "messages": [{"role": "user", "content": "Hello from OpenAI API!"}]
  }'
```

### 2. Anthropic Messages (Claude Code CLI format)
```bash
curl http://127.0.0.1:9610/v1/messages \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude-3-5-sonnet",
    "messages": [{"role": "user", "content": "Hello from Claude!"}],
    "max_tokens": 100
  }'
```

### 3. Google-shaped v1beta adapter
```bash
curl http://127.0.0.1:9610/v1beta/models/gemini-3.7-flash:generateContent \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [{"parts": [{"text": "Hello from Gemini API!"}]}]
  }'
```

---

## 🐍 Python & TypeScript SDK Examples

### Python (`openai` SDK)
```python
from openai import OpenAI

client = OpenAI(
    base_url="http://127.0.0.1:9610/v1",
    api_key="none"
)

response = client.chat.completions.create(
    model="gemini-3.7-flash",
    messages=[{"role": "user", "content": "Explain quantum superposition in 2 sentences."}]
)

print(response.choices[0].message.content)
```

### TypeScript / Node.js (`openai` SDK)
```typescript
import OpenAI from "openai";

const openai = new OpenAI({
  baseURL: "http://127.0.0.1:9610/v1",
  apiKey: "none",
});

async function main() {
  const completion = await openai.chat.completions.create({
    model: "gemini-3.7-flash-thinking",
    messages: [{ role: "user", content: "Solve: 17 * 19" }],
  });

  console.log(completion.choices[0].message.content);
}

main();
```

---

## ⚙️ Background Service Management (24/7 Autostart on Reboot)

Run BOB Gemini Free quietly in the background without needing open terminal windows:

```bash
# Register native OS daemon (macOS launchd, Linux systemd, Windows Startup)
./bob-gemini-free service install

# Check background daemon health and service definition
./bob-gemini-free service status

# Start / Stop / Uninstall daemon
./bob-gemini-free service start
./bob-gemini-free service stop
./bob-gemini-free service uninstall
```

---

## 🔄 Signed CLI updater

Keep the standalone CLI up to date with an explicit, signed GitHub release
check:

```bash
./bob-gemini-free --update
```

The CLI refuses an unsigned update. The public Preview 2 native beta is a separate
manual beta path; its native auto-update is not enabled until a desktop build
embeds the release public key and the release publishes a signed manifest.

---

## 🧪 Testing & Diagnostics

Run the built-in 13-point automated diagnostic suite:

```bash
./bob-gemini-free --test --test-url http://127.0.0.1:9610
```

Run a live stress benchmark:

```bash
./bob-gemini-free --bench --bench-concurrency 3 --bench-requests 6
```
