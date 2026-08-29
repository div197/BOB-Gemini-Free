package format

import (
	"strings"
	"testing"

	"github.com/div197/bob-gemini-free/internal/models"
)

func TestAnthropicToOpenAIChatRequest(t *testing.T) {
	req := models.AnthropicMessagesRequest{
		Model:  "gemini-3.7-flash",
		System: "You are a senior Claude engineer.",
		Messages: []models.AnthropicMessage{
			{
				Role:    "user",
				Content: "Run test suite",
			},
			{
				Role: "assistant",
				Content: []any{
					map[string]any{
						"type": "text",
						"text": "I will execute the test tool.",
					},
					map[string]any{
						"type":  "tool_use",
						"id":    "toolu_01",
						"name":  "run_tests",
						"input": map[string]any{"flags": "-v"},
					},
				},
			},
			{
				Role: "user",
				Content: []any{
					map[string]any{
						"type":        "tool_result",
						"tool_use_id": "toolu_01",
						"content":     "PASS",
					},
				},
			},
		},
		Tools: []models.AnthropicTool{
			{
				Name:        "run_tests",
				Description: "Run test suite",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"flags": map[string]string{"type": "string"},
					},
				},
			},
		},
	}

	chatReq, err := AnthropicToOpenAIChatRequest(req)
	if err != nil {
		t.Fatalf("AnthropicToOpenAIChatRequest failed: %v", err)
	}
	if len(chatReq.Messages) < 4 {
		t.Fatalf("Expected at least 4 converted messages, got %d", len(chatReq.Messages))
	}
	if chatReq.Messages[0].Role != "system" {
		t.Errorf("Expected first message role system, got %s", chatReq.Messages[0].Role)
	}
	if len(chatReq.Tools) != 1 {
		t.Fatalf("Expected 1 tool converted, got %d", len(chatReq.Tools))
	}
	if chatReq.Tools[0].Function.Name != "run_tests" {
		t.Errorf("Expected tool name run_tests, got %s", chatReq.Tools[0].Function.Name)
	}

	// Test ConvertToolCallsToAnthropicBlocks
	toolCalls := []models.OpenAIToolCall{
		{
			ID:   "toolu_123",
			Type: "function",
			Function: models.OpenAIToolCallFunction{
				Name:      "run_tests",
				Arguments: `{"flags":"-v"}`,
			},
		},
	}
	blocks := ConvertToolCallsToAnthropicBlocks("Tests initiated", toolCalls)
	if len(blocks) != 2 {
		t.Fatalf("Expected 2 content blocks (text + tool_use), got %d", len(blocks))
	}
	if blocks[0].Type != "text" || !strings.Contains(blocks[0].Text, "Tests initiated") {
		t.Errorf("Unexpected text block: %+v", blocks[0])
	}
	if blocks[1].Type != "tool_use" || blocks[1].Name != "run_tests" {
		t.Errorf("Unexpected tool_use block: %+v", blocks[1])
	}
}

func TestAnthropicThinkingConversion(t *testing.T) {
	req := models.AnthropicMessagesRequest{
		Model: "claude-3-7-sonnet",
		Thinking: &models.AnthropicThinking{
			Type:         "enabled",
			BudgetTokens: 2048,
		},
		Messages: []models.AnthropicMessage{
			{Role: "user", Content: "Calculate complex equation"},
		},
	}
	chatReq, err := AnthropicToOpenAIChatRequest(req)
	if err != nil {
		t.Fatalf("AnthropicToOpenAIChatRequest failed: %v", err)
	}
	if chatReq.ReasoningEffort != "high" {
		t.Errorf("Expected reasoning effort 'high', got %s", chatReq.ReasoningEffort)
	}

	blocks := ConvertToolCallsAndThinkingToAnthropicBlocks("Thinking steps here", "Result: 42", nil)
	if len(blocks) != 2 {
		t.Fatalf("Expected 2 blocks (thinking + text), got %d", len(blocks))
	}
	if blocks[0].Type != "thinking" || blocks[0].Thinking != "Thinking steps here" {
		t.Errorf("Unexpected thinking block: %+v", blocks[0])
	}
	if blocks[1].Type != "text" || blocks[1].Text != "Result: 42" {
		t.Errorf("Unexpected text block: %+v", blocks[1])
	}
}

