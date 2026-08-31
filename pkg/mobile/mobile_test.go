package mobile

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
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

func TestMobileGatewayRejectsNonLoopbackBinding(t *testing.T) {
	gw := &MobileGateway{}
	if _, err := gw.Start(0, "0.0.0.0", ""); err == nil || !strings.Contains(err.Error(), "must be loopback") {
		t.Fatalf("non-loopback host was accepted: %v", err)
	}
	if gw.IsRunning() {
		t.Fatal("mobile gateway became running after rejecting a non-loopback host")
	}
}

func TestNilMobileGatewayMethodsFailClosed(t *testing.T) {
	var gw *MobileGateway

	if _, err := gw.Start(0, "127.0.0.1", ""); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("nil Start() error = %v, want initialization error", err)
	}
	if err := gw.Stop(); err != nil {
		t.Fatalf("nil Stop() = %v, want nil", err)
	}
	if gw.IsRunning() {
		t.Fatal("nil IsRunning() = true, want false")
	}
	if got := gw.GetURL(); got != "" {
		t.Fatalf("nil GetURL() = %q, want empty", got)
	}
	if _, err := gw.Generate("hello", "gemini-3.7-flash"); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("nil Generate() error = %v, want initialization error", err)
	}
	if err := gw.GenerateStream("hello", "gemini-3.7-flash", nil); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("nil GenerateStream() error = %v, want initialization error", err)
	}
	if _, err := gw.Refine("hello"); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("nil Refine() error = %v, want initialization error", err)
	}
}

func TestMobileGatewayUsesEphemeralCookieFileAndCleansItUp(t *testing.T) {
	gw := &MobileGateway{}
	if _, err := gw.Start(0, "localhost", "SID=test-session; SAPISID=test-sapisid"); err != nil {
		t.Fatalf("Start() with cookie content: %v", err)
	}
	cookiePath := gw.cookiePath
	if cookiePath == "" || strings.HasSuffix(cookiePath, "bob_mobile_cookie.txt") {
		t.Fatalf("cookie path is predictable: %q", cookiePath)
	}
	info, err := os.Stat(cookiePath)
	if err != nil {
		t.Fatalf("stat temporary cookie: %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("temporary cookie permissions = %o, want 600", got)
	}
	if err := gw.Stop(); err != nil {
		t.Fatalf("Stop(): %v", err)
	}
	if _, err := os.Stat(cookiePath); !os.IsNotExist(err) {
		t.Fatalf("temporary cookie still exists after Stop(): %v", err)
	}
}

func TestMobileGatewayDoesNotSilentlySubstituteInvalidModel(t *testing.T) {
	gw := &MobileGateway{}
	if _, err := gw.Start(0, "127.0.0.1", ""); err != nil {
		t.Fatalf("Start(): %v", err)
	}
	defer func() { _ = gw.Stop() }()

	if _, err := gw.Generate("hello", "not-a-real-model"); err == nil {
		t.Fatal("Generate() silently substituted a model for invalid input")
	}
	if err := gw.GenerateStream("hello", "not-a-real-model", nil); err == nil {
		t.Fatal("GenerateStream() silently substituted a model for invalid input")
	}

	if err := gw.Stop(); err != nil {
		t.Fatalf("Stop(): %v", err)
	}
	if _, err := gw.Generate("hello", "gemini-3.7-flash"); err == nil || !strings.Contains(err.Error(), "not running") {
		t.Fatalf("Generate() after Stop() = %v, want stopped error", err)
	}
}

func TestMobileGatewayStopCancelsOwnedRunContext(t *testing.T) {
	gw := &MobileGateway{}
	if _, err := gw.Start(0, "127.0.0.1", ""); err != nil {
		t.Fatalf("Start(): %v", err)
	}
	runCtx := gw.runCtx
	if runCtx == nil {
		t.Fatal("Start() did not create an owned run context")
	}
	if err := gw.Stop(); err != nil {
		t.Fatalf("Stop(): %v", err)
	}
	if !errors.Is(runCtx.Err(), context.Canceled) {
		t.Fatalf("owned run context error = %v, want context.Canceled", runCtx.Err())
	}
}

func TestResolveMobileModelUsesDefaultOnlyWhenOmitted(t *testing.T) {
	if _, err := resolveMobileModel(""); err != nil {
		t.Fatalf("omitted model should use default: %v", err)
	}
	if _, err := resolveMobileModel("not-a-real-model"); err == nil {
		t.Fatal("unknown model was accepted by the strict mobile resolver")
	}
}
