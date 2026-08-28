package mobile

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

type testCallback struct {
	deltas      []string
	completed   bool
	totalTokens int
	errStr      string
}

func (t *testCallback) OnDelta(delta string) {
	t.deltas = append(t.deltas, delta)
}

func (t *testCallback) OnComplete(totalTokens int, errStr string) {
	t.completed = true
	t.totalTokens = totalTokens
	t.errStr = errStr
}

func TestMobileGatewayLifecycle(t *testing.T) {
	gw := GetDefaultGateway()

	if gw.IsRunning() {
		t.Fatal("expected gateway to initially be stopped")
	}

	url, err := gw.Start(0, "127.0.0.1", "")
	if err != nil {
		t.Fatalf("failed to start mobile gateway: %v", err)
	}
	defer func() {
		_ = gw.Stop()
	}()

	if !gw.IsRunning() {
		t.Fatal("expected gateway to be running")
	}

	if !strings.HasPrefix(url, "http://127.0.0.1:") {
		t.Fatalf("unexpected base URL: %s", url)
	}

	if gw.GetURL() != url {
		t.Fatalf("expected GetURL %q, got %q", url, gw.GetURL())
	}

	// Test calling the health endpoint over loopback
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("%s/healthz", url))
	if err != nil {
		t.Fatalf("failed to query healthz on mobile gateway: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", resp.StatusCode)
	}

	// Test CountTokens
	tokens := gw.CountTokens("Hello World! नमस्ते 123")
	if tokens <= 0 {
		t.Fatalf("expected positive token count, got %d", tokens)
	}

	// Test Version
	v := Version()
	if !strings.Contains(v, "mobile") {
		t.Fatalf("expected mobile version string, got %s", v)
	}

	// Test Stop
	err = gw.Stop()
	if err != nil {
		t.Fatalf("failed to stop gateway: %v", err)
	}
	if gw.IsRunning() {
		t.Fatal("expected gateway to be stopped after Stop()")
	}
}
