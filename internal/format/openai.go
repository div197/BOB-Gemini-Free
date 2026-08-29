package format

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/div197/bob-gemini-free/internal/models"
)

const (
	MaxToolDefinitions      = 64
	MaxToolDefinitionBytes  = 1 << 20
	MaxToolArgumentBytes    = 1 << 20
	MaxToolNameBytes        = 256
	MaxToolDescriptionBytes = 64 << 10
)

// ValidateToolCall bounds and validates a structured tool call before it is
// forwarded to another adapter or emitted in a response. Provider responses
// are untrusted input too: a malformed function call must not be turned into
// an apparently successful OpenAI/Anthropic response.
func ValidateToolCall(name, arguments string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("tool call name is empty")
	}
	if len(name) > MaxToolNameBytes {
		return fmt.Errorf("tool call name exceeds %d bytes", MaxToolNameBytes)
	}
	arguments = strings.TrimSpace(arguments)
	if arguments == "" {
		arguments = "{}"
	}
	if len(arguments) > MaxToolArgumentBytes {
		return fmt.Errorf("tool call %q arguments exceed %d bytes", name, MaxToolArgumentBytes)
	}
	if !json.Valid([]byte(arguments)) {
		return fmt.Errorf("tool call %q has invalid JSON arguments", name)
	}
	return nil
}

// ValidateToolChoice rejects values that would otherwise silently fall back to
// automatic tool selection. A named choice must reference a declared tool so
// prompt and direct-provider adapters cannot claim to honor a tool that does
// not exist in the request.
func ValidateToolChoice(choice any, tools []models.OpenAITool) error {
	switch value := choice.(type) {
	case nil:
		return nil
	case string:
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "", "auto", "none", "required", "any":
			return nil
		default:
			return fmt.Errorf("unsupported tool_choice %q", value)
		}
	case map[string]any:
		rawType, typePresent := value["type"]
		if typePresent {
			typeName, ok := rawType.(string)
			if !ok {
				return errors.New("tool_choice type must be a string")
			}
			if typeName != "" && !strings.EqualFold(typeName, "function") {
				return fmt.Errorf("unsupported tool_choice type %q", typeName)
			}
		}
		rawFunction, functionPresent := value["function"]
		if !functionPresent {
			return errors.New("tool_choice function is missing")
		}
		function, ok := rawFunction.(map[string]any)
		if !ok {
			return errors.New("tool_choice function must be an object")
		}
		rawName, namePresent := function["name"]
		if !namePresent {
			return errors.New("tool_choice function name is missing")
		}
		name, ok := rawName.(string)
		if !ok || strings.TrimSpace(name) == "" {
			return errors.New("tool_choice function name is empty")
		}
		if len(name) > MaxToolNameBytes {
			return fmt.Errorf("tool_choice function name exceeds %d bytes", MaxToolNameBytes)
		}
		for _, tool := range tools {
			fn := tool.Function
			if fn.Name == "" {
				fn.Name = tool.Name
			}
			if fn.Name == name {
				return nil
			}
		}
		return fmt.Errorf("tool_choice references undeclared tool %q", name)
	default:
		return fmt.Errorf("unsupported tool_choice type %T", choice)
	}
}

func RandHex(n int) string {
	if n <= 0 {
		return ""
	}
	var hexBuf [32]byte
	if n > len(hexBuf) {
		n = len(hexBuf)
	}
	var b [16]byte // 16 bytes = 32 hex chars max
	reqBytes := (n + 1) / 2
	_, _ = rand.Read(b[:reqBytes])
	hex.Encode(hexBuf[:], b[:reqBytes])
	return string(hexBuf[:n])
}

