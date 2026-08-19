# Docker & OrbStack Deployment Guide

BOB Gemini Free includes a lightweight multi-stage Docker build producing an ultra-compact **<25MB image** based on `alpine:3.21` with native healthchecks.

---

## 🚀 Quick Run (Anonymous Mode)

```bash
docker run -d \
  --name bob-gemini-free \
  -p 9610:9610 \
  bob-gemini-free:local
```

---

## 🔑 Authenticated Mode (With Session Cookie / Pool)

Mount your `cookie.txt` (or `cookies/` folder) read-only into `/app/`:

```bash
docker run -d \
  --name bob-gemini-free \
  -p 9610:9610 \
  -v $(pwd)/cookie.txt:/app/cookie.txt:ro \
  -e BOB_GEMINI_FREE_COOKIE_FILE=/app/cookie.txt \
  bob-gemini-free:local
```

---

## 🐳 Docker Compose Deployment

Use the pre-configured `docker-compose.yml`:

```yaml
services:
  bob-gemini-free:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: bob-gemini-free
    environment:
      - BOB_GEMINI_FREE_HOST=0.0.0.0
      - BOB_GEMINI_FREE_PORT=9610
    ports:
      - "9610:9610"
    volumes:
      - ./config.json:/app/config.json:ro
      - ./cookie.txt:/app/cookie.txt:ro
    healthcheck:
      test: ["CMD-SHELL", "wget -qO- http://127.0.0.1:9610/ >/dev/null 2>&1 || exit 1"]
      interval: 20s
      timeout: 3s
      retries: 3
      start_period: 3s
    restart: unless-stopped
```

Start with:
```bash
docker compose up -d
```

---

## ⚡ OrbStack Specific Advantages

When running under **OrbStack** on macOS:
- **Instant Cold Boot**: Gateway starts in **<3ms**.
- **Domain Access**: Access directly via `http://bob-gemini-free.orb.local/` or `http://127.0.0.1:9610`.
- **Zero Overhead**: CPU emulation and battery utilization are negligible compared to standard Docker Desktop.
- **Native Healthcheck Indicator**: OrbStack shows a green health badge immediately.
