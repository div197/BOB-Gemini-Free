package gemini

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/metrics"
)

type UpstreamError struct {
	Status     int
	Kind       string
	Msg        string
	RetryAfter time.Duration
}

func (e *UpstreamError) Error() string {
	return e.Msg
}

type Requester interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	Cfg     config.Config
	HTTP    Requester
	Cookies *CookieCache
	Pool    *CookiePool
	Flight  *StreamFlight
	Logf    func(format string, args ...any)
	Metrics *metrics.Registry
}

const (
	maxUpstreamResponseBytes   = 32 << 20
	maxUpstreamStreamLineBytes = 16 << 20
	maxUpstreamStreamBytes     = 32 << 20
)

var errUpstreamStreamLineTooLarge = errors.New("upstream stream line exceeded configured limit")
var errUpstreamStreamTooLarge = errors.New("upstream stream exceeded configured limit")

func NewClient(cfg config.Config) *Client {
	config.Normalize(&cfg)
	timeoutSec := cfg.RequestTimeoutSec
	if timeoutSec <= 0 {
		timeoutSec = config.DefaultRequestTimeoutSec
	}
	var req Requester
	if cfg.Impersonate != "" && cfg.Proxy == "" {
		tlsAdapter, err := getTLSClient(cfg.Impersonate, timeoutSec)
		if err != nil {
			log.Printf("Failed to create TLS client for profile %s: %v, falling back to stdlib", cfg.Impersonate, err)
		} else {
			req = tlsAdapter
		}
	}

	if req == nil {
		transport := &http.Transport{
			DisableCompression:    true,
			ForceAttemptHTTP2:     true,
			ResponseHeaderTimeout: time.Duration(timeoutSec) * time.Second,
			IdleConnTimeout:       90 * time.Second,
			MaxIdleConns:          2000,
			MaxIdleConnsPerHost:   500,
			MaxConnsPerHost:       2000,
		}

		if cfg.Proxy != "" {
			if proxyURL, err := url.Parse(cfg.Proxy); err == nil {
				transport.Proxy = http.ProxyURL(proxyURL)
			}
		} else {
			transport.Proxy = http.ProxyFromEnvironment
		}

		req = &http.Client{
			Transport: transport,
			Timeout:   time.Duration(timeoutSec) * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}

	logFn := func(format string, args ...any) {
		if cfg.LogRequests {
			log.Printf(format, args...)
		}
	}

	pool := NewCookiePool()
	if len(cfg.CookiePool) > 0 {
		loaded := pool.LoadFromFiles(cfg.CookiePool)
		if loaded > 0 && cfg.LogRequests {
			log.Printf("[CookiePool] Loaded %d accounts from cookie pool configuration", loaded)
		}
	}
	if cfg.CookieFile != "" {
		pool.LoadFromFiles([]string{cfg.CookieFile})
	}

	return &Client{
		Cfg:     cfg,
		HTTP:    req,
		Cookies: NewCookieCache(cfg.CookieFile),
		Pool:    pool,
		Flight:  NewStreamFlight(),
		Logf:    logFn,
	}
}

func (c *Client) buildHeaders(session *AccountSession, guestCookies string) http.Header {
	authUser := c.Cfg.AuthUser
	if session != nil && session.AuthUser != "" {
		authUser = session.AuthUser
	}

	prefix := AccountPrefix(authUser)
	h := make(http.Header)
	h.Set("Content-Type", "application/x-www-form-urlencoded")
	h.Set("Origin", "https://gemini.google.com")
	h.Set("Referer", fmt.Sprintf("https://gemini.google.com%s/app", prefix))
	h.Set("X-Same-Domain", "1")

	// Dynamic Fingerprint matching (Mobile/iOS High-Trust Rotation)
	fp := ResolveFingerprint(c.Cfg.Impersonate)
	for k, v := range fp.Headers {
		h.Set(k, v)
	}

	if prefix != "" && authUser != "" {
		h.Set("X-Goog-AuthUser", authUser)
	}

	if session != nil && session.Cookie != "" {
		h.Set("Cookie", session.Cookie)
		if session.SAPISID != "" {
			hash := SAPISIDHash(session.SAPISID)
			if hash != "" {
				h.Set("Authorization", hash)
			}
		}
	} else if c.configuredSessionRoute() {
		// A configured session pool is an explicit authenticated route. If no
		// healthy account was selected, do not silently downgrade the request to
		// guest cookies. This also covers a configured cookie file that is
		// unreadable or malformed and therefore did not enter the pool.
	} else {
		cookieInfo, _ := c.Cookies.Load()
		if cookieInfo.Cookie != "" {
			h.Set("Cookie", cookieInfo.Cookie)
			if cookieInfo.SAPISID != "" {
				hash := SAPISIDHash(cookieInfo.SAPISID)
				if hash != "" {
					h.Set("Authorization", hash)
				}
			}
		} else if guestCookies != "" {
			h.Set("Cookie", guestCookies)
		} else if cookieInfo.GuestCookies != "" {
			h.Set("Cookie", cookieInfo.GuestCookies)
		}
	}

	if c.Cfg.Impersonate != "" {
		h.Set("Accept-Encoding", "identity")
	}

	return h
}

func (c *Client) doUpstream(req *http.Request) (*http.Response, error) {
	started := time.Now()
	if c.Metrics != nil {
		c.Metrics.UpstreamRequests.Add(1)
	}
	resp, err := c.HTTP.Do(req)
	if c.Metrics != nil {
		c.Metrics.UpstreamLatency.Observe(time.Since(started))
		if err != nil {
			c.Metrics.UpstreamErrors.Add(1)
		}
		if resp != nil && resp.StatusCode >= http.StatusBadRequest {
			c.Metrics.UpstreamErrors.Add(1)
			if resp.StatusCode == http.StatusTooManyRequests {
				c.Metrics.Upstream429.Add(1)
			}
		}
	}
	return resp, err
}

func (c *Client) triageStatus(resp *http.Response) error {
	retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
		return &UpstreamError{
			Status:     resp.StatusCode,
			Kind:       "http",
			Msg:        "upstream policy/network rejection (redirected to sorry/index); automatic retries stopped",
			RetryAfter: retryAfter,
		}
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return &UpstreamError{
			Status:     resp.StatusCode,
			Kind:       "http",
			Msg:        "upstream rate limited (HTTP 429); retry later",
			RetryAfter: retryAfter,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return &UpstreamError{
			Status:     resp.StatusCode,
			Kind:       "http",
			Msg:        fmt.Sprintf("upstream HTTP %d", resp.StatusCode),
			RetryAfter: retryAfter,
		}
	}
	return nil
}

// shouldRetryUpstream classifies failures where another attempt could have a
// reasonable chance of succeeding without amplifying a provider policy
// decision. The generation operation applies the stricter idempotency policy
// in canRetryGeneration: classification alone does not authorize replaying a
// POST after request delivery may have started. In particular, 3xx/401/403/429
// and Bard rejection frames must surface immediately; rotating identities or
// retrying harder is not a legitimate quota or access-control strategy.
func shouldRetryUpstream(err error) bool {
	var upstreamErr *UpstreamError
	if !errors.As(err, &upstreamErr) {
		return false
	}

	switch upstreamErr.Kind {
	case "transport":
		return true
	case "http":
		return upstreamErr.Status >= http.StatusInternalServerError
	default:
		// Parser/Bard rejection errors describe the provider response, not a
		// transient connection failure. Stream deduplication remains active
		// for transport retries only.
		return false
	}
}

// retryableBeforeRequest reports whether a transport error is known to have
// happened before a connection could be established. A POST to Google's
// generation endpoint has no idempotency contract: once a connection may have
// accepted request bytes, retrying can create a second generation even when
// the client never receives its response. Keep automatic retries limited to
// errors whose operation is still known not to have reached the provider.
func retryableBeforeRequest(err error) bool {
	if err == nil {
		return false
	}
	var opErr *net.OpError
	if !errors.As(err, &opErr) {
		return false
	}
	switch strings.ToLower(opErr.Op) {
	case "dial", "connect", "lookup":
		return true
	default:
		return false
	}
}

// canRetryGeneration applies the operation-level idempotency policy after
// shouldRetryUpstream has classified the error. HTTP 5xx responses and stream
// read failures remain ambiguous for a POST, so they are surfaced once rather
// than replayed. Only a pre-connection dial/lookup failure is safe to repeat.
func canRetryGeneration(err error, safeTransportFailure bool) bool {
	if !shouldRetryUpstream(err) {
		return false
	}
	var upstreamErr *UpstreamError
	if !errors.As(err, &upstreamErr) {
		return false
	}
	return upstreamErr.Kind == "transport" && safeTransportFailure
}

const maxUpstreamRetryBackoff = 60 * time.Second

func parseRetryAfter(value string) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds <= 0 {
			return 0
		}
		delay := time.Duration(seconds) * time.Second
		if delay < 0 || delay > maxUpstreamRetryBackoff {
			return maxUpstreamRetryBackoff
		}
		return delay
	}
	when, err := http.ParseTime(value)
	if err != nil {
		return 0
	}
	delay := time.Until(when)
	if delay <= 0 {
		return 0
	}
	if delay > maxUpstreamRetryBackoff {
		return maxUpstreamRetryBackoff
	}
	return delay
}

