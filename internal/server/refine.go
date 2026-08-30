package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/models"
	"github.com/div197/bob-gemini-free/internal/refiner"
)

type RefineRequest struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model,omitempty"`
}

func (a *App) handleRefine(w http.ResponseWriter, r *http.Request) {
	if a == nil || a.Gem == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"error": map[string]any{
				"message": "gateway engine is not initialized",
				"type":    "api_error",
			},
		})
		return
	}
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
	a.observeRoute(routeRefineWebRPC)

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
	timeoutSec := a.Cfg.RequestTimeoutSec
	if timeoutSec <= 0 {
		timeoutSec = config.DefaultRequestTimeoutSec
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	result, err := engine.Refine(ctx, req.Prompt, func(ctx context.Context, p string) (string, error) {
		return a.Gem.GenerateContext(ctx, p, resolved.Mode, resolved.Think, nil, resolved.Extra)
	})

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": publicUpstreamErrorMessage(err),
				"type":    "api_error",
			},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}
