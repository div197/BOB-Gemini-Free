package gemini

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	reAt = regexp.MustCompile(`"(?:SNlM0e|thykhd)":"([^"]+)"`)
	reBL = regexp.MustCompile(`"cfb2h":"([^"]+)"`)
)

type CookieInfo struct {
	Cookie       string
	GuestCookies string
	SAPISID      string
	At           string
	BL           string
	AtTime       time.Time
}

type CookieCache struct {
	mu         sync.Mutex
	file       string
	mtime      time.Time
	info       CookieInfo
	refreshing bool
	refreshCh  chan struct{}
}

func NewCookieCache(file string) *CookieCache {
	return &CookieCache{file: file}
}

func (c *CookieCache) loadUnlocked() error {
	if c.file == "" {
		return nil
	}
	stat, err := os.Stat(c.file)
	if err != nil {
		if c.info.Cookie != "" {
			return nil
		}
		return err
	}

	if stat.ModTime().Equal(c.mtime) && c.info.Cookie != "" {
		return nil
	}

	contentBytes, err := os.ReadFile(c.file)
	if err != nil {
		return err
	}

	c.mtime = stat.ModTime()
	content := strings.TrimSpace(string(contentBytes))

	c.info.Cookie = ""
	c.info.SAPISID = ""

	if strings.HasPrefix(content, "{") {
		var jsonCookies map[string]any
		if err := json.Unmarshal(contentBytes, &jsonCookies); err == nil {
			if cookieVal, ok := jsonCookies["cookie"].(string); ok {
				c.info.Cookie = cookieVal
			}
			if sapisidVal, ok := jsonCookies["sapisid"].(string); ok {
				c.info.SAPISID = sapisidVal
			}
		}
	} else {
		extracted, err := ExtractCookies(content)
		if err == nil && extracted != nil {
			c.info.Cookie = extracted.RawCookie
			c.info.SAPISID = extracted.SAPISID
		} else {
			c.info.Cookie = content
			for _, part := range strings.Split(content, ";") {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "SAPISID=") {
					c.info.SAPISID = strings.TrimPrefix(part, "SAPISID=")
					break
				}
			}
		}
	}

	return nil
}

// GetSessionInfo dynamically resolves active SNlM0e `at` token, `bl` build, and cookies
// for both authenticated user sessions and anonymous/guest school lab sessions.
// It includes Stale-While-Revalidate and Thundering-Herd protection for 500+ concurrent students.
func (c *CookieCache) GetSessionInfo(ctx context.Context, httpClient Requester, authUser string) (atToken string, blToken string, cookieHeader string, sapisid string) {
	c.mu.Lock()
	_ = c.loadUnlocked()

	hasCookie := c.info.Cookie != ""
	if time.Since(c.info.AtTime) < 10*time.Minute && c.info.At != "" {
		at := c.info.At
		bl := c.info.BL
		cookie := c.info.Cookie
		if cookie == "" {
			cookie = c.info.GuestCookies
		}
		sapi := c.info.SAPISID
		c.mu.Unlock()
		return at, bl, cookie, sapi
	}

	// If another goroutine is already refreshing /app:
	if c.refreshing {
		// If we already have a previous token, serve it stale immediately (zero blocking!)
		if c.info.At != "" {
			at := c.info.At
			bl := c.info.BL
			cookie := c.info.Cookie
			if cookie == "" {
				cookie = c.info.GuestCookies
			}
			sapi := c.info.SAPISID
			c.mu.Unlock()
			return at, bl, cookie, sapi
		}

		// Initial startup with no token yet: wait for the active refresh to finish
		waitCh := c.refreshCh
		c.mu.Unlock()
		if waitCh != nil {
			select {
			case <-waitCh:
			case <-ctx.Done():
				return "", "", "", ""
			}
		}
		c.mu.Lock()
		at := c.info.At
		bl := c.info.BL
		cookie := c.info.Cookie
		if cookie == "" {
			cookie = c.info.GuestCookies
		}
		sapi := c.info.SAPISID
		c.mu.Unlock()
		return at, bl, cookie, sapi
	}

	// Designated refresher goroutine
	c.refreshing = true
	c.refreshCh = make(chan struct{})
	cookieStr := c.info.Cookie
	sapi := c.info.SAPISID
	lastAt := c.info.At
	lastBL := c.info.BL
	lastGuest := c.info.GuestCookies
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.refreshing = false
		if c.refreshCh != nil {
			close(c.refreshCh)
			c.refreshCh = nil
		}
		c.mu.Unlock()
	}()

	if httpClient == nil {
		return lastAt, lastBL, cookieStr, sapi
	}

	reqURL := fmt.Sprintf("https://gemini.google.com%s/app", AccountPrefix(authUser))
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return lastAt, lastBL, cookieStr, sapi
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")

	if hasCookie {
		req.Header.Set("Cookie", cookieStr)
		if sapi != "" {
			req.Header.Set("Authorization", SAPISIDHash(sapi))
		}
	} else if lastGuest != "" {
		req.Header.Set("Cookie", lastGuest)
	}

	resp, err := httpClient.Do(req)
	if err != nil || resp == nil {
		return lastAt, lastBL, cookieStr, sapi
	}
	defer resp.Body.Close()

	// Extract Set-Cookie headers for guest session
	var extractedGuestCookies []string
	if resp.Header != nil {
		for _, rawCookie := range resp.Header.Values("Set-Cookie") {
			parts := strings.SplitN(rawCookie, ";", 2)
			if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
				extractedGuestCookies = append(extractedGuestCookies, strings.TrimSpace(parts[0]))
			}
		}
	}
	newGuestCookies := strings.Join(extractedGuestCookies, "; ")

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return lastAt, lastBL, cookieStr, sapi
	}
	html := string(bodyBytes)

	c.mu.Lock()
	defer c.mu.Unlock()
	if m := reAt.FindStringSubmatch(html); len(m) > 1 {
		c.info.At = m[1]
		c.info.AtTime = time.Now()
	}
	if m := reBL.FindStringSubmatch(html); len(m) > 1 {
		c.info.BL = m[1]
	}
	if !hasCookie && newGuestCookies != "" {
		c.info.GuestCookies = newGuestCookies
	}

	finalCookie := c.info.Cookie
	if finalCookie == "" {
		finalCookie = c.info.GuestCookies
	}
	return c.info.At, c.info.BL, finalCookie, c.info.SAPISID
}

