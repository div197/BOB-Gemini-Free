#!/usr/bin/env bash
# Real-Time Streaming with Deep Reasoning Thinking Tokens
curl -N http://127.0.0.1:9610/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-3.7-flash-thinking",
    "stream": true,
    "messages": [
      {"role": "user", "content": "Solve step by step: What is 17 * 19?"}
    ]
  }'
