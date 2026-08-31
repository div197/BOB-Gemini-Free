package gemini

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

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

func TestUpstreamLogMessageDoesNotExposeTransportURL(t *testing.T) {
	secretURL := "https://gemini.google.com/StreamGenerate?bl=short-lived-token"
	err := &UpstreamError{Kind: "transport", Msg: "Post " + secretURL + ": connection reset"}
	message := upstreamLogMessage(err)
	if strings.Contains(message, "short-lived-token") || strings.Contains(message, "StreamGenerate") {
		t.Fatalf("transport URL leaked through retry log message: %q", message)
	}
	if message != "transport failure" {
		t.Fatalf("retry log message = %q", message)
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

func TestTriageStatusCapturesBoundedRetryAfter(t *testing.T) {
	client := &Client{}
	resp := &http.Response{
		StatusCode: http.StatusServiceUnavailable,
		Header:     http.Header{"Retry-After": []string{"7"}},
	}
	err := client.triageStatus(resp)
	var upstreamErr *UpstreamError
	if !errors.As(err, &upstreamErr) {
		t.Fatalf("triage error = %v, want UpstreamError", err)
	}
	if upstreamErr.RetryAfter != 7*time.Second {
		t.Fatalf("retry-after = %s, want 7s", upstreamErr.RetryAfter)
	}
	if got := parseRetryAfter("999999999"); got != maxUpstreamRetryBackoff {
		t.Fatalf("oversized retry-after = %s, want %s", got, maxUpstreamRetryBackoff)
	}
	if got := parseRetryAfter("not-a-delay"); got != 0 {
		t.Fatalf("invalid retry-after = %s, want 0", got)
	}
}

func TestTriageStatusClassifiesSessionAndQuotaFailures(t *testing.T) {
	for _, test := range []struct {
		status int
		kind   string
		text   string
	}{
		{status: http.StatusUnauthorized, kind: "auth", text: "session or request authentication"},
		{status: http.StatusForbidden, kind: "auth", text: "session or request authentication"},
		{status: http.StatusTooManyRequests, kind: "quota", text: "rate limited"},
	} {
		resp := &http.Response{StatusCode: test.status, Header: make(http.Header)}
		err := (&Client{}).triageStatus(resp)
		var upstreamErr *UpstreamError
		if !errors.As(err, &upstreamErr) {
			t.Fatalf("status %d error = %v, want UpstreamError", test.status, err)
		}
		if upstreamErr.Kind != test.kind || !strings.Contains(upstreamErr.Msg, test.text) {
			t.Fatalf("status %d error = %#v, want kind %q and text %q", test.status, upstreamErr, test.kind, test.text)
		}
	}
}

func TestUpstreamRetryDelayIsCappedAndJittered(t *testing.T) {
	for attempt := 0; attempt < 8; attempt++ {
		got := upstreamRetryDelay(attempt, 2)
		if got < time.Second || got > maxUpstreamRetryBackoff {
			t.Fatalf("attempt %d delay = %s, want [1s, %s]", attempt, got, maxUpstreamRetryBackoff)
		}
	}
	if got := upstreamRetryDelay(4, 0); got != 0 {
		t.Fatalf("zero-base retry delay = %s, want 0", got)
	}
}

func TestWaitForUpstreamRetryHonorsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	started := time.Now()
	err := waitForUpstreamRetry(ctx, &UpstreamError{RetryAfter: 5 * time.Second}, 0, 2)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("wait error = %v, want context.Canceled", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("cancelled retry wait took %s", elapsed)
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
	primeCachedSession(client.Cookies)

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
	assertSessionTokens(t, client.Cookies, "cached-at", "cached-bl")
}

func TestGenerateInvalidatesCachedSessionAfterHTTP401(t *testing.T) {
	requester := &goldenRequester{responses: []goldenHTTPResponse{{
		response: &http.Response{
			StatusCode: http.StatusUnauthorized,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("unauthorized")),
		},
	}}}
	client := testClientWithRequester(requester, 3)
	primeCachedSession(client.Cookies)

	_, err := client.GenerateContext(context.Background(), "fixture", 1, 4, nil, nil)
	var upstreamErr *UpstreamError
	if !errors.As(err, &upstreamErr) || upstreamErr.Kind != "auth" {
		t.Fatalf("HTTP 401 error = %#v, want auth UpstreamError", err)
	}
	assertSessionTokens(t, client.Cookies, "", "")
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

func TestGenerateRetriesPreRequestDialFailure(t *testing.T) {
	requester := &goldenRequester{responses: []goldenHTTPResponse{
		{err: &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}},
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
		t.Fatalf("pre-request dial retry request count = %d, want 2", called)
	}
}

func TestGenerateRetryWithNilLoggerDoesNotPanic(t *testing.T) {
	requester := &goldenRequester{responses: []goldenHTTPResponse{
		{err: &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}},
		{err: &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}},
	}}
	client := testClientWithRequester(requester, 2)
	client.Logf = nil

	_, err := client.GenerateContext(context.Background(), "fixture", 1, 4, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "connection refused") {
		t.Fatalf("GenerateContext error = %v, want the final dial failure", err)
	}
}

