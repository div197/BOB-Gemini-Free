# Classroom LAN Deployment Guide

This guide is for computer labs, classrooms, workshops, and training rooms where many students send requests from the same physical location at the same time.

The key decision is simple:

- Use the public Cloudflare Pages studio for demos, exploration, and geographically distributed visitors.
- Use the local Go gateway on the lab network for a real classroom session with many simultaneous students.

## Why Cloudflare Can Rate Limit in a Classroom Burst

The Cloudflare Pages studio is useful because it is zero-install and globally accessible. It is not the strongest path for a concentrated 30-student burst from one classroom.

When the serverless function calls Gemini from Cloudflare, the upstream request originates from Cloudflare datacenter IP space, not from the student's Mac or the school's local network. If many anonymous requests arrive together from the same datacenter egress pool, Google may classify the pattern as automated datacenter traffic and return an upstream HTTP `429` or `403`.

That does not prove the school network, the teacher's Mac, or the local binary is blocked. It means the Cloudflare egress path is being rate limited or challenged.

## Recommended Classroom Topology

Run one BOB Gemini Free process on a teacher machine or local lab server:

```bash
./bob-gemini-free --host 0.0.0.0 --port 9610 --cookie-pool-dir ./cookies
```

Students connect to that machine over the LAN:

```text
http://TEACHER_LAN_IP:9610/playground
```

Example:

```text
http://192.168.1.50:9610/playground
```

For API clients:

```text
http://192.168.1.50:9610/v1
```

This keeps classroom traffic inside the LAN until the single local gateway makes upstream Gemini web requests.

## StreamFlight Deduplication & High-Concurrency Lab Scaling (v0.1.8)

In a live computer lab, multiple students often submit similar or identical prompts (e.g. *"Build a 2D CyberSnake game in HTML5 Canvas"* or *"Derive the Schrödinger wave equation"*).

In BOB Gemini Free v0.1.8:
- **StreamFlight Multiplexer (`internal/gemini/flight.go`)**: Automatically detects concurrent in-flight requests with matching normalized prompts and multiplexes a single upstream Google connection across unlimited downstream students over non-blocking atomic channels.
- **Thundering-Herd Protection (`internal/gemini/auth.go`)**: Guest session tokens are managed via Stale-While-Revalidate with a single background refresh lock, guaranteeing 0ms blocking when 500 students connect simultaneously.
- **Dynamic Keep-Alive Heartbeat (`internal/server/helpers.go`)**: Periodic `: keepalive\n\n` comments keep student browser sockets open during deep thinking phases, permanently eliminating gateway timeouts.
- **Auto-Guest Fallback on Cookie Cooldown**: If an account cookie in `./cookies/` expires or triggers a temporary cooldown, BOB automatically routes traffic to live anonymous guest mode, preventing HTTP 503 errors.

## Cookie Pool Setup for Labs

For authenticated Pro or Vision workloads, prepare a small pool of browser sessions:

```text
./cookies/
├── account_1.txt
├── account_2.txt
├── account_3.txt
└── account_4.txt
```

Then start:

```bash
./bob-gemini-free --host 0.0.0.0 --port 9610 --cookie-pool-dir ./cookies
```

The pool rotates between sessions and applies a 60-second cooldown on transient anomalies. If all pool accounts are in cooldown, BOB automatically falls back to live guest mode.

Use only accounts and sessions you are authorized to use. Do not publish cookie files. Do not commit `cookie.txt` or `cookies/*.txt`.

## Pre-Class Verification

Run these checks before students arrive.

### 1. Confirm the gateway is reachable from the teacher machine

```bash
curl http://127.0.0.1:9610/
```

Expected result: JSON with `"status":"ok"`.

### 2. Confirm the gateway is reachable from a student machine

Replace `TEACHER_LAN_IP` with the teacher machine's LAN IP:

```bash
curl http://TEACHER_LAN_IP:9610/
```

Expected result: JSON with `"status":"ok"`.

If this fails but localhost works on the teacher machine, check macOS firewall, Windows firewall, router client isolation, and whether the server was started with `--host 0.0.0.0`.

### 3. Run a small live request

```bash
curl http://TEACHER_LAN_IP:9610/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-3.7-flash",
    "messages": [{"role": "user", "content": "Reply with exactly one word."}],
    "stream": false
  }'
```

Expected result: an OpenAI-compatible JSON response.

### 4. Run a bounded benchmark

Start with a conservative benchmark:

```bash
./bob-gemini-free --bench --bench-url http://127.0.0.1:9610 --bench-concurrency 3 --bench-requests 6
```

If that passes, increase gradually. Do not start by blasting 30 or 100 concurrent requests. Upstream rate limits are dynamic and depend on account state, IP reputation, time, and recent traffic.

## Cloudflare Demo Stress Test Boundary

A Cloudflare stress test can tell you whether the public demo is currently being rate-limited from Cloudflare egress. It should not be treated as proof that the local gateway will fail.

If the public demo returns `429` or `403` during a 30-student burst, use the local LAN gateway for class. Keep the public demo for global discovery and lightweight demonstrations.

## Operational Decision Table

| Scenario | Recommended Mode | Why |
| --- | --- | --- |
| One visitor trying the project from GitHub | Cloudflare Pages studio | Zero install, low friction |
| Global visitors spread across locations | Cloudflare Pages studio | Traffic is distributed across edge locations |
| 30 students in one lab at the same time | Local Go gateway on LAN | Avoids concentrated Cloudflare datacenter egress |
| Authenticated vision / Imagen / Pro routing | Local Go gateway with cookies | Requires browser session cookies |
| Paid workshop or exam-like session | Local Go gateway with cookie pool | Better observability and operational control |

## What This Does Not Prove

This topology improves the classroom architecture, but it does not mathematically guarantee unlimited upstream capacity. Google can still apply dynamic limits to accounts, sessions, IPs, or traffic patterns.

Before a live class, prove the current environment with the pre-class checks above.
