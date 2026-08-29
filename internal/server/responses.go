package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/div197/bob-gemini-free/internal/format"
	"github.com/div197/bob-gemini-free/internal/models"
)

func (a *App) handleResponses(w http.ResponseWriter, r *http.Request) {
	if a.rejectDeveloperAPIOnRoute(w, r, "/v1/responses") {
		return
	}
	bodyBytes, err := readRequestBody(r)
	if err != nil || len(bodyBytes) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "invalid JSON"}})
		return
	}

	var req map[string]any
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "invalid JSON"}})
		return
	}

	modelStr, _ := req["model"].(string)
	if modelStr == "" {
		modelStr = a.Cfg.DefaultModel
	}

	resolved, err := models.Resolve(modelStr, a.Cfg.DefaultModel)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": err.Error()}})
		return
	}

	if reStr, ok := req["reasoning_effort"].(string); ok && reStr != "" {
		switch strings.ToLower(reStr) {
		case "high", "xhigh":
			resolved.Think = 0
		case "medium":
			resolved.Think = 2
		case "low", "minimal", "none":
			resolved.Think = 4
		}
	}

	inputRaw := req["input"]
	instructions, _ := req["instructions"].(string)

	messages, err := format.ResponsesInputToMessages(inputRaw, instructions)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "invalid input structure"}})
		return
	}

	var reqMessages []models.OpenAIMessage
	for _, m := range messages {
		role, _ := m["role"].(string)
		content := m["content"]
		reqMessages = append(reqMessages, models.OpenAIMessage{Role: role, Content: content})
	}

	var reqTools []models.OpenAITool
	if toolsRaw, ok := req["tools"].([]any); ok {
		for _, t := range toolsRaw {
			if tm, ok := t.(map[string]any); ok {
				var fn models.OpenAIFunction
				if f, ok := tm["function"].(map[string]any); ok {
					fn.Name, _ = f["name"].(string)
					fn.Description, _ = f["description"].(string)
					fn.Parameters = f["parameters"]
				} else {
					fn.Name, _ = tm["name"].(string)
					fn.Description, _ = tm["description"].(string)
					fn.Parameters = tm["parameters"]
				}
				reqTools = append(reqTools, models.OpenAITool{Function: fn})
			}
		}
	}

	toolChoice := req["tool_choice"]
	if toolChoice == nil {
		toolChoice = "auto"
	}

	chatReq := models.OpenAIChatRequest{
		Messages:   reqMessages,
		Tools:      reqTools,
		ToolChoice: toolChoice,
	}

	prompt, images, err := format.MessagesToPromptAndImages(chatReq)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": err.Error(), "type": "invalid_request_error"}})
		return
	}
	if strings.TrimSpace(prompt) == "" && len(images) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "empty input", "type": "invalid_request_error"}})
		return
	}
	if strings.TrimSpace(prompt) == "" && len(images) > 0 {
		prompt = "Please analyze the attached image."
	}

	fileRefs, err := a.uploadImagesContext(r.Context(), images)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": map[string]any{"message": err.Error(), "type": "api_error"}})
		return
	}

	strChoice, isStr := toolChoice.(string)
	isToolNone := isStr && strChoice == "none"

	rid := fmt.Sprintf("resp_%s", format.RandHex(16))
	mid := fmt.Sprintf("msg_%s", format.RandHex(12))
	stream, _ := req["stream"].(bool)
	promptTokens := format.EstimateTokens(prompt)

	a.RequestsServed.Add(1)

	// --- Real-time streaming path (no tools) ---
	if stream && (len(reqTools) == 0 || isToolNone) {
		if !startSSE(w) {
			return
		}

		// 1. response.created
		_ = writeSSEEvent(w, "response.created", map[string]any{
			"type": "response.created",
			"response": map[string]any{
				"id":     rid,
				"object": "response",
				"status": "in_progress",
				"model":  resolved.Name,
				"output": []any{},
			},
		})

		// 2. output_item.added (message item)
		itemID := fmt.Sprintf("item_%s", format.RandHex(8))
		_ = writeSSEEvent(w, "response.output_item.added", map[string]any{
			"type":         "response.output_item.added",
			"output_index": 0,
			"item": map[string]any{
				"id":      itemID,
				"type":    "message",
				"role":    "assistant",
				"status":  "in_progress",
				"content": []any{},
			},
		})

		// 3. content_part.added
		_ = writeSSEEvent(w, "response.content_part.added", map[string]any{
			"type":          "response.content_part.added",
			"item_id":       itemID,
			"output_index":  0,
			"content_index": 0,
			"part": map[string]any{
				"type": "output_text",
				"text": "",
			},
		})

		// 4. Real-time delta stream
		var fullStreamText string
		streamErr := StreamWithKeepAlive(r.Context(), w, 2500*time.Millisecond, func(emitDelta func(string) error) error {
			return a.Gem.GenerateStreamContext(r.Context(), prompt, resolved.Mode, resolved.Think, fileRefs, resolved.Extra, emitDelta)
		}, func(delta string) error {
			fullStreamText += delta
			return writeSSEEvent(w, "response.output_text.delta", map[string]any{
				"type":          "response.output_text.delta",
				"item_id":       itemID,
				"output_index":  0,
				"content_index": 0,
				"delta":         delta,
			})
		})
		if streamErr != nil {
			_ = writeSSEEvent(w, "error", map[string]any{
				"type":    "error",
				"message": streamErr.Error(),
			})
			return
		}

		outputTokens := format.EstimateTokens(fullStreamText)
		if outputTokens == 0 {
			outputTokens = 1
		}
		a.addEstimatedTokens(uint64(promptTokens + outputTokens))

		// 5. content_part.done
		_ = writeSSEEvent(w, "response.content_part.done", map[string]any{
			"type":          "response.content_part.done",
			"item_id":       itemID,
			"output_index":  0,
			"content_index": 0,
			"part": map[string]any{
				"type": "output_text",
				"text": fullStreamText,
			},
		})

		// 6. output_text.done
		_ = writeSSEEvent(w, "response.output_text.done", map[string]any{
			"type":          "response.output_text.done",
			"item_id":       itemID,
			"output_index":  0,
			"content_index": 0,
			"text":          fullStreamText,
		})

		// 7. output_item.done
		_ = writeSSEEvent(w, "response.output_item.done", map[string]any{
			"type":         "response.output_item.done",
			"output_index": 0,
			"item": map[string]any{
				"id":     itemID,
				"type":   "message",
				"role":   "assistant",
				"status": "completed",
				"content": []map[string]any{
					{"type": "output_text", "text": fullStreamText},
				},
			},
		})

		// 8. response.completed
		outputItems := format.BuildResponseOutput(fullStreamText, nil, mid)
		_ = writeSSEEvent(w, "response.completed", map[string]any{
			"type": "response.completed",
			"response": map[string]any{
				"id":          rid,
				"object":      "response",
				"status":      "completed",
				"model":       resolved.Name,
				"output":      outputItems,
				"output_text": fullStreamText,
				"usage": map[string]any{
					"input_tokens":  promptTokens,
					"output_tokens": outputTokens,
					"total_tokens":  promptTokens + outputTokens,
				},
			},
		})
		return
	}

	// --- Non-streaming path (or streaming with tools: buffer then replay) ---
	text, err := a.Gem.GenerateContext(r.Context(), prompt, resolved.Mode, resolved.Think, fileRefs, resolved.Extra)
	if err != nil {
		writeJSON(w, ErrorToStatusCode(err), map[string]any{"error": map[string]any{"message": fmt.Sprintf("upstream error: %v", err)}})
		return
	}

	// Strip thinking blocks from the raw response before exposing to user.
	// This mirrors chat.go and anthropic.go which both call ExtractThinking.
	if thinking, cleanText := format.ExtractThinking(text); thinking != "" {
		text = cleanText
	}

	var toolCalls []models.OpenAIToolCall
	if len(reqTools) > 0 && text != "" && !isToolNone {
		text, toolCalls = format.ParseToolCalls(text)
	}

	outputItems := format.BuildResponseOutput(text, toolCalls, mid)
	outputTokens := format.EstimateTokens(text)
	if outputTokens == 0 {
		outputTokens = 1
	}
	a.addEstimatedTokens(uint64(promptTokens + outputTokens))

	if stream {
		// Tool-call streaming: replay synchronously after buffering
		if !startSSE(w) {
			return
		}
		_ = writeSSEEvent(w, "response.created", map[string]any{
			"type": "response.created",
			"response": map[string]any{
				"id":     rid,
				"object": "response",
				"status": "in_progress",
				"model":  resolved.Name,
				"output": []any{},
			},
		})
		for _, item := range outputItems {
			iType, _ := item["type"].(string)
			if iType == "function_call" {
				_ = writeSSEEvent(w, "response.function_call_arguments.done", map[string]any{
					"type":      "response.function_call_arguments.done",
					"item_id":   item["id"],
					"call_id":   item["call_id"],
					"name":      item["name"],
					"arguments": item["arguments"],
				})
			} else if iType == "message" {
				if content, ok := item["content"].([]map[string]any); ok {
					for ci, cp := range content {
						_ = writeSSEEvent(w, "response.output_text.done", map[string]any{
							"type":          "response.output_text.done",
							"item_id":       item["id"],
							"content_index": ci,
							"text":          cp["text"],
						})
					}
				}
			}
		}
		_ = writeSSEEvent(w, "response.completed", map[string]any{
			"type": "response.completed",
			"response": map[string]any{
				"id":          rid,
				"object":      "response",
				"status":      "completed",
				"model":       resolved.Name,
				"output":      outputItems,
				"output_text": text,
				"usage": map[string]any{
					"input_tokens":  promptTokens,
					"output_tokens": outputTokens,
					"total_tokens":  promptTokens + outputTokens,
				},
			},
		})
	} else {
		writeJSON(w, http.StatusOK, map[string]any{
			"id":          rid,
			"object":      "response",
			"created_at":  time.Now().Unix(),
			"status":      "completed",
			"model":       resolved.Name,
			"output":      outputItems,
			"output_text": text,
			"usage": map[string]any{
				"input_tokens":  promptTokens,
				"output_tokens": outputTokens,
				"total_tokens":  promptTokens + outputTokens,
			},
		})
	}
}
