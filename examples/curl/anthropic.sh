#!/usr/bin/env bash
# Anthropic Messages API (Claude Code standard)
curl -s http://127.0.0.1:9610/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-7-sonnet",
    "max_tokens": 1024,
    "messages": [
      {"role": "user", "content": "What is the capital of France?"}
    ]
  }' | jq .
