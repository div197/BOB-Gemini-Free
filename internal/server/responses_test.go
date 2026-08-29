package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/format"
	"github.com/div197/bob-gemini-free/internal/models"
)

func TestResponsesInputMapToOpenAIMessagePreservesToolFields(t *testing.T) {
	input := []any{
		map[string]any{
			"type":      "function_call",
			"call_id":   "call_lookup",
			"name":      "lookup",
			"arguments": `{"query":"भारत"}`,
		},
		map[string]any{
			"type":    "function_call_output",
			"call_id": "call_lookup",
			"name":    "lookup",
			"output":  `{"result":"ok"}`,
		},
	}
	responseMessages, err := format.ResponsesInputToMessages(input, "")
	if err != nil {
		t.Fatalf("ResponsesInputToMessages error: %v", err)
	}

	messages := make([]models.OpenAIMessage, 0, len(responseMessages))
	for _, responseMessage := range responseMessages {
		message, err := responsesInputMapToOpenAIMessage(responseMessage)
		if err != nil {
			t.Fatalf("responsesInputMapToOpenAIMessage error: %v", err)
		}
		messages = append(messages, message)
	}

	if len(messages) != 2 || len(messages[0].ToolCalls) != 1 {
		t.Fatalf("tool call fields were lost during server conversion: %#v", messages)
	}
	call := messages[0].ToolCalls[0]
	if call.ID != "call_lookup" || call.Function.Name != "lookup" || call.Function.Arguments != `{"query":"भारत"}` {
		t.Fatalf("tool call fields were not preserved: %#v", call)
	}
	if messages[1].Role != "tool" || messages[1].ToolCallID != "call_lookup" || messages[1].Name != "lookup" {
		t.Fatalf("tool result correlation fields were not preserved: %#v", messages[1])
	}
	if err := format.ValidateToolResultReferences(messages); err != nil {
		t.Fatalf("preserved continuation failed correlation validation: %v", err)
	}
	prompt, _, err := format.MessagesToPromptAndImages(models.OpenAIChatRequest{Messages: messages})
	if err != nil {
		t.Fatalf("preserved continuation failed prompt translation: %v", err)
	}
	if !strings.Contains(prompt, "[Tool result for lookup]") || !strings.Contains(prompt, "[Assistant]") {
		t.Fatalf("prompt lost continuation context: %q", prompt)
	}
}

func TestResponsesRejectsUnsupportedInputItemBeforeUpstream(t *testing.T) {
	app := New(config.Default(), "test-version")
	req := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(`{
		"model":"gemini-3.7-flash",
		"input":[{"type":"reasoning","summary":[]}]
	}`))
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "unsupported input item type") {
		t.Fatalf("client-visible validation error missing: %s", rec.Body.String())
	}
}
