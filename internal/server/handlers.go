package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/div197/bob-gemini-free/internal/format"
	"github.com/div197/bob-gemini-free/internal/metrics"
	"github.com/div197/bob-gemini-free/internal/models"
	"github.com/div197/bob-gemini-free/internal/updater"
)

// HealthzProtocolVersion identifies the local gateway handshake understood by
// desktop wrappers. It is not an authentication credential.
const HealthzProtocolVersion = "1"

// HealthzVersionHeader carries the running gateway's release identity for
// desktop coexistence checks. It contains no credentials or user data.
const HealthzVersionHeader = "X-BOB-Version"

const (
	routeOpenAIChatWebRPC      = metrics.RouteOpenAIChatWebRPC
	routeOpenAIResponsesWebRPC = metrics.RouteOpenAIResponsesWebRPC
	routeAnthropicWebRPC       = metrics.RouteAnthropicWebRPC
	routeGoogleWebRPC          = metrics.RouteGoogleWebRPC
	routeGeminiDeveloperAPI    = metrics.RouteGeminiDeveloperAPI
	routeImageGenerationWebRPC = metrics.RouteImageGenerationWebRPC
	routeRefineWebRPC          = metrics.RouteRefineWebRPC
)

func (a *App) observeRoute(route metrics.Route) {
	if a != nil && a.Metrics != nil {
		a.Metrics.ObserveRoute(route)
	}
}

// handleHealthz is intentionally smaller than the human-facing / telemetry
// route. It performs no upstream, cookie, or GitHub work and is safe for
// container orchestration probes even when API keys protect the main API.
func (a *App) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	version := strings.TrimSpace(a.Version)
	if version == "" {
		version = "dev"
	}
	w.Header().Set("X-BOB-Gateway", "bob-gemini-free")
	w.Header().Set("X-BOB-Protocol", HealthzProtocolVersion)
	w.Header().Set(HealthzVersionHeader, version)
	w.Header().Set("X-BOB-Auth-Required", strconv.FormatBool(len(a.Cfg.APIKeys) > 0))
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	if a.Metrics != nil && a.Gem != nil && a.Gem.Pool != nil {
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

	modelCatalog := models.GetAllModels()
	modelNames := make([]string, 0, len(modelCatalog))
	for k := range modelCatalog {
		modelNames = append(modelNames, k)
	}
	sort.Strings(modelNames)

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
	poolTotal, poolHealthy := 0, 0
	if a.Gem != nil && a.Gem.Pool != nil {
		poolTotal = a.Gem.Pool.Count()
		poolHealthy = a.Gem.Pool.CountHealthy()
	}
	if a.Metrics != nil {
		a.Metrics.SessionPoolTotal.Store(int64(poolTotal))
		a.Metrics.SessionPoolHealthy.Store(int64(poolHealthy))
	}
	uptimeSeconds := 0
	if !a.StartTime.IsZero() {
		uptimeSeconds = int(time.Since(a.StartTime).Seconds())
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
		"uptime_seconds":        uptimeSeconds,
		"pool_sessions_total":   poolTotal,
		"pool_sessions_healthy": poolHealthy,
		"metrics":               a.Metrics.Snapshot(),
	})
}

func (a *App) handleModels(w http.ResponseWriter, r *http.Request) {
	var data []map[string]any
	modelCatalog := models.GetAllModels()
	modelNames := make([]string, 0, len(modelCatalog))
	for name := range modelCatalog {
		modelNames = append(modelNames, name)
	}
	sort.Strings(modelNames)
	for _, name := range modelNames {
		m := modelCatalog[name]
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
	if _, exists := models.GetModel(lookupName); !exists {
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
	m, _ := models.GetModel(resolved.Name)

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
	modelCatalog := models.GetAllModels()
	modelNames := make([]string, 0, len(modelCatalog))
	for name := range modelCatalog {
		modelNames = append(modelNames, name)
	}
	sort.Strings(modelNames)
	for _, name := range modelNames {
		m := modelCatalog[name]
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
	if !updater.IsDesktopVersionCheckable(a.Version) {
		writeJSON(w, http.StatusOK, map[string]any{
			"current_version": a.Version,
			"has_update":      false,
			"error":           publicUpdateCheckErrorMessage(fmt.Errorf("current version %q is not a published release", a.Version)),
		})
		return
	}

	check := a.updateCheck
	if check == nil {
		check = updater.CheckLatestDesktopForChannelContext
	}
	channel := a.updateChannel
	if channel == "" {
		channel = updater.DesktopChannelStable
	}
	res, err := check(r.Context(), a.Version, channel)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"current_version": a.Version,
			"has_update":      false,
			"error":           publicUpdateCheckErrorMessage(err),
		})
		return
	}
	writeJSON(w, http.StatusOK, res)
}