func TestAnthropicImageAndStructuredSystemConversion(t *testing.T) {
	req := models.AnthropicMessagesRequest{
		Model: "claude-3-7-sonnet",
		System: map[string]any{
			"type": "text",
			"text": "System instructions in map",
		},
		Messages: []models.AnthropicMessage{
			{
				Role: "user",
				Content: []any{
					map[string]any{
						"type": "text",
						"text": "What is in this image?",
					},
					map[string]any{
						"type": "image",
						"source": map[string]any{
							"type":       "base64",
							"media_type": "image/jpeg",
							"data":       "aGVsbG8=",
						},
					},
				},
			},
		},
	}

	chatReq, err := AnthropicToOpenAIChatRequest(req)
	if err != nil {
		t.Fatalf("AnthropicToOpenAIChatRequest failed: %v", err)
	}
	if len(chatReq.Messages) != 2 {
		t.Fatalf("Expected 2 messages (system + user), got %d", len(chatReq.Messages))
	}
	if chatReq.Messages[0].Content != "System instructions in map" {
		t.Errorf("Expected system text, got %v", chatReq.Messages[0].Content)
	}

	parts, ok := chatReq.Messages[1].Content.([]any)
	if !ok || len(parts) != 2 {
		t.Fatalf("Expected 2 parts in user message, got %v", chatReq.Messages[1].Content)
	}
}

