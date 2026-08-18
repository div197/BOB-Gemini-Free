package gateway

import (
	"context"
	"net/http"

	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/models"
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

// WithCookiePool configures a pool of account cookie files for automatic round-robin dispatch.
func WithCookiePool(paths ...string) Option {
	return func(c *config.Config) {
		c.CookiePool = paths
	}
}

// WithAuthUser sets the target Google multi-account index (e.g. "0", "1").
func WithAuthUser(authUser string) Option {
	return func(c *config.Config) {
		c.AuthUser = authUser
	}
}

// WithImpersonate sets the TLS browser fingerprint profile (e.g. "chrome", "firefox", "safari").
func WithImpersonate(profile string) Option {
	return func(c *config.Config) {
		c.Impersonate = profile
	}
}

// WithLogRequests enables or disables request lifecycle logging.
func WithLogRequests(enabled bool) Option {
	return func(c *config.Config) {
		c.LogRequests = enabled
	}
}

// Engine represents an embedded in-process Gemini inference engine.
type Engine struct {
	app *server.App
}

// NewEngine creates an embedded in-process engine for direct programmatic Go inference.
func NewEngine(opts ...Option) *Engine {
	cfg := config.Default()
	for _, opt := range opts {
		opt(&cfg)
	}
	app := server.New(cfg, "v0.1.1")
	return &Engine{app: app}
}

// Handler returns the standard http.Handler for mounting into HTTP servers or routers.
func (e *Engine) Handler() http.Handler {
	return e.app.Handler()
}

// Generate performs a synchronous, single-turn text generation request directly in Go.
func (e *Engine) Generate(ctx context.Context, prompt string, model string) (string, error) {
	if model == "" {
		model = e.app.Cfg.DefaultModel
	}
	resolved, err := models.Resolve(model, e.app.Cfg.DefaultModel)
	if err != nil {
		return "", err
	}
	return e.app.Gem.GenerateContext(ctx, prompt, resolved.Mode, resolved.Think, nil, resolved.Extra)
}

// GenerateStream performs real-time streaming text generation directly in Go, invoking onDelta for each token chunk.
func (e *Engine) GenerateStream(ctx context.Context, prompt string, model string, onDelta func(delta string) error) error {
	if model == "" {
		model = e.app.Cfg.DefaultModel
	}
	resolved, err := models.Resolve(model, e.app.Cfg.DefaultModel)
	if err != nil {
		return err
	}
	return e.app.Gem.GenerateStreamContext(ctx, prompt, resolved.Mode, resolved.Think, nil, resolved.Extra, onDelta)
}

// NewHandler creates a standalone standard http.Handler that can be mounted into any Go HTTP server or router.
func NewHandler(opts ...Option) http.Handler {
	engine := NewEngine(opts...)
	return engine.Handler()
}