func upstreamRetryDelay(attempt, baseDelaySec int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	if baseDelaySec <= 0 {
		return 0
	}
	delay := time.Duration(baseDelaySec) * time.Second
	if delay <= 0 {
		return 0
	}
	for step := 0; step < attempt && delay < maxUpstreamRetryBackoff; step++ {
		if delay > maxUpstreamRetryBackoff/2 {
			delay = maxUpstreamRetryBackoff
			break
		}
		delay *= 2
	}
	if delay > maxUpstreamRetryBackoff {
		delay = maxUpstreamRetryBackoff
	}
	// Keep retries from synchronizing across a classroom while retaining a
	// deterministic upper bound. The package-level math/rand functions are
	// concurrency-safe; a zero configured delay remains entirely immediate for
	// hermetic tests and explicit low-latency configurations.
	jitterWindow := delay / 2
	if jitterWindow <= 0 {
		return delay
	}
	return delay/2 + time.Duration(rand.Int63n(int64(jitterWindow)+1))
}

func waitForUpstreamRetry(ctx context.Context, err error, attempt, baseDelaySec int) error {
	delay := upstreamRetryDelay(attempt, baseDelaySec)
	var upstreamErr *UpstreamError
	if errors.As(err, &upstreamErr) && upstreamErr.RetryAfter > 0 && upstreamErr.RetryAfter <= maxUpstreamRetryBackoff {
		delay = upstreamErr.RetryAfter
	}
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (c *Client) Generate(prompt string, modelID, thinkMode int, fileRefs []string, extra map[int]any) (string, error) {
	return c.GenerateContext(context.Background(), prompt, modelID, thinkMode, fileRefs, extra)
}

func (c *Client) GenerateContext(ctx context.Context, prompt string, modelID, thinkMode int, fileRefs []string, extra map[int]any) (string, error) {
	if c.Flight != nil {
		flightKey := c.requestFlightKey(prompt, modelID, thinkMode, fileRefs)
		return c.Flight.ExecuteContext(ctx, flightKey, func() (string, error) {
			return c.generateContextDirect(ctx, prompt, modelID, thinkMode, fileRefs, extra)
		})
	}
	return c.generateContextDirect(ctx, prompt, modelID, thinkMode, fileRefs, extra)
}

func (c *Client) generateContextDirect(ctx context.Context, prompt string, modelID, thinkMode int, fileRefs []string, extra map[int]any) (string, error) {
	atToken, blToken, guestCookies, _ := c.Cookies.GetSessionInfo(ctx, c.HTTP, c.Cfg.AuthUser)
	bodyStr := BuildBodyWithAt(prompt, modelID, thinkMode, fileRefs, extra, c.Cfg, atToken)
	reqURL := BuildURLWithBL(c.Cfg, blToken)

	var lastErr error
	for attempt := 0; attempt < c.Cfg.RetryAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		retrySafe := false
		session, sessionErr := c.healthySession()
		if sessionErr != nil {
			lastErr = sessionErr
			break
		}
		headers := c.buildHeaders(session, guestCookies)

		req, err := http.NewRequestWithContext(ctx, "POST", reqURL, strings.NewReader(bodyStr))
		if err != nil {
			return "", err
		}
		req.Header = headers.Clone()

		resp, err := c.doUpstream(req)
		if err != nil {
			if ctx.Err() != nil {
				return "", ctx.Err()
			}
			lastErr = &UpstreamError{Kind: "transport", Msg: err.Error()}
			retrySafe = retryableBeforeRequest(err)
			if session != nil {
				c.Pool.MarkFailure(session.ID)
			}
		} else {
			if resp == nil {
				lastErr = &UpstreamError{Kind: "protocol", Msg: "upstream returned an empty response"}
				if session != nil {
					c.Pool.MarkFailure(session.ID)
				}
			} else if err := c.triageStatus(resp); err != nil {
				closeResponseBody(resp)
				lastErr = err
				if session != nil {
					c.Pool.MarkFailure(session.ID)
				}
			} else {
				if resp.Body == nil {
					lastErr = &UpstreamError{Kind: "protocol", Msg: "upstream returned an empty response body"}
					if session != nil {
						c.Pool.MarkFailure(session.ID)
					}
				} else {
					rawBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxUpstreamResponseBytes+1))
					closeResponseBody(resp)
					if err != nil {
						lastErr = &UpstreamError{Kind: "transport", Msg: err.Error()}
						if session != nil {
							c.Pool.MarkFailure(session.ID)
						}
					} else if len(rawBytes) > maxUpstreamResponseBytes {
						lastErr = &UpstreamError{Kind: "protocol", Msg: fmt.Sprintf("upstream response exceeded %d bytes", maxUpstreamResponseBytes)}
						if session != nil {
							c.Pool.MarkFailure(session.ID)
						}
					} else {
						text, err := ExtractResponseText(string(rawBytes))
						if err != nil {
							lastErr = &UpstreamError{Kind: "bard", Msg: err.Error()}
							if session != nil {
								c.Pool.MarkFailure(session.ID)
							}
						} else {
							if session != nil {
								c.Pool.MarkSuccess(session.ID)
							}
							return text, nil
						}
					}
				}
			}
		}

		if attempt >= c.Cfg.RetryAttempts-1 || !canRetryGeneration(lastErr, retrySafe) {
			break
		}
		if attempt < c.Cfg.RetryAttempts-1 {
			c.Logf("Retry %d/%d: %v", attempt+1, c.Cfg.RetryAttempts, lastErr)
			if err := waitForUpstreamRetry(ctx, lastErr, attempt, c.Cfg.RetryDelaySec); err != nil {
				return "", err
			}
		}
	}

	return "", lastErr
}

