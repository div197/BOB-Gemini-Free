package gemini

import (
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
	Cookie  string
	SAPISID string
	At      string
	BL      string
	AtTime  time.Time
}

type CookieCache struct {
	mu    sync.Mutex
	file  string
	mtime time.Time
	info  CookieInfo
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
		if c.info.Cookie != "" {
			return nil
		}
		return err
	}

	content := strings.TrimSpace(string(contentBytes))
	var cookieStr string
	var sapisid string

	if strings.HasPrefix(content, "{") {
		var data struct {
			Cookie  string `json:"cookie"`
			SAPISID string `json:"sapisid"`
		}
		if err := json.Unmarshal([]byte(content), &data); err == nil {
			cookieStr = data.Cookie
			sapisid = data.SAPISID
		}
	} else {
		cookieStr = content
		pairs := strings.Split(cookieStr, "; ")
		for _, pair := range pairs {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 && parts[0] == "SAPISID" {
				sapisid = parts[1]
				break
			}
		}
	}

	c.mtime = stat.ModTime()
	c.info.Cookie = cookieStr
	c.info.SAPISID = sapisid

	return nil
}

func (c *CookieCache) GetAtToken(httpClient Requester, authUser string) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	_ = c.loadUnlocked()

	if c.info.Cookie == "" {
		return ""
	}

	if time.Since(c.info.AtTime) < 10*time.Minute && c.info.At != "" {
		return c.info.At
	}

	reqURL := fmt.Sprintf("https://gemini.google.com%s/app", AccountPrefix(authUser))
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return c.info.At
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Cookie", c.info.Cookie)
	if c.info.SAPISID != "" {
		req.Header.Set("Authorization", SAPISIDHash(c.info.SAPISID))
	}

	resp, err := httpClient.Do(req)
	if err != nil || resp == nil {
		return c.info.At
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.info.At
	}
	html := string(bodyBytes)

	if m := reAt.FindStringSubmatch(html); len(m) > 1 {
		c.info.At = m[1]
		c.info.AtTime = time.Now()
	}
	if m := reBL.FindStringSubmatch(html); len(m) > 1 {
		c.info.BL = m[1]
	}

	return c.info.At
}

func (c *CookieCache) Load() (CookieInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.loadUnlocked()
	return c.info, err
}

func SAPISIDHash(sapisid string) string {
	if sapisid == "" {
		return ""
	}
	ts := time.Now().Unix()
	input := fmt.Sprintf("%d %s https://gemini.google.com", ts, sapisid)
	h := sha1.Sum([]byte(input))
	return fmt.Sprintf("SAPISIDHASH %d_%s", ts, hex.EncodeToString(h[:]))
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
