package server

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/div197/bob-gemini-free/internal/config"
)

// anthropicSSEData returns the decoded JSON payload of every `data:` line in an
// SSE body, in order.
func anthropicSSEData(t *testing.T, body string) []map[string]any {
	t.Helper()
	var events []map[string]any
	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		var event map[string]any
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &event); err != nil {
			t.Fatalf("invalid SSE data %q: %v", line, err)
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan SSE body: %v", err)
	}
	return events
}

// TestAnthropicStreamWithToolsDoesNotLeakToolFenceAsText is the H1 regression:
// with tools declared, the streaming branch used to emit the raw tool_call
// fence as text_delta and then append a real tool_use block for the same call.
// Chat and Responses already buffer in this case; Anthropic must do the same.
func TestAnthropicStreamWithToolsDoesNotLeakToolFenceAsText(t *testing.T) {
	upstream := mockGeminiBody("Let me check the weather.\n```tool_call\n{\"name\": \"get_weather\", \"arguments\": {\"location\": \"Bengaluru\"}}\n```")
	app := New(config.Default(), "test-version")
	app.Gem.HTTP = fakeGeminiRequester{body: upstream}

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
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(payload))
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var textBuf strings.Builder
	var toolUseCount int
	for _, event := range anthropicSSEData(t, rec.Body.String()) {
		switch event["type"] {
		case "content_block_delta":
			delta, _ := event["delta"].(map[string]any)
			if delta["type"] == "text_delta" {
				text, _ := delta["text"].(string)
				textBuf.WriteString(text)
			}
		case "content_block_start":
			block, _ := event["content_block"].(map[string]any)
			if block["type"] == "tool_use" {
				toolUseCount++
			}
		}
	}

	streamedText := textBuf.String()
	if strings.Contains(streamedText, "tool_call") || strings.Contains(streamedText, "```") {
		t.Fatalf("tool-call fence leaked into text_delta: %q", streamedText)
	}
	if !strings.Contains(streamedText, "Let me check the weather.") {
		t.Fatalf("preamble text missing from text_delta: %q", streamedText)
	}
	if toolUseCount != 1 {
		t.Fatalf("tool_use blocks = %d, want 1", toolUseCount)
	}
	if !strings.Contains(rec.Body.String(), `"stop_reason":"tool_use"`) {
		t.Fatalf("stop_reason tool_use missing: %s", rec.Body.String())
	}
}

// unpaddedGeminiBody wraps text so the answer is exactly text. The parser needs
// a long enough wire line, so the filler goes in an unused inner slot rather
// than in the answer as mockGeminiBody does.
func unpaddedGeminiBody(text string) string {
	inner := []any{nil, nil, nil, nil, []any{[]any{nil, []any{text}}}, nil, strings.Repeat("x", 200)}
	innerBytes, _ := json.Marshal(inner)
	outer, _ := json.Marshal([]any{[]any{"wrb.fr", nil, string(innerBytes)}})
	return string(outer) + "\n"
}

func anthropicStreamWithTools(t *testing.T, upstream string) (blocks []map[string]any, body string, stopReason string) {
	t.Helper()
	app := New(config.Default(), "test-version")
	app.Gem.HTTP = fakeGeminiRequester{body: upstream}
	payload := `{
		"model": "claude-3-7-sonnet",
		"max_tokens": 1000,
		"stream": true,
		"tools": [{"name": "get_weather", "description": "Get weather info", "input_schema": {"type": "object"}}],
		"messages": [{"role": "user", "content": "What is the weather in Bengaluru?"}]
	}`
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(payload)))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	body = rec.Body.String()
	for _, event := range anthropicSSEData(t, body) {
		switch event["type"] {
		case "content_block_start":
			blocks = append(blocks, event)
		case "message_delta":
			delta, _ := event["delta"].(map[string]any)
			stopReason, _ = delta["stop_reason"].(string)
		}
	}
	return blocks, body, stopReason
}

// A tool-only answer must not open an empty text block before tool_use, and the
// tool block must start at index 0.
func TestAnthropicStreamToolOnlyAnswerHasNoEmptyTextBlock(t *testing.T) {
	upstream := unpaddedGeminiBody("```tool_call\n{\"name\": \"get_weather\", \"arguments\": {\"location\": \"Bengaluru\"}}\n```")
	blocks, body, stopReason := anthropicStreamWithTools(t, upstream)

	if len(blocks) != 1 {
		t.Fatalf("content blocks = %d, want exactly one tool_use: %s", len(blocks), body)
	}
	block, _ := blocks[0]["content_block"].(map[string]any)
	if block["type"] != "tool_use" {
		t.Fatalf("only block is %v, want tool_use: %s", block["type"], body)
	}
	if index, _ := blocks[0]["index"].(float64); index != 0 {
		t.Fatalf("tool_use index = %v, want 0: %s", index, body)
	}
	if stopReason != "tool_use" {
		t.Fatalf("stop_reason = %q, want tool_use", stopReason)
	}
}

// With tools declared but no call made, the plain answer streams as text with
// end_turn and no tool_use block.
func TestAnthropicStreamToolsDeclaredPlainAnswer(t *testing.T) {
	blocks, body, stopReason := anthropicStreamWithTools(t, unpaddedGeminiBody("It is sunny in Bengaluru today."))

	if stopReason != "end_turn" {
		t.Fatalf("stop_reason = %q, want end_turn: %s", stopReason, body)
	}
	for _, block := range blocks {
		content, _ := block["content_block"].(map[string]any)
		if content["type"] == "tool_use" {
			t.Fatalf("plain answer produced a tool_use block: %s", body)
		}
	}
	if !strings.Contains(body, "It is sunny in Bengaluru today.") {
		t.Fatalf("plain answer missing from stream: %s", body)
	}
}
