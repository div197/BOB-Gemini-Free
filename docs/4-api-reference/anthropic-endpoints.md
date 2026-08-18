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
  "system": "You are a software architect.",
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

### Streaming Event Lifecycle
1. `event: message_start` $\rightarrow$ Initializes message ID, model name, and input usage.
2. `event: content_block_start` $\rightarrow$ Starts text or thinking block.
3. `event: content_block_delta` $\rightarrow$ Delivers incremental `text_delta` chunks.
4. `event: content_block_stop` $\rightarrow$ Concludes current block.
5. `event: message_delta` $\rightarrow$ Emits `stop_reason: "end_turn"` and final `output_tokens`.
6. `event: message_stop` $\rightarrow$ Closes stream.
