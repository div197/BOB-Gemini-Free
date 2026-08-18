# Connecting Cursor, Windsurf, & Continue.dev

Connect modern AI coding IDEs directly to BOB Gemini Free using the standard OpenAI protocol.

---

## 🛠️ Cursor Configuration

1. Open **Cursor Settings** $\rightarrow$ **Models** $\rightarrow$ **OpenAI API Key**.
2. Set:
   * **OpenAI Base URL**: `http://127.0.0.1:8081/v1`
   * **API Key**: `none` (or your configured `api_keys`)
3. Add custom model names:
   * `gemini-3.7-flash` (Ultra-fast code generation)
   * `gemini-3.7-flash-thinking` (Deep reasoning & multi-step planning)
   * `gemini-3.1-pro` (Pro model for complex refactoring)
4. Enable the models in your Cursor dropdown.

---

## 🌊 Windsurf (Codeium) Configuration

1. Open **Windsurf Settings** $\rightarrow$ **Cascade** $\rightarrow$ **Custom OpenAI Provider**.
2. Configure:
   ```json
   {
     "base_url": "http://127.0.0.1:8081/v1",
     "api_key": "none",
     "model": "gemini-3.7-flash"
   }
   ```

---

## 🚀 Continue.dev (`config.json`)

In `~/.continue/config.json`:

```json
{
  "models": [
    {
      "title": "Gemini 3.7 Flash",
      "provider": "openai",
      "model": "gemini-3.7-flash",
      "apiBase": "http://127.0.0.1:8081/v1",
      "apiKey": "none"
    },
    {
      "title": "Gemini 3.7 Flash Thinking",
      "provider": "openai",
      "model": "gemini-3.7-flash-thinking",
      "apiBase": "http://127.0.0.1:8081/v1",
      "apiKey": "none"
    }
  ]
}
```
