<p align="center">
  <img src="assets/bob-gemini-free-banner.jpg" alt="BOB Gemini Free Banner" width="100%">
</p>

# BOB Gemini Free

<p align="center">
  <strong>Break Ordinary Boundaries — High-Performance Local OpenAI & Gemini Gateway</strong><br>
  <em>Powered by Google Gemini Web UI</em>
</p>

<p align="center">
  <a href="https://abcsteps.com/"><img src="https://img.shields.io/badge/Powered%20by-ABCsteps.com-blue?style=for-the-badge" alt="ABCsteps"></a>
  <a href="https://github.com/div197/bob-gemini-free"><img src="https://img.shields.io/badge/Author-Divyanshu%20Singh%20Chouhan%20(@div197)-green?style=for-the-badge" alt="Author"></a>
  <img src="https://img.shields.io/badge/Release-v0.1.0-blueviolet?style=for-the-badge" alt="Release">
  <img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=for-the-badge&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT-orange?style=for-the-badge" alt="License">
</p>

---

[English](README.md) | [हिंदी (Hindi)](README.hi.md) | [Changelog](CHANGELOG.md)

**BOB Gemini Free** is part of the **BOB Series** (*Break Ordinary Boundaries*) developed by [**ABCsteps.com**](https://abcsteps.com/) — an online AI engineering school founded by **Divyanshu Singh Chouhan** ([@div197](https://github.com/div197)).

It is a high-performance, single-binary Go gateway that converts Google Gemini's web interface into standard **OpenAI-compatible** (`/v1/chat/completions`, `/v1/models`, `/v1/responses`) and **Gemini-native** (`/v1beta/models`) API endpoints.

---

## The BOB Series (*Break Ordinary Boundaries*)

The **BOB Series** by **ABCsteps** is a developer-first suite of high-impact runtimes, proxies, and automation engines designed to remove paywalls and artificial constraints from modern AI workflows:

* 🎥 [**BOB YouTube**](https://github.com/div197/BOB-Youtube) — Docker-first YouTube ingestion runtime for developers, products, bulk workflows, and AI agents.
* ⚡ [**BOB Gemini Free**](https://github.com/div197/bob-gemini-free) — High-performance OpenAI & Gemini gateway unlocking Google Gemini Web for agents, IDEs, and developer tools.

---

## Key Features

* **Free for Every Gmail User**: Out of the box, every Google account includes free Gemini access with high-speed Flash, adaptive Flash Lite, and deep Flash Thinking (up to 20,000+ characters of reasoning).
* **Gemini Advanced ($20/mo) Integration**: Attach your session cookie to legitimately route to Google's flagship **Pro** model for deep mathematical and coding capabilities.
* **OpenAI Drop-In Replacement**: Seamlessly works with Cherry Studio, ChatBox, Codex CLI, Cursor, OpenAI Python/TypeScript SDKs, and custom AI agents.
* **Full Multimodal Vision**: Send base64 images or image URLs via standard OpenAI payloads — automatically uploaded via Google's Scotty Resumable Upload protocol with automatic compression.
* **Reasoning Control**: Tune thinking depth dynamically via `@think=N` (0 = deepest reasoning, 4 = fast concise output).
* **Zero Cost & Local First**: Single static binary with near-zero memory footprint (<15MB RAM baseline), safe local-first binding (`127.0.0.1`), and high concurrency throughput.
* **TLS Browser Impersonation**: Built-in support to mimic Chrome/Firefox/Safari TLS fingerprints via `tls-client` for network environments facing WAF restrictions.

---

## Quick Start (Zero-Friction for All Users)

### Option 1: Automatic Installer (No Go Required)

#### macOS & Linux
```bash
chmod +x install.sh
./install.sh
```

#### Windows (PowerShell)
```powershell
.\install.ps1
```

---

### Option 2: Docker & Docker Compose

If you have Docker installed, no Go toolchain is needed:

```bash
# Using Docker
docker build -t bob-gemini-free .
docker run -d --name bob-gemini-free -p 8081:8081 bob-gemini-free

# Using Docker Compose
docker compose up -d
```

---

### Option 3: Build from Source with Make or Go (Go 1.22+)

```bash
# Clone the repository
git clone https://github.com/div197/bob-gemini-free.git
cd bob-gemini-free

# Build with Make
make build

# Start the gateway on port 8081
./bob-gemini-free --port 8081
```

The gateway will start listening at `http://127.0.0.1:8081/v1`.

---

## Client Integration

### OpenAI Python SDK

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://127.0.0.1:8081/v1",
    api_key="none"  # Or your configured api_key
)

# Text Chat Completion with Deep Thinking
response = client.chat.completions.create(
    model="gemini-3.5-flash-thinking",
    messages=[
        {"role": "user", "content": "Explain quantum error correction step by step."}
    ]
)
print(response.choices[0].message.content)
```

### Multimodal Vision (OpenAI Format)

```python
import base64
from openai import OpenAI

client = OpenAI(base_url="http://127.0.0.1:8081/v1", api_key="none")

with open("diagram.png", "rb") as img_file:
    b64_img = base64.b64encode(img_file.read()).decode("utf-8")

response = client.chat.completions.create(
    model="gemini-3.6-flash",
    messages=[
        {
            "role": "user",
            "content": [
                {"type": "text", "text": "What architecture is depicted in this image?"},
                {"type": "image_url", "image_url": {"url": f"data:image/png;base64,{b64_img}"}}
            ]
        }
    ]
)
print(response.choices[0].message.content)
```

### cURL

```bash
curl http://127.0.0.1:8081/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-3.6-flash",
    "messages": [{"role": "user", "content": "Hello BOB Gemini Free!"}]
  }'
```

### Gemini CLI (Native Google API)

```bash
export GEMINI_API_KEY=none
export GOOGLE_GEMINI_BASE_URL=http://127.0.0.1:8081
gemini
```

---

## Model Matrix & Reasoning Controls

| Local Model Alias | Backend Mode | Default Think Depth | Output Profile | Auth Requirement |
| :--- | :---: | :---: | :--- | :--- |
| `gemini-3.6-flash` | Mode 1 | `@think=4` | Fast all-around model (~12k chars) | Free / Anonymous |
| `gemini-3.5-flash` | Mode 1 | `@think=4` | Alias for `gemini-3.6-flash` | Free / Anonymous |
| `gemini-3.5-flash-thinking` | Mode 2 | `@think=0` | **Deep thinking mode** (~20k+ chars) | Free / Anonymous |
| `gemini-3.5-flash-thinking-lite` | Mode 5 | `@think=0` | Adaptive thinking depth (~15k chars) | Free / Anonymous |
| `gemini-flash-lite` | Mode 6 | `@think=4` | Ultra-low latency responses | Free / Anonymous |
| `gemini-auto` | Mode 4 | `@think=4` | Google server-side auto routing | Free / Anonymous |
| `gemini-3.1-pro` | Mode 3 | `@think=4` | Flagship Pro reasoning & code | **Gemini Advanced Cookie** |

### Dynamic Thinking Depth Override

Append `@think=N` to any model alias in your client to control reasoning depth on the fly:

```text
gemini-3.6-flash@think=0    # Deepest step-by-step reasoning tokens
gemini-3.6-flash@think=2    # Balanced medium reasoning
gemini-3.6-flash@think=4    # Direct fast response (shallowest reasoning)
```

---

## Unlocking Pro: Gemini Advanced ($20/mo) Cookies

Anonymous and standard free accounts have full access to Flash, Thinking, and Lite models. If you have an active **Google AI / Gemini Advanced** subscription, configure your session cookie to activate real **Pro** model routing:

1. Open Chrome, go to [gemini.google.com](https://gemini.google.com) and sign in with your subscribed Google account.
2. Open DevTools (`F12`) → **Application** → **Cookies** → `https://gemini.google.com`.
3. Copy your cookie values: `SID`, `HSID`, `SSID`, `APISID`, `SAPISID`, `__Secure-1PSID`.
4. Create a secure local cookie file (e.g. `~/.config/bob-gemini-free/cookie.txt`):

```text
SID=your_sid; HSID=your_hsid; SSID=your_ssid; APISID=your_apisid; SAPISID=your_sapisid; __Secure-1PSID=your_1psid
```

5. Set restrictive POSIX permissions so only your OS user can read it:

```bash
chmod 600 ~/.config/bob-gemini-free/cookie.txt
```

6. Start the server with the cookie flag:

```bash
./bob-gemini-free --cookie-file ~/.config/bob-gemini-free/cookie.txt
```

---

## Configuration (`config.json`)

Create `config.json` or place it in `~/.config/bob-gemini-free/config.json`:

```json
{
  "port": 8081,
  "host": "127.0.0.1",
  "retry_attempts": 3,
  "retry_delay_sec": 2,
  "request_timeout_sec": 180,
  "default_model": "gemini-3.6-flash",
  "api_keys": ["sk-your-secret-key"],
  "cookie_file": null,
  "proxy": null,
  "impersonate": "chrome_133",
  "log_requests": true
}
```

* **`host`**: Defaults to `127.0.0.1` for maximum security.
* **`api_keys`**: When empty `[]`, authentication is disabled. When configured, requests require `Authorization: Bearer <key>`.
* **`impersonate`**: Mimic browser TLS signatures (`chrome_120`, `chrome_131`, `chrome_133`, `chrome_146`, `firefox_147`, `safari_16_0`).
* **`proxy`**: Route requests through an HTTP/HTTPS/SOCKS5 proxy (e.g. `http://127.0.0.1:7890`).

---

## Developer Commands (Makefile)

| Command | Description |
| :--- | :--- |
| `make build` | Compile static binary for host architecture into `./bob-gemini-free` |
| `make run` | Build and start server immediately on port 8081 |
| `make test` | Run complete unit test suite with verbose logging |
| `make test-cover` | Run test suite with line-by-line coverage report |
| `make dist` | Cross-compile standalone binaries for macOS (ARM64/Intel), Linux (AMD64/ARM64), and Windows |
| `make clean` | Remove local binaries and distribution directory |

---

## About ABCsteps

[**ABCsteps**](https://abcsteps.com/) is an online AI engineering school founded by **Divyanshu Singh Chouhan** in Jodhpur, Rajasthan, India.

ABCsteps provides a complete publicly readable foundation for AI engineers and developers, featuring:
* [**20-Lesson AI Engineering Curriculum**](https://abcsteps.com/offerings/)
* [**Practical Engineering Blog & Tutorials**](https://abcsteps.com/blog/)
* [**Curated Reading Paths & Glossary**](https://abcsteps.com/blog/paths/)
* Founder-led mentorship, architecture reviews, and institutional workshops.

Learn more at [https://abcsteps.com/](https://abcsteps.com/).

---

## License

MIT License. Developed with pride by [ABCsteps.com](https://abcsteps.com/).
