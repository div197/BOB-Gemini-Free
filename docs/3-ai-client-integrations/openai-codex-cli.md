# OpenAI Codex CLI (Responses API)

BOB Gemini Free natively implements OpenAI's next-generation **Responses API** (`POST /v1/responses`), designed for terminal coding agents, repository indexers, and autonomous developer loops.

---

## ⚡ 1-Minute Quickstart

Set the environment variables and launch Codex CLI:

### macOS & Linux (Terminal)
```bash
export OPENAI_BASE_URL="http://127.0.0.1:8081/v1"
export OPENAI_API_KEY="none"
codex
```

### Windows (PowerShell)
```powershell
$env:OPENAI_BASE_URL="http://127.0.0.1:8081/v1"
$env:OPENAI_API_KEY="none"
codex
```

---

## 📁 Project-Specific `.env` Setup

You can also drop a `.env` file into your project's root folder:

```ini
OPENAI_BASE_URL=http://127.0.0.1:8081/v1
OPENAI_API_KEY=none
```

Codex CLI will automatically detect and load these variables when launched in that directory.

> **Session Reset**: If Codex CLI was previously logged into an OpenAI account via browser OAuth, switch to local API mode by running:
> ```bash
> codex /logout
> ```
> Then relaunch Codex with the environment variables set.

---

## 🔄 Supported Responses API Features

- **`input` String / Multi-Part Array**: Automatically parsed into system, user, and tool turns.
- **`instructions` System Prompt**: Injected into top-level system context for precise role adherence.
- **`output_text` Streaming Lifecycle**:
  - `response.created`
  - `response.output_text.done`
  - `response.completed`
- **Tool Calling & Function Output**: Intercepts `function_call` output items and resolves `function_call_output` returns seamlessly.

---

## 💻 Permanent Shell Configuration

### For Zsh (`~/.zshrc`):
```bash
echo 'export OPENAI_BASE_URL="http://127.0.0.1:8081/v1"' >> ~/.zshrc
echo 'export OPENAI_API_KEY="none"' >> ~/.zshrc
source ~/.zshrc
```

### For Bash (`~/.bashrc`):
```bash
echo 'export OPENAI_BASE_URL="http://127.0.0.1:8081/v1"' >> ~/.bashrc
echo 'export OPENAI_API_KEY="none"' >> ~/.bashrc
source ~/.bashrc
```
