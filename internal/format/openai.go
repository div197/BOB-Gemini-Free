package format

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/div197/bob-gemini-free/internal/models"
)

func RandHex(n int) string {
	bytes := make([]byte, (n+1)/2)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)[:n]
}

func BuildToolChoiceInstruction(toolChoice any) string {
	if strChoice, ok := toolChoice.(string); ok {
		if strChoice == "none" {
			return "\n\nIMPORTANT: Do NOT call any tools. Respond with text only."
		}
		if strChoice == "required" {
			return "\n\nIMPORTANT: You MUST call at least one tool. Do not respond with text only."
		}
	} else if mapChoice, ok := toolChoice.(map[string]any); ok {
		if fn, ok := mapChoice["function"].(map[string]any); ok {
			if name, ok := fn["name"].(string); ok && name != "" {
				return fmt.Sprintf("\n\nIMPORTANT: You MUST call the tool \"%s\". Do not call other tools.", name)
			}
		}
	}
	return ""
}

func MessagesToPromptAndImages(req models.OpenAIChatRequest) (string, []Image, error) {
	var parts []string
	var images []Image

	strChoice, isStr := req.ToolChoice.(string)
	if !(isStr && strChoice == "none") && len(req.Tools) > 0 {
		var toolDefs []models.OpenAIFunction
		for _, tool := range req.Tools {
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
			if fn.Parameters == nil {
				fn.Parameters = map[string]any{}
			}

			toolDefs = append(toolDefs, fn)
		}

		if len(toolDefs) > 0 {
			constraint := BuildToolChoiceInstruction(req.ToolChoice)
			defsJSON, _ := json.Marshal(toolDefs)
			parts = append(parts, fmt.Sprintf(
				"# Tool Use\n\n"+
					"You can call the following tools. Call format:\n"+
					"```tool_call\n{\"name\": \"func_name\", \"arguments\": {...}}\n```\n"+
					"When calling tools, output ONLY the tool_call block(s).\n\n"+
					"Available tools:\n%s%s",
				string(defsJSON),
				constraint,
			))
		}
	}

	for _, msg := range req.Messages {
		role := msg.Role
		if role == "" {
			role = "user"
		}

		var contentStr string
		if strContent, ok := msg.Content.(string); ok {
			contentStr = strContent
		} else if contentList, ok := msg.Content.([]any); ok {
			var textParts []string
			for _, item := range contentList {
				if mapItem, ok := item.(map[string]any); ok {
					iType, _ := mapItem["type"].(string)
					if iType == "text" || iType == "input_text" {
						if t, ok := mapItem["text"].(string); ok {
							textParts = append(textParts, t)
						}
					} else if iType == "image_url" || iType == "image" {
						var urlStr string
						if imgURLMap, ok := mapItem["image_url"].(map[string]any); ok {
							urlStr, _ = imgURLMap["url"].(string)
						} else if strURL, ok := mapItem["image_url"].(string); ok {
							urlStr = strURL
						} else if strURL, ok := mapItem["url"].(string); ok {
							urlStr = strURL
						}

						if strings.HasPrefix(urlStr, "data:") {
							commaIdx := strings.Index(urlStr, ",")
							if commaIdx != -1 {
								meta := urlStr[:commaIdx]
								b64Data := urlStr[commaIdx+1:]
								mime := "image/png"
								if semiIdx := strings.Index(meta, ";"); semiIdx != -1 {
									mime = strings.TrimPrefix(meta[:semiIdx], "data:")
								}
								if dec, err := base64.StdEncoding.DecodeString(b64Data); err == nil && len(dec) > 0 {
									images = append(images, Image{Data: dec, MIME: mime})
								}
							}
						} else if dec, err := base64.StdEncoding.DecodeString(urlStr); err == nil && len(dec) > 0 {
							images = append(images, Image{Data: dec, MIME: "image/png"})
						}
					}
				}
			}
			contentStr = strings.Join(textParts, " ")
		}

		switch role {
		case "system", "developer":
			parts = append(parts, fmt.Sprintf("[System instruction]: %s", contentStr))
		case "assistant":
			if len(msg.ToolCalls) > 0 {
				var tcStrs []string
				for _, tc := range msg.ToolCalls {
					argsStr := tc.Function.Arguments
					if argsStr == "" {
						argsStr = "{}"
					}
					tcStrs = append(tcStrs, fmt.Sprintf("```tool_call\n{\"name\": \"%s\", \"arguments\": %s}\n```", tc.Function.Name, argsStr))
				}
				parts = append(parts, fmt.Sprintf("[Assistant]: %s\n%s", contentStr, strings.Join(tcStrs, "\n")))
			} else {
				parts = append(parts, fmt.Sprintf("[Assistant]: %s", contentStr))
			}
		case "tool":
			identifier := msg.Name
			if identifier == "" {
				identifier = msg.ToolCallID
			}
			if identifier == "" {
				identifier = "tool"
			}
			parts = append(parts, fmt.Sprintf("[Tool result for %s]: %s", identifier, contentStr))
		default:
			if contentStr != "" {
				parts = append(parts, contentStr)
			}
		}
	}

	if req.ResponseFormat != nil && (req.ResponseFormat.Type == "json_object" || req.ResponseFormat.Type == "json_schema") {
		parts = append(parts, "[System instruction]: You must respond strictly with valid JSON output.")
	}

	return strings.Join(parts, "\n\n"), images, nil
}

func MessagesToPrompt(req models.OpenAIChatRequest) (string, error) {
	prompt, _, err := MessagesToPromptAndImages(req)
	return prompt, err
}

var reToolCall = regexp.MustCompile(`(?s)\x60\x60\x60tool_call\s*\n(.*?)\n\x60\x60\x60`)

// ParseToolCalls extracts tool calls from the raw string output of the model.
// The model is instructed to output tool calls in markdown blocks e.g. ```tool_call\n{...}\n```.
// This function parses those blocks, removes them from the text, and returns the cleaned text along with structured ToolCalls.
func ParseToolCalls(text string) (string, []models.OpenAIToolCall) {
	var toolCalls []models.OpenAIToolCall
	matches := reToolCall.FindAllStringSubmatchIndex(text, -1)
	if len(matches) == 0 {
		return text, nil
	}

	var cleanParts []string
	lastEnd := 0

	for _, m := range matches {
		cleanParts = append(cleanParts, text[lastEnd:m[0]])
		lastEnd = m[1]

		content := strings.TrimSpace(text[m[2]:m[3]])
		var data struct {
			Name      string `json:"name"`
			Arguments any    `json:"arguments"`
		}
		if err := json.Unmarshal([]byte(content), &data); err == nil && data.Name != "" {
			var argsStr string
			if str, ok := data.Arguments.(string); ok {
				argsStr = str
			} else if data.Arguments != nil {
				b, _ := json.Marshal(data.Arguments)
				argsStr = string(b)
			} else {
				argsStr = "{}"
			}

			toolCalls = append(toolCalls, models.OpenAIToolCall{
				ID:   fmt.Sprintf("call_%s", RandHex(8)),
				Type: "function",
				Function: models.OpenAIToolCallFunction{
					Name:      data.Name,
					Arguments: argsStr,
				},
			})
		}
	}

	cleanParts = append(cleanParts, text[lastEnd:])
	clean := strings.TrimSpace(strings.Join(cleanParts, ""))
	return clean, toolCalls
}
