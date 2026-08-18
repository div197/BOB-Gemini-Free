# Multi-Account Cookie Pool & Auto-Rotation

BOB Gemini Free includes an enterprise-grade **Multi-Account Cookie Pool Engine** (`CookiePool`) that provides high-throughput load balancing and transparent failover across multiple Google accounts.

---

## 🌟 Why Use a Cookie Pool?

1. **High Concurrency**: Distribute large batches of agent queries across multiple Google accounts.
2. **Auto-Failover**: If an account hits a temporary rate limit (HTTP 429), BOB automatically puts that account on a 60-second cooldown and transparently retries on the next healthy account.
3. **Zero Interruption**: The client's streaming response never breaks during failover.

---

## 📁 Directory Setup (`cookies/`)

Create a `cookies/` directory in your working folder and drop individual cookie files:

```
./cookies/
├── primary_account.txt
├── secondary_account.txt
└── team_backup.txt
```

BOB Gemini Free automatically scans `./cookies/*.txt` and `~/.config/bob-gemini-free/cookies/*.txt` on startup.

---

## ⚙️ Configuration Setup (`config.json`)

You can also specify explicit paths or accounts in `config.json`:

```json
{
  "cookie_pool": [
    "./cookies/primary_account.txt",
    "./cookies/secondary_account.txt",
    "/shared/team_cookie.txt"
  ]
}
```

Or via environment variable:
```bash
export BOB_GEMINI_FREE_COOKIE_POOL="./cookies/acc1.txt,./cookies/acc2.txt"
./bob-gemini-free
```

---

## 🔄 Load Balancing Mechanics

- **Algorithm**: Thread-safe atomic round-robin (`cursor.Add(1) % total`).
- **Health Tracking**: Accounts that trigger upstream transport or HTTP errors are placed on temporary cooldown.
- **Auto-Recovery**: As cooldown expires or requests succeed, failure counters reset automatically.
