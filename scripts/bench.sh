#!/usr/bin/env bash
set -e

# ==============================================================================
# BOB Gemini Free - Live Performance & Stress Benchmark Runner
# Break Ordinary Boundaries | Powered by ABCsteps (https://abcsteps.com)
# Author: Divyanshu Singh Chouhan (@div197)
# ==============================================================================

TARGET_URL="${1:-http://127.0.0.1:8081}"
CONCURRENCY="${2:-3}"
REQUESTS="${3:-6}"
API_KEY="${4:-}"

if [ -f "./bob-gemini-free" ]; then
    ./bob-gemini-free --bench --test-url "$TARGET_URL" --bench-concurrency "$CONCURRENCY" --bench-requests "$REQUESTS" --test-key "$API_KEY"
else
    go run . --bench --test-url "$TARGET_URL" --bench-concurrency "$CONCURRENCY" --bench-requests "$REQUESTS" --test-key "$API_KEY"
fi
