package server

import (
	"net/http"
)

const playgroundHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>BOB Gemini Free — Universal AI Gateway</title>
<style>
:root {
  --bg-primary: #0a0e17;
  --bg-card: #121826;
  --bg-input: #1a2234;
  --border: #26334d;
  --accent-cyan: #00f0ff;
  --accent-blue: #3b82f6;
  --accent-purple: #a855f7;
  --accent-green: #10b981;
  --text-primary: #f1f5f9;
  --text-secondary: #94a3b8;
  --font-mono: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
* { box-sizing: border-box; margin: 0; padding: 0; }
body {
  background: var(--bg-primary);
  color: var(--text-primary);
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  line-height: 1.5;
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}
header {
  background: var(--bg-card);
  border-bottom: 1px solid var(--border);
  padding: 12px 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.brand { display: flex; align-items: center; gap: 12px; }
.logo {
  font-weight: 800;
  font-size: 1.2rem;
  letter-spacing: 0.5px;
  background: linear-gradient(135deg, var(--accent-cyan), var(--accent-purple));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}
.badge {
  font-size: 0.75rem;
  background: rgba(0, 240, 255, 0.1);
  color: var(--accent-cyan);
  border: 1px solid rgba(0, 240, 255, 0.3);
  padding: 2px 8px;
  border-radius: 9999px;
  font-weight: 600;
}
.telemetry { display: flex; gap: 16px; font-size: 0.8rem; color: var(--text-secondary); }
.telemetry span b { color: var(--accent-cyan); }
main {
  flex: 1;
  display: grid;
  grid-template-columns: 320px 1fr 340px;
  overflow: hidden;
}
.sidebar-left, .sidebar-right {
  background: var(--bg-card);
  padding: 16px;
  overflow-y: auto;
  border-right: 1px solid var(--border);
}
.sidebar-right { border-right: none; border-left: 1px solid var(--border); }
.section-title {
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 1px;
  color: var(--text-secondary);
  margin-bottom: 12px;
  font-weight: 700;
}
.control-group { margin-bottom: 16px; }
label { display: block; font-size: 0.8rem; color: var(--text-secondary); margin-bottom: 6px; }
select, input, textarea {
  width: 100%;
  background: var(--bg-input);
  border: 1px solid var(--border);
  color: var(--text-primary);
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 0.9rem;
  outline: none;
}
select:focus, input:focus, textarea:focus { border-color: var(--accent-cyan); }
.chat-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--bg-primary);
}
.messages {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.msg {
  max-width: 80%;
  padding: 12px 16px;
  border-radius: 8px;
  font-size: 0.95rem;
  line-height: 1.6;
}
.msg-user {
  background: #1e293b;
  align-self: flex-end;
  border: 1px solid #334155;
}
.msg-assistant {
  background: var(--bg-card);
  align-self: flex-start;
  border: 1px solid var(--border);
  width: 100%;
  max-width: 90%;
}
.thought-box {
  background: #0f172a;
  border-left: 3px solid var(--accent-purple);
  padding: 10px 12px;
  margin-bottom: 12px;
  border-radius: 4px;
  font-size: 0.85rem;
  color: #cbd5e1;
  font-family: var(--font-mono);
  white-space: pre-wrap;
  display: none;
}
.thought-box.active { display: block; }
.thought-header {
  color: var(--accent-purple);
  font-weight: 700;
  font-size: 0.75rem;
  text-transform: uppercase;
  margin-bottom: 4px;
  display: flex;
  align-items: center;
  gap: 6px;
}
.input-area {
  padding: 16px 20px;
  background: var(--bg-card);
  border-top: 1px solid var(--border);
  display: flex;
  gap: 12px;
}
.btn {
  background: linear-gradient(135deg, var(--accent-cyan), var(--accent-blue));
  color: #000;
  font-weight: 700;
  border: none;
  padding: 10px 20px;
  border-radius: 6px;
  cursor: pointer;
  transition: opacity 0.2s;
}
.btn:hover { opacity: 0.9; }
.btn:disabled { opacity: 0.5; cursor: not-allowed; }
.snippet-box {
  background: var(--bg-input);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 10px;
  font-family: var(--font-mono);
  font-size: 0.75rem;
  color: #a5f3fc;
  overflow-x: auto;
  white-space: pre;
  margin-bottom: 16px;
}
.copy-btn {
  background: transparent;
  color: var(--text-secondary);
  border: 1px solid var(--border);
  padding: 4px 8px;
  font-size: 0.7rem;
  border-radius: 4px;
  cursor: pointer;
  margin-top: 6px;
}
.copy-btn:hover { color: var(--text-primary); border-color: var(--text-secondary); }
</style>
</head>
<body>
<header>
  <div class="brand">
    <div class="logo">⚡ BOB GEMINI FREE</div>
    <div class="badge">Universal AI Gateway</div>
  </div>
  <div class="telemetry">
    <span>Uptime: <b id="stat-uptime">0s</b></span>
    <span>Requests: <b id="stat-requests">0</b></span>
    <span>Tokens: <b id="stat-tokens">0</b></span>
    <span>Estimated Savings: <b id="stat-savings">$0.00</b></span>
  </div>
</header>
<main>
  <div class="sidebar-left">
    <div class="section-title">Model & Reasoning</div>
    <div class="control-group">
      <label>Target Model</label>
      <select id="model-select">
        <option value="gemini-3.7-flash" selected>gemini-3.7-flash (Flagship Fast)</option>
        <option value="gemini-3.7-flash-thinking">gemini-3.7-flash-thinking (Deep Reasoning)</option>
        <option value="gemini-3.6-flash">gemini-3.6-flash</option>
        <option value="gemini-3.1-pro">gemini-3.1-pro (Pro Routing)</option>
        <option value="gemini-flash-lite">gemini-flash-lite (Ultra Fast)</option>
        <option value="imagen-3">imagen-3 (Image Synthesis)</option>
      </select>
    </div>
    <div class="control-group">
      <label>Reasoning Effort (@think)</label>
      <select id="think-select">
        <option value="default" selected>Model Default</option>
        <option value="@think=0">Deep Reasoning (@think=0)</option>
        <option value="@think=2">Medium Thinking (@think=2)</option>
        <option value="@think=4">Zero Thinking / Fast (@think=4)</option>
      </select>
    </div>
    <div class="control-group">
      <label>Protocol Endpoint</label>
      <select id="proto-select" onchange="updateSnippet()">
        <option value="openai" selected>OpenAI Standard (/v1/chat/completions)</option>
        <option value="anthropic">Anthropic Standard (/v1/messages)</option>
        <option value="google">Google Gemini (/v1beta/models/...)</option>
      </select>
    </div>
    <div class="control-group">
      <label>System Prompt (Optional)</label>
      <textarea id="system-prompt" rows="3" placeholder="You are a helpful AI assistant..."></textarea>
    </div>
  </div>

  <div class="chat-container">
    <div class="messages" id="messages-list">
      <div class="msg msg-assistant">
        👋 Welcome to <b>BOB Gemini Free</b>. All requests are translated locally to Google's internal web RPC. Zero cloud bill, zero credit card, 100% free. Type a message to start!
      </div>
    </div>
    <div class="input-area">
      <input type="text" id="user-input" placeholder="Type your message here... (e.g. Write a quicksort in Go with reasoning)" autofocus onkeydown="if(event.key==='Enter') sendMessage()">
      <button class="btn" id="send-btn" onclick="sendMessage()">Send</button>
    </div>
  </div>

  <div class="sidebar-right">
    <div class="section-title">Drop-in Integration Code</div>
    
    <label>Python SDK (OpenAI)</label>
    <div class="snippet-box" id="code-python">from openai import OpenAI
client = OpenAI(base_url="http://127.0.0.1:8081/v1", api_key="none")
res = client.chat.completions.create(
    model="gemini-3.7-flash",
    messages=[{"role": "user", "content": "Hello!"}]
)
print(res.choices[0].message.content)</div>
    <button class="copy-btn" onclick="copySnippet('code-python')">Copy Python</button>

    <div style="height: 16px;"></div>

    <label>Claude Code CLI</label>
    <div class="snippet-box" id="code-claude">export ANTHROPIC_BASE_URL=http://127.0.0.1:8081
export ANTHROPIC_API_KEY=none
claude</div>
    <button class="copy-btn" onclick="copySnippet('code-claude')">Copy Claude Code</button>

    <div style="height: 16px;"></div>

    <label>cURL Stream</label>
    <div class="snippet-box" id="code-curl">curl http://127.0.0.1:8081/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-3.7-flash","stream":true,"messages":[{"role":"user","content":"Hi"}]}'</div>
    <button class="copy-btn" onclick="copySnippet('code-curl')">Copy cURL</button>
  </div>
</main>

<script>
async function refreshTelemetry() {
  try {
    const res = await fetch("/");
    if (res.ok) {
      const data = await res.json();
      document.getElementById("stat-uptime").innerText = data.uptime_seconds + "s";
      document.getElementById("stat-requests").innerText = data.requests_served || 0;
      document.getElementById("stat-tokens").innerText = data.tokens_processed || 0;
      document.getElementById("stat-savings").innerText = data.estimated_savings_usd || "$0.00";
    }
  } catch(e) {}
}
setInterval(refreshTelemetry, 2500);
refreshTelemetry();

function copySnippet(id) {
  const text = document.getElementById(id).innerText;
  navigator.clipboard.writeText(text);
}

let chatHistory = [];

async function sendMessage() {
  const inputEl = document.getElementById("user-input");
  const text = inputEl.value.trim();
  if (!text) return;

  const btn = document.getElementById("send-btn");
  inputEl.value = "";
  inputEl.disabled = true;
  btn.disabled = true;

  const msgList = document.getElementById("messages-list");

  // User bubble
  const userDiv = document.createElement("div");
  userDiv.className = "msg msg-user";
  userDiv.innerText = text;
  msgList.appendChild(userDiv);

  chatHistory.push({role: "user", content: text});

  // Assistant bubble
  const asstDiv = document.createElement("div");
  asstDiv.className = "msg msg-assistant";

  const thoughtBox = document.createElement("div");
  thoughtBox.className = "thought-box";
  thoughtBox.innerHTML = '<div class="thought-header">🧠 Reasoning Process</div><span class="thought-content"></span>';
  asstDiv.appendChild(thoughtBox);

  const contentSpan = document.createElement("div");
  asstDiv.appendChild(contentSpan);
  msgList.appendChild(asstDiv);
  msgList.scrollTop = msgList.scrollHeight;

  let model = document.getElementById("model-select").value;
  const thinkMod = document.getElementById("think-select").value;
  if (thinkMod !== "default") {
    model = model + thinkMod;
  }

  const sysPrompt = document.getElementById("system-prompt").value.trim();
  const reqMessages = [];
  if (sysPrompt) {
    reqMessages.push({role: "system", content: sysPrompt});
  }
  for (const m of chatHistory) {
    reqMessages.push(m);
  }

  try {
    const res = await fetch("/v1/chat/completions", {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({
        model: model,
        stream: true,
        messages: reqMessages
      })
    });

    if (!res.ok) {
      const errJson = await res.json();
      contentSpan.innerText = "Error: " + JSON.stringify(errJson);
      return;
    }

    const reader = res.body.getReader();
    const decoder = new TextDecoder();
    let partial = "";
    let asstFullText = "";
    let thoughtFullText = "";

    while (true) {
      const {done, value} = await reader.read();
      if (done) break;
      partial += decoder.decode(value, {stream: true});

      const lines = partial.split("\n");
      partial = lines.pop() || "";

      for (const line of lines) {
        const trimmed = line.trim();
        if (!trimmed || trimmed === "data: [DONE]") continue;
        if (trimmed.startsWith("data: ")) {
          try {
            const data = JSON.parse(trimmed.slice(6));
            const delta = data.choices && data.choices[0] ? data.choices[0].delta : null;
            if (delta) {
              if (delta.reasoning_content) {
                thoughtBox.classList.add("active");
                thoughtFullText += delta.reasoning_content;
                thoughtBox.querySelector(".thought-content").innerText = thoughtFullText;
              }
              if (delta.content) {
                asstFullText += delta.content;
                contentSpan.innerText = asstFullText;
              }
            }
          } catch(e) {}
        }
      }
      msgList.scrollTop = msgList.scrollHeight;
    }
    chatHistory.push({role: "assistant", content: asstFullText});
  } catch(err) {
    contentSpan.innerText = "Connection error: " + err.message;
  } finally {
    inputEl.disabled = false;
    btn.disabled = false;
    inputEl.focus();
    refreshTelemetry();
  }
}
</script>
</body>
</html>`

func (a *App) handlePlayground(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(playgroundHTML))
}
