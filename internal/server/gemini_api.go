package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/div197/bob-gemini-free/internal/format"
	"github.com/div197/bob-gemini-free/internal/geminiapi"
	"github.com/div197/bob-gemini-free/internal/models"
)

const geminiProviderKeyHeader = "X-BOB-Gemini-API-Key"

func (a *App) rejectDeveloperAPIOnRoute(w http.ResponseWriter, r *http.Request, route string) bool {
	_, selected, err := a.geminiAPIKeyForRequest(r)
	if err != nil {
		writeGeminiAPIError(w, err)
		return true
	}
	if selected {
		writeGeminiAPIError(w, fmt.Errorf("an explicit Gemini Developer API key is not supported on %s yet; use /v1/chat/completions or the native /v1beta/models/... route", route))
		return true
	}
	return false
}

// geminiAPIKeyForRequest resolves exactly one provider key. A request header
// takes precedence over the process-level key, but a blank or repeated header
// is an error. There is intentionally no pool/rotation behavior: that would
// obscure quota ownership and could be used to evade provider limits.
func (a *App) geminiAPIKeyForRequest(r *http.Request) (string, bool, error) {
	var values []string
	for name, headerValues := range r.Header {
		if strings.EqualFold(name, geminiProviderKeyHeader) {
			values = append(values, headerValues...)
		}
	}
	if len(values) > 1 {
		return "", false, errors.New("only one Gemini Developer API key may be supplied")
	}
	if len(values) == 1 {
		key := strings.TrimSpace(values[0])
		if err := geminiapi.ValidateAPIKey(key); err != nil {
			return "", false, errors.New("invalid Gemini Developer API key")
		}
		return key, true, nil
	}

	key := strings.TrimSpace(a.Cfg.GeminiAPIKey)
	if key == "" {
		return "", false, nil
	}
	if err := geminiapi.ValidateAPIKey(key); err != nil {
		return "", false, errors.New("invalid configured Gemini Developer API key")
	}
	return key, true, nil
}

// directGeminiModel keeps BOB's web-RPC alias catalog out of the Developer API
// path. Known local convenience aliases are mapped explicitly; other
// provider-shaped Gemini IDs are forwarded unchanged so a newly published
// public model can work before BOB ships a catalog update. Google remains the
// authority for whether that ID exists, is available to the project, or is
// included in the selected tier.
func directGeminiModel(requested, defaultModel string) (string, error) {
	name := strings.TrimSpace(requested)
	if name == "" {
		name = strings.TrimSpace(defaultModel)
	}
	name = strings.TrimPrefix(name, "models/")
	if idx := strings.LastIndex(name, "@think="); idx >= 0 {
		return "", errors.New("thinking suffixes are not supported on the explicit Gemini Developer API route")
	}
	switch name {
	case "gemini-3.7-flash", "gemini-3.6-flash":
		return name, nil
	case "gemini-3.5-flash", "gemini-flash":
		return "gemini-3.6-flash", nil
	}
	if !isProviderGeminiModelID(name) {
		return "", fmt.Errorf("model %q is a BOB web-RPC alias or invalid Developer API model ID; use a provider Gemini model ID", name)
	}
	return name, nil
}

func isProviderGeminiModelID(name string) bool {
	if len(name) < len("gemini-x") || len(name) > 256 || !strings.HasPrefix(name, "gemini-") {
		return false
	}
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			continue
		}
		return false
	}
	return true
}

