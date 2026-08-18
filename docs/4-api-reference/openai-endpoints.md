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

### Thinking / Reasoning Depth (`reasoning_effort`)
You can control reasoning depth dynamically:
* `"reasoning_effort": "high"` $\rightarrow$ Deep reasoning
* `"reasoning_effort": "medium"` $\rightarrow$ Moderate reasoning
* `"reasoning_effort": "low"` $\rightarrow$ Concise direct answer

---

## 2. Models Catalog (`GET /v1/models`)

Returns the full catalog of registered Gemini, OpenAI alias, and Anthropic alias models with permission metadata.

```bash
curl http://127.0.0.1:8081/v1/models
```

---

## 3. Single Model Lookup (`GET /v1/models/{model}`)

Returns metadata for a specific model ID.

```bash
curl http://127.0.0.1:8081/v1/models/gemini-3.7-flash
```
