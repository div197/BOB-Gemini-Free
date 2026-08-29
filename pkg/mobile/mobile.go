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
	"path/filepath"
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
}

var (
	defaultGateway *MobileGateway
	once           sync.Once
)

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

	if host == "" {
		host = "127.0.0.1"
	}
	if port <= 0 {
		port = 9610
	}

	cfg := config.Default()
	cfg.Host = host
	cfg.Port = port
	cfg.LogRequests = false

	if cookieContent != "" {
		tmpDir := os.TempDir()
		cookieFile := filepath.Join(tmpDir, "bob_mobile_cookie.txt")
		_ = os.WriteFile(cookieFile, []byte(cookieContent), 0600)
		cfg.CookieFile = cookieFile
		m.cookiePath = cookieFile
	}

	app := server.New(cfg, "v0.2.0-mobile")

	addr := fmt.Sprintf("%s:%d", host, port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		// Fallback to random available port if specified port is occupied
		ln, err = net.Listen("tcp", fmt.Sprintf("%s:0", host))
		if err != nil {
			return "", fmt.Errorf("failed to bind mobile port: %w", err)
		}
	}

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

	if m.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = m.httpSrv.Shutdown(ctx)
	}

	if m.cookiePath != "" {
		_ = os.Remove(m.cookiePath)
		m.cookiePath = ""
	}

	m.running = false
	m.baseURL = ""
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
	m.mu.RUnlock()

	if app == nil {
		return "", fmt.Errorf("mobile gateway is not running; call Start() first")
	}

	resolved, err := models.Resolve(modelName, "gemini-3.7-flash")
	if err != nil {
		resolved = models.Resolved{Mode: 1, Think: 4}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	return app.Gem.GenerateContext(ctx, prompt, resolved.Mode, resolved.Think, nil, resolved.Extra)
}

// GenerateStream executes a streaming request, invoking the StreamCallback on each incoming chunk.
func (m *MobileGateway) GenerateStream(prompt string, modelName string, cb StreamCallback) error {
	m.mu.RLock()
	app := m.app
	m.mu.RUnlock()

	if app == nil {
		return fmt.Errorf("mobile gateway is not running; call Start() first")
	}

	resolved, err := models.Resolve(modelName, "gemini-3.7-flash")
	if err != nil {
		resolved = models.Resolved{Mode: 1, Think: 4}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
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
	m.mu.RUnlock()

	if app == nil {
		return "", fmt.Errorf("mobile gateway is not running; call Start() first")
	}

	engine := refiner.NewEngine()
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	resolved, err := models.Resolve("gemini-3.7-flash-thinking", "gemini-3.7-flash")
	if err != nil {
		resolved = models.Resolved{Mode: 2, Think: 0}
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
