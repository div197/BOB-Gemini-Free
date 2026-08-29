# Global School, College & Hackathon Lab Deployment Guide

**BOB Gemini Free** can be evaluated as a local gateway for a classroom or
lab. This guide describes a possible topology; it does not guarantee a fixed
student count, provider capacity, zero operating cost, or unlimited use. Those
depend on the host machine, network, account/session state, and Google's
current policies.

---

## 🏫 Classroom Topology: Single Master Hub Mode

```
                                  ┌──────────────────────────────────────────────┐
                                  │      CLASSROOM MASTER NODE (Mac / Linux / PC)│
                                  │      • IP: 192.168.1.100:9610                │
                                  │      • BOB Gemini Free v0.2.0                │
                                  │      • StreamFlight High-Concurrency Mux     │
                                  │      • Stale-While-Revalidate Session Cache  │
                                  └──────────────────────┬───────────────────────┘
                                                         │
                                   ┌─────────────────────┴─────────────────────┐
                                   │                                           │
                                   ▼                                           ▼
                    ┌──────────────────────────────┐            ┌──────────────────────────────┐
                    │    STUDENT WORKSTATION 1     │            │    STUDENT WORKSTATION 30+   │
                    │    • Cursor / Windsurf       │            │    • Python / Jupyter Lab    │
                    │    • http://192.168.1.100    │            │    • http://192.168.1.100    │
                    └──────────────────────────────┘            └──────────────────────────────┘
```

---

## 🚀 1-Command Master Host Deployment

### Option A: Direct Binary
```bash
./bob-gemini-free --host 0.0.0.0 --port 9610
```

### Option B: Docker / OrbStack
```bash
docker run -d \
  --name bob-classroom-hub \
  --restart always \
  -p 9610:9610 \
  -e BOB_GEMINI_FREE_HOST=0.0.0.0 \
  -e BOB_GEMINI_FREE_PORT=9610 \
  bob-gemini-free:local
```

---

## 💻 Connecting Student Machines

Students on the local Wi-Fi / Ethernet LAN connect their tools directly to the Master Node IP:

When a gateway is bound to a LAN interface, configure `api_keys` and require a
Bearer key for clients. `API key: none` is appropriate only for a trusted
loopback-only development instance; it is not a secure classroom-LAN default.
Never mount or distribute a teacher's cookie to students.

### 1. Browser Web Studio
Students open their browser to:
`http://192.168.1.100:9610/playground`

### 2. VS Code / Cursor / Windsurf
* **OpenAI Base URL**: `http://192.168.1.100:9610/v1`
* **API Key**: the locally configured classroom key
* **Model**: `gemini-3.7-flash`

### 3. Claude Code CLI
```bash
export ANTHROPIC_BASE_URL=http://192.168.1.100:9610
export ANTHROPIC_API_KEY="$BOB_GEMINI_FREE_API_KEYS"
claude
```

### 4. Python SDK in Jupyter Notebooks
```python
from openai import OpenAI

client = OpenAI(
    base_url="http://192.168.1.100:9610/v1",
    api_key="your-local-classroom-key"
)

response = client.chat.completions.create(
    model="gemini-3.7-flash",
    messages=[{"role": "user", "content": "Explain binary search in Python"}]
)
print(response.choices[0].message.content)
```

---

## 🛡️ Why BOB Excels in Classroom Environments

1. **StreamFlight Deduplication**: Identical concurrent requests may share one
   upstream request. The current fixture suite tests multiplexing, but it does
   not certify 50/100 students, zero lag, or unlimited downstream capacity.
2. **Local process boundary**: The gateway has local aggregate metrics and no
   automatic telemetry; RAM and long-running stability still need a dated
   measurement on the target host.
3. **No gateway billing**: BOB does not add a commercial API bill, but Google
   account/session limits, network costs, and provider policy still apply.
