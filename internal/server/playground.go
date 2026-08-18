package server

import (
	_ "embed"
	"net/http"
)

//go:embed playground.html
var playgroundHTML []byte

func (a *App) handlePlayground(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(playgroundHTML)
}
