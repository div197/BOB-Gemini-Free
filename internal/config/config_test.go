package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error loading empty path: %v", err)
	}
	if cfg.Port != 8081 {
		t.Errorf("expected port 8081, got %d", cfg.Port)
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", cfg.Host)
	}
	if cfg.DefaultModel != "gemini-3.6-flash" {
		t.Errorf("expected default model gemini-3.6-flash, got %s", cfg.DefaultModel)
	}
	if cfg.RequestTimeoutSec != 180 {
		t.Errorf("expected timeout 180, got %d", cfg.RequestTimeoutSec)
	}
}

func TestLoadConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	content := `{
		"port": 9090,
		"host": "0.0.0.0",
		"default_model": "gemini-3.5-flash-thinking",
		"api_keys": ["sk-test-key-1", "sk-test-key-2"],
		"impersonate": "chrome_133",
		"cookie_file": "/path/to/cookie.txt"
	}`
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("failed to load config file: %v", err)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected host 0.0.0.0, got %s", cfg.Host)
	}
	if cfg.DefaultModel != "gemini-3.5-flash-thinking" {
		t.Errorf("expected model gemini-3.5-flash-thinking, got %s", cfg.DefaultModel)
	}
	if len(cfg.APIKeys) != 2 || cfg.APIKeys[0] != "sk-test-key-1" {
		t.Errorf("unexpected api keys: %v", cfg.APIKeys)
	}
	if cfg.CookieFile != "/path/to/cookie.txt" {
		t.Errorf("expected cookie file /path/to/cookie.txt, got %s", cfg.CookieFile)
	}
}

func TestFindCookie(t *testing.T) {
	tmpDir := t.TempDir()
	cookieFile := filepath.Join(tmpDir, "cookie.txt")
	_ = os.WriteFile(cookieFile, []byte("SID=test; SAPISID=test;"), 0600)

	// Set env var to test direct resolution
	t.Setenv("BOB_GEMINI_FREE_COOKIE_FILE", cookieFile)
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if cfg.CookieFile != cookieFile {
		t.Errorf("expected %s, got %s", cookieFile, cfg.CookieFile)
	}
}
