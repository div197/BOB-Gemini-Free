"""
BOB Gemini Free - Anthropic Python SDK Example
Run BOB Gemini Free gateway: ./bob-gemini-free --port 9610
Then run: pip install anthropic && python anthropic_messages.py
"""

from anthropic import Anthropic

# 1. Initialize client pointing to local BOB gateway
client = Anthropic(
    base_url="http://127.0.0.1:9610",
    api_key="none",
)

print("--- 1. Synchronous Anthropic Messages API ---")
response = client.messages.create(
    model="claude-3-7-sonnet",
    max_tokens=1024,
    messages=[
        {"role": "user", "content": "What are 3 key principles of distributed systems?"}
    ],
)
for block in response.content:
    if getattr(block, "type", "") == "thinking":
        print(f"[Thinking]: {block.thinking}")
    elif getattr(block, "type", "") == "text":
        print(f"[Text]: {block.text}")

print("\n--- 2. Real-Time Streaming with Extended Thinking ---")
with client.messages.stream(
    model="claude-3-7-sonnet",
    max_tokens=2048,
    thinking={"type": "enabled", "budget_tokens": 4000},
    messages=[
        {"role": "user", "content": "How many 'r' letters are in strawberry?"}
    ],
) as stream:
    for event in stream:
        if event.type == "content_block_delta":
            delta = event.delta
            if getattr(delta, "type", "") == "thinking_delta":
                print(delta.thinking, end="", flush=True)
            elif getattr(delta, "type", "") == "text_delta":
                print(delta.text, end="", flush=True)

print("\n\nDone!")