func (a *App) handleDirectGeminiChat(w http.ResponseWriter, r *http.Request, req models.OpenAIChatRequest, key string) {
	if req.ReasoningEffort != "" {
		writeGeminiAPIError(w, errors.New("reasoning_effort is not translated on the explicit Gemini Developer API route"))
		return
	}
	model, err := directGeminiModel(req.Model, a.Cfg.DefaultModel)
	if err != nil {
		writeGeminiAPIError(w, err)
		return
	}
	translated, err := geminiapi.FromOpenAI(req)
	if err != nil {
		writeGeminiAPIError(w, err)
		return
	}
	if len(translated.Contents) == 0 {
		writeGeminiAPIError(w, errors.New("empty prompt"))
		return
	}
	if a.GeminiAPI == nil {
		writeGeminiAPIError(w, errors.New("Gemini Developer API route is not initialized"))
		return
	}

	a.RequestsServed.Add(1)
	if req.Stream {
		a.handleDirectGeminiStream(w, r, req, translated, model, key)
		return
	}

	response, err := a.GeminiAPI.Generate(r.Context(), model, key, translated)
	if err != nil {
		writeGeminiAPIError(w, err)
		return
	}
	result, err := geminiapi.ToOpenAIResult(response)
	if err != nil {
		writeGeminiAPIError(w, err)
		return
	}
	usage := directUsage(req, result.PromptTokens, result.CompletionTokens, result.TotalTokens, result.ReasoningTokens, result.Text, result.Thinking)
	a.addEstimatedTokens(uint64(usage.TotalTokens))

	message := models.OpenAIMessage{
		Role:             "assistant",
		Content:          result.Text,
		ReasoningContent: result.Thinking,
		ToolCalls:        result.ToolCalls,
	}
	if message.Content == "" {
		message.Content = nil
	}
	finish := result.FinishReason
	responseModel := model
	if response.ModelVersion != "" {
		responseModel = response.ModelVersion
	}
	writeJSON(w, http.StatusOK, models.OpenAIChatResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", format.RandHex(12)),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   responseModel,
		Choices: []models.OpenAIChoice{{Index: 0, Message: &message, FinishReason: &finish}},
		Usage:   &usage,
	})
}

func (a *App) handleDirectGoogleGenerate(w http.ResponseWriter, r *http.Request, requestedModel, action string, body []byte, key string) {
	model, err := directGeminiModel(requestedModel, a.Cfg.DefaultModel)
	if err != nil {
		writeGeminiAPIError(w, err)
		return
	}
	if a.GeminiAPI == nil {
		writeGeminiAPIError(w, errors.New("Gemini Developer API route is not initialized"))
		return
	}
	a.RequestsServed.Add(1)
	rawBody := json.RawMessage(body)
	switch action {
	case "countTokens":
		response, err := a.GeminiAPI.CountTokensRaw(r.Context(), model, key, rawBody)
		if err != nil {
			writeGeminiAPIError(w, err)
			return
		}
		var usage struct {
			TotalTokenCount int `json:"totalTokens"`
		}
		if json.Unmarshal(response, &usage) == nil && usage.TotalTokenCount > 0 {
			a.addEstimatedTokens(uint64(usage.TotalTokenCount))
		}
		writeRawJSON(w, http.StatusOK, response)
	case "generateContent":
		response, err := a.GeminiAPI.GenerateRaw(r.Context(), model, key, rawBody)
		if err != nil {
			writeGeminiAPIError(w, err)
			return
		}
		writeRawJSON(w, http.StatusOK, response)
	case "streamGenerateContent":
		if !startSSE(w) {
			return
		}
		err := streamGeminiRawWithKeepAlive(r.Context(), w, 2500*time.Millisecond, func(emit func(json.RawMessage) error) error {
			return a.GeminiAPI.StreamRaw(r.Context(), model, key, rawBody, emit)
		}, func(event json.RawMessage) error {
			return writeSSEData(w, event)
		})
		if err != nil {
			a.Logf("Gemini Developer API Google stream error: %v", err)
			_ = writeSSEError(w, err)
		}
	default:
		writeGeminiAPIError(w, fmt.Errorf("unsupported Gemini Developer API action %q", action))
	}
}

func directUsage(req models.OpenAIChatRequest, promptTokens, completionTokens, totalTokens, reasoningTokens int, text, thinking string) models.OpenAIUsage {
	if promptTokens <= 0 {
		promptTokens = format.CountOpenAITokens(req)
	}
	if completionTokens <= 0 {
		completionTokens = format.EstimateTokens(text + thinking)
		if completionTokens == 0 {
			completionTokens = 1
		}
	}
	if totalTokens <= 0 {
		totalTokens = promptTokens + completionTokens
	}
	usage := models.OpenAIUsage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
	}
	if reasoningTokens > 0 {
		usage.CompletionTokensDetails = &models.CompletionTokensDetails{ReasoningTokens: reasoningTokens}
	}
	return usage
}

type directStreamEvent struct {
	message      models.OpenAIMessage
	finishReason string
	finished     bool
	usage        *geminiapi.UsageMetadata
}

// directStreamToolAssembler keeps native Gemini function calls bounded and
// deterministic across SSE events. Gemini represents function-call arguments
// as decoded JSON objects, while OpenAI clients expect a single JSON argument
// stream. Repeated events for the same candidate/part are therefore treated as
// cumulative snapshots and the latest valid snapshot is retained. Calls are
// emitted only after the provider stream completes, so a client can never see
// a partial or concatenated invalid JSON document.
type directStreamToolAssembler struct {
	order map[string]int
	calls map[string]models.OpenAIToolCall
}

