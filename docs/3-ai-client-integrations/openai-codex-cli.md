# OpenAI Codex CLI (Responses API)

BOB Gemini Free natively implements OpenAI's next-generation **Responses API** (`POST /v1/responses`), designed for terminal coding agents and autonomous workflows.

---

## ⚡ Setup for OpenAI Codex CLI

Set the base URL and API key:

```bash
export OPENAI_BASE_URL="http://127.0.0.1:8081/v1"
export OPENAI_API_KEY="none"
```

Run Codex CLI:

```bash
codex
```

---

## 🔄 Supported Responses API Features

- **`input` String / Multi-Part Array**: Automatically parsed into system, user, and tool turns.
- **`instructions` System Prompt**: Injected into top-level system context.
- **`output_text` Streaming Lifecycle**:
  - `response.created`
  - `response.output_text.done`
  - `response.completed`
- **Tool Calling**: Intercepts `function_call` output items and resolves `function_call_output` returns.
