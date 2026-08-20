with open('internal/gemini/client.go', 'r') as f:
    code = f.read()

code = code.replace("atToken := c.Cookies.GetAtToken(c.HTTP, c.Cfg.AuthUser)", "atToken := c.Cookies.GetAtToken(ctx, c.HTTP, c.Cfg.AuthUser)")

# Fix the time.After leak
code = code.replace("""		if attempt < c.Cfg.RetryAttempts-1 {
			c.Logf("Stream retry %d/%d: %v", attempt+1, c.Cfg.RetryAttempts, lastErr)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(c.Cfg.RetryDelaySec) * time.Second):
			}
		}""", """		if attempt < c.Cfg.RetryAttempts-1 {
			c.Logf("Stream retry %d/%d: %v", attempt+1, c.Cfg.RetryAttempts, lastErr)
			timer := time.NewTimer(time.Duration(c.Cfg.RetryDelaySec) * time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}""")

with open('internal/gemini/client.go', 'w') as f:
    f.write(code)
