package geminiapi

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/div197/bob-gemini-free/internal/format"
	"github.com/div197/bob-gemini-free/internal/models"
)

const (
	MaxInlineImageBytes      = 20 * 1024 * 1024
	MaxInlineImageBase64Size = (MaxInlineImageBytes*4 + 2) / 3
)

// FromOpenAI translates an OpenAI-shaped chat request into the public Gemini
// request schema. It preserves system instructions, multimodal data URLs,
// native function declarations, and tool-choice semantics.
func FromOpenAI(req models.OpenAIChatRequest) (GenerateContentRequest, error) {
	if err := format.ValidateToolResultReferences(req.Messages); err != nil {
		return GenerateContentRequest{}, err
	}
	var result GenerateContentRequest
	var systemParts []Part

	for _, message := range req.Messages {
		role := strings.ToLower(strings.TrimSpace(message.Role))
		if role == "system" || role == "developer" {
			parts, err := partsFromContent(message.Content)
			if err != nil {
				return GenerateContentRequest{}, err
			}
			systemParts = append(systemParts, parts...)
			continue
		}

		content := Content{Role: "user"}
		switch role {
		case "assistant", "model":
			content.Role = "model"
		case "user":
			content.Role = "user"
		case "tool":
			content.Role = "user"
			response := toolResponse(message)
			content.Parts = []Part{{FunctionResponse: &response}}
			result.Contents = append(result.Contents, content)
			continue
		default:
			return GenerateContentRequest{}, fmt.Errorf("unsupported OpenAI message role %q for Gemini Developer API", message.Role)
		}

		parts, err := partsFromContent(message.Content)
		if err != nil {
			return GenerateContentRequest{}, err
		}
		content.Parts = append(content.Parts, parts...)
		for _, call := range message.ToolCalls {
			args := map[string]any{}
			arguments := strings.TrimSpace(call.Function.Arguments)
			if len(arguments) > format.MaxToolArgumentBytes {
				return GenerateContentRequest{}, fmt.Errorf("tool call %q arguments exceed %d bytes", call.Function.Name, format.MaxToolArgumentBytes)
			}
			if arguments != "" {
				if err := json.Unmarshal([]byte(arguments), &args); err != nil {
					return GenerateContentRequest{}, fmt.Errorf("tool call %q has invalid JSON arguments: %w", call.Function.Name, err)
				}
			}
			content.Parts = append(content.Parts, Part{FunctionCall: &FunctionCall{Name: call.Function.Name, Args: args}})
		}
		if len(content.Parts) > 0 {
			result.Contents = append(result.Contents, content)
		}
	}

	if len(systemParts) > 0 {
		result.SystemInstruction = &Content{Parts: systemParts}
	}

	if len(req.Tools) > 0 {
		if len(req.Tools) > format.MaxToolDefinitions {
			return GenerateContentRequest{}, fmt.Errorf("tool definition count exceeds %d", format.MaxToolDefinitions)
		}
		declarations := make([]FunctionDeclaration, 0, len(req.Tools))
		for _, tool := range req.Tools {
			if tool.Type != "" && !strings.EqualFold(tool.Type, "function") {
				return GenerateContentRequest{}, fmt.Errorf("Gemini Developer API direct routing supports function tools only, not %q", tool.Type)
			}
			fn := tool.Function
			if fn.Name == "" {
				fn.Name = tool.Name
			}
			if fn.Description == "" {
				fn.Description = tool.Description
			}
			if fn.Parameters == nil {
				fn.Parameters = tool.Parameters
			}
			if strings.TrimSpace(fn.Name) == "" {
				return GenerateContentRequest{}, errors.New("Gemini Developer API tool name is empty")
			}
			if len(fn.Name) > format.MaxToolNameBytes {
				return GenerateContentRequest{}, fmt.Errorf("Gemini Developer API tool name exceeds %d bytes", format.MaxToolNameBytes)
			}
			if len(fn.Description) > format.MaxToolDescriptionBytes {
				return GenerateContentRequest{}, fmt.Errorf("Gemini Developer API tool %q description exceeds %d bytes", fn.Name, format.MaxToolDescriptionBytes)
			}
			if err := format.ValidateToolSchema(fn.Parameters); err != nil {
				return GenerateContentRequest{}, fmt.Errorf("Gemini Developer API tool %q schema: %w", fn.Name, err)
			}
			declarations = append(declarations, FunctionDeclaration{
				Name:        fn.Name,
				Description: fn.Description,
				Parameters:  fn.Parameters,
			})
		}
		result.Tools = []Tool{{FunctionDeclarations: declarations}}
		if encoded, err := json.Marshal(result.Tools); err != nil {
			return GenerateContentRequest{}, fmt.Errorf("encode Gemini Developer API tool declarations: %w", err)
		} else if len(encoded) > format.MaxToolDefinitionBytes {
			return GenerateContentRequest{}, fmt.Errorf("Gemini Developer API tool declarations exceed %d bytes", format.MaxToolDefinitionBytes)
		}
		var err error
		result.ToolConfig, err = toolConfig(req.ToolChoice)
		if err != nil {
			return GenerateContentRequest{}, err
		}
	}

	maxTokens := req.MaxCompletionTokens
	if maxTokens == nil {
		maxTokens = req.MaxTokens
	}
	if req.ResponseFormat != nil {
		switch strings.ToLower(strings.TrimSpace(req.ResponseFormat.Type)) {
		case "json_object":
			// The public Gemini API supports MIME-type constrained JSON output.
		case "text":
			// Text is the normal default and needs no extra provider setting.
		case "json_schema":
			// OpenAI's schema payload is not represented in this request type;
			// silently dropping it would produce a false structured-output claim.
			return GenerateContentRequest{}, errors.New("json_schema response format is not supported on the explicit Gemini Developer API route")
		default:
			return GenerateContentRequest{}, fmt.Errorf("unsupported response format %q for Gemini Developer API", req.ResponseFormat.Type)
		}
	}
	if req.Temperature != nil || req.TopP != nil || maxTokens != nil || req.ResponseFormat != nil {
		config := &GenerationConfig{Temperature: req.Temperature, TopP: req.TopP, MaxOutputTokens: maxTokens}
		if req.ResponseFormat != nil && strings.EqualFold(strings.TrimSpace(req.ResponseFormat.Type), "json_object") {
			config.ResponseMimeType = "application/json"
		}
		result.GenerationConfig = config
	}
	return result, nil
}

