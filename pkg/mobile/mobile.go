// Package mobile provides an experimental Go bridge for future Android/iOS
// bindings. It is not itself a native mobile application and currently starts
// a local HTTP gateway backed by the existing Google web client.
package mobile

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/format"
	"github.com/div197/bob-gemini-free/internal/models"
	"github.com/div197/bob-gemini-free/internal/refiner"
	"github.com/div197/bob-gemini-free/internal/server"
)

// StreamCallback defines the interface an eventual mobile host can implement
// to receive real-time token deltas.
type StreamCallback interface {
	OnDelta(delta string)
	OnComplete(totalTokens int, errStr string)
}

// MobileGateway provides an experimental local HTTP bridge to the BOB Gemini
// Free core engine.
type MobileGateway struct {
	mu         sync.RWMutex
	app        *server.App
	httpSrv    *http.Server
	listener   net.Listener
	baseURL    string
	running    bool
	cookiePath string
	runCtx     context.Context
	runCancel  context.CancelFunc
}

var (
	defaultGateway *MobileGateway
	once           sync.Once
)

const maxMobileCookieBytes = 1 << 20

func normalizeMobileHost(host string) (string, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return "127.0.0.1", nil
	}
	cleanHost := strings.Trim(host, "[]")
	if strings.EqualFold(cleanHost, "localhost") {
		return "127.0.0.1", nil
	}
	ip := net.ParseIP(cleanHost)
	if ip == nil || !ip.IsLoopback() {
		return "", fmt.Errorf("mobile gateway host must be loopback, got %q", host)
	}
	return cleanHost, nil
}

func resolveMobileModel(modelName string) (models.Resolved, error) {
	if strings.TrimSpace(modelName) == "" {
		modelName = "gemini-3.7-flash"
	}
	return models.ResolveStrict(modelName, "gemini-3.7-flash")
}

// GetDefaultGateway returns the singleton mobile gateway instance.
func GetDefaultGateway() *MobileGateway {
	once.Do(func() {
		defaultGateway = &MobileGateway{}
	})
	return defaultGateway
}

// Start boots the local HTTP gateway. If cookieContent is provided, the current
// experimental implementation writes it to a temporary file for the Go client;
// it does not provide platform keystore integration.
func (m *MobileGateway) Start(port int, host string, cookieContent string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return m.baseURL, nil
	}

	var err error
	host, err = normalizeMobileHost(host)
	if err != nil {
		return "", err
	}
	if port <= 0 {
		port = 9610
	}
	if port > 65535 {
		return "", fmt.Errorf("mobile gateway port %d is out of range", port)
	}

	cfg := config.Default()
	cfg.Host = host
	cfg.Port = port
	cfg.LogRequests = false

	if cookieContent != "" {
		if len(cookieContent) > maxMobileCookieBytes {
			return "", fmt.Errorf("mobile cookie content exceeds %d bytes", maxMobileCookieBytes)
		}
		cookieFile, createErr := os.CreateTemp("", "bob-mobile-cookie-*.txt")
		if createErr != nil {
			return "", fmt.Errorf("create mobile cookie file: %w", createErr)
		}
		cookiePath := cookieFile.Name()
		cleanupCookie := func() {
			_ = cookieFile.Close()
			_ = os.Remove(cookiePath)
		}
		if _, writeErr := cookieFile.WriteString(cookieContent); writeErr != nil {
			cleanupCookie()
			return "", fmt.Errorf("write mobile cookie file: %w", writeErr)
		}
		if closeErr := cookieFile.Close(); closeErr != nil {
			_ = os.Remove(cookiePath)
			return "", fmt.Errorf("close mobile cookie file: %w", closeErr)
		}
		if chmodErr := os.Chmod(cookiePath, 0600); chmodErr != nil {
			_ = os.Remove(cookiePath)
			return "", fmt.Errorf("protect mobile cookie file: %w", chmodErr)
		}
		cfg.CookieFile = cookiePath
		m.cookiePath = cookiePath
	}

	app := server.New(cfg, "v0.2.0-mobile")

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		// Fallback to random available port if specified port is occupied
		ln, err = net.Listen("tcp", net.JoinHostPort(host, "0"))
		if err != nil {
			app.Close()
			if m.cookiePath != "" {
				_ = os.Remove(m.cookiePath)
				m.cookiePath = ""
			}
			return "", fmt.Errorf("failed to bind mobile port: %w", err)
		}
	}

	m.runCtx, m.runCancel = context.WithCancel(context.Background())
	m.listener = ln
	m.app = app
	m.baseURL = fmt.Sprintf("http://%s", ln.Addr().String())

	httpSrv := &http.Server{
		Handler: app.Handler(),
	}
	m.httpSrv = httpSrv
	m.running = true

	go func() {
		_ = httpSrv.Serve(ln)
	}()

	return m.baseURL, nil
}

