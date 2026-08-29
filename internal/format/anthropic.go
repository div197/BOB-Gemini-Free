package format

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/div197/bob-gemini-free/internal/models"
)

// AnthropicToOpenAIChatRequest converts an Anthropic Messages request into an OpenAI Chat request.
func AnthropicToOpenAIChatRequest(req models.AnthropicMessagesRequest) (models.OpenAIChatRequest, error) {
	var openAIMessages []models.OpenAIMessage

	// 1. Extract system prompt
	if req.System != nil {
		sysText, err := anthropicSystemText(req.System)
		if err != nil {
			return models.OpenAIChatRequest{}, err
		}
		if strings.TrimSpace(sysText) != "" {
			openAIMessages = append(openAIMessages, models.OpenAIMessage{
				Role:    "system",
				Content: sysText,
			})
		}
	}

	// 2. Parse Anthropic messages
	for messageIndex, msg := range req.Messages {
		role := msg.Role
		if role == "" {
			return models.OpenAIChatRequest{}, fmt.Errorf("Anthropic message %d is missing role", messageIndex)
		}
		if role != "user" && role != "assistant" {
			return models.OpenAIChatRequest{}, fmt.Errorf("Anthropic message %d has unsupported role %q", messageIndex, role)
		}

		if strContent, ok := msg.Content.(string); ok {
			openAIMessages = append(openAIMessages, models.OpenAIMessage{
				Role:    role,
				Content: strContent,
			})
		} else if blockList, ok := msg.Content.([]any); ok {
			var contentParts []any
			var toolCalls []models.OpenAIToolCall

			flushContent := func() {
				if len(contentParts) == 0 && len(toolCalls) == 0 {
					return
				}
				var content any
				if len(contentParts) > 0 {
					content = contentParts
				}
				openAIMessages = append(openAIMessages, models.OpenAIMessage{
					Role:      role,
					Content:   content,
					ToolCalls: toolCalls,
				})
				contentParts = nil
				toolCalls = nil
			}

			for blockIndex, b := range blockList {
				bMap, ok := b.(map[string]any)
				if !ok {
					return models.OpenAIChatRequest{}, fmt.Errorf("Anthropic message %d content block %d has unsupported type %T", messageIndex, blockIndex, b)
				}
				bType, err := anthropicRequiredString(bMap, "type", true, MaxToolNameBytes)
				if err != nil {
					return models.OpenAIChatRequest{}, fmt.Errorf("Anthropic message %d content block %d: %w", messageIndex, blockIndex, err)
				}

				switch bType {
				case "text":
					txt, err := anthropicRequiredString(bMap, "text", false, 0)
					if err != nil {
						return models.OpenAIChatRequest{}, fmt.Errorf("Anthropic message %d content block %d: %w", messageIndex, blockIndex, err)
					}
					contentParts = append(contentParts, map[string]any{
						"type": "text",
						"text": txt,
					})
				case "image":
					imagePart, err := anthropicImagePart(bMap)
					if err != nil {
						return models.OpenAIChatRequest{}, fmt.Errorf("Anthropic message %d content block %d: %w", messageIndex, blockIndex, err)
					}
					contentParts = append(contentParts, imagePart)
				case "tool_use":
					if role != "assistant" {
						return models.OpenAIChatRequest{}, fmt.Errorf("Anthropic message %d content block %d: tool_use requires assistant role", messageIndex, blockIndex)
					}
					toolCall, err := anthropicToolUseCall(bMap)
					if err != nil {
						return models.OpenAIChatRequest{}, fmt.Errorf("Anthropic message %d content block %d: %w", messageIndex, blockIndex, err)
					}
					toolCalls = append(toolCalls, toolCall)
				case "tool_result":
					if role != "user" {
						return models.OpenAIChatRequest{}, fmt.Errorf("Anthropic message %d content block %d: tool_result requires user role", messageIndex, blockIndex)
					}
					flushContent()
					toolCallID, err := anthropicRequiredString(bMap, "tool_use_id", true, maxResponsesIdentifierBytes)
					if err != nil {
						return models.OpenAIChatRequest{}, fmt.Errorf("Anthropic message %d content block %d: %w", messageIndex, blockIndex, err)
					}
					resStr, err := anthropicToolResultContent(bMap)
					if err != nil {
						return models.OpenAIChatRequest{}, fmt.Errorf("Anthropic message %d content block %d: %w", messageIndex, blockIndex, err)
					}
					openAIMessages = append(openAIMessages, models.OpenAIMessage{
						Role:       "tool",
						ToolCallID: toolCallID,
						Content:    resStr,
					})
				default:
					return models.OpenAIChatRequest{}, fmt.Errorf("Anthropic message %d content block %d: unsupported content block type %q", messageIndex, blockIndex, bType)
				}
			}
			flushContent()
			if len(blockList) == 0 {
				openAIMessages = append(openAIMessages, models.OpenAIMessage{Role: role, Content: ""})
			}
		} else {
			return models.OpenAIChatRequest{}, fmt.Errorf("Anthropic message %d content must be a string or array", messageIndex)
		}
	}

	// 3. Convert Tools
	openAITools, err := anthropicTools(req.Tools)
	if err != nil {
		return models.OpenAIChatRequest{}, err
	}
	toolChoice, err := anthropicToolChoice(req.ToolChoice, openAITools)
	if err != nil {
		return models.OpenAIChatRequest{}, err
	}

	reasoningEffort := ""
	modelName := req.Model
	if req.Thinking != nil && req.Thinking.Type == "enabled" {
		if req.Thinking.BudgetTokens > 0 && req.Thinking.BudgetTokens < 1000 {
			reasoningEffort = "low"
		} else {
			reasoningEffort = "high"
		}
		if !strings.Contains(modelName, "thinking") && !strings.Contains(modelName, "@think=") {
			if modelName == "claude-3-7-sonnet" || modelName == "claude-3-5-sonnet" || modelName == "claude-code" {
				modelName = "gemini-3.7-flash-thinking"
			}
		}
	}

	return models.OpenAIChatRequest{
		Model:           modelName,
		Messages:        openAIMessages,
		Tools:           openAITools,
		ToolChoice:      toolChoice,
		ReasoningEffort: reasoningEffort,
		Stream:          req.Stream,
	}, nil
}

