package diag

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormalizeBenchmarkSettings(t *testing.T) {
	if gotConcurrency, gotRequests := normalizeBenchmarkSettings(0, 0); gotConcurrency != 3 || gotRequests != 6 {
		t.Fatalf("defaults = %d/%d, want 3/6", gotConcurrency, gotRequests)
	}
	if gotConcurrency, gotRequests := normalizeBenchmarkSettings(maxBenchmarkConcurrency+1, maxBenchmarkRequests+1); gotConcurrency != maxBenchmarkConcurrency || gotRequests != maxBenchmarkRequests {
		t.Fatalf("caps = %d/%d, want %d/%d", gotConcurrency, gotRequests, maxBenchmarkConcurrency, maxBenchmarkRequests)
	}
}

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
	if !report.TokenCountsMeasured {
		t.Fatal("provider-reported token counts should be marked measured")
	}
}

func TestBenchmarkDoesNotInventMissingTokenCounts(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "response without usage"}},
			},
		})
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	report := RunBenchmark(ts.URL, "test-key", 1, 2)
	if report.Successful != 2 || report.Failed != 0 {
		t.Fatalf("report = %#v", report)
	}
	if report.TotalTokens != 0 || report.TokensPerSecond != 0 || report.TokenCountsMeasured {
		t.Fatalf("benchmark invented token measurements: %#v", report)
	}
}

func TestBenchmarkRejectsInvalidOrEmptyResponses(t *testing.T) {
	requestCount := 0
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount == 1 {
			_, _ = w.Write([]byte("not-json"))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{}})
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	report := RunBenchmark(ts.URL, "test-key", 1, 2)
	if report.Successful != 0 || report.Failed != 2 {
		t.Fatalf("invalid/empty responses were counted as successful: %#v", report)
	}
}
