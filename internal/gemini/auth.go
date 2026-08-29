package gemini

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	reAt = regexp.MustCompile(`"(?:SNlM0e|thykhd)":"([^"]+)"`)
	reBL = regexp.MustCompile(`"cfb2h":"([^"]+)"`)
)

const (
	// Cookie files and the Gemini bootstrap document are untrusted local or
	// network inputs. Keep their limits well above normal values without
	// allowing an accidental large file/page to consume unbounded memory.
	MaxCookieFileBytes  int64 = 1 << 20
	maxSessionHTMLBytes int64 = 4 << 20
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
	mu             sync.Mutex
	file           string
	mtime          time.Time
	fileSize       int64
	contentHash    [32]byte
	hasContentHash bool
	info           CookieInfo
	refreshing     bool
	refreshCh      chan struct{}
}

func NewCookieCache(file string) *CookieCache {
	return &CookieCache{file: file}
}

func (c *CookieCache) loadUnlocked() error {
	if c.file == "" {
		return nil
	}
	stat, contentBytes, err := readSecureCookieFile(c.file)
	if err != nil {
		// Do not continue sending a deleted or unreadable credential file. Keep
		// guest cookies separate so anonymous access can still be retried.
		c.info.Cookie = ""
		c.info.SAPISID = ""
		c.info.At = ""
		c.info.BL = ""
		c.info.AtTime = time.Time{}
		c.fileSize = 0
		c.contentHash = [32]byte{}
		c.hasContentHash = false
		return err
	}
	contentHash := sha256.Sum256(contentBytes)
	if stat.ModTime().Equal(c.mtime) && stat.Size() == c.fileSize && c.hasContentHash && contentHash == c.contentHash {
		return nil
	}

	c.mtime = stat.ModTime()
	c.fileSize = stat.Size()
	c.contentHash = contentHash
	c.hasContentHash = true
	c.info.Cookie = ""
	c.info.SAPISID = ""
	c.info.At = ""
	c.info.BL = ""
	c.info.AtTime = time.Time{}

	extracted, err := ExtractCookies(strings.TrimSpace(string(contentBytes)))
	if err != nil {
		return fmt.Errorf("parse cookie file %s: %w", c.file, err)
	}
	c.info.Cookie = extracted.RawCookie
	c.info.SAPISID = extracted.SAPISID
	return nil
}

// GetSessionInfo dynamically resolves active SNlM0e `at` token, `bl` build, and cookies
// for both authenticated user sessions and anonymous/guest school lab sessions.
// It includes Stale-While-Revalidate and Thundering-Herd protection for 500+ concurrent students.
func (c *CookieCache) GetSessionInfo(ctx context.Context, httpClient Requester, authUser string) (atToken string, blToken string, cookieHeader string, sapisid string) {
	if ctx == nil {
		ctx = context.Background()
	}
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
	if resp.Body == nil || resp.StatusCode != http.StatusOK || resp.ContentLength > maxSessionHTMLBytes {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
		return lastAt, lastBL, cookieStr, sapi
	}

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

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxSessionHTMLBytes+1))
	_ = resp.Body.Close()
	if err != nil {
		return lastAt, lastBL, cookieStr, sapi
	}
	if int64(len(bodyBytes)) > maxSessionHTMLBytes {
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
	if strings.TrimSpace(filePath) == "" {
		return fmt.Errorf("cookie file path is empty")
	}
	if int64(len(cookieStr)) > MaxCookieFileBytes {
		return fmt.Errorf("cookie input exceeds %d bytes", MaxCookieFileBytes)
	}
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(dir, 0700); err != nil {
			return fmt.Errorf("failed to restrict cookie directory: %w", err)
		}
	}
	if err := os.WriteFile(filePath, []byte(cookieStr), 0600); err != nil {
		return err
	}
	if err := os.Chmod(filePath, 0600); err != nil {
		return fmt.Errorf("failed to restrict cookie file: %w", err)
	}
	return nil
}

func readSecureCookieFile(filePath string) (os.FileInfo, []byte, error) {
	stat, err := os.Lstat(filePath)
	if err != nil {
		return nil, nil, err
	}
	if stat.Mode()&os.ModeSymlink != 0 {
		return nil, nil, fmt.Errorf("cookie file is a symlink: %s", filePath)
	}
	if !stat.Mode().IsRegular() {
		return nil, nil, fmt.Errorf("cookie file is not a regular file: %s", filePath)
	}
	if stat.Size() > MaxCookieFileBytes {
		return nil, nil, fmt.Errorf("cookie file exceeds %d bytes", MaxCookieFileBytes)
	}
	if runtime.GOOS != "windows" && stat.Mode().Perm()&0077 != 0 {
		return nil, nil, fmt.Errorf("cookie file permissions are too broad: %o", stat.Mode().Perm())
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}
	if int64(len(content)) > MaxCookieFileBytes {
		return nil, nil, fmt.Errorf("cookie file exceeds %d bytes", MaxCookieFileBytes)
	}
	return stat, content, nil
}
