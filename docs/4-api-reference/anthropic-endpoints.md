# Anthropic Messages Endpoints Reference

BOB Gemini Free implements Anthropic's **Messages API** (`POST /v1/messages`) with full Server-Sent Events (SSE) streaming and thinking blocks.

---

## Endpoint: `POST /v1/messages`

### Headers
```http
Content-Type: application/json
x-api-key: none
anthropic-version: 2023-06-01
```

### Request Payload
```json
{
  "model": "claude-3-7-sonnet",
  "system": "You are an expert system architect.",
  "messages": [
    {"role": "user", "content": "Explain raft consensus in simple terms."}
  ],
  "max_tokens": 1024,
  "stream": true,
  "thinking": {
    "type": "enabled",
    "budget_tokens": 2048
  }
}
```

---

## 🌊 SSE Event Lifecycle Stream

When `stream: true` is requested, BOB Gemini Free emits standard Anthropic SSE events:

```text
event: message_start
data: {"type":"message_start","message":{"id":"msg_123","type":"message","role":"assistant","content":[],"model":"claude-3-7-sonnet","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":14,"output_tokens":0}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"thinking","thinking":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"Raft consensus relies on leader election and log replication..."}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: content_block_start
data: {"type":"content_block_start","index":1,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":"Raft is an algorithm used to manage a replicated log..."}}

event: content_block_stop
data: {"type":"content_block_stop","index":1}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":85}}

event: message_stop
data: {"type":"message_stop"}
```
