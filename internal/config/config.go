package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/div197/bob-gemini-free/internal/models"
)

type Config struct {
	Port              int      `json:"port"`
	Host              string   `json:"host"`
	RetryAttempts     int      `json:"retry_attempts"`
	RetryDelaySec     int      `json:"retry_delay_sec"`
	RequestTimeoutSec int      `json:"request_timeout_sec"`
	GeminiBL          string   `json:"gemini_bl"`
	AuthUser          string   `json:"auth_user"`
	XSRFToken         string   `json:"xsrf_token"`
	DefaultModel      string   `json:"default_model"`
	LogRequests       bool     `json:"log_requests"`
	CookieFile        string   `json:"cookie_file"`
	CookiePool        []string `json:"cookie_pool,omitempty"`
	Proxy             string   `json:"proxy"`
	APIKeys           []string `json:"api_keys"`
	// GeminiAPIKey is a single, opt-in Developer API credential. It is loaded
	// only from the process environment and deliberately excluded from JSON so
	// config saves/exported diagnostics cannot persist it.
	GeminiAPIKey   string                        `json:"-"`
	AllowedOrigins []string                      `json:"allowed_origins,omitempty"`
	Impersonate    string                        `json:"impersonate"`
	Version        string                        `json:"version,omitempty"`
	CustomModels   map[string]models.Model       `json:"custom_models,omitempty"`
	CustomPricing  map[string]models.PricingInfo `json:"custom_pricing,omitempty"`
}

const (
	DefaultRequestTimeoutSec = 180
	MaxRetryAttempts         = 10
	MaxRetryDelaySec         = 60
	MaxRequestTimeoutSec     = 3600
)

func Default() Config {
	return Config{
		Port:              9610,
		Host:              "127.0.0.1",
		RetryAttempts:     3,
		RetryDelaySec:     2,
		RequestTimeoutSec: DefaultRequestTimeoutSec,
		GeminiBL:          "boq_assistant-bard-web-server_20260716.08_p0",
		AuthUser:          "",
		XSRFToken:         "",
		DefaultModel:      "gemini-3.6-flash",
		LogRequests:       true,
		CookieFile:        "",
		CookiePool:        []string{},
		Proxy:             "",
		APIKeys:           []string{},
		AllowedOrigins:    []string{},
		Impersonate:       "",
	}
}

