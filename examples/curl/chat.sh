#!/usr/bin/env bash
# Standard OpenAI Chat Completion
curl -s http://127.0.0.1:9610/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-3.7-flash",
    "messages": [
      {"role": "user", "content": "Explain Archimedes principle in simple terms."}
    ]
  }' | jq .
