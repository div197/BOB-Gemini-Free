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
