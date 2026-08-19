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
		var reqBody map[string]any
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
		if isStream, _ := reqBody["stream"].(bool); isStream {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"1, 2\"}}]}\n\ndata: [DONE]\n\n"))
			return
		}
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

	mux.HandleFunc("POST /v1/messages", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"role": "assistant",
			"content": []map[string]any{
				{"type": "text", "text": "Claude OK"},
			},
		})
	})

	mux.HandleFunc("POST /v1/images/generations", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"created": 1700000000,
			"data": []map[string]any{
				{"url": "https://lh3.googleusercontent.com/test.png"},
			},
		})
	})

	mux.HandleFunc("POST /v1beta/models/gemini-3.7-flash:countTokens", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"totalTokens": 12,
		})
	})

	mux.HandleFunc("POST /v1/tokens/count", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model":         "gemini-3.7-flash",
			"prompt_tokens": 10,
			"total_tokens":  10,
		})
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	results := RunDiagnostics(ts.URL, "test-key")
	if len(results) != 13 {
		t.Fatalf("expected 13 diagnostic test results, got %d", len(results))
	}

	for _, res := range results {
		if !res.Passed {
			t.Errorf("test %q failed: %v", res.Name, res.Error)
		}
	}
}
