package diag

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDiagnosticsRunner(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "version": "v0.1.0"})
	})

	mux.HandleFunc("GET /v1/models", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "list",
			"data": []map[string]any{
				{"id": "gemini-3.7-flash", "object": "model"},
			},
		})
	})

	mux.HandleFunc("GET /v1/models/gemini-3.7-flash", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "gemini-3.7-flash", "object": "model"})
	})

	mux.HandleFunc("POST /v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{"content": "OK", "reasoning_content": "7*8=56"},
				},
			},
		})
	})

	mux.HandleFunc("POST /v1beta/models/gemini-3.7-flash:generateContent", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"candidates": []map[string]any{
				{"content": map[string]any{"parts": []map[string]string{{"text": "Ping"}}}},
			},
		})
	})

	mux.HandleFunc("POST /v1/responses", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "completed"})
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	results := RunDiagnostics(ts.URL, "test-key")
	if len(results) != 9 {
		t.Fatalf("expected 9 diagnostic test results, got %d", len(results))
	}

	for _, res := range results {
		if !res.Passed {
			t.Errorf("test %q failed: %v", res.Name, res.Error)
		}
	}
}
