package multimodal

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/gemini"
)

const (
	DefaultPushID             = "feeds/mcudyrk2a4khkz"
	DefaultPctx               = "CgcSBWjK7pYx"
	TokenCacheTTL             = 600 * time.Second
	TokenCacheRetryDelay      = 15 * time.Second
	MaxPageTokenResponseBytes = 8 << 20
	MaxPageTokenValueBytes    = 4096
)

// Internal regex patterns used to extract hidden session tokens embedded in Gemini HTML responses:
// - `qKIAYe`: Push-ID used as tenant identifier for image uploads
// - `Ylro7b`: Client-Pctx (context token) required for image upload authorization
// - `SNlM0e` / `thykhd`: XSRF/AT token required for state-changing POST requests
// - `cfb2h`: Build label version required for RPC stream endpoints
var (
	rePushID = regexp.MustCompile(`"(?:qKIAYe|KnDnFf)":"([^"]+)"`)
	rePctx   = regexp.MustCompile(`"Ylro7b":"([^"]+)"`)
	reAt     = regexp.MustCompile(`"(?:SNlM0e|thykhd)":"([^"]+)"`)
	reBL     = regexp.MustCompile(`"cfb2h":"([^"]+)"`)
)

type PageTokens struct {
	PushID string
	Pctx   string
	At     string
	BL     string
}

type TokenCache struct {
	mu     sync.Mutex
	cfg    config.Config
	cookie *gemini.CookieCache
	client *http.Client
	ts     time.Time
	tokens PageTokens
	// retryAt prevents a failed refresh from becoming either a ten-minute
	// outage or an immediate retry storm. refreshing keeps the timestamp
	// reservation from creating a dog-pile while the request is in flight.
	retryAt    time.Time
	refreshing bool
	nowFn      func() time.Time
}

func NewTokenCache(cfg config.Config, cookie *gemini.CookieCache, client *http.Client) *TokenCache {
	return &TokenCache{
		cfg:    cfg,
		cookie: cookie,
		client: client,
		tokens: PageTokens{
			PushID: DefaultPushID,
			Pctx:   DefaultPctx,
			At:     "",
			BL:     cfg.GeminiBL,
		},
	}
}

func (c *TokenCache) fetchPageTokens(ctx context.Context) (PageTokens, bool) {
	if c == nil {
		return PageTokens{PushID: DefaultPushID, Pctx: DefaultPctx}, false
	}
	tokens := PageTokens{
		PushID: DefaultPushID,
		Pctx:   DefaultPctx,
		At:     "",
		BL:     c.cfg.GeminiBL,
	}
	if c.client == nil {
		return tokens, false
	}
	if ctx == nil {
		ctx = context.Background()
	}

	reqURL := fmt.Sprintf("https://gemini.google.com%s/app", gemini.AccountPrefix(c.cfg.AuthUser))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return tokens, false
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	var cookieInfo gemini.CookieInfo
	if c.cookie != nil {
		cookieInfo, _ = c.cookie.Load()
	}
	if cookieInfo.Cookie != "" {
		req.Header.Set("Cookie", cookieInfo.Cookie)
	}
	if cookieInfo.SAPISID != "" {
		req.Header.Set("Authorization", gemini.SAPISIDHash(cookieInfo.SAPISID))
	}

	// Do not allow a page-token request carrying session credentials to follow
	// a redirect to an untrusted or unexpected host. The shared application
	// client is cloned so image-fetch behavior and caller configuration remain
	// unchanged.
	client := *c.client
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := client.Do(req)
	if err != nil {
		// The request carries the user's session authorization; keep transport
		// details and URLs out of logs.
		log.Printf("Page token fetch failed")
		return tokens, false
	}
	if resp == nil || resp.Body == nil {
		return tokens, false
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return tokens, false
	}
	if resp.ContentLength > MaxPageTokenResponseBytes {
		return tokens, false
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, MaxPageTokenResponseBytes+1))
	if err != nil || len(bodyBytes) > MaxPageTokenResponseBytes {
		return tokens, false
	}

	html := string(bodyBytes)
	found := false
	if m := rePushID.FindStringSubmatch(html); len(m) > 1 {
		if tokenValueUsable(m[1]) {
			tokens.PushID = m[1]
			found = true
		}
	}
	if m := rePctx.FindStringSubmatch(html); len(m) > 1 {
		if tokenValueUsable(m[1]) {
			tokens.Pctx = m[1]
			found = true
		}
	}
	if m := reAt.FindStringSubmatch(html); len(m) > 1 {
		if tokenValueUsable(m[1]) {
			tokens.At = m[1]
			found = true
		}
	}
	if m := reBL.FindStringSubmatch(html); len(m) > 1 {
		if tokenValueUsable(m[1]) {
			tokens.BL = m[1]
			found = true
		}
	}

	return tokens, found
}

func (c *TokenCache) Get() PageTokens {
	return c.GetContext(context.Background())
}

func (c *TokenCache) now() time.Time {
	if c != nil && c.nowFn != nil {
		return c.nowFn()
	}
	return time.Now()
}

// GetContext refreshes page tokens without allowing a failed refresh to erase
// the last known-good token set. This matters for Scotty uploads: a transient
// login-page, redirect, or oversized response should fail the current upload,
// but should not turn every subsequent request into a request with empty
// authorization state. A failed refresh is retried after a short bounded
// delay, rather than being cached as a ten-minute success.
func (c *TokenCache) GetContext(ctx context.Context) PageTokens {
	if c == nil {
		return PageTokens{PushID: DefaultPushID, Pctx: DefaultPctx}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	now := c.now()
	c.mu.Lock()
	refreshDue := c.ts.IsZero() || now.Sub(c.ts) > TokenCacheTTL || (!c.retryAt.IsZero() && !now.Before(c.retryAt))
	if !c.refreshing && refreshDue {
		// Immediately update the timestamp to prevent a dog-pile of concurrent fetches.
		// Other requests will safely fall back to the stale tokens while we fetch in the background.
		c.ts = now
		c.refreshing = true
		c.mu.Unlock()

		newTokens, ok := c.fetchPageTokens(ctx)

		c.mu.Lock()
		c.refreshing = false
		if ok {
			c.tokens = newTokens
			c.retryAt = time.Time{}
		} else {
			c.retryAt = now.Add(TokenCacheRetryDelay)
		}
		tokens := c.tokens
		c.mu.Unlock()
		return tokens
	}

	t := c.tokens
	c.mu.Unlock()
	return t
}

func tokenValueUsable(value string) bool {
	return value != "" && len(value) <= MaxPageTokenValueBytes && strings.IndexFunc(value, func(r rune) bool {
		return r < 0x20 || r == 0x7f
	}) == -1
}
