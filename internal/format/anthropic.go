package format

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/div197/bob-gemini-free/internal/models"
)

// AnthropicToOpenAIChatRequest converts an Anthropic Messages request into an OpenAI Chat request.
func AnthropicToOpenAIChatRequest(req models.AnthropicMessagesRequest) models.OpenAIChatRequest {
	var openAIMessages []models.OpenAIMessage

	// 1. Extract system prompt
	if req.System != nil {
		var sysText string
		if sStr, ok := req.System.(string); ok {
			sysText = sStr
		} else if sList, ok := req.System.([]any); ok {
			var parts []string
			for _, item := range sList {
				if m, ok := item.(map[string]any); ok {
					if txt, ok := m["text"].(string); ok {
						parts = append(parts, txt)
					}
				}
			}
			sysText = strings.Join(parts, "\n")
		}
		if strings.TrimSpace(sysText) != "" {
			openAIMessages = append(openAIMessages, models.OpenAIMessage{
				Role:    "system",
				Content: sysText,
			})
		}
	}

	// 2. Parse Anthropic messages
	for _, msg := range req.Messages {
		role := msg.Role
		if role == "" {
			role = "user"
		}

		if strContent, ok := msg.Content.(string); ok {
			openAIMessages = append(openAIMessages, models.OpenAIMessage{
				Role:    role,
				Content: strContent,
			})
		} else if blockList, ok := msg.Content.([]any); ok {
			var contentParts []any
			var toolCalls []models.OpenAIToolCall

			for _, b := range blockList {
				bMap, ok := b.(map[string]any)
				if !ok {
					continue
				}
				bType, _ := bMap["type"].(string)

				switch bType {
				case "text":
					if txt, ok := bMap["text"].(string); ok {
						contentParts = append(contentParts, map[string]any{
							"type": "text",
							"text": txt,
						})
					}
				case "image":
					if src, ok := bMap["source"].(map[string]any); ok {
						data, _ := src["data"].(string)
						mediaType, _ := src["media_type"].(string)
						if mediaType == "" {
							mediaType = "image/png"
						}
						if data != "" {
							contentParts = append(contentParts, map[string]any{
								"type": "image_url",
								"image_url": map[string]string{
									"url": fmt.Sprintf("data:%s;base64,%s", mediaType, data),
								},
							})
						}
					}
				case "tool_use":
					toolID, _ := bMap["id"].(string)
					toolName, _ := bMap["name"].(string)
					inputArgs := bMap["input"]
					argsJSON, _ := json.Marshal(inputArgs)

					toolCalls = append(toolCalls, models.OpenAIToolCall{
						ID:   toolID,
						Type: "function",
						Function: models.OpenAIToolCallFunction{
							Name:      toolName,
							Arguments: string(argsJSON),
						},
					})
				case "tool_result":
					toolCallID, _ := bMap["tool_use_id"].(string)
					rawContent := bMap["content"]
					var resStr string
					if s, ok := rawContent.(string); ok {
						resStr = s
					} else {
						cb, _ := json.Marshal(rawContent)
						resStr = string(cb)
					}
					openAIMessages = append(openAIMessages, models.OpenAIMessage{
						Role:       "tool",
						ToolCallID: toolCallID,
						Content:    resStr,
					})
				}
			}

			if len(contentParts) > 0 || len(toolCalls) > 0 {
				openAIMessages = append(openAIMessages, models.OpenAIMessage{
					Role:      role,
					Content:   contentParts,
					ToolCalls: toolCalls,
				})
			}
		}
	}

	// 3. Convert Tools
	var openAITools []models.OpenAITool
	for _, t := range req.Tools {
		openAITools = append(openAITools, models.OpenAITool{
			Type: "function",
			Function: models.OpenAIFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.InputSchema,
			},
		})
	}

	reasoningEffort := ""
	if req.Thinking != nil && req.Thinking.Type == "enabled" {
		reasoningEffort = "high"
	}

	return models.OpenAIChatRequest{
		Model:           req.Model,
		Messages:        openAIMessages,
		Tools:           openAITools,
		ToolChoice:      req.ToolChoice,
		ReasoningEffort: reasoningEffort,
		Stream:          req.Stream,
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
