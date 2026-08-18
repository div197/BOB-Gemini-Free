# Claude Code CLI (Anthropic Protocol)

BOB Gemini Free includes full native support for **Claude Code CLI** and Anthropic Messages SDK clients with real-time Server-Sent Events (SSE) streaming and reasoning blocks.

---

## ⚡ 1-Minute Claude Code Setup

Set the base URL and API key environment variables in your terminal:

```bash
export ANTHROPIC_BASE_URL="http://127.0.0.1:8081"
export ANTHROPIC_API_KEY="none"
```

Then run Claude Code:

```bash
claude
```

---

## 🧠 Native Thinking Support (`thinking`)

Claude Code automatically negotiates reasoning tokens via:
```json
{
  "thinking": {
    "type": "enabled",
    "budget_tokens": 2048
  }
}
```

BOB Gemini Free intercepts this parameter:
1. Translates `thinking: enabled` into deep Gemini reasoning depth.
2. Emits real-time `type: "thinking"` and `type: "text"` content blocks.
3. Automatically maps tools (`tool_use` and `tool_result`) back and forth seamlessly.

---

## 🛠️ Persisting Configuration in `.zshrc` / `.bashrc`

Add to your `~/.zshrc` or `~/.bashrc`:

```bash
# BOB Gemini Free - Claude Code Integration
alias claude-free='ANTHROPIC_BASE_URL=http://127.0.0.1:8081 ANTHROPIC_API_KEY=none claude'
```
