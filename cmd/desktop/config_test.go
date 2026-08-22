package main

import (
	"path/filepath"
	"testing"
)

func TestLoadDesktopConfigKeepsUserSessionAndLoopbackBoundary(t *testing.T) {
	cookieFile := filepath.Join(t.TempDir(), "cookie.txt")
	t.Setenv("BOB_GEMINI_FREE_COOKIE_FILE", cookieFile)
	t.Setenv("BOB_GEMINI_FREE_HOST", "0.0.0.0")
	t.Setenv("BOB_GEMINI_FREE_API_KEYS", "desktop-secret")
	t.Setenv("BOB_GEMINI_FREE_ALLOWED_ORIGINS", "https://example.invalid")

	cfg := loadDesktopConfig()

	if cfg.CookieFile != cookieFile {
		t.Fatalf("desktop cookie file = %q, want %q", cfg.CookieFile, cookieFile)
	}
	if cfg.Host != "127.0.0.1" {
		t.Fatalf("desktop host = %q, want loopback", cfg.Host)
	}
	if len(cfg.APIKeys) != 0 {
		t.Fatalf("desktop API keys = %v, want disabled", cfg.APIKeys)
	}
	if len(cfg.AllowedOrigins) != 0 {
		t.Fatalf("desktop allowed origins = %v, want disabled", cfg.AllowedOrigins)
	}
}
