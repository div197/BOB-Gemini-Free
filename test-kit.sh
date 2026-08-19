#!/usr/bin/env bash
set -e

# ==============================================================================
# BOB Gemini Free - Universal Diagnostic & Test Kit
# Break Ordinary Boundaries | Powered by ABCsteps (https://abcsteps.com)
# Author: Divyanshu Singh Chouhan (@div197)
# ==============================================================================

TARGET_URL="${1:-http://127.0.0.1:9610}"
API_KEY="${2:-}"

if [ -f "./bob-gemini-free" ]; then
    if [ -n "$API_KEY" ]; then
        ./bob-gemini-free --test --test-url "$TARGET_URL" --test-key "$API_KEY"
    else
        ./bob-gemini-free --test --test-url "$TARGET_URL"
    fi
else
    echo "[*] Compiling BOB Gemini Free binary..."
    CGO_ENABLED=0 go build -ldflags="-s -w" -o bob-gemini-free .
    if [ -n "$API_KEY" ]; then
        ./bob-gemini-free --test --test-url "$TARGET_URL" --test-key "$API_KEY"
    else
        ./bob-gemini-free --test --test-url "$TARGET_URL"
    fi
fi
