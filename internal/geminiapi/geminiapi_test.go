package geminiapi

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/div197/bob-gemini-free/internal/format"
	"github.com/div197/bob-gemini-free/internal/models"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func (f roundTripFunc) Do(r *http.Request) (*http.Response, error)        { return f(r) }

func response(status int, contentType, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{contentType}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestGenerateUsesHeaderAndPublicEndpoint(t *testing.T) {
	const apiKey = "test-provider-key"
	var got *http.Request
	var gotBody string
	httpClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		got = req
		body, _ := io.ReadAll(req.Body)
		gotBody = string(body)
		return response(http.StatusOK, "application/json", `{"candidates":[{"content":{"role":"model","parts":[{"text":"hello"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":2,"candidatesTokenCount":1,"totalTokenCount":3}}`), nil
	})}

	client := NewClient(httpClient)
	client.BaseURL = "https://provider.test"
	result, err := client.Generate(context.Background(), "gemini-3.7-flash", apiKey, GenerateContentRequest{
		Contents: []Content{{Role: "user", Parts: []Part{{Text: "hi"}}}},
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if got == nil {
		t.Fatal("transport was not called")
	}
	if got.URL.String() != "https://provider.test/v1beta/models/gemini-3.7-flash:generateContent" {
		t.Fatalf("URL = %s", got.URL)
	}
	if got.URL.Query().Get("key") != "" || got.URL.Query().Get("api_key") != "" {
		t.Fatalf("provider key leaked into query: %s", got.URL.RawQuery)
	}
	if got.Header.Get(APIKeyHeader) != apiKey {
		t.Fatalf("provider header = %q", got.Header.Get(APIKeyHeader))
	}
	if strings.Contains(gotBody, apiKey) {
		t.Fatalf("provider key leaked into body: %s", gotBody)
	}
	if result.Candidates[0].Content.Parts[0].Text != "hello" || result.UsageMetadata.TotalTokenCount != 3 {
		t.Fatalf("decoded response = %#v", result)
	}
}

func TestStreamParsesSSEAndIgnoresCommentsAndDone(t *testing.T) {
	var events []GenerateContentResponse
	client := NewClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.RawQuery != "alt=sse" {
			t.Fatalf("stream query = %q", req.URL.RawQuery)
		}
		if req.Header.Get("Accept") != "text/event-stream" {
			t.Fatalf("stream Accept = %q", req.Header.Get("Accept"))
		}
		return response(http.StatusOK, "text/event-stream", ": keepalive\n\ndata: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"a\"}]}}]}\n\ndata: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"b\"}]},\"finishReason\":\"STOP\"}]}\n\ndata: [DONE]\n\n"), nil
	})})
	client.BaseURL = "https://provider.test"
	err := client.Stream(context.Background(), "models/gemini-3.7-flash", "stream-key", GenerateContentRequest{}, func(event GenerateContentResponse) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}
	if len(events) != 2 || events[0].Candidates[0].Content.Parts[0].Text != "a" || events[1].Candidates[0].FinishReason != "STOP" {
		t.Fatalf("events = %#v", events)
	}
}

func TestMalformedStreamEventFailsClosed(t *testing.T) {
	client := NewClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return response(http.StatusOK, "text/event-stream", "data: {not-json}\n\n"), nil
	})})
	err := client.Stream(context.Background(), "gemini-3.7-flash", "stream-key", GenerateContentRequest{}, func(GenerateContentResponse) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "invalid stream JSON") {
		t.Fatalf("error = %v", err)
	}
}

func TestEmptyStreamFailsClosed(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "empty body", body: ""},
		{name: "done only", body: ": keepalive\n\ndata: [DONE]\n\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := NewClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return response(http.StatusOK, "text/event-stream", test.body), nil
			})})
			err := client.Stream(context.Background(), "gemini-3.7-flash", "stream-key", GenerateContentRequest{}, func(GenerateContentResponse) error { return nil })
			if err == nil || !strings.Contains(err.Error(), "empty stream") {
				t.Fatalf("error = %v, want empty stream failure", err)
			}
		})
	}
}

