package format

import (
	"fmt"
	"strings"
	"testing"

	"github.com/div197/bob-gemini-free/internal/models"
)

// Clients that enable extended thinking echo the previous assistant turn,
// including its thinking blocks, on the next request. Those blocks must not
// make the whole request fail, and their reasoning must not be replayed.
func TestAnthropicEchoedThinkingBlocksAreDroppedNotRejected(t *testing.T) {
	req := models.AnthropicMessagesRequest{
		Model: "claude-3-7-sonnet",
		Messages: []models.AnthropicMessage{
			{Role: "user", Content: "What is 2+2?"},
			{
				Role: "assistant",
				Content: []any{
					map[string]any{"type": "thinking", "thinking": "private reasoning", "signature": "sig"},
					map[string]any{"type": "redacted_thinking", "data": "opaque"},
					map[string]any{"type": "text", "text": "4"},
				},
			},
			{Role: "user", Content: "And 3+3?"},
		},
	}

	chatReq, err := AnthropicToOpenAIChatRequest(req)
	if err != nil {
		t.Fatalf("echoed thinking blocks must not fail the request: %v", err)
	}
	prompt, _, err := MessagesToPromptAndImages(chatReq)
	if err != nil {
		t.Fatalf("MessagesToPromptAndImages failed: %v", err)
	}
	if !strings.Contains(prompt, "4") {
		t.Fatalf("visible assistant text was dropped: %q", prompt)
	}
	if strings.Contains(prompt, "private reasoning") || strings.Contains(prompt, "opaque") {
		t.Fatalf("prior reasoning leaked into the upstream prompt: %q", prompt)
	}
}

// Extended-thinking tool loops echo [thinking, tool_use] from the assistant,
// then send the matching tool_result. The dropped thinking block must not break
// tool-result reference validation or leave a dangling tool call.
func TestAnthropicThinkingToolUseRoundTripStaysValid(t *testing.T) {
	req := models.AnthropicMessagesRequest{
		Model: "claude-3-7-sonnet",
		Messages: []models.AnthropicMessage{
			{Role: "user", Content: "Weather in Jodhpur?"},
			{
				Role: "assistant",
				Content: []any{
					map[string]any{"type": "thinking", "thinking": "need the tool", "signature": "sig"},
					map[string]any{
						"type":  "tool_use",
						"id":    "toolu_1",
						"name":  "get_weather",
						"input": map[string]any{"city": "Jodhpur"},
					},
				},
			},
			{
				Role: "user",
				Content: []any{
					map[string]any{"type": "tool_result", "tool_use_id": "toolu_1", "content": "sunny"},
				},
			},
		},
	}

	chatReq, err := AnthropicToOpenAIChatRequest(req)
	if err != nil {
		t.Fatalf("thinking+tool_use round trip failed: %v", err)
	}
	if err := ValidateToolResultReferences(chatReq.Messages); err != nil {
		t.Fatalf("tool result no longer references its tool call: %v", err)
	}
	if strings.Contains(fmt.Sprint(chatReq.Messages), "need the tool") {
		t.Fatal("prior reasoning leaked into the converted messages")
	}
}

// A thinking-only assistant turn has no visible text or tool call. It must be
// dropped quietly rather than becoming an empty assistant message or an error.
func TestAnthropicThinkingOnlyAssistantTurnIsNotRejected(t *testing.T) {
	req := models.AnthropicMessagesRequest{
		Model: "claude-3-7-sonnet",
		Messages: []models.AnthropicMessage{
			{Role: "user", Content: "Hello"},
			{
				Role: "assistant",
				Content: []any{
					map[string]any{"type": "thinking", "thinking": "just thinking", "signature": "sig"},
				},
			},
			{Role: "user", Content: "Continue"},
		},
	}

	chatReq, err := AnthropicToOpenAIChatRequest(req)
	if err != nil {
		t.Fatalf("thinking-only assistant turn must not fail: %v", err)
	}
	for _, msg := range chatReq.Messages {
		if msg.Role == "assistant" {
			t.Fatalf("thinking-only turn produced an assistant message: %+v", msg)
		}
	}
}
