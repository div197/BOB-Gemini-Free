package format

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/div197/bob-gemini-free/internal/models"
)

func TestRandHexBoundsInvalidLengths(t *testing.T) {
	for _, length := range []int{-1, 0, 33, 1000000} {
		got := RandHex(length)
		if len(got) > 32 {
			t.Fatalf("RandHex(%d) returned %d characters", length, len(got))
		}
	}
	if got := RandHex(16); len(got) != 16 {
		t.Fatalf("RandHex(16) length = %d, want 16", len(got))
	}
}

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

func TestValidateToolResultReferences(t *testing.T) {
	valid := []models.OpenAIMessage{
		{Role: "assistant", ToolCalls: []models.OpenAIToolCall{{ID: "call_1", Function: models.OpenAIToolCallFunction{Name: "lookup"}}}},
		{Role: "tool", ToolCallID: "call_1", Name: "lookup", Content: `{"ok":true}`},
	}
	if err := ValidateToolResultReferences(valid); err != nil {
		t.Fatalf("valid tool continuation rejected: %v", err)
	}
	tests := []struct {
		name string
		msgs []models.OpenAIMessage
		want string
	}{
		{
			name: "unknown id",
			msgs: []models.OpenAIMessage{{Role: "tool", ToolCallID: "missing", Content: "result"}},
			want: "unknown tool_call_id",
		},
		{
			name: "mismatched name",
			msgs: []models.OpenAIMessage{
				{Role: "assistant", ToolCalls: []models.OpenAIToolCall{{ID: "call_1", Function: models.OpenAIToolCallFunction{Name: "lookup"}}}},
				{Role: "tool", ToolCallID: "call_1", Name: "delete", Content: "result"},
			},
			want: "does not match",
		},
		{
			name: "ambiguous name",
			msgs: []models.OpenAIMessage{
				{Role: "assistant", ToolCalls: []models.OpenAIToolCall{
					{ID: "call_1", Function: models.OpenAIToolCallFunction{Name: "lookup"}},
					{ID: "call_2", Function: models.OpenAIToolCallFunction{Name: "lookup"}},
				}},
				{Role: "tool", Name: "lookup", Content: "result"},
			},
			want: "ambiguous",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := ValidateToolResultReferences(test.msgs); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want substring %q", err, test.want)
			}
		})
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

func TestMessagesToPromptAndImagesRejectsInvalidInlineImages(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{name: "invalid base64", url: "data:image/png;base64,not-base64", want: "base64 is invalid"},
		{name: "missing base64 marker", url: "data:image/png,hello", want: "must use base64"},
		{name: "non-image MIME", url: "data:text/plain;base64,aGk=", want: "not an image"},
		{name: "missing payload", url: "data:image/png;base64", want: "missing its payload"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := MessagesToPromptAndImages(models.OpenAIChatRequest{Messages: []models.OpenAIMessage{{
				Role:    "user",
				Content: []any{map[string]any{"type": "image_url", "image_url": map[string]any{"url": test.url}}},
			}}})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
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

func TestMessagesToPromptAndImagesRejectsDroppedContent(t *testing.T) {
	tests := []struct {
		name string
		req  models.OpenAIChatRequest
		want string
	}{
		{
			name: "scalar message content",
			req:  models.OpenAIChatRequest{Messages: []models.OpenAIMessage{{Role: "user", Content: 42}}},
			want: "message content must be a string",
		},
		{
			name: "missing text field",
			req: models.OpenAIChatRequest{Messages: []models.OpenAIMessage{{
				Role: "user", Content: []any{map[string]any{"type": "text"}},
			}}},
			want: "text is missing",
		},
		{
			name: "wrong text field type",
			req: models.OpenAIChatRequest{Messages: []models.OpenAIMessage{{
				Role: "user", Content: []any{map[string]any{"type": "text", "text": 42}},
			}}},
			want: "text must be a string",
		},
		{
			name: "unknown role",
			req:  models.OpenAIChatRequest{Messages: []models.OpenAIMessage{{Role: "reviewer", Content: "ignored"}}},
			want: "unsupported OpenAI message role",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, _, err := MessagesToPromptAndImages(test.req); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want substring %q", err, test.want)
			}
		})
	}

	prompt, _, err := MessagesToPromptAndImages(models.OpenAIChatRequest{
		Messages: []models.OpenAIMessage{{Role: " USER ", Content: "normalized role"}},
	})
	if err != nil || !strings.Contains(prompt, "normalized role") {
		t.Fatalf("normalized valid role prompt = %q, error = %v", prompt, err)
	}
}

