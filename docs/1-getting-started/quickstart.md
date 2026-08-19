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

---

## 🎯 Model Tiers

| Tier | Required Setup | Available Models | Pricing |
| :--- | :--- | :--- | :--- |
| **Anonymous Free Tier** | Zero setup (no cookies) | `gemini-3.7-flash`, `gemini-3.6-flash`, `gemini-3.5-flash-thinking`, `gemini-flash-lite` | $0.00 / Free |
| **Authenticated Pro Tier** | 1-Click Login (`--login`) | `gemini-3.1-pro`, `gemini-3.1-pro-enhanced`, `imagen-3`, `gemini-nano-banana-2`, `gemini-nano-banana-pro` | $0.00 / Free ($20/mo subscription) |

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

### 3. Google Gemini Native v1beta
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

## 🧪 Testing & Diagnostics

Run the built-in 12-point automated diagnostic suite:

```bash
./bob-gemini-free --test --test-url http://127.0.0.1:9610
```

Run a live stress benchmark:

```bash
./bob-gemini-free --bench --bench-concurrency 3 --bench-requests 6
```
