import re

with open('internal/gemini/auth.go', 'r') as f:
    code = f.read()
    
# Make sure context is imported
if '"context"' not in code:
    code = code.replace('"crypto/sha1"', '"context"\n\t"crypto/sha1"')

bad_func = """func (c *CookieCache) GetAtToken(httpClient Requester, authUser string) string {
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
	req.Header.Set("Cookie", cookie)
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
}"""

good_func = """func (c *CookieCache) GetAtToken(ctx context.Context, httpClient Requester, authUser string) string {
	c.mu.Lock()
	_ = c.loadUnlocked()

	if c.info.Cookie == "" {
		c.mu.Unlock()
		return ""
	}

	if time.Since(c.info.AtTime) < 10*time.Minute && c.info.At != "" {
		at := c.info.At
		c.mu.Unlock()
		return at
	}

	cookieStr := c.info.Cookie
	sapi := c.info.SAPISID
	lastAt := c.info.At
	c.mu.Unlock() // Release lock for network request!

	reqURL := fmt.Sprintf("https://gemini.google.com%s/app", AccountPrefix(authUser))
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return lastAt
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Cookie", cookieStr)
	if sapi != "" {
		req.Header.Set("Authorization", SAPISIDHash(sapi))
	}

	resp, err := httpClient.Do(req)
	if err != nil || resp == nil {
		return lastAt
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return lastAt
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

	return c.info.At
}"""

# Fallback regex sub in case exact text doesn't match
pattern = re.compile(r'func \(c \*CookieCache\) GetAtToken[\s\S]*?return c\.info\.At\n}')
code = pattern.sub(good_func, code)

with open('internal/gemini/auth.go', 'w') as f:
    f.write(code)
