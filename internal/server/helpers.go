package server

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/div197/bob-gemini-free/internal/format"
	"github.com/div197/bob-gemini-free/internal/gemini"
	"github.com/div197/bob-gemini-free/internal/geminiapi"
	"github.com/div197/bob-gemini-free/internal/multimodal"
)

// ErrorToStatusCode maps an upstream error to an appropriate HTTP status code.
// Default is 502 Bad Gateway if the error isn't a recognized UpstreamError.
func ErrorToStatusCode(err error) int {
	var upErr *gemini.UpstreamError
	if errors.As(err, &upErr) && upErr.Status > 0 {
		return upErr.Status
	}
	var apiErr *geminiapi.APIError
	if errors.As(err, &apiErr) {
		if apiErr.Status > 0 {
			return apiErr.Status
		}
		if apiErr.Kind == "request" {
			return http.StatusBadRequest
		}
	}
	return http.StatusBadGateway
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(data)
}

func writeRawJSON(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}

var bufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

const maxPooledSSEBufferBytes = 1 << 20

var errRequestBodyTooLarge = errors.New("request body exceeds the 32 MB limit")

// readRequestBody keeps the endpoint handlers safe even when they are called
// directly in an embedding or through a future mux that does not install
// withAuthAndLogging. The middleware applies the same limit to normal HTTP
// traffic, while this second boundary prevents an unbounded io.ReadAll at the
// handler seam.
func readRequestBody(r *http.Request) ([]byte, error) {
	if r == nil || r.Body == nil {
		return nil, errors.New("request body is empty")
	}
	if r.ContentLength > maxRequestBodySize {
		return nil, errRequestBodyTooLarge
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBodySize+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxRequestBodySize {
		return nil, errRequestBodyTooLarge
	}
	return body, nil
}

func startSSE(w http.ResponseWriter) bool {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if ok {
		flusher.Flush()
	}
	return ok
}

func writeSSEData(w http.ResponseWriter, data any) error {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer func() {
		if buf.Cap() <= maxPooledSSEBufferBytes {
			bufPool.Put(buf)
		}
	}()

	buf.WriteString("data: ")
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(data); err != nil {
		return err
	}
	buf.WriteByte('\n') // json.Encode adds one \n. We need two for SSE

	_, err := w.Write(buf.Bytes())
	if err != nil {
		return err
	}

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}

func writeSSEEvent(w http.ResponseWriter, event string, data any) error {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer func() {
		if buf.Cap() <= maxPooledSSEBufferBytes {
			bufPool.Put(buf)
		}
	}()

	buf.WriteString("event: ")
	buf.WriteString(event)
	buf.WriteString("\ndata: ")

	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(data); err != nil {
		return err
	}
	buf.WriteByte('\n')

	_, err := w.Write(buf.Bytes())
	if err != nil {
		return err
	}

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}

func writeSSEComment(w http.ResponseWriter, comment string) error {
	_, err := fmt.Fprintf(w, ": %s\n\n", comment)
	if err != nil {
		return err
	}
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}

// StreamWithKeepAlive periodically flushes standard SSE comment heartbeats (`: keepalive\n\n`)
// while waiting for upstream reasoning or generation tokens, preventing classroom client timeouts
// during lengthy code generation (e.g. 2D CyberSnake games).
func StreamWithKeepAlive(ctx context.Context, w http.ResponseWriter, interval time.Duration, runStream func(emit func(string) error) error, emitDelta func(string) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if runStream == nil {
		return errors.New("stream runner is nil")
	}
	if emitDelta == nil {
		return errors.New("stream emitter is nil")
	}
	if interval <= 0 {
		interval = 2500 * time.Millisecond
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	stopTicker := make(chan struct{})
	tickerDone := make(chan struct{})

	var mu sync.Mutex
	lastWrite := time.Now()

	go func() {
		defer close(tickerDone)
		for {
			select {
			case <-stopTicker:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				mu.Lock()
				if time.Since(lastWrite) >= interval-200*time.Millisecond {
					_ = writeSSEComment(w, "keepalive")
					lastWrite = time.Now()
				}
				mu.Unlock()
			}
		}
	}()

	err := runStream(func(delta string) error {
		mu.Lock()
		defer mu.Unlock()
		lastWrite = time.Now()
		return emitDelta(delta)
	})

	close(stopTicker)
	<-tickerDone
	return err
}

func writeSSEDone(w http.ResponseWriter) error {
	_, err := w.Write([]byte("data: [DONE]\n\n"))
	if err != nil {
		return err
	}
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}

func (a *App) addEstimatedTokens(count uint64) {
	a.TokensProcessed.Add(count)
	if a.Metrics != nil {
		a.Metrics.TokensEstimated.Add(count)
	}
}

func (a *App) uploadImages(images []format.Image) ([]string, error) {
	return a.uploadImagesContext(context.Background(), images)
}

func (a *App) uploadImagesContext(ctx context.Context, images []format.Image) ([]string, error) {
	if len(images) == 0 {
		return nil, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var fileRefs []string
	var tokens multimodal.PageTokens
	var haveTokens bool

	var requester gemini.Requester = a.Gem.HTTP
	if requester == nil {
		if a.HTTPClient != nil {
			requester = a.HTTPClient
		} else {
			requester = createHTTPClient(a.Cfg)
		}
	}

	cacheScope, cacheEnabled := a.authenticatedImageCacheScope()
	for _, img := range images {
		data := img.Data
		mime := img.MIME
		if len(data) == 0 && img.URL != "" {
			fetched, err := multimodal.FetchImageBytesContext(ctx, requester, img.URL)
			if err != nil {
				return nil, fmt.Errorf("image fetch failed for %s: %w", img.URL, err)
			}
			data = fetched
			mime = http.DetectContentType(data)
		}
		if len(data) == 0 {
			continue
		}

		// Fast path: Check image cache to avoid redundant uploads to Scotty
		hashStr := imageCacheKey(data, cacheScope)
		if cacheEnabled {
			if cachedRef, ok := a.ImageCache.Load(hashStr); ok {
				if a.Metrics != nil {
					a.Metrics.ImageCacheHits.Add(1)
				}
				fileRefs = append(fileRefs, cachedRef)
				continue
			}
			if a.Metrics != nil {
				a.Metrics.ImageCacheMisses.Add(1)
			}
		}

		if !haveTokens {
			tokens = a.Tokens.Get()
			haveTokens = true
		}

		ref, err := multimodal.UploadImageContext(ctx, requester, tokens, data, mime, a.Gem.Cookies, a.Cfg.AuthUser)
		if err != nil {
			return nil, fmt.Errorf("image upload failed: %w", err)
		}
		if ref != "" {
			if a.Metrics != nil {
				a.Metrics.ImageUploads.Add(1)
			}
			if cacheEnabled {
				a.ImageCache.Store(hashStr, ref)
			}
			fileRefs = append(fileRefs, ref)
		}
	}

	if len(fileRefs) == 0 {
		return nil, fmt.Errorf("no image attachments could be uploaded")
	}
	return fileRefs, nil
}

func (a *App) authenticatedImageCacheScope() (string, bool) {
	if a == nil || a.Gem == nil || a.Gem.Cookies == nil {
		return "", false
	}
	// A Scotty reference is session-bound. Disable reuse when a pool can select
	// different accounts because the current upload and later generation would
	// otherwise be able to cross account boundaries.
	if len(a.Cfg.CookiePool) > 0 {
		return "", false
	}
	info, err := a.Gem.Cookies.Load()
	if err != nil || info.Cookie == "" {
		return "", false
	}
	scope := sha256.Sum256([]byte(info.Cookie + "\x00" + info.SAPISID + "\x00" + a.Cfg.AuthUser))
	return hex.EncodeToString(scope[:]), true
}

func imageCacheKey(data []byte, scope string) string {
	hash := sha256.New()
	hash.Write(data)
	hash.Write([]byte("\x00"))
	hash.Write([]byte(scope))
	return hex.EncodeToString(hash.Sum(nil))
}
