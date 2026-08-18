# Embedded Go Library Guide (`pkg/gateway`)

BOB Gemini Free exposes its entire gateway engine as a clean, reusable Go standard library package (`pkg/gateway`). You can embed the gateway directly inside existing Go web services, microservices, or custom proxies without running a separate binary.

---

## 📦 Installation

```bash
go get github.com/div197/bob-gemini-free
```

---

## 🚀 Quick Example (Standard `net/http`)

```go
package main

import (
	"log"
	"net/http"

	"github.com/div197/bob-gemini-free/pkg/gateway"
)

func main() {
	// Create embedded handler with functional options
	handler := gateway.NewHandler(
		gateway.WithDefaultModel("gemini-3.7-flash"),
		gateway.WithCookieFile("./cookie.txt"),
		gateway.WithLogRequests(true),
		gateway.WithAPIKeys("sk-internal-secret"),
	)

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
| `WithPort(port int)` | `int` | Default server listening port |
| `WithHost(host string)` | `string` | Local network binding interface (default: `127.0.0.1`) |
| `WithCookieFile(path string)` | `string` | Specify authenticated session `cookie.txt` |
| `WithCookiePool(paths ...string)` | `...string` | Configure round-robin pool of multiple account cookie files |
| `WithAuthUser(index string)` | `string` | Multi-account Google profile index (`"0"`, `"1"`) |
| `WithDefaultModel(model string)` | `string` | Fallback model if omitted by client |
| `WithAPIKeys(keys ...string)` | `...string` | Enforce API key authorization |
| `WithProxy(proxyURL string)` | `string` | Outbound HTTP/SOCKS5 proxy |
| `WithImpersonate(profile string)` | `string` | Browser TLS fingerprint profile (`chrome`, `firefox`, `safari`) |
| `WithLogRequests(enabled bool)` | `bool` | Enable request lifecycle logging |

---

## 🔒 Custom Context Cancellation & Timeouts

The embedded handler respects standard Go `context.Context` cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

req, _ := http.NewRequestWithContext(ctx, "POST", "/v1/chat/completions", body)
handler.ServeHTTP(rec, req)
```
