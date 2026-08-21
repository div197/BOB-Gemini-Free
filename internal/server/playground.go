package server

import (
	_ "embed"
	"net/http"
)

//go:embed playground.html
var playgroundHTML []byte

//go:embed favicon.ico
var faviconICO []byte

func (a *App) handlePlayground(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(playgroundHTML)
}

func (a *App) handleFavicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(faviconICO)
}
