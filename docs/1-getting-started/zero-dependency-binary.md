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

Students should use the reviewed local-file installer above. A direct binary
download is not authenticated by its SHA-256 value alone: download
`SHA256SUMS` and `SHA256SUMS.sig` from the same release and verify the complete
directory with the repository's `cmd/release-verify` tool before execution.
Never treat a checksum copied from an unrelated page as authenticity proof.

| OS / Architecture | Target Hardware | Direct Binary Link |
| :--- | :--- | :--- |
| **macOS ARM64** | Apple Silicon (M1 / M2 / M3 / M4 / M5) | `bob-gemini-free-darwin-arm64` |
| **macOS AMD64** | Intel Macs | `bob-gemini-free-darwin-amd64` |
| **Linux AMD64** | Ubuntu, Debian, CentOS, Fedora (x86_64) | `bob-gemini-free-linux-amd64` |
| **Linux ARM64** | Raspberry Pi 4/5, AWS Graviton, Ampere | `bob-gemini-free-linux-arm64` |
| **Windows AMD64** | Windows 10, Windows 11 (64-bit) | `bob-gemini-free-windows-amd64.exe` |

---

## Direct execution (operator-verified release directory only)

The links above are asset names, not proof that an arbitrary downloaded file is
authentic. Before execution, download the matching `SHA256SUMS` and
`SHA256SUMS.sig` into the same directory and run the repository's
`cmd/release-verify` directory verifier with the embedded public key. If that
verification cannot be performed, stop and use the authenticated installer or
do not execute the file. Do not replace this check with a checksum copied from
an unrelated page.

For student distribution, the reviewed `install.sh`/`install.ps1` flow is the
supported path: download the script as a local file, inspect it, and run it.
The installers fail closed when a signed release cannot be verified. They do
not pipe a remote script into a shell, compile the current directory, or accept
an unsigned binary by default.

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
