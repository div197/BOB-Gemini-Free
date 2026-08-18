# OpenWebUI, Cherry Studio, & ChatBox

Connect standalone desktop clients and web UIs to BOB Gemini Free for interactive chatting, file attachments, and image synthesis.

---

## 🍒 Cherry Studio / ChatBox Setup

1. Open **Settings** $\rightarrow$ **Model Provider** $\rightarrow$ **OpenAI**.
2. Configure:
   * **API Host / Base URL**: `http://127.0.0.1:8081/v1`
   * **API Key**: `none`
3. Click **Fetch Models** to populate all 56+ registered Gemini models automatically.
4. Select `gemini-3.7-flash` or `gemini-3.7-flash-thinking`.

---

## 🌐 OpenWebUI Setup

In your OpenWebUI environment variables or settings:

```bash
OPENAI_API_BASE_URL="http://127.0.0.1:8081/v1"
OPENAI_API_KEY="none"
```

OpenWebUI will automatically detect all registered models (`GET /v1/models`), support multimodal vision uploads, and render math equations seamlessly.
