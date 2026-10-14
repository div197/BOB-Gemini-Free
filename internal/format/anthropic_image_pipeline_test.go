package format

import (
	"strings"
	"testing"

	"github.com/div197/bob-gemini-free/internal/models"
)

// Anthropic image blocks must survive the full conversion into the prompt
// pipeline. The converter builds image_url as map[string]string while the
// OpenAI extractor only reads map[string]any, so a valid base64 image used to
// fail with "has no image URL" before reaching Gemini.
func TestAnthropicBase64ImageReachesPromptPipeline(t *testing.T) {
	req := models.AnthropicMessagesRequest{
		Model: "claude-3-7-sonnet",
		Messages: []models.AnthropicMessage{
			{
				Role: "user",
				Content: []any{
					map[string]any{"type": "text", "text": "Describe this image"},
					map[string]any{
						"type": "image",
						"source": map[string]any{
							"type":       "base64",
							"media_type": "image/png",
							"data":       "iVBORw0KGgo=",
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
	prompt, images, err := MessagesToPromptAndImages(chatReq)
	if err != nil {
		t.Fatalf("MessagesToPromptAndImages rejected a valid Anthropic image: %v", err)
	}
	if !strings.Contains(prompt, "Describe this image") {
		t.Fatalf("prompt lost the text part: %q", prompt)
	}
	if len(images) != 1 {
		t.Fatalf("expected 1 extracted image, got %d", len(images))
	}
}
