package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/format"
	"github.com/div197/bob-gemini-free/internal/geminiapi"
)

type developerRoundTripFunc func(*http.Request) (*http.Response, error)

func (f developerRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func developerResponse(status int, contentType, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{contentType}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestChatDeveloperAPIUsesExplicitRouteAndNativeResponse(t *testing.T) {
	const providerKey = "test-provider-key"
	var gotRequest *http.Request
	provider := &http.Client{Transport: developerRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotRequest = req
		return developerResponse(http.StatusOK, "application/json", `{"candidates":[{"index":0,"content":{"role":"model","parts":[{"text":"नमस्ते विद्यार्थी"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":4,"candidatesTokenCount":7,"totalTokenCount":11}}`), nil
	})}
	cfg := config.Default()
	app := New(cfg, "test-version")
	app.GeminiAPI.HTTP = provider
	app.GeminiAPI.BaseURL = "https://provider.test"
	app.Gem.HTTP = fakeGeminiRequester{body: mockGeminiBody("web route should not be called")}

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"gemini-3.7-flash","messages":[{"role":"user","content":"Say hello"}]}`))
	req.Header.Set(geminiProviderKeyHeader, providerKey)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if gotRequest == nil || gotRequest.URL.String() != "https://provider.test/v1beta/models/gemini-3.7-flash:generateContent" {
		t.Fatalf("provider request = %#v", gotRequest)
	}
	if gotRequest.Header.Get("x-goog-api-key") != providerKey {
		t.Fatalf("upstream provider key header = %q", gotRequest.Header.Get("x-goog-api-key"))
	}
	if rec.Header().Get("x-goog-api-key") != "" || strings.Contains(rec.Body.String(), providerKey) {
		t.Fatalf("provider credential leaked in gateway response: headers=%v body=%s", rec.Header(), rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "नमस्ते विद्यार्थी") || !strings.Contains(rec.Body.String(), `"total_tokens":11`) {
		t.Fatalf("unexpected OpenAI response: %s", rec.Body.String())
	}
}

func TestChatDeveloperAPIRejectsEmptyNonstreamOutput(t *testing.T) {
	provider := &http.Client{Transport: developerRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return developerResponse(http.StatusOK, "application/json", `{"candidates":[{"finishReason":"STOP"}],"usageMetadata":{"totalTokenCount":2}}`), nil
	})}
	app := New(config.Default(), "test-version")
	app.GeminiAPI.HTTP = provider
	app.GeminiAPI.BaseURL = "https://provider.test"

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"gemini-3.7-flash","messages":[{"role":"user","content":"hello"}]}`))
	req.Header.Set(geminiProviderKeyHeader, "test-provider-key")
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway || !strings.Contains(rec.Body.String(), "no usable output") {
		t.Fatalf("empty nonstream output was accepted: status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestChatDeveloperAPIStreamPreservesDeltasAndUsage(t *testing.T) {
	provider := &http.Client{Transport: developerRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return developerResponse(http.StatusOK, "text/event-stream", ": waiting\n\ndata: {\"candidates\":[{\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"hello \"}]}}]}\n\ndata: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"world\"}]},\"finishReason\":\"STOP\"}],\"usageMetadata\":{\"promptTokenCount\":2,\"candidatesTokenCount\":2,\"totalTokenCount\":4}}\n\n"), nil
	})}
	app := New(config.Default(), "test-version")
	app.GeminiAPI.HTTP = provider
	app.GeminiAPI.BaseURL = "https://provider.test"

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"gemini-3.7-flash","stream":true,"stream_options":{"include_usage":true},"messages":[{"role":"user","content":"Say hello"}]}`))
	req.Header.Set(geminiProviderKeyHeader, "test-provider-key")
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, "hello ") || !strings.Contains(body, "world") || !strings.Contains(body, `"finish_reason":"stop"`) || !strings.Contains(body, `"total_tokens":4`) || !strings.Contains(body, "[DONE]") {
		t.Fatalf("stream response = status %d %s", rec.Code, body)
	}
}

func TestChatDeveloperAPIStreamEmitsStructuredErrorWithoutAssistantText(t *testing.T) {
	provider := &http.Client{Transport: developerRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("provider connection failed")
	})}
	app := New(config.Default(), "test-version")
	app.GeminiAPI.HTTP = provider
	app.GeminiAPI.BaseURL = "https://provider.test"

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"gemini-3.7-flash","stream":true,"messages":[{"role":"user","content":"hello"}]}`))
	req.Header.Set(geminiProviderKeyHeader, "test-provider-key")
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, `"error"`) || !strings.Contains(body, `"type":"api_error"`) {
		t.Fatalf("structured direct stream error = status %d %s", rec.Code, body)
	}
	if strings.Contains(body, "Gemini Developer API Error") || strings.Contains(body, "⚠️") {
		t.Fatalf("direct stream error was emitted as assistant text: %s", body)
	}
}

