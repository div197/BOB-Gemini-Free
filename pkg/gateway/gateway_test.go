package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHandlerEmbedded(t *testing.T) {
	handler := NewHandler(
		WithPort(9999),
		WithDefaultModel("gemini-3.7-flash"),
		WithAPIKeys("sk-embed-key"),
		WithHost("127.0.0.1"),
		WithCookieFile("cookie.txt"),
		WithCookiePool("cookies/acc1.txt", "cookies/acc2.txt"),
		WithAuthUser("1"),
		WithImpersonate("chrome"),
		WithProxy("http://127.0.0.1:8080"),
	)

	if handler == nil {
		t.Fatalf("Expected non-nil handler")
	}

	// 1. Test unauthorized request
	req1 := httptest.NewRequest("GET", "/", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 on unauthorized request, got %d", rec1.Code)
	}

	// 2. Test authorized request
	req2 := httptest.NewRequest("GET", "/?key=sk-embed-key", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("Expected 200 on authorized request, got %d", rec2.Code)
	}

	// 3. Test models lookup
	req3 := httptest.NewRequest("GET", "/v1/models?key=sk-embed-key", nil)
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Errorf("Expected 200 on models endpoint, got %d", rec3.Code)
	}
}

func TestNewEngine(t *testing.T) {
	tmpDir := t.TempDir()
	engine := NewEngine(
		WithDefaultModel("gemini-3.7-flash"),
		WithLogRequests(false),
		WithVersion("v1.0.0-test"),
		WithRetry(5, 3),
		WithTimeout(120),
		WithCookiePoolDir(tmpDir),
	)
	if engine == nil {
		t.Fatalf("Expected non-nil Engine")
	}

	h := engine.Handler()
	if h == nil {
		t.Fatalf("Expected non-nil Handler from Engine")
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200 on health check via Engine handler, got %d", rec.Code)
	}

	if engine.app.Cfg.RetryAttempts != 5 {
		t.Errorf("Expected retry attempts 5, got %d", engine.app.Cfg.RetryAttempts)
	}
	if engine.app.Cfg.RequestTimeoutSec != 120 {
		t.Errorf("Expected request timeout 120, got %d", engine.app.Cfg.RequestTimeoutSec)
	}
	if engine.app.Version != "v1.0.0-test" {
		t.Errorf("Expected version v1.0.0-test, got %s", engine.app.Version)
	}
}

func TestEngineGenerateMethods(t *testing.T) {
	engine := NewEngine(
		WithDefaultModel("gemini-3.7-flash"),
		WithVersion("v1.0.0-media-test"),
	)
	if engine == nil {
		t.Fatalf("Expected non-nil Engine")
	}

	// Verify methods return context error when context is canceled
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := engine.Generate(ctx, "hello", "gemini-3.7-flash")
	if err == nil {
		t.Errorf("Expected context error for canceled context in Generate")
	}

	_, err = engine.GenerateWithMedia(ctx, "hello", "gemini-3.7-flash", []string{"ref1"})
	if err == nil {
		t.Errorf("Expected context error for canceled context in GenerateWithMedia")
	}

	err = engine.GenerateStream(ctx, "hello", "gemini-3.7-flash", func(delta string) error {
		return nil
	})
	if err == nil {
		t.Errorf("Expected context error for canceled context in GenerateStream")
	}

	err = engine.GenerateStreamWithMedia(ctx, "hello", "gemini-3.7-flash", []string{"ref1"}, func(delta string) error {
		return nil
	})
	if err == nil {
		t.Errorf("Expected context error for canceled context in GenerateStreamWithMedia")
	}
}
