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
		Tools: []models.GoogleTool{{FunctionDeclarations: []models.GoogleFunctionDeclaration{{Name: "search_web"}}}},
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

func TestGoogleContentsToPromptRejectsInvalidInlineImages(t *testing.T) {
	tests := []struct {
		name string
		part models.GooglePart
		want string
	}{
		{
			name: "invalid base64",
			part: models.GooglePart{InlineData: &models.GoogleInlineData{MIMEType: "image/png", Data: "not-base64"}},
			want: "base64 is invalid",
		},
		{
			name: "non-image MIME",
			part: models.GooglePart{InlineData: &models.GoogleInlineData{MIMEType: "text/plain", Data: "aGk="}},
			want: "not an image",
		},
		{
			name: "oversized encoded input",
			part: models.GooglePart{InlineData: &models.GoogleInlineData{MIMEType: "image/png", Data: strings.Repeat("A", 28<<20)}},
			want: "exceeds",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := GoogleContentsToPrompt(models.GoogleGenerateRequest{Contents: []models.GoogleContent{{Parts: []models.GooglePart{test.part}}}})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestGoogleFunctionCallingModeFailsClosedAndNormalizes(t *testing.T) {
	declaration := models.GoogleFunctionDeclaration{Name: "lookup", Parameters: map[string]any{"type": "object"}}
	base := models.GoogleGenerateRequest{
		Contents: []models.GoogleContent{{Role: "user", Parts: []models.GooglePart{{Text: "lookup"}}}},
		Tools:    []models.GoogleTool{{FunctionDeclarations: []models.GoogleFunctionDeclaration{declaration}}},
	}

	base.ToolConfig = &models.GoogleToolConfig{FunctionCallingConfig: &models.GoogleFunctionCallingConfig{Mode: " none "}}
	mode, err := GoogleFunctionCallingMode(base)
	if err != nil || mode != "NONE" {
		t.Fatalf("normalized mode = %q, error = %v; want NONE", mode, err)
	}
	prompt, _, err := GoogleContentsToPrompt(base)
	if err != nil || strings.Contains(prompt, "# Tool Use") {
		t.Fatalf("NONE prompt = %q, error = %v", prompt, err)
	}

	base.ToolConfig = &models.GoogleToolConfig{FunctionCallingConfig: &models.GoogleFunctionCallingConfig{
		Mode:                 "any",
		AllowedFunctionNames: []string{"lookup"},
	}}
	mode, err = GoogleFunctionCallingMode(base)
	if err != nil || mode != "ANY" {
		t.Fatalf("normalized ANY mode = %q, error = %v", mode, err)
	}
	if !strings.Contains(GoogleToolChoiceInstruction(base), "lookup") {
		t.Fatalf("ANY instruction lost allowed function: %q", GoogleToolChoiceInstruction(base))
	}

	invalid := []struct {
		name string
		mode string
		want string
	}{
		{name: "unknown mode", mode: "maybe", want: "unsupported Google function-calling mode"},
		{name: "undeclared name", mode: "ANY", want: "is not declared"},
		{name: "allowed name with auto", mode: "AUTO", want: "require Google function-calling mode ANY"},
	}
	for _, test := range invalid {
		t.Run(test.name, func(t *testing.T) {
			request := base
			request.ToolConfig = &models.GoogleToolConfig{FunctionCallingConfig: &models.GoogleFunctionCallingConfig{Mode: test.mode}}
			if test.name == "undeclared name" {
				request.ToolConfig.FunctionCallingConfig.AllowedFunctionNames = []string{"delete_all"}
			} else if test.name == "allowed name with auto" {
				request.ToolConfig.FunctionCallingConfig.AllowedFunctionNames = []string{"lookup"}
			}
			if _, _, err := GoogleContentsToPrompt(request); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestGoogleContentsToPromptRejectsSilentLossBoundaries(t *testing.T) {
	base := models.GoogleGenerateRequest{
		Contents: []models.GoogleContent{{Role: "user", Parts: []models.GooglePart{{Text: "hello"}}}},
	}
	tests := []struct {
		name   string
		mutate func(*models.GoogleGenerateRequest)
		want   string
	}{
		{
			name: "empty part",
			mutate: func(req *models.GoogleGenerateRequest) {
				req.Contents[0].Parts = []models.GooglePart{{}}
			},
			want: "no supported value",
		},
		{
			name: "ambiguous part",
			mutate: func(req *models.GoogleGenerateRequest) {
				req.Contents[0].Parts = []models.GooglePart{
					{Text: "hello", FunctionCall: &models.GoogleFunctionCall{Name: "lookup"}},
				}
			},
			want: "multiple values",
		},
		{
			name: "unknown role",
			mutate: func(req *models.GoogleGenerateRequest) {
				req.Contents[0].Role = "assistant"
			},
			want: "unsupported Google content role",
		},
		{
			name: "malformed function call",
			mutate: func(req *models.GoogleGenerateRequest) {
				req.Contents[0].Parts = []models.GooglePart{{FunctionCall: &models.GoogleFunctionCall{Name: "  "}}}
			},
			want: "function call name is empty",
		},
		{
			name: "duplicate declarations",
			mutate: func(req *models.GoogleGenerateRequest) {
				req.Tools = []models.GoogleTool{{FunctionDeclarations: []models.GoogleFunctionDeclaration{
					{Name: "lookup"}, {Name: "lookup"},
				}}}
			},
			want: "declared more than once",
		},
		{
			name: "non-text system instruction",
			mutate: func(req *models.GoogleGenerateRequest) {
				req.SystemInstruction = &models.GoogleContent{Parts: []models.GooglePart{{InlineData: &models.GoogleInlineData{MIMEType: "image/png", Data: "aGk="}}}}
			},
			want: "only text is supported",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := models.GoogleGenerateRequest{
				Contents: []models.GoogleContent{{
					Role:  base.Contents[0].Role,
					Parts: append([]models.GooglePart(nil), base.Contents[0].Parts...),
				}},
			}
			test.mutate(&request)
			if _, _, err := GoogleContentsToPrompt(request); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want substring %q", err, test.want)
			}
		})
	}
}