func (c *Client) GenerateStream(prompt string, modelID, thinkMode int, fileRefs []string, extra map[int]any, emit func(string) error) error {
	return c.GenerateStreamContext(context.Background(), prompt, modelID, thinkMode, fileRefs, extra, emit)
}

func (c *Client) GenerateStreamContext(ctx context.Context, prompt string, modelID, thinkMode int, fileRefs []string, extra map[int]any, emit func(string) error) error {
	if c.Flight != nil {
		flightKey := c.requestFlightKey(prompt, modelID, thinkMode, fileRefs)
		return c.Flight.ExecuteStreamContext(ctx, flightKey, func(streamEmit func(string) error) error {
			return c.generateStreamContextDirect(ctx, prompt, modelID, thinkMode, fileRefs, extra, streamEmit)
		}, emit)
	}
	return c.generateStreamContextDirect(ctx, prompt, modelID, thinkMode, fileRefs, extra, emit)
}

// requestFlightKey keeps request coalescing limited to anonymous upstream
// work. A response produced with a browser session can vary by account,
// entitlement, experiment, or session state; sharing it across concurrent
// requests would be a correctness and privacy hazard. Anonymous requests can
// still use the existing deduplication path to reduce duplicate bursts.
func (c *Client) requestFlightKey(prompt string, modelID, thinkMode int, fileRefs []string) string {
	if c == nil || c.Flight == nil || c.sessionBound() {
		return ""
	}
	return c.Flight.KeyWithScope("anonymous", prompt, modelID, thinkMode, fileRefs)
}

