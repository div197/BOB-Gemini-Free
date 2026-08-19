package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"net/http"

	"github.com/div197/bob-gemini-free/internal/format"
	"github.com/div197/bob-gemini-free/internal/gemini"
	"github.com/div197/bob-gemini-free/internal/multimodal"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(data)
}

var bufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
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
	defer bufPool.Put(buf)

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
	defer bufPool.Put(buf)

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

func (a *App) uploadImages(images []format.Image) []string {
	if len(images) == 0 {
		return nil
	}

	tokens := a.Tokens.Get()
	var fileRefs []string

	var requester gemini.Requester = a.Gem.HTTP
	if requester == nil {
		if a.HTTPClient != nil {
			requester = a.HTTPClient
		} else {
			requester = createHTTPClient(a.Cfg)
		}
	}

	for _, img := range images {
		data := img.Data
		if len(data) == 0 {
			continue
		}

		// Fast path: Check image cache to avoid redundant uploads to Scotty
		hashBytes := sha256.Sum256(data)
		hashStr := hex.EncodeToString(hashBytes[:])
		if cachedRef, ok := a.ImageCache.Load(hashStr); ok {
			fileRefs = append(fileRefs, cachedRef.(string))
			continue
		}

		ref, err := multimodal.UploadImage(requester, tokens, data, img.MIME, a.Gem.Cookies, a.Cfg.AuthUser)
		if err != nil {
			a.Logf("Image upload failed: %v", err)
			continue
		}
		if ref != "" {
			a.ImageCache.Store(hashStr, ref)
			fileRefs = append(fileRefs, ref)
		}
	}

	if len(fileRefs) == 0 {
		return nil
	}
	return fileRefs
}
