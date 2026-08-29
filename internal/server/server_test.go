package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/format"
	"github.com/div197/bob-gemini-free/internal/geminiapi"
)

type fakeGeminiRequester struct {
	body string
}

func (f fakeGeminiRequester) Do(req *http.Request) (*http.Response, error) {
	if req.Method == http.MethodGet {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`<html></html>`)),
		}, nil
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(f.body)),
	}, nil
}

type streamFailureRequester struct{}

func (streamFailureRequester) Do(req *http.Request) (*http.Response, error) {
	if req.Method == http.MethodGet {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("<html></html>")),
		}, nil
	}
	return nil, errors.New("connection reset after request write")
}

func mockGeminiBody(text string) string {
	padded := text + " " + strings.Repeat("x", 220)
	inner := []any{nil, nil, nil, nil, []any{[]any{nil, []any{padded}}}}
	innerBytes, _ := json.Marshal(inner)
	outer := []any{[]any{"wrb.fr", nil, string(innerBytes)}}
	outerBytes, _ := json.Marshal(outer)
	return string(outerBytes) + "\n"
}

func TestReadRequestBodyHasIndependentLimit(t *testing.T) {
	oversized := strings.NewReader(strings.Repeat("x", maxRequestBodySize+1))
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", oversized)

	_, err := readRequestBody(req)
	if !errors.Is(err, errRequestBodyTooLarge) {
		t.Fatalf("readRequestBody error = %v, want %v", err, errRequestBodyTooLarge)
	}
}

func TestHealthEndpoint(t *testing.T) {
	testVer := "test-version-1.0"
	cfg := config.Default()
	app := New(cfg, testVer)
	handler := app.Handler()

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("Invalid JSON response: %v", err)
	}

	if body["status"] != "ok" || body["version"] != testVer {
		t.Errorf("Unexpected health response: %v", body)
	}
	if _, ok := body["estimated_savings_usd"]; !ok {
		t.Errorf("Expected estimated_savings_usd in health response")
	}
}

func TestAuthMatrix(t *testing.T) {
	cfg := config.Default()
	cfg.APIKeys = []string{"sk-secret-key"}
	app := New(cfg, "test-version")
	handler := app.Handler()

	// 1. No key -> 401
	req1 := httptest.NewRequest("GET", "/v1/models", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 without auth, got %d", rec1.Code)
	}

	// 2. Wrong key -> 401
	req2 := httptest.NewRequest("GET", "/v1/models", nil)
	req2.Header.Set("Authorization", "Bearer wrong-key")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 with wrong key, got %d", rec2.Code)
	}

	// 3. Valid Bearer token -> 200
	req3 := httptest.NewRequest("GET", "/v1/models", nil)
	req3.Header.Set("Authorization", "Bearer sk-secret-key")
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Errorf("Expected 200 with Bearer token, got %d", rec3.Code)
	}

	// 4. Valid x-api-key header -> 200
	req4 := httptest.NewRequest("GET", "/v1/models", nil)
	req4.Header.Set("x-api-key", "sk-secret-key")
	rec4 := httptest.NewRecorder()
	handler.ServeHTTP(rec4, req4)
	if rec4.Code != http.StatusOK {
		t.Errorf("Expected 200 with x-api-key, got %d", rec4.Code)
	}

	// 5. Valid x-goog-api-key header -> 200
	req5 := httptest.NewRequest("GET", "/v1/models", nil)
	req5.Header.Set("x-goog-api-key", "sk-secret-key")
	rec5 := httptest.NewRecorder()
	handler.ServeHTTP(rec5, req5)
	if rec5.Code != http.StatusOK {
		t.Errorf("Expected 200 with x-goog-api-key, got %d", rec5.Code)
	}

	// 6. Valid query param ?key= -> 200
	req6 := httptest.NewRequest("GET", "/v1/models?key=sk-secret-key", nil)
	rec6 := httptest.NewRecorder()
	handler.ServeHTTP(rec6, req6)
	if rec6.Code != http.StatusOK {
		t.Errorf("Expected 200 with ?key= query, got %d", rec6.Code)
	}
}