func (c *Client) sessionBound() bool {
	if c == nil {
		return false
	}
	if strings.TrimSpace(c.Cfg.CookieFile) != "" || len(c.Cfg.CookiePool) > 0 {
		return true
	}
	return c.Pool != nil && c.Pool.Count() > 0
}

// configuredSessionRoute reports an explicit authenticated configuration even
// when its files failed validation and the pool is currently empty. Keeping
// this distinction prevents a bad/expired cookie from silently changing the
// provider route to anonymous guest traffic.
func (c *Client) configuredSessionRoute() bool {
	if c == nil {
		return false
	}
	return strings.TrimSpace(c.Cfg.CookieFile) != "" || len(c.Cfg.CookiePool) > 0 || (c.Pool != nil && c.Pool.Count() > 0)
}

func (c *Client) generateStreamContextDirect(ctx context.Context, prompt string, modelID, thinkMode int, fileRefs []string, extra map[int]any, emit func(string) error) error {
	atToken, blToken, guestCookies, _ := c.Cookies.GetSessionInfo(ctx, c.HTTP, c.Cfg.AuthUser)
	bodyStr := BuildBodyWithAt(prompt, modelID, thinkMode, fileRefs, extra, c.Cfg, atToken)
	reqURL := BuildURLWithBL(c.Cfg, blToken)

	parser := NewStreamParser()
	var lastErr error
	var streamBytes int64

	for attempt := 0; attempt < c.Cfg.RetryAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Reset the chunk buffer on each retry, but preserve the prevText state.
		// This ensures we don't emit duplicate deltas to the client if a stream connection drops halfway.
		parser.ResetBuffer()
		if attempt > 0 && c.Metrics != nil {
			c.Metrics.StreamRetries.Add(1)
			c.Metrics.SessionFailovers.Add(1)
		}

		retrySafe := false
		session, sessionErr := c.healthySession()
		if sessionErr != nil {
			lastErr = sessionErr
			break
		}
		headers := c.buildHeaders(session, guestCookies)

		req, err := http.NewRequestWithContext(ctx, "POST", reqURL, strings.NewReader(bodyStr))
		if err != nil {
			return err
		}
		req.Header = headers.Clone()

		resp, err := c.doUpstream(req)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			lastErr = &UpstreamError{Kind: "transport", Msg: err.Error()}
			retrySafe = retryableBeforeRequest(err)
			if session != nil {
				c.Pool.MarkFailure(session.ID)
			}
		} else {
			if resp == nil {
				lastErr = &UpstreamError{Kind: "protocol", Msg: "upstream returned an empty response"}
				if session != nil {
					c.Pool.MarkFailure(session.ID)
				}
			} else if err := c.triageStatus(resp); err != nil {
				closeResponseBody(resp)
				lastErr = err
				if session != nil {
					c.Pool.MarkFailure(session.ID)
				}
			} else {
				err = c.streamAttempt(resp.Body, parser, emit, &streamBytes)
				closeResponseBody(resp)
				if err == nil {
					if session != nil {
						c.Pool.MarkSuccess(session.ID)
					}
					return nil
				}
				if isClientDisconnect(err) || ctx.Err() != nil {
					return err
				}
				if session != nil {
					c.Pool.MarkFailure(session.ID)
				}
				lastErr = err
			}
		}

		if attempt >= c.Cfg.RetryAttempts-1 || !canRetryGeneration(lastErr, retrySafe) {
			break
		}
		if attempt < c.Cfg.RetryAttempts-1 {
			c.Logf("Stream retry %d/%d: %v", attempt+1, c.Cfg.RetryAttempts, lastErr)
			if err := waitForUpstreamRetry(ctx, lastErr, attempt, c.Cfg.RetryDelaySec); err != nil {
				return err
			}
		}
	}

	return lastErr
}

