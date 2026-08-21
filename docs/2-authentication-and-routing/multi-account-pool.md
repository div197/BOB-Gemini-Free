# Multi-Account Cookie Pool & Auto-Rotation

BOB Gemini Free includes a local **Multi-Account Cookie Pool Engine**
(`CookiePool`) that rotates configured sessions and applies cooldown/failover
logic. It does not establish higher upstream quotas, provider approval, or a
guarantee of uninterrupted streams.

---

## 🌟 Why Use a Cookie Pool?

1. **Session Distribution**: Distribute requests across explicitly configured sessions; actual safe concurrency remains provider-dependent.
2. **Cooldown/Failover Logic**: A selected account can be backed off for 60 seconds after a classified failure and the retry path can select another healthy session; stream continuity is not guaranteed.
3. **Atomic Selection**: Uses atomic round-robin selection in the local pool; no standalone latency claim is made.

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
