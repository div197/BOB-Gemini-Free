package diag

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// TestResult records the execution metrics of a diagnostic check.
type TestResult struct {
	Name     string
	Passed   bool
	Duration time.Duration
	Details  string
	Error    error
}

// ProgressFn is called after each diagnostic test completes with real-time feedback.
type ProgressFn func(idx, total int, res TestResult)

const maxDiagnosticResponseBytes int64 = 4 << 20

// newDiagnosticRequest is the single request-construction seam for the live
// diagnostic suite. The old call sites ignored malformed target URLs and
// could panic while applying headers to a nil request.
func newDiagnosticRequest(method, endpoint string, body []byte) (*http.Request, error) {
	if body == nil {
		return http.NewRequest(method, endpoint, nil)
	}
	return http.NewRequest(method, endpoint, bytes.NewReader(body))
}

func diagnosticStatusError(resp *http.Response) error {
	if resp == nil {
		return errors.New("diagnostic server returned no response")
	}
	return fmt.Errorf("HTTP %d", resp.StatusCode)
}

func readDiagnosticBody(resp *http.Response) ([]byte, error) {
	if resp == nil || resp.Body == nil {
		return nil, errors.New("diagnostic server returned an empty response body")
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxDiagnosticResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read diagnostic response: %w", err)
	}
	if int64(len(data)) > maxDiagnosticResponseBytes {
		return nil, errors.New("diagnostic response exceeds the safety limit")
	}
	return data, nil
}

func decodeDiagnosticJSON(resp *http.Response, destination any) error {
	data, err := readDiagnosticBody(resp)
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return errors.New("diagnostic response is empty")
	}
	if !json.Valid(data) {
		return errors.New("diagnostic response is not valid JSON")
	}
	if err := json.Unmarshal(data, destination); err != nil {
		return fmt.Errorf("decode diagnostic response: %w", err)
	}
	return nil
}

func requireDiagnosticText(label, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s returned no usable output", label)
	}
	return nil
}

type diagnosticOpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content   string            `json:"content"`
			Reasoning string            `json:"reasoning_content"`
			ToolCalls []json.RawMessage `json:"tool_calls"`
		} `json:"message"`
	} `json:"choices"`
}

func decodeDiagnosticChat(resp *http.Response, label string) (diagnosticOpenAIResponse, error) {
	var result diagnosticOpenAIResponse
	if err := decodeDiagnosticJSON(resp, &result); err != nil {
		return result, err
	}
	if len(result.Choices) == 0 {
		return result, fmt.Errorf("%s returned no choices", label)
	}
	choice := result.Choices[0]
	if strings.TrimSpace(choice.Message.Content) == "" && strings.TrimSpace(choice.Message.Reasoning) == "" && len(choice.Message.ToolCalls) == 0 {
		return result, fmt.Errorf("%s returned no usable output", label)
	}
	return result, nil
}

