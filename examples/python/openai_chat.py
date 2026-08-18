"""
BOB Gemini Free - OpenAI Python SDK Example
Run BOB Gemini Free gateway: ./bob-gemini-free --port 8081
Then run: pip install openai && python openai_chat.py
"""

from openai import OpenAI

# 1. Initialize client pointing to local BOB gateway
client = OpenAI(
    base_url="http://127.0.0.1:8081/v1",
    api_key="none",  # or your configured BOB_GEMINI_FREE_API_KEYS
)

print("--- 1. Synchronous Completion ---")
response = client.chat.completions.create(
    model="gemini-3.7-flash",
    messages=[
        {"role": "user", "content": "Explain quantum superposition in 2 sentences."}
    ],
)
print("Response:\n", response.choices[0].message.content)

print("\n--- 2. Real-Time Streaming with Deep Reasoning ---")
stream = client.chat.completions.create(
    model="gemini-3.7-flash-thinking",
    messages=[
        {"role": "user", "content": "Solve: What is the sum of the first 10 prime numbers?"}
    ],
    stream=True,
)

for chunk in stream:
    delta = chunk.choices[0].delta if chunk.choices else None
    if delta:
        # Check for real-time reasoning content
        if getattr(delta, "reasoning_content", None):
            print(f"[Thinking]: {delta.reasoning_content}", end="", flush=True)
        # Check for clean answer content
        if delta.content:
            print(delta.content, end="", flush=True)

print("\n\nDone!")
