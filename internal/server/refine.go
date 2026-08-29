package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/div197/bob-gemini-free/internal/models"
	"github.com/div197/bob-gemini-free/internal/refiner"
)

type RefineRequest struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model,omitempty"`
}

func (a *App) handleRefine(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := readRequestBody(r)
	var req RefineRequest
	if err != nil || json.Unmarshal(bodyBytes, &req) != nil || req.Prompt == "" {
		http.Error(w, `{"error":{"message":"Invalid JSON or missing 'prompt' field","type":"invalid_request_error"}}`, http.StatusBadRequest)
		return
	}

	modelName := req.Model
	if modelName == "" {
		modelName = a.Cfg.DefaultModel
	}

	resolved, err := models.ResolveStrict(modelName, a.Cfg.DefaultModel)
	if err != nil {
		http.Error(w, `{"error":{"message":"invalid model","type":"invalid_request_error"}}`, http.StatusBadRequest)
		return
	}

	engine := refiner.NewEngine()
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(a.Cfg.RequestTimeoutSec)*time.Second)
	defer cancel()

	result, err := engine.Refine(ctx, req.Prompt, func(ctx context.Context, p string) (string, error) {
		return a.Gem.GenerateContext(ctx, p, resolved.Mode, resolved.Think, nil, resolved.Extra)
	})

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": err.Error(),
				"type":    "api_error",
			},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}
