# Embedded Go Library Guide (`pkg/gateway`)

BOB Gemini Free exposes its entire gateway engine as a clean, reusable Go standard library package (`pkg/gateway`). You can embed the gateway directly inside existing Go web services, microservices, or custom proxies without running a separate binary.

---

## 📦 Installation

```bash
go get github.com/div197/bob-gemini-free
```

---

## 🚀 Quick Example

```go
package main

import (
	"log"
	"net/http"

	"github.com/div197/bob-gemini-free/pkg/gateway"
)

func main() {
	// Create embedded handler with functional options
	handler, err := gateway.NewHandler(
		gateway.WithDefaultModel("gemini-3.7-flash"),
		gateway.WithCookieFile("./cookie.txt"),
		gateway.WithLogRequests(true),
		gateway.WithAPIKeys("sk-internal-secret"),
	)
	if err != nil {
		log.Fatalf("Failed to initialize gateway: %v", err)
	}

	// Mount into your existing HTTP server
	http.Handle("/v1/", handler)
	http.Handle("/v1beta/", handler)
	http.Handle("/", handler)

	log.Println("Server running on http://127.0.0.1:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

---

## 🛠️ Available Functional Options

| Option | Type | Description |
| :--- | :--- | :--- |
| `WithConfigFile(path string)` | `string` | Load base settings from `config.json` |
| `WithCookieFile(path string)` | `string` | Specify authenticated session `cookie.txt` |
| `WithAuthUser(index string)` | `string` | Multi-account Google profile index (`"0"`, `"1"`) |
| `WithDefaultModel(model string)` | `string` | Fallback model if omitted by client |
| `WithAPIKeys(keys ...string)` | `...string` | Enforce API key authorization |
| `WithProxy(proxyURL string)` | `string` | Outbound HTTP/SOCKS5 proxy |
| `WithImpersonate(profile string)` | `string` | Browser TLS fingerprint profile |
| `WithLogRequests(enabled bool)` | `bool` | Enable request lifecycle logging |
