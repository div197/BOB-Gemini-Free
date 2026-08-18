package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHandlerEmbedded(t *testing.T) {
	handler := NewHandler(
		WithPort(9999),
		WithDefaultModel("gemini-3.7-flash"),
		WithAPIKeys("sk-embed-key"),
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
