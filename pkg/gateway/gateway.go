package gateway

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

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

// WithRetry configures the maximum upstream retry attempts and delay in seconds.
func WithRetry(attempts int, delaySec int) Option {
	return func(c *config.Config) {
		if attempts >= 0 {
			c.RetryAttempts = attempts
		}
		if delaySec >= 0 {
			c.RetryDelaySec = delaySec
		}
	}
}

// WithTimeout sets the upstream HTTP client per-request timeout in seconds.
func WithTimeout(timeoutSec int) Option {
	return func(c *config.Config) {
		if timeoutSec > 0 {
			c.RequestTimeoutSec = timeoutSec
		}
	}
}

// WithCookiePoolDir scans a directory for *.txt cookie files to populate the pool.
func WithCookiePoolDir(dir string) Option {
	return func(c *config.Config) {
		if entries, err := os.ReadDir(dir); err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".txt") {
					c.CookiePool = append(c.CookiePool, filepath.Join(dir, e.Name()))
				}
			}
		}
	}
}

// resolveEngineVersion returns the build-time module version, falling back to "dev".
func resolveEngineVersion(override string) string {
	if override != "" {
		return override
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

// WithVersion sets an explicit version string reported by the health endpoint.
// When omitted, the version is resolved from Go build info automatically.
func WithVersion(v string) Option {
	return func(c *config.Config) {
		c.Version = v
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
	version := resolveEngineVersion(cfg.Version)
	app := server.New(cfg, version)
	return &Engine{app: app}
}

// Handler returns the standard http.Handler for mounting into HTTP servers or routers.
func (e *Engine) Handler() http.Handler {
	return e.app.Handler()
}

// Generate performs a synchronous, single-turn text generation request directly in Go.
func (e *Engine) Generate(ctx context.Context, prompt string, model string) (string, error) {
	return e.GenerateWithMedia(ctx, prompt, model, nil)
}

// GenerateWithMedia performs a synchronous text generation request with multimodal file references directly in Go.
func (e *Engine) GenerateWithMedia(ctx context.Context, prompt string, model string, fileRefs []string) (string, error) {
	if model == "" {
		model = e.app.Cfg.DefaultModel
	}
	resolved, err := models.Resolve(model, e.app.Cfg.DefaultModel)
	if err != nil {
		return "", err
	}
	return e.app.Gem.GenerateContext(ctx, prompt, resolved.Mode, resolved.Think, fileRefs, resolved.Extra)
}

// GenerateStream performs real-time streaming text generation directly in Go, invoking onDelta for each token chunk.
func (e *Engine) GenerateStream(ctx context.Context, prompt string, model string, onDelta func(delta string) error) error {
	return e.GenerateStreamWithMedia(ctx, prompt, model, nil, onDelta)
}

// GenerateStreamWithMedia performs real-time streaming text generation with multimodal file references directly in Go.
func (e *Engine) GenerateStreamWithMedia(ctx context.Context, prompt string, model string, fileRefs []string, onDelta func(delta string) error) error {
	if model == "" {
		model = e.app.Cfg.DefaultModel
	}
	resolved, err := models.Resolve(model, e.app.Cfg.DefaultModel)
	if err != nil {
		return err
	}
	return e.app.Gem.GenerateStreamContext(ctx, prompt, resolved.Mode, resolved.Think, fileRefs, resolved.Extra, onDelta)
}

// NewHandler creates a standalone standard http.Handler that can be mounted into any Go HTTP server or router.
func NewHandler(opts ...Option) http.Handler {
	engine := NewEngine(opts...)
	return engine.Handler()
}