func newDirectStreamToolAssembler() *directStreamToolAssembler {
	return &directStreamToolAssembler{
		order: make(map[string]int),
		calls: make(map[string]models.OpenAIToolCall),
	}
}

func (a *directStreamToolAssembler) Observe(calls []models.OpenAIToolCall) error {
	if a == nil {
		return errors.New("direct stream tool assembler is unavailable")
	}
	for _, call := range calls {
		if err := format.ValidateToolCall(call.Function.Name, call.Function.Arguments); err != nil {
			return fmt.Errorf("invalid Gemini function-call stream event: %w", err)
		}
		if strings.TrimSpace(call.ID) == "" {
			return errors.New("Gemini function-call stream event has an empty ID")
		}
		if existing, ok := a.calls[call.ID]; ok {
			if existing.Type != call.Type || existing.Function.Name != call.Function.Name {
				return fmt.Errorf("Gemini function-call stream ID %q changed name or type", call.ID)
			}
			// The provider may resend a complete snapshot, or may first expose
			// an empty/partial object and then a fuller object. Keep only the
			// latest valid snapshot; never concatenate JSON documents.
			a.calls[call.ID] = call
			continue
		}
		if len(a.calls) >= format.MaxToolDefinitions {
			return fmt.Errorf("Gemini Developer API returned more than %d tool calls in one stream", format.MaxToolDefinitions)
		}
		a.order[call.ID] = len(a.order)
		a.calls[call.ID] = call
	}
	return nil
}

func (a *directStreamToolAssembler) Len() int {
	if a == nil {
		return 0
	}
	return len(a.calls)
}

func (a *directStreamToolAssembler) Finalize() ([]models.OpenAIToolCall, error) {
	if a == nil {
		return nil, errors.New("direct stream tool assembler is unavailable")
	}
	result := make([]models.OpenAIToolCall, len(a.calls))
	for id, call := range a.calls {
		index := a.order[id]
		if index < 0 || index >= len(result) {
			return nil, errors.New("direct stream tool-call ordering state is corrupt")
		}
		if err := format.ValidateToolCall(call.Function.Name, call.Function.Arguments); err != nil {
			return nil, fmt.Errorf("invalid finalized Gemini function call: %w", err)
		}
		result[index] = call
	}
	return result, nil
}

func openAIStreamDelta(response geminiapi.GenerateContentResponse) (directStreamEvent, error) {
	var delta directStreamEvent
	var text, thinking string
	toolCount := 0
	if len(response.Candidates) > 1 {
		return directStreamEvent{}, fmt.Errorf("Gemini Developer API returned %d candidates; this OpenAI stream adapter supports exactly one", len(response.Candidates))
	}
	if response.UsageMetadata != nil {
		usage := *response.UsageMetadata
		delta.usage = &usage
	}
	for candidateIndex, candidate := range response.Candidates {
		if candidate.FinishReason != "" {
			var err error
			delta.finishReason, err = directFinishReason(candidate.FinishReason)
			if err != nil {
				return directStreamEvent{}, err
			}
			delta.finished = true
		}
		if candidate.Content == nil {
			continue
		}
		for partIndex, part := range candidate.Content.Parts {
			if part.Text != "" {
				if part.Thought {
					thinking += part.Text
				} else {
					text += part.Text
				}
			}
			if part.FunctionCall != nil {
				args, err := json.Marshal(part.FunctionCall.Args)
				if err != nil {
					return directStreamEvent{}, fmt.Errorf("could not encode Gemini function-call arguments: %w", err)
				}
				if toolCount >= format.MaxToolDefinitions {
					return directStreamEvent{}, fmt.Errorf("Gemini Developer API returned more than %d tool calls in one stream event", format.MaxToolDefinitions)
				}
				if err := format.ValidateToolCall(part.FunctionCall.Name, string(args)); err != nil {
					return directStreamEvent{}, fmt.Errorf("invalid Gemini function-call stream event: %w", err)
				}
				toolCount++
				delta.message.ToolCalls = append(delta.message.ToolCalls, models.OpenAIToolCall{
					ID:   fmt.Sprintf("call_gemini_%d_%d", candidateIndex, partIndex),
					Type: "function",
					Function: models.OpenAIToolCallFunction{
						Name:      part.FunctionCall.Name,
						Arguments: string(args),
					},
				})
			}
		}
	}
	delta.message.Content = text
	delta.message.ReasoningContent = thinking
	if len(delta.message.ToolCalls) > 0 && delta.finished {
		delta.finishReason = "tool_calls"
	}
	return delta, nil
}