func TestProviderEmptyResponseFailsClosed(t *testing.T) {
	client := NewClient(roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, nil
	}))
	if _, err := client.Generate(context.Background(), "gemini-3.7-flash", "stream-key", GenerateContentRequest{}); err == nil || !strings.Contains(err.Error(), "empty response") {
		t.Fatalf("nil Generate response error = %v", err)
	}
	if err := client.Stream(context.Background(), "gemini-3.7-flash", "stream-key", GenerateContentRequest{}, func(GenerateContentResponse) error { return nil }); err == nil || !strings.Contains(err.Error(), "empty response") {
		t.Fatalf("nil Stream response error = %v", err)
	}

	client.HTTP = roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header)}, nil
	})
	if _, err := client.Generate(context.Background(), "gemini-3.7-flash", "stream-key", GenerateContentRequest{}); err == nil || !strings.Contains(err.Error(), "empty response body") {
		t.Fatalf("nil Body response error = %v", err)
	}
}

func TestClientNormalizesNilContext(t *testing.T) {
	client := NewClient(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Context() == nil {
			t.Fatal("request context is nil")
		}
		return response(http.StatusOK, "application/json", `{"candidates":[]}`), nil
	}))
	if _, err := client.Generate(nil, "gemini-3.7-flash", "context-key", GenerateContentRequest{}); err != nil {
		t.Fatalf("Generate(nil context) error = %v", err)
	}

	client.HTTP = roundTripFunc(func(*http.Request) (*http.Response, error) {
		return response(http.StatusOK, "text/event-stream", "data: {\"candidates\":[]}\n\ndata: [DONE]\n\n"), nil
	})
	if err := client.Stream(nil, "gemini-3.7-flash", "context-key", GenerateContentRequest{}, func(GenerateContentResponse) error { return nil }); err != nil {
		t.Fatalf("Stream(nil context) error = %v", err)
	}
}

func TestStreamRejectsAggregateSSEBodyLimit(t *testing.T) {
	client := NewClient(roundTripFunc(func(*http.Request) (*http.Response, error) {
		return response(http.StatusOK, "text/event-stream", strings.Repeat("data: {}\n\n", 10)), nil
	}))
	client.MaxResponseBody = 32
	err := client.Stream(context.Background(), "gemini-3.7-flash", "stream-key", GenerateContentRequest{}, func(GenerateContentResponse) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "stream is too large") {
		t.Fatalf("aggregate stream limit error = %v", err)
	}
}

func TestProviderErrorDoesNotEchoKey(t *testing.T) {
	const apiKey = "secret-test-provider-key"
	client := NewClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return response(http.StatusForbidden, "application/json", `{"error":{"message":"bad key secret-test-provider-key","status":"PERMISSION_DENIED"}}`), nil
	})})
	_, err := client.Generate(context.Background(), "gemini-3.7-flash", apiKey, GenerateContentRequest{})
	if err == nil || strings.Contains(err.Error(), apiKey) || !strings.Contains(err.Error(), "PERMISSION_DENIED") {
		t.Fatalf("sanitized error = %v", err)
	}
}

func TestFromOpenAIPreservesSystemImageToolsAndChoice(t *testing.T) {
	request := models.OpenAIChatRequest{
		Messages: []models.OpenAIMessage{
			{Role: "system", Content: "Be concise."},
			{Role: "user", Content: []any{
				map[string]any{"type": "text", "text": "What is this?"},
				map[string]any{"type": "image_url", "image_url": map[string]any{"url": "data:image/png;base64,aGk="}},
			}},
		},
		Tools: []models.OpenAITool{{Type: "function", Function: models.OpenAIFunction{
			Name: "lookup", Description: "Look something up.", Parameters: map[string]any{
				"type": "object", "properties": map[string]any{"q": map[string]any{"type": "string"}},
			},
		}}},
		ToolChoice:  map[string]any{"type": "function", "function": map[string]any{"name": "lookup"}},
		Temperature: func() *float64 { value := 0.2; return &value }(),
		MaxTokens:   func() *int { value := 256; return &value }(),
	}

	translated, err := FromOpenAI(request)
	if err != nil {
		t.Fatalf("FromOpenAI() error = %v", err)
	}
	if translated.SystemInstruction == nil || translated.SystemInstruction.Parts[0].Text != "Be concise." {
		t.Fatalf("system instruction = %#v", translated.SystemInstruction)
	}
	if len(translated.Contents) != 1 || len(translated.Contents[0].Parts) != 2 || translated.Contents[0].Parts[1].InlineData == nil {
		t.Fatalf("contents = %#v", translated.Contents)
	}
	if translated.Tools[0].FunctionDeclarations[0].Name != "lookup" || translated.ToolConfig.FunctionCallingConfig.Mode != "ANY" || translated.ToolConfig.FunctionCallingConfig.AllowedFunctionNames[0] != "lookup" {
		t.Fatalf("tools/config = %#v %#v", translated.Tools, translated.ToolConfig)
	}
	if translated.GenerationConfig == nil || translated.GenerationConfig.MaxOutputTokens == nil || *translated.GenerationConfig.MaxOutputTokens != 256 {
		t.Fatalf("generation config = %#v", translated.GenerationConfig)
	}
}