func TestGenerateStreamRetryWithNilLoggerDoesNotPanic(t *testing.T) {
	requester := &goldenRequester{responses: []goldenHTTPResponse{
		{err: &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}},
		{response: goldenOK(io.NopCloser(strings.NewReader(goldenBoQLine("recovered stream"))))},
	}}
	client := testClientWithRequester(requester, 2)
	client.Logf = nil

	var got []string
	err := client.GenerateStreamContext(context.Background(), "fixture", 1, 4, nil, nil, func(delta string) error {
		got = append(got, delta)
		return nil
	})
	if err != nil {
		t.Fatalf("GenerateStreamContext: %v", err)
	}
	if strings.Join(got, "") != "recovered stream" {
		t.Fatalf("stream output = %q, want recovered stream", strings.Join(got, ""))
	}
}

func TestGenerateDoesNotRetryAmbiguousTransportFailure(t *testing.T) {
	requester := &goldenRequester{responses: []goldenHTTPResponse{
		{err: errors.New("connection reset after request write")},
		{response: goldenOK(io.NopCloser(strings.NewReader(goldenBoQLine("must not replay"))))},
	}}
	client := testClientWithRequester(requester, 2)

	_, err := client.GenerateContext(context.Background(), "fixture", 1, 4, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "connection reset") {
		t.Fatalf("ambiguous transport error = %v", err)
	}
	requester.mu.Lock()
	called := requester.called
	requester.mu.Unlock()
	if called != 1 {
		t.Fatalf("ambiguous transport request count = %d, want 1", called)
	}
}

func TestGenerateRejectsOversizedUpstreamResponse(t *testing.T) {
	requester := &goldenRequester{responses: []goldenHTTPResponse{{
		response: goldenOK(io.NopCloser(strings.NewReader(strings.Repeat("x", maxUpstreamResponseBytes+1)))),
	}}}
	client := testClientWithRequester(requester, 1)

	_, err := client.GenerateContext(context.Background(), "fixture", 1, 4, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "exceeded") {
		t.Fatalf("oversized response error = %v", err)
	}
	requester.mu.Lock()
	called := requester.called
	requester.mu.Unlock()
	if called != 1 {
		t.Fatalf("oversized response request count = %d, want 1", called)
	}
}

func TestGenerateRejectsNilUpstreamResponseWithoutRetry(t *testing.T) {
	requester := &goldenRequester{responses: []goldenHTTPResponse{{}}}
	client := testClientWithRequester(requester, 3)

	_, err := client.GenerateContext(context.Background(), "fixture", 1, 4, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "empty response") {
		t.Fatalf("nil response error = %v", err)
	}
	requester.mu.Lock()
	called := requester.called
	requester.mu.Unlock()
	if called != 1 {
		t.Fatalf("nil response request count = %d, want 1", called)
	}
}

func TestGenerateRejectsEmptyUpstreamText(t *testing.T) {
	requester := &goldenRequester{responses: []goldenHTTPResponse{{
		response: goldenOK(io.NopCloser(strings.NewReader("[]"))),
	}}}
	client := testClientWithRequester(requester, 3)

	_, err := client.GenerateContext(context.Background(), "fixture", 1, 4, nil, nil)
	var upstreamErr *UpstreamError
	if !errors.As(err, &upstreamErr) || upstreamErr.Kind != "protocol" || !strings.Contains(err.Error(), "no usable text") {
		t.Fatalf("empty upstream text error = %#v, want protocol failure", err)
	}
	requester.mu.Lock()
	called := requester.called
	requester.mu.Unlock()
	if called != 1 {
		t.Fatalf("empty upstream text request count = %d, want 1", called)
	}
}

func TestGenerateStreamRejectsOversizedLineWithoutRetry(t *testing.T) {
	requester := &goldenRequester{responses: []goldenHTTPResponse{{
		response: goldenOK(io.NopCloser(strings.NewReader(strings.Repeat("x", maxUpstreamStreamLineBytes+1) + "\n"))),
	}}}
	client := testClientWithRequester(requester, 3)

	err := client.GenerateStreamContext(context.Background(), "fixture", 1, 4, nil, nil, func(string) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "stream line exceeded") {
		t.Fatalf("oversized stream line error = %v", err)
	}
	requester.mu.Lock()
	called := requester.called
	requester.mu.Unlock()
	if called != 1 {
		t.Fatalf("oversized stream line request count = %d, want 1", called)
	}
}

func TestStreamAttemptRejectsAggregateByteLimit(t *testing.T) {
	client := &Client{}
	streamBytes := int64(maxUpstreamStreamBytes)
	err := client.streamAttempt(strings.NewReader("ignored\n"), NewStreamParser(), func(string) error { return nil }, &streamBytes)
	if err == nil || !strings.Contains(err.Error(), "stream exceeded") {
		t.Fatalf("aggregate stream error = %v", err)
	}
}