func TestChatDeveloperAPIStreamFinalizesCumulativeNativeToolCallOnce(t *testing.T) {
	provider := &http.Client{Transport: developerRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return developerResponse(http.StatusOK, "text/event-stream", "data: {\"candidates\":[{\"content\":{\"parts\":[{\"functionCall\":{\"name\":\"lookup\",\"args\":{\"city\":\"Delhi\"}}}]}}]}\n\ndata: {\"candidates\":[{\"content\":{\"parts\":[{\"functionCall\":{\"name\":\"lookup\",\"args\":{\"city\":\"Delhi\",\"language\":\"हिन्दी\"}}}]},\"finishReason\":\"STOP\"}]}\n\n"), nil
	})}
	app := New(config.Default(), "test-version")
	app.GeminiAPI.HTTP = provider
	app.GeminiAPI.BaseURL = "https://provider.test"

	requestBody := `{"model":"gemini-3.7-flash","stream":true,"tools":[{"type":"function","function":{"name":"lookup","parameters":{"type":"object","properties":{"city":{"type":"string"}}}}}],"messages":[{"role":"user","content":"lookup"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(requestBody))
	req.Header.Set(geminiProviderKeyHeader, "test-provider-key")
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	body := rec.Body.String()
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, body)
	}
	if got := strings.Count(body, `call_gemini_0_0`); got != 1 {
		t.Fatalf("native tool call ID appeared %d times, want exactly once: %s", got, body)
	}
	if !strings.Contains(body, `language`) || !strings.Contains(body, `हिन्दी`) {
		t.Fatalf("latest cumulative tool arguments were not emitted: %s", body)
	}
	if !strings.Contains(body, `"finish_reason":"tool_calls"`) || !strings.Contains(body, "[DONE]") {
		t.Fatalf("tool-call stream did not terminate with tool_calls: %s", body)
	}
}

func TestDeveloperAPIKeyDoesNotSilentlyUseWebAlias(t *testing.T) {
	app := New(config.Default(), "test-version")
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"gpt-5.6-sol","messages":[{"role":"user","content":"hello"}]}`))
	req.Header.Set(geminiProviderKeyHeader, "test-provider-key")
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "web-RPC alias") {
		t.Fatalf("status/body = %d %s", rec.Code, rec.Body.String())
	}
}

func TestDeveloperAPIRouteForwardsFutureGeminiModelID(t *testing.T) {
	var gotURL string
	provider := &http.Client{Transport: developerRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotURL = req.URL.String()
		return developerResponse(http.StatusOK, "application/json", `{"candidates":[{"content":{"parts":[{"text":"future"}]}}]}`), nil
	})}
	app := New(config.Default(), "test-version")
	app.GeminiAPI.HTTP = provider
	app.GeminiAPI.BaseURL = "https://provider.test"
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"gemini-future-flash-2027-01","messages":[{"role":"user","content":"hello"}]}`))
	req.Header.Set(geminiProviderKeyHeader, "test-provider-key")
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || gotURL != "https://provider.test/v1beta/models/gemini-future-flash-2027-01:generateContent" || !strings.Contains(rec.Body.String(), "future") {
		t.Fatalf("status/url/body = %d %s %s", rec.Code, gotURL, rec.Body.String())
	}
}

func TestDeveloperAPIKeyHeaderIsNotLocalGatewayAuthHeader(t *testing.T) {
	cfg := config.Default()
	cfg.APIKeys = []string{"local-gateway-key"}
	app := New(cfg, "test-version")
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"gemini-3.7-flash","messages":[{"role":"user","content":"hello"}]}`))
	req.Header.Set(geminiProviderKeyHeader, "test-provider-key")
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s, provider key must not satisfy local auth", rec.Code, rec.Body.String())
	}
}