func TestMessagesToPromptEscapesToolCallNamesAsJSON(t *testing.T) {
	name := "lookup\"\\\nuser"
	prompt, err := MessagesToPrompt(models.OpenAIChatRequest{Messages: []models.OpenAIMessage{{
		Role: "assistant",
		ToolCalls: []models.OpenAIToolCall{{
			Function: models.OpenAIToolCallFunction{Name: name, Arguments: `{"ok":true}`},
		}},
	}}})
	if err != nil {
		t.Fatalf("tool-call prompt error = %v", err)
	}
	start := strings.Index(prompt, "```tool_call\n")
	if start < 0 {
		t.Fatalf("tool-call fence missing from prompt: %q", prompt)
	}
	start += len("```tool_call\n")
	end := strings.Index(prompt[start:], "\n```")
	if end < 0 {
		t.Fatalf("tool-call fence was not closed: %q", prompt[start:])
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(prompt[start:start+end]), &payload); err != nil {
		t.Fatalf("tool-call payload is not valid JSON: %v; payload=%q", err, prompt[start:start+end])
	}
	if payload["name"] != name {
		t.Fatalf("decoded tool name = %#v, want %q", payload["name"], name)
	}
	if !strings.Contains(BuildToolChoiceInstruction(map[string]any{
		"function": map[string]any{"name": name},
	}), `\\`) {
		t.Fatalf("specific tool instruction did not JSON-escape the name")
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
	if !strings.Contains(BuildToolChoiceInstruction(" ANY "), "MUST call") {
		t.Errorf("Expected normalized 'any' choice to require a tool")
	}
	if !strings.Contains(BuildToolChoiceInstruction(" NONE "), "Do NOT call") {
		t.Errorf("Expected normalized 'none' choice to forbid tools")
	}
	if !IsToolChoiceNone(" NONE ") || IsToolChoiceNone("auto") || IsToolChoiceNone(map[string]any{"type": "none"}) {
		t.Errorf("IsToolChoiceNone did not preserve canonical string-only behavior")
	}

	prompt, _, err := MessagesToPromptAndImages(models.OpenAIChatRequest{
		Messages:   []models.OpenAIMessage{{Role: "user", Content: "do not use tools"}},
		Tools:      []models.OpenAITool{{Type: "function", Function: models.OpenAIFunction{Name: "lookup"}}},
		ToolChoice: " NONE ",
	})
	if err != nil || strings.Contains(prompt, "# Tool Use") {
		t.Fatalf("normalized none choice prompt = %q, error = %v", prompt, err)
	}
}

func TestValidateToolChoiceRejectsUnsupportedOrUndeclaredChoices(t *testing.T) {
	tools := []models.OpenAITool{{
		Type:     "function",
		Function: models.OpenAIFunction{Name: "lookup_user"},
	}}
	valid := []any{
		nil,
		"",
		"auto",
		"none",
		"required",
		"any",
		map[string]any{
			"type":     "function",
			"function": map[string]any{"name": "lookup_user"},
		},
		map[string]any{
			"function": map[string]any{"name": "lookup_user"},
		},
	}
	for _, choice := range valid {
		if err := ValidateToolChoice(choice, tools); err != nil {
			t.Errorf("valid tool choice %#v rejected: %v", choice, err)
		}
	}

	invalid := []struct {
		name   string
		choice any
		want   string
	}{
		{name: "unknown mode", choice: "maybe", want: "unsupported tool_choice"},
		{name: "wrong type", choice: 1, want: "unsupported tool_choice type"},
		{name: "missing function", choice: map[string]any{"type": "function"}, want: "function is missing"},
		{name: "wrong function shape", choice: map[string]any{"type": "function", "function": "lookup_user"}, want: "function must be an object"},
		{name: "missing function name", choice: map[string]any{"type": "function", "function": map[string]any{}}, want: "function name is missing"},
		{name: "undeclared function", choice: map[string]any{"type": "function", "function": map[string]any{"name": "delete_all"}}, want: "undeclared tool"},
		{name: "unsupported type", choice: map[string]any{"type": "computer", "function": map[string]any{"name": "lookup_user"}}, want: "unsupported tool_choice type"},
	}
	for _, test := range invalid {
		t.Run(test.name, func(t *testing.T) {
			if err := ValidateToolChoice(test.choice, tools); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want substring %q", err, test.want)
			}
		})
	}

	if _, _, err := MessagesToPromptAndImages(models.OpenAIChatRequest{
		Messages:   []models.OpenAIMessage{{Role: "user", Content: "use the requested tool"}},
		Tools:      tools,
		ToolChoice: map[string]any{"type": "function", "function": map[string]any{"name": "delete_all"}},
	}); err == nil || !strings.Contains(err.Error(), "undeclared tool") {
		t.Fatalf("shared prompt path error = %v, want undeclared tool failure", err)
	}
}

