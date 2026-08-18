# OpenAI Standard Endpoints Reference

BOB Gemini Free implements standard OpenAI REST endpoints on `/v1/*`.

---

## 1. Chat Completions (`POST /v1/chat/completions`)

### Request
```json
{
  "model": "gemini-3.7-flash",
  "messages": [
    {"role": "system", "content": "You are a helpful coding assistant."},
    {"role": "user", "content": "Write a quicksort in Go."}
  ],
  "stream": true,
  "stream_options": {
    "include_usage": true
  }
}
```

### Thinking & Reasoning Control (`reasoning_effort` or `@think=N`)
* `"reasoning_effort": "high"` or `gemini-3.7-flash@think=0` $\rightarrow$ Deep step-by-step reasoning (up to 20k+ chars).
* `"reasoning_effort": "medium"` $\rightarrow$ Balanced thinking.
* `"reasoning_effort": "low"` or `gemini-3.7-flash@think=4` $\rightarrow$ Fast, concise output.

### Streaming Delta Response Format (`data: ...`)
```json
data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1700000000,"model":"gemini-3.7-flash","choices":[{"index":0,"delta":{"role":"assistant","content":"package main\n\nfunc quicksort..."}}]}
data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1700000000,"model":"gemini-3.7-flash","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":12,"completion_tokens":85,"total_tokens":97}}
data: [DONE]
```

---

## 2. Models Catalog (`GET /v1/models`)

Returns the complete catalog of registered Gemini, OpenAI alias, and Anthropic alias models with permission metadata.

```bash
curl http://127.0.0.1:8081/v1/models
```

### Response
```json
{
  "object": "list",
  "data": [
    {
      "id": "gemini-3.7-flash",
      "object": "model",
      "created": 1700000000,
      "owned_by": "google",
      "permission": [{"id":"modelperm-123","object":"model_permission","allow_create_engine":false,"allow_sampling":true,"allow_logprobs":true,"allow_search_indices":false,"allow_view":true,"allow_fine_tuning":false,"organization":"*","group":null,"is_blocking":false}],
      "root": "gemini-3.7-flash",
      "parent": null
    }
  ]
}
```

---

## 3. Single Model Lookup (`GET /v1/models/{model}`)

```bash
curl http://127.0.0.1:8081/v1/models/gemini-3.7-flash
```

---

## 4. Responses API (`POST /v1/responses`)

For OpenAI Codex CLI and terminal agents:

```bash
curl http://127.0.0.1:8081/v1/responses \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-3.7-flash",
    "input": "Write a python fibonacci function",
    "instructions": "Return clean code only"
  }'
```
