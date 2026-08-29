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

func (a *App) handleChat(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := readRequestBody(r)
	if err != nil || len(bodyBytes) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "invalid JSON"}})
		return
	}

	var req models.OpenAIChatRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "invalid JSON"}})
		return
	}

	// Resolve an explicit Developer API key before consulting BOB's web-RPC
	// alias catalog. The public provider may publish a new gemini-* model ID
	// before this local web catalog is updated; that route must be forwarded to
	// Google unchanged rather than rejected by the web adapter.
	providerKey, useDeveloperAPI, keyErr := a.geminiAPIKeyForRequest(r)
	if keyErr != nil {
		writeGeminiAPIError(w, keyErr)
		return
	}
	if useDeveloperAPI {
		a.handleDirectGeminiChat(w, r, req, providerKey)
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
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": err.Error(), "type": "invalid_request_error"}})
		return
	}
	if strings.TrimSpace(prompt) == "" && len(images) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "empty prompt", "type": "invalid_request_error"}})
		return
	}
	if strings.TrimSpace(prompt) == "" && len(images) > 0 {
		prompt = "Please analyze the attached image."
	}

	fileRefs, err := a.uploadImagesContext(r.Context(), images)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": map[string]any{"message": publicAttachmentErrorMessage(err), "type": "api_error"}})
		return
	}
	cid := fmt.Sprintf("chatcmpl-%s", format.RandHex(12))
	a.RequestsServed.Add(1)

	isToolNone := format.IsToolChoiceNone(req.ToolChoice)

	if req.Stream && (len(req.Tools) == 0 || isToolNone) {
		if !startSSE(w) {
			return
		}

		splitter := format.NewThinkingStreamSplitter()

		baseMsg := &models.OpenAIMessage{}
		baseChunk := models.OpenAIChatResponse{
			ID:                cid,
			Object:            "chat.completion.chunk",
			Created:           time.Now().Unix(),
			Model:             resolved.Name,
			SystemFingerprint: "fp_bob_gemini",
			Choices: []models.OpenAIChoice{
				{
					Index:        0,
					Delta:        baseMsg,
					FinishReason: nil,
				},
			},
		}

		emitChunk := func(ch format.StreamChunk) error {
			if ch.Text == "" {
				return nil
			}

			// Reset the reused message
			baseMsg.Content = nil
			baseMsg.ReasoningContent = ""

			if ch.Type == format.DeltaThinking {
				baseMsg.ReasoningContent = ch.Text
			} else if ch.Type == format.DeltaContent {
				baseMsg.Content = ch.Text
			} else {
				return nil
			}
			return writeSSEData(w, baseChunk)
		}

		emitErr := StreamWithKeepAlive(r.Context(), w, 2500*time.Millisecond, func(emitDelta func(string) error) error {
			return a.Gem.GenerateStreamContext(r.Context(), prompt, resolved.Mode, resolved.Think, fileRefs, resolved.Extra, emitDelta)
		}, func(delta string) error {
			for _, ch := range splitter.Feed(delta) {
				if err := emitChunk(ch); err != nil {
					return err
				}
			}
			return nil
		})

		if emitErr == nil {
			// Flush any remaining buffered tokens
			for _, ch := range splitter.Flush() {
				_ = emitChunk(ch)
			}

			fullThinking := splitter.GetFullThinking()
			fullContent := splitter.GetFullContent()

			pTokens := format.EstimateTokens(prompt)
			cTokens := format.EstimateTokens(fullContent)
			var rTokens int
			if fullThinking != "" {
				rTokens = format.EstimateTokens(fullThinking)
				cTokens += rTokens
			}
			if cTokens == 0 {
				cTokens = 1
			}
			a.addEstimatedTokens(uint64(pTokens + cTokens))

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
			a.Logf("Chat stream error: %s", publicUpstreamErrorMessage(emitErr))
			// Headers have already been sent, so preserve any partial model
			// output but never turn a transport/provider failure into
			// assistant-authored Markdown. The browser and standard SSE
			// consumers can classify this structured error without storing it as
			// successful model output.
			_ = writeSSEError(w, emitErr)
			_ = writeSSEDone(w)
		}
		return
	}

	text, err := a.Gem.GenerateContext(r.Context(), prompt, resolved.Mode, resolved.Think, fileRefs, resolved.Extra)
	if err != nil {
		writeJSON(w, ErrorToStatusCode(err), map[string]any{"error": map[string]any{"message": publicUpstreamErrorMessage(err), "type": "api_error"}})
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

	a.addEstimatedTokens(uint64(usage.TotalTokens))

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