// Stop shuts down the in-process mobile gateway and cleans up temporary credentials.
func (m *MobileGateway) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	app := m.app
	if m.runCancel != nil {
		m.runCancel()
	}
	if m.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = m.httpSrv.Shutdown(ctx)
	}
	if app != nil {
		app.Close()
	}

	if m.cookiePath != "" {
		_ = os.Remove(m.cookiePath)
		m.cookiePath = ""
	}

	m.running = false
	m.baseURL = ""
	m.app = nil
	m.httpSrv = nil
	m.listener = nil
	m.runCtx = nil
	m.runCancel = nil
	return nil
}

// IsRunning returns whether the mobile gateway is actively listening.
func (m *MobileGateway) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// GetURL returns the active local HTTP endpoint (e.g. http://127.0.0.1:9610).
func (m *MobileGateway) GetURL() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.baseURL
}

// Generate executes a synchronous generation request through the configured
// upstream client.
func (m *MobileGateway) Generate(prompt string, modelName string) (string, error) {
	m.mu.RLock()
	app := m.app
	running := m.running
	runCtx := m.runCtx
	m.mu.RUnlock()

	if !running || app == nil {
		return "", fmt.Errorf("mobile gateway is not running; call Start() first")
	}

	resolved, err := resolveMobileModel(modelName)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(runCtx, 120*time.Second)
	defer cancel()

	return app.Gem.GenerateContext(ctx, prompt, resolved.Mode, resolved.Think, nil, resolved.Extra)
}

// GenerateStream executes a streaming request, invoking the StreamCallback on each incoming chunk.
func (m *MobileGateway) GenerateStream(prompt string, modelName string, cb StreamCallback) error {
	m.mu.RLock()
	app := m.app
	running := m.running
	runCtx := m.runCtx
	m.mu.RUnlock()

	if !running || app == nil {
		return fmt.Errorf("mobile gateway is not running; call Start() first")
	}

	resolved, err := resolveMobileModel(modelName)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(runCtx, 180*time.Second)
	defer cancel()

	var totalTokens int
	err = app.Gem.GenerateStreamContext(ctx, prompt, resolved.Mode, resolved.Think, nil, resolved.Extra, func(chunk string) error {
		totalTokens += format.EstimateTokens(chunk)
		if cb != nil {
			cb.OnDelta(chunk)
		}
		return nil
	})

	if cb != nil {
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}
		cb.OnComplete(totalTokens, errStr)
	}

	return err
}

// CountTokens provides a local subword token estimate for typed mobile inputs.
func (m *MobileGateway) CountTokens(text string) int {
	return format.EstimateTokens(text)
}

// Refine executes the three-stage reasoning orchestration through the configured
// upstream client.
func (m *MobileGateway) Refine(prompt string) (string, error) {
	m.mu.RLock()
	app := m.app
	running := m.running
	runCtx := m.runCtx
	m.mu.RUnlock()

	if !running || app == nil {
		return "", fmt.Errorf("mobile gateway is not running; call Start() first")
	}

	engine := refiner.NewEngine()
	ctx, cancel := context.WithTimeout(runCtx, 180*time.Second)
	defer cancel()

	resolved, err := models.ResolveStrict("gemini-3.7-flash-thinking", "gemini-3.7-flash")
	if err != nil {
		return "", err
	}

	res, err := engine.Refine(ctx, prompt, func(ctx context.Context, p string) (string, error) {
		return app.Gem.GenerateContext(ctx, p, resolved.Mode, resolved.Think, nil, resolved.Extra)
	})
	if err != nil {
		return "", err
	}

	return res.FinalOutput, nil
}

// Version returns the mobile gateway semantic version.
func Version() string {
	return "v0.2.0-mobile"
}
