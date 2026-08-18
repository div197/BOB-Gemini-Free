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

## 🚀 Example Requests

### 1. Standard URL Format (cURL)
```bash
curl -X POST http://127.0.0.1:8081/v1/images/generations \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "A majestic snow leopard on a Himalayan cliff at sunset, cinematic lighting, 8k",
    "model": "gemini-nano-banana-2"
  }'
```

### 2. Base64 Encoded Format (`response_format: "b64_json"`)
```bash
curl -X POST http://127.0.0.1:8081/v1/images/generations \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "An intricate brass astronomical clock with Vedic constellations",
    "model": "imagen-3",
    "response_format": "b64_json"
  }'
```

---

## 🐍 Python SDK (`openai.images.generate`)

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://127.0.0.1:8081/v1",
    api_key="none"
)

response = client.images.generate(
    model="gemini-nano-banana-2",
    prompt="A futuristic solar powered smart city in Rajasthan with green gardens",
    n=1,
    size="1024x1024"
)

print("Generated Image URL:", response.data[0].url)
```