func directFinishReason(reason string) (string, error) {
	reason = strings.ToUpper(strings.TrimSpace(reason))
	switch reason {
	case "STOP":
		return "stop", nil
	case "MAX_TOKENS":
		return "length", nil
	case "SAFETY", "RECITATION", "BLOCKLIST", "PROHIBITED_CONTENT", "SPII", "LANGUAGE", "IMAGE_SAFETY", "IMAGE_PROHIBITED_CONTENT", "IMAGE_RECITATION":
		return "content_filter", nil
	case "MALFORMED_FUNCTION_CALL":
		return "error", nil
	default:
		return "", fmt.Errorf("unsupported Gemini finish reason %q", reason)
	}
}

func (a *App) handleDirectGeminiStream(w http.ResponseWriter, r *http.Request, req models.OpenAIChatRequest, translated geminiapi.GenerateContentRequest, model, key string) {
	if !startSSE(w) {
		return
	}
	cid := fmt.Sprintf("chatcmpl-%s", format.RandHex(12))
	created := time.Now().Unix()
	var fullText, fullThinking string
	var lastUsage *geminiapi.UsageMetadata
	toolAssembler := newDirectStreamToolAssembler()
	sentFinish := false
	streamFailed := false

	emit := func(response geminiapi.GenerateContentResponse) error {
		delta, err := openAIStreamDelta(response)
		if err != nil {
			return err
		}
		if err := toolAssembler.Observe(delta.message.ToolCalls); err != nil {
			return err
		}
		delta.message.ToolCalls = nil
		if delta.usage != nil {
			lastUsage = delta.usage
		}
		if text, ok := delta.message.Content.(string); ok {
			fullText += text
		}
		fullThinking += delta.message.ReasoningContent
		if delta.message.Content == "" && delta.message.ReasoningContent == "" && !delta.finished {
			return nil
		}
		finish := (*string)(nil)
		if delta.finished && toolAssembler.Len() == 0 {
			reason := delta.finishReason
			finish = &reason
			sentFinish = true
		}
		return writeSSEData(w, models.OpenAIChatResponse{
			ID:      cid,
			Object:  "chat.completion.chunk",
			Created: created,
			Model:   model,
			Choices: []models.OpenAIChoice{{Index: 0, Delta: &delta.message, FinishReason: finish}},
		})
	}

	err := streamGeminiWithKeepAlive(r.Context(), w, 2500*time.Millisecond, func(emitResponse func(geminiapi.GenerateContentResponse) error) error {
		return a.GeminiAPI.Stream(r.Context(), model, key, translated, emitResponse)
	}, emit)
	if err != nil {
		a.Logf("Gemini Developer API stream error: %v", err)
		// Do not serialize provider/transport failures as assistant Markdown.
		// OpenAI stream consumers can classify the structured error after any
		// already-received partial deltas, and the browser can keep it out of
		// successful conversation history.
		_ = writeSSEError(w, err)
		streamFailed = true
	}
	if !streamFailed && toolAssembler.Len() > 0 {
		toolCalls, finalizeErr := toolAssembler.Finalize()
		if finalizeErr != nil {
			a.Logf("Gemini Developer API tool-call finalization error: %v", finalizeErr)
			_ = writeSSEData(w, map[string]any{
				"error": map[string]any{
					"message": finalizeErr.Error(),
					"type":    "api_error",
				},
			})
			streamFailed = true
		} else {
			toolMessage := models.OpenAIMessage{ToolCalls: toolCalls}
			if err := writeSSEData(w, models.OpenAIChatResponse{
				ID: cid, Object: "chat.completion.chunk", Created: created, Model: model,
				Choices: []models.OpenAIChoice{{Index: 0, Delta: &toolMessage, FinishReason: nil}},
			}); err != nil {
				a.Logf("Gemini Developer API tool-call stream write error: %v", err)
			}
			reason := "tool_calls"
			_ = writeSSEData(w, models.OpenAIChatResponse{
				ID: cid, Object: "chat.completion.chunk", Created: created, Model: model,
				Choices: []models.OpenAIChoice{{Index: 0, Delta: &models.OpenAIMessage{}, FinishReason: &reason}},
			})
			sentFinish = true
		}
	}
	if !streamFailed && !sentFinish {
		reason := "stop"
		_ = writeSSEData(w, models.OpenAIChatResponse{
			ID: cid, Object: "chat.completion.chunk", Created: created, Model: model,
			Choices: []models.OpenAIChoice{{Index: 0, Delta: &models.OpenAIMessage{}, FinishReason: &reason}},
		})
	}

	promptTokens, completionTokens, totalTokens, reasoningTokens := 0, 0, 0, 0
	if lastUsage != nil {
		promptTokens = lastUsage.PromptTokenCount
		completionTokens = lastUsage.CandidatesTokenCount
		totalTokens = lastUsage.TotalTokenCount
		reasoningTokens = lastUsage.ThoughtsTokenCount
	}
	usage := directUsage(req, promptTokens, completionTokens, totalTokens, reasoningTokens, fullText, fullThinking)
	a.addEstimatedTokens(uint64(usage.TotalTokens))
	if !streamFailed && req.StreamOptions != nil && req.StreamOptions.IncludeUsage {
		_ = writeSSEData(w, models.OpenAIChatResponse{
			ID: cid, Object: "chat.completion.chunk", Created: created, Model: model,
			Choices: []models.OpenAIChoice{}, Usage: &usage,
		})
	}
	_ = writeSSEDone(w)
}

