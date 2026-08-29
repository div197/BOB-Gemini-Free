package format

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/div197/bob-gemini-free/internal/models"
)

type Image struct {
	Data []byte
	MIME string
	URL  string
}

// GoogleFunctionCallingMode normalizes the public Google function-calling
// mode and rejects values that the prompt adapter cannot honor. An invalid
// mode must not silently become AUTO.
func GoogleFunctionCallingMode(req models.GoogleGenerateRequest) (string, error) {
	mode := "AUTO"
	var allowed []string
	if req.ToolConfig != nil && req.ToolConfig.FunctionCallingConfig != nil {
		config := req.ToolConfig.FunctionCallingConfig
		if strings.TrimSpace(config.Mode) != "" {
			mode = strings.ToUpper(strings.TrimSpace(config.Mode))
		}
		allowed = config.AllowedFunctionNames
	}
	switch mode {
	case "AUTO", "NONE", "ANY":
	default:
		return "", fmt.Errorf("unsupported Google function-calling mode %q", mode)
	}
	if len(allowed) > MaxToolDefinitions {
		return "", fmt.Errorf("allowed function name count exceeds %d", MaxToolDefinitions)
	}
	if len(allowed) == 0 {
		return mode, nil
	}
	if mode != "ANY" {
		return "", fmt.Errorf("allowed function names require Google function-calling mode ANY")
	}
	declared := make(map[string]struct{})
	for _, group := range req.Tools {
		for _, declaration := range group.FunctionDeclarations {
			declared[declaration.Name] = struct{}{}
		}
	}
	seen := make(map[string]struct{}, len(allowed))
	for _, name := range allowed {
		if strings.TrimSpace(name) == "" {
			return "", fmt.Errorf("allowed function name is empty")
		}
		if len(name) > MaxToolNameBytes {
			return "", fmt.Errorf("allowed function name exceeds %d bytes", MaxToolNameBytes)
		}
		if _, duplicate := seen[name]; duplicate {
			return "", fmt.Errorf("allowed function name %q is duplicated", name)
		}
		seen[name] = struct{}{}
		if _, ok := declared[name]; !ok {
			return "", fmt.Errorf("allowed function name %q is not declared", name)
		}
	}
	return mode, nil
}

func BuildToolPrompt(defs []models.GoogleFunctionDeclaration) string {
	specBytes, _ := json.Marshal(defs)
	return fmt.Sprintf(
		"# Tool Use\n\n"+
			"You can call the following tools to help accomplish tasks. "+
			"These tools connect to the user's local environment and will execute when called.\n\n"+
			"Call format (use this exact format):\n"+
			"```function_call\n"+
			"{\"name\": \"<tool_name>\", \"args\": {<arguments>}}\n"+
			"```\n\n"+
			"When calling tools:\n"+
			"- Output ONLY the function_call block(s), nothing else\n"+
			"- You may call multiple tools with multiple blocks\n"+
			"- After receiving a [Tool result for ...], use that data to answer the user\n\n"+
			"Available tools:\n%s",
		string(specBytes),
	)
}

func GoogleToolChoiceInstruction(req models.GoogleGenerateRequest) string {
	mode, err := GoogleFunctionCallingMode(req)
	if err != nil {
		return ""
	}

	if mode == "NONE" {
		return "\n\nIMPORTANT: Do NOT call any tools. Respond with text only."
	}
	if mode == "ANY" {
		if req.ToolConfig != nil && req.ToolConfig.FunctionCallingConfig != nil && len(req.ToolConfig.FunctionCallingConfig.AllowedFunctionNames) > 0 {
			var names []string
			for _, s := range req.ToolConfig.FunctionCallingConfig.AllowedFunctionNames {
				encoded, _ := json.Marshal(s)
				names = append(names, string(encoded))
			}
			if len(names) > 0 {
				return fmt.Sprintf("\n\nIMPORTANT: You MUST call one of these tools: %s. Do not respond with text only.", strings.Join(names, ", "))
			}
		}
		return "\n\nIMPORTANT: You MUST call at least one tool. Do not respond with text only."
	}
	return ""
}

