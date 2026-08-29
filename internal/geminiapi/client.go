package geminiapi

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	DefaultBaseURL    = "https://generativelanguage.googleapis.com"
	DefaultMaxBody    = 32 << 20
	APIKeyHeader      = "x-goog-api-key"
	MaxAPIKeyLength   = 4096
	MaxErrorBodyBytes = 64 << 10
)

// Requester is the small transport seam used by deterministic tests and by
// the production http.Client. It keeps the provider path independently
// testable without network access or a Google account.
type Requester interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	HTTP            Requester
	BaseURL         string
	MaxResponseBody int64
}

func NewClient(httpClient Requester) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		HTTP:            httpClient,
		BaseURL:         DefaultBaseURL,
		MaxResponseBody: DefaultMaxBody,
	}
}

// APIError is intentionally sanitized: it never contains request headers or
// the provider key. Status is suitable for mapping to the gateway response.
type APIError struct {
	Status  int
	Kind    string
	Message string
	Err     error
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *APIError) Unwrap() error { return e.Err }

func ValidateAPIKey(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("Gemini Developer API key is empty")
	}
	if strings.ContainsAny(key, "\r\n") {
		return errors.New("Gemini Developer API key contains an invalid newline")
	}
	if len(key) > MaxAPIKeyLength {
		return errors.New("Gemini Developer API key is too long")
	}
	return nil
}

func (c *Client) Generate(ctx context.Context, model, apiKey string, req GenerateContentRequest) (GenerateContentResponse, error) {
	ctx = normalizeContext(ctx)
	body, err := c.GenerateRaw(ctx, model, apiKey, req)
	if err != nil {
		return GenerateContentResponse{}, err
	}

	var response GenerateContentResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return GenerateContentResponse{}, &APIError{
			Kind:    "protocol",
			Message: "Gemini Developer API returned invalid JSON",
			Err:     err,
		}
	}
	return response, nil
}

func (c *Client) GenerateRaw(ctx context.Context, model, apiKey string, body any) ([]byte, error) {
	return c.GenerateRawAction(ctx, model, apiKey, "generateContent", "", body)
}

func (c *Client) GenerateRawAction(ctx context.Context, model, apiKey, action, query string, body any) ([]byte, error) {
	ctx = normalizeContext(ctx)
	if action == "" {
		return nil, &APIError{Kind: "request", Message: "Gemini Developer API action is empty"}
	}
	return c.doJSON(ctx, model, apiKey, action, query, body)
}

func (c *Client) CountTokensRaw(ctx context.Context, model, apiKey string, body any) ([]byte, error) {
	return c.GenerateRawAction(ctx, model, apiKey, "countTokens", "", body)
}

// Stream calls the public Gemini SSE endpoint and decodes each data event.
// Keepalive comments are ignored; malformed provider events fail closed.
func (c *Client) Stream(ctx context.Context, model, apiKey string, req GenerateContentRequest, emit func(GenerateContentResponse) error) error {
	ctx = normalizeContext(ctx)
	if emit == nil {
		return errors.New("Gemini Developer API stream callback is nil")
	}
	return c.StreamRaw(ctx, model, apiKey, req, func(data json.RawMessage) error {
		var response GenerateContentResponse
		if err := json.Unmarshal(data, &response); err != nil {
			return &APIError{Kind: "protocol", Message: "Gemini Developer API returned invalid stream JSON", Err: err}
		}
		return emit(response)
	})
}

func (c *Client) StreamRaw(ctx context.Context, model, apiKey string, body any, emit func(json.RawMessage) error) error {
	return c.StreamRawAction(ctx, model, apiKey, "streamGenerateContent", "alt=sse", body, emit)
}

func (c *Client) StreamRawAction(ctx context.Context, model, apiKey, action, query string, body any, emit func(json.RawMessage) error) error {
	ctx = normalizeContext(ctx)
	if emit == nil {
		return errors.New("Gemini Developer API stream callback is nil")
	}

	validatedKey, err := validatedKey(apiKey)
	if err != nil {
		return err
	}
	endpoint, err := c.endpoint(model, action, query)
	if err != nil {
		return err
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return &APIError{Kind: "request", Message: "could not encode Gemini Developer API request", Err: err}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return &APIError{Kind: "request", Message: "could not create Gemini Developer API request", Err: err}
	}
	setHeaders(req, validatedKey, true)

	response, err := c.requester().Do(req)
	if err != nil {
		return transportError(ctx, err)
	}
	if response == nil {
		return &APIError{Kind: "protocol", Message: "Gemini Developer API returned an empty response"}
	}
	if response.Body == nil {
		return &APIError{Kind: "protocol", Message: "Gemini Developer API returned an empty response body"}
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return c.readHTTPError(response, validatedKey)
	}

	return parseSSE(response.Body, c.maxBody(), emit)
}

func (c *Client) doJSON(ctx context.Context, model, apiKey, action, query string, body any) ([]byte, error) {
	ctx = normalizeContext(ctx)
	validatedKey, err := validatedKey(apiKey)
	if err != nil {
		return nil, err
	}
	endpoint, err := c.endpoint(model, action, query)
	if err != nil {
		return nil, err
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, &APIError{Kind: "request", Message: "could not encode Gemini Developer API request", Err: err}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, &APIError{Kind: "request", Message: "could not create Gemini Developer API request", Err: err}
	}
	setHeaders(req, validatedKey, false)

	response, err := c.requester().Do(req)
	if err != nil {
		return nil, transportError(ctx, err)
	}
	if response == nil {
		return nil, &APIError{Kind: "protocol", Message: "Gemini Developer API returned an empty response"}
	}
	if response.Body == nil {
		return nil, &APIError{Kind: "protocol", Message: "Gemini Developer API returned an empty response body"}
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, c.readHTTPError(response, validatedKey)
	}

	limit := c.maxBody()
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return nil, &APIError{Kind: "transport", Message: "could not read Gemini Developer API response", Err: err}
	}
	if int64(len(data)) > limit {
		return nil, &APIError{Kind: "protocol", Message: "Gemini Developer API response is too large"}
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, &APIError{Kind: "protocol", Message: "Gemini Developer API returned an empty response"}
	}
	return data, nil
}

func normalizeContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func (c *Client) endpoint(model, action, query string) (string, error) {
	model = strings.TrimSpace(strings.TrimPrefix(model, "models/"))
	if model == "" {
		return "", &APIError{Kind: "request", Message: "Gemini Developer API model is empty"}
	}
	if strings.ContainsAny(model, "/?#\r\n") {
		return "", &APIError{Kind: "request", Message: "Gemini Developer API model is invalid"}
	}
	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		base = DefaultBaseURL
	}
	parsed, err := url.Parse(base + "/v1beta/models/" + url.PathEscape(model) + ":" + action)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", &APIError{Kind: "request", Message: "Gemini Developer API endpoint is invalid", Err: err}
	}
	if query != "" {
		parsed.RawQuery = query
	}
	return parsed.String(), nil
}

func (c *Client) requester() Requester {
	if c != nil && c.HTTP != nil {
		return c.HTTP
	}
	return http.DefaultClient
}

func (c *Client) maxBody() int64 {
	if c != nil && c.MaxResponseBody > 0 {
		return c.MaxResponseBody
	}
	return DefaultMaxBody
}

func setHeaders(req *http.Request, key string, stream bool) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if stream {
		req.Header.Set("Accept", "text/event-stream")
	}
	// The public API contract specifies this header. Do not put the key in the
	// URL, where it could leak through proxy logs, browser history, or referrers.
	req.Header.Set(APIKeyHeader, key)
}

func validatedKey(key string) (string, error) {
	key = strings.TrimSpace(key)
	if err := ValidateAPIKey(key); err != nil {
		return "", &APIError{Kind: "request", Message: err.Error(), Err: err}
	}
	return key, nil
}

func transportError(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return &APIError{Kind: "transport", Message: "Gemini Developer API request failed", Err: err}
}

func (c *Client) readHTTPError(response *http.Response, key string) error {
	if response == nil {
		return &APIError{Kind: "protocol", Message: "Gemini Developer API returned an empty response"}
	}
	var data []byte
	if response.Body != nil {
		data, _ = io.ReadAll(io.LimitReader(response.Body, MaxErrorBodyBytes))
	}
	message := redactMessage(providerErrorMessage(data), key)
	if message == "" {
		message = "Gemini Developer API request failed"
	}
	kind := "provider"
	switch response.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		kind = "auth"
	case http.StatusTooManyRequests:
		kind = "quota"
	}
	return &APIError{Status: response.StatusCode, Kind: kind, Message: message}
}

func providerErrorMessage(data []byte) string {
	var envelope struct {
		Error struct {
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}
	if json.Unmarshal(data, &envelope) == nil && envelope.Error.Message != "" {
		message := strings.TrimSpace(envelope.Error.Message)
		if envelope.Error.Status != "" {
			message += " (" + envelope.Error.Status + ")"
		}
		return sanitizeMessage(message)
	}
	return sanitizeMessage(strings.TrimSpace(string(data)))
}

func sanitizeMessage(message string) string {
	message = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' || (r >= 0x20 && r != 0x7f) {
			return r
		}
		return -1
	}, message)
	message = strings.TrimSpace(message)
	if len(message) > 500 {
		message = message[:500] + "..."
	}
	return message
}

func redactMessage(message, secret string) string {
	if secret != "" {
		message = strings.ReplaceAll(message, secret, "[redacted]")
	}
	return sanitizeMessage(message)
}

func parseSSE(reader io.Reader, maxBody int64, emit func(json.RawMessage) error) error {
	if reader == nil {
		return &APIError{Kind: "protocol", Message: "Gemini Developer API returned an empty stream body"}
	}
	scanner := bufio.NewScanner(reader)
	bufferSize := 64 * 1024
	if maxBody > int64(bufferSize) && maxBody < int64(16<<20) {
		bufferSize = int(maxBody)
	}
	scanner.Buffer(make([]byte, 0, bufferSize), int(maxBody))
	var dataLines []string
	var eventBytes int64
	var streamBytes int64
	flush := func() error {
		if len(dataLines) == 0 {
			return nil
		}
		data := strings.TrimSpace(strings.Join(dataLines, "\n"))
		dataLines = nil
		eventBytes = 0
		if data == "" || data == "[DONE]" {
			return nil
		}
		return emit(json.RawMessage(data))
	}
	for scanner.Scan() {
		streamBytes += int64(len(scanner.Bytes())) + 1
		if streamBytes > maxBody {
			return &APIError{Kind: "protocol", Message: "Gemini Developer API stream is too large"}
		}
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if line == "" {
			if err := flush(); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		value := strings.TrimPrefix(line, "data:")
		value = strings.TrimPrefix(value, " ")
		eventBytes += int64(len(value))
		if eventBytes > maxBody {
			return &APIError{Kind: "protocol", Message: "Gemini Developer API stream event is too large"}
		}
		dataLines = append(dataLines, value)
	}
	if err := scanner.Err(); err != nil {
		return &APIError{Kind: "transport", Message: "could not read Gemini Developer API stream", Err: err}
	}
	return flush()
}
