package server

import (
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/div197/bob-gemini-free/internal/format"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Flush() {
	if flusher, ok := lrw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func authorize(r *http.Request, apiKeys []string, allowQueryAPIKey bool) bool {
	if len(apiKeys) == 0 {
		return true
	}

	authHeader := r.Header.Get("Authorization")
	if len(authHeader) >= len("Bearer ") && strings.EqualFold(authHeader[:len("Bearer ")], "Bearer ") {
		token := authHeader[len("Bearer "):]
		for _, key := range apiKeys {
			if subtle.ConstantTimeCompare([]byte(token), []byte(key)) == 1 {
				return true
			}
		}
	}

	for _, h := range []string{"x-api-key", "x-goog-api-key"} {
		headerVal := r.Header.Get(h)
		if headerVal != "" {
			for _, key := range apiKeys {
				if subtle.ConstantTimeCompare([]byte(headerVal), []byte(key)) == 1 {
					return true
				}
			}
		}
	}

	if allowQueryAPIKey {
		if queryKey := r.URL.Query().Get("key"); queryKey != "" {
			for _, key := range apiKeys {
				if subtle.ConstantTimeCompare([]byte(queryKey), []byte(key)) == 1 {
					return true
				}
			}
		}
	}

	return false
}

func (a *App) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		requestScheme := "http"
		if r.TLS != nil {
			requestScheme = "https"
		}
		if origin != "" && !isAllowedOrigin(origin, a.Cfg.AllowedOrigins, r.Host, requestScheme) {
			w.Header().Set("Vary", "Origin")
			writeJSON(w, http.StatusForbidden, map[string]any{
				"error": map[string]any{
					"message": "origin is not allowed",
					"type":    "invalid_request_error",
				},
			})
			return
		}

		if origin == "" {
			// Native clients do not send Origin. Retain the historical wildcard
			// header for them; browsers always send an origin on CORS requests.
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept, Origin, User-Agent, x-api-key, anthropic-version, anthropic-beta, x-goog-api-key, x-goog-api-client, x-client-request-id, x-bob-gemini-api-key, *")
		w.Header().Set("Access-Control-Expose-Headers", "x-request-id, openai-processing-ms, openai-version, content-length")
		w.Header().Set("Access-Control-Allow-Private-Network", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isAllowedOrigin(origin string, configured []string, requestHost, requestScheme string) bool {
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil {
		return false
	}

	normalized := strings.TrimRight(origin, "/")
	for _, candidate := range configured {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" || candidate == "*" {
			continue
		}
		if strings.TrimRight(candidate, "/") == normalized {
			return true
		}
	}

	// A local browser page needs no cross-origin permission when it talks to
	// the exact gateway origin that served it. Do not trust every loopback
	// port by default: another local web server can be malicious and the
	// gateway may hold privileged Google session credentials.
	if strings.EqualFold(parsed.Scheme, requestScheme) && strings.EqualFold(parsed.Host, strings.TrimSpace(requestHost)) {
		return true
	}
	return false
}

const maxRequestBodySize = 32 << 20 // 32 MB limit

const maxRequestIDBytes = 128

func requestIDFor(r *http.Request) string {
	if r != nil {
		requestID := strings.TrimSpace(r.Header.Get("X-Client-Request-Id"))
		if requestID != "" && len(requestID) <= maxRequestIDBytes && strings.IndexFunc(requestID, func(r rune) bool {
			return r < 0x20 || r == 0x7f
		}) < 0 {
			return requestID
		}
	}
	return fmt.Sprintf("req_%s", format.RandHex(16))
}

func (a *App) withAuthAndLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestStart := time.Now()
		if a.Metrics != nil {
			a.Metrics.RequestsTotal.Add(1)
			a.Metrics.RequestsInFlight.Add(1)
			defer func() {
				a.Metrics.RequestsInFlight.Add(-1)
				a.Metrics.RequestLatency.Observe(time.Since(requestStart))
			}()
		}
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
		}
		start := time.Now()
		lrw := newLoggingResponseWriter(w)

		// Set standard request metadata. Do not advertise rate limits that this
		// process does not actually enforce.
		reqID := requestIDFor(r)
		w.Header().Set("x-request-id", reqID)
		w.Header().Set("x-powered-by", "BOB-Gemini-Free / ABCsteps (div197)")
		w.Header().Set("openai-version", "2020-10-01")
		w.Header().Set("x-content-type-options", "nosniff")
		w.Header().Set("x-frame-options", "SAMEORIGIN")
		w.Header().Set("x-xss-protection", "1; mode=block")
		w.Header().Set("referrer-policy", "strict-origin-when-cross-origin")
		w.Header().Set("cross-origin-opener-policy", "same-origin-allow-popups")
		isPublicRoute := false
		// Routes that must always be publicly accessible (never blocked by api_keys):
		// - /playground and /ui: the embedded Web Studio must always load
		// - /sw.js, /manifest.json, /favicon.ico, /icons/: PWA assets
		// - /v1/update/check: version check used by the UI without credentials
		// NOTE: "/" (health) is intentionally NOT in this list — it is protected
		// by api_keys when configured, matching the OpenAI API behavior.
		publicRoutes := []string{"/playground", "/ui", "/favicon.ico", "/manifest.json", "/sw.js", "/v1/update/check", "/healthz"}
		for _, route := range publicRoutes {
			if r.URL.Path == route {
				isPublicRoute = true
				break
			}
		}
		if strings.HasPrefix(r.URL.Path, "/icons/") {
			isPublicRoute = true
		}

		if len(a.Cfg.APIKeys) > 0 && !isPublicRoute && !authorize(r, a.Cfg.APIKeys, a.Cfg.AllowQueryAPIKey) {
			writeJSON(lrw, http.StatusUnauthorized, map[string]any{
				"error": map[string]any{
					"message": "invalid api key",
				},
			})
			a.logRequest(r, lrw.statusCode, time.Since(start))
			return
		}

		next.ServeHTTP(lrw, r)
		duration := time.Since(start)
		w.Header().Set("openai-processing-ms", fmt.Sprintf("%d", duration.Milliseconds()))
		a.logRequest(r, lrw.statusCode, duration)
	})
}

func (a *App) logRequest(r *http.Request, statusCode int, duration time.Duration) {
	if !a.Cfg.LogRequests {
		return
	}
	clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		clientIP = r.RemoteAddr
	}
	a.Logf("%s %s %s -> %d (%dms)", clientIP, r.Method, r.URL.Path, statusCode, duration.Milliseconds())
}
