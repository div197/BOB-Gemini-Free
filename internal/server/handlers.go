package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/div197/bob-gemini-free/internal/format"
	"github.com/div197/bob-gemini-free/internal/models"
)

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
		return
	}

	modelNames := make([]string, 0, len(models.MODELS))
	for k := range models.MODELS {
		modelNames = append(modelNames, k)
	}

	reqs := a.RequestsServed.Load()
	tokens := a.TokensProcessed.Load()
	savingsVal := float64(tokens) / 1_000_000.0 * 3.75
	var savingsUSD string
	if savingsVal == 0 {
		savingsUSD = "$0.00"
	} else if savingsVal < 0.01 {
		savingsUSD = fmt.Sprintf("$%.4f", savingsVal)
	} else {
		savingsUSD = fmt.Sprintf("$%.2f", savingsVal)
	}
	poolTotal := a.Gem.Pool.Count()
	poolHealthy := a.Gem.Pool.CountHealthy()

	writeJSON(w, http.StatusOK, map[string]any{
		"status":                "ok",
		"engine":                "BOB Gemini Free",
		"author":                "Divyanshu Singh Chouhan (@div197)",
		"organization":          "ABCsteps (abcsteps.com)",
		"version":               a.Version,
		"models":                modelNames,
		"requests_served":       reqs,
		"tokens_processed":      tokens,
		"estimated_savings_usd": savingsUSD,
		"uptime_seconds":        int(time.Since(a.StartTime).Seconds()),
		"pool_sessions_total":   poolTotal,
		"pool_sessions_healthy": poolHealthy,
	})
}

func (a *App) handleModels(w http.ResponseWriter, r *http.Request) {
	var data []map[string]any
	for name, m := range models.MODELS {
		data = append(data, map[string]any{
			"id":          name,
			"object":      "model",
			"created":     1700000000,
			"owned_by":    "google",
			"description": m.Desc,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"object": "list",
		"data":   data,
	})
}

func (a *App) handleSingleModel(w http.ResponseWriter, r *http.Request) {
	modelName := r.PathValue("model")
	m, exists := models.MODELS[modelName]
	if !exists {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": map[string]any{
				"message": "The model '" + modelName + "' does not exist",
				"type":    "invalid_request_error",
				"param":   "model",
				"code":    "model_not_found",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":          modelName,
		"object":      "model",
		"created":     1700000000,
		"owned_by":    "google",
		"description": m.Desc,
		"permission": []map[string]any{
			{
				"id":                   "modelperm-" + modelName,
				"object":               "model_permission",
				"created":              1700000000,
				"allow_create_engine":  false,
				"allow_sampling":       true,
				"allow_logprobs":       true,
				"allow_search_indices": false,
				"allow_view":           true,
				"allow_fine_tuning":    false,
				"organization":         "*",
				"group":                nil,
				"is_blocking":          false,
			},
		},
	})
}

func (a *App) handleGoogleModels(w http.ResponseWriter, r *http.Request) {
	var modelList []map[string]any
	for name, m := range models.MODELS {
		modelList = append(modelList, map[string]any{
			"name":                       "models/" + name,
			"displayName":                name,
			"description":                m.Desc,
			"supportedGenerationMethods": []string{"generateContent", "streamGenerateContent"},
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"models": modelList,
	})
}

func (a *App) handleCountTokens(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil || len(bodyBytes) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "invalid JSON body"}})
		return
	}

	var req models.OpenAIChatRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "invalid JSON payload"}})
		return
	}

	totalTokens := format.CountOpenAITokens(req)
	a.RequestsServed.Add(1)
	a.TokensProcessed.Add(uint64(totalTokens))

	writeJSON(w, http.StatusOK, map[string]any{
		"prompt_tokens": totalTokens,
		"total_tokens":  totalTokens,
		"model":         req.Model,
	})
}
