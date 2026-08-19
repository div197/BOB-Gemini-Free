# BOB Gemini Free — Static Web Studio (`web/`)

This directory contains the zero-server, 100% static, local-first production web studio for **BOB Gemini Free** by **ABCsteps** ([abcsteps.com](https://abcsteps.com/)) and **Divyanshu Singh Chouhan** ([@div197](https://github.com/div197)).

---

## 🌟 100% Unlimited Scalability: The Local-First Architecture

Unlike server-side proxy architectures that incur cloud bandwidth bills, API request limits (e.g. Cloudflare Workers 100k req/day free limit), or centralized data privacy concerns:

* 🚀 **Zero Server Compute Costs**: The frontend is 100% static HTML, CSS, and JavaScript. It can be hosted on **Cloudflare Pages**, **GitHub Pages**, **Vercel**, **Netlify**, or AWS S3 for **unlimited requests at $0.00 cost**.
* 🔒 **100% On-Device Privacy & Zero Cloud Tracking**: All prompts, system instructions, thinking reasoning streams, and multi-modal attachments stream directly between the user's browser and their local gateway engine (`http://127.0.0.1:9610`). No conversations ever touch an intermediate server.
* ⚡ **Chrome Private Network Access (PNA) Compliant**: Fully compliant with modern browser security standards. The static HTTPS web application (`https://bob-gemini-free.abcsteps.com`) connects to private loopback addresses (`http://127.0.0.1:9610`) seamlessly.
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
