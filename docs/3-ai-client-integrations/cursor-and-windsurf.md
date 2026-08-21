# Connecting Cursor, Windsurf, & Agentic IDEs

Connect modern AI coding editors and agentic IDE extensions directly to BOB Gemini Free using standard OpenAI protocol formatting.

---

## 🛠️ 1. Cursor IDE Setup (Agent & Composer Mode)

Connecting BOB can enable selected Cursor Agent/Composer requests through the
OpenAI-shaped adapter. Actual client features, tool execution, quotas, and
model access remain endpoint- and session-dependent:

1. Open **Cursor Settings** (`Cmd + ,` or `Ctrl + ,`) $\rightarrow$ **Models**.
2. Under **OpenAI API Key**, configure:
   * **OpenAI Base URL**: `http://127.0.0.1:9610/v1`
   * **API Key**: `none` (or your configured `api_keys`)
3. Click **Add Model** and add the following models:
   * `gemini-3.7-flash` (fast-mode alias)
   * `gemini-3.7-flash-thinking` (thinking-mode alias)
   * `gemini-3.1-pro` (experimental/session-dependent Pro alias)
4. Enable the models in your Cursor dropdown.

---

## 🌊 2. Windsurf (Codeium Cascade) Setup

1. Open **Windsurf Settings** $\rightarrow$ **Cascade** $\rightarrow$ **Custom OpenAI Provider**.
2. Enter:
   * **Base URL**: `http://127.0.0.1:9610/v1`
   * **API Key**: `none`
   * **Model**: `gemini-3.7-flash`
3. Cascade may route supported requests through the local adapter; verify the
specific model and tool behavior on the target setup.

---

## 🚀 3. Continue.dev & VS Code Extensions

In `~/.continue/config.json`:

```json
{
  "models": [
    {
      "title": "Gemini 3.7 Flash",
      "provider": "openai",
      "model": "gemini-3.7-flash",
      "apiBase": "http://127.0.0.1:9610/v1",
      "apiKey": "none"
    },
    {
      "title": "Gemini 3.7 Flash Thinking",
      "provider": "openai",
      "model": "gemini-3.7-flash-thinking",
      "apiBase": "http://127.0.0.1:9610/v1",
      "apiKey": "none"
    },
    {
      "title": "Gemini 3.1 Pro",
      "provider": "openai",
      "model": "gemini-3.1-pro",
      "apiBase": "http://127.0.0.1:9610/v1",
      "apiKey": "none"
    }
  ],
  "tabAutocompleteModel": {
    "title": "Gemini Flash Lite",
    "provider": "openai",
    "model": "gemini-flash-lite",
    "apiBase": "http://127.0.0.1:9610/v1",
    "apiKey": "none"
  }
}
```
