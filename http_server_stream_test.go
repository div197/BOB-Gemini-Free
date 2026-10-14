package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestGatewayStreamOutlivesReadTimeout checks that a streamed answer keeps its
// request context for longer than ReadTimeout. The gateway reads the JSON body
// first and then streams, which is the path chat and messages use. If Go
// cancelled the context once ReadTimeout passed, long streams would be cut.
func TestGatewayStreamOutlivesReadTimeout(t *testing.T) {
	const readTimeout = 300 * time.Millisecond
	const ticks = 12
	cancelled := make(chan struct{}, 1)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		flusher := w.(http.Flusher)
		for i := 0; i < ticks; i++ {
			select {
			case <-r.Context().Done():
				cancelled <- struct{}{}
				return
			case <-time.After(100 * time.Millisecond):
				_, _ = io.WriteString(w, "data: tick\n\n")
				flusher.Flush()
			}
		}
	})

	srv := newGatewayHTTPServer("", handler)
	srv.ReadTimeout = readTimeout
	ts := httptest.NewUnstartedServer(srv.Handler)
	ts.Config = srv
	ts.Start()
	defer ts.Close()

	start := time.Now()
	resp, err := http.Post(ts.URL, "application/json", strings.NewReader(`{"stream":true}`))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	elapsed := time.Since(start)

	if elapsed <= readTimeout {
		t.Fatalf("stream finished in %v, too short to prove it outlived ReadTimeout", elapsed)
	}
	select {
	case <-cancelled:
		t.Fatalf("request context cancelled mid-stream after %v", elapsed)
	default:
	}
	if got := strings.Count(string(body), "data: tick"); got != ticks {
		t.Fatalf("received %d ticks, want %d", got, ticks)
	}
}
