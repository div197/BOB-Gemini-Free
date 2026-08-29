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
		header.Set("X-Goog-Upload-URL", "https://content-push.googleapis.com/upload/fixture")
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

type scriptedUploadClient struct {
	mu        sync.Mutex
	responses []*http.Response
	requests  []uploadFixtureRequest
}

func (c *scriptedUploadClient) Do(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.requests = append(c.requests, uploadFixtureRequest{method: req.Method, url: req.URL.String(), header: req.Header.Clone(), body: body})
	if len(c.responses) == 0 {
		c.mu.Unlock()
		return nil, io.ErrUnexpectedEOF
	}
	response := c.responses[0]
	c.responses = c.responses[1:]
	c.mu.Unlock()
	return response, nil
}

func uploadStartResponse(uploadURL string, status int) *http.Response {
	header := make(http.Header)
	if uploadURL != "" {
		header.Set("X-Goog-Upload-URL", uploadURL)
	}
	return &http.Response{StatusCode: status, Header: header, Body: io.NopCloser(bytes.NewReader(nil))}
}

func uploadFinishResponse(body []byte, status int) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(bytes.NewReader(body))}
}

func TestUploadImageRejectsUntrustedUploadURL(t *testing.T) {
	client := &scriptedUploadClient{responses: []*http.Response{
		uploadStartResponse("https://attacker.example.test/upload/fixture", http.StatusOK),
	}}
	_, err := UploadImage(client, PageTokens{}, createTestImage(2, 2), "image/png", nil, "")
	if err == nil {
		t.Fatal("untrusted upload URL was accepted")
	}
	client.mu.Lock()
	requestCount := len(client.requests)
	client.mu.Unlock()
	if requestCount != 1 {
		t.Fatalf("upload request count = %d, want 1 after URL validation", requestCount)
	}
}

func TestUploadImageRejectsNonSuccessResponsesAndOversizedReferences(t *testing.T) {
	tests := []struct {
		name      string
		responses []*http.Response
		wantCalls int
	}{
		{
			name:      "start status",
			responses: []*http.Response{uploadStartResponse("https://content-push.googleapis.com/upload/fixture", http.StatusTooManyRequests)},
			wantCalls: 1,
		},
		{
			name: "finish body",
			responses: []*http.Response{
				uploadStartResponse("https://content-push.googleapis.com/upload/fixture", http.StatusOK),
				uploadFinishResponse(bytes.Repeat([]byte("x"), MaxUploadResponseBytes+1), http.StatusOK),
			},
			wantCalls: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &scriptedUploadClient{responses: test.responses}
			_, err := UploadImage(client, PageTokens{}, createTestImage(2, 2), "image/png", nil, "")
			if err == nil {
				t.Fatal("unsafe upload response was accepted")
			}
			client.mu.Lock()
			calls := len(client.requests)
			client.mu.Unlock()
			if calls != test.wantCalls {
				t.Fatalf("upload request count = %d, want %d", calls, test.wantCalls)
			}
		})
	}
}

func TestGoldenUploadFixtureCoversResumableProtocol(t *testing.T) {
	client := &uploadFixtureClient{}
	imageBytes := createTestImage(2, 2)
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