func partsFromContent(content any) ([]Part, error) {
	if content == nil {
		return nil, nil
	}
	switch value := content.(type) {
	case string:
		if value == "" {
			return nil, nil
		}
		return []Part{{Text: value}}, nil
	case []any:
		var parts []Part
		for _, item := range value {
			obj, ok := item.(map[string]any)
			if !ok {
				return nil, errors.New("unsupported OpenAI content part")
			}
			part, err := contentPart(obj)
			if err != nil {
				return nil, err
			}
			if part.Text != "" || part.InlineData != nil {
				parts = append(parts, part)
			}
		}
		return parts, nil
	default:
		return nil, errors.New("unsupported OpenAI message content")
	}
}

func contentPart(obj map[string]any) (Part, error) {
	typeName, _ := obj["type"].(string)
	switch typeName {
	case "text", "input_text":
		text, _ := obj["text"].(string)
		return Part{Text: text}, nil
	case "image_url", "input_image":
		var rawURL string
		switch image := obj["image_url"].(type) {
		case string:
			rawURL = image
		case map[string]any:
			rawURL, _ = image["url"].(string)
		}
		data, err := dataURLToInlineData(rawURL)
		if err != nil {
			return Part{}, err
		}
		return Part{InlineData: data}, nil
	default:
		return Part{}, fmt.Errorf("unsupported OpenAI content part type %q for Gemini Developer API", typeName)
	}
}

