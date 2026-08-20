import re

with open('internal/gemini/auth.go', 'r') as f:
    code = f.read()

bad_get = """func (c *CookieCache) GetAtToken(httpClient Requester, authUser string) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	_ = c.loadUnlocked()

	if c.info.Cookie == "" {
		return ""
	}

	if time.Since(c.info.AtTime) < 10*time.Minute && c.info.At != "" {
		return c.info.At
	}

	reqURL := "https://gemini.google.com/app"
	if authUser != "" && authUser != "0" {
		reqURL = "https://gemini.google.com/u/" + authUser + "/app"
	}
	req, err := http.NewRequest("GET", reqURL, nil)"""

good_get = """func (c *CookieCache) GetAtToken(ctx context.Context, httpClient Requester, authUser string) string {
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
	
	cookie := c.info.Cookie
	c.mu.Unlock() // Release lock during network I/O to prevent massive contention

	reqURL := "https://gemini.google.com/app"
	if authUser != "" && authUser != "0" {
		reqURL = "https://gemini.google.com/u/" + authUser + "/app"
	}
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)"""

code = code.replace(bad_get, good_get)
code = code.replace("req.Header.Set(\"Cookie\", c.info.Cookie)", "req.Header.Set(\"Cookie\", cookie)")

# Also lock again before saving
bad_save = """	c.info.At = at
	c.info.AtTime = time.Now()
	_ = c.saveUnlocked()

	return at
}"""
good_save = """	c.mu.Lock()
	c.info.At = at
	c.info.AtTime = time.Now()
	_ = c.saveUnlocked()
	c.mu.Unlock()

	return at
}"""
code = code.replace(bad_save, good_save)

with open('internal/gemini/auth.go', 'w') as f:
    f.write(code)
