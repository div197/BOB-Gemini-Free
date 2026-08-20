#!/bin/bash
# Build the Go binary
make -C .. build

# Tauri passes the target triple in the TARGET environment variable
# If TARGET is empty, fallback to aarch64-apple-darwin for local testing
TARGET_TRIPLE=${TARGET:-aarch64-apple-darwin}

echo "Bundling sidecar for target: $TARGET_TRIPLE"
mkdir -p src-tauri/binaries
cp ../bob-gemini-free src-tauri/binaries/bob-gemini-free-$TARGET_TRIPLE
