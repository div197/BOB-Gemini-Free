package server

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"
)

//go:embed playground.html
var playgroundHTML []byte

//go:embed favicon.ico
var faviconICO []byte

// These assets are intentionally embedded separately from web/ because the
// local gateway's application entry point is /playground rather than /. The
// hosted static bundle keeps its own root-relative PWA contract.
//
//go:embed manifest.json
var localManifestJSON []byte

//go:embed sw.js
var localServiceWorkerJS []byte

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

func (a *App) handleLocalManifest(w http.ResponseWriter, r *http.Request) {
	manifest := append([]byte(nil), localManifestJSON...)
	w.Header().Set("Content-Type", "application/manifest+json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(manifest)
}

func (a *App) handleLocalServiceWorker(w http.ResponseWriter, r *http.Request) {
	version := strings.TrimSpace(a.Version)
	if version == "" {
		version = "dev"
	}
	encodedVersion, err := json.Marshal(version)
	if err != nil {
		http.Error(w, "could not prepare service worker", http.StatusInternalServerError)
		return
	}
	worker := bytes.ReplaceAll(localServiceWorkerJS, []byte("__BOB_CACHE_VERSION__"), encodedVersion)
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Service-Worker-Allowed", "/")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(worker)
}
