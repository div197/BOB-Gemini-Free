package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDesktopConfigKeepsUserSessionAndLoopbackBoundary(t *testing.T) {
	cookieFile := filepath.Join(t.TempDir(), "cookie.txt")
	t.Setenv("BOB_GEMINI_FREE_COOKIE_FILE", cookieFile)
	t.Setenv("BOB_GEMINI_FREE_HOST", "0.0.0.0")
	t.Setenv("BOB_GEMINI_FREE_API_KEYS", "desktop-secret")
	t.Setenv("BOB_GEMINI_FREE_ALLOWED_ORIGINS", "https://example.invalid")

	cfg, err := loadDesktopConfig()
	if err != nil {
		t.Fatalf("loadDesktopConfig: %v", err)
	}

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

func TestLoadDesktopConfigFailsClosedOnInvalidDiscoveredConfig(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	path := filepath.Join(root, "config.json")
	if err := os.WriteFile(path, []byte("{invalid"), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if _, err := loadDesktopConfig(); err == nil {
		t.Fatal("invalid discovered config was accepted")
	}
}
