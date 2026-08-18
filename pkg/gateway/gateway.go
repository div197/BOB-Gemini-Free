package gateway

import (
	"net/http"

	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/server"
)

// Option configures the embedded gateway engine.
type Option func(*config.Config)

// WithPort sets the default listening port.
func WithPort(port int) Option {
	return func(c *config.Config) {
		c.Port = port
	}
}

// WithHost sets the binding host interface.
func WithHost(host string) Option {
	return func(c *config.Config) {
		c.Host = host
	}
}

// WithCookieFile sets the session cookie file path for Gemini Pro and Imagen access.
func WithCookieFile(path string) Option {
	return func(c *config.Config) {
		c.CookieFile = path
	}
}

// WithDefaultModel sets the default model when none is specified.
func WithDefaultModel(model string) Option {
	return func(c *config.Config) {
		c.DefaultModel = model
	}
}

// WithAPIKeys sets authorized API keys for request authentication.
func WithAPIKeys(keys ...string) Option {
	return func(c *config.Config) {
		c.APIKeys = keys
	}
}

// WithProxy sets an upstream HTTP / SOCKS5 proxy URL.
func WithProxy(proxy string) Option {
	return func(c *config.Config) {
		c.Proxy = proxy
	}
}

// NewHandler creates a standalone standard http.Handler that can be mounted into any Go HTTP server or router.
func NewHandler(opts ...Option) http.Handler {
	cfg := config.Default()
	for _, opt := range opts {
		opt(&cfg)
	}
	app := server.New(cfg, "v0.1.0")
	return app.Handler()
}
