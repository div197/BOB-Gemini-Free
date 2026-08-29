package format

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/div197/bob-gemini-free/internal/models"
)

func ResponsesInputToMessages(input any, instructions string) ([]map[string]any, error) {
	var messages []map[string]any

	if instructions != "" {
		messages = append(messages, map[string]any{
			"role":    "system",
			"content": instructions,
		})
	}

	if strInput, ok := input.(string); ok {
		messages = append(messages, map[string]any{
			"role":    "user",
			"content": strInput,
		})
	} else if listInput, ok := input.([]any); ok {
		for itemIndex, item := range listInput {
			if strItem, ok := item.(string); ok {
				messages = append(messages, map[string]any{
					"role":    "user",
					"content": strItem,
				})
			} else if mapItem, ok := item.(map[string]any); ok {
				itemType, err := responsesOptionalString(mapItem, "type", maxResponsesIdentifierBytes)
				if err != nil {
					return nil, fmt.Errorf("responses input item %d: %w", itemIndex, err)
				}
				itemRole, err := responsesOptionalString(mapItem, "role", maxResponsesIdentifierBytes)
				if err != nil {
					return nil, fmt.Errorf("responses input item %d: %w", itemIndex, err)
				}

				if itemType == "function_call_output" {
					callID, err := responsesRequiredString(mapItem, "call_id", true)
					if err != nil {
						return nil, fmt.Errorf("responses input item %d: %w", itemIndex, err)
					}
					output, err := responsesToolResultContent(mapItem)
					if err != nil {
						return nil, fmt.Errorf("responses input item %d: %w", itemIndex, err)
					}
					toolMessage := map[string]any{
						"role":         "tool",
						"tool_call_id": callID,
						"content":      output,
					}
					if name, err := responsesOptionalString(mapItem, "name", MaxToolNameBytes); err != nil {
						return nil, fmt.Errorf("responses input item %d: %w", itemIndex, err)
					} else if name != "" {
						toolMessage["name"] = name
					}
					messages = append(messages, toolMessage)
				} else if itemType == "function_call" {
					toolCall, err := responsesToolCall(mapItem)
					if err != nil {
						return nil, fmt.Errorf("responses input item %d: %w", itemIndex, err)
					}
					messages = append(messages, map[string]any{
						"role":       "assistant",
						"content":    nil,
						"tool_calls": []map[string]any{toolCall},
					})
				} else if itemRole == "assistant" || (itemType == "message" && itemRole == "assistant") {
					if itemType != "" && itemType != "message" {
						return nil, fmt.Errorf("responses input item %d: unsupported input item type %q", itemIndex, itemType)
					}
					textAcc, tcList, err := responsesAssistantContent(mapItem["content"])
					if err != nil {
						return nil, fmt.Errorf("responses input item %d: %w", itemIndex, err)
					}

					m := map[string]any{
						"role":    "assistant",
						"content": textAcc,
					}
					if textAcc == "" {
						m["content"] = nil
					}

					if len(tcList) > 0 {
						m["tool_calls"] = tcList
					}
					if name, err := responsesOptionalString(mapItem, "name", MaxToolNameBytes); err != nil {
						return nil, fmt.Errorf("responses input item %d: %w", itemIndex, err)
					} else if name != "" {
						m["name"] = name
					}
					messages = append(messages, m)
				} else {
					role := itemRole
					if role == "" {
						if itemType == "message" {
							return nil, fmt.Errorf("responses input item %d: message item is missing role", itemIndex)
						}
						role = "user"
					}
					switch role {
					case "user", "system", "developer":
					default:
						return nil, fmt.Errorf("responses input item %d: unsupported message role %q", itemIndex, role)
					}
					if itemType != "" && itemType != "message" {
						return nil, fmt.Errorf("responses input item %d: unsupported input item type %q", itemIndex, itemType)
					}
					if _, exists := mapItem["content"]; !exists {
						return nil, fmt.Errorf("responses input item %d: message item is missing content", itemIndex)
					}
					content, err := responsesMessageContent(mapItem["content"])
					if err != nil {
						return nil, fmt.Errorf("responses input item %d: %w", itemIndex, err)
					}
					messages = append(messages, map[string]any{
						"role":    role,
						"content": content,
					})
				}
			} else {
				return nil, fmt.Errorf("responses input item %d: unsupported item type %T", itemIndex, item)
			}
		}
	} else if input != nil {
		return nil, fmt.Errorf("responses input must be a string or array")
	}

	return messages, nil
}

const maxResponsesIdentifierBytes = 256

func responsesRequiredString(item map[string]any, field string, nonEmpty bool) (string, error) {
	raw, ok := item[field]
	if !ok {
		return "", fmt.Errorf("responses item is missing %q", field)
	}
	value, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("responses item field %q must be a string", field)
	}
	if nonEmpty && strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("responses item field %q must not be empty", field)
	}
	return value, nil
}

func responsesOptionalString(item map[string]any, field string, maxBytes int) (string, error) {
	raw, ok := item[field]
	if !ok || raw == nil {
		return "", nil
	}
	value, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("responses item field %q must be a string", field)
	}
	if len(value) > maxBytes {
		return "", fmt.Errorf("responses item field %q exceeds %d bytes", field, maxBytes)
	}
	return value, nil
}

