#!/usr/bin/env bash
# OpenAI Codex CLI Responses API (POST /v1/responses)
curl -s http://127.0.0.1:9610/v1/responses \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-5-codex",
    "input": "Write a bash command to count lines in all .go files."
  }' | jq .
