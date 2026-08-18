# Connecting Cursor, Windsurf, Aider & Continue.dev

Connect modern AI coding IDEs, agents, and extensions directly to BOB Gemini Free using standard OpenAI protocol formatting.

---

## 🛠️ 1. Cursor IDE Setup

1. Open **Cursor Settings** (`Cmd + ,` or `Ctrl + ,`) $\rightarrow$ **Models**.
2. Under **OpenAI API Key**, configure:
   * **OpenAI Base URL**: `http://127.0.0.1:8081/v1`
   * **API Key**: `none` (or your configured `api_keys`)
3. Click **Add Model** and enter the following models:
   * `gemini-3.7-flash` (Fast code edits, autocomplete, and chat)
   * `gemini-3.7-flash-thinking` (Multi-file reasoning and complex bug triage)
   * `gemini-3.1-pro` (Flagship Pro model for architectural planning)
4. Toggle them **ON** and disable any default remote models you do not wish to use.

---

## 🌊 2. Windsurf (Codeium Cascade) Setup

1. Open **Windsurf Settings** $\rightarrow$ **Cascade** $\rightarrow$ **Custom OpenAI Provider**.
2. Enter:
   * **Base URL**: `http://127.0.0.1:8081/v1`
   * **API Key**: `none`
   * **Model**: `gemini-3.7-flash`
3. Cascade will immediately start generating diffs and resolving repo contexts locally!

---

## 💻 3. Aider CLI Setup

Aider is a popular terminal-based pair programming assistant:

```bash
# Run Aider with Gemini 3.7 Flash Thinking
aider \
  --openai-api-base http://127.0.0.1:8081/v1 \
  --openai-api-key none \
  --model openai/gemini-3.7-flash-thinking
```

---

## 🚀 4. Continue.dev Setup (`config.json`)

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
    },
    {
      "title": "Gemini 3.1 Pro",
      "provider": "openai",
      "model": "gemini-3.1-pro",
      "apiBase": "http://127.0.0.1:8081/v1",
      "apiKey": "none"
    }
  ],
  "tabAutocompleteModel": {
    "title": "Gemini Flash Lite",
    "provider": "openai",
    "model": "gemini-flash-lite",
    "apiBase": "http://127.0.0.1:8081/v1",
    "apiKey": "none"
  }
}
```