func TestFromOpenAIRejectsRemoteImagesAndInvalidToolArguments(t *testing.T) {
	_, err := FromOpenAI(models.OpenAIChatRequest{Messages: []models.OpenAIMessage{{
		Role: "user", Content: []any{map[string]any{"type": "image_url", "image_url": map[string]any{"url": "https://example.test/image.png"}}},
	}}})
	if err == nil || !strings.Contains(err.Error(), "data URLs only") {
		t.Fatalf("remote image error = %v", err)
	}

	_, err = FromOpenAI(models.OpenAIChatRequest{Messages: []models.OpenAIMessage{{
		Role: "assistant", ToolCalls: []models.OpenAIToolCall{{Function: models.OpenAIToolCallFunction{Name: "lookup", Arguments: "{"}}},
	}}})
	if err == nil || !strings.Contains(err.Error(), "invalid JSON arguments") {
		t.Fatalf("tool argument error = %v", err)
	}
}

func TestFromOpenAIRejectsUnsafeInlineImages(t *testing.T) {
	for _, imageURL := range []string{
		"data:text/plain;base64,aGk=",
		"data:image/png;base64," + strings.Repeat("A", MaxInlineImageBase64Size+1),
	} {
		_, err := FromOpenAI(models.OpenAIChatRequest{Messages: []models.OpenAIMessage{{
			Role: "user", Content: []any{map[string]any{"type": "image_url", "image_url": map[string]any{"url": imageURL}}},
		}}})
		if err == nil {
			t.Fatalf("unsafe inline image was accepted: %q", imageURL[:min(len(imageURL), 32)])
		}
	}
}

func TestFromOpenAIRejectsUnsupportedStructuredOutputAndToolChoice(t *testing.T) {
	base := models.OpenAIChatRequest{
		Messages: []models.OpenAIMessage{{Role: "user", Content: "hello"}},
		Tools:    []models.OpenAITool{{Type: "function", Function: models.OpenAIFunction{Name: "lookup"}}},
	}
	base.ResponseFormat = &models.OpenAIResponseFormat{Type: "json_schema"}
	if _, err := FromOpenAI(base); err == nil || !strings.Contains(err.Error(), "json_schema") {
		t.Fatalf("json_schema error = %v", err)
	}

	base.ResponseFormat = nil
	base.ToolChoice = "unsupported"
	if _, err := FromOpenAI(base); err == nil || !strings.Contains(err.Error(), "tool_choice") {
		t.Fatalf("tool_choice error = %v", err)
	}

	base.ToolChoice = map[string]any{"type": "computer"}
	if _, err := FromOpenAI(base); err == nil || !strings.Contains(err.Error(), "tool_choice type") {
		t.Fatalf("tool_choice type error = %v", err)
	}

	base.ToolChoice = map[string]any{"type": "function", "function": map[string]any{"name": "delete_all"}}
	if _, err := FromOpenAI(base); err == nil || !strings.Contains(err.Error(), "undeclared tool") {
		t.Fatalf("undeclared tool choice error = %v", err)
	}
}

