package format

import (
	"strings"
	"testing"

	"github.com/div197/bob-gemini-free/internal/models"
)

func TestParseToolCalls(t *testing.T) {
	input := "Here is a call:\n```tool_call\n{\"name\": \"get_weather\", \"arguments\": {\"city\": \"Jakarta\"}}\n```\nDone."
	clean, calls := ParseToolCalls(input)

	if len(calls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(calls))
	}

	if calls[0].Function.Name != "get_weather" {
		t.Errorf("Expected name get_weather, got %s", calls[0].Function.Name)
	}

	if !strings.Contains(clean, "Here is a call:") || !strings.Contains(clean, "Done.") {
		t.Errorf("Unexpected clean text: %q", clean)
	}
}

func TestMessagesToPrompt(t *testing.T) {
	req := models.OpenAIChatRequest{
		Messages: []models.OpenAIMessage{
			{Role: "system", Content: "Be helpful."},
			{Role: "user", Content: "Hello!"},
		},
		ToolChoice: "auto",
	}

	prompt, err := MessagesToPrompt(req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(prompt, "[System instruction]: Be helpful.") {
		t.Errorf("Prompt missing system instruction: %q", prompt)
	}
	if !strings.Contains(prompt, "Hello!") {
		t.Errorf("Prompt missing user message: %q", prompt)
	}
}

func TestResponsesInputToMessagesMultiPartText(t *testing.T) {
	input := []any{
		map[string]any{
			"role": "user",
			"content": []any{
				map[string]any{"type": "text", "text": "Hello"},
				map[string]any{"type": "text", "text": "world"},
			},
		},
	}

	messages, err := ResponsesInputToMessages(input, "System instructions")
	if err != nil {
		t.Fatalf("ResponsesInputToMessages error: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages (system + user), got %d", len(messages))
	}

	userContent, _ := messages[1]["content"].(string)
	expected := "Hello world"
	if userContent != expected {
		t.Errorf("Got content %q, want %q", userContent, expected)
	}
}

func TestMessagesToPromptAndImages(t *testing.T) {
	fakeBase64 := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
	req := models.OpenAIChatRequest{
		Messages: []models.OpenAIMessage{
			{
				Role: "user",
				Content: []any{
					map[string]any{"type": "text", "text": "Describe this image:"},
					map[string]any{
						"type": "image_url",
						"image_url": map[string]any{
							"url": "data:image/png;base64," + fakeBase64,
						},
					},
				},
			},
		},
	}

	prompt, images, err := MessagesToPromptAndImages(req)
	if err != nil {
		t.Fatalf("MessagesToPromptAndImages failed: %v", err)
	}

	if !strings.Contains(prompt, "Describe this image:") {
		t.Errorf("Expected prompt to contain text, got %q", prompt)
	}

	if len(images) != 1 {
		t.Fatalf("Expected 1 image extracted, got %d", len(images))
	}

	if images[0].MIME != "image/png" {
		t.Errorf("Expected image MIME image/png, got %s", images[0].MIME)
	}

	if len(images[0].Data) == 0 {
		t.Errorf("Expected non-empty image bytes")
	}
}

func TestExtractThinking(t *testing.T) {
	raw := "```thought\n1. Analyze input\n2. Compute step\n```\nHere is the final answer."
	thinking, clean := ExtractThinking(raw)
	expectedThinking := "1. Analyze input\n2. Compute step"
	expectedClean := "Here is the final answer."

	if thinking != expectedThinking {
		t.Errorf("got thinking %q, want %q", thinking, expectedThinking)
	}
	if clean != expectedClean {
		t.Errorf("got clean %q, want %q", clean, expectedClean)
	}

	// Plain text without thought blocks
	plain := "Just a direct response"
	noThinking, plainClean := ExtractThinking(plain)
	if noThinking != "" {
		t.Errorf("expected empty thinking, got %q", noThinking)
	}
	if plainClean != plain {
		t.Errorf("expected plain text unchanged, got %q", plainClean)
	}
}
