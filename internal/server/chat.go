package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/div197/bob-gemini-free/internal/format"
	"github.com/div197/bob-gemini-free/internal/models"
)

func (a *App) handleChat(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil || len(bodyBytes) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "invalid JSON"}})
		return
	}

	var req models.OpenAIChatRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "invalid JSON"}})
		return
	}

	modelStr := req.Model
	if modelStr == "" {
		modelStr = a.Cfg.DefaultModel
	}

	resolved, err := models.Resolve(modelStr, a.Cfg.DefaultModel)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": err.Error()}})
		return
	}

	if req.ReasoningEffort != "" {
		switch strings.ToLower(req.ReasoningEffort) {
		case "high", "xhigh":
			resolved.Think = 0
		case "medium":
			resolved.Think = 2
		case "low", "minimal", "none":
			resolved.Think = 4
		}
	}

	if req.ToolChoice == nil {
		req.ToolChoice = "auto"
	}

	prompt, images, err := format.MessagesToPromptAndImages(req)
	if err != nil || strings.TrimSpace(prompt) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "empty prompt"}})
		return
	}

	fileRefs := a.uploadImages(images)
	cid := fmt.Sprintf("chatcmpl-%s", format.RandHex(12))
	a.RequestsServed.Add(1)

	strChoice, isStr := req.ToolChoice.(string)
	isToolNone := isStr && strChoice == "none"

	if req.Stream && (len(req.Tools) == 0 || isToolNone) {
		if !startSSE(w) {
			return
		}

		var fullStreamText string
		emitErr := a.Gem.GenerateStreamContext(r.Context(), prompt, resolved.Mode, resolved.Think, fileRefs, resolved.Extra, func(delta string) error {
			fullStreamText += delta
			chunk := models.OpenAIChatResponse{
				ID:                cid,
				Object:            "chat.completion.chunk",
				Created:           time.Now().Unix(),
				Model:             resolved.Name,
				SystemFingerprint: "fp_bob_gemini",
				Choices: []models.OpenAIChoice{
					{
						Index: 0,
						Delta: &models.OpenAIMessage{
							Content: delta,
						},
						FinishReason: nil,
					},
				},
			}
			return writeSSEData(w, chunk)
		})

		if emitErr == nil {
			// Extract thinking from the accumulated stream text (if model emitted a thought block)
			thinking, cleanStreamText := format.ExtractThinking(fullStreamText)

			// If thinking was found, emit a preceding reasoning_content delta chunk
			if thinking != "" {
				thinkChunk := models.OpenAIChatResponse{
					ID:                cid,
					Object:            "chat.completion.chunk",
					Created:           time.Now().Unix(),
					Model:             resolved.Name,
					SystemFingerprint: "fp_bob_gemini",
					Choices: []models.OpenAIChoice{
						{
							Index: 0,
							Delta: &models.OpenAIMessage{
								ReasoningContent: thinking,
							},
							FinishReason: nil,
						},
					},
				}
				_ = writeSSEData(w, thinkChunk)
			}

			pTokens := format.EstimateTokens(prompt)
			cTokens := format.EstimateTokens(cleanStreamText)
			var rTokens int
			if thinking != "" {
				rTokens = format.EstimateTokens(thinking)
				cTokens += rTokens
			}
			if cTokens == 0 {
				cTokens = 1
			}
			a.TokensProcessed.Add(uint64(pTokens + cTokens))

			stopReason := "stop"
			endChunk := models.OpenAIChatResponse{
				ID:                cid,
				Object:            "chat.completion.chunk",
				Created:           time.Now().Unix(),
				Model:             resolved.Name,
				SystemFingerprint: "fp_bob_gemini",
				Choices: []models.OpenAIChoice{
					{
						Index:        0,
						Delta:        &models.OpenAIMessage{},
						FinishReason: &stopReason,
					},
				},
			}
			_ = writeSSEData(w, endChunk)

			if req.StreamOptions != nil && req.StreamOptions.IncludeUsage {
				usage := &models.OpenAIUsage{
					PromptTokens:     pTokens,
					CompletionTokens: cTokens,
					TotalTokens:      pTokens + cTokens,
				}
				if rTokens > 0 {
					usage.CompletionTokensDetails = &models.CompletionTokensDetails{
						ReasoningTokens: rTokens,
					}
				}
				usageChunk := models.OpenAIChatResponse{
					ID:                cid,
					Object:            "chat.completion.chunk",
					Created:           time.Now().Unix(),
					Model:             resolved.Name,
					SystemFingerprint: "fp_bob_gemini",
					Choices:           []models.OpenAIChoice{},
					Usage:             usage,
				}
				_ = writeSSEData(w, usageChunk)
			}

			_ = writeSSEDone(w)
		} else {
			a.Logf("Chat stream error: %v", emitErr)
		}
		return
	}

	text, err := a.Gem.GenerateContext(r.Context(), prompt, resolved.Mode, resolved.Think, fileRefs, resolved.Extra)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": map[string]any{"message": fmt.Sprintf("upstream error: %v", err)}})
		return
	}

	var toolCalls []models.OpenAIToolCall
	if len(req.Tools) > 0 && text != "" && !isToolNone {
		text, toolCalls = format.ParseToolCalls(text)
	}

	thinking, cleanText := format.ExtractThinking(text)
	if thinking != "" {
		text = cleanText
	}

	msg := models.OpenAIMessage{
		Role:             "assistant",
		Content:          text,
		ReasoningContent: thinking,
	}
	if text == "" {
		msg.Content = nil
	}

	if len(toolCalls) > 0 {
		msg.ToolCalls = toolCalls
	}

	finish := "stop"
	if len(toolCalls) > 0 {
		finish = "tool_calls"
	}

	promptTokens := format.EstimateTokens(prompt)
	completionTokens := format.EstimateTokens(text)
	var reasoningTokens int
	if thinking != "" {
		reasoningTokens = format.EstimateTokens(thinking)
		completionTokens += reasoningTokens
	}

	usage := &models.OpenAIUsage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
	}
	if reasoningTokens > 0 {
		usage.CompletionTokensDetails = &models.CompletionTokensDetails{
			ReasoningTokens: reasoningTokens,
		}
	}

	a.TokensProcessed.Add(uint64(usage.TotalTokens))

	if req.Stream {
		if !startSSE(w) {
			return
		}
		chunk := models.OpenAIChatResponse{
			ID:                cid,
			Object:            "chat.completion.chunk",
			Created:           time.Now().Unix(),
			Model:             resolved.Name,
			SystemFingerprint: "fp_bob_gemini",
			Choices: []models.OpenAIChoice{
				{
					Index:        0,
					Delta:        &msg,
					FinishReason: &finish,
				},
			},
		}
		_ = writeSSEData(w, chunk)

		if req.StreamOptions != nil && req.StreamOptions.IncludeUsage {
			usageChunk := models.OpenAIChatResponse{
				ID:                cid,
				Object:            "chat.completion.chunk",
				Created:           time.Now().Unix(),
				Model:             resolved.Name,
				SystemFingerprint: "fp_bob_gemini",
				Choices:           []models.OpenAIChoice{},
				Usage:             usage,
			}
			_ = writeSSEData(w, usageChunk)
		}

		_ = writeSSEDone(w)
	} else {
		resp := models.OpenAIChatResponse{
			ID:                cid,
			Object:            "chat.completion",
			Created:           time.Now().Unix(),
			Model:             resolved.Name,
			SystemFingerprint: "fp_bob_gemini",
			Choices: []models.OpenAIChoice{
				{
					Index:        0,
					Message:      &msg,
					FinishReason: &finish,
				},
			},
			Usage: usage,
		}
		writeJSON(w, http.StatusOK, resp)
	}
}