func TestToolPromptBudgetsRejectOversizedDeclarationsAndArguments(t *testing.T) {
	tooMany := make([]models.OpenAITool, MaxToolDefinitions+1)
	for i := range tooMany {
		tooMany[i] = models.OpenAITool{Type: "function", Function: models.OpenAIFunction{Name: "tool"}}
	}
	if _, err := MessagesToPrompt(models.OpenAIChatRequest{
		Messages: []models.OpenAIMessage{{Role: "user", Content: "use tools"}},
		Tools:    tooMany,
	}); err == nil {
		t.Fatal("oversized tool definition count was accepted")
	}

	if _, err := MessagesToPrompt(models.OpenAIChatRequest{
		Messages: []models.OpenAIMessage{{Role: "user", Content: "use tools"}},
		Tools: []models.OpenAITool{{Type: "function", Function: models.OpenAIFunction{
			Name:        "large_schema",
			Description: strings.Repeat("x", MaxToolDescriptionBytes+1),
		}}},
	}); err == nil {
		t.Fatal("oversized tool description was accepted")
	}

	largeArgs := strings.Repeat("x", MaxToolArgumentBytes+1)
	if _, err := MessagesToPrompt(models.OpenAIChatRequest{
		Messages: []models.OpenAIMessage{{Role: "assistant", ToolCalls: []models.OpenAIToolCall{{
			Function: models.OpenAIToolCallFunction{Name: "large_args", Arguments: largeArgs},
		}}}},
	}); err == nil {
		t.Fatal("oversized tool arguments were accepted")
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

func TestResponsesInputToMessagesPreservesToolContinuations(t *testing.T) {
	input := []any{
		map[string]any{
			"type":      "function_call",
			"call_id":   "call_weather",
			"name":      "get_weather",
			"arguments": `{"city":"Jodhpur","units":"metric"}`,
		},
		map[string]any{
			"type":    "function_call_output",
			"call_id": "call_weather",
			"output": map[string]any{
				"temperature": 31,
				"condition":   "clear",
			},
		},
	}

	messages, err := ResponsesInputToMessages(input, "")
	if err != nil {
		t.Fatalf("ResponsesInputToMessages error: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected assistant call and tool result, got %d messages", len(messages))
	}

	calls, ok := messages[0]["tool_calls"].([]map[string]any)
	if !ok || len(calls) != 1 {
		t.Fatalf("expected one preserved tool call, got %#v", messages[0]["tool_calls"])
	}
	if calls[0]["id"] != "call_weather" {
		t.Fatalf("tool call ID = %#v, want call_weather", calls[0]["id"])
	}
	function, ok := calls[0]["function"].(map[string]any)
	if !ok || function["name"] != "get_weather" || function["arguments"] != `{"city":"Jodhpur","units":"metric"}` {
		t.Fatalf("tool call function was not preserved: %#v", calls[0]["function"])
	}
	if messages[1]["tool_call_id"] != "call_weather" {
		t.Fatalf("tool result ID = %#v, want call_weather", messages[1]["tool_call_id"])
	}
	if _, ok := messages[1]["name"]; ok {
		t.Fatalf("tool result unexpectedly synthesized a name: %#v", messages[1]["name"])
	}
	result, ok := messages[1]["content"].(string)
	if !ok || !strings.Contains(result, `"temperature":31`) {
		t.Fatalf("tool result output was not encoded safely: %#v", messages[1]["content"])
	}
}

func TestResponsesInputToMessagesRejectsMalformedItems(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{
			name:  "unsupported item type",
			input: []any{map[string]any{"type": "reasoning", "summary": []any{}}},
			want:  "unsupported input item type",
		},
		{
			name:  "missing function call ID",
			input: []any{map[string]any{"type": "function_call", "name": "lookup", "arguments": `{}`}},
			want:  `missing "call_id"`,
		},
		{
			name:  "missing function call arguments",
			input: []any{map[string]any{"type": "function_call", "call_id": "call_1", "name": "lookup"}},
			want:  `missing "arguments"`,
		},
		{
			name:  "missing tool output",
			input: []any{map[string]any{"type": "function_call_output", "call_id": "call_1"}},
			want:  `missing "output"`,
		},
		{
			name: "missing input text",
			input: []any{map[string]any{
				"type":    "message",
				"role":    "user",
				"content": []any{map[string]any{"type": "input_text"}},
			}},
			want: `missing "text"`,
		},
		{
			name:  "unsupported scalar",
			input: []any{42},
			want:  "unsupported item type",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ResponsesInputToMessages(test.input, "")
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestResponsesInputToMessagesRejectsInvalidFunctionArguments(t *testing.T) {
	_, err := ResponsesInputToMessages([]any{map[string]any{
		"type":      "function_call",
		"call_id":   "call_1",
		"name":      "lookup",
		"arguments": "not-json",
	}}, "")
	if err == nil || !strings.Contains(err.Error(), "invalid JSON arguments") {
		t.Fatalf("error = %v, want invalid JSON arguments", err)
	}
}