func anthropicRequiredString(item map[string]any, field string, nonEmpty bool, maxBytes int) (string, error) {
	raw, ok := item[field]
	if !ok {
		return "", fmt.Errorf("item is missing %q", field)
	}
	value, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("item field %q must be a string", field)
	}
	if nonEmpty && strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("item field %q must not be empty", field)
	}
	if maxBytes > 0 && len(value) > maxBytes {
		return "", fmt.Errorf("item field %q exceeds %d bytes", field, maxBytes)
	}
	return value, nil
}

func anthropicSystemText(raw any) (string, error) {
	if value, ok := raw.(string); ok {
		return value, nil
	}
	if item, ok := raw.(map[string]any); ok {
		if typ, err := anthropicOptionalString(item, "type", MaxToolNameBytes); err != nil {
			return "", err
		} else if typ != "" && typ != "text" {
			return "", fmt.Errorf("unsupported system block type %q", typ)
		}
		return anthropicRequiredString(item, "text", false, 0)
	}
	items, ok := raw.([]any)
	if !ok {
		return "", fmt.Errorf("system must be a string, text block, or array")
	}
	parts := make([]string, 0, len(items))
	for index, item := range items {
		if value, ok := item.(string); ok {
			parts = append(parts, value)
			continue
		}
		block, ok := item.(map[string]any)
		if !ok {
			return "", fmt.Errorf("system block %d has unsupported type %T", index, item)
		}
		typ, err := anthropicOptionalString(block, "type", MaxToolNameBytes)
		if err != nil {
			return "", fmt.Errorf("system block %d: %w", index, err)
		}
		if typ != "" && typ != "text" {
			return "", fmt.Errorf("system block %d has unsupported type %q", index, typ)
		}
		text, err := anthropicRequiredString(block, "text", false, 0)
		if err != nil {
			return "", fmt.Errorf("system block %d: %w", index, err)
		}
		parts = append(parts, text)
	}
	return strings.Join(parts, "\n"), nil
}

func anthropicOptionalString(item map[string]any, field string, maxBytes int) (string, error) {
	raw, ok := item[field]
	if !ok || raw == nil {
		return "", nil
	}
	value, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("item field %q must be a string", field)
	}
	if maxBytes > 0 && len(value) > maxBytes {
		return "", fmt.Errorf("item field %q exceeds %d bytes", field, maxBytes)
	}
	return value, nil
}