func TestCORSPreflight(t *testing.T) {
	cfg := config.Default()
	app := New(cfg, "test-version")
	handler := app.Handler()

	req := httptest.NewRequest("OPTIONS", "/v1/chat/completions", nil)
	req.Header.Set("Access-Control-Request-Headers", "content-type, x-bob-gemini-api-key")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("Expected status 204 for OPTIONS, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected CORS origin *, got %s", rec.Header().Get("Access-Control-Allow-Origin"))
	}
	if !strings.Contains(strings.ToLower(rec.Header().Get("Access-Control-Allow-Headers")), "x-bob-gemini-api-key") {
		t.Errorf("Expected CORS headers to allow x-bob-gemini-api-key, got %s", rec.Header().Get("Access-Control-Allow-Headers"))
	}
}

func TestDefaultHost(t *testing.T) {
	cfg := config.Default()
	if cfg.Host != "127.0.0.1" {
		t.Errorf("Expected default host 127.0.0.1, got %s", cfg.Host)
	}
}

func TestHealthAuthCheck(t *testing.T) {
	cfg := config.Default()
	cfg.APIKeys = []string{"sk-secret"}
	app := New(cfg, "test-version")
	handler := app.Handler()

	// Without key -> 401
	req1 := httptest.NewRequest("GET", "/", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 on health check with API keys configured, got %d", rec1.Code)
	}

	// With key -> 200
	req2 := httptest.NewRequest("GET", "/?key=sk-secret", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("Expected 200 on health check with valid API key, got %d", rec2.Code)
	}
}

func TestSingleModelEndpoint(t *testing.T) {
	cfg := config.Default()
	app := New(cfg, "test-version")
	handler := app.Handler()

	// 1. Valid model -> 200
	req1 := httptest.NewRequest("GET", "/v1/models/gemini-3.7-flash", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("Expected 200 for valid model, got %d", rec1.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(rec1.Body.Bytes(), &res); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	if res["id"] != "gemini-3.7-flash" || res["object"] != "model" {
		t.Errorf("Unexpected model payload: %v", res)
	}

	// 2. Unknown model -> 404
	req2 := httptest.NewRequest("GET", "/v1/models/non-existent-model", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for unknown model, got %d", rec2.Code)
	}
}

func TestGoogleModelsEndpoint(t *testing.T) {
	cfg := config.Default()
	app := New(cfg, "test-version")
	handler := app.Handler()

	req := httptest.NewRequest("GET", "/v1beta/models", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 for Google models, got %d", rec.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	modelsList, ok := res["models"].([]any)
	if !ok || len(modelsList) == 0 {
		t.Errorf("Expected non-empty models array in Google models response")
	}
}

func TestBadRequestHandling(t *testing.T) {
	cfg := config.Default()
	app := New(cfg, "test-version")
	handler := app.Handler()

	// 1. Invalid JSON in chat completions -> 400
	req1 := httptest.NewRequest("POST", "/v1/chat/completions", strings.NewReader("invalid-json"))
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid JSON in chat completions, got %d", rec1.Code)
	}

	// 2. Empty prompt in chat completions -> 400
	req2 := httptest.NewRequest("POST", "/v1/chat/completions", strings.NewReader(`{"model":"gemini-3.7-flash","messages":[]}`))
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for empty prompt in chat completions, got %d", rec2.Code)
	}

	// 3. Invalid JSON in responses API -> 400
	req3 := httptest.NewRequest("POST", "/v1/responses", strings.NewReader("invalid-json"))
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid JSON in responses API, got %d", rec3.Code)
	}

	// 4. Invalid JSON in Google generateContent -> 400
	req4 := httptest.NewRequest("POST", "/v1beta/models/gemini-3.7-flash:generateContent", strings.NewReader("invalid-json"))
	rec4 := httptest.NewRecorder()
	handler.ServeHTTP(rec4, req4)
	if rec4.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid JSON in Google generateContent, got %d", rec4.Code)
	}

	// 5. Invalid JSON in Anthropic Messages -> 400
	req5 := httptest.NewRequest("POST", "/v1/messages", strings.NewReader("invalid-json"))
	rec5 := httptest.NewRecorder()
	handler.ServeHTTP(rec5, req5)
	if rec5.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid JSON in Anthropic messages, got %d", rec5.Code)
	}

	// 6. Invalid JSON in Image Generations -> 400
	req6 := httptest.NewRequest("POST", "/v1/images/generations", strings.NewReader("invalid-json"))
	rec6 := httptest.NewRecorder()
	handler.ServeHTTP(rec6, req6)
	if rec6.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid JSON in image generations, got %d", rec6.Code)
	}

	// 7. Empty prompt in Image Generations -> 400
	req7 := httptest.NewRequest("POST", "/v1/images/generations", strings.NewReader(`{"prompt":""}`))
	rec7 := httptest.NewRecorder()
	handler.ServeHTTP(rec7, req7)
	if rec7.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for empty prompt in image generations, got %d", rec7.Code)
	}
}

