# Standalone Binaries (No Separately Managed Runtime)

BOB Gemini Free can be distributed as a single native executable (`CGO_ENABLED=0`)
that does not require users to separately install Go, Python, Node.js, SQLite,
or a memory/database service. The binary still contains the Go module
dependencies declared in `go.mod`.

It requires:
- ❌ **No Go runtime**
- ❌ **No Python interpreter**
- ❌ **No Node.js or npm**
- ❌ **No dynamic C libraries**

---

## Direct Binary Downloads

| OS / Architecture | Target Hardware | Direct Binary Link |
| :--- | :--- | :--- |
| **macOS ARM64** | Apple Silicon (M1 / M2 / M3 / M4 / M5) | `bob-gemini-free-darwin-arm64` |
| **macOS AMD64** | Intel Macs | `bob-gemini-free-darwin-amd64` |
| **Linux AMD64** | Ubuntu, Debian, CentOS, Fedora (x86_64) | `bob-gemini-free-linux-amd64` |
| **Linux ARM64** | Raspberry Pi 4/5, AWS Graviton, Ampere | `bob-gemini-free-linux-arm64` |
| **Windows AMD64** | Windows 10, Windows 11 (64-bit) | `bob-gemini-free-windows-amd64.exe` |

---

## ⚡ 10-Second Quick Execution

### macOS & Linux
```bash
# 1. Download binary (e.g. macOS Apple Silicon)
curl -fsSL -o bob-gemini-free https://github.com/div197/bob-gemini-free/releases/latest/download/bob-gemini-free-darwin-arm64

# 2. Grant executable permission
chmod +x bob-gemini-free

# 3. Launch
./bob-gemini-free
```

### Windows (PowerShell / CMD)
```powershell
# 1. Download
Invoke-WebRequest -Uri "https://github.com/div197/bob-gemini-free/releases/latest/download/bob-gemini-free-windows-amd64.exe" -OutFile "bob-gemini-free.exe"

# 2. Launch
.\bob-gemini-free.exe
```

---

## Cross-Compiling from Source (`make dist`)

If you have Go installed on your developer workstation, you can build all binaries at once:

```bash
make dist
```

Outputs will be generated in `./dist/`:
```
dist/
├── bob-gemini-free-darwin-arm64
├── bob-gemini-free-darwin-amd64
├── bob-gemini-free-linux-amd64
├── bob-gemini-free-linux-arm64
└── bob-gemini-free-windows-amd64.exe
```

---

## 🔄 Self-Updating Standalone Binaries (`--update`)

Once downloaded, you can upgrade your standalone binary in place at any time:

```bash
./bob-gemini-free --update
```

---

## ⚙️ Running as a 24/7 OS Service Daemon (`service`)

Run BOB Gemini Free silently in the background across system reboots:

```bash
# Register & start background daemon
./bob-gemini-free service install

# Check background daemon health
./bob-gemini-free service status

# Stop / Uninstall daemon
./bob-gemini-free service stop
./bob-gemini-free service uninstall
```
