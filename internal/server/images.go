package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/div197/bob-gemini-free/internal/format"
	"github.com/div197/bob-gemini-free/internal/models"
	"github.com/div197/bob-gemini-free/internal/multimodal"
)

func (a *App) handleImageGenerations(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil || len(bodyBytes) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]any{
				"message": "invalid JSON",
				"type":    "invalid_request_error",
				"code":    "bad_request",
			},
		})
		return
	}

	var req models.OpenAIImageGenerationRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil || strings.TrimSpace(req.Prompt) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]any{
				"message": "empty prompt or invalid JSON",
				"type":    "invalid_request_error",
				"code":    "bad_request",
			},
		})
		return
	}

	imagePrompt := fmt.Sprintf("Generate an image of: %s. Return the generated image directly.", req.Prompt)
	resolved, _ := models.Resolve("gemini-3.7-flash", a.Cfg.DefaultModel)

	text, err := a.Gem.GenerateContext(r.Context(), imagePrompt, resolved.Mode, resolved.Think, nil, resolved.Extra)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"error": map[string]any{
				"message": fmt.Sprintf("upstream error: %v", err),
				"type":    "api_error",
			},
		})
		return
	}

	extracted := format.ExtractImageURLsFromText(text)
	var imageObjects []models.OpenAIImageObject

	for _, img := range extracted {
		obj := models.OpenAIImageObject{
			RevisedPrompt: img.RevisedPrompt,
		}
		if obj.RevisedPrompt == "" {
			obj.RevisedPrompt = req.Prompt
		}

		if strings.ToLower(req.ResponseFormat) == "b64_json" {
			imgBytes, err := multimodal.FetchImageBytes(a.HTTPClient, img.URL)
			if err == nil && len(imgBytes) > 0 {
				obj.B64JSON = base64.StdEncoding.EncodeToString(imgBytes)
			} else {
				obj.URL = img.URL
			}
		} else {
			obj.URL = img.URL
		}
		imageObjects = append(imageObjects, obj)
	}

	if len(imageObjects) == 0 {
		imageObjects = append(imageObjects, models.OpenAIImageObject{
			URL:           "https://generativelanguage.googleapis.com/v1beta/images/placeholder.png",
			RevisedPrompt: text,
		})
	}

	resp := models.OpenAIImageGenerationResponse{
		Created: time.Now().Unix(),
		Data:    imageObjects,
	}

	writeJSON(w, http.StatusOK, resp)
}