func BuildToolChoiceInstruction(toolChoice any) string {
	if strChoice, ok := toolChoice.(string); ok {
		switch strings.ToLower(strings.TrimSpace(strChoice)) {
		case "none":
			return "\n\nIMPORTANT: Do NOT call any tools. Respond with text only."
		case "required", "any":
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

// ValidateToolResultReferences makes tool continuations explicit before they
// are flattened into a prompt or translated into Gemini function responses.
// An uncorrelated result can otherwise be attached to the wrong call, which
// is especially dangerous when multiple or parallel tools are in flight.
func ValidateToolResultReferences(messages []models.OpenAIMessage) error {
	callsByID := make(map[string]string)
	nameCounts := make(map[string]int)
	for _, msg := range messages {
		switch strings.ToLower(strings.TrimSpace(msg.Role)) {
		case "assistant":
			for _, call := range msg.ToolCalls {
				id := strings.TrimSpace(call.ID)
				name := strings.TrimSpace(call.Function.Name)
				if name == "" {
					return errors.New("assistant tool call is missing a name")
				}
				if id != "" {
					if previous, exists := callsByID[id]; exists && previous != name {
						return fmt.Errorf("tool call ID %q changed name", id)
					}
					callsByID[id] = name
				}
				nameCounts[name]++
			}
		case "tool":
			id := strings.TrimSpace(msg.ToolCallID)
			name := strings.TrimSpace(msg.Name)
			if id == "" && name == "" {
				return errors.New("tool result is missing tool_call_id")
			}
			if id != "" {
				expectedName, exists := callsByID[id]
				if !exists {
					return fmt.Errorf("tool result references unknown tool_call_id %q", id)
				}
				if name != "" && name != expectedName {
					return fmt.Errorf("tool result %q does not match tool call %q", name, id)
				}
				continue
			}
			if nameCounts[name] == 0 {
				return fmt.Errorf("tool result references unknown tool %q", name)
			}
			if nameCounts[name] != 1 {
				return fmt.Errorf("tool result for tool %q is ambiguous; provide tool_call_id", name)
			}
		}
	}
	return nil
}

func MessagesToPromptAndImages(req models.OpenAIChatRequest) (string, []Image, error) {
	if err := ValidateToolChoice(req.ToolChoice, req.Tools); err != nil {
		return "", nil, err
	}
	if err := ValidateToolResultReferences(req.Messages); err != nil {
		return "", nil, err
	}
	var parts []string
	var images []Image

	strChoice, isStr := req.ToolChoice.(string)
	if !(isStr && strings.EqualFold(strings.TrimSpace(strChoice), "none")) && len(req.Tools) > 0 {
		if len(req.Tools) > MaxToolDefinitions {
			return "", nil, fmt.Errorf("tool definition count exceeds %d", MaxToolDefinitions)
		}
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
			if strings.TrimSpace(fn.Name) == "" {
				return "", nil, fmt.Errorf("tool name is empty")
			}
			if len(fn.Name) > MaxToolNameBytes {
				return "", nil, fmt.Errorf("tool %q name exceeds %d bytes", fn.Name[:MaxToolNameBytes], MaxToolNameBytes)
			}
			if len(fn.Description) > MaxToolDescriptionBytes {
				return "", nil, fmt.Errorf("tool %q description exceeds %d bytes", fn.Name, MaxToolDescriptionBytes)
			}
			if err := ValidateToolSchema(fn.Parameters); err != nil {
				return "", nil, fmt.Errorf("tool %q schema: %w", fn.Name, err)
			}

			toolDefs = append(toolDefs, fn)
		}

		if len(toolDefs) > 0 {
			constraint := BuildToolChoiceInstruction(req.ToolChoice)
			defsJSON, err := json.Marshal(toolDefs)
			if err != nil {
				return "", nil, fmt.Errorf("encode tool definitions: %w", err)
			}
			if len(defsJSON) > MaxToolDefinitionBytes {
				return "", nil, fmt.Errorf("tool definitions exceed %d bytes", MaxToolDefinitionBytes)
			}
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
		role := strings.ToLower(strings.TrimSpace(msg.Role))
		if role == "" {
			role = "user"
		}

		var contentStr string
		switch value := msg.Content.(type) {
		case nil:
			// Empty content is valid for assistant tool-call turns and is
			// represented as an empty prompt segment for other roles.
		case string:
			contentStr = value
		case []any:
			contentList := value
			var textParts []string
			for itemIndex, item := range contentList {
				mapItem, ok := item.(map[string]any)
				if !ok {
					return "", nil, fmt.Errorf("unsupported multimodal content item %d type %T", itemIndex, item)
				}
				rawType, exists := mapItem["type"]
				if !exists {
					return "", nil, fmt.Errorf("multimodal content item %d is missing type", itemIndex)
				}
				iType, ok := rawType.(string)
				if !ok || strings.TrimSpace(iType) == "" {
					return "", nil, fmt.Errorf("multimodal content item %d type must be a non-empty string", itemIndex)
				}
				switch iType {
				case "text", "input_text":
					rawText, exists := mapItem["text"]
					if !exists {
						return "", nil, fmt.Errorf("multimodal content item %d text is missing", itemIndex)
					}
					text, ok := rawText.(string)
					if !ok {
						return "", nil, fmt.Errorf("multimodal content item %d text must be a string", itemIndex)
					}
					textParts = append(textParts, text)
				case "image_url", "image":
					var urlStr string
					if imgURLMap, ok := mapItem["image_url"].(map[string]any); ok {
						urlStr, _ = imgURLMap["url"].(string)
					} else if strURL, ok := mapItem["image_url"].(string); ok {
						urlStr = strURL
					} else if strURL, ok := mapItem["url"].(string); ok {
						urlStr = strURL
					}

					if urlStr == "" {
						return "", nil, fmt.Errorf("image content item %d has no image URL", itemIndex)
					}
					if strings.HasPrefix(strings.ToLower(urlStr), "data:") {
						commaIdx := strings.Index(urlStr, ",")
						if commaIdx == -1 {
							return "", nil, fmt.Errorf("image data URL is missing its payload")
						}
						meta := urlStr[:commaIdx]
						metaParts := strings.Split(meta, ";")
						if len(metaParts) < 2 || !strings.EqualFold(strings.TrimSpace(metaParts[len(metaParts)-1]), "base64") {
							return "", nil, fmt.Errorf("image data URL must use base64 encoding")
						}
						mimeType := strings.TrimPrefix(strings.TrimSpace(metaParts[0]), "data:")
						dec, normalizedMIME, err := decodeInlineImageData(urlStr[commaIdx+1:], mimeType)
						if err != nil {
							return "", nil, fmt.Errorf("invalid image data URL: %w", err)
						}
						images = append(images, Image{Data: dec, MIME: normalizedMIME})
					} else if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") {
						images = append(images, Image{URL: urlStr})
					} else if strings.Contains(urlStr, "://") {
						images = append(images, Image{URL: urlStr})
					} else {
						dec, mimeType, err := decodeInlineImageData(urlStr, "image/png")
						if err != nil {
							return "", nil, fmt.Errorf("invalid inline image data: %w", err)
						}
						images = append(images, Image{Data: dec, MIME: mimeType})
					}
				default:
					return "", nil, fmt.Errorf("unsupported multimodal content type %q", iType)
				}
			}
			contentStr = strings.Join(textParts, " ")
		default:
			return "", nil, fmt.Errorf("message content must be a string, array, or null; got %T", msg.Content)
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
					if err := ValidateToolCall(tc.Function.Name, argsStr); err != nil {
						return "", nil, err
					}
					tcStrs = append(tcStrs, fmt.Sprintf("```tool_call\n{\"name\": \"%s\", \"arguments\": %s}\n```", tc.Function.Name, argsStr))
				}
				parts = append(parts, fmt.Sprintf("[Assistant]: %s\n%s", contentStr, strings.Join(tcStrs, "\n")))
			} else {
				parts = append(parts, fmt.Sprintf("[Assistant]: %s", contentStr))
			}
		case "user":
			parts = append(parts, contentStr)
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
			return "", nil, fmt.Errorf("unsupported OpenAI message role %q", msg.Role)
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

var reToolCall = regexp.MustCompile(`(?s)\x60\x60\x60(?:tool_call|function_call)\s*\n(.*?)\n\x60\x60\x60`)

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
		rawBlock := text[m[0]:m[1]]
		parsed := false

		content := strings.TrimSpace(text[m[2]:m[3]])
		var data struct {
			Name      string `json:"name"`
			Arguments any    `json:"arguments"`
		}
		if err := json.Unmarshal([]byte(content), &data); err == nil && data.Name != "" && len(data.Name) <= MaxToolNameBytes {
			var argsStr string
			if str, ok := data.Arguments.(string); ok {
				argsStr = strings.TrimSpace(str)
				if argsStr == "" {
					argsStr = "{}"
				}
			} else if data.Arguments != nil {
				if b, err := json.Marshal(data.Arguments); err == nil {
					argsStr = string(b)
				}
			} else {
				argsStr = "{}"
			}
			if len(argsStr) <= MaxToolArgumentBytes && json.Valid([]byte(argsStr)) {
				toolCalls = append(toolCalls, models.OpenAIToolCall{
					ID:   fmt.Sprintf("call_%s", RandHex(8)),
					Type: "function",
					Function: models.OpenAIToolCallFunction{
						Name:      data.Name,
						Arguments: argsStr,
					},
				})
				parsed = true
			}
		}
		if !parsed {
			// A malformed or oversized fence is ordinary model text until it
			// has passed all validation. Never silently delete user-visible
			// Markdown that was not converted into a structured call.
			cleanParts = append(cleanParts, rawBlock)
		}
	}

	cleanParts = append(cleanParts, text[lastEnd:])
	clean := strings.TrimSpace(strings.Join(cleanParts, ""))
	return clean, toolCalls
}
