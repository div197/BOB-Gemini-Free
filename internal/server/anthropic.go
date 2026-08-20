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

	if req.Thinking != nil {
		if req.Thinking.Type == "enabled" {
			if req.Thinking.BudgetTokens > 4000 {
				resolved.Think = 0
			} else if req.Thinking.BudgetTokens > 0 {
				resolved.Think = 2
			} else {
				resolved.Think = 0
			}
		} else if req.Thinking.Type == "disabled" {
			resolved.Think = 4
		}
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
	promptTokens := format.EstimateTokens(prompt)
	a.RequestsServed.Add(1)

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
					"input_tokens":                promptTokens,
					"output_tokens":               1,
					"cache_creation_input_tokens": 0,
					"cache_read_input_tokens":     0,
				},
			},
		}
		_ = writeSSEEvent(w, "message_start", evStart)

		splitter := format.NewThinkingStreamSplitter()
		var startedThinkingBlock bool
		var startedTextBlock bool
		var currentBlockIndex int

		startThinkingBlock := func() error {
			if !startedThinkingBlock {
				startedThinkingBlock = true
				return writeSSEEvent(w, "content_block_start", map[string]any{
					"type":          "content_block_start",
					"index":         currentBlockIndex,
					"content_block": map[string]any{"type": "thinking", "thinking": ""},
				})
			}
			return nil
		}

		startTextBlock := func() error {
			if !startedTextBlock {
				startedTextBlock = true
				return writeSSEEvent(w, "content_block_start", map[string]any{
					"type":          "content_block_start",
					"index":         currentBlockIndex,
					"content_block": map[string]any{"type": "text", "text": ""},
				})
			}
			return nil
		}

		stopCurrentBlock := func() error {
			return writeSSEEvent(w, "content_block_stop", map[string]any{
				"type":  "content_block_stop",
				"index": currentBlockIndex,
			})
		}

		emitChunk := func(ch format.StreamChunk) error {
			if ch.TransitionToContent {
				if startedThinkingBlock && !startedTextBlock {
					if err := stopCurrentBlock(); err != nil {
						return err
					}
					currentBlockIndex++
				}
			}

			if ch.Type == format.DeltaThinking {
				if err := startThinkingBlock(); err != nil {
					return err
				}
				if ch.Text != "" {
					return writeSSEEvent(w, "content_block_delta", map[string]any{
						"type":  "content_block_delta",
						"index": currentBlockIndex,
						"delta": map[string]any{
							"type":     "thinking_delta",
							"thinking": ch.Text,
						},
					})
				}
			} else if ch.Type == format.DeltaContent {
				if err := startTextBlock(); err != nil {
					return err
				}
				if ch.Text != "" {
					return writeSSEEvent(w, "content_block_delta", map[string]any{
						"type":  "content_block_delta",
						"index": currentBlockIndex,
						"delta": map[string]any{
							"type": "text_delta",
							"text": ch.Text,
						},
					})
				}
			}
			return nil
		}

		streamErr := a.Gem.GenerateStreamContext(r.Context(), prompt, resolved.Mode, resolved.Think, fileRefs, resolved.Extra, func(delta string) error {
			for _, ch := range splitter.Feed(delta) {
				if err := emitChunk(ch); err != nil {
					return err
				}
			}
			return nil
		})

		if streamErr != nil {
			_ = writeSSEEvent(w, "error", map[string]any{
				"type": "error",
				"error": map[string]any{
					"type":    "api_error",
					"message": streamErr.Error(),
				},
			})
			return
		}

		// Flush remaining tokens
		for _, ch := range splitter.Flush() {
			_ = emitChunk(ch)
		}

		// Stop active block
		if startedTextBlock || startedThinkingBlock {
			_ = stopCurrentBlock()
		} else {
			_ = startTextBlock()
			_ = stopCurrentBlock()
		}

		fullThinking := splitter.GetFullThinking()
		fullContent := splitter.GetFullContent()

		outTokens := format.EstimateTokens(fullContent)
		if fullThinking != "" {
			outTokens += format.EstimateTokens(fullThinking)
		}
		if outTokens == 0 {
			outTokens = 1
		}

		_ = writeSSEEvent(w, "message_delta", map[string]any{
			"type": "message_delta",
			"delta": map[string]any{
				"stop_reason":   "end_turn",
				"stop_sequence": nil,
			},
			"usage": map[string]any{
				"output_tokens": outTokens,
			},
		})

		a.TokensProcessed.Add(uint64(promptTokens + outTokens))

		_ = writeSSEEvent(w, "message_stop", map[string]any{
			"type": "message_stop",
		})
		return
	}

	// Non-streaming response
	text, err := a.Gem.GenerateContext(r.Context(), prompt, resolved.Mode, resolved.Think, fileRefs, resolved.Extra)
	if err != nil {
		writeJSON(w, ErrorToStatusCode(err), map[string]any{
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
	outputTokens := format.EstimateTokens(text)
	if thinking != "" {
		outputTokens += format.EstimateTokens(thinking)
	}
	if outputTokens == 0 {
		outputTokens = 1
	}

	a.TokensProcessed.Add(uint64(promptTokens + outputTokens))

	resp := models.AnthropicMessagesResponse{
		ID:           msgID,
		Type:         "message",
		Role:         "assistant",
		Model:        resolved.Name,
		Content:      contentBlocks,
		StopReason:   stopReason,
		StopSequence: nil,
		Usage: models.AnthropicUsage{
			InputTokens:              promptTokens,
			OutputTokens:             outputTokens,
			CacheCreationInputTokens: 0,
			CacheReadInputTokens:     0,
		},
	}

	writeJSON(w, http.StatusOK, resp)
}
