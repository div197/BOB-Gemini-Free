# Global School, College & Hackathon Lab Deployment Guide

**BOB Gemini Free** enables schools, universities, coding bootcamps, and hackathons to deploy a **₹0 / $0 cost, high-concurrency local AI hub** across 100+ classroom workstations using a single master node.

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

### Option B: Docker / OrbStack (10.8 MB Container)
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

### 1. Browser Web Studio
Students open their browser to:
`http://192.168.1.100:9610/playground`

### 2. VS Code / Cursor / Windsurf
* **OpenAI Base URL**: `http://192.168.1.100:9610/v1`
* **API Key**: `none`
* **Model**: `gemini-3.7-flash`

### 3. Claude Code CLI
```bash
export ANTHROPIC_BASE_URL=http://192.168.1.100:9610
export ANTHROPIC_API_KEY=none
claude
```

### 4. Python SDK in Jupyter Notebooks
```python
from openai import OpenAI

client = OpenAI(
    base_url="http://192.168.1.100:9610/v1",
    api_key="none"
)

response = client.chat.completions.create(
    model="gemini-3.7-flash",
    messages=[{"role": "user", "content": "Explain binary search in Python"}]
)
print(response.choices[0].message.content)
```

---

## 🛡️ Why BOB Excels in Classroom Environments

1. **StreamFlight Deduplication**: If 50 students submit the same programming assignment question simultaneously, BOB executes **1 single upstream request** and broadcasts real-time SSE chunks to all 50 students in parallel without lag.
2. **Zero Memory Leaks**: Runs continuously 24/7 in < 25 MB RAM.
3. **No Financial Risk**: Completely eliminates accidental API key leaks or student overdraft charges on commercial cloud platforms.
