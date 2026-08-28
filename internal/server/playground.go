package server

import (
	"bytes"
	_ "embed"
	"net/http"
	"strings"
)

//go:embed playground.html
var playgroundHTML []byte

//go:embed favicon.ico
var faviconICO []byte

func (a *App) handlePlayground(w http.ResponseWriter, r *http.Request) {
	version := strings.TrimSpace(a.Version)
	if version == "" {
		version = "dev"
	}
	html := bytes.ReplaceAll(playgroundHTML, []byte("__BOB_DESKTOP_VERSION__"), []byte(version))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("Content-Security-Policy", "default-src 'self' 'unsafe-inline' 'unsafe-eval' data: blob: https:; frame-src 'self' data: blob:; connect-src *;")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(html)
}

func (a *App) handleFavicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(faviconICO)
}