func GoogleContentsToPrompt(req models.GoogleGenerateRequest) (string, []Image, error) {
	fcMode, err := GoogleFunctionCallingMode(req)
	if err != nil {
		return "", nil, err
	}
	var parts []string
	var images []Image

	var toolDefs []models.GoogleFunctionDeclaration
	if fcMode != "NONE" {
		for _, toolGroup := range req.Tools {
			toolDefs = append(toolDefs, toolGroup.FunctionDeclarations...)
		}
		if len(toolDefs) > MaxToolDefinitions {
			return "", nil, fmt.Errorf("tool definition count exceeds %d", MaxToolDefinitions)
		}
		for _, tool := range toolDefs {
			name := strings.TrimSpace(tool.Name)
			if name == "" {
				return "", nil, fmt.Errorf("tool name is empty")
			}
			if name != tool.Name {
				return "", nil, fmt.Errorf("tool %q name has surrounding whitespace", tool.Name)
			}
			if len(tool.Name) > MaxToolNameBytes {
				return "", nil, fmt.Errorf("tool %q name exceeds %d bytes", tool.Name[:MaxToolNameBytes], MaxToolNameBytes)
			}
			if len(tool.Description) > MaxToolDescriptionBytes {
				return "", nil, fmt.Errorf("tool %q description exceeds %d bytes", tool.Name, MaxToolDescriptionBytes)
			}
			if err := ValidateToolSchema(tool.Parameters); err != nil {
				return "", nil, fmt.Errorf("tool %q schema: %w", tool.Name, err)
			}
		}
		seenTools := make(map[string]struct{}, len(toolDefs))
		for _, tool := range toolDefs {
			if _, exists := seenTools[tool.Name]; exists {
				return "", nil, fmt.Errorf("tool %q is declared more than once", tool.Name)
			}
			seenTools[tool.Name] = struct{}{}
		}
		encoded, err := json.Marshal(toolDefs)
		if err != nil {
			return "", nil, fmt.Errorf("encode tool definitions: %w", err)
		}
		if len(encoded) > MaxToolDefinitionBytes {
			return "", nil, fmt.Errorf("tool definitions exceed %d bytes", MaxToolDefinitionBytes)
		}
	}

	var sysText string
	if req.SystemInstruction != nil {
		var tParts []string
		for index, p := range req.SystemInstruction.Parts {
			kind, partErr := googlePartKind(p)
			if partErr != nil {
				return "", nil, fmt.Errorf("system instruction part %d: %w", index, partErr)
			}
			if kind != "text" {
				return "", nil, fmt.Errorf("system instruction part %d is %s; only text is supported", index, kind)
			}
			tParts = append(tParts, p.Text)
		}
		sysText = strings.Join(tParts, " ")
	}

	if sysText != "" {
		if len(toolDefs) > 0 {
			constraint := GoogleToolChoiceInstruction(req)
			parts = append(parts, sysText+"\n\n"+BuildToolPrompt(toolDefs)+constraint)
		} else {
			parts = append(parts, sysText)
		}
	} else if len(toolDefs) > 0 {
		constraint := GoogleToolChoiceInstruction(req)
		parts = append(parts, BuildToolPrompt(toolDefs)+constraint)
	}

	for _, content := range req.Contents {
		role := strings.ToLower(strings.TrimSpace(content.Role))
		if role == "" {
			role = "user"
		}
		switch role {
		case "user", "model", "function", "tool":
		default:
			return "", nil, fmt.Errorf("unsupported Google content role %q", content.Role)
		}

		var msgParts []string
		for index, p := range content.Parts {
			kind, partErr := googlePartKind(p)
			if partErr != nil {
				return "", nil, fmt.Errorf("content role %q part %d: %w", role, index, partErr)
			}
			switch kind {
			case "text":
				msgParts = append(msgParts, p.Text)
			case "inlineData":
				dec, mimeType, err := decodeInlineImageData(p.InlineData.Data, p.InlineData.MIMEType)
				if err != nil {
					return "", nil, fmt.Errorf("invalid inline image: %w", err)
				}
				images = append(images, Image{Data: dec, MIME: mimeType})
			case "functionCall":
				if err := validateGoogleFunctionCall(p.FunctionCall); err != nil {
					return "", nil, fmt.Errorf("content role %q part %d: %w", role, index, err)
				}
				args := p.FunctionCall.Args
				if args == nil {
					args = map[string]any{}
				}
				payload, err := json.Marshal(map[string]any{"name": p.FunctionCall.Name, "args": args})
				if err != nil {
					return "", nil, fmt.Errorf("encode function call %q: %w", p.FunctionCall.Name, err)
				}
				msgParts = append(msgParts, fmt.Sprintf("```function_call\n%s\n```", string(payload)))
			case "functionResponse":
				if err := validateGoogleFunctionResponse(p.FunctionResponse); err != nil {
					return "", nil, fmt.Errorf("content role %q part %d: %w", role, index, err)
				}
				resp := p.FunctionResponse.Response
				if resp == nil {
					resp = map[string]any{}
				}
				payload, err := json.Marshal(resp)
				if err != nil {
					return "", nil, fmt.Errorf("encode function response %q: %w", p.FunctionResponse.Name, err)
				}
				msgParts = append(msgParts, fmt.Sprintf("[Tool result for %s]: %s", p.FunctionResponse.Name, string(payload)))
			}
		}

		text := strings.Join(msgParts, "\n")
		if role == "model" {
			parts = append(parts, fmt.Sprintf("[Assistant]: %s", text))
		} else {
			parts = append(parts, text)
		}
	}

	return strings.Join(parts, "\n\n"), images, nil
}

func googlePartKind(part models.GooglePart) (string, error) {
	count := 0
	kind := ""
	if part.Text != "" {
		count++
		kind = "text"
	}
	if part.InlineData != nil {
		count++
		kind = "inlineData"
	}
	if part.FunctionCall != nil {
		count++
		kind = "functionCall"
	}
	if part.FunctionResponse != nil {
		count++
		kind = "functionResponse"
	}
	if count == 0 {
		return "", fmt.Errorf("part has no supported value")
	}
	if count > 1 {
		return "", fmt.Errorf("part has multiple values")
	}
	return kind, nil
}