func Normalize(cfg *Config) {
	if cfg == nil {
		return
	}
	if cfg.RetryAttempts < 1 {
		cfg.RetryAttempts = 1
	}
	if cfg.RetryAttempts > MaxRetryAttempts {
		cfg.RetryAttempts = MaxRetryAttempts
	}
	if cfg.RetryDelaySec < 0 {
		cfg.RetryDelaySec = 0
	}
	if cfg.RetryDelaySec > MaxRetryDelaySec {
		cfg.RetryDelaySec = MaxRetryDelaySec
	}
	if cfg.RequestTimeoutSec <= 0 {
		cfg.RequestTimeoutSec = DefaultRequestTimeoutSec
	}
	if cfg.RequestTimeoutSec > MaxRequestTimeoutSec {
		cfg.RequestTimeoutSec = MaxRequestTimeoutSec
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return cfg, err
		}

		var aux struct {
			Port              *int                           `json:"port"`
			Host              *string                        `json:"host"`
			RetryAttempts     *int                           `json:"retry_attempts"`
			RetryDelaySec     *int                           `json:"retry_delay_sec"`
			RequestTimeoutSec *int                           `json:"request_timeout_sec"`
			GeminiBL          *string                        `json:"gemini_bl"`
			AuthUser          *string                        `json:"auth_user"`
			XSRFToken         *string                        `json:"xsrf_token"`
			DefaultModel      *string                        `json:"default_model"`
			LogRequests       *bool                          `json:"log_requests"`
			CookieFile        *string                        `json:"cookie_file"`
			CookiePool        *[]string                      `json:"cookie_pool"`
			Proxy             *string                        `json:"proxy"`
			APIKeys           *[]string                      `json:"api_keys"`
			AllowedOrigins    *[]string                      `json:"allowed_origins"`
			Impersonate       *string                        `json:"impersonate"`
			CustomModels      *map[string]models.Model       `json:"custom_models"`
			CustomPricing     *map[string]models.PricingInfo `json:"custom_pricing"`
		}

		if err := json.Unmarshal(data, &aux); err != nil {
			return cfg, err
		}

		if aux.Port != nil {
			cfg.Port = *aux.Port
		}
		if aux.Host != nil {
			cfg.Host = *aux.Host
		}
		if aux.RetryAttempts != nil {
			cfg.RetryAttempts = *aux.RetryAttempts
		}
		if aux.RetryDelaySec != nil {
			cfg.RetryDelaySec = *aux.RetryDelaySec
		}
		if aux.RequestTimeoutSec != nil {
			cfg.RequestTimeoutSec = *aux.RequestTimeoutSec
		}
		if aux.GeminiBL != nil {
			cfg.GeminiBL = *aux.GeminiBL
		}
		if aux.AuthUser != nil {
			cfg.AuthUser = *aux.AuthUser
		}
		if aux.XSRFToken != nil {
			cfg.XSRFToken = *aux.XSRFToken
		}
		if aux.DefaultModel != nil {
			cfg.DefaultModel = *aux.DefaultModel
		}
		if aux.LogRequests != nil {
			cfg.LogRequests = *aux.LogRequests
		}
		if aux.CookieFile != nil {
			cfg.CookieFile = *aux.CookieFile
		}
		if aux.CookiePool != nil {
			cfg.CookiePool = *aux.CookiePool
		}
		if aux.Proxy != nil {
			cfg.Proxy = *aux.Proxy
		}
		if aux.APIKeys != nil {
			cfg.APIKeys = *aux.APIKeys
		}
		if aux.AllowedOrigins != nil {
			cfg.AllowedOrigins = *aux.AllowedOrigins
		}
		if aux.Impersonate != nil {
			cfg.Impersonate = *aux.Impersonate
		}
		if aux.CustomModels != nil {
			cfg.CustomModels = *aux.CustomModels
			for name, m := range cfg.CustomModels {
				models.RegisterModel(name, m)
			}
		}
		if aux.CustomPricing != nil {
			cfg.CustomPricing = *aux.CustomPricing
			for name, p := range cfg.CustomPricing {
				models.RegisterPricing(name, p)
			}
		}
	}

	// Environment variable overrides (useful for Docker & cloud environments)
	if envHost := os.Getenv("BOB_GEMINI_FREE_HOST"); envHost != "" {
		cfg.Host = envHost
	}
	if envPort := os.Getenv("BOB_GEMINI_FREE_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 {
			cfg.Port = p
		}
	}
	if envCookie := os.Getenv("BOB_GEMINI_FREE_COOKIE_FILE"); envCookie != "" {
		cfg.CookieFile = envCookie
	}
	if envPool := os.Getenv("BOB_GEMINI_FREE_COOKIE_POOL"); envPool != "" {
		for _, f := range strings.Split(envPool, ",") {
			if tf := strings.TrimSpace(f); tf != "" {
				cfg.CookiePool = append(cfg.CookiePool, tf)
			}
		}
	}
	if envPoolDir := os.Getenv("BOB_GEMINI_FREE_COOKIE_POOL_DIR"); envPoolDir != "" {
		if entries, err := os.ReadDir(envPoolDir); err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".txt") {
					cfg.CookiePool = append(cfg.CookiePool, filepath.Join(envPoolDir, e.Name()))
				}
			}
		}
	}
	if cfg.CookieFile == "" {
		cfg.CookieFile = FindCookie()
	}
	if len(cfg.CookiePool) == 0 {
		cfg.CookiePool = FindCookiePool()
	}
	if envProxy := os.Getenv("BOB_GEMINI_FREE_PROXY"); envProxy != "" {
		cfg.Proxy = envProxy
	}
	if envImpersonate := os.Getenv("BOB_GEMINI_FREE_IMPERSONATE"); envImpersonate != "" {
		cfg.Impersonate = envImpersonate
	}
	if envKeys := os.Getenv("BOB_GEMINI_FREE_API_KEYS"); envKeys != "" {
		var keys []string
		for _, k := range strings.Split(envKeys, ",") {
			if trimmed := strings.TrimSpace(k); trimmed != "" {
				keys = append(keys, trimmed)
			}
		}
		if len(keys) > 0 {
			cfg.APIKeys = keys
		}
	}
	if envGeminiAPIKey := os.Getenv("BOB_GEMINI_FREE_GEMINI_API_KEY"); envGeminiAPIKey != "" {
		// Keep this a single key. Never parse a comma-separated list: key
		// rotation would evade provider quotas and make student usage opaque.
		cfg.GeminiAPIKey = strings.TrimSpace(envGeminiAPIKey)
	}
	if envOrigins := os.Getenv("BOB_GEMINI_FREE_ALLOWED_ORIGINS"); envOrigins != "" {
		var origins []string
		for _, origin := range strings.Split(envOrigins, ",") {
			if trimmed := strings.TrimSpace(origin); trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
		if len(origins) > 0 {
			cfg.AllowedOrigins = origins
		}
	}
	if envModel := os.Getenv("BOB_GEMINI_FREE_DEFAULT_MODEL"); envModel != "" {
		cfg.DefaultModel = envModel
	}
	if envAuthUser := os.Getenv("BOB_GEMINI_FREE_AUTH_USER"); envAuthUser != "" {
		cfg.AuthUser = envAuthUser
	}
	if envLog := os.Getenv("BOB_GEMINI_FREE_LOG_REQUESTS"); envLog != "" {
		cfg.LogRequests = envLog == "true" || envLog == "1" || envLog == "yes"
	}
	if envRetry := os.Getenv("BOB_GEMINI_FREE_RETRY_ATTEMPTS"); envRetry != "" {
		if n, err := strconv.Atoi(envRetry); err == nil {
			cfg.RetryAttempts = n
		}
	}
	if envRetryDelay := os.Getenv("BOB_GEMINI_FREE_RETRY_DELAY_SEC"); envRetryDelay != "" {
		if n, err := strconv.Atoi(envRetryDelay); err == nil && n >= 0 {
			cfg.RetryDelaySec = n
		}
	}
	if envTimeout := os.Getenv("BOB_GEMINI_FREE_REQUEST_TIMEOUT_SEC"); envTimeout != "" {
		if n, err := strconv.Atoi(envTimeout); err == nil && n > 0 {
			cfg.RequestTimeoutSec = n
		}
	}

	Normalize(&cfg)
	return cfg, nil
}

func Find() string {
	paths := []string{"./config.json"}
	home, err := os.UserHomeDir()
	if err == nil {
		paths = append(paths, filepath.Join(home, ".config", "bob-gemini-free", "config.json"))
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func FindCookie() string {
	paths := []string{"./cookie.txt"}
	home, err := os.UserHomeDir()
	if err == nil {
		paths = append(paths, filepath.Join(home, ".config", "bob-gemini-free", "cookie.txt"))
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func FindCookiePool() []string {
	var pool []string
	dirs := []string{"./cookies"}
	home, err := os.UserHomeDir()
	if err == nil {
		dirs = append(dirs, filepath.Join(home, ".config", "bob-gemini-free", "cookies"))
	}

	for _, d := range dirs {
		if entries, err := os.ReadDir(d); err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".txt") {
					pool = append(pool, filepath.Join(d, e.Name()))
				}
			}
		}
	}
	return pool
}
