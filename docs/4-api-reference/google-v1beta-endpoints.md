# Google Native Gemini v1beta Endpoints Reference

BOB Gemini Free implements Google's official Gemini REST endpoints on `/v1beta/*`.

---

## 1. List Models (`GET /v1beta/models`)

```bash
curl http://127.0.0.1:8081/v1beta/models
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
