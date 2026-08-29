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
	defer func() {
		if closeable, ok := handler.(gateway.CloseableHandler); ok {
			_ = closeable.Close()
		}
	}()

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
| `WithCookiePoolDir(dir string)` | `string` | Auto-discover all `*.txt` cookie files from a directory |
| `WithAuthUser(index string)` | `string` | Multi-account Google profile index (`"0"`, `"1"`) |
| `WithDefaultModel(model string)` | `string` | Default model when omitted by client; unknown explicit model names fail closed |
| `WithAPIKeys(keys ...string)` | `...string` | Enforce API key authorization |
| `WithProxy(proxyURL string)` | `string` | Outbound HTTP/SOCKS5 proxy |
| `WithImpersonate(profile string)` | `string` | Browser TLS fingerprint profile (`chrome`, `firefox`, `safari`) |
| `WithLogRequests(enabled bool)` | `bool` | Enable request lifecycle logging |
| `WithRetry(attempts, delaySec int)` | `int, int` | Max upstream retries and delay between attempts |
| `WithTimeout(sec int)` | `int` | Per-request upstream timeout in seconds |
| `WithVersion(version string)` | `string` | Explicit version string for telemetry endpoint |

## ⚡ Direct In-Process Programmatic Go Inference (`NewEngine`)

When building autonomous CLI agents, background bots, or microservices, you can execute inference directly in Go without HTTP overhead:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/div197/bob-gemini-free/pkg/gateway"
)

func main() {
	// Instantiate the in-process inference engine
	engine := gateway.NewEngine(
		gateway.WithDefaultModel("gemini-3.7-flash"),
		gateway.WithCookieFile("./cookie.txt"),
	)
	defer engine.Close()

	ctx := context.Background()

	// 1. Direct synchronous generation
	response, err := engine.Generate(ctx, "Explain quantum error correction briefly.", "gemini-3.7-flash")
	if err != nil {
		log.Fatalf("Inference error: %v", err)
	}
	fmt.Println("Response:\n", response)

	// 2. Direct real-time streaming in Go
	fmt.Println("\nStreaming:")
	err = engine.GenerateStream(ctx, "Write a quicksort in Go.", "gemini-3.7-flash", func(delta string) error {
		fmt.Print(delta)
		return nil
	})
	if err != nil {
		log.Fatalf("Stream error: %v", err)
	}

	// 3. Multimodal vision with image attachments directly in Go
	imageRefs := []string{"/test/remy/skills/dropzone/diagram.png"}
	analysis, err := engine.GenerateWithMedia(ctx, "Analyze this architecture diagram.", "gemini-3.7-flash", imageRefs)
	if err != nil {
		log.Fatalf("Vision error: %v", err)
	}
	fmt.Println("\nVision Analysis:\n", analysis)
}
```

---

## 🔒 Custom Context Cancellation & Timeouts

The embedded engine and handler respect standard Go `context.Context` cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

req, _ := http.NewRequestWithContext(ctx, "POST", "/v1/chat/completions", body)
handler.ServeHTTP(rec, req)
```

`NewEngine` owns the gateway's session-reload worker and should be closed when
the embedding process shuts down. `NewHandler` returns an ordinary
`http.Handler` that also implements `gateway.CloseableHandler`; close that
value when cookie files or a cookie pool are configured. Guest-only engines do
not start an idle session-reload worker.