func TestFromOpenAIEnforcesToolBudgets(t *testing.T) {
	tooMany := make([]models.OpenAITool, format.MaxToolDefinitions+1)
	for i := range tooMany {
		tooMany[i] = models.OpenAITool{Type: "function", Function: models.OpenAIFunction{Name: "tool"}}
	}
	if _, err := FromOpenAI(models.OpenAIChatRequest{Tools: tooMany}); err == nil {
		t.Fatal("oversized Developer API tool count was accepted")
	}

	if _, err := FromOpenAI(models.OpenAIChatRequest{Tools: []models.OpenAITool{{Type: "function", Function: models.OpenAIFunction{
		Name: "large", Description: strings.Repeat("x", format.MaxToolDescriptionBytes+1),
	}}}}); err == nil {
		t.Fatal("oversized Developer API tool description was accepted")
	}

	if _, err := FromOpenAI(models.OpenAIChatRequest{Messages: []models.OpenAIMessage{{
		Role: "assistant", ToolCalls: []models.OpenAIToolCall{{Function: models.OpenAIToolCallFunction{
			Name: "large", Arguments: strings.Repeat("x", format.MaxToolArgumentBytes+1),
		}}},
	}}}); err == nil {
		t.Fatal("oversized Developer API tool arguments were accepted")
	}
}

func TestToOpenAIResultPreservesThinkingAndNativeToolCalls(t *testing.T) {
	response := GenerateContentResponse{
		Candidates: []Candidate{{
			Content: &Content{Parts: []Part{
				{Text: "internal", Thought: true},
				{Text: "answer"},
				{FunctionCall: &FunctionCall{Name: "lookup", Args: map[string]any{"q": "भारत"}}},
			}},
			FinishReason: "STOP",
		}},
		UsageMetadata: &UsageMetadata{PromptTokenCount: 4, CandidatesTokenCount: 5, ThoughtsTokenCount: 2, TotalTokenCount: 9},
	}
	result, err := ToOpenAIResult(response)
	if err != nil {
		t.Fatalf("ToOpenAIResult() error = %v", err)
	}
	if result.Thinking != "internal" || result.Text != "answer" || len(result.ToolCalls) != 1 || result.FinishReason != "tool_calls" {
		t.Fatalf("result = %#v", result)
	}
	if !strings.Contains(result.ToolCalls[0].Function.Arguments, "भारत") || result.ReasoningTokens != 2 {
		t.Fatalf("tool/result usage = %#v", result)
	}
}

func TestToOpenAIResultRejectsUntrustedToolOutput(t *testing.T) {
	tests := []struct {
		name     string
		response GenerateContentResponse
		want     string
	}{
		{
			name: "empty name",
			response: GenerateContentResponse{Candidates: []Candidate{{Content: &Content{Parts: []Part{{
				FunctionCall: &FunctionCall{Args: map[string]any{}},
			}}}}}},
			want: "name is empty",
		},
		{
			name: "oversized arguments",
			response: GenerateContentResponse{Candidates: []Candidate{{Content: &Content{Parts: []Part{{
				FunctionCall: &FunctionCall{Name: "large", Args: map[string]any{"value": strings.Repeat("x", format.MaxToolArgumentBytes)}},
			}}}}}},
			want: "arguments exceed",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ToOpenAIResult(test.response); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("ToOpenAIResult() error = %v, want %q", err, test.want)
			}
		})
	}

	parts := make([]Part, format.MaxToolDefinitions+1)
	for i := range parts {
		parts[i].FunctionCall = &FunctionCall{Name: "lookup", Args: map[string]any{"index": i}}
	}
	if _, err := ToOpenAIResult(GenerateContentResponse{Candidates: []Candidate{{Content: &Content{Parts: parts}}}}); err == nil || !strings.Contains(err.Error(), "more than") {
		t.Fatalf("ToOpenAIResult() tool-count error = %v", err)
	}
}

func TestToOpenAIResultRejectsMultipleCandidatesAndUnknownFinishReason(t *testing.T) {
	multiple := GenerateContentResponse{Candidates: []Candidate{
		{Content: &Content{Parts: []Part{{Text: "one"}}}},
		{Content: &Content{Parts: []Part{{Text: "two"}}}},
	}}
	if _, err := ToOpenAIResult(multiple); err == nil || !strings.Contains(err.Error(), "exactly one") {
		t.Fatalf("multiple candidate result = %v", err)
	}
	unknown := GenerateContentResponse{Candidates: []Candidate{{FinishReason: "NEW_PROVIDER_REASON"}}}
	if _, err := ToOpenAIResult(unknown); err == nil || !strings.Contains(err.Error(), "unsupported Gemini finish reason") {
		t.Fatalf("unknown finish reason result = %v", err)
	}
}
