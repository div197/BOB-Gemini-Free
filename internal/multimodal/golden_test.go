package multimodal

import (
	"bytes"
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/div197/bob-gemini-free/internal/gemini"
)

type uploadFixtureRequest struct {
	method string
	url    string
	header http.Header
	body   []byte
}

type uploadFixtureClient struct {
	mu       sync.Mutex
	requests []uploadFixtureRequest
}

func (c *uploadFixtureClient) Do(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.requests = append(c.requests, uploadFixtureRequest{
		method: req.Method,
		url:    req.URL.String(),
		header: req.Header.Clone(),
		body:   append([]byte(nil), body...),
	})
	call := len(c.requests)
	c.mu.Unlock()

	if call == 1 {
		header := make(http.Header)
		header.Set("X-Goog-Upload-URL", "https://upload.example.test/fixture")
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     header,
			Body:       io.NopCloser(bytes.NewReader(nil)),
		}, nil
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString("/file_refs/fixture-image")),
	}, nil
}

func TestGoldenUploadFixtureCoversResumableProtocol(t *testing.T) {
	client := &uploadFixtureClient{}
	imageBytes := []byte("fixture-image-bytes")
	ref, err := UploadImage(client, PageTokens{PushID: "fixture-push", Pctx: "fixture-pctx"}, imageBytes, "image/png", gemini.NewCookieCache(""), "1")
	if err != nil {
		t.Fatalf("UploadImage: %v", err)
	}
	if ref != "/file_refs/fixture-image" {
		t.Fatalf("file reference = %q", ref)
	}

	client.mu.Lock()
	requests := append([]uploadFixtureRequest(nil), client.requests...)
	client.mu.Unlock()
	if len(requests) != 2 {
		t.Fatalf("upload requests = %d, want 2", len(requests))
	}
	if requests[0].method != http.MethodPost || requests[0].url != "https://content-push.googleapis.com/upload/" {
		t.Fatalf("start request = %#v", requests[0])
	}
	if got := requests[0].header.Get("X-Goog-Upload-Protocol"); got != "resumable" {
		t.Errorf("upload protocol = %q", got)
	}
	if got := requests[0].header.Get("X-Goog-Upload-Command"); got != "start" {
		t.Errorf("start command = %q", got)
	}
	if got := requests[0].header.Get("X-Goog-AuthUser"); got != "1" {
		t.Errorf("auth user = %q", got)
	}
	if len(requests[0].body) != 0 {
		t.Errorf("start body length = %d", len(requests[0].body))
	}
	if !bytes.Equal(requests[1].body, imageBytes) {
		t.Errorf("uploaded bytes = %q, want %q", requests[1].body, imageBytes)
	}
	if got := requests[1].header.Get("X-Goog-Upload-Command"); got != "upload, finalize" {
		t.Errorf("finalize command = %q", got)
	}
}
