# Claude Code CLI (Anthropic Protocol)

BOB Gemini Free implements Anthropic's **Messages API** (`POST /v1/messages`) natively with complete Server-Sent Events (SSE) streaming and reasoning blocks, making it 100% compatible with **Claude Code CLI**.

---

## ⚡ 1-Minute Claude Code Quickstart

Set the environment variables and run `claude`:

```bash
export ANTHROPIC_BASE_URL="http://127.0.0.1:8081"
export ANTHROPIC_API_KEY="none"
claude
```

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
│  BOB Gateway: Injects schema & extracts tool call           │
└──────────────────────────────┬──────────────────────────────┘
                               │ Emits tool_use block
                               ▼
┌─────────────────────────────────────────────────────────────┐
│  Claude Code CLI executes tool & sends tool_result          │
└─────────────────────────────────────────────────────────────┘
```

---

## 💻 Shell Aliases & Permanent Configuration

### For Zsh (`~/.zshrc`) or Bash (`~/.bashrc`):
```bash
alias claude-free='ANTHROPIC_BASE_URL=http://127.0.0.1:8081 ANTHROPIC_API_KEY=none claude'
```

### For Fish Shell (`~/.config/fish/config.fish`):
```fish
alias claude-free='env ANTHROPIC_BASE_URL=http://127.0.0.1:8081 ANTHROPIC_API_KEY=none claude'
```
