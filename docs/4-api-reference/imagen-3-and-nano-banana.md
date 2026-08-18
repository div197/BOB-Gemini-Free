# Image Generation: Imagen 3 & Gemini Nano Banana

BOB Gemini Free exposes standard OpenAI image generation endpoints (`POST /v1/images/generations`) backed by Google's state-of-the-art **Imagen 3** and **Gemini Nano Banana 2 / Pro** visual engines.

---

## 🎨 Supported Image Models

| Model ID | Visual Engine | Description |
| :--- | :--- | :--- |
| `imagen-3` | Google Imagen 3 | Ultra-high-fidelity photorealistic generation |
| `imagen-3-fast` | Google Imagen 3 Fast | Low-latency rapid image generation |
| `gemini-nano-banana` | Gemini Nano Banana | Google's native multimodal image generation model |
| `gemini-nano-banana-2` | Gemini Nano Banana 2 | Latest native multimodal Banana 2 image model |
| `gemini-nano-banana-pro` | Gemini Nano Banana Pro | High-resolution Mode 3 visual synthesis |
| `dall-e-3` | DALL-E 3 Alias | OpenAI drop-in alias routed to Imagen 3 |
| `dall-e-2` | DALL-E 2 Alias | OpenAI drop-in alias routed to Imagen 3 Fast |

---

## 🚀 Example Request

### URL Format Response (Default)
```bash
curl -X POST http://127.0.0.1:8081/v1/images/generations \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "A serene Himalayan mountain landscape at sunrise, watercolor style",
    "model": "gemini-nano-banana-2"
  }'
```

### Base64 JSON Format (`response_format: "b64_json"`)
```bash
curl -X POST http://127.0.0.1:8081/v1/images/generations \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "A golden medallion with intricate Vedic geometry",
    "model": "imagen-3",
    "response_format": "b64_json"
  }'
```
