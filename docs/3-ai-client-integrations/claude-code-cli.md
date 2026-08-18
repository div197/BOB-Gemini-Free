# Claude Code CLI (Anthropic Protocol)

BOB Gemini Free implements Anthropic's **Messages API** (`POST /v1/messages`) natively with complete Server-Sent Events (SSE) streaming and reasoning blocks, making it 100% compatible with **Claude Code CLI**.

---

## ⚡ 1-Minute Claude Code Quickstart

Set the environment variables and run `claude`:

```bash
# macOS / Linux (Terminal)
export ANTHROPIC_BASE_URL="http://127.0.0.1:8081"
export ANTHROPIC_API_KEY="none"
claude
```

### Windows (PowerShell)
```powershell
$env:ANTHROPIC_BASE_URL="http://127.0.0.1:8081"
$env:ANTHROPIC_API_KEY="none"
claude
```

---

## 🔍 Verifying the Connection

Once inside Claude Code, verify that traffic is routing locally through BOB:

```bash
claude /status
```
*(Confirms that `ANTHROPIC_BASE_URL` is pointing to `http://127.0.0.1:8081`).*

> **Tip**: If you were previously logged in via OAuth subscription and want to force local routing, run:
> ```bash
> claude /logout
> ```
> Then relaunch with `ANTHROPIC_BASE_URL` and `ANTHROPIC_API_KEY` set.

---

## 🧠 Extended Thinking Support (`thinking`)

When Claude Code CLI asks for thinking tokens (e.g. `thinking: { type: "enabled", budget_tokens: 2048 }`), BOB Gemini Free:

1. Translates the thinking request into deep Gemini reasoning depth (`think=0`).
2. Streams real-time `type: "thinking"` content blocks for intermediate thoughts.
3. Concludes with `type: "text"` for the final response and tool execution.

---

## 🛠️ Tool Calling & Terminal Execution Lifecycle

Claude Code CLI uses tool calls to inspect files, edit code, and run bash commands. BOB Gemini Free translates this seamlessly:

```
┌─────────────────────────────────────────────────────────────┐
│  Claude Code CLI: "I will run `git status` via bash tool"   │
└──────────────────────────────┬──────────────────────────────┘
                               │ POST /v1/messages
                               ▼
┌─────────────────────────────────────────────────────────────┐
│  ⚡ BOB Gateway: Injects schema & extracts tool call         │
└──────────────────────────────┬──────────────────────────────┘
                               │ Emits tool_use block
                               ▼
┌─────────────────────────────────────────────────────────────┐
│  Claude Code CLI executes tool & sends tool_result          │
└─────────────────────────────────────────────────────────────┘
```

---

## 💻 Permanent Shell Configuration

### For Zsh (`~/.zshrc`):
```bash
echo 'export ANTHROPIC_BASE_URL="http://127.0.0.1:8081"' >> ~/.zshrc
echo 'export ANTHROPIC_API_KEY="none"' >> ~/.zshrc
source ~/.zshrc
```

### For Bash (`~/.bashrc`):
```bash
echo 'export ANTHROPIC_BASE_URL="http://127.0.0.1:8081"' >> ~/.bashrc
echo 'export ANTHROPIC_API_KEY="none"' >> ~/.bashrc
source ~/.bashrc
```

### Handy One-Word Alias:
```bash
alias claude-free='ANTHROPIC_BASE_URL=http://127.0.0.1:8081 ANTHROPIC_API_KEY=none claude'
```