func validateGoogleFunctionCall(call *models.GoogleFunctionCall) error {
	if call == nil {
		return fmt.Errorf("function call is missing")
	}
	if strings.TrimSpace(call.Name) == "" {
		return fmt.Errorf("function call name is empty")
	}
	if strings.TrimSpace(call.Name) != call.Name {
		return fmt.Errorf("function call name has surrounding whitespace")
	}
	if len(call.Name) > MaxToolNameBytes {
		return fmt.Errorf("function call name exceeds %d bytes", MaxToolNameBytes)
	}
	args := call.Args
	if args == nil {
		args = map[string]any{}
	}
	encoded, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("encode function call %q: %w", call.Name, err)
	}
	if len(encoded) > MaxToolArgumentBytes {
		return fmt.Errorf("function call %q arguments exceed %d bytes", call.Name, MaxToolArgumentBytes)
	}
	return nil
}

func validateGoogleFunctionResponse(response *models.GoogleFunctionCallResp) error {
	if response == nil {
		return fmt.Errorf("function response is missing")
	}
	if strings.TrimSpace(response.Name) == "" {
		return fmt.Errorf("function response name is empty")
	}
	if strings.TrimSpace(response.Name) != response.Name {
		return fmt.Errorf("function response name has surrounding whitespace")
	}
	if len(response.Name) > MaxToolNameBytes {
		return fmt.Errorf("function response name exceeds %d bytes", MaxToolNameBytes)
	}
	result := response.Response
	if result == nil {
		result = map[string]any{}
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("encode function response %q: %w", response.Name, err)
	}
	if len(encoded) > MaxToolArgumentBytes {
		return fmt.Errorf("function response %q exceeds %d bytes", response.Name, MaxToolArgumentBytes)
	}
	return nil
}

var (
	reFencedGoogleCall   = regexp.MustCompile(`(?s)\x60\x60\x60function_call\s*\n(.*?)\n\x60\x60\x60`)
	reUnfencedGoogleCall = regexp.MustCompile(`(?s)(?:^|\n)function_call\s*\n(\{[^\x60]*?\})`)
)

func ParseGoogleFunctionCalls(text string) (string, []models.GoogleFunctionCall) {
	var calls []models.GoogleFunctionCall
	clean := text

	var fenced strings.Builder
	lastEnd := 0
	for _, m := range reFencedGoogleCall.FindAllStringSubmatchIndex(clean, -1) {
		fenced.WriteString(clean[lastEnd:m[0]])
		call, ok := parseGoogleFunctionCallPayload(strings.TrimSpace(clean[m[2]:m[3]]))
		if ok {
			calls = append(calls, call)
		} else {
			// Preserve invalid/ambiguous model Markdown instead of silently
			// deleting content that was not converted into a call.
			fenced.WriteString(clean[m[0]:m[1]])
		}
		lastEnd = m[1]
	}
	fenced.WriteString(clean[lastEnd:])
	clean = strings.TrimSpace(fenced.String())

	var unfenced strings.Builder
	lastEnd = 0
	for _, m := range reUnfencedGoogleCall.FindAllStringSubmatchIndex(clean, -1) {
		unfenced.WriteString(clean[lastEnd:m[0]])
		call, ok := parseGoogleFunctionCallPayload(strings.TrimSpace(clean[m[2]:m[3]]))
		if ok {
			calls = append(calls, call)
		} else {
			unfenced.WriteString(clean[m[0]:m[1]])
		}
		lastEnd = m[1]
	}
	unfenced.WriteString(clean[lastEnd:])
	clean = strings.TrimSpace(unfenced.String())

	if len(calls) == 0 && strings.HasPrefix(strings.TrimSpace(clean), "{") {
		if call, ok := parseGoogleFunctionCallPayload(strings.TrimSpace(clean)); ok {
			calls = append(calls, call)
			clean = ""
		}
	}

	return clean, calls
}

func parseGoogleFunctionCallPayload(payload string) (models.GoogleFunctionCall, bool) {
	var data map[string]any
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return models.GoogleFunctionCall{}, false
	}
	name, ok := data["name"].(string)
	if !ok || strings.TrimSpace(name) == "" || len(name) > MaxToolNameBytes {
		return models.GoogleFunctionCall{}, false
	}
	args := map[string]any{}
	rawArgs, hasArgs := data["args"]
	if !hasArgs {
		rawArgs, hasArgs = data["arguments"]
	}
	if hasArgs {
		var ok bool
		args, ok = rawArgs.(map[string]any)
		if !ok {
			return models.GoogleFunctionCall{}, false
		}
	}
	encoded, err := json.Marshal(args)
	if err != nil || len(encoded) > MaxToolArgumentBytes {
		return models.GoogleFunctionCall{}, false
	}
	return models.GoogleFunctionCall{Name: name, Args: args}, true
}