func TestChatReportsFormatValidationErrors(t *testing.T) {
	cfg := config.Default()
	app := New(cfg, "test-version")
	req := httptest.NewRequest("POST", "/v1/chat/completions", strings.NewReader(`{"model":"gemini-3.7-flash","messages":[{"role":"user","content":"use tools"}],"tools":[{"type":"function","function":{"name":"large","description":"`+strings.Repeat("x", format.MaxToolDescriptionBytes+1)+`"}}]}`))
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "empty prompt") || !strings.Contains(rec.Body.String(), "description exceeds") {
		t.Fatalf("format validation error was hidden: %s", rec.Body.String())
	}
}

func TestGoogleReportsFormatValidationErrors(t *testing.T) {
	cfg := config.Default()
	app := New(cfg, "test-version")
	payload := `{"contents":[{"role":"user","parts":[{"text":"use tools"}]}],"tools":[{"functionDeclarations":[{"name":"large","description":"` + strings.Repeat("x", format.MaxToolDescriptionBytes+1) + `"}]}]}`
	req := httptest.NewRequest("POST", "/v1beta/models/gemini-3.7-flash:generateContent", strings.NewReader(payload))
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "empty content") || !strings.Contains(rec.Body.String(), "description exceeds") {
		t.Fatalf("format validation error was hidden: %s", rec.Body.String())
	}
}

func TestResponsesReportsFormatValidationErrors(t *testing.T) {
	cfg := config.Default()
	app := New(cfg, "test-version")
	payload := `{"model":"gpt-5.6-sol","input":[{"role":"user","content":"use tools"}],"tools":[{"type":"function","function":{"name":"large","description":"` + strings.Repeat("x", format.MaxToolDescriptionBytes+1) + `"}}]}`
	req := httptest.NewRequest("POST", "/v1/responses", strings.NewReader(payload))
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "empty input") || !strings.Contains(rec.Body.String(), "description exceeds") {
		t.Fatalf("format validation error was hidden: %s", rec.Body.String())
	}
}

func TestUploadImagesRejectsUnsupportedRemoteURL(t *testing.T) {
	cfg := config.Default()
	app := New(cfg, "test-version")

	_, err := app.uploadImages([]format.Image{{URL: "file:///etc/passwd"}})
	if err == nil {
		t.Fatalf("expected unsupported image URL error")
	}
	if !strings.Contains(err.Error(), "unsupported image URL scheme") {
		t.Fatalf("expected unsupported image URL scheme error, got %v", err)
	}
}

func TestResponsesRejectsUnsupportedInputImageURL(t *testing.T) {
	cfg := config.Default()
	app := New(cfg, "test-version")
	handler := app.Handler()

	body := `{
		"model": "gemini-3.7-flash",
		"input": [{
			"role": "user",
			"content": [
				{"type": "input_text", "text": "Describe this."},
				{"type": "input_image", "image_url": "file:///etc/passwd"}
			]
		}]
	}`
	req := httptest.NewRequest("POST", "/v1/responses", strings.NewReader(body))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502 for unsupported input image URL, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "unsupported image URL scheme") {
		t.Fatalf("expected unsupported image URL scheme error, got %s", rec.Body.String())
	}
}

