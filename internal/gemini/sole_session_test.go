package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/div197/bob-gemini-free/internal/config"
)

// soleSessionRequester fails the first generation POST at the transport layer
// (no HTTP response), then answers normally. Token-page GETs return empty HTML.
type soleSessionRequester struct {
	posts int
}

func (r *soleSessionRequester) Do(req *http.Request) (*http.Response, error) {
	if req.Method == http.MethodGet {
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("<html></html>"))}, nil
	}
	r.posts++
	if r.posts == 1 {
		return nil, errors.New("connection reset by peer")
	}
	return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(soleSessionBody("recovered answer")))}, nil
}

func soleSessionBody(text string) string {
	padded := text + " " + strings.Repeat("x", 220)
	inner := []any{nil, nil, nil, nil, []any{[]any{nil, []any{padded}}}}
	innerBytes, _ := json.Marshal(inner)
	outer, _ := json.Marshal([]any{[]any{"wrb.fr", nil, string(innerBytes)}})
	return string(outer) + "\n"
}

// TestSoleAccountRecoversAfterTransientFailure is the M4 regression. With one
// pool session, a single connection reset used to park the only account for
// 60s, so every following request reported "all sessions unavailable" without
// reaching Google.
func TestSoleAccountRecoversAfterTransientFailure(t *testing.T) {
	cfg := config.Default()
	cfg.RetryAttempts = 1
	cfg.RetryDelaySec = 0
	pool := NewCookiePool()
	pool.AddSession("student.txt", "SID=sid; SAPISID=sapisid", "sapisid", "")
	requester := &soleSessionRequester{}
	client := &Client{Cfg: cfg, HTTP: requester, Cookies: NewCookieCache(""), Pool: pool}

	if _, err := client.GenerateContext(context.Background(), "hello", 0, 0, nil, nil); err == nil {
		t.Fatal("first request should surface the transport failure")
	}
	text, err := client.GenerateContext(context.Background(), "hello", 0, 0, nil, nil)
	if err != nil {
		t.Fatalf("second request on the sole account failed: %v", err)
	}
	if !strings.Contains(text, "recovered answer") {
		t.Fatalf("second request text = %q", text)
	}
}

// midStreamRequester answers the first streamed POST with a body that fails
// before any line arrives, which is a transport read error mid-stream. The
// second POST streams a normal answer.
type midStreamRequester struct {
	posts int
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("connection reset mid-stream") }

func (r *midStreamRequester) Do(req *http.Request) (*http.Response, error) {
	if req.Method == http.MethodGet {
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("<html></html>"))}, nil
	}
	r.posts++
	if r.posts == 1 {
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(failingReader{})}, nil
	}
	return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(soleSessionBody("recovered stream")))}, nil
}

// TestSoleAccountRecoversAfterMidStreamTransportFailure extends the M4 regression
// to streaming, the main path. A read reset during the stream used to park the
// sole account for 60s.
func TestSoleAccountRecoversAfterMidStreamTransportFailure(t *testing.T) {
	cfg := config.Default()
	cfg.RetryAttempts = 1
	cfg.RetryDelaySec = 0
	pool := NewCookiePool()
	pool.AddSession("student.txt", "SID=sid; SAPISID=sapisid", "sapisid", "")
	client := &Client{Cfg: cfg, HTTP: &midStreamRequester{}, Cookies: NewCookieCache(""), Pool: pool}

	if err := client.GenerateStreamContext(context.Background(), "hello", 0, 0, nil, nil, func(string) error { return nil }); err == nil {
		t.Fatal("first stream should surface the mid-stream failure")
	}
	var got strings.Builder
	err := client.GenerateStreamContext(context.Background(), "hello", 0, 0, nil, nil, func(s string) error {
		got.WriteString(s)
		return nil
	})
	if err != nil {
		t.Fatalf("second stream on the sole account failed: %v", err)
	}
	if !strings.Contains(got.String(), "recovered stream") {
		t.Fatalf("second stream text = %q", got.String())
	}
}
