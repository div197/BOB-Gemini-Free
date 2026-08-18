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

	chatReq := AnthropicToOpenAIChatRequest(req)
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
	chatReq := AnthropicToOpenAIChatRequest(req)
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

	chatReq := AnthropicToOpenAIChatRequest(req)
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
