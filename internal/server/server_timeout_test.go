package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/div197/bob-gemini-free/internal/config"
)

// TestDeveloperAPIHonoursRequestTimeout is the M1 regression: the Developer API
// client shared a hard-coded 30s http.Client, so request_timeout_sec never
// applied to Developer API calls, while the web path did honour it.
func TestDeveloperAPIHonoursRequestTimeout(t *testing.T) {
	cfg := config.Default()
	cfg.RequestTimeoutSec = 300
	app := New(cfg, "timeout-test")

	// Developer API calls use the configured budget.
	developer, ok := app.GeminiAPI.HTTP.(*http.Client)
	if !ok {
		t.Fatalf("Developer API requester is %T, want *http.Client", app.GeminiAPI.HTTP)
	}
	if want := 300 * time.Second; developer.Timeout != want {
		t.Fatalf("Developer API timeout = %s, want %s", developer.Timeout, want)
	}
	// The shared client (image fetch, Scotty upload, token refresh) keeps 30s.
	if want := 30 * time.Second; app.HTTPClient.Timeout != want {
		t.Fatalf("shared HTTPClient timeout = %s, want %s", app.HTTPClient.Timeout, want)
	}
}
