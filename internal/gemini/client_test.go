package gemini

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/div197/bob-gemini-free/internal/config"
)

func testClientWithRequester(requester Requester, attempts int) *Client {
	cfg := config.Default()
	cfg.RetryAttempts = attempts
	cfg.RetryDelaySec = 0
	return &Client{
		Cfg:     cfg,
		HTTP:    requester,
		Cookies: NewCookieCache(""),
		Pool:    NewCookiePool(),
		Logf:    func(string, ...any) {},
	}
}

func TestShouldRetryUpstreamOnlyRetriesTransientFailures(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "transport", err: &UpstreamError{Kind: "transport", Msg: "connection reset"}, want: true},
		{name: "server", err: &UpstreamError{Kind: "http", Status: http.StatusBadGateway, Msg: "upstream HTTP 502"}, want: true},
		{name: "redirect", err: &UpstreamError{Kind: "http", Status: http.StatusFound, Msg: "policy rejection"}, want: false},
		{name: "unauthorized", err: &UpstreamError{Kind: "http", Status: http.StatusUnauthorized, Msg: "upstream HTTP 401"}, want: false},
		{name: "rate-limit", err: &UpstreamError{Kind: "http", Status: http.StatusTooManyRequests, Msg: "rate limited"}, want: false},
		{name: "bard-rejection", err: &UpstreamError{Kind: "bard", Msg: "BardErrorInfo [42901]"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldRetryUpstream(tt.err); got != tt.want {
				t.Fatalf("shouldRetryUpstream(%v) = %t, want %t", tt.err, got, tt.want)
			}
		})
	}
}

func TestGenerateDoesNotAmplifyHTTP429(t *testing.T) {
	requester := &goldenRequester{responses: []goldenHTTPResponse{{
		response: &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("rate limited")),
		},
	}}}
	client := testClientWithRequester(requester, 3)

	_, err := client.GenerateContext(context.Background(), "fixture", 1, 4, nil, nil)
	if err == nil {
		t.Fatal("HTTP 429 unexpectedly succeeded")
	}
	var upstreamErr *UpstreamError
	if !errors.As(err, &upstreamErr) || upstreamErr.Status != http.StatusTooManyRequests {
		t.Fatalf("error = %#v, want UpstreamError status 429", err)
	}
	requester.mu.Lock()
	called := requester.called
	requester.mu.Unlock()
	if called != 1 {
		t.Fatalf("HTTP 429 request count = %d, want 1", called)
	}
}

func TestGenerateDoesNotAmplifyBardRateLimit(t *testing.T) {
	requester := &goldenRequester{responses: []goldenHTTPResponse{{
		response: goldenOK(io.NopCloser(strings.NewReader("BardErrorInfo [42901]"))),
	}}}
	client := testClientWithRequester(requester, 3)

	_, err := client.GenerateContext(context.Background(), "fixture", 1, 4, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "42901") {
		t.Fatalf("Bard rate-limit error = %v", err)
	}
	requester.mu.Lock()
	called := requester.called
	requester.mu.Unlock()
	if called != 1 {
		t.Fatalf("Bard rate-limit request count = %d, want 1", called)
	}
}

func TestGenerateRetriesTransportFailure(t *testing.T) {
	requester := &goldenRequester{responses: []goldenHTTPResponse{
		{err: errors.New("temporary connection reset")},
		{response: goldenOK(io.NopCloser(strings.NewReader(goldenBoQLine("recovered response"))))},
	}}
	client := testClientWithRequester(requester, 2)

	got, err := client.GenerateContext(context.Background(), "fixture", 1, 4, nil, nil)
	if err != nil {
		t.Fatalf("GenerateContext: %v", err)
	}
	if got != "recovered response" {
		t.Fatalf("response = %q, want recovered response", got)
	}
	requester.mu.Lock()
	called := requester.called
	requester.mu.Unlock()
	if called != 2 {
		t.Fatalf("transport retry request count = %d, want 2", called)
	}
}

func TestGenerateStreamDoesNotAmplifyHTTP403(t *testing.T) {
	requester := &goldenRequester{responses: []goldenHTTPResponse{{
		response: &http.Response{
			StatusCode: http.StatusForbidden,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("forbidden")),
		},
	}}}
	client := testClientWithRequester(requester, 3)

	err := client.GenerateStreamContext(context.Background(), "fixture", 1, 4, nil, nil, func(string) error { return nil })
	if err == nil {
		t.Fatal("HTTP 403 unexpectedly succeeded")
	}
	requester.mu.Lock()
	called := requester.called
	requester.mu.Unlock()
	if called != 1 {
		t.Fatalf("HTTP 403 stream request count = %d, want 1", called)
	}
}
