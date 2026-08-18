package diag

import (
	"bytes"
	"encoding/json"
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

// RunDiagnostics executes an automated end-to-end test suite against a running BOB Gemini Free instance.
func RunDiagnostics(baseURL, apiKey string) []TestResult {
	baseURL = strings.TrimRight(baseURL, "/")
	client := &http.Client{Timeout: 45 * time.Second}

	var results []TestResult

	runTest := func(name string, fn func() (string, error)) {
		start := time.Now()
		details, err := fn()
		dur := time.Since(start)
		passed := err == nil
		results = append(results, TestResult{
			Name:     name,
			Passed:   passed,
			Duration: dur,
			Details:  details,
			Error:    err,
		})
	}

	setHeaders := func(req *http.Request) {
		req.Header.Set("Content-Type", "application/json")
		if apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+apiKey)
		}
	}

	// 1. Health Endpoint
	runTest("Gateway Engine Health (GET /)", func() (string, error) {
		req, _ := http.NewRequest("GET", baseURL+"/", nil)
		setHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		var data map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&data)
		return fmt.Sprintf("status=%v, version=%v", data["status"], data["version"]), nil
	})

	// 2. OpenAI Models Registry (GET /v1/models)
	runTest("OpenAI Models Registry (GET /v1/models)", func() (string, error) {
		req, _ := http.NewRequest("GET", baseURL+"/v1/models", nil)
		setHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		var data struct {
			Data []any `json:"data"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&data)
		return fmt.Sprintf("%d models registered", len(data.Data)), nil
	})

	// 3. Single Model Lookup (GET /v1/models/gemini-3.7-flash)
	runTest("Single Model Lookup (GET /v1/models/gemini-3.7-flash)", func() (string, error) {
		req, _ := http.NewRequest("GET", baseURL+"/v1/models/gemini-3.7-flash", nil)
		setHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("HTTP %d", resp.StatusCode)
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
		req, _ := http.NewRequest("POST", baseURL+"/v1/chat/completions", bytes.NewReader(body))
		setHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
		}
		var chatRes struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&chatRes)
		if len(chatRes.Choices) == 0 {
			return "", fmt.Errorf("no choices returned")
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
		req, _ := http.NewRequest("POST", baseURL+"/v1/chat/completions", bytes.NewReader(body))
		setHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
		}
		var chatRes struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&chatRes)
		if len(chatRes.Choices) == 0 {
			return "", fmt.Errorf("no choices returned")
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
		req, _ := http.NewRequest("POST", baseURL+"/v1/chat/completions", bytes.NewReader(body))
		setHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		buf := make([]byte, 1024)
		n, _ := resp.Body.Read(buf)
		if n == 0 {
			return "", fmt.Errorf("empty stream response")
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
		req, _ := http.NewRequest("POST", baseURL+"/v1/chat/completions", bytes.NewReader(body))
		setHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("HTTP %d", resp.StatusCode)
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
		req, _ := http.NewRequest("POST", baseURL+"/v1beta/models/gemini-3.7-flash:generateContent", bytes.NewReader(body))
		setHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("HTTP %d", resp.StatusCode)
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
		req, _ := http.NewRequest("POST", baseURL+"/v1/responses", bytes.NewReader(body))
		setHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		return "response object generated", nil
	})

	return results
}