func anthropicImagePart(block map[string]any) (map[string]any, error) {
	rawSource, ok := block["source"]
	if !ok {
		return nil, fmt.Errorf("image block is missing %q", "source")
	}
	source, ok := rawSource.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("image source must be an object")
	}
	sourceType, err := anthropicOptionalString(source, "type", MaxToolNameBytes)
	if err != nil {
		return nil, err
	}
	if sourceType == "" {
		sourceType = "base64"
	}
	switch sourceType {
	case "base64":
		data, err := anthropicRequiredString(source, "data", true, 0)
		if err != nil {
			return nil, err
		}
		mediaType, err := anthropicOptionalString(source, "media_type", MaxInlineImageMIMEBytes)
		if err != nil {
			return nil, err
		}
		if mediaType == "" {
			mediaType = "image/png"
		}
		if _, normalizedMIME, err := decodeInlineImageData(data, mediaType); err != nil {
			return nil, err
		} else {
			mediaType = normalizedMIME
		}
		return map[string]any{
			"type": "image_url",
			"image_url": map[string]string{
				"url": fmt.Sprintf("data:%s;base64,%s", mediaType, data),
			},
		}, nil
	case "url":
		url, err := anthropicRequiredString(source, "url", true, 0)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"type": "image_url",
			"image_url": map[string]string{
				"url": url,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported image source type %q", sourceType)
	}
}

func anthropicToolUseCall(block map[string]any) (models.OpenAIToolCall, error) {
	id, err := anthropicRequiredString(block, "id", true, maxResponsesIdentifierBytes)
	if err != nil {
		return models.OpenAIToolCall{}, err
	}
	name, err := anthropicRequiredString(block, "name", true, MaxToolNameBytes)
	if err != nil {
		return models.OpenAIToolCall{}, err
	}
	rawInput, ok := block["input"]
	if !ok || rawInput == nil {
		return models.OpenAIToolCall{}, fmt.Errorf("item is missing %q", "input")
	}
	argsJSON, err := json.Marshal(rawInput)
	if err != nil {
		return models.OpenAIToolCall{}, fmt.Errorf("encode tool input: %w", err)
	}
	if len(argsJSON) == 0 || argsJSON[0] != '{' {
		return models.OpenAIToolCall{}, fmt.Errorf("tool input must be a JSON object")
	}
	if err := ValidateToolCall(name, string(argsJSON)); err != nil {
		return models.OpenAIToolCall{}, err
	}
	return models.OpenAIToolCall{
		ID:   id,
		Type: "function",
		Function: models.OpenAIToolCallFunction{
			Name:      name,
			Arguments: string(argsJSON),
		},
	}, nil
}

func anthropicToolResultContent(block map[string]any) (string, error) {
	rawContent, ok := block["content"]
	if !ok {
		return "", fmt.Errorf("item is missing %q", "content")
	}
	if value, ok := rawContent.(string); ok {
		if len(value) > MaxToolArgumentBytes {
			return "", fmt.Errorf("item field %q exceeds %d bytes", "content", MaxToolArgumentBytes)
		}
		return value, nil
	}
	encoded, err := json.Marshal(rawContent)
	if err != nil {
		return "", fmt.Errorf("encode tool result: %w", err)
	}
	if len(encoded) > MaxToolArgumentBytes {
		return "", fmt.Errorf("item field %q exceeds %d bytes", "content", MaxToolArgumentBytes)
	}
	return string(encoded), nil
}

func anthropicTools(tools []models.AnthropicTool) ([]models.OpenAITool, error) {
	if len(tools) > MaxToolDefinitions {
		return nil, fmt.Errorf("tool definition count exceeds %d", MaxToolDefinitions)
	}
	var openAITools []models.OpenAITool
	for index, tool := range tools {
		name := strings.TrimSpace(tool.Name)
		if name == "" {
			return nil, fmt.Errorf("Anthropic tool %d name is empty", index)
		}
		if len(name) > MaxToolNameBytes {
			return nil, fmt.Errorf("tool %q name exceeds %d bytes", name[:MaxToolNameBytes], MaxToolNameBytes)
		}
		if len(tool.Description) > MaxToolDescriptionBytes {
			return nil, fmt.Errorf("tool %q description exceeds %d bytes", name, MaxToolDescriptionBytes)
		}
		parameters := tool.InputSchema
		if parameters == nil {
			parameters = map[string]any{}
		}
		if err := ValidateToolSchema(parameters); err != nil {
			return nil, fmt.Errorf("tool %q schema: %w", name, err)
		}
		openAITools = append(openAITools, models.OpenAITool{
			Type: "function",
			Function: models.OpenAIFunction{
				Name:        name,
				Description: tool.Description,
				Parameters:  parameters,
			},
		})
	}
	encoded, err := json.Marshal(openAITools)
	if err != nil {
		return nil, fmt.Errorf("encode Anthropic tools: %w", err)
	}
	if len(encoded) > MaxToolDefinitionBytes {
		return nil, fmt.Errorf("tool definitions exceed %d bytes", MaxToolDefinitionBytes)
	}
	return openAITools, nil
}

