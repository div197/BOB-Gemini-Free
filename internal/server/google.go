package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/div197/bob-gemini-free/internal/format"
	"github.com/div197/bob-gemini-free/internal/gemini"
	"github.com/div197/bob-gemini-free/internal/models"
)

func (a *App) handleGoogleGenerate(w http.ResponseWriter, r *http.Request) {
	target := r.PathValue("target")
	if target == "" {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
		return
	}

	modelName := target
	stream := false
	action := "generateContent"

	isCountTokens := false
	if idx := strings.LastIndex(target, ":"); idx != -1 {
		action = target[idx+1:]
		modelName = target[:idx]
		if action == "streamGenerateContent" {
			stream = true
		} else if action == "countTokens" {
			isCountTokens = true
		} else if action != "generateContent" {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
			return
		}
	}

	bodyBytes, err := readRequestBody(r)
	if err != nil || len(bodyBytes) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "invalid JSON"}})
		return
	}
	if !json.Valid(bodyBytes) {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "invalid JSON"}})
		return
	}

	providerKey, useDeveloperAPI, keyErr := a.geminiAPIKeyForRequest(r)
	if keyErr != nil {
		writeGeminiAPIError(w, keyErr)
		return
	}
	if useDeveloperAPI {
		a.observeRoute(routeGeminiDeveloperAPI)
		a.handleDirectGoogleGenerate(w, r, modelName, action, bodyBytes, providerKey)
		return
	}
	a.observeRoute(routeGoogleWebRPC)

	var req models.GoogleGenerateRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "invalid JSON"}})
		return
	}

	if isCountTokens {
		totalTokens := format.CountGoogleTokens(req)
		a.RequestsServed.Add(1)
		a.addEstimatedTokens(uint64(totalTokens))
		writeJSON(w, http.StatusOK, map[string]any{
			"totalTokens": totalTokens,
		})
		return
	}

	if modelName == "" {
		modelName = a.Cfg.DefaultModel
	}

	resolved, err := models.Resolve(modelName, a.Cfg.DefaultModel)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": err.Error()}})
		return
	}

	fcMode, err := format.GoogleFunctionCallingMode(req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": err.Error(), "type": "invalid_request_error"}})
		return
	}

	hasTools := len(req.Tools) > 0 && fcMode != "NONE"

	prompt, images, err := format.GoogleContentsToPrompt(req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": err.Error(), "type": "invalid_request_error"}})
		return
	}
	if strings.TrimSpace(prompt) == "" && len(images) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "empty content", "type": "invalid_request_error"}})
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
	a.RequestsServed.Add(1) // Track all Google API requests, not just countTokens
	a.logf("Google API: model=%s stream=%t tools=%t prompt_len=%d", resolved.Name, stream, hasTools, len(prompt))

	if stream && !hasTools {
		if !startSSE(w) {
			return
		}

		var fullText string
		emitErr := StreamWithKeepAlive(r.Context(), w, 2500*time.Millisecond, func(emitDelta func(string) error) error {
			return a.Gem.GenerateStreamContext(r.Context(), prompt, resolved.Mode, resolved.Think, fileRefs, resolved.Extra, emitDelta)
		}, func(delta string) error {
			if delta == "" {
				return nil
			}
			fullText += delta
			chunkObj := models.GoogleGenerateResponse{
				Candidates: []models.GoogleCandidate{
					{
						Index: 0,
						Content: &models.GoogleContent{
							Role: "model",
							Parts: []models.GooglePart{
								{Text: delta},
							},
						},
					},
				},
				ModelVersion: resolved.Name,
			}
			return writeSSEData(w, chunkObj)
		})

		if emitErr == nil && strings.TrimSpace(fullText) == "" {
			emitErr = &gemini.UpstreamError{Kind: "protocol", Msg: "upstream response contained no usable text"}
		}
		if emitErr == nil {
			promptTokens := format.EstimateTokens(prompt)
			candidatesTokens := format.EstimateTokens(fullText)
			a.addEstimatedTokens(uint64(promptTokens + candidatesTokens))
			// Final chunk: must include Content with empty parts so Google client SDKs don't nil-pointer
			finalChunk := models.GoogleGenerateResponse{
				Candidates: []models.GoogleCandidate{
					{
						Index:        0,
						FinishReason: "STOP",
						Content: &models.GoogleContent{
							Role:  "model",
							Parts: []models.GooglePart{{Text: ""}},
						},
					},
				},
				UsageMetadata: &models.GoogleUsageMetadata{
					PromptTokenCount:     promptTokens,
					CandidatesTokenCount: candidatesTokens,
					TotalTokenCount:      promptTokens + candidatesTokens,
				},
				ModelVersion: resolved.Name,
			}
			_ = writeSSEData(w, finalChunk)
		} else {
			a.logf("Google stream error: %s", publicUpstreamErrorMessage(emitErr))
			// Headers have already been sent. A top-level error preserves the
			// native Google-shaped stream contract without pretending that a
			// provider failure is model-authored Markdown. Native Google SSE
			// streams terminate with HTTP EOF rather than OpenAI's [DONE].
			_ = writeSSEError(w, emitErr)
		}
		return
	}

	text, err := a.Gem.GenerateContext(r.Context(), prompt, resolved.Mode, resolved.Think, fileRefs, resolved.Extra)
	if err != nil {
		writeJSON(w, ErrorToStatusCode(err), map[string]any{"error": map[string]any{"message": publicUpstreamErrorMessage(err), "type": "api_error"}})
		return
	}
	if strings.TrimSpace(text) == "" {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": map[string]any{
			"message": "upstream response contained no usable text",
			"type":    "api_error",
		}})
		return
	}

	var responseParts []models.GooglePart
	if hasTools && text != "" {
		cleanText, fnCalls := format.ParseGoogleFunctionCalls(text)
		if len(fnCalls) > 0 {
			if cleanText != "" {
				responseParts = append(responseParts, models.GooglePart{Text: cleanText})
			}
			for _, fc := range fnCalls {
				responseParts = append(responseParts, models.GooglePart{
					FunctionCall: &models.GoogleFunctionCall{
						Name: fc.Name,
						Args: fc.Args,
					},
				})
			}
		} else {
			responseParts = append(responseParts, models.GooglePart{Text: text})
		}
	} else {
		responseParts = append(responseParts, models.GooglePart{Text: text})
	}

	promptTokens := format.EstimateTokens(prompt)
	candidatesTokens := format.EstimateTokens(text)
	a.addEstimatedTokens(uint64(promptTokens + candidatesTokens))

	responseObj := models.GoogleGenerateResponse{
		Candidates: []models.GoogleCandidate{
			{
				Index: 0,
				Content: &models.GoogleContent{
					Role:  "model",
					Parts: responseParts,
				},
				FinishReason: "STOP",
			},
		},
		UsageMetadata: &models.GoogleUsageMetadata{
			PromptTokenCount:     promptTokens,
			CandidatesTokenCount: candidatesTokens,
			TotalTokenCount:      promptTokens + candidatesTokens,
		},
		ModelVersion: resolved.Name,
	}

	if stream {
		if !startSSE(w) {
			return
		}
		_ = writeSSEData(w, responseObj)
	} else {
		writeJSON(w, http.StatusOK, responseObj)
	}
}
