package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/div197/bob-gemini-free/internal/config"
)

func TestAnthropicRejectsUnsupportedContentBlockBeforeUpstream(t *testing.T) {
	app := New(config.Default(), "test-version")
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(`{
		"model":"claude-code",
		"messages":[{"role":"user","content":[{"type":"document","source":{}}]}]
	}`))
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "unsupported content block type") {
		t.Fatalf("client-visible validation error missing: %s", rec.Body.String())
	}
}
