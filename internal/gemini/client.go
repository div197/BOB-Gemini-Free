package gemini

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/metrics"
)

type UpstreamError struct {
	Status int
	Kind   string
	Msg    string
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

func NewClient(cfg config.Config) *Client {
	var req Requester
	if cfg.Impersonate != "" && cfg.Proxy == "" {
		tlsAdapter, err := getTLSClient(cfg.Impersonate, cfg.RequestTimeoutSec)
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
			ResponseHeaderTimeout: time.Duration(cfg.RequestTimeoutSec) * time.Second,
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
	} else if c.Pool != nil && c.Pool.Count() > 0 {
		// Configured pool session is in failure cooldown -> seamlessly fallback to live guest session!
		if guestCookies != "" {
			h.Set("Cookie", guestCookies)
		}
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
	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
		return &UpstreamError{
			Status: resp.StatusCode,
			Kind:   "http",
			Msg:    "upstream policy/network rejection (redirected to sorry/index); automatic retries stopped",
		}
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return &UpstreamError{
			Status: resp.StatusCode,
			Kind:   "http",
			Msg:    "upstream rate limited (HTTP 429); retry later",
		}
	}
	if resp.StatusCode != http.StatusOK {
		return &UpstreamError{
			Status: resp.StatusCode,
			Kind:   "http",
			Msg:    fmt.Sprintf("upstream HTTP %d", resp.StatusCode),
		}
	}
	return nil
}

// shouldRetryUpstream limits automatic retries to failures where another
// attempt has a reasonable chance of succeeding without amplifying a provider
// policy decision. In particular, 3xx/401/403/429 and Bard rejection frames
// must surface immediately; rotating identities or retrying harder is not a
// legitimate quota or access-control strategy.
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

func (c *Client) Generate(prompt string, modelID, thinkMode int, fileRefs []string, extra map[int]any) (string, error) {
	return c.GenerateContext(context.Background(), prompt, modelID, thinkMode, fileRefs, extra)
}

func (c *Client) GenerateContext(ctx context.Context, prompt string, modelID, thinkMode int, fileRefs []string, extra map[int]any) (string, error) {
	if c.Flight != nil {
		flightKey := c.Flight.Key(prompt, modelID, thinkMode, fileRefs)
		return c.Flight.Execute(flightKey, func() (string, error) {
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

		session := c.Pool.GetHealthySession()
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
			if session != nil {
				c.Pool.MarkFailure(session.ID)
			}
		} else {
			if err := c.triageStatus(resp); err != nil {
				_ = resp.Body.Close()
				lastErr = err
				if session != nil {
					c.Pool.MarkFailure(session.ID)
				}
			} else {
				rawBytes, err := io.ReadAll(resp.Body)
				_ = resp.Body.Close()
				if err != nil {
					lastErr = &UpstreamError{Kind: "transport", Msg: err.Error()}
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

		if attempt >= c.Cfg.RetryAttempts-1 || !shouldRetryUpstream(lastErr) {
			break
		}
		if attempt < c.Cfg.RetryAttempts-1 {
			c.Logf("Retry %d/%d: %v", attempt+1, c.Cfg.RetryAttempts, lastErr)
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(time.Duration(c.Cfg.RetryDelaySec) * time.Second):
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
		flightKey := c.Flight.Key(prompt, modelID, thinkMode, fileRefs)
		return c.Flight.ExecuteStream(flightKey, func(streamEmit func(string) error) error {
			return c.generateStreamContextDirect(ctx, prompt, modelID, thinkMode, fileRefs, extra, streamEmit)
		}, emit)
	}
	return c.generateStreamContextDirect(ctx, prompt, modelID, thinkMode, fileRefs, extra, emit)
}

func (c *Client) generateStreamContextDirect(ctx context.Context, prompt string, modelID, thinkMode int, fileRefs []string, extra map[int]any, emit func(string) error) error {
	atToken, blToken, guestCookies, _ := c.Cookies.GetSessionInfo(ctx, c.HTTP, c.Cfg.AuthUser)
	bodyStr := BuildBodyWithAt(prompt, modelID, thinkMode, fileRefs, extra, c.Cfg, atToken)
	reqURL := BuildURLWithBL(c.Cfg, blToken)

	parser := NewStreamParser()
	var lastErr error

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

		session := c.Pool.GetHealthySession()
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
			if session != nil {
				c.Pool.MarkFailure(session.ID)
			}
		} else {
			if err := c.triageStatus(resp); err != nil {
				_ = resp.Body.Close()
				lastErr = err
				if session != nil {
					c.Pool.MarkFailure(session.ID)
				}
			} else {
				err = c.streamAttempt(resp.Body, parser, emit)
				_ = resp.Body.Close()
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

		if attempt >= c.Cfg.RetryAttempts-1 || !shouldRetryUpstream(lastErr) {
			break
		}
		if attempt < c.Cfg.RetryAttempts-1 {
			c.Logf("Stream retry %d/%d: %v", attempt+1, c.Cfg.RetryAttempts, lastErr)
			timer := time.NewTimer(time.Duration(c.Cfg.RetryDelaySec) * time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
	}

	return lastErr
}

func (c *Client) streamAttempt(body io.Reader, parser *StreamParser, emit func(string) error) error {
	reader := bufio.NewReader(body)
	for {
		lineBytes, err := reader.ReadBytes('\n')
		if len(lineBytes) > 0 {
			deltas, parseErr := parser.Feed(string(lineBytes))
			if parseErr != nil {
				return &UpstreamError{Kind: "bard", Msg: parseErr.Error()}
			}
			for _, delta := range deltas {
				if err := emit(delta); err != nil {
					return err
				}
			}
		}
		if err != nil {
			if err == io.EOF {
				deltas, parseErr := parser.Flush()
				if parseErr != nil {
					return &UpstreamError{Kind: "bard", Msg: parseErr.Error()}
				}
				for _, delta := range deltas {
					if emitErr := emit(delta); emitErr != nil {
						return emitErr
					}
				}
				return nil
			}
			return &UpstreamError{Kind: "transport", Msg: err.Error()}
		}
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
