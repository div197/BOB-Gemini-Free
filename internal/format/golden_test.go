package format

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/div197/bob-gemini-free/internal/models"
)

func collectSplitterOutput(splitter *ThinkingStreamSplitter, input string, oneByte bool) (string, string, bool) {
	var thinking, content strings.Builder
	transition := false
	feed := func(delta string) {
		for _, chunk := range splitter.Feed(delta) {
			if chunk.TransitionToContent {
				transition = true
			}
			switch chunk.Type {
			case DeltaThinking:
				thinking.WriteString(chunk.Text)
			case DeltaContent:
				content.WriteString(chunk.Text)
			}
		}
	}
	if oneByte {
		for i := 0; i < len(input); i++ {
			feed(input[i : i+1])
		}
	} else {
		feed(input)
	}
	for _, chunk := range splitter.Flush() {
		if chunk.TransitionToContent {
			transition = true
		}
		switch chunk.Type {
		case DeltaThinking:
			thinking.WriteString(chunk.Text)
		case DeltaContent:
			content.WriteString(chunk.Text)
		}
	}
	return thinking.String(), content.String(), transition
}

func TestGoldenThinkingSplitterSurvivesOneByteFenceBoundaries(t *testing.T) {
	input := "```thinking\nपहला विचार.\nदूसरा विचार.\n```उत्तर: नमस्ते 🌻"
	thinking, content, transition := collectSplitterOutput(NewThinkingStreamSplitter(), input, true)
	if thinking != "पहला विचार.\nदूसरा विचार." {
		t.Fatalf("thinking = %q", thinking)
	}
	if content != "उत्तर: नमस्ते 🌻" {
		t.Fatalf("content = %q", content)
	}
	if !transition {
		t.Fatal("expected transition from thinking to content")
	}
}

func TestGoldenThinkingSplitterFlushesMissingClosingFenceAsThinking(t *testing.T) {
	thinking, content, transition := collectSplitterOutput(NewThinkingStreamSplitter(), "```thought\nunfinished reasoning", true)
	if thinking != "unfinished reasoning" || content != "" {
		t.Fatalf("missing-close output = thinking %q, content %q", thinking, content)
	}
	if transition {
		t.Fatal("missing closing fence must not invent a content transition")
	}
}

func TestGoldenToolCallTortureFixtures(t *testing.T) {
	fixtures := []struct {
		name      string
		input     string
		wantCalls int
		wantNames []string
		preserve  string
	}{
		{
			name:      "one nested unicode call",
			input:     "```tool_call\n{\"name\":\"search\",\"arguments\":{\"query\":\"चाय\",\"items\":[{\"n\":1}],\"nullable\":null}}\n```",
			wantCalls: 1,
			wantNames: []string{"search"},
		},
		{
			name:      "multiple calls with text",
			input:     "before\n```tool_call\n{\"name\":\"one\",\"arguments\":{}}\n```\nafter\n```tool_call\n{\"name\":\"two\",\"arguments\":{\"values\":[1,2]}}\n```\nend",
			wantCalls: 2,
			wantNames: []string{"one", "two"},
			preserve:  "before",
		},
		{
			name:      "invalid json",
			input:     "before\n```tool_call\n{not valid}\n```\nafter",
			wantCalls: 0,
			preserve:  "after",
		},
		{
			name:      "accidental json markdown",
			input:     "```json\n{\"name\":\"not_a_tool\"}\n```",
			wantCalls: 0,
			preserve:  "```json",
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			clean, calls := ParseToolCalls(fixture.input)
			if len(calls) != fixture.wantCalls {
				t.Fatalf("calls = %#v, want %d", calls, fixture.wantCalls)
			}
			for i, name := range fixture.wantNames {
				if calls[i].Function.Name != name {
					t.Errorf("call %d name = %q, want %q", i, calls[i].Function.Name, name)
				}
				var args any
				if err := json.Unmarshal([]byte(calls[i].Function.Arguments), &args); err != nil {
					t.Errorf("call %d arguments are not JSON: %v", i, err)
				}
			}
			if fixture.preserve != "" && !strings.Contains(clean, fixture.preserve) {
				t.Errorf("clean text lost %q: %q", fixture.preserve, clean)
			}
		})
	}
}

func TestToolParsersPreserveRejectedBlocks(t *testing.T) {
	openAIInput := "before\n```tool_call\n{\"name\":\"lookup\",\"arguments\":\"not-json\"}\n```\nafter"
	clean, calls := ParseToolCalls(openAIInput)
	if len(calls) != 0 || clean != openAIInput {
		t.Fatalf("OpenAI rejected block changed: clean %q calls %#v", clean, calls)
	}

	googleInput := "before\n```function_call\n{\"name\":\"lookup\",\"args\":[] }\n```\nafter"
	clean, googleCalls := ParseGoogleFunctionCalls(googleInput)
	if len(googleCalls) != 0 || clean != googleInput {
		t.Fatalf("Google rejected block changed: clean %q calls %#v", clean, googleCalls)
	}
}

