package server

import (
	"log"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/gemini"
	"github.com/div197/bob-gemini-free/internal/metrics"
	"github.com/div197/bob-gemini-free/internal/multimodal"
)

type App struct {
	Cfg             config.Config
	Gem             *gemini.Client
	Tokens          *multimodal.TokenCache
	HTTPClient      *http.Client
	Logf            func(format string, args ...any)
	Version         string
	RequestsServed  atomic.Uint64
	TokensProcessed atomic.Uint64
	StartTime       time.Time
	ImageCache      sync.Map // Caches SHA256 -> Scotty FileRef to prevent redundant uploads in long multi-turn vision chats
	Metrics         *metrics.Registry
}

func createHTTPClient(cfg config.Config) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     1000,
		IdleConnTimeout:     90 * time.Second,
	}
	if cfg.Proxy != "" {
		if proxyURL, err := url.Parse(cfg.Proxy); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	} else {
		transport.Proxy = http.ProxyFromEnvironment
	}
	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}

func New(cfg config.Config, version string) *App {
	config.Normalize(&cfg)
	gemClient := gemini.NewClient(cfg)
	registry := metrics.New()
	gemClient.Metrics = registry

	logFn := func(format string, args ...any) {
		if cfg.LogRequests {
			log.Printf(format, args...)
		}
	}

	httpClient := createHTTPClient(cfg)

	app := &App{
		Cfg:        cfg,
		Gem:        gemClient,
		Tokens:     multimodal.NewTokenCache(cfg, gemClient.Cookies, httpClient),
		HTTPClient: httpClient,
		Logf:       logFn,
		Version:    version,
		StartTime:  time.Now(),
		Metrics:    registry,
	}

	if gemClient.Pool != nil {
		gemClient.Pool.StartAutoReload(30 * time.Second)
	}

	return app
}

func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", a.handleHealth)
	mux.HandleFunc("GET /healthz", a.handleHealthz)
	mux.HandleFunc("GET /v1/metrics", a.handleMetrics)
	mux.HandleFunc("GET /playground", a.handlePlayground)
	mux.HandleFunc("GET /ui", a.handlePlayground)
	mux.HandleFunc("GET /favicon.ico", a.handleFavicon)
	mux.HandleFunc("GET /v1/models", a.handleModels)
	mux.HandleFunc("GET /v1/models/{model}", a.handleSingleModel)
	mux.HandleFunc("POST /v1/chat/completions", a.handleChat)
	mux.HandleFunc("POST /v1/responses", a.handleResponses)
	mux.HandleFunc("POST /v1/messages", a.handleAnthropicMessages)
	mux.HandleFunc("POST /v1/images/generations", a.handleImageGenerations)
	mux.HandleFunc("POST /v1/tokens/count", a.handleCountTokens)
	mux.HandleFunc("GET /v1/update/check", a.handleUpdateCheck)

	mux.HandleFunc("GET /v1beta/models", a.handleGoogleModels)
	mux.HandleFunc("POST /v1beta/models/{target}", a.handleGoogleGenerate)

	handler := a.withAuthAndLogging(mux)
	handler = a.withCORS(handler)

	return handler
}
