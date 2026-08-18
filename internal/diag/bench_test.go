package diag

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBenchmarkRunner(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "apple banana orange grape pear"}},
			},
			"usage": map[string]any{
				"total_tokens": 25,
			},
		})
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	report := RunBenchmark(ts.URL, "test-key", 2, 4)
	if report.Successful != 4 {
		t.Errorf("Expected 4 successful requests, got %d", report.Successful)
	}
	if report.Failed != 0 {
		t.Errorf("Expected 0 failed requests, got %d", report.Failed)
	}
	if report.TokensPerSecond <= 0 {
		t.Errorf("Expected positive TPS, got %f", report.TokensPerSecond)
	}
}
