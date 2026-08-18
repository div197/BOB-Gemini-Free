package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/div197/bob-gemini-free/internal/config"
)

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
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("Expected status 204 for OPTIONS, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected CORS origin *, got %s", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestMarshalNoEscapeHTML(t *testing.T) {
	data := map[string]string{
		"text": "<hello & world>",
	}
	b, err := marshalNoEscapeHTML(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected := `{"text":"<hello & world>"}`
	if string(b) != expected {
		t.Errorf("Got %q, want %q", string(b), expected)
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

