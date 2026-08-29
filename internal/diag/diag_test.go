package diag

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
		content := "OK"
		if _, wantsJSON := reqBody["response_format"]; wantsJSON {
			content = `{"result":42}`
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{"content": content, "reasoning_content": "7*8=56"},
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
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "resp_test",
			"object":      "response",
			"status":      "completed",
			"output_text": "hello",
			"output":      []any{},
		})
	})

	mux.HandleFunc("POST /v1/messages", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Stream bool `json:"stream"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Stream {
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("event: message_start\ndata: {\"type\":\"message_start\"}\n\n" +
				"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"content_block\":{\"type\":\"text\"}}\n\n" +
				"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\"Claude OK\"}}\n\n" +
				"event: content_block_stop\ndata: {\"type\":\"content_block_stop\"}\n\n" +
				"event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"}}\n\n" +
				"event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n"))
			return
		}
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
	if len(results) != 15 {
		t.Fatalf("expected 15 diagnostic test results, got %d", len(results))
	}

	for _, res := range results {
		if !res.Passed {
			t.Errorf("test %q failed: %v", res.Name, res.Error)
		}
	}
}

func TestDiagnosticsRejectsMalformedTargetURL(t *testing.T) {
	results := RunDiagnostics("http://[::1", "test-key")
	if len(results) != 15 {
		t.Fatalf("expected 15 diagnostic results, got %d", len(results))
	}
	for _, result := range results {
		if result.Passed {
			t.Fatalf("malformed target URL was reported as passed by %q", result.Name)
		}
	}
}

func TestDiagnosticStreamRejectsDoneOnly(t *testing.T) {
	resp := &http.Response{Body: io.NopCloser(strings.NewReader("data: [DONE]\n\n"))}
	if err := scanDiagnosticOpenAIStream(resp); err == nil {
		t.Fatal("expected [DONE]-only stream to fail")
	}
}

func TestDiagnosticStreamRejectsFinishOnly(t *testing.T) {
	resp := &http.Response{Body: io.NopCloser(strings.NewReader("data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n"))}
	if err := scanDiagnosticOpenAIStream(resp); err == nil {
		t.Fatal("expected finish-only stream to fail")
	}
}

func TestDiagnosticStreamRejectsOversizedBody(t *testing.T) {
	body := strings.Repeat(": keepalive\n", int(maxDiagnosticResponseBytes/11)+2)
	resp := &http.Response{Body: io.NopCloser(strings.NewReader(body))}
	if err := scanDiagnosticOpenAIStream(resp); err == nil || !strings.Contains(err.Error(), "safety limit") {
		t.Fatalf("oversized stream error = %v, want safety-limit failure", err)
	}
}

func TestDiagnosticAnthropicStreamRejectsIncompleteLifecycle(t *testing.T) {
	resp := &http.Response{Body: io.NopCloser(strings.NewReader("event: message_start\ndata: {\"type\":\"message_start\"}\n\n"))}
	if err := scanDiagnosticAnthropicStream(resp); err == nil {
		t.Fatal("expected incomplete Anthropic lifecycle to fail")
	}
}

func TestDiagnosticsDoNotPassUnavailableImageGeneration(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer ts.Close()

	results := RunDiagnostics(ts.URL, "test-key")
	for _, result := range results {
		if strings.Contains(result.Name, "Image Generation") {
			if result.Passed {
				t.Fatal("unavailable image generation was reported as passed")
			}
			if result.Error == nil || !strings.Contains(result.Error.Error(), "provider-dependent") {
				t.Fatalf("image failure was not identified as provider-dependent: %v", result.Error)
			}
			return
		}
	}
	t.Fatal("image generation diagnostic was not found")
}
