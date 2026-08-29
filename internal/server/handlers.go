package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/div197/bob-gemini-free/internal/format"
	"github.com/div197/bob-gemini-free/internal/models"
	"github.com/div197/bob-gemini-free/internal/updater"
)

// HealthzProtocolVersion identifies the local gateway handshake understood by
// desktop wrappers. It is not an authentication credential.
const HealthzProtocolVersion = "1"

// handleHealthz is intentionally smaller than the human-facing / telemetry
// route. It performs no upstream, cookie, or GitHub work and is safe for
// container orchestration probes even when API keys protect the main API.
func (a *App) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("X-BOB-Gateway", "bob-gemini-free")
	w.Header().Set("X-BOB-Protocol", HealthzProtocolVersion)
	w.Header().Set("X-BOB-Auth-Required", strconv.FormatBool(len(a.Cfg.APIKeys) > 0))
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	if a.Metrics != nil {
		a.Metrics.SessionPoolTotal.Store(int64(a.Gem.Pool.Count()))
		a.Metrics.SessionPoolHealthy.Store(int64(a.Gem.Pool.CountHealthy()))
	}
	writeJSON(w, http.StatusOK, a.Metrics.Snapshot())
}

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
	if a.Metrics != nil {
		a.Metrics.SessionPoolTotal.Store(int64(poolTotal))
		a.Metrics.SessionPoolHealthy.Store(int64(poolHealthy))
	}

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
		"metrics":               a.Metrics.Snapshot(),
	})
}

func (a *App) handleModels(w http.ResponseWriter, r *http.Request) {
	var data []map[string]any
	for name, m := range models.MODELS {
		pricing := models.GetModelPricing(name)
		data = append(data, map[string]any{
			"id":          name,
			"object":      "model",
			"created":     1700000000,
			"owned_by":    pricing.Provider,
			"description": m.Desc,
			"category":    pricing.Category,
			"pricing":     pricing,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"object": "list",
		"data":   data,
	})
}

func (a *App) handleSingleModel(w http.ResponseWriter, r *http.Request) {
	modelName := r.PathValue("model")
	// Strip prefixes/suffixes to check actual existence before resolving
	lookupName := strings.TrimPrefix(modelName, "models/")
	if idx := strings.LastIndex(lookupName, "@think="); idx != -1 {
		lookupName = lookupName[:idx]
	}
	if _, exists := models.MODELS[lookupName]; !exists {
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

	// Now safe to use Resolve() to handle aliases correctly
	resolved, err := models.Resolve(modelName, a.Cfg.DefaultModel)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": map[string]any{
				"message": err.Error(),
				"type":    "invalid_request_error",
				"param":   "model",
				"code":    "model_not_found",
			},
		})
		return
	}

	// Look up description from the canonical resolved model
	m := models.MODELS[resolved.Name]

	writeJSON(w, http.StatusOK, map[string]any{
		"id":          modelName,
		"object":      "model",
		"created":     1700000000,
		"owned_by":    "google",
		"description": m.Desc,
		"permission": []map[string]any{
			{
				"id":                   "modelperm-" + resolved.Name,
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
	bodyBytes, err := readRequestBody(r)
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
	a.addEstimatedTokens(uint64(totalTokens))

	writeJSON(w, http.StatusOK, map[string]any{
		"prompt_tokens": totalTokens,
		"total_tokens":  totalTokens,
		"model":         req.Model,
	})
}

func (a *App) handleUpdateCheck(w http.ResponseWriter, r *http.Request) {
	res, err := updater.CheckLatest(a.Version)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"current_version": a.Version,
			"has_update":      false,
			"error":           err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, res)
}