func (c *CookieCache) GetAtToken(ctx context.Context, httpClient Requester, authUser string) string {
	at, _, _, _ := c.GetSessionInfo(ctx, httpClient, authUser)
	return at
}

func (c *CookieCache) Load() (CookieInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.loadUnlocked()
	return c.info, err
}

func SAPISIDHash(sapisid string) string {
	return SAPISIDHashAt(sapisid, time.Now().Unix())
}

// SAPISIDHashAt computes the Google SAPISIDHASH authorization value for a
// caller-supplied Unix timestamp. Keeping the timestamped primitive separate
// makes the protocol invariant deterministic in tests while preserving the
// existing wall-clock behavior for production requests.
func SAPISIDHashAt(sapisid string, timestamp int64) string {
	if sapisid == "" {
		return ""
	}
	input := fmt.Sprintf("%d %s https://gemini.google.com", timestamp, sapisid)
	h := sha1.Sum([]byte(input))
	return fmt.Sprintf("SAPISIDHASH %d_%s", timestamp, hex.EncodeToString(h[:]))
}

func AccountPrefix(authUser string) string {
	if authUser == "" {
		return ""
	}
	return "/u/" + authUser
}

// ExtractedCookie holds verified Google session cookie tokens.
type ExtractedCookie struct {
	RawCookie string
	SAPISID   string
	Tokens    map[string]string
}

// ExtractCookies parses raw cookie strings, JSON, or DevTools headers,
// extracting critical Google session tokens (SID, HSID, SSID, APISID, SAPISID, __Secure-1PSID).
func ExtractCookies(raw string) (*ExtractedCookie, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("cookie input is empty")
	}

	tokens := make(map[string]string)
	var cookieStr string

	if strings.HasPrefix(raw, "{") {
		var data struct {
			Cookie  string `json:"cookie"`
			SAPISID string `json:"sapisid"`
		}
		if err := json.Unmarshal([]byte(raw), &data); err == nil && data.Cookie != "" {
			cookieStr = data.Cookie
		}
	}
	if cookieStr == "" {
		cookieStr = raw
	}

	// Remove leading "Cookie: " if user copied header prefix
	cookieStr = strings.TrimPrefix(cookieStr, "Cookie: ")
	cookieStr = strings.TrimPrefix(cookieStr, "cookie: ")

	// Split by ';' or newline
	items := strings.FieldsFunc(cookieStr, func(r rune) bool {
		return r == ';' || r == '\n' || r == '\r'
	})

	var formattedPairs []string
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		parts := strings.SplitN(item, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			if k != "" && v != "" {
				tokens[k] = v
				formattedPairs = append(formattedPairs, fmt.Sprintf("%s=%s", k, v))
			}
		}
	}

	if len(tokens) == 0 {
		return nil, fmt.Errorf("no valid key=value cookie pairs found")
	}

	sapisid := tokens["SAPISID"]
	if sapisid == "" {
		sapisid = tokens["__Secure-3PAPISID"]
	}

	finalCookie := strings.Join(formattedPairs, "; ")

	return &ExtractedCookie{
		RawCookie: finalCookie,
		SAPISID:   sapisid,
		Tokens:    tokens,
	}, nil
}

// SaveCookieFile writes cookies to disk with restricted 0600 POSIX permissions.
func SaveCookieFile(filePath string, cookieStr string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return os.WriteFile(filePath, []byte(cookieStr), 0600)
}
