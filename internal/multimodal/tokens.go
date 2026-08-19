package multimodal

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/gemini"
)

const (
	DefaultPushID = "feeds/mcudyrk2a4khkz"
	DefaultPctx   = "CgcSBWjK7pYx"
	TokenCacheTTL = 600 * time.Second
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

func (c *TokenCache) fetchPageTokens() PageTokens {
	tokens := PageTokens{
		PushID: DefaultPushID,
		Pctx:   DefaultPctx,
		At:     "",
		BL:     c.cfg.GeminiBL,
	}

	reqURL := fmt.Sprintf("https://gemini.google.com%s/app", gemini.AccountPrefix(c.cfg.AuthUser))
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return tokens
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	cookieInfo, _ := c.cookie.Load()
	if cookieInfo.Cookie != "" {
		req.Header.Set("Cookie", cookieInfo.Cookie)
	}
	if cookieInfo.SAPISID != "" {
		req.Header.Set("Authorization", gemini.SAPISIDHash(cookieInfo.SAPISID))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("Page token fetch failed: %v", err)
		return tokens
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return tokens
	}

	html := string(bodyBytes)
	if m := rePushID.FindStringSubmatch(html); len(m) > 1 {
		tokens.PushID = m[1]
	}
	if m := rePctx.FindStringSubmatch(html); len(m) > 1 {
		tokens.Pctx = m[1]
	}
	if m := reAt.FindStringSubmatch(html); len(m) > 1 {
		tokens.At = m[1]
	}
	if m := reBL.FindStringSubmatch(html); len(m) > 1 {
		tokens.BL = m[1]
	}

	return tokens
}

func (c *TokenCache) Get() PageTokens {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Since(c.ts) > TokenCacheTTL {
		c.tokens = c.fetchPageTokens()
		c.ts = time.Now()
	}

	return c.tokens
}
