package server

import (
	"crypto/subtle"
	"net"
	"net/http"
	"strings"
	"time"
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
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

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
		a.logRequest(r, lrw.statusCode, time.Since(start))
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
