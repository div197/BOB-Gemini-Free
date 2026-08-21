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
	if cfg.Port != 9610 {
		t.Errorf("expected port 9610, got %d", cfg.Port)
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
		"allowed_origins": ["https://studio.example.test"],
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
	if len(cfg.AllowedOrigins) != 1 || cfg.AllowedOrigins[0] != "https://studio.example.test" {
		t.Errorf("unexpected allowed origins: %v", cfg.AllowedOrigins)
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

func TestEnvVarDefaultModel(t *testing.T) {
	t.Setenv("BOB_GEMINI_FREE_DEFAULT_MODEL", "gemini-3.7-flash")
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if cfg.DefaultModel != "gemini-3.7-flash" {
		t.Errorf("expected gemini-3.7-flash from env, got %s", cfg.DefaultModel)
	}
}

func TestEnvVarAuthUser(t *testing.T) {
	t.Setenv("BOB_GEMINI_FREE_AUTH_USER", "1")
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if cfg.AuthUser != "1" {
		t.Errorf("expected auth_user 1 from env, got %s", cfg.AuthUser)
	}
}

func TestEnvVarAllowedOrigins(t *testing.T) {
	t.Setenv("BOB_GEMINI_FREE_ALLOWED_ORIGINS", "https://one.example, https://two.example")
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if len(cfg.AllowedOrigins) != 2 || cfg.AllowedOrigins[1] != "https://two.example" {
		t.Fatalf("unexpected allowed origins from env: %v", cfg.AllowedOrigins)
	}
}

func TestEnvVarLogRequests(t *testing.T) {
	for _, val := range []string{"true", "1", "yes"} {
		t.Setenv("BOB_GEMINI_FREE_LOG_REQUESTS", val)
		cfg, err := Load("")
		if err != nil {
			t.Fatalf("unexpected load error: %v", err)
		}
		if !cfg.LogRequests {
			t.Errorf("expected log_requests=true for env value %q", val)
		}
	}
}

func TestEnvVarRetryAttempts(t *testing.T) {
	t.Setenv("BOB_GEMINI_FREE_RETRY_ATTEMPTS", "5")
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if cfg.RetryAttempts != 5 {
		t.Errorf("expected retry_attempts 5, got %d", cfg.RetryAttempts)
	}
}

func TestRetryAttemptsClampedFromEnv(t *testing.T) {
	t.Setenv("BOB_GEMINI_FREE_RETRY_ATTEMPTS", "0")
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if cfg.RetryAttempts != 1 {
		t.Errorf("expected retry_attempts to be clamped to 1, got %d", cfg.RetryAttempts)
	}
}

func TestRetryAttemptsClampedFromConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	content := `{"retry_attempts": 0}`
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("failed to load config file: %v", err)
	}
	if cfg.RetryAttempts != 1 {
		t.Errorf("expected retry_attempts to be clamped to 1, got %d", cfg.RetryAttempts)
	}
}

func TestEnvVarRequestTimeout(t *testing.T) {
	t.Setenv("BOB_GEMINI_FREE_REQUEST_TIMEOUT_SEC", "300")
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if cfg.RequestTimeoutSec != 300 {
		t.Errorf("expected request_timeout_sec 300, got %d", cfg.RequestTimeoutSec)
	}
}

func TestEnvVarCookiePoolDir(t *testing.T) {
	tmpDir := t.TempDir()
	c1 := filepath.Join(tmpDir, "account1.txt")
	c2 := filepath.Join(tmpDir, "account2.txt")
	_ = os.WriteFile(c1, []byte("cookie1"), 0600)
	_ = os.WriteFile(c2, []byte("cookie2"), 0600)

	t.Setenv("BOB_GEMINI_FREE_COOKIE_POOL_DIR", tmpDir)
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if len(cfg.CookiePool) < 2 {
		t.Errorf("expected at least 2 cookie pool files from dir, got %d", len(cfg.CookiePool))
	}
}
