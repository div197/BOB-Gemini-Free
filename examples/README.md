# BOB Gemini Free — Code Examples & SDK Integrations

This directory provides copy-pasteable, production-ready examples demonstrating how to connect AI coding agents, autonomous scripts, and backend microservices to **BOB Gemini Free**.

---

## 📂 Available Examples

### 1. Python
* **[`examples/python/openai_chat.py`](python/openai_chat.py)** — OpenAI Python SDK (`openai`) with streaming and live `reasoning_content` thinking extraction.
* **[`examples/python/anthropic_messages.py`](python/anthropic_messages.py)** — Anthropic Python SDK (`anthropic`) with streaming and extended thinking.

### 2. Node.js / TypeScript
* **[`examples/nodejs/openai_chat.mjs`](nodejs/openai_chat.mjs)** — OpenAI Node.js SDK (`openai`) with real-time stream decoding.
* **[`examples/nodejs/anthropic_messages.mjs`](nodejs/anthropic_messages.mjs)** — Anthropic Node.js SDK (`@anthropic-ai/sdk`) Messages API.

### 3. Go (Embedded Programmatic Engine)
* **[`examples/go/embedded_sdk.go`](go/embedded_sdk.go)** — Direct in-process Go programmatic inference (`pkg/gateway.NewEngine()`) without HTTP networking.

### 4. cURL & Shell Scripts
* **[`examples/curl/chat.sh`](curl/chat.sh)** — Standard non-streaming chat completion.
* **[`examples/curl/stream_thinking.sh`](curl/stream_thinking.sh)** — Real-time reasoning stream with SSE chunks.
* **[`examples/curl/anthropic.sh`](curl/anthropic.sh)** — Anthropic `/v1/messages` protocol.
* **[`examples/curl/responses_codex.sh`](curl/responses_codex.sh)** — OpenAI Codex CLI `/v1/responses` endpoint.

---

## 🚀 Quick Run

1. Start your local BOB Gemini Free gateway:
   ```bash
   ./bob-gemini-free --port 8081
   ```

2. Open the built-in Web Playground in your browser:
   ```
   http://127.0.0.1:8081/playground
   ```

3. Run any Python/Node/Shell example against the running gateway!
