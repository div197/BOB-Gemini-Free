/**
 * BOB Gemini Free - Cloudflare Pages Edge Function
 * POST /v1/chat/completions
 * 100% Free Serverless Edge Proxy to Google Gemini Web RPC
 */

function uuidV4() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

function cleanText(text, strip = false) {
  text = text.replace(/```(?:python|javascript|text)\?code_(?:reference|stdout)&code_event_index=\d+\n[\s\S]*?```\n?/g, '');
  text = text.replace(/http:\/\/googleusercontent\.com\/card_content\/\d+\n?/g, '');
  return strip ? text.trim() : text;
}

function extractTextsFromLine(line) {
  if (!line.includes('"wrb.fr"') || line.length < 200) return [];
  try {
    const arr = JSON.parse(line);
    if (!Array.isArray(arr) || arr.length === 0) return [];
    const firstElem = arr[0];
    if (!Array.isArray(firstElem) || firstElem.length < 3) return [];
    const innerStr = firstElem[2];
    if (typeof innerStr !== 'string' || innerStr.length < 50) return [];
    const inner = JSON.parse(innerStr);
    if (!Array.isArray(inner) || inner.length <= 4 || !inner[4]) return [];
    const parts = inner[4];
    if (!Array.isArray(parts)) return [];
    
    const texts = [];
    for (const part of parts) {
      if (Array.isArray(part) && part.length > 1 && part[1]) {
        const tList = part[1];
        if (Array.isArray(tList)) {
          for (const t of tList) {
            if (typeof t === 'string' && t) texts.push(t);
          }
        }
      }
    }
    return texts;
  } catch (e) {
    return [];
  }
}

class StreamParser {
  constructor() {
    this.prevText = "";
    this.buf = "";
  }

  feed(chunk) {
    this.buf += chunk;
    const deltas = [];

    while (this.buf.includes("\n")) {
      const idx = this.buf.indexOf("\n");
      const line = this.buf.slice(0, idx);
      this.buf = this.buf.slice(idx + 1);

      const texts = extractTextsFromLine(line);
      for (const t of texts) {
        if (t === this.prevText || this.prevText.startsWith(t)) continue;

        if (t.startsWith(this.prevText)) {
          const delta = cleanText(t.slice(this.prevText.length), false);
          this.prevText = t;
          if (delta) deltas.push(delta);
          break;
        } else if (this.prevText === "") {
          const delta = cleanText(t, false);
          this.prevText = t;
          if (delta) deltas.push(delta);
          break;
        } else if (t.length > this.prevText.length) {
          const delta = cleanText(t.slice(this.prevText.length), false);
          this.prevText = t;
          if (delta) deltas.push(delta);
          break;
        }
      }
    }
    return deltas;
  }
}

function buildBody(prompt, modelID = -1, thinkMode = 1) {
  const inner = new Array(102).fill(null);
  inner[0] = [prompt, 0, null, null, null, null, 0];
  inner[1] = ["en"];
  inner[2] = ["", "", "", null, null, null, null, null, null, ""];
  inner[6] = [0];
  inner[7] = 1;
  inner[10] = 1;
  inner[11] = 0;
  inner[17] = [[thinkMode]];
  inner[18] = 0;
  inner[27] = 1;
  inner[30] = [4];
  inner[41] = [2];
  inner[53] = 0;
  inner[59] = uuidV4();
  inner[61] = [];
  inner[68] = 1;
  inner[79] = modelID;

  const outer = [null, JSON.stringify(inner)];
  const form = new URLSearchParams();
  form.set("f.req", JSON.stringify(outer));
  return form.toString();
}

function corsHeaders() {
  return {
    "Access-Control-Allow-Origin": "*",
    "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
    "Access-Control-Allow-Headers": "Content-Type, Authorization, X-Requested-With",
    "Access-Control-Allow-Private-Network": "true"
  };
}

export async function onRequestOptions() {
  return new Response(null, {
    status: 204,
    headers: corsHeaders()
  });
}