func streamGeminiWithKeepAlive(ctx context.Context, w http.ResponseWriter, interval time.Duration, run func(func(geminiapi.GenerateContentResponse) error) error, emit func(geminiapi.GenerateContentResponse) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if run == nil {
		return errors.New("stream runner is nil")
	}
	if emit == nil {
		return errors.New("stream emitter is nil")
	}
	if interval <= 0 {
		interval = 2500 * time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	stop := make(chan struct{})
	done := make(chan struct{})
	var mu sync.Mutex
	lastWrite := time.Now()
	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				mu.Lock()
				if time.Since(lastWrite) >= interval-200*time.Millisecond {
					_ = writeSSEComment(w, "keepalive")
					lastWrite = time.Now()
				}
				mu.Unlock()
			}
		}
	}()
	runErr := run(func(response geminiapi.GenerateContentResponse) error {
		mu.Lock()
		defer mu.Unlock()
		lastWrite = time.Now()
		return emit(response)
	})
	close(stop)
	<-done
	return runErr
}

func streamGeminiRawWithKeepAlive(ctx context.Context, w http.ResponseWriter, interval time.Duration, run func(func(json.RawMessage) error) error, emit func(json.RawMessage) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if run == nil {
		return errors.New("stream runner is nil")
	}
	if emit == nil {
		return errors.New("stream emitter is nil")
	}
	if interval <= 0 {
		interval = 2500 * time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	stop := make(chan struct{})
	done := make(chan struct{})
	var mu sync.Mutex
	lastWrite := time.Now()
	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				mu.Lock()
				if time.Since(lastWrite) >= interval-200*time.Millisecond {
					_ = writeSSEComment(w, "keepalive")
					lastWrite = time.Now()
				}
				mu.Unlock()
			}
		}
	}()
	runErr := run(func(event json.RawMessage) error {
		mu.Lock()
		defer mu.Unlock()
		lastWrite = time.Now()
		return emit(event)
	})
	close(stop)
	<-done
	return runErr
}

func writeGeminiAPIError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	var apiErr *geminiapi.APIError
	if errors.As(err, &apiErr) {
		status = ErrorToStatusCode(err)
	}
	if status < http.StatusBadRequest || status > 599 {
		status = http.StatusBadGateway
	}
	message := err.Error()
	if errors.As(err, &apiErr) {
		switch apiErr.Kind {
		case "auth":
			message = "Gemini Developer API rejected the key or project"
		case "quota":
			message = "Gemini Developer API quota or rate limit reached; check Google AI Studio for the project limits"
		case "transport":
			message = "Gemini Developer API could not be reached"
		}
	}
	typeName := "api_error"
	if status == http.StatusBadRequest {
		typeName = "invalid_request_error"
	}
	writeJSON(w, status, map[string]any{"error": map[string]any{
		"message": message,
		"type":    typeName,
	}})
}
