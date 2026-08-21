package diag

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/server"
)

func TestSingleModelRetrieve(t *testing.T) {
	cfg := config.Default()
	app := server.New(cfg, "v0.1.0")
	handler := app.Handler()

	req := httptest.NewRequest(http.MethodGet, "/v1/models/gemini-3.7-flash-thinking", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for model retrieval, got %d", rec.Code)
	}

	var res map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode model response: %v", err)
	}
	if res["id"] != "gemini-3.7-flash-thinking" {
		t.Errorf("expected id gemini-3.7-flash-thinking, got %v", res["id"])
	}
	modelID, ok := res["id"].(string)
	if !ok || !strings.Contains(modelID, "thinking") {
		t.Errorf("expected thinking model name, got %v", res["id"])
	}
}