func TestStreamAttemptRejectsMissingParserOrCallback(t *testing.T) {
	client := &Client{}
	if err := client.streamAttempt(strings.NewReader("ignored\n"), nil, func(string) error { return nil }, nil); err == nil || !strings.Contains(err.Error(), "parser") {
		t.Fatalf("nil parser error = %v", err)
	}
	if err := client.streamAttempt(strings.NewReader("ignored\n"), NewStreamParser(), nil, nil); err == nil || !strings.Contains(err.Error(), "callback") {
		t.Fatalf("nil callback error = %v", err)
	}
}

func TestStreamAttemptRejectsEmptyUsableOutput(t *testing.T) {
	for _, body := range []string{"", "ignored\n"} {
		err := (&Client{}).streamAttempt(strings.NewReader(body), NewStreamParser(), func(string) error { return nil }, nil)
		if err == nil || !strings.Contains(err.Error(), "no usable text") {
			t.Fatalf("body %q error = %v, want explicit empty-output failure", body, err)
		}
	}
}

func TestClientStopsWhenAllConfiguredSessionsAreCoolingDown(t *testing.T) {
	requester := &goldenRequester{}
	client := testClientWithRequester(requester, 3)
	client.Pool.AddSession("student.txt", "SID=sid; SAPISID=sapisid", "sapisid", "")
	session := client.Pool.GetHealthySession()
	if session == nil {
		t.Fatal("configured session was not available before failure")
	}
	client.Pool.MarkFailure(session.ID)

	_, err := client.GenerateContext(context.Background(), "fixture", 1, 4, nil, nil)
	var upstreamErr *UpstreamError
	if !errors.As(err, &upstreamErr) || upstreamErr.Kind != "session" {
		t.Fatalf("cooldown error = %#v, want session UpstreamError", err)
	}
	requester.mu.Lock()
	called := requester.called
	requester.mu.Unlock()
	if called != 0 {
		t.Fatalf("cooldown bypass made %d upstream requests", called)
	}
}

func TestNewClientConfiguresTotalTimeout(t *testing.T) {
	cfg := config.Default()
	cfg.RequestTimeoutSec = 7
	client := NewClient(cfg)
	httpClient, ok := client.HTTP.(*http.Client)
	if !ok {
		t.Fatalf("HTTP requester type = %T, want *http.Client", client.HTTP)
	}
	if got := httpClient.Timeout; got != 7*time.Second {
		t.Fatalf("HTTP client timeout = %s, want 7s", got)
	}
}

func TestClientOnlyCoalescesAnonymousRequests(t *testing.T) {
	client := testClientWithRequester(&goldenRequester{}, 1)
	client.Flight = NewStreamFlight()

	if key := client.requestFlightKey("same prompt", 1, 4, nil); key == "" {
		t.Fatal("anonymous request unexpectedly disabled from-flight coalescing")
	}

	client.Cfg.CookieFile = "student-session.txt"
	if key := client.requestFlightKey("same prompt", 1, 4, nil); key != "" {
		t.Fatalf("configured cookie request received a coalescing key %q", key)
	}

	client.Cfg.CookieFile = ""
	client.Pool.AddSession("student-session.txt", "SID=sid; SAPISID=sapisid", "sapisid", "")
	if key := client.requestFlightKey("same prompt", 1, 4, nil); key != "" {
		t.Fatalf("loaded session request received a coalescing key %q", key)
	}
}

func TestConfiguredSessionPoolNeverFallsBackToGuestCookie(t *testing.T) {
	client := testClientWithRequester(&goldenRequester{}, 1)
	client.Pool.AddSession("student-session.txt", "SID=student; SAPISID=student-sapi", "student-sapi", "")

	headers := client.buildHeaders(nil, "guest-cookie=1")
	if got := headers.Get("Cookie"); got != "" {
		t.Fatalf("configured session pool downgraded to guest cookie %q", got)
	}
}

func TestConfiguredInvalidCookieNeverFallsBackToGuestCookie(t *testing.T) {
	client := testClientWithRequester(&goldenRequester{}, 1)
	client.Cfg.CookieFile = "missing-or-invalid-cookie.txt"

	headers := client.buildHeaders(nil, "guest-cookie=1")
	if got := headers.Get("Cookie"); got != "" {
		t.Fatalf("configured invalid cookie downgraded to guest cookie %q", got)
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
	primeCachedSession(client.Cookies)

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
	assertSessionTokens(t, client.Cookies, "", "")
}

func primeCachedSession(cache *CookieCache) {
	cache.mu.Lock()
	cache.info.At = "cached-at"
	cache.info.BL = "cached-bl"
	cache.info.AtTime = time.Now()
	cache.mu.Unlock()
}

func assertSessionTokens(t *testing.T, cache *CookieCache, wantAt, wantBL string) {
	t.Helper()
	info, err := cache.Load()
	if err != nil {
		t.Fatalf("Load session cache: %v", err)
	}
	if info.At != wantAt || info.BL != wantBL {
		t.Fatalf("session tokens = %q %q, want %q %q", info.At, info.BL, wantAt, wantBL)
	}
}
