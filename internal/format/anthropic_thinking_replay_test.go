package format

import (
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
