# Connecting Cursor, Windsurf, & Agentic IDEs

Connect modern AI coding editors and agentic IDE extensions directly to BOB Gemini Free using standard OpenAI protocol formatting.

---

## 🛠️ 1. Cursor IDE Setup (Agent & Composer Mode)

Cursor is the leading AI-first code editor. Connecting BOB enables full **Agent Mode** and multi-file reasoning for free:

1. Open **Cursor Settings** (`Cmd + ,` or `Ctrl + ,`) $\rightarrow$ **Models**.
2. Under **OpenAI API Key**, configure:
   * **OpenAI Base URL**: `http://127.0.0.1:9610/v1`
   * **API Key**: `none` (or your configured `api_keys`)
3. Click **Add Model** and add the following models:
   * `gemini-3.7-flash` (Ultra-fast code edits, Composer, and autocomplete)
   * `gemini-3.7-flash-thinking` (Deep multi-file reasoning and complex bug triage)
   * `gemini-3.1-pro` (Flagship Pro model for architectural refactors)
4. Enable the models in your Cursor dropdown.

---

## 🌊 2. Windsurf (Codeium Cascade) Setup

1. Open **Windsurf Settings** $\rightarrow$ **Cascade** $\rightarrow$ **Custom OpenAI Provider**.
2. Enter:
   * **Base URL**: `http://127.0.0.1:9610/v1`
   * **API Key**: `none`
   * **Model**: `gemini-3.7-flash`
3. Cascade will immediately start generating diffs and resolving repo contexts locally!

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