export async function onRequestPost(context) {
  try {
    const req = context.request;
    const body = await req.json().catch(() => ({}));
    const messages = body.messages || [];
    const model = body.model || "gemini-3.7-flash";
    const stream = body.stream !== false; // default true

    // Extract consolidated prompt
    let promptParts = [];
    for (const m of messages) {
      if (m.role === 'system') {
        promptParts.push(`System Instruction: ${m.content}`);
      } else if (m.role === 'user') {
        if (typeof m.content === 'string') {
          promptParts.push(m.content);
        } else if (Array.isArray(m.content)) {
          const t = m.content.find(c => c.type === 'text');
          if (t && t.text) promptParts.push(t.text);
        }
      } else if (m.role === 'assistant') {
        promptParts.push(`Previous Assistant Response: ${m.content}`);
      }
    }
    const fullPrompt = promptParts.join("\n\n") || "Hello";

    // Model parameters
    let modelID = -1;
    let thinkMode = 1;
    if (model.includes("thinking") || model.includes("think")) {
      thinkMode = 2; // Deep reasoning
    } else if (model.includes("pro")) {
      modelID = 1;
    }

    const reqid = Date.now() % 1000000;
    const upstreamUrl = `https://gemini.google.com/_/BardChatUi/data/assistant.lamda.BardFrontendService/StreamGenerate?bl=boq_assistant-bard-web-server_20250218.06_p0&hl=en&_reqid=${reqid}&rt=c`;

    const upstreamRes = await fetch(upstreamUrl, {
      method: "POST",
      headers: {
        "Content-Type": "application/x-www-form-urlencoded",
        "Origin": "https://gemini.google.com",
        "Referer": "https://gemini.google.com/app",
        "X-Same-Domain": "1",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
      },
      body: buildBody(fullPrompt, modelID, thinkMode)
    });

    if (!upstreamRes.ok) {
      return new Response(JSON.stringify({
        error: {
          message: `Upstream Gemini error: HTTP ${upstreamRes.status}`,
          type: "api_error"
        }
      }), {
        status: 502,
        headers: { ...corsHeaders(), "Content-Type": "application/json" }
      });
    }

    const chatId = "chatcmpl-" + Math.random().toString(36).substring(2, 14);
    const created = Math.floor(Date.now() / 1000);

    if (stream) {
      const { readable, writable } = new TransformStream();
      const writer = writable.getWriter();
      const encoder = new TextEncoder();
      const decoder = new TextDecoder();
      const reader = upstreamRes.body.getReader();
      const parser = new StreamParser();

      // Background stream processing with context.waitUntil to keep Cloudflare isolate alive
      const streamPromise = (async () => {
        try {
          while (true) {
            const { done, value } = await reader.read();
            if (done) break;
            const chunk = decoder.decode(value, { stream: true });
            const deltas = parser.feed(chunk);
            for (const delta of deltas) {
              const sseObj = {
                id: chatId,
                object: "chat.completion.chunk",
                created: created,
                model: model,
                system_fingerprint: "fp_bob_gemini_edge",
                choices: [
                  {
                    index: 0,
                    delta: { content: delta },
                    finish_reason: null
                  }
                ]
              };
              await writer.write(encoder.encode(`data: ${JSON.stringify(sseObj)}\n\n`));
            }
          }

          // Send [DONE]
          const finalSse = {
            id: chatId,
            object: "chat.completion.chunk",
            created: created,
            model: model,
            system_fingerprint: "fp_bob_gemini_edge",
            choices: [
              {
                index: 0,
                delta: {},
                finish_reason: "stop"
              }
            ]
          };
          await writer.write(encoder.encode(`data: ${JSON.stringify(finalSse)}\n\ndata: [DONE]\n\n`));
        } catch (e) {
          console.error("Stream pump error:", e);
        } finally {
          try {
            await writer.close();
          } catch (e) {}
        }
      })();

      if (context && typeof context.waitUntil === 'function') {
        context.waitUntil(streamPromise);
      }

      return new Response(readable, {
        headers: {
          ...corsHeaders(),
          "Content-Type": "text/event-stream; charset=utf-8",
          "Cache-Control": "no-cache, no-transform",
          "Connection": "keep-alive"
        }
      });
    } else {
      // Non-streaming response
      const parser = new StreamParser();
      const reader = upstreamRes.body.getReader();
      const decoder = new TextDecoder();
      let fullText = "";

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        const chunk = decoder.decode(value, { stream: true });
        const deltas = parser.feed(chunk);
        for (const d of deltas) {
          fullText += d;
        }
      }

      return new Response(JSON.stringify({
        id: chatId,
        object: "chat.completion",
        created: created,
        model: model,
        system_fingerprint: "fp_bob_gemini_edge",
        choices: [
          {
            index: 0,
            message: {
              role: "assistant",
              content: fullText.trim()
            },
            finish_reason: "stop"
          }
        ],
        usage: {
          prompt_tokens: Math.round(fullPrompt.length / 4),
          completion_tokens: Math.round(fullText.length / 4),
          total_tokens: Math.round((fullPrompt.length + fullText.length) / 4)
        }
      }), {
        headers: { ...corsHeaders(), "Content-Type": "application/json" }
      });
    }
  } catch (err) {
    return new Response(JSON.stringify({
      error: {
        message: "Internal Edge Error: " + err.message,
        type: "api_error"
      }
    }), {
      status: 500,
      headers: { ...corsHeaders(), "Content-Type": "application/json" }
    });
  }
}