func TestDeveloperAPIKeyValidationDoesNotEchoCredential(t *testing.T) {
	const providerKey = "test-provider-key"
	app := New(config.Default(), "test-version")
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"gemini-3.7-flash","messages":[{"role":"user","content":"hello"}]}`))
	req.Header.Add(geminiProviderKeyHeader, providerKey)
	req.Header.Add(geminiProviderKeyHeader, providerKey)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || strings.Contains(rec.Body.String(), providerKey) {
		t.Fatalf("status/body = %d %s", rec.Code, rec.Body.String())
	}
}

func TestCredentialRoutingMatrixKeepsGatewayAndProviderKeysSeparate(t *testing.T) {
	const (
		localKey    = "local-gateway-key"
		providerKey = "test-provider-key"
	)

	tests := []struct {
		name             string
		path             string
		body             string
		gatewayAuth      string
		configuredKey    string
		requestProvider  bool
		wantStatus       int
		wantProviderCall bool
		wantUnsupported  bool
	}{
		{
			name:             "chat explicit provider with local gateway auth",
			path:             "/v1/chat/completions",
			body:             `{"model":"gemini-3.7-flash","messages":[{"role":"user","content":"hello"}]}`,
			gatewayAuth:      localKey,
			requestProvider:  true,
			wantStatus:       http.StatusOK,
			wantProviderCall: true,
		},
		{
			name:             "native Google explicit provider with local gateway auth",
			path:             "/v1beta/models/gemini-3.7-flash:generateContent",
			body:             `{"contents":[{"role":"user","parts":[{"text":"hello"}]}]}`,
			gatewayAuth:      localKey,
			requestProvider:  true,
			wantStatus:       http.StatusOK,
			wantProviderCall: true,
		},
		{
			name:            "anthropic rejects explicit provider",
			path:            "/v1/messages",
			body:            `{"model":"claude-3-7-sonnet","messages":[]}`,
			gatewayAuth:     localKey,
			requestProvider: true,
			wantStatus:      http.StatusBadRequest,
			wantUnsupported: true,
		},
		{
			name:            "responses rejects explicit provider",
			path:            "/v1/responses",
			body:            `{"model":"gpt-5.6-sol","input":"hello"}`,
			gatewayAuth:     localKey,
			requestProvider: true,
			wantStatus:      http.StatusBadRequest,
			wantUnsupported: true,
		},
		{
			name:            "images rejects explicit provider",
			path:            "/v1/images/generations",
			body:            `{"prompt":"draw a lotus"}`,
			gatewayAuth:     localKey,
			requestProvider: true,
			wantStatus:      http.StatusBadRequest,
			wantUnsupported: true,
		},
		{
			name:             "configured provider key selects direct chat route",
			path:             "/v1/chat/completions",
			body:             `{"model":"gemini-3.7-flash","messages":[{"role":"user","content":"hello"}]}`,
			gatewayAuth:      localKey,
			configuredKey:    providerKey,
			wantStatus:       http.StatusOK,
			wantProviderCall: true,
		},
		{
			name:            "provider key alone cannot satisfy gateway auth",
			path:            "/v1/chat/completions",
			body:            `{"model":"gemini-3.7-flash","messages":[{"role":"user","content":"hello"}]}`,
			gatewayAuth:     "",
			requestProvider: true,
			wantStatus:      http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Default()
			cfg.APIKeys = []string{localKey}
			cfg.GeminiAPIKey = tt.configuredKey
			app := New(cfg, "test-version")
			providerCalls := 0
			var upstreamKey string
			app.GeminiAPI.HTTP = &http.Client{Transport: developerRoundTripFunc(func(req *http.Request) (*http.Response, error) {
				providerCalls++
				upstreamKey = req.Header.Get("x-goog-api-key")
				return developerResponse(http.StatusOK, "application/json", `{"candidates":[{"content":{"parts":[{"text":"provider response"}]}}]}`), nil
			})}
			app.GeminiAPI.BaseURL = "https://provider.test"

			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
			if tt.gatewayAuth != "" {
				req.Header.Set("Authorization", "Bearer "+tt.gatewayAuth)
			}
			if tt.requestProvider {
				req.Header.Set(geminiProviderKeyHeader, providerKey)
			}
			rec := httptest.NewRecorder()
			app.Handler().ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
			}
			if got := providerCalls > 0; got != tt.wantProviderCall {
				t.Fatalf("provider call = %v, want %v; calls=%d", got, tt.wantProviderCall, providerCalls)
			}
			if tt.wantProviderCall && upstreamKey != providerKey {
				t.Fatalf("upstream provider key = %q, want %q", upstreamKey, providerKey)
			}
			if !tt.wantProviderCall && upstreamKey != "" {
				t.Fatalf("unexpected upstream provider key = %q", upstreamKey)
			}
			if tt.wantUnsupported && !strings.Contains(rec.Body.String(), "not supported on") {
				t.Fatalf("unsupported provider route did not return sanitized error: %s", rec.Body.String())
			}
			if strings.Contains(rec.Body.String(), providerKey) {
				t.Fatalf("provider key leaked in response: %s", rec.Body.String())
			}
		})
	}
}

func TestGoogleDeveloperAPIRouteForwardsNativeJSON(t *testing.T) {
	var gotURL string
	provider := &http.Client{Transport: developerRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotURL = req.URL.String()
		return developerResponse(http.StatusOK, "application/json", `{"candidates":[{"content":{"parts":[{"text":"native"}]}}]}`), nil
	})}
	app := New(config.Default(), "test-version")
	app.GeminiAPI.HTTP = provider
	app.GeminiAPI.BaseURL = "https://provider.test"
	req := httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-3.7-flash:generateContent", strings.NewReader(`{"contents":[{"role":"user","parts":[{"text":"hello"}]}]}`))
	req.Header.Set(geminiProviderKeyHeader, "test-provider-key")
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || gotURL != "https://provider.test/v1beta/models/gemini-3.7-flash:generateContent" || !strings.Contains(rec.Body.String(), "native") {
		t.Fatalf("status/url/body = %d %s %s", rec.Code, gotURL, rec.Body.String())
	}
}

func TestGoogleDeveloperAPIRouteRejectsSemanticEmptyJSON(t *testing.T) {
	for _, body := range []string{
		`{"candidates":[{"finishReason":"STOP"}],"usageMetadata":{"totalTokenCount":2}}`,
		`{"usageMetadata":{"totalTokenCount":2}}`,
	} {
		provider := &http.Client{Transport: developerRoundTripFunc(func(*http.Request) (*http.Response, error) {
			return developerResponse(http.StatusOK, "application/json", body), nil
		})}
		app := New(config.Default(), "test-version")
		app.GeminiAPI.HTTP = provider
		app.GeminiAPI.BaseURL = "https://provider.test"
		req := httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-3.7-flash:generateContent", strings.NewReader(`{"contents":[{"role":"user","parts":[{"text":"hello"}]}]}`))
		req.Header.Set(geminiProviderKeyHeader, "test-provider-key")
		rec := httptest.NewRecorder()
		app.Handler().ServeHTTP(rec, req)
		if rec.Code != http.StatusBadGateway || !strings.Contains(rec.Body.String(), "no") {
			t.Fatalf("semantic empty native response was accepted: status=%d body=%s", rec.Code, rec.Body.String())
		}
	}
}

func TestGoogleDeveloperAPIRouteRejectsSemanticEmptyStream(t *testing.T) {
	provider := &http.Client{Transport: developerRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return developerResponse(http.StatusOK, "text/event-stream", "data: {\"candidates\":[{\"finishReason\":\"STOP\"}]}\n\ndata: {\"usageMetadata\":{\"totalTokenCount\":2}}\n\n"), nil
	})}
	app := New(config.Default(), "test-version")
	app.GeminiAPI.HTTP = provider
	app.GeminiAPI.BaseURL = "https://provider.test"
	req := httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-3.7-flash:streamGenerateContent", strings.NewReader(`{"contents":[{"role":"user","parts":[{"text":"hello"}]}]}`))
	req.Header.Set(geminiProviderKeyHeader, "test-provider-key")
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, `"error"`) || !strings.Contains(body, "no usable stream content") {
		t.Fatalf("semantic empty native stream was accepted: status=%d body=%s", rec.Code, body)
	}
}

func TestGoogleDeveloperAPIRoutePreservesNativeStreamEvents(t *testing.T) {
	provider := &http.Client{Transport: developerRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return developerResponse(http.StatusOK, "text/event-stream", "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"hello\"}]}}]}\n\ndata: {\"candidates\":[{\"finishReason\":\"STOP\"}]}\n\n"), nil
	})}
	app := New(config.Default(), "test-version")
	app.GeminiAPI.HTTP = provider
	app.GeminiAPI.BaseURL = "https://provider.test"
	req := httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-3.7-flash:streamGenerateContent", strings.NewReader(`{"contents":[{"role":"user","parts":[{"text":"hello"}]}]}`))
	req.Header.Set(geminiProviderKeyHeader, "test-provider-key")
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, `"text":"hello"`) || !strings.Contains(body, `"finishReason":"STOP"`) || strings.Contains(body, `"no usable stream content"`) {
		t.Fatalf("native stream events were not preserved: status=%d body=%s", rec.Code, body)
	}
}

func TestUnsupportedDeveloperAPIRouteIsExplicitlyRejected(t *testing.T) {
	app := New(config.Default(), "test-version")
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(`{"model":"claude-3-7-sonnet","messages":[]}`))
	req.Header.Set(geminiProviderKeyHeader, "test-provider-key")
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "not supported") {
		t.Fatalf("status/body = %d %s", rec.Code, rec.Body.String())
	}
}

func TestOpenAIStreamDeltaRejectsUntrustedToolOutput(t *testing.T) {
	if _, err := openAIStreamDelta(geminiapi.GenerateContentResponse{Candidates: []geminiapi.Candidate{{Content: &geminiapi.Content{Parts: []geminiapi.Part{{
		FunctionCall: &geminiapi.FunctionCall{Name: "", Args: map[string]any{}},
	}}}}}}); err == nil || !strings.Contains(err.Error(), "name is empty") {
		t.Fatalf("empty tool name error = %v", err)
	}
	if _, err := openAIStreamDelta(geminiapi.GenerateContentResponse{Candidates: []geminiapi.Candidate{{Content: &geminiapi.Content{Parts: []geminiapi.Part{{
		FunctionCall: &geminiapi.FunctionCall{Name: "large", Args: map[string]any{"value": strings.Repeat("x", format.MaxToolArgumentBytes)}},
	}}}}}}); err == nil || !strings.Contains(err.Error(), "arguments exceed") {
		t.Fatalf("oversized tool arguments error = %v", err)
	}

	parts := make([]geminiapi.Part, format.MaxToolDefinitions+1)
	for i := range parts {
		parts[i].FunctionCall = &geminiapi.FunctionCall{Name: "lookup", Args: map[string]any{"index": i}}
	}
	if _, err := openAIStreamDelta(geminiapi.GenerateContentResponse{Candidates: []geminiapi.Candidate{{Content: &geminiapi.Content{Parts: parts}}}}); err == nil || !strings.Contains(err.Error(), "more than") {
		t.Fatalf("tool-count error = %v", err)
	}

	encoded, err := json.Marshal(geminiapi.GenerateContentResponse{Candidates: []geminiapi.Candidate{{Content: &geminiapi.Content{Parts: []geminiapi.Part{{
		FunctionCall: &geminiapi.FunctionCall{Name: "lookup", Args: map[string]any{"q": "भारत"}},
	}}}}}})
	if err != nil || !strings.Contains(string(encoded), "भारत") {
		t.Fatalf("fixture sanity check failed: %v %s", err, encoded)
	}
}

func TestOpenAIStreamDeltaRejectsMultipleCandidatesAndUnknownFinishReason(t *testing.T) {
	response := geminiapi.GenerateContentResponse{Candidates: []geminiapi.Candidate{
		{Content: &geminiapi.Content{Parts: []geminiapi.Part{{Text: "one"}}}},
		{Content: &geminiapi.Content{Parts: []geminiapi.Part{{Text: "two"}}}},
	}}
	if _, err := openAIStreamDelta(response); err == nil || !strings.Contains(err.Error(), "exactly one") {
		t.Fatalf("multiple candidate result = %v", err)
	}
	unknown := geminiapi.GenerateContentResponse{Candidates: []geminiapi.Candidate{{FinishReason: "NEW_PROVIDER_REASON"}}}
	if _, err := openAIStreamDelta(unknown); err == nil || !strings.Contains(err.Error(), "unsupported Gemini finish reason") {
		t.Fatalf("unknown finish reason result = %v", err)
	}
}
