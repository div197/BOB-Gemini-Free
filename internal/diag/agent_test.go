package diag

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/models"
	"github.com/div197/bob-gemini-free/internal/server"
)

func TestAgenticMultiTurnWorkflow(t *testing.T) {
	cfg := config.Default()
	app := server.New(cfg, "v0.1.0")
	handler := app.Handler()

	// 1. Initial agent request with tool declaration
	tools := []models.OpenAITool{
		{
			Type: "function",
			Function: models.OpenAIFunction{
				Name:        "get_stock_price",
				Description: "Retrieve current stock price for a symbol",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"symbol": map[string]string{"type": "string"},
					},
					"required": []string{"symbol"},
				},
			},
		},
	}

	chatReqTurn1 := models.OpenAIChatRequest{
		Model: "gemini-3.7-flash",
		Messages: []models.OpenAIMessage{
			{Role: "developer", Content: "You are a financial analysis autonomous agent."},
			{Role: "user", Content: "What is the stock price of AAPL?"},
		},
		Tools:      tools,
		ToolChoice: "auto",
	}

	body1, _ := json.Marshal(chatReqTurn1)
	req1 := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()

	handler.ServeHTTP(rec1, req1)
	// Server returns 200 (or 502 in unit test without upstream live network)
	if rec1.Code != http.StatusOK && rec1.Code != http.StatusBadGateway {
		t.Errorf("Unexpected status on agent turn 1: %d", rec1.Code)
	}

	// 2. Turn 2: Simulating tool result loop
	chatReqTurn2 := models.OpenAIChatRequest{
		Model: "gemini-3.7-flash",
		Messages: []models.OpenAIMessage{
			{Role: "developer", Content: "You are a financial analysis autonomous agent."},
			{Role: "user", Content: "What is the stock price of AAPL?"},
			{
				Role: "assistant",
				ToolCalls: []models.OpenAIToolCall{
					{
						ID:   "call_aapl_123",
						Type: "function",
						Function: models.OpenAIToolCallFunction{
							Name:      "get_stock_price",
							Arguments: `{"symbol": "AAPL"}`,
						},
					},
				},
			},
			{
				Role:       "tool",
				Name:       "get_stock_price",
				ToolCallID: "call_aapl_123",
				Content:    `{"symbol": "AAPL", "price": 245.50, "currency": "USD"}`,
			},
		},
		Tools: tools,
	}

	body2, _ := json.Marshal(chatReqTurn2)
	req2 := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()

	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK && rec2.Code != http.StatusBadGateway {
		t.Errorf("Unexpected status on agent turn 2: %d", rec2.Code)
	}
}

func TestCodexResponsesAPIWorkflow(t *testing.T) {
	cfg := config.Default()
	app := server.New(cfg, "v0.1.0")
	handler := app.Handler()

	codexReq := map[string]any{
		"model":        "gemini-3.7-flash",
		"instructions": "You are Codex CLI backend.",
		"input": []map[string]any{
			{"role": "user", "content": "Generate a Go hello world function"},
		},
		"stream": false,
	}

	bodyBytes, _ := json.Marshal(codexReq)
	req := httptest.NewRequest("POST", "/v1/responses", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK && rec.Code != http.StatusBadGateway {
		t.Errorf("Unexpected status on Codex Responses API: %d", rec.Code)
	}
}

func TestSingleModelRetrieve(t *testing.T) {
	cfg := config.Default()
	app := server.New(cfg, "v0.1.0")
	handler := app.Handler()

	req := httptest.NewRequest("GET", "/v1/models/gemini-3.7-flash-thinking", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 for model retrieval, got %d", rec.Code)
	}

	var res map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&res)
	if res["id"] != "gemini-3.7-flash-thinking" {
		t.Errorf("Expected id gemini-3.7-flash-thinking, got %v", res["id"])
	}
	if !strings.Contains(res["id"].(string), "thinking") {
		t.Errorf("Expected thinking model name")
	}
}

func TestClaudeCodeAnthropicWorkflow(t *testing.T) {
	cfg := config.Default()
	app := server.New(cfg, "v0.1.0")
	handler := app.Handler()

	claudeReq := models.AnthropicMessagesRequest{
		Model:     "gemini-3.7-flash",
		System:    "You are Claude Code assistant.",
		MaxTokens: 4096,
		Messages: []models.AnthropicMessage{
			{Role: "user", Content: "Hello Claude Code! How do I build this codebase?"},
		},
		Tools: []models.AnthropicTool{
			{
				Name:        "bash",
				Description: "Execute bash command",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"command": map[string]string{"type": "string"},
					},
					"required": []string{"command"},
				},
			},
		},
		Stream: false,
	}

	bodyBytes, _ := json.Marshal(claudeReq)
	req := httptest.NewRequest("POST", "/v1/messages", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK && rec.Code != http.StatusBadGateway {
		t.Errorf("Unexpected status on Claude Code Anthropic endpoint: %d", rec.Code)
	}
}
