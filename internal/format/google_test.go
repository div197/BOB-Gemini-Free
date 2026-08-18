package format

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/div197/bob-gemini-free/internal/models"
)

func TestGoogleFormatting(t *testing.T) {
	// 1. BuildToolPrompt
	tools := []models.GoogleFunctionDeclaration{
		{
			Name:        "get_weather",
			Description: "Get the current weather",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]string{"type": "string"},
				},
			},
		},
	}
	toolPrompt := BuildToolPrompt(tools)
	if !strings.Contains(toolPrompt, "get_weather") {
		t.Errorf("Expected tool prompt to contain 'get_weather', got: %s", toolPrompt)
	}

	// 2. GoogleToolChoiceInstruction
	reqAuto := models.GoogleGenerateRequest{}
	if GoogleToolChoiceInstruction(reqAuto) != "" {
		t.Errorf("Expected empty instruction for AUTO mode")
	}

	reqNone := models.GoogleGenerateRequest{
		ToolConfig: &models.GoogleToolConfig{
			FunctionCallingConfig: &models.GoogleFunctionCallingConfig{
				Mode: "NONE",
			},
		},
	}
	if !strings.Contains(GoogleToolChoiceInstruction(reqNone), "Do NOT call any tools") {
		t.Errorf("Expected NONE mode instruction")
	}

	reqAny := models.GoogleGenerateRequest{
		ToolConfig: &models.GoogleToolConfig{
			FunctionCallingConfig: &models.GoogleFunctionCallingConfig{
				Mode:                 "ANY",
				AllowedFunctionNames: []string{"search_web"},
			},
		},
	}
	if !strings.Contains(GoogleToolChoiceInstruction(reqAny), "search_web") {
		t.Errorf("Expected ANY mode instruction with search_web")
	}

	// 3. GoogleContentsToPrompt with Text, SystemInstruction, and Images
	b64Sample := base64.StdEncoding.EncodeToString([]byte("fake-png-data"))
	reqFull := models.GoogleGenerateRequest{
		SystemInstruction: &models.GoogleContent{
			Parts: []models.GooglePart{
				{Text: "You are a helpful assistant."},
			},
		},
		Tools: []models.GoogleTool{
			{FunctionDeclarations: tools},
		},
		Contents: []models.GoogleContent{
			{
				Role: "user",
				Parts: []models.GooglePart{
					{Text: "What's the weather in Tokyo?"},
					{
						InlineData: &models.GoogleInlineData{
							MIMEType: "image/png",
							Data:     b64Sample,
						},
					},
				},
			},
			{
				Role: "model",
				Parts: []models.GooglePart{
					{
						FunctionCall: &models.GoogleFunctionCall{
							Name: "get_weather",
							Args: map[string]any{"location": "Tokyo"},
						},
					},
				},
			},
			{
				Role: "function",
				Parts: []models.GooglePart{
					{
						FunctionResponse: &models.GoogleFunctionCallResp{
							Name: "get_weather",
							Response: map[string]any{
								"result": "22C Sunny",
							},
						},
					},
				},
			},
		},
	}

	prompt, imgs, err := GoogleContentsToPrompt(reqFull)
	if err != nil {
		t.Fatalf("GoogleContentsToPrompt failed: %v", err)
	}
	if len(imgs) != 1 {
		t.Errorf("Expected 1 image, got %d", len(imgs))
	}
	if !strings.Contains(prompt, "You are a helpful assistant.") {
		t.Errorf("Missing system instruction in prompt")
	}
	if !strings.Contains(prompt, "Tokyo") {
		t.Errorf("Missing user query in prompt")
	}
	if !strings.Contains(prompt, "22C Sunny") {
		t.Errorf("Missing function response in prompt")
	}

	// 4. ParseGoogleFunctionCalls
	rawModelOutput := "Here is the tool call:\n```function_call\n{\"name\": \"get_weather\", \"args\": {\"location\": \"Tokyo\"}}\n```"
	cleanText, calls := ParseGoogleFunctionCalls(rawModelOutput)
	if len(calls) != 1 {
		t.Fatalf("Expected 1 parsed function call, got %d", len(calls))
	}
	if calls[0].Name != "get_weather" {
		t.Errorf("Expected name get_weather, got %s", calls[0].Name)
	}
	if strings.Contains(cleanText, "```function_call") {
		t.Errorf("Expected clean text to have function_call block removed, got: %s", cleanText)
	}
}