func responsesToolCall(item map[string]any) (map[string]any, error) {
	callID, err := responsesRequiredString(item, "call_id", true)
	if err != nil {
		return nil, err
	}
	if len(callID) > maxResponsesIdentifierBytes {
		return nil, fmt.Errorf("responses item field %q exceeds %d bytes", "call_id", maxResponsesIdentifierBytes)
	}
	name, err := responsesRequiredString(item, "name", true)
	if err != nil {
		return nil, err
	}
	arguments, err := responsesRequiredString(item, "arguments", false)
	if err != nil {
		return nil, err
	}
	if err := ValidateToolCall(name, arguments); err != nil {
		return nil, err
	}
	if strings.TrimSpace(arguments) == "" {
		arguments = "{}"
	}
	return map[string]any{
		"id":   callID,
		"type": "function",
		"function": map[string]any{
			"name":      name,
			"arguments": arguments,
		},
	}, nil
}

func responsesToolResultContent(item map[string]any) (string, error) {
	raw, ok := item["output"]
	if !ok {
		return "", fmt.Errorf("responses item is missing %q", "output")
	}
	if value, ok := raw.(string); ok {
		if len(value) > MaxToolArgumentBytes {
			return "", fmt.Errorf("responses item field %q exceeds %d bytes", "output", MaxToolArgumentBytes)
		}
		return value, nil
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return "", fmt.Errorf("encode responses tool output: %w", err)
	}
	if len(encoded) > MaxToolArgumentBytes {
		return "", fmt.Errorf("responses item field %q exceeds %d bytes", "output", MaxToolArgumentBytes)
	}
	return string(encoded), nil
}

func responsesAssistantContent(raw any) (string, []map[string]any, error) {
	if raw == nil {
		return "", nil, nil
	}
	if value, ok := raw.(string); ok {
		return value, nil, nil
	}
	parts, ok := raw.([]any)
	if !ok {
		return "", nil, fmt.Errorf("assistant content must be a string or array")
	}
	var textParts []string
	var toolCalls []map[string]any
	for index, part := range parts {
		partMap, ok := part.(map[string]any)
		if !ok {
			return "", nil, fmt.Errorf("assistant content item %d has unsupported type %T", index, part)
		}
		partType, err := responsesRequiredString(partMap, "type", true)
		if err != nil {
			return "", nil, err
		}
		switch partType {
		case "output_text":
			text, err := responsesRequiredString(partMap, "text", false)
			if err != nil {
				return "", nil, err
			}
			textParts = append(textParts, text)
		case "function_call":
			toolCall, err := responsesToolCall(partMap)
			if err != nil {
				return "", nil, err
			}
			toolCalls = append(toolCalls, toolCall)
		default:
			return "", nil, fmt.Errorf("unsupported assistant content type %q", partType)
		}
	}
	return strings.Join(textParts, ""), toolCalls, nil
}

func responsesMessageContent(raw any) (any, error) {
	if raw == nil {
		return "", nil
	}
	if value, ok := raw.(string); ok {
		return value, nil
	}
	parts, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("message content must be a string or array")
	}
	var textParts []string
	var contentParts []any
	hasImage := false
	for index, part := range parts {
		partMap, ok := part.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("message content item %d has unsupported type %T", index, part)
		}
		partType, err := responsesRequiredString(partMap, "type", true)
		if err != nil {
			return nil, err
		}
		switch partType {
		case "text", "input_text":
			text, err := responsesRequiredString(partMap, "text", false)
			if err != nil {
				return nil, err
			}
			textParts = append(textParts, text)
			contentParts = append(contentParts, map[string]any{"type": "text", "text": text})
		case "image_url", "input_image":
			imageURL, err := responsesImageURL(partMap)
			if err != nil {
				return nil, err
			}
			hasImage = true
			contentParts = append(contentParts, map[string]any{
				"type": "image_url",
				"image_url": map[string]any{
					"url": imageURL,
				},
			})
		default:
			return nil, fmt.Errorf("unsupported message content type %q", partType)
		}
	}
	if hasImage {
		return contentParts, nil
	}
	return strings.Join(textParts, " "), nil
}

func responsesImageURL(item map[string]any) (string, error) {
	var imageURL string
	if raw, ok := item["image_url"]; ok {
		switch value := raw.(type) {
		case string:
			imageURL = value
		case map[string]any:
			var fieldOK bool
			imageURL, fieldOK = value["url"].(string)
			if !fieldOK {
				return "", fmt.Errorf("image content item field %q must be a string", "url")
			}
		default:
			return "", fmt.Errorf("image content item field %q must be a string or object", "image_url")
		}
	} else if value, ok := item["url"].(string); ok {
		imageURL = value
	} else {
		return "", fmt.Errorf("image content item is missing image URL")
	}
	if strings.TrimSpace(imageURL) == "" {
		return "", fmt.Errorf("image content item has an empty image URL")
	}
	return imageURL, nil
}

func BuildResponseOutput(text string, toolCalls []models.OpenAIToolCall, msgID string) []map[string]any {
	var output []map[string]any

	if len(toolCalls) > 0 {
		for _, tc := range toolCalls {
			output = append(output, map[string]any{
				"type":      "function_call",
				"id":        tc.ID,
				"call_id":   tc.ID,
				"name":      tc.Function.Name,
				"arguments": tc.Function.Arguments,
				"status":    "completed",
			})
		}
	}

	if text != "" || len(toolCalls) == 0 {
		output = append(output, map[string]any{
			"type":   "message",
			"id":     msgID,
			"role":   "assistant",
			"status": "completed",
			"content": []map[string]any{
				{
					"type":        "output_text",
					"text":        text,
					"annotations": []any{},
				},
			},
		})
	}

	return output
}
