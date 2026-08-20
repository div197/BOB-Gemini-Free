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

	// Test developer role and response_format
	reqDev := models.OpenAIChatRequest{
		Messages: []models.OpenAIMessage{
			{Role: "developer", Content: "Formatting developer rule."},
			{Role: "user", Content: "Return JSON."},
		},
		ResponseFormat: &models.OpenAIResponseFormat{Type: "json_object"},
	}
	promptDev, errDev := MessagesToPrompt(reqDev)
	if errDev != nil {
		t.Fatalf("Unexpected error: %v", errDev)
	}
	if !strings.Contains(promptDev, "[System instruction]: Formatting developer rule.") {
		t.Errorf("Prompt missing developer instruction: %q", promptDev)
	}
	if !strings.Contains(promptDev, "You must respond strictly with valid JSON output.") {
		t.Errorf("Prompt missing JSON instruction: %q", promptDev)
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

func TestResponsesInputToMessagesPreservesInputImage(t *testing.T) {
	input := []any{
		map[string]any{
			"role": "user",
			"content": []any{
				map[string]any{"type": "input_text", "text": "Describe this."},
				map[string]any{"type": "input_image", "image_url": "https://example.com/image.png"},
			},
		},
	}

	messages, err := ResponsesInputToMessages(input, "")
	if err != nil {
		t.Fatalf("ResponsesInputToMessages error: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	content, ok := messages[0]["content"].([]any)
	if !ok {
		t.Fatalf("Expected multipart content, got %#v", messages[0]["content"])
	}
	if len(content) != 2 {
		t.Fatalf("Expected 2 content parts, got %d", len(content))
	}

	chatReq := models.OpenAIChatRequest{
		Messages: []models.OpenAIMessage{{Role: "user", Content: content}},
	}
	prompt, images, err := MessagesToPromptAndImages(chatReq)
	if err != nil {
		t.Fatalf("MessagesToPromptAndImages failed: %v", err)
	}
	if !strings.Contains(prompt, "Describe this.") {
		t.Fatalf("Expected prompt text to be preserved, got %q", prompt)
	}
	if len(images) != 1 || images[0].URL != "https://example.com/image.png" {
		t.Fatalf("Expected remote input image to be preserved, got %#v", images)
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

func TestMessagesToPromptAndImagesRemoteURL(t *testing.T) {
	req := models.OpenAIChatRequest{
		Messages: []models.OpenAIMessage{
			{
				Role: "user",
				Content: []any{
					map[string]any{
						"type": "image_url",
						"image_url": map[string]any{
							"url": "https://example.com/diagram.png",
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
	if strings.TrimSpace(prompt) != "" {
		t.Errorf("expected empty prompt for image-only request, got %q", prompt)
	}
	if len(images) != 1 {
		t.Fatalf("expected 1 remote image, got %d", len(images))
	}
	if images[0].URL != "https://example.com/diagram.png" {
		t.Errorf("expected remote image URL to be preserved, got %q", images[0].URL)
	}
	if len(images[0].Data) != 0 {
		t.Errorf("expected remote image data to be fetched later, got %d bytes", len(images[0].Data))
	}
}

func TestExtractThinking(t *testing.T) {
	raw := "```thought\n1. Analyze input\n2. Compute step\n```\nHere is the final answer."
	thinking, clean := ExtractThinking(raw)
	expectedThinking := "1. Analyze input\n2. Compute step"
	expectedClean := "Here is the final answer."

	if thinking != expectedThinking {
		t.Errorf("Got thinking %q, want %q", thinking, expectedThinking)
	}
	if clean != expectedClean {
		t.Errorf("Got clean text %q, want %q", clean, expectedClean)
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

func TestBuildToolChoiceInstruction(t *testing.T) {
	if BuildToolChoiceInstruction("none") != "\n\nIMPORTANT: Do NOT call any tools. Respond with text only." {
		t.Errorf("Unexpected tool choice instruction for 'none'")
	}
	if BuildToolChoiceInstruction("required") != "\n\nIMPORTANT: You MUST call at least one tool. Do not respond with text only." {
		t.Errorf("Unexpected tool choice instruction for 'required'")
	}
	specificChoice := map[string]any{
		"type": "function",
		"function": map[string]any{
			"name": "lookup_user",
		},
	}
	if !strings.Contains(BuildToolChoiceInstruction(specificChoice), "lookup_user") {
		t.Errorf("Expected specific choice to mention 'lookup_user'")
	}
}

func TestBuildResponseOutput(t *testing.T) {
	toolCalls := []models.OpenAIToolCall{
		{
			ID:   "call_123",
			Type: "function",
			Function: models.OpenAIToolCallFunction{
				Name:      "calc",
				Arguments: `{"expr":"2+2"}`,
			},
		},
	}
	output := BuildResponseOutput("Done calculating", toolCalls, "msg_abc")
	if len(output) != 2 {
		t.Fatalf("Expected 2 output items (1 tool call + 1 message), got %d", len(output))
	}
	if output[0]["type"] != "function_call" || output[0]["name"] != "calc" {
		t.Errorf("Unexpected function call item: %v", output[0])
	}
	if output[1]["type"] != "message" || output[1]["id"] != "msg_abc" {
		t.Errorf("Unexpected message item: %v", output[1])
	}
}

func TestResponsesInputString(t *testing.T) {
	msgs, err := ResponsesInputToMessages("Simple query string", "Be concise")
	if err != nil {
		t.Fatalf("ResponsesInputToMessages failed: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("Expected 2 messages (system + user), got %d", len(msgs))
	}
	if msgs[0]["role"] != "system" || msgs[1]["role"] != "user" {
		t.Errorf("Unexpected roles: %v", msgs)
	}
}
