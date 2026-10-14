package main

import (
	"net/http"
	"testing"
	"time"
)

// TestGatewayHTTPServerTimeouts is the M2 regression. The gateway had only
// ReadHeaderTimeout, so a client could hold a connection open with a slow
// body or an idle keep-alive indefinitely. WriteTimeout must stay zero, since
// SSE responses are long-lived.
func TestGatewayHTTPServerTimeouts(t *testing.T) {
	srv := newGatewayHTTPServer("127.0.0.1:9610", http.NotFoundHandler())

	if srv.ReadHeaderTimeout != 10*time.Second {
		t.Fatalf("ReadHeaderTimeout = %s, want 10s", srv.ReadHeaderTimeout)
	}
	if srv.ReadTimeout < 5*time.Minute {
		t.Fatalf("ReadTimeout = %s, want at least 5m so a maximum-size body can upload", srv.ReadTimeout)
	}
	if srv.IdleTimeout != 2*time.Minute {
		t.Fatalf("IdleTimeout = %s, want 2m", srv.IdleTimeout)
	}
	if srv.WriteTimeout != 0 {
		t.Fatalf("WriteTimeout = %s, want 0 so streaming responses are not cut", srv.WriteTimeout)
	}
}