func TestImageGenerationNoExtractedURLReturnsError(t *testing.T) {
	cfg := config.Default()
	cfg.RetryAttempts = 1
	app := New(cfg, "test-version")
	app.Gem.HTTP = fakeGeminiRequester{body: mockGeminiBody("No image was generated.")}
	handler := app.Handler()

	req := httptest.NewRequest("POST", "/v1/images/generations", strings.NewReader(`{"prompt":"draw a lotus","model":"imagen-3"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502 when upstream returns no image URL, got %d body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "placeholder.png") {
		t.Fatalf("image generation must not return placeholder URL as success: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "upstream response did not contain a generated image URL") {
		t.Fatalf("expected explicit missing-image error, got %s", rec.Body.String())
	}
}

func TestImageGenerationB64DoesNotFallbackToURL(t *testing.T) {
	cfg := config.Default()
	cfg.RetryAttempts = 1
	app := New(cfg, "test-version")
	app.Gem.HTTP = fakeGeminiRequester{body: mockGeminiBody("![generated](https://127.0.0.1/generated.png)")}
	req := httptest.NewRequest("POST", "/v1/images/generations", strings.NewReader(`{"prompt":"draw a lotus","response_format":"b64_json"}`))
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "generated.png") {
		t.Fatalf("b64_json failure returned the source URL: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "image download failed") {
		t.Fatalf("missing explicit image download failure: %s", rec.Body.String())
	}
}

func TestImageGenerationRejectsUnsupportedFormatAndModel(t *testing.T) {
	app := New(config.Default(), "test-version")
	for name, payload := range map[string]string{
		"format": `{"prompt":"draw a lotus","response_format":"binary"}`,
		"model":  `{"prompt":"draw a lotus","model":"not-a-gemini-model"}`,
	} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/v1/images/generations", strings.NewReader(payload))
			rec := httptest.NewRecorder()
			app.Handler().ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "invalid_request_error") {
				t.Fatalf("status/body = %d %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestObservabilityHeaders(t *testing.T) {
	cfg := config.Default()
	app := New(cfg, "test-version")
	handler := app.Handler()

	req := httptest.NewRequest("GET", "/v1/models", nil)
	req.Header.Set("X-Client-Request-Id", "custom-trace-123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("x-request-id") != "custom-trace-123" {
		t.Errorf("Expected x-request-id custom-trace-123, got %s", rec.Header().Get("x-request-id"))
	}
	if rec.Header().Get("openai-version") != "2020-10-01" {
		t.Errorf("Expected openai-version 2020-10-01, got %s", rec.Header().Get("openai-version"))
	}
	if rec.Header().Get("x-ratelimit-limit-requests") == "" {
		t.Errorf("Missing x-ratelimit-limit-requests header")
	}
	if rec.Header().Get("openai-processing-ms") == "" {
		t.Errorf("Missing openai-processing-ms header")
	}
}

func TestTokenCountingEndpoints(t *testing.T) {
	cfg := config.Default()
	app := New(cfg, "test-version")
	handler := app.Handler()

	// 1. Google Native :countTokens
	googlePayload := `{"contents":[{"role":"user","parts":[{"text":"Explain the theory of relativity in simple words."}]}]}`
	req1 := httptest.NewRequest("POST", "/v1beta/models/gemini-3.7-flash:countTokens", strings.NewReader(googlePayload))
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("Expected 200 for Google countTokens, got %d (body: %s)", rec1.Code, rec1.Body.String())
	}

	var gResp map[string]any
	if err := json.Unmarshal(rec1.Body.Bytes(), &gResp); err != nil {
		t.Fatalf("Failed to parse Google countTokens JSON: %v", err)
	}
	totalTokens, ok := gResp["totalTokens"].(float64)
	if !ok || totalTokens < 5 {
		t.Errorf("Expected valid totalTokens >= 5, got %v", gResp["totalTokens"])
	}

	// 2. OpenAI /v1/tokens/count
	openAIPayload := `{"model":"gemini-3.7-flash","messages":[{"role":"user","content":"How does quantum teleportation work?"}]}`
	req2 := httptest.NewRequest("POST", "/v1/tokens/count", strings.NewReader(openAIPayload))
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("Expected 200 for OpenAI /v1/tokens/count, got %d", rec2.Code)
	}

	var oResp map[string]any
	if err := json.Unmarshal(rec2.Body.Bytes(), &oResp); err != nil {
		t.Fatalf("Failed to parse OpenAI tokens count JSON: %v", err)
	}
	promptTokens, ok := oResp["prompt_tokens"].(float64)
	if !ok || promptTokens < 5 {
		t.Errorf("Expected valid prompt_tokens >= 5, got %v", oResp["prompt_tokens"])
	}
}

func TestPlaygroundEndpoint(t *testing.T) {
	cfg := config.Default()
	app := New(cfg, "test-version")
	handler := app.Handler()

	for _, path := range []string{"/playground", "/ui", "/favicon.ico", "/manifest.json", "/sw.js"} {
		req := httptest.NewRequest("GET", path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected 200 for %s, got %d", path, rec.Code)
		}
		if path != "/favicon.ico" && path != "/manifest.json" && path != "/sw.js" && !strings.Contains(rec.Body.String(), "BOB GEMINI FREE") {
			t.Errorf("Expected playground HTML content for %s", path)
		}
		if path != "/favicon.ico" && path != "/manifest.json" && path != "/sw.js" && (strings.Contains(rec.Body.String(), "sql.js") || strings.Contains(rec.Body.String(), "SQLite WASM")) {
			t.Errorf("Removed SQLite WASM studio still advertised by %s", path)
		}
		if path != "/favicon.ico" && path != "/manifest.json" && path != "/sw.js" && !strings.Contains(rec.Body.String(), "test-version") {
			t.Errorf("expected served playground %s to inject the application version", path)
		}
		if path != "/favicon.ico" && path != "/manifest.json" && path != "/sw.js" && strings.Contains(rec.Body.String(), "__BOB_DESKTOP_VERSION__") {
			t.Errorf("served playground %s leaked its version placeholder", path)
		}
		if path == "/favicon.ico" && len(rec.Body.Bytes()) == 0 {
			t.Error("favicon response was empty")
		}
		if path == "/manifest.json" {
			if got := rec.Header().Get("Content-Type"); got != "application/manifest+json; charset=utf-8" {
				t.Errorf("manifest Content-Type = %q", got)
			}
			if !strings.Contains(rec.Body.String(), `"start_url": "/playground"`) {
				t.Error("local manifest must launch the local playground")
			}
		}
		if path == "/sw.js" {
			if got := rec.Header().Get("Content-Type"); got != "application/javascript; charset=utf-8" {
				t.Errorf("service worker Content-Type = %q", got)
			}
			if !strings.Contains(rec.Body.String(), `"bob-gemini-local-studio-" + "test-version"`) {
				t.Error("service worker must contain the served application version")
			}
			if strings.Contains(rec.Body.String(), "__BOB_CACHE_VERSION__") {
				t.Error("service worker leaked its cache version placeholder")
			}
		}
		if (path == "/playground" || path == "/ui") && rec.Header().Get("Content-Type") != "text/html; charset=utf-8" {
			t.Errorf("Expected text/html Content-Type for %s, got %s", path, rec.Header().Get("Content-Type"))
		}
	}
}

func TestStreamWithKeepAlive(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := t.Context()

	// Slow upstream generation that takes 80ms while keepalive interval is 20ms
	runStream := func(emit func(string) error) error {
		time.Sleep(50 * time.Millisecond)
		if err := emit("chunk1"); err != nil {
			return err
		}
		time.Sleep(50 * time.Millisecond)
		return emit("chunk2")
	}

	var emitted []string
	err := StreamWithKeepAlive(ctx, rec, 20*time.Millisecond, runStream, func(delta string) error {
		emitted = append(emitted, delta)
		return nil
	})

	if err != nil {
		t.Fatalf("StreamWithKeepAlive failed: %v", err)
	}

	if len(emitted) != 2 || emitted[0] != "chunk1" || emitted[1] != "chunk2" {
		t.Errorf("unexpected emitted chunks: %#v", emitted)
	}

	body := rec.Body.String()
	if !strings.Contains(body, ": keepalive") {
		t.Errorf("expected keepalive comment in stream output, got %q", body)
	}
}

func TestStreamWithKeepAliveNormalizesNilContext(t *testing.T) {
	rec := httptest.NewRecorder()
	err := StreamWithKeepAlive(nil, rec, time.Hour, func(emit func(string) error) error {
		return emit("one")
	}, func(string) error { return nil })
	if err != nil {
		t.Fatalf("StreamWithKeepAlive with nil context failed: %v", err)
	}
}

func TestStreamWithKeepAliveRejectsNilCallbacks(t *testing.T) {
	rec := httptest.NewRecorder()
	if err := StreamWithKeepAlive(context.Background(), rec, time.Hour, nil, func(string) error { return nil }); err == nil {
		t.Fatal("nil stream runner was accepted")
	}
	if err := StreamWithKeepAlive(context.Background(), rec, time.Hour, func(func(string) error) error { return nil }, nil); err == nil {
		t.Fatal("nil stream emitter was accepted")
	}
}

func TestChatStreamEmitsStructuredErrorWithoutAssistantMarkdown(t *testing.T) {
	app := New(config.Default(), "stream-error-test")
	app.Gem.HTTP = streamFailureRequester{}
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"gemini-3.7-flash","stream":true,"messages":[{"role":"user","content":"hello"}]}`))
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	body := rec.Body.String()
	if rec.Code != http.StatusOK {
		t.Fatalf("stream status = %d body=%s", rec.Code, body)
	}
	if !strings.Contains(body, `"error"`) || !strings.Contains(body, `"type":"api_error"`) {
		t.Fatalf("stream error was not structured: %s", body)
	}
	if strings.Contains(body, "Upstream Error") || strings.Contains(body, `"finish_reason":"error"`) {
		t.Fatalf("stream error was serialized as assistant/error-choice content: %s", body)
	}
	if !strings.Contains(body, "[DONE]") {
		t.Fatalf("stream error did not terminate with [DONE]: %s", body)
	}
}

