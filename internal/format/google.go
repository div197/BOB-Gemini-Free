package format

import (
	"encoding/base64"
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
	mode := "AUTO"
	if req.ToolConfig != nil && req.ToolConfig.FunctionCallingConfig != nil {
		if req.ToolConfig.FunctionCallingConfig.Mode != "" {
			mode = req.ToolConfig.FunctionCallingConfig.Mode
		}
	}

	if mode == "NONE" {
		return "\n\nIMPORTANT: Do NOT call any tools. Respond with text only."
	}
	if mode == "ANY" {
		if req.ToolConfig != nil && req.ToolConfig.FunctionCallingConfig != nil && len(req.ToolConfig.FunctionCallingConfig.AllowedFunctionNames) > 0 {
			var names []string
			for _, s := range req.ToolConfig.FunctionCallingConfig.AllowedFunctionNames {
				names = append(names, fmt.Sprintf("\"%s\"", s))
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
	var parts []string
	var images []Image

	fcMode := "AUTO"
	if req.ToolConfig != nil && req.ToolConfig.FunctionCallingConfig != nil && req.ToolConfig.FunctionCallingConfig.Mode != "" {
		fcMode = req.ToolConfig.FunctionCallingConfig.Mode
	}

	var toolDefs []models.GoogleFunctionDeclaration
	if fcMode != "NONE" {
		for _, toolGroup := range req.Tools {
			toolDefs = append(toolDefs, toolGroup.FunctionDeclarations...)
		}
		if len(toolDefs) > MaxToolDefinitions {
			return "", nil, fmt.Errorf("tool definition count exceeds %d", MaxToolDefinitions)
		}
		for _, tool := range toolDefs {
			if strings.TrimSpace(tool.Name) == "" {
				return "", nil, fmt.Errorf("tool name is empty")
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
		for _, p := range req.SystemInstruction.Parts {
			if p.Text != "" {
				tParts = append(tParts, p.Text)
			}
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
		role := content.Role
		if role == "" {
			role = "user"
		}

		var msgParts []string
		for _, p := range content.Parts {
			if p.Text != "" {
				msgParts = append(msgParts, p.Text)
			} else if p.InlineData != nil {
				mime := p.InlineData.MIMEType
				if mime == "" {
					mime = "image/png"
				}
				if dec, err := base64.StdEncoding.DecodeString(p.InlineData.Data); err == nil {
					images = append(images, Image{Data: dec, MIME: mime})
				}
			} else if p.FunctionCall != nil {
				args := p.FunctionCall.Args
				if args == nil {
					args = map[string]any{}
				}
				payload, err := json.Marshal(map[string]any{"name": p.FunctionCall.Name, "args": args})
				if err != nil {
					return "", nil, fmt.Errorf("encode function call %q: %w", p.FunctionCall.Name, err)
				}
				msgParts = append(msgParts, fmt.Sprintf("```function_call\n%s\n```", string(payload)))
			} else if p.FunctionResponse != nil {
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