func TestGoldenToolChoiceModes(t *testing.T) {
	if !strings.Contains(BuildToolChoiceInstruction("required"), "MUST call") {
		t.Fatal("required tool choice lost its constraint")
	}
	if !strings.Contains(BuildToolChoiceInstruction(map[string]any{
		"function": map[string]any{"name": "lookup"},
	}), "lookup") {
		t.Fatal("specific tool choice lost the requested name")
	}
	if BuildToolChoiceInstruction("none") == "" {
		t.Fatal("none tool choice must produce an explicit constraint")
	}
}

func TestGoldenToolSchemaAndLargeArgumentFixtures(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"mode":     map[string]any{"type": "string", "enum": []string{"fast", "deep"}},
			"items":    map[string]any{"type": "array", "items": map[string]any{"type": "integer"}},
			"optional": map[string]any{"type": []any{"string", "null"}},
		},
		"required": []string{"mode"},
	}
	prompt, err := MessagesToPrompt(models.OpenAIChatRequest{
		Messages: []models.OpenAIMessage{{Role: "user", Content: "Use the schema"}},
		Tools: []models.OpenAITool{{Type: "function", Function: models.OpenAIFunction{
			Name: "schema_tool", Parameters: schema,
		}}},
	})
	if err != nil {
		t.Fatalf("schema prompt: %v", err)
	}
	for _, expected := range []string{"schema_tool", "enum", "fast", "items", "null"} {
		if !strings.Contains(prompt, expected) {
			t.Errorf("schema prompt lost %q: %s", expected, prompt)
		}
	}

	largeValue := strings.Repeat("अ", 128*1024)
	largeInput := "```tool_call\n{" + `"name":"large_tool","arguments":{"value":"` + largeValue + `"}}` + "\n```"
	clean, calls := ParseToolCalls(largeInput)
	if len(calls) != 1 || calls[0].Function.Name != "large_tool" {
		t.Fatalf("large argument calls = %#v", calls)
	}
	if clean != "" {
		t.Fatalf("large argument clean text = %q", clean)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(calls[0].Function.Arguments), &decoded); err != nil {
		t.Fatalf("large arguments JSON: %v", err)
	}
	if len(decoded["value"].(string)) != len(largeValue) {
		t.Fatalf("large argument length = %d, want %d", len(decoded["value"].(string)), len(largeValue))
	}
}

func TestGoldenGoogleFunctionCallTortureFixtures(t *testing.T) {
	input := "```function_call\n{\"name\":\"search\",\"args\":{\"q\":\"नमस्ते\",\"values\":[1,2]}}\n```\n```function_call\n{\"name\":\"open\",\"args\":{\"url\":\"https://example.test\"}}\n```"
	clean, calls := ParseGoogleFunctionCalls(input)
	if len(calls) != 2 || calls[0].Name != "search" || calls[1].Name != "open" {
		t.Fatalf("Google calls = %#v", calls)
	}
	if clean != "" {
		t.Fatalf("clean Google call text = %q", clean)
	}

	invalid := "```function_call\n{invalid}\n```"
	clean, calls = ParseGoogleFunctionCalls(invalid)
	if len(calls) != 0 || clean != invalid {
		t.Fatalf("invalid Google call result = clean %q calls %#v", clean, calls)
	}
	accidental := "```json\n{\"name\":\"not_a_function_call\"}\n```"
	clean, calls = ParseGoogleFunctionCalls(accidental)
	if len(calls) != 0 || !strings.Contains(clean, "```json") {
		t.Fatalf("accidental Google markdown changed: clean %q calls %#v", clean, calls)
	}
}

func TestGoldenAdaptersPreserveEquivalentSemanticFixture(t *testing.T) {
	semantic := "नमस्ते — cite https://example.test/source\n```go\nfmt.Println(\"hello\")\n```"

	openPrompt, err := MessagesToPrompt(models.OpenAIChatRequest{
		Messages: []models.OpenAIMessage{{Role: "user", Content: semantic}},
	})
	if err != nil {
		t.Fatalf("OpenAI adapter: %v", err)
	}
	anthropicPrompt, err := MessagesToPrompt(AnthropicToOpenAIChatRequest(models.AnthropicMessagesRequest{
		Messages: []models.AnthropicMessage{{Role: "user", Content: semantic}},
	}))
	if err != nil {
		t.Fatalf("Anthropic adapter: %v", err)
	}
	googlePrompt, _, err := GoogleContentsToPrompt(models.GoogleGenerateRequest{
		Contents: []models.GoogleContent{{Role: "user", Parts: []models.GooglePart{{Text: semantic}}}},
	})
	if err != nil {
		t.Fatalf("Google adapter: %v", err)
	}
	for name, prompt := range map[string]string{
		"openai": openPrompt, "anthropic": anthropicPrompt, "google": googlePrompt,
	} {
		if !strings.Contains(prompt, semantic) {
			t.Errorf("%s adapter changed semantic fixture: %q", name, prompt)
		}
	}

	if !reflect.DeepEqual(openPrompt, anthropicPrompt) {
		t.Logf("OpenAI and Anthropic prompts intentionally differ only in adapter framing")
	}
}