func (c *Client) streamAttempt(body io.Reader, parser *StreamParser, emit func(string) error, streamBytes *int64) error {
	if body == nil {
		return &UpstreamError{Kind: "protocol", Msg: "upstream returned an empty stream body"}
	}
	if parser == nil {
		return &UpstreamError{Kind: "protocol", Msg: "upstream stream parser is unavailable"}
	}
	if emit == nil {
		return &UpstreamError{Kind: "protocol", Msg: "upstream stream callback is unavailable"}
	}
	reader := bufio.NewReader(body)
	for {
		lineBytes, err := readBoundedStreamLine(reader)
		if streamBytes != nil {
			*streamBytes += int64(len(lineBytes))
			if *streamBytes > maxUpstreamStreamBytes {
				return &UpstreamError{Kind: "protocol", Msg: fmt.Sprintf("upstream stream exceeded %d bytes", maxUpstreamStreamBytes)}
			}
		}
		if len(lineBytes) > 0 {
			deltas, parseErr := parser.Feed(string(lineBytes))
			if parseErr != nil {
				if errors.Is(parseErr, errStreamTextTooLarge) {
					return &UpstreamError{Kind: "protocol", Msg: parseErr.Error()}
				}
				return &UpstreamError{Kind: "bard", Msg: parseErr.Error()}
			}
			for _, delta := range deltas {
				if err := emit(delta); err != nil {
					return err
				}
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				deltas, parseErr := parser.Flush()
				if parseErr != nil {
					if errors.Is(parseErr, errStreamTextTooLarge) {
						return &UpstreamError{Kind: "protocol", Msg: parseErr.Error()}
					}
					return &UpstreamError{Kind: "bard", Msg: parseErr.Error()}
				}
				for _, delta := range deltas {
					if emitErr := emit(delta); emitErr != nil {
						return emitErr
					}
				}
				return nil
			}
			if errors.Is(err, errUpstreamStreamLineTooLarge) || errors.Is(err, errUpstreamStreamTooLarge) {
				return &UpstreamError{Kind: "protocol", Msg: err.Error()}
			}
			return &UpstreamError{Kind: "transport", Msg: err.Error()}
		}
	}
}

func (c *Client) healthySession() (*AccountSession, error) {
	if c.Pool == nil || c.Pool.Count() == 0 {
		return nil, nil
	}
	session := c.Pool.GetHealthySession()
	if session == nil {
		return nil, &UpstreamError{
			Kind: "session",
			Msg:  "all configured Google sessions are unavailable or in cooldown; retry later",
		}
	}
	return session, nil
}

func closeResponseBody(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
}

func readBoundedStreamLine(reader *bufio.Reader) ([]byte, error) {
	var line []byte
	for {
		fragment, err := reader.ReadSlice('\n')
		line = append(line, fragment...)
		if len(line) > maxUpstreamStreamLineBytes {
			return nil, fmt.Errorf("%w: %d bytes", errUpstreamStreamLineTooLarge, maxUpstreamStreamLineBytes)
		}
		if err == bufio.ErrBufferFull {
			continue
		}
		return line, err
	}
}

// isClientDisconnect helps distinguish between upstream server errors (which we might want to retry)
// and actual client-side disconnections or non-retryable transport errors.
// If the client disconnected, we should abort the retry loop immediately to save resources.
func isClientDisconnect(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(*UpstreamError); ok {
		return false
	}
	return true
}