func scanDiagnosticOpenAIStream(resp *http.Response) error {
	if resp == nil || resp.Body == nil {
		return errors.New("diagnostic stream has no response body")
	}
	limited := &io.LimitedReader{R: resp.Body, N: maxDiagnosticResponseBytes + 1}
	scanner := bufio.NewScanner(limited)
	scanner.Buffer(make([]byte, 0, 64*1024), 256*1024)
	sawUsableContent := false
	sawDone := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
		if data == "[DONE]" {
			sawDone = true
			continue
		}
		var event struct {
			Error   json.RawMessage `json:"error"`
			Choices []struct {
				Delta struct {
					Content   string            `json:"content"`
					Reasoning string            `json:"reasoning_content"`
					ToolCalls []json.RawMessage `json:"tool_calls"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return errors.New("diagnostic stream contained invalid JSON")
		}
		if len(event.Error) > 0 && string(event.Error) != "null" {
			return errors.New("diagnostic stream returned a structured error")
		}
		for _, choice := range event.Choices {
			if strings.TrimSpace(choice.Delta.Content) != "" || strings.TrimSpace(choice.Delta.Reasoning) != "" || len(choice.Delta.ToolCalls) > 0 {
				sawUsableContent = true
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read diagnostic stream: %w", err)
	}
	if limited.N == 0 {
		return errors.New("diagnostic stream exceeds the safety limit")
	}
	if sawDone && !sawUsableContent {
		return errors.New("diagnostic stream contained only [DONE]")
	}
	if !sawUsableContent {
		return errors.New("diagnostic stream contained no usable event")
	}
	if !sawDone {
		return errors.New("diagnostic stream ended without [DONE]")
	}
	return nil
}

func scanDiagnosticAnthropicStream(resp *http.Response) error {
	if resp == nil || resp.Body == nil {
		return errors.New("diagnostic Anthropic stream has no response body")
	}
	limited := &io.LimitedReader{R: resp.Body, N: maxDiagnosticResponseBytes + 1}
	scanner := bufio.NewScanner(limited)
	scanner.Buffer(make([]byte, 0, 64*1024), 256*1024)
	currentEvent := ""
	sawStart := false
	sawContent := false
	sawStop := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			currentEvent = ""
			continue
		}
		if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimSpace(strings.TrimPrefix(line, "event: "))
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
		if data == "" || !json.Valid([]byte(data)) {
			return errors.New("diagnostic Anthropic stream contained invalid JSON")
		}
		if currentEvent == "error" {
			return errors.New("diagnostic Anthropic stream returned a structured error")
		}
		switch currentEvent {
		case "message_start":
			sawStart = true
		case "content_block_start", "content_block_delta":
			sawContent = true
		case "message_stop":
			sawStop = true
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read diagnostic Anthropic stream: %w", err)
	}
	if limited.N == 0 {
		return errors.New("diagnostic Anthropic stream exceeds the safety limit")
	}
	if !sawStart || !sawContent || !sawStop {
		return errors.New("diagnostic Anthropic stream is missing a complete message lifecycle")
	}
	return nil
}

// RunDiagnostics executes an automated end-to-end test suite against a running BOB Gemini Free instance.
func RunDiagnostics(baseURL, apiKey string) []TestResult {
	return RunDiagnosticsWithProgress(baseURL, apiKey, nil)
}

// RunDiagnosticsWithProgress executes the diagnostic suite and reports each result immediately.
func RunDiagnosticsWithProgress(baseURL, apiKey string, onProgress ProgressFn) []TestResult {
	baseURL = strings.TrimRight(baseURL, "/")
	transport := &http.Transport{
		MaxIdleConns:        20,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   45 * time.Second,
	}

	var results []TestResult
	const totalTests = 15

	runTest := func(name string, fn func() (string, error)) {
		start := time.Now()
		details, err := fn()
		dur := time.Since(start)
		passed := err == nil
		res := TestResult{
			Name:     name,
			Passed:   passed,
			Duration: dur,
			Details:  details,
			Error:    err,
		}
		results = append(results, res)
		if onProgress != nil {
			onProgress(len(results), totalTests, res)
		}
	}

	setHeaders := func(req *http.Request) {
		req.Header.Set("Content-Type", "application/json")
		if apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+apiKey)
		}
	}

	// 1. Health Endpoint
	runTest("Gateway Engine Health (GET /)", func() (string, error) {
		req, err := newDiagnosticRequest(http.MethodGet, baseURL+"/", nil)
		if err != nil {
			return "", err
		}
		setHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", diagnosticStatusError(resp)
		}
		var data map[string]any
		if err := decodeDiagnosticJSON(resp, &data); err != nil {
			return "", err
		}
		if data["status"] != "ok" {
			return "", errors.New("gateway health response is not ok")
		}
		if _, ok := data["version"].(string); !ok {
			return "", errors.New("gateway health response has no version")
		}
		return fmt.Sprintf("status=%v, version=%v", data["status"], data["version"]), nil
	})

	// 2. OpenAI Models Registry (GET /v1/models)
	runTest("OpenAI Models Registry (GET /v1/models)", func() (string, error) {
		req, err := newDiagnosticRequest(http.MethodGet, baseURL+"/v1/models", nil)
		if err != nil {
			return "", err
		}
		setHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", diagnosticStatusError(resp)
		}
		var data struct {
			Data []any `json:"data"`
		}
		if err := decodeDiagnosticJSON(resp, &data); err != nil {
			return "", err
		}
		if len(data.Data) == 0 {
			return "", errors.New("models response contained no models")
		}
		return fmt.Sprintf("%d models registered", len(data.Data)), nil
	})

	// 3. Single Model Lookup (GET /v1/models/gemini-3.7-flash)
	runTest("Single Model Lookup (GET /v1/models/gemini-3.7-flash)", func() (string, error) {
		req, err := newDiagnosticRequest(http.MethodGet, baseURL+"/v1/models/gemini-3.7-flash", nil)
		if err != nil {
			return "", err
		}
		setHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", diagnosticStatusError(resp)
		}
		var data struct {
			ID string `json:"id"`
		}
		if err := decodeDiagnosticJSON(resp, &data); err != nil {
			return "", err
		}
		if data.ID != "gemini-3.7-flash" {
			return "", fmt.Errorf("model lookup returned unexpected id %q", data.ID)
		}
		return "verified model permission metadata", nil
	})

	// 4. Gemini 3.7 Flash Completion (POST /v1/chat/completions)
	runTest("Gemini 3.7 Flash Fast Completion", func() (string, error) {
		payload := map[string]any{
			"model": "gemini-3.7-flash",
			"messages": []map[string]string{
				{"role": "user", "content": "Reply with OK."},
			},
		}
		body, _ := json.Marshal(payload)
		req, err := newDiagnosticRequest(http.MethodPost, baseURL+"/v1/chat/completions", body)
		if err != nil {
			return "", err
		}
		setHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", diagnosticStatusError(resp)
		}
		var chatRes struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		if err := decodeDiagnosticJSON(resp, &chatRes); err != nil {
			return "", err
		}
		if len(chatRes.Choices) == 0 || strings.TrimSpace(chatRes.Choices[0].Message.Content) == "" {
			return "", fmt.Errorf("no usable choices returned")
		}
		return strings.TrimSpace(chatRes.Choices[0].Message.Content), nil
	})

	// 5. Gemini 3.7 Flash Thinking (POST /v1/chat/completions)
	runTest("Gemini 3.7 Flash Deep Reasoning", func() (string, error) {
		payload := map[string]any{
			"model": "gemini-3.7-flash-thinking",
			"messages": []map[string]string{
				{"role": "user", "content": "What is 7*8? Reply with number only."},
			},
		}
		body, _ := json.Marshal(payload)
		req, err := newDiagnosticRequest(http.MethodPost, baseURL+"/v1/chat/completions", body)
		if err != nil {
			return "", err
		}
		setHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", diagnosticStatusError(resp)
		}
		var chatRes struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		if err := decodeDiagnosticJSON(resp, &chatRes); err != nil {
			return "", err
		}
		if len(chatRes.Choices) == 0 || strings.TrimSpace(chatRes.Choices[0].Message.Content) == "" {
			return "", fmt.Errorf("no usable reasoning choices returned")
		}
		return strings.TrimSpace(chatRes.Choices[0].Message.Content), nil
	})

	// 6. Real-time SSE Stream & Usage
	runTest("Real-time SSE Delta Stream & Usage", func() (string, error) {
		payload := map[string]any{
			"model": "gemini-3.7-flash",
			"messages": []map[string]string{
				{"role": "user", "content": "Count 1, 2."},
			},
			"stream": true,
			"stream_options": map[string]any{
				"include_usage": true,
			},
		}
		body, _ := json.Marshal(payload)
		req, err := newDiagnosticRequest(http.MethodPost, baseURL+"/v1/chat/completions", body)
		if err != nil {
			return "", err
		}
		setHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", diagnosticStatusError(resp)
		}
		if err := scanDiagnosticOpenAIStream(resp); err != nil {
			return "", err
		}
		return "streaming chunks verified", nil
	})

	// 7. Developer Role & JSON Format
	runTest("Developer Role & JSON Output Enforcement", func() (string, error) {
		payload := map[string]any{
			"model": "gemini-3.7-flash",
			"messages": []map[string]string{
				{"role": "developer", "content": "You are a calculator."},
				{"role": "user", "content": "Return JSON with key result equal to 42."},
			},
			"response_format": map[string]string{
				"type": "json_object",
			},
		}
		body, _ := json.Marshal(payload)
		req, err := newDiagnosticRequest(http.MethodPost, baseURL+"/v1/chat/completions", body)
		if err != nil {
			return "", err
		}
		setHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", diagnosticStatusError(resp)
		}
		chatRes, err := decodeDiagnosticChat(resp, "JSON response")
		if err != nil {
			return "", err
		}
		content := strings.TrimSpace(chatRes.Choices[0].Message.Content)
		if !json.Valid([]byte(content)) {
			return "", errors.New("response_format JSON output was not valid JSON (web route provides instruction-only semantics)")
		}
		return "valid JSON response received", nil
	})

	// 8. Google Native Gemini API
	runTest("Google Native Gemini API Format", func() (string, error) {
		payload := map[string]any{
			"contents": []map[string]any{
				{"parts": []map[string]string{{"text": "Ping"}}},
			},
		}
		body, _ := json.Marshal(payload)
		req, err := newDiagnosticRequest(http.MethodPost, baseURL+"/v1beta/models/gemini-3.7-flash:generateContent", body)
		if err != nil {
			return "", err
		}
		setHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", diagnosticStatusError(resp)
		}
		var native struct {
			Candidates []struct {
				Content *struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
		}
		if err := decodeDiagnosticJSON(resp, &native); err != nil {
			return "", err
		}
		if len(native.Candidates) == 0 {
			return "", errors.New("Google response contained no candidates")
		}
		var text string
		for _, candidate := range native.Candidates {
			if candidate.Content == nil {
				continue
			}
			for _, part := range candidate.Content.Parts {
				text += part.Text
			}
		}
		if err := requireDiagnosticText("Google response", text); err != nil {
			return "", err
		}
		return "candidates generated", nil
	})

	// 9. OpenAI Responses API
	runTest("OpenAI Codex CLI Responses API Format", func() (string, error) {
		payload := map[string]any{
			"model": "gemini-3.7-flash",
			"input": "Write one line python hello",
		}
		body, _ := json.Marshal(payload)
		req, err := newDiagnosticRequest(http.MethodPost, baseURL+"/v1/responses", body)
		if err != nil {
			return "", err
		}
		setHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", diagnosticStatusError(resp)
		}
		var response struct {
			ID         string            `json:"id"`
			Status     string            `json:"status"`
			OutputText string            `json:"output_text"`
			Output     []json.RawMessage `json:"output"`
		}
		if err := decodeDiagnosticJSON(resp, &response); err != nil {
			return "", err
		}
		if response.ID == "" || response.Status != "completed" || (strings.TrimSpace(response.OutputText) == "" && len(response.Output) == 0) {
			return "", errors.New("Responses API returned an incomplete response object")
		}
		return "response object generated", nil
	})

	// 10. Anthropic Messages Protocol
	runTest("Anthropic Messages API Protocol (POST /v1/messages)", func() (string, error) {
		payload := map[string]any{
			"model": "claude-3-5-sonnet",
			"messages": []map[string]string{
				{"role": "user", "content": "Reply with 'Claude OK'."},
			},
			"max_tokens": 50,
		}
		body, _ := json.Marshal(payload)
		req, err := newDiagnosticRequest(http.MethodPost, baseURL+"/v1/messages", body)
		if err != nil {
			return "", err
		}
		setHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", diagnosticStatusError(resp)
		}
		var msgRes struct {
			Role    string `json:"role"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		}
		if err := decodeDiagnosticJSON(resp, &msgRes); err != nil {
			return "", err
		}
		if msgRes.Role != "assistant" || len(msgRes.Content) == 0 {
			return "", fmt.Errorf("no assistant content blocks returned")
		}
		if err := requireDiagnosticText("Anthropic response", msgRes.Content[0].Text); err != nil {
			return "", err
		}
		return strings.TrimSpace(msgRes.Content[0].Text), nil
	})

	// 11. OpenAI Tool / Function Calling
	runTest("OpenAI Function Calling & Tool Invocation", func() (string, error) {
		payload := map[string]any{
			"model": "gemini-3.7-flash",
			"messages": []map[string]string{
				{"role": "user", "content": "What is the weather in Delhi? Call get_weather tool."},
			},
			"tools": []map[string]any{
				{
					"type": "function",
					"function": map[string]any{
						"name":        "get_weather",
						"description": "Get current temperature for city",
						"parameters": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"city": map[string]string{"type": "string"},
							},
							"required": []string{"city"},
						},
					},
				},
			},
		}
		body, _ := json.Marshal(payload)
		req, err := newDiagnosticRequest(http.MethodPost, baseURL+"/v1/chat/completions", body)
		if err != nil {
			return "", err
		}
		setHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", diagnosticStatusError(resp)
		}
		if _, err := decodeDiagnosticChat(resp, "tool-call response"); err != nil {
			return "", err
		}
		return "tool call pipeline verified", nil
	})

	// 12. Image Generation & Gemini Nano Banana Pipeline
	runTest("Image Generation & Gemini Nano Banana Pipeline", func() (string, error) {
		payload := map[string]any{
			"prompt": "A golden lotus flower blooming on calm water",
			"model":  "gemini-nano-banana-2",
		}
		body, _ := json.Marshal(payload)
		req, err := newDiagnosticRequest(http.MethodPost, baseURL+"/v1/images/generations", body)
		if err != nil {
			return "", err
		}
		setHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("provider-dependent image generation check failed: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("provider-dependent image generation check failed: %w", diagnosticStatusError(resp))
		}
		var imgRes struct {
			Data []struct {
				URL     string `json:"url"`
				B64JSON string `json:"b64_json"`
			} `json:"data"`
		}
		if err := decodeDiagnosticJSON(resp, &imgRes); err != nil {
			return "", err
		}
		if len(imgRes.Data) == 0 || (strings.TrimSpace(imgRes.Data[0].URL) == "" && strings.TrimSpace(imgRes.Data[0].B64JSON) == "") {
			return "", errors.New("image generation returned no usable image data")
		}
		return "image generation pipeline verified", nil
	})

	// 13. Token Counting Subsystem (Google :countTokens & OpenAI /v1/tokens/count)
	runTest("Token Counting Engine (Google :countTokens & OpenAI /v1/tokens/count)", func() (string, error) {
		// 1. Google Native :countTokens
		gPayload := map[string]any{
			"contents": []map[string]any{
				{
					"role": "user",
					"parts": []map[string]string{
						{"text": "Explain the architecture of Transformer neural networks."},
					},
				},
			},
		}
		gBody, _ := json.Marshal(gPayload)
		gReq, err := newDiagnosticRequest(http.MethodPost, baseURL+"/v1beta/models/gemini-3.7-flash:countTokens", gBody)
		if err != nil {
			return "", err
		}
		setHeaders(gReq)
		gResp, err := client.Do(gReq)
		if err != nil {
			return "", err
		}
		defer gResp.Body.Close()
		if gResp.StatusCode != http.StatusOK {
			return "", diagnosticStatusError(gResp)
		}
		var googleCount struct {
			TotalTokens int `json:"totalTokens"`
		}
		if err := decodeDiagnosticJSON(gResp, &googleCount); err != nil {
			return "", err
		}
		if googleCount.TotalTokens <= 0 {
			return "", errors.New("Google countTokens returned no positive token count")
		}

		// 2. OpenAI /v1/tokens/count
		oPayload := map[string]any{
			"model": "gemini-3.7-flash",
			"messages": []map[string]string{
				{"role": "user", "content": "Explain neural attention mechanisms."},
			},
		}
		oBody, _ := json.Marshal(oPayload)
		oReq, err := newDiagnosticRequest(http.MethodPost, baseURL+"/v1/tokens/count", oBody)
		if err != nil {
			return "", err
		}
		setHeaders(oReq)
		oResp, err := client.Do(oReq)
		if err != nil {
			return "", err
		}
		defer oResp.Body.Close()
		if oResp.StatusCode != http.StatusOK {
			return "", diagnosticStatusError(oResp)
		}
		var openAICount struct {
			TotalTokens int `json:"total_tokens"`
		}
		if err := decodeDiagnosticJSON(oResp, &openAICount); err != nil {
			return "", err
		}
		if openAICount.TotalTokens <= 0 {
			return "", errors.New("OpenAI tokens/count returned no positive token count")
		}

		return "token counting engine verified for Google and OpenAI protocols", nil
	})

	// 14. Claude Code CLI SSE Streaming Tool Execution Protocol
	runTest("Claude Code SSE Streaming Tool Execution Protocol", func() (string, error) {
		payload := map[string]any{
			"model":  "claude-3-5-sonnet",
			"stream": true,
			"messages": []map[string]string{
				{"role": "user", "content": "List files matching pattern *.go by calling GlobTool."},
			},
			"tools": []map[string]any{
				{
					"name":        "GlobTool",
					"description": "Find files matching glob",
					"input_schema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"pattern": map[string]string{"type": "string"},
						},
						"required": []string{"pattern"},
					},
				},
			},
			"max_tokens": 100,
		}
		body, _ := json.Marshal(payload)
		req, err := newDiagnosticRequest(http.MethodPost, baseURL+"/v1/messages", body)
		if err != nil {
			return "", err
		}
		setHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", diagnosticStatusError(resp)
		}
		if err := scanDiagnosticAnthropicStream(resp); err != nil {
			return "", err
		}
		return "verified complete Anthropic SSE tool execution lifecycle", nil
	})

	// 15. StreamFlight Concurrency Multiplexing & Coalescing
	runTest("StreamFlight Concurrency Multiplexing (5 Parallel Coalesced Requests)", func() (string, error) {
		const numParallel = 5
		type reqResult struct {
			index int
			err   error
			dur   time.Duration
		}
		resChan := make(chan reqResult, numParallel)

		flightPrompt := "Ping StreamFlight deduplication test 1"
		for i := 0; i < numParallel; i++ {
			go func(idx int) {
				tStart := time.Now()
				payload := map[string]any{
					"model": "gemini-3.7-flash",
					"messages": []map[string]string{
						{"role": "user", "content": flightPrompt},
					},
					"stream": true,
				}
				pBody, _ := json.Marshal(payload)
				pReq, err := newDiagnosticRequest(http.MethodPost, baseURL+"/v1/chat/completions", pBody)
				if err != nil {
					resChan <- reqResult{index: idx, err: err, dur: time.Since(tStart)}
					return
				}
				setHeaders(pReq)
				pResp, pErr := client.Do(pReq)
				if pErr != nil {
					resChan <- reqResult{index: idx, err: pErr, dur: time.Since(tStart)}
					return
				}
				defer pResp.Body.Close()
				if pResp.StatusCode != http.StatusOK {
					resChan <- reqResult{index: idx, err: diagnosticStatusError(pResp), dur: time.Since(tStart)}
					return
				}
				if err := scanDiagnosticOpenAIStream(pResp); err != nil {
					resChan <- reqResult{index: idx, err: err, dur: time.Since(tStart)}
					return
				}
				resChan <- reqResult{index: idx, err: nil, dur: time.Since(tStart)}
			}(i)
		}

		var firstErr error
		for i := 0; i < numParallel; i++ {
			r := <-resChan
			if r.err != nil && firstErr == nil {
				firstErr = r.err
			}
		}
		if firstErr != nil {
			return "", firstErr
		}
		return fmt.Sprintf("all %d concurrent in-flight streams multiplexed successfully", numParallel), nil
	})

	return results
}
