package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/div197/bob-gemini-free/internal/format"
	"github.com/div197/bob-gemini-free/internal/models"
)

func (a *App) handleAnthropicMessages(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil || len(bodyBytes) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"type": "error",
			"error": map[string]any{
				"type":    "invalid_request_error",
				"message": "invalid JSON body",
			},
		})
		return
	}

	var req models.AnthropicMessagesRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"type": "error",
			"error": map[string]any{
				"type":    "invalid_request_error",
				"message": fmt.Sprintf("invalid JSON payload: %v", err),
			},
		})
		return
	}

	if req.Model == "" {
		req.Model = a.Cfg.DefaultModel
	}

	resolved, err := models.Resolve(req.Model, a.Cfg.DefaultModel)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"type": "error",
			"error": map[string]any{
				"type":    "not_found_error",
				"message": err.Error(),
			},
		})
		return
	}

	chatReq := format.AnthropicToOpenAIChatRequest(req)
	prompt, images, err := format.MessagesToPromptAndImages(chatReq)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"type": "error",
			"error": map[string]any{
				"type":    "invalid_request_error",
				"message": err.Error(),
			},
		})
		return
	}

	if strings.TrimSpace(prompt) == "" && len(images) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"type": "error",
			"error": map[string]any{
				"type":    "invalid_request_error",
				"message": "empty messages prompt",
			},
		})
		return
	}

	// Upload any multimodal image attachments if present
	var fileRefs []string
	if len(images) > 0 {
		fileRefs = a.uploadImages(images)
	}

	msgID := fmt.Sprintf("msg_%s", format.RandHex(24))
	promptTokens := len(prompt) / 4

	if req.Stream {
		if !startSSE(w) {
			return
		}

		// 1. message_start event
		evStart := map[string]any{
			"type": "message_start",
			"message": map[string]any{
				"id":            msgID,
				"type":          "message",
				"role":          "assistant",
				"model":         resolved.Name,
				"content":       []any{},
				"stop_reason":   nil,
				"stop_sequence": nil,
				"usage": map[string]any{
					"input_tokens":  promptTokens,
					"output_tokens": 1,
				},
			},
		}
		_ = writeSSEEvent(w, "message_start", evStart)

		// 2. content_block_start event
		evBlockStart := map[string]any{
			"type":          "content_block_start",
			"index":         0,
			"content_block": map[string]any{"type": "text", "text": ""},
		}
		_ = writeSSEEvent(w, "content_block_start", evBlockStart)

		var fullText string
		streamErr := a.Gem.GenerateStreamContext(r.Context(), prompt, resolved.Mode, resolved.Think, fileRefs, resolved.Extra, func(delta string) error {
			fullText += delta
			evDelta := map[string]any{
				"type":  "content_block_delta",
				"index": 0,
				"delta": map[string]any{
					"type": "text_delta",
					"text": delta,
				},
			}
			return writeSSEEvent(w, "content_block_delta", evDelta)
		})

		if streamErr != nil {
			evErr := map[string]any{
				"type": "error",
				"error": map[string]any{
					"type":    "api_error",
					"message": streamErr.Error(),
				},
			}
			_ = writeSSEEvent(w, "error", evErr)
			return
		}

		// 3. content_block_stop event
		_ = writeSSEEvent(w, "content_block_stop", map[string]any{
			"type":  "content_block_stop",
			"index": 0,
		})

		// 4. message_delta event
		outTokens := len(fullText) / 4
		if outTokens == 0 {
			outTokens = 1
		}
		evMsgDelta := map[string]any{
			"type": "message_delta",
			"delta": map[string]any{
				"stop_reason":   "end_turn",
				"stop_sequence": nil,
			},
			"usage": map[string]any{
				"output_tokens": outTokens,
			},
		}
		_ = writeSSEEvent(w, "message_delta", evMsgDelta)

		// 5. message_stop event
		_ = writeSSEEvent(w, "message_stop", map[string]any{
			"type": "message_stop",
		})
		return
	}

	// Non-streaming response
	text, err := a.Gem.GenerateContext(r.Context(), prompt, resolved.Mode, resolved.Think, fileRefs, resolved.Extra)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"type": "error",
			"error": map[string]any{
				"type":    "api_error",
				"message": fmt.Sprintf("upstream error: %v", err),
			},
		})
		return
	}

	var toolCalls []models.OpenAIToolCall
	if len(chatReq.Tools) > 0 && text != "" {
		text, toolCalls = format.ParseToolCalls(text)
	}

	thinking, cleanText := format.ExtractThinking(text)
	if thinking != "" {
		text = cleanText
	}

	stopReason := "end_turn"
	if len(toolCalls) > 0 {
		stopReason = "tool_use"
	}

	contentBlocks := format.ConvertToolCallsAndThinkingToAnthropicBlocks(thinking, text, toolCalls)
	outputTokens := len(text) / 4
	if outputTokens == 0 {
		outputTokens = 1
	}

	resp := models.AnthropicMessagesResponse{
		ID:           msgID,
		Type:         "message",
		Role:         "assistant",
		Model:        resolved.Name,
		Content:      contentBlocks,
		StopReason:   stopReason,
		StopSequence: nil,
		Usage: models.AnthropicUsage{
			InputTokens:  promptTokens,
			OutputTokens: outputTokens,
		},
	}

	writeJSON(w, http.StatusOK, resp)
}