func TestAnthropicConversionRejectsDroppedContent(t *testing.T) {
	tests := []struct {
		name string
		req  models.AnthropicMessagesRequest
		want string
	}{
		{
			name: "unsupported content block",
			req: models.AnthropicMessagesRequest{Messages: []models.AnthropicMessage{{
				Role:    "user",
				Content: []any{map[string]any{"type": "document", "source": map[string]any{}}},
			}}},
			want: "unsupported content block type",
		},
		{
			name: "missing text",
			req: models.AnthropicMessagesRequest{Messages: []models.AnthropicMessage{{
				Role:    "user",
				Content: []any{map[string]any{"type": "text"}},
			}}},
			want: `missing "text"`,
		},
		{
			name: "missing tool result ID",
			req: models.AnthropicMessagesRequest{Messages: []models.AnthropicMessage{{
				Role:    "user",
				Content: []any{map[string]any{"type": "tool_result", "content": "done"}},
			}}},
			want: `missing "tool_use_id"`,
		},
		{
			name: "invalid image data",
			req: models.AnthropicMessagesRequest{Messages: []models.AnthropicMessage{{
				Role: "user",
				Content: []any{map[string]any{
					"type": "image",
					"source": map[string]any{
						"type":       "base64",
						"media_type": "image/png",
						"data":       "not-base64",
					},
				}},
			}}},
			want: "base64 is invalid",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := AnthropicToOpenAIChatRequest(test.req)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestAnthropicConversionPreservesToolResultOrderAndFields(t *testing.T) {
	req := models.AnthropicMessagesRequest{Messages: []models.AnthropicMessage{
		{
			Role: "assistant",
			Content: []any{map[string]any{
				"type":  "tool_use",
				"id":    "toolu_lookup",
				"name":  "lookup",
				"input": map[string]any{"query": "भारत"},
			}},
		},
		{
			Role: "user",
			Content: []any{
				map[string]any{"type": "text", "text": "Before result"},
				map[string]any{"type": "tool_result", "tool_use_id": "toolu_lookup", "content": "PASS"},
				map[string]any{"type": "text", "text": "After result"},
			},
		},
	}}

	chatReq, err := AnthropicToOpenAIChatRequest(req)
	if err != nil {
		t.Fatalf("AnthropicToOpenAIChatRequest failed: %v", err)
	}
	if len(chatReq.Messages) != 4 {
		t.Fatalf("expected assistant, user, tool, user messages, got %#v", chatReq.Messages)
	}
	if chatReq.Messages[0].Role != "assistant" || len(chatReq.Messages[0].ToolCalls) != 1 {
		t.Fatalf("assistant tool call was not preserved: %#v", chatReq.Messages[0])
	}
	if chatReq.Messages[0].ToolCalls[0].ID != "toolu_lookup" || chatReq.Messages[0].ToolCalls[0].Function.Name != "lookup" {
		t.Fatalf("assistant tool identity was not preserved: %#v", chatReq.Messages[0].ToolCalls[0])
	}
	if chatReq.Messages[1].Role != "user" || chatReq.Messages[2].Role != "tool" || chatReq.Messages[3].Role != "user" {
		t.Fatalf("tool result changed message order: %#v", chatReq.Messages)
	}
	if chatReq.Messages[2].ToolCallID != "toolu_lookup" || chatReq.Messages[2].Content != "PASS" {
		t.Fatalf("tool result fields were not preserved: %#v", chatReq.Messages[2])
	}
	if err := ValidateToolResultReferences(chatReq.Messages); err != nil {
		t.Fatalf("converted Anthropic continuation failed validation: %v", err)
	}
}

func TestAnthropicConversionValidatesTools(t *testing.T) {
	_, err := AnthropicToOpenAIChatRequest(models.AnthropicMessagesRequest{
		Messages: []models.AnthropicMessage{{Role: "user", Content: "use tools"}},
		Tools:    []models.AnthropicTool{{Name: "", InputSchema: map[string]any{"type": "object"}}},
	})
	if err == nil || !strings.Contains(err.Error(), "name is empty") {
		t.Fatalf("error = %v, want empty tool name validation", err)
	}
}

func TestAnthropicConversionNormalizesToolChoice(t *testing.T) {
	tool := models.AnthropicTool{
		Name:        "lookup_user",
		Description: "Look up a user.",
		InputSchema: map[string]any{"type": "object"},
	}
	base := models.AnthropicMessagesRequest{
		Messages: []models.AnthropicMessage{{Role: "user", Content: "find the user"}},
		Tools:    []models.AnthropicTool{tool},
	}

	base.ToolChoice = map[string]any{"type": "any", "disable_parallel_tool_use": false}
	chatReq, err := AnthropicToOpenAIChatRequest(base)
	if err != nil || chatReq.ToolChoice != "required" {
		t.Fatalf("any tool choice = %#v, error = %v; want required", chatReq.ToolChoice, err)
	}

	base.ToolChoice = map[string]any{
		"type": "tool",
		"name": "lookup_user",
	}
	chatReq, err = AnthropicToOpenAIChatRequest(base)
	if err != nil {
		t.Fatalf("named tool choice error = %v", err)
	}
	choice, ok := chatReq.ToolChoice.(map[string]any)
	if !ok {
		t.Fatalf("named tool choice type = %T, want object", chatReq.ToolChoice)
	}
	function, ok := choice["function"].(map[string]any)
	if !ok || function["name"] != "lookup_user" {
		t.Fatalf("named tool choice = %#v", chatReq.ToolChoice)
	}

	invalid := []struct {
		name   string
		choice any
		want   string
	}{
		{name: "unknown type", choice: map[string]any{"type": "server_tool"}, want: "unsupported Anthropic tool_choice type"},
		{name: "undeclared tool", choice: map[string]any{"type": "tool", "name": "delete_all"}, want: "undeclared tool"},
		{name: "parallel disabled", choice: map[string]any{"type": "any", "disable_parallel_tool_use": true}, want: "disable_parallel_tool_use"},
		{name: "wrong parallel flag", choice: map[string]any{"type": "any", "disable_parallel_tool_use": "yes"}, want: "must be a boolean"},
	}
	for _, test := range invalid {
		t.Run(test.name, func(t *testing.T) {
			base.ToolChoice = test.choice
			if _, err := AnthropicToOpenAIChatRequest(base); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want substring %q", err, test.want)
			}
		})
	}
}
