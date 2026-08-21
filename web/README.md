# BOB Gemini Free — Static Web Studio (`web/`)

This directory contains the static, local-first web studio for **BOB Gemini Free** by **ABCsteps** ([abcsteps.com](https://abcsteps.com/)) and **Divyanshu Singh Chouhan** ([@div197](https://github.com/div197)). The static bundle does not replace the Go gateway or Google's upstream/session limits.

---

## 🌟 Static Hosting with a Local-First Architecture

Unlike server-side proxy architectures that incur cloud bandwidth bills, API request limits (e.g. Cloudflare Workers 100k req/day free limit), or centralized data privacy concerns:

* 🚀 **Static Hosting**: The frontend is static HTML, CSS, and JavaScript and can be hosted on **Cloudflare Pages**, **GitHub Pages**, **Vercel**, **Netlify**, or AWS S3. Hosting scale does not establish unlimited gateway or Google capacity.
* 🔒 **Local Gateway Privacy Boundary**: When configured to use a local gateway, prompts and attachments go from the browser to that gateway; the Go gateway sends no automatic telemetry. CDN assets and browser input tools remain separate network dependencies.
* ⚡ **Origin-Gated Private Network Access**: A hosted HTTPS application must be explicitly listed in the gateway's `allowed_origins` configuration. PNA is a browser permission mechanism, not authentication; configure API keys for application-level access.
* 📱 **PWA & Offline Ready**: Installable as a standalone native-feeling desktop app on macOS Dock, Windows Taskbar, iOS Home Screen, or Android.

---

## 🚀 1-Click Deployment

### Option 1: Deploy to Cloudflare Pages (Recommended for `bob-gemini-free.abcsteps.com`)

1. Go to [Cloudflare Dashboard](https://dash.cloudflare.com/) → **Workers & Pages** → **Create application** → **Pages** → **Connect to Git**.
2. Select repository: `div197/bob-gemini-free`.
3. Configure Build Settings:
   - **Framework preset**: `None`
   - **Build command**: `make web` (or leave blank)
   - **Build output directory**: `web`
4. Click **Save and Deploy**.
5. Under **Custom domains**, add `bob-gemini-free.abcsteps.com`.

### Option 2: Deploy to GitHub Pages

1. In repository settings → **Pages** → **Build and deployment**.
2. Source: **Deploy from a branch**.
3. Branch: `main` / Folder: `/web` (or deploy via GitHub Actions).
4. Custom domain: `bob-gemini-free.abcsteps.com`.

---

## 🛠️ Local Development & Testing

To preview the static studio locally:

```bash
# Start BOB Gateway Daemon in background
./bob-gemini-free --port 9610 &

# Serve the static web directory with Python HTTP server
cd web && python3 -m http.server 8080
```

Open `http://127.0.0.1:8080` in your browser. The connection badge will automatically detect the local engine on port 9610 and turn green.