func dataURLToInlineData(rawURL string) (*InlineData, error) {
	rawURL = strings.TrimSpace(rawURL)
	if !strings.HasPrefix(rawURL, "data:") {
		return nil, errors.New("Gemini Developer API direct image routing supports data URLs only; remote image URLs are not fetched")
	}
	parts := strings.SplitN(rawURL, ",", 2)
	if len(parts) != 2 {
		return nil, errors.New("invalid image data URL")
	}
	metadata := strings.TrimPrefix(parts[0], "data:")
	segments := strings.Split(metadata, ";")
	if len(segments) < 2 || segments[len(segments)-1] != "base64" {
		return nil, errors.New("image data URL must use base64 encoding")
	}
	mimeType := segments[0]
	if mimeType == "" {
		return nil, errors.New("image data URL is missing its MIME type")
	}
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	switch mimeType {
	case "image/png", "image/jpeg", "image/jpg", "image/gif", "image/webp", "image/avif", "image/heic", "image/heif":
	default:
		return nil, fmt.Errorf("unsupported image data URL MIME type %q", mimeType)
	}
	encoded := strings.TrimSpace(parts[1])
	if len(encoded) > MaxInlineImageBase64Size || int64(base64.StdEncoding.DecodedLen(len(encoded))) > MaxInlineImageBytes {
		return nil, fmt.Errorf("image data URL exceeds %d decoded bytes", MaxInlineImageBytes)
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid image data URL: %w", err)
	}
	if len(decoded) == 0 {
		return nil, errors.New("image data URL is empty")
	}
	return &InlineData{MimeType: mimeType, Data: base64.StdEncoding.EncodeToString(decoded)}, nil
}

func toolConfig(choice any) (*ToolConfig, error) {
	config := FunctionCallingConfig{Mode: "AUTO"}
	switch value := choice.(type) {
	case nil:
		// OpenAI defaults tool choice to automatic selection when tools exist.
	case string:
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "none":
			config.Mode = "NONE"
		case "required", "any":
			config.Mode = "ANY"
		case "auto", "":
			config.Mode = "AUTO"
		default:
			return nil, fmt.Errorf("unsupported tool_choice %q for Gemini Developer API", value)
		}
	case map[string]any:
		if typeName, ok := value["type"].(string); ok && typeName != "" && !strings.EqualFold(typeName, "function") {
			return nil, fmt.Errorf("unsupported tool_choice type %q for Gemini Developer API", typeName)
		}
		if function, ok := value["function"].(map[string]any); ok {
			if name, ok := function["name"].(string); ok && strings.TrimSpace(name) != "" {
				config.Mode = "ANY"
				config.AllowedFunctionNames = []string{name}
			} else {
				return nil, errors.New("Gemini Developer API tool_choice function name is empty")
			}
		} else {
			return nil, errors.New("Gemini Developer API tool_choice function is missing")
		}
	default:
		return nil, errors.New("unsupported tool_choice for Gemini Developer API")
	}
	return &ToolConfig{FunctionCallingConfig: config}, nil
}

func toolResponse(message models.OpenAIMessage) FunctionResponse {
	name := message.Name
	if name == "" {
		name = message.ToolCallID
	}
	response := map[string]any{"result": message.Content}
	if text, ok := message.Content.(string); ok {
		var decoded any
		if json.Unmarshal([]byte(text), &decoded) == nil {
			if object, ok := decoded.(map[string]any); ok {
				response = object
			}
		}
	}
	return FunctionResponse{Name: name, Response: response}
}

type OpenAIResult struct {
	Text             string
	Thinking         string
	ToolCalls        []models.OpenAIToolCall
	FinishReason     string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	ReasoningTokens  int
}