func anthropicToolChoice(choice any, tools []models.OpenAITool) (any, error) {
	if choice == nil {
		return nil, nil
	}
	switch value := choice.(type) {
	case string:
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "", "auto":
			return "auto", nil
		case "none":
			return "none", nil
		case "any", "required":
			return "required", nil
		default:
			return nil, fmt.Errorf("unsupported Anthropic tool_choice %q", value)
		}
	case map[string]any:
		typeName, err := anthropicOptionalString(value, "type", MaxToolNameBytes)
		if err != nil {
			return nil, err
		}
		if rawDisabled, ok := value["disable_parallel_tool_use"]; ok {
			disabled, ok := rawDisabled.(bool)
			if !ok {
				return nil, errors.New("Anthropic tool_choice disable_parallel_tool_use must be a boolean")
			}
			if disabled {
				return nil, errors.New("Anthropic tool_choice disable_parallel_tool_use is not supported by the emulated tool route")
			}
		}
		switch strings.ToLower(strings.TrimSpace(typeName)) {
		case "", "function":
			if typeName == "" {
				return nil, errors.New("Anthropic tool_choice type is missing")
			}
			if err := ValidateToolChoice(value, tools); err != nil {
				return nil, err
			}
			return value, nil
		case "auto":
			return "auto", nil
		case "none":
			return "none", nil
		case "any":
			return "required", nil
		case "tool":
			name, err := anthropicRequiredString(value, "name", true, MaxToolNameBytes)
			if err != nil {
				return nil, err
			}
			converted := map[string]any{
				"type": "function",
				"function": map[string]any{
					"name": name,
				},
			}
			if err := ValidateToolChoice(converted, tools); err != nil {
				return nil, err
			}
			return converted, nil
		default:
			return nil, fmt.Errorf("unsupported Anthropic tool_choice type %q", typeName)
		}
	default:
		return nil, fmt.Errorf("unsupported Anthropic tool_choice type %T", choice)
	}
}

// ConvertToolCallsToAnthropicBlocks transforms parsed OpenAI tool calls and thinking into Anthropic content blocks.
func ConvertToolCallsToAnthropicBlocks(text string, toolCalls []models.OpenAIToolCall) []models.AnthropicContentBlock {
	return ConvertToolCallsAndThinkingToAnthropicBlocks("", text, toolCalls)
}

// ConvertToolCallsAndThinkingToAnthropicBlocks transforms thinking traces, tool calls, and user text into standard Anthropic content blocks.
func ConvertToolCallsAndThinkingToAnthropicBlocks(thinking, text string, toolCalls []models.OpenAIToolCall) []models.AnthropicContentBlock {
	var blocks []models.AnthropicContentBlock

	if strings.TrimSpace(thinking) != "" {
		blocks = append(blocks, models.AnthropicContentBlock{
			Type:     "thinking",
			Thinking: thinking,
		})
	}

	if strings.TrimSpace(text) != "" {
		blocks = append(blocks, models.AnthropicContentBlock{
			Type: "text",
			Text: text,
		})
	}

	for i, tc := range toolCalls {
		var inputMap map[string]any
		_ = json.Unmarshal([]byte(tc.Function.Arguments), &inputMap)
		if inputMap == nil {
			inputMap = make(map[string]any)
		}

		tID := tc.ID
		if tID == "" {
			tID = fmt.Sprintf("toolu_%s%d", RandHex(8), i)
		}

		blocks = append(blocks, models.AnthropicContentBlock{
			Type:  "tool_use",
			ID:    tID,
			Name:  tc.Function.Name,
			Input: inputMap,
		})
	}

	if len(blocks) == 0 {
		blocks = append(blocks, models.AnthropicContentBlock{
			Type: "text",
			Text: "",
		})
	}

	return blocks
}
