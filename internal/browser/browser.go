package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bogdanfinn/websocket"
	"github.com/div197/bob-gemini-free/internal/gemini"
)

// BrowserInfo stores the detected browser type and binary path.
type BrowserInfo struct {
	Name string
	Path string
}

// FindBrowser locates an installed Chromium-based browser on the host machine.
func FindBrowser() (*BrowserInfo, error) {
	switch runtime.GOOS {
	case "darwin":
		candidates := []struct {
			name string
			path string
		}{
			{"Google Chrome", "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"},
			{"Arc", "/Applications/Arc.app/Contents/MacOS/Arc"},
			{"Brave Browser", "/Applications/Brave Browser.app/Contents/MacOS/Brave Browser"},
			{"Microsoft Edge", "/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge"},
			{"Chromium", "/Applications/Chromium.app/Contents/MacOS/Chromium"},
		}
		for _, c := range candidates {
			if _, err := os.Stat(c.path); err == nil {
				return &BrowserInfo{Name: c.name, Path: c.path}, nil
			}
		}

	case "linux":
		candidates := []string{
			"google-chrome",
			"google-chrome-stable",
			"brave-browser",
			"microsoft-edge",
			"chromium-browser",
			"chromium",
		}
		for _, c := range candidates {
			if path, err := exec.LookPath(c); err == nil {
				return &BrowserInfo{Name: c, Path: path}, nil
			}
		}

	case "windows":
		roots := []string{
			os.Getenv("ProgramFiles"),
			os.Getenv("ProgramFiles(x86)"),
			os.Getenv("LocalAppData"),
		}
		relPaths := []struct {
			name string
			rel  string
		}{
			{"Google Chrome", `Google\Chrome\Application\chrome.exe`},
			{"Microsoft Edge", `Microsoft\Edge\Application\msedge.exe`},
			{"Brave Browser", `BraveSoftware\Brave-Browser\Application\brave.exe`},
		}
		for _, root := range roots {
			if root == "" {
				continue
			}
			for _, p := range relPaths {
				full := filepath.Join(root, p.rel)
				if _, err := os.Stat(full); err == nil {
					return &BrowserInfo{Name: p.name, Path: full}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no Chromium-based browser found on this system (Google Chrome, Brave, Arc, or Edge required)")
}

func getFreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

type cdpVersionResponse struct {
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

type cdpRequest struct {
	ID     int64  `json:"id"`
	Method string `json:"method"`
	Params any    `json:"params,omitempty"`
}

type cdpCookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	HTTPOnly bool   `json:"httpOnly"`
	Secure   bool   `json:"secure"`
}

type cdpCookiesResult struct {
	Cookies []cdpCookie `json:"cookies"`
}

type cdpResponse struct {
	ID     int64            `json:"id"`
	Result cdpCookiesResult `json:"result"`
	Error  any              `json:"error,omitempty"`
}

// LaunchLoginSession opens a standalone Google Gemini sign-in window,
// connects to the Chrome DevTools Protocol, and monitors for active session tokens.
func LaunchLoginSession(ctx context.Context, logFn func(format string, args ...any)) (*gemini.ExtractedCookie, error) {
	if logFn == nil {
		logFn = func(format string, args ...any) {
			log.Printf(format, args...)
		}
	}

	browser, err := FindBrowser()
	if err != nil {
		return nil, err
	}

	port, err := getFreePort()
	if err != nil {
		return nil, fmt.Errorf("failed to allocate free port for browser automation: %w", err)
	}

	tempDir, err := os.MkdirTemp("", "bob-gemini-login-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary profile directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	args := []string{
		fmt.Sprintf("--remote-debugging-port=%d", port),
		fmt.Sprintf("--user-data-dir=%s", tempDir),
		"--app=https://gemini.google.com",
		"--disable-blink-features=AutomationControlled",
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-features=Translate,OptimizationHints",
		"--window-size=1024,800",
	}

	logFn("[*] Found browser: %s (%s)", browser.Name, browser.Path)
	logFn("[*] Launching standalone Google Gemini sign-in window on port %d...", port)

	cmd := exec.CommandContext(ctx, browser.Path, args...)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start %s: %w", browser.Name, err)
	}

	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}()

	// Poll until CDP HTTP endpoint becomes available
	var wsURL string
	httpClient := &http.Client{Timeout: 1 * time.Second}
	deadline := time.Now().Add(15 * time.Second)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		resp, err := httpClient.Get(fmt.Sprintf("http://127.0.0.1:%d/json/version", port))
		if err == nil && resp.StatusCode == http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()

			var v cdpVersionResponse
			if err := json.Unmarshal(bodyBytes, &v); err == nil && v.WebSocketDebuggerURL != "" {
				wsURL = v.WebSocketDebuggerURL
				break
			}
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		time.Sleep(300 * time.Millisecond)
	}

	if wsURL == "" {
		return nil, fmt.Errorf("timed out waiting for browser DevTools interface to start")
	}

	// Connect to DevTools WebSocket
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to browser CDP WebSocket: %w", err)
	}
	defer conn.Close()

	logFn("[*] Connected to browser session. Waiting for Google account sign-in...")

	var reqID int64 = 1
	var lastTokenCount int

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			// Check if user manually closed the browser window
			if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
				return nil, fmt.Errorf("browser window was closed before sign-in completed")
			}

			currID := atomic.AddInt64(&reqID, 1)
			reqMsg := cdpRequest{
				ID:     currID,
				Method: "Storage.getCookies",
			}
			reqBytes, _ := json.Marshal(reqMsg)

			if err := conn.WriteMessage(websocket.TextMessage, reqBytes); err != nil {
				return nil, fmt.Errorf("failed to query cookies via CDP: %w", err)
			}

			_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			_, respBytes, err := conn.ReadMessage()
			if err != nil {
				continue
			}

			var resp cdpResponse
			if err := json.Unmarshal(respBytes, &resp); err != nil {
				continue
			}

			var pairs []string
			hasSAPISID := false
			hasPSID := false
			hasPSIDTS := false

			for _, c := range resp.Result.Cookies {
				if strings.Contains(c.Domain, "google.com") {
					pairs = append(pairs, fmt.Sprintf("%s=%s", c.Name, c.Value))
					if c.Name == "SAPISID" || c.Name == "__Secure-3PAPISID" {
						hasSAPISID = true
					}
					if strings.Contains(c.Name, "PSID") || c.Name == "SID" {
						hasPSID = true
					}
					if strings.Contains(c.Name, "PSIDTS") {
						hasPSIDTS = true
					}
				}
			}

			if len(pairs) != lastTokenCount && len(pairs) > 0 {
				lastTokenCount = len(pairs)
				logFn("[*] Detected %d active session tokens...", len(pairs))
			}

			// Ensure user has fully loaded Gemini and dynamic timestamp session token is present
			if hasSAPISID && hasPSID && (hasPSIDTS || len(pairs) >= 15) {
				rawCookie := strings.Join(pairs, "; ")
				extracted, err := gemini.ExtractCookies(rawCookie)
				if err == nil && extracted.SAPISID != "" {
					logFn("[✔] Successfully captured authenticated session with active SAPISID & PSIDTS!")
					return extracted, nil
				}
			}
		}
	}
}

// LaunchStudioWindow opens the local BOB Gemini Free playground in a standalone,
// chromeless native browser window (App Mode).
func LaunchStudioWindow(ctx context.Context, port int, logFn func(format string, args ...any)) error {
	if logFn == nil {
		logFn = func(format string, args ...any) {
			log.Printf(format, args...)
		}
	}

	browser, err := FindBrowser()
	if err != nil {
		return err
	}

	appURL := fmt.Sprintf("http://127.0.0.1:%d/playground", port)
	args := []string{
		fmt.Sprintf("--app=%s", appURL),
		"--window-size=1200,900",
		"--disable-features=Translate,OptimizationHints",
		"--no-first-run",
		"--no-default-browser-check",
	}

	logFn("🖥️  Opening Native Studio Window using %s...", browser.Name)

	cmd := exec.CommandContext(ctx, browser.Path, args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", browser.Name, err)
	}

	// We don't block. The window runs independently.
	// We detach it so the Go server can continue running.
	go func() {
		_ = cmd.Wait()
		logFn("🖥️  Native Studio Window closed.")
	}()

	return nil
}