func ToOpenAIResult(response GenerateContentResponse) (OpenAIResult, error) {
	if len(response.Candidates) == 0 {
		if response.PromptFeedback != nil && response.PromptFeedback.BlockReason != "" {
			return OpenAIResult{}, fmt.Errorf("Gemini Developer API blocked the prompt: %s", response.PromptFeedback.BlockReason)
		}
		return OpenAIResult{}, errors.New("Gemini Developer API returned no candidates")
	}
	if len(response.Candidates) > 1 {
		return OpenAIResult{}, fmt.Errorf("Gemini Developer API returned %d candidates; this OpenAI adapter supports exactly one", len(response.Candidates))
	}
	var result OpenAIResult
	var finishErr error
	toolCount := 0
	for candidateIndex, candidate := range response.Candidates {
		if candidate.Content == nil {
			result.FinishReason, finishErr = openAIFinishReason(candidate.FinishReason, false)
			if finishErr != nil {
				return OpenAIResult{}, finishErr
			}
			continue
		}
		for partIndex, part := range candidate.Content.Parts {
			if part.Text != "" {
				if part.Thought {
					result.Thinking += part.Text
				} else {
					result.Text += part.Text
				}
			}
			if part.FunctionCall != nil {
				arguments, err := json.Marshal(part.FunctionCall.Args)
				if err != nil {
					return OpenAIResult{}, fmt.Errorf("could not encode Gemini function-call arguments: %w", err)
				}
				if toolCount >= format.MaxToolDefinitions {
					return OpenAIResult{}, fmt.Errorf("Gemini Developer API returned more than %d tool calls", format.MaxToolDefinitions)
				}
				if err := format.ValidateToolCall(part.FunctionCall.Name, string(arguments)); err != nil {
					return OpenAIResult{}, fmt.Errorf("invalid Gemini function-call response: %w", err)
				}
				toolCount++
				result.ToolCalls = append(result.ToolCalls, models.OpenAIToolCall{
					ID:   fmt.Sprintf("call_gemini_%d_%d", candidateIndex, partIndex),
					Type: "function",
					Function: models.OpenAIToolCallFunction{
						Name:      part.FunctionCall.Name,
						Arguments: string(arguments),
					},
				})
			}
		}
		if result.FinishReason == "" {
			result.FinishReason, finishErr = openAIFinishReason(candidate.FinishReason, len(result.ToolCalls) > 0)
			if finishErr != nil {
				return OpenAIResult{}, finishErr
			}
		}
	}
	if result.FinishReason == "" {
		result.FinishReason, finishErr = openAIFinishReason("STOP", len(result.ToolCalls) > 0)
		if finishErr != nil {
			return OpenAIResult{}, finishErr
		}
	}
	if response.UsageMetadata != nil {
		result.PromptTokens = response.UsageMetadata.PromptTokenCount
		result.CompletionTokens = response.UsageMetadata.CandidatesTokenCount
		result.TotalTokens = response.UsageMetadata.TotalTokenCount
		result.ReasoningTokens = response.UsageMetadata.ThoughtsTokenCount
	}
	return result, nil
}

func openAIFinishReason(reason string, hasToolCalls bool) (string, error) {
	reason = strings.ToUpper(strings.TrimSpace(reason))
	if reason == "" {
		if hasToolCalls {
			return "tool_calls", nil
		}
		return "stop", nil
	}
	switch reason {
	case "STOP":
		if hasToolCalls {
			return "tool_calls", nil
		}
		return "stop", nil
	case "MAX_TOKENS":
		return "length", nil
	case "SAFETY", "RECITATION", "BLOCKLIST", "PROHIBITED_CONTENT", "SPII", "LANGUAGE", "IMAGE_SAFETY", "IMAGE_PROHIBITED_CONTENT", "IMAGE_RECITATION":
		return "content_filter", nil
	case "MALFORMED_FUNCTION_CALL":
		return "error", nil
	default:
		return "", fmt.Errorf("unsupported Gemini finish reason %q", reason)
	}
}
