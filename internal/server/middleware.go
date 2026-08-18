package server

import (
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
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

func authorize(r *http.Request, apiKeys []string) bool {
	if len(apiKeys) == 0 {
		return true
	}

	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		token := authHeader[7:]
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

	if queryKey := r.URL.Query().Get("key"); queryKey != "" {
		for _, key := range apiKeys {
			if subtle.ConstantTimeCompare([]byte(queryKey), []byte(key)) == 1 {
				return true
			}
		}
	}

	return false
}

func (a *App) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept, Origin, User-Agent, x-api-key, anthropic-version, anthropic-beta, x-goog-api-key, x-goog-api-client, x-client-request-id, *")
		w.Header().Set("Access-Control-Expose-Headers", "x-request-id, openai-processing-ms, openai-version, x-ratelimit-limit-requests, x-ratelimit-remaining-requests, x-ratelimit-reset-requests, content-length, *")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

const maxRequestBodySize = 32 << 20 // 32 MB limit

func (a *App) withAuthAndLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
		}
		start := time.Now()
		lrw := newLoggingResponseWriter(w)

		// Set standard request metadata and rate limit headers
		reqID := r.Header.Get("X-Client-Request-Id")
		if reqID == "" {
			reqID = fmt.Sprintf("req_%s", format.RandHex(16))
		}
		w.Header().Set("x-request-id", reqID)
		w.Header().Set("x-powered-by", "BOB-Gemini-Free / ABCsteps (div197)")
		w.Header().Set("openai-version", "2020-10-01")
		w.Header().Set("x-ratelimit-limit-requests", "1000")
		w.Header().Set("x-ratelimit-remaining-requests", "999")
		w.Header().Set("x-ratelimit-reset-requests", "1s")

		if len(a.Cfg.APIKeys) > 0 && !authorize(r, a.Cfg.APIKeys) {
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