func TestDirectGeminiStreamWithKeepAliveRejectsNilCallbacks(t *testing.T) {
	rec := httptest.NewRecorder()
	if err := streamGeminiWithKeepAlive(context.Background(), rec, time.Hour, nil, func(geminiapi.GenerateContentResponse) error { return nil }); err == nil {
		t.Fatal("nil direct stream runner was accepted")
	}
	if err := streamGeminiWithKeepAlive(context.Background(), rec, time.Hour, func(func(geminiapi.GenerateContentResponse) error) error { return nil }, nil); err == nil {
		t.Fatal("nil direct stream emitter was accepted")
	}
	if err := streamGeminiWithKeepAlive(nil, rec, time.Hour, func(emit func(geminiapi.GenerateContentResponse) error) error {
		return emit(geminiapi.GenerateContentResponse{})
	}, func(geminiapi.GenerateContentResponse) error { return nil }); err != nil {
		t.Fatalf("nil direct stream context failed: %v", err)
	}
}

func TestAnthropicMessagesStreamWithToolCalls(t *testing.T) {
	fakeToolBody := mockGeminiBody("```tool_call\n{\"name\": \"get_weather\", \"arguments\": {\"location\": \"Bengaluru\"}}\n```")
	cfg := config.Default()
	app := New(cfg, "v0.1.8")
	app.Gem.HTTP = fakeGeminiRequester{body: fakeToolBody}
	handler := app.Handler()

	payload := `{
		"model": "claude-3-7-sonnet",
		"max_tokens": 1000,
		"stream": true,
		"tools": [
			{
				"name": "get_weather",
				"description": "Get weather info",
				"input_schema": {"type": "object", "properties": {"location": {"type": "string"}}}
			}
		],
		"messages": [{"role": "user", "content": "What is the weather in Bengaluru?"}]
	}`

	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(payload))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, "event: content_block_start") || !strings.Contains(body, "\"tool_use\"") {
		t.Errorf("Expected tool_use block start in SSE stream, got:\n%s", body)
	}
	if !strings.Contains(body, "\"stop_reason\":\"tool_use\"") {
		t.Errorf("Expected stop_reason tool_use in message_delta, got:\n%s", body)
	}
}

func TestRefineEndpoint(t *testing.T) {
	fakeBody := mockGeminiBody("1. Objective: Build game.\n2. Invariants: 60 FPS.\n3. Verified output.")
	cfg := config.Default()
	app := New(cfg, "v0.2.0")
	app.Gem.HTTP = fakeGeminiRequester{body: fakeBody}
	handler := app.Handler()

	payload := `{"prompt": "Build a Cyberpunk Snake game"}`
	req := httptest.NewRequest("POST", "/v1/refine", strings.NewReader(payload))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if resp["original_prompt"] != "Build a Cyberpunk Snake game" {
		t.Errorf("Expected original_prompt, got %v", resp["original_prompt"])
	}
}

func TestRefineRejectsUnknownModel(t *testing.T) {
	app := New(config.Default(), "test-version")
	req := httptest.NewRequest("POST", "/v1/refine", strings.NewReader(`{"prompt":"Improve this","model":"not-a-gemini-model"}`))
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "invalid_request_error") {
		t.Fatalf("status/body = %d %s", rec.Code, rec.Body.String())
	}
}
