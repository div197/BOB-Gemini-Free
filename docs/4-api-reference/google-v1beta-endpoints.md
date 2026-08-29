# Google-shaped Gemini v1beta Endpoints Reference

BOB Gemini Free exposes Google-shaped adapter routes on `/v1beta/*`. They are
translated into Google's undocumented web RPC and are not Google's official
API implementation; live behavior remains session/provider-dependent.

There are two explicit modes. Without `X-BOB-Gemini-API-Key`, these adapter
routes use BOB's existing cookie/guest web-RPC path. With that header, BOB
forwards the request to Google's documented Gemini Developer API using the
student's own project key. The latter route is opt-in, does not rotate keys,
and does not silently fall back to web RPC. See
[`GEMINI-API-ROUTING.md`](../engineering/GEMINI-API-ROUTING.md) for the
student setup link, current-limit policy, and supported model-ID boundary.

Example explicit provider request (use only your own key; do not commit it):

```bash
curl -X POST http://127.0.0.1:9610/v1beta/models/gemini-3.7-flash:generateContent \
  -H 'Content-Type: application/json' \
  -H 'X-BOB-Gemini-API-Key: YOUR_OWN_AI_STUDIO_KEY' \
  -d '{"contents":[{"role":"user","parts":[{"text":"Hello"}]}]}'
```

---

## 1. List Models (`GET /v1beta/models`)

```bash
curl http://127.0.0.1:9610/v1beta/models
```

---

## 2. Generate Content (`POST /v1beta/models/{model}:generateContent`)

### Request
```json
{
  "contents": [
    {
      "role": "user",
      "parts": [
        {"text": "What is the distance between Sun and Earth?"}
      ]
    }
  ]
}
```

### Response
```json
{
  "candidates": [
    {
      "index": 0,
      "content": {
        "role": "model",
        "parts": [
          {"text": "The average distance between the Sun and the Earth is approximately 149.6 million kilometers (93 million miles)."}
        ]
      },
      "finishReason": "STOP"
    }
  ],
  "usageMetadata": {
    "promptTokenCount": 9,
    "candidatesTokenCount": 24,
    "totalTokenCount": 33
  },
  "modelVersion": "gemini-3.7-flash"
}
```

---

## 3. Streaming Generate Content (`POST /v1beta/models/{model}:streamGenerateContent`)

Streams real-time Server-Sent Events with candidates array chunks directly to Google SDK clients.

---

## 4. Token Counting (`POST /v1beta/models/{model}:countTokens`)

Estimate token usage for prompt text and multimodal inputs before sending
inference requests. Counts are local estimates, not Google's authoritative
tokenizer, and have no provider quota authority:

### Request
```bash
curl -X POST http://127.0.0.1:9610/v1beta/models/gemini-3.7-flash:countTokens \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [
      {
        "role": "user",
        "parts": [
          {"text": "Explain the architecture of Transformer neural networks."}
        ]
      }
    ]
  }'
```

### Response
```json
{
  "totalTokens": 10
}
```
