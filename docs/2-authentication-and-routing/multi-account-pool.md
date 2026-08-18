# Multi-Account Cookie Pool & Auto-Rotation

BOB Gemini Free includes an enterprise-grade **Multi-Account Cookie Pool Engine** (`CookiePool`) that provides high-throughput load balancing, dynamic failover, and automatic cooldown recovery across multiple Google accounts.

---

## 🌟 Why Use a Cookie Pool?

1. **High Concurrency**: Distribute hundreds of agent queries across multiple Google accounts without hitting individual account burst rate limits.
2. **Transparent Failover (Zero Stream Drops)**: If an account encounters a burst rate-limit (HTTP 429) or session anomaly, BOB automatically backs off that account for 60 seconds and transparently retries on the next healthy account without interrupting the client connection.
3. **Lock-Free Atomic Dispatch**: Uses atomic CPU instructions (`atomic.Uint64`) for ultra-low latency round-robin dispatching (<0.01ms overhead).

---

## 🔄 Lifecycle & State Flow

```
                     ┌─────────────────────────┐
                     │ Incoming Client Request │
                     └────────────┬────────────┘
                                  │
                                  ▼
                     ┌─────────────────────────┐
                     │ Atomic Lock-Free Select │
                     │   (Next Healthy Acct)   │
                     └────────────┬────────────┘
                                  │
                  ┌───────────────┴───────────────┐
                  ▼                               ▼
       ┌───────────────────────┐       ┌───────────────────────┐
       │   Request Succeeded   │       │ Rate Limit (HTTP 429) │
       │     (Return Stream)   │       │   or Session Anomaly  │
       └───────────────────────┘       └──────────┬────────────┘
                                                  │
                                                  ▼
                                       ┌───────────────────────┐
                                       │ Mark 60s Backoff      │
                                       │ Transparent Retry on  │
                                       │ Next Healthy Account  │
                                       └───────────────────────┘
```

---

## 📁 Directory Setup (`cookies/`)

Simply create a `cookies/` directory in your project folder and drop individual `.txt` cookie files:

```
./cookies/
├── primary_account.txt
├── secondary_account.txt
├── team_alpha.txt
└── team_beta.txt
```

BOB Gemini Free automatically scans `./cookies/*.txt` and `~/.config/bob-gemini-free/cookies/*.txt` on startup.

---

## ⚙️ Configuration Setup (`config.json`)

You can also specify explicit paths in `config.json`:

```json
{
  "cookie_pool": [
    "./cookies/primary_account.txt",
    "./cookies/secondary_account.txt",
    "/shared/secrets/team_cookie.txt"
  ]
}
```

Or pass via environment variable:

```bash
export BOB_GEMINI_FREE_COOKIE_POOL="./cookies/acc1.txt,./cookies/acc2.txt,/shared/acc3.txt"
./bob-gemini-free
```

---

## 📊 Pool Telemetry & Logging

When accounts rotate or recover, BOB Gemini Free logs clean operational status:

```text
2026/08/18 14:06:16 [CookiePool] Loaded 3 active Google account sessions
2026/08/18 14:06:25 [CookiePool] Session primary_account marked with failure (cooldown 60s)
2026/08/18 14:06:25 [CookiePool] Transparently routing request to secondary_account...
2026/08/18 14:07:25 [CookiePool] Session primary_account cooldown expired, restored to active pool
```
