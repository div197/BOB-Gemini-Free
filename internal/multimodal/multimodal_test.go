package multimodal

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/gemini"
)

type contextCheckingRequester struct{}

func (contextCheckingRequester) Do(req *http.Request) (*http.Response, error) {
	if err := req.Context().Err(); err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"image/png"}},
		Body:       io.NopCloser(bytes.NewReader(createTestImage(2, 2))),
	}, nil
}

func createTestImage(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 255), G: uint8(y % 255), B: 100, A: 255})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func TestCompressImageBytesIfNeeded(t *testing.T) {
	// Small image should not be compressed
	smallImg := createTestImage(50, 50)
	outBytes, mime, err := CompressImageBytesIfNeeded(smallImg, "image/png", 500000)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(outBytes) != len(smallImg) {
		t.Errorf("Expected small image to remain uncompressed, len=%d, got=%d", len(smallImg), len(outBytes))
	}
	if mime != "image/png" {
		t.Errorf("Expected mime image/png, got %s", mime)
	}

	// Large image (> MaxImageDimension) should be downscaled to MaxImageDimension
	largeImg := createTestImage(2000, 1500)
	outBytes2, mime2, err := CompressImageBytesIfNeeded(largeImg, "image/png", 1000)
	if err != nil {
		t.Fatalf("Unexpected error on large image: %v", err)
	}
	if mime2 != "image/jpeg" {
		t.Errorf("Expected mime image/jpeg, got %s", mime2)
	}

	decodedImg, _, err := image.Decode(bytes.NewReader(outBytes2))
	if err != nil {
		t.Fatalf("Failed to decode compressed image: %v", err)
	}
	if decodedImg.Bounds().Dx() > MaxImageDimension || decodedImg.Bounds().Dy() > MaxImageDimension {
		t.Errorf("Expected dimensions <= %d, got %dx%d", MaxImageDimension, decodedImg.Bounds().Dx(), decodedImg.Bounds().Dy())
	}
}

func TestCompressImageBytesGIF(t *testing.T) {
	gifImg := createTestImage(80, 80)
	outBytes, mime, err := CompressImageBytesIfNeeded(gifImg, "image/gif", 50)
	if err != nil {
		t.Fatalf("Failed to compress GIF image: %v", err)
	}
	if mime != "image/jpeg" {
		t.Errorf("Expected mime image/jpeg for re-encoded GIF, got %s", mime)
	}
	if len(outBytes) == 0 {
		t.Errorf("Expected non-empty output bytes")
	}
}

func TestCompressIfNeededBase64(t *testing.T) {
	largeImg := createTestImage(2000, 1500)
	b64 := base64.StdEncoding.EncodeToString(largeImg)

	compressedB64, err := CompressIfNeeded(b64, 1000)
	if err != nil {
		t.Fatalf("CompressIfNeeded error: %v", err)
	}

	dec, err := base64.StdEncoding.DecodeString(compressedB64)
	if err != nil {
		t.Fatalf("Failed to decode base64 output: %v", err)
	}
	decodedImg, _, err := image.Decode(bytes.NewReader(dec))
	if err != nil {
		t.Fatalf("Failed to decode compressed image from base64: %v", err)
	}
	if decodedImg.Bounds().Dx() > MaxImageDimension || decodedImg.Bounds().Dy() > MaxImageDimension {
		t.Errorf("Expected dimensions <= %d, got %dx%d", MaxImageDimension, decodedImg.Bounds().Dx(), decodedImg.Bounds().Dy())
	}
}

func TestImageInputBounds(t *testing.T) {
	if _, _, err := CompressImageBytesIfNeeded([]byte("not-an-image"), "image/png", MaxImageByteSize); err == nil {
		t.Fatal("invalid image bytes were accepted")
	}
	if _, _, err := CompressImageBytesIfNeeded(make([]byte, MaxImageInputBytes+1), "image/png", MaxImageByteSize); err == nil {
		t.Fatal("oversized image input was accepted")
	}
}

func TestFetchImageBytes(t *testing.T) {
	_, err := FetchImageBytes(nil, "file:///etc/passwd")
	if err == nil {
		t.Error("Expected error for file:// scheme, got nil")
	}

	imageBytes := createTestImage(4, 4)
	fixtureClient := imageFixtureRequester{body: imageBytes}
	b, err := fetchImageBytes(fixtureClient, "https://images.example.test/image.png", func(string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("203.0.113.10")}, nil
	})
	if err != nil {
		t.Fatalf("Expected image fetch to succeed, got %v", err)
	}
	if len(b) == 0 {
		t.Fatalf("Expected non-empty image bytes")
	}
}

func TestFetchImageBytesRejectsNilClientAndResponse(t *testing.T) {
	_, err := fetchImageBytes(nil, "https://images.example.test/image.png", func(string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("203.0.113.10")}, nil
	})
	if err == nil || !strings.Contains(err.Error(), "HTTP client") {
		t.Fatalf("nil client error = %v", err)
	}

	_, err = fetchImageBytes(emptyResponseRequester{}, "https://images.example.test/image.png", func(string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("203.0.113.10")}, nil
	})
	if err == nil || !strings.Contains(err.Error(), "empty response") {
		t.Fatalf("nil response error = %v", err)
	}
}

func TestFetchImageBytesRejectsUnguardedRequester(t *testing.T) {
	_, err := FetchImageBytes(imageFixtureRequester{body: createTestImage(2, 2)}, "https://images.example.test/image.png")
	if err == nil || !strings.Contains(err.Error(), "guardable HTTP client") {
		t.Fatalf("unguarded requester error = %v", err)
	}
}

type emptyResponseRequester struct{}

func (emptyResponseRequester) Do(*http.Request) (*http.Response, error) { return nil, nil }

type imageFixtureRequester struct {
	body []byte
}

func (f imageFixtureRequester) Do(_ *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"image/png"}},
		Body:       io.NopCloser(bytes.NewReader(f.body)),
	}, nil
}

func TestFetchImageBytesRejectsHTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "missing", http.StatusNotFound)
	}))
	defer ts.Close()

	_, err := FetchImageBytes(ts.Client(), ts.URL)
	if err == nil {
		t.Fatalf("Expected error for non-2xx image fetch")
	}
}

func TestFetchImageBytesContextHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := fetchImageBytesContext(ctx, contextCheckingRequester{}, "https://images.example.test/image.png", func(string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("203.0.113.10")}, nil
	})
	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("fetch error = %v, want context cancellation", err)
	}
}

func TestFetchImageBytesRejectsPrivateAndLocalHosts(t *testing.T) {
	for _, imageURL := range []string{
		"http://127.0.0.1/image.png",
		"http://169.254.169.254/latest/meta-data",
		"http://localhost/image.png",
	} {
		if _, err := FetchImageBytes(nil, imageURL); err == nil {
			t.Errorf("expected private/local URL %q to be rejected", imageURL)
		}
	}
}

func TestFetchImageBytesRejectsPrivateDNSResolution(t *testing.T) {
	_, err := fetchImageBytes(nil, "https://images.example.test/image.png", func(string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("10.0.0.7")}, nil
	})
	if err == nil {
		t.Fatal("private DNS result was accepted")
	}
}

func TestFetchImageBytesRevalidatesDNSBeforeDial(t *testing.T) {
	calls := 0
	client := &http.Client{Transport: &http.Transport{
		DialContext: func(context.Context, string, string) (net.Conn, error) {
			t.Fatal("guarded image dial should reject the private second DNS answer before opening a socket")
			return nil, nil
		},
	}}
	_, err := fetchImageBytes(client, "https://images.example.test/image.png", func(string) ([]net.IP, error) {
		calls++
		if calls == 1 {
			return []net.IP{net.ParseIP("203.0.113.10")}, nil
		}
		return []net.IP{net.ParseIP("10.0.0.7")}, nil
	})
	if err == nil || !strings.Contains(err.Error(), "private or local") {
		t.Fatalf("DNS rebinding result = %v, want private-address rejection", err)
	}
	if calls != 2 {
		t.Fatalf("resolver calls = %d, want validation plus final dial lookup", calls)
	}
}

func TestFetchImageBytesDoesNotFollowCrossHostRedirect(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(createTestImage(2, 2))
	}))
	defer target.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	defer redirect.Close()

	_, err := fetchImageBytes(redirect.Client(), redirect.URL, func(string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("203.0.113.10")}, nil
	})
	if err == nil {
		t.Fatal("cross-host redirect was accepted")
	}
}

func TestFetchImageBytesRejectsNonImageContent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html>not an image</html>"))
	}))
	defer ts.Close()

	_, err := FetchImageBytes(ts.Client(), ts.URL)
	if err == nil {
		t.Fatalf("Expected error for non-image fetch response")
	}
}

func TestFetchImageBytesRejectsSpoofedImageHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("<html>not actually an image</html>"))
	}))
	defer ts.Close()

	_, err := FetchImageBytes(ts.Client(), ts.URL)
	if err == nil {
		t.Fatalf("Expected error for spoofed image content type")
	}
}

func TestTokenCache(t *testing.T) {
	cfg := config.Default()
	cookie := gemini.NewCookieCache(cfg.CookieFile)
	client := &http.Client{Timeout: 5 * time.Second}

	tc := NewTokenCache(cfg, cookie, client)
	if tc == nil {
		t.Fatalf("Expected non-nil TokenCache")
	}

	// Keep the core suite hermetic. The live page-token path is covered by
	// TestLiveImageUpload when a caller explicitly provides a cookie/session.
	tc.mu.Lock()
	tc.ts = time.Now()
	tc.mu.Unlock()
	tokens := tc.Get()
	if tokens.PushID == "" || tokens.Pctx == "" {
		t.Errorf("Expected valid PushID and Pctx tokens, got: %+v", tokens)
	}
}

func TestLiveImageUpload(t *testing.T) {
	cfg := config.Default()
	cookieCache := gemini.NewCookieCache("../../cookie.txt")
	cinfo, err := cookieCache.Load()
	if err != nil || cinfo.Cookie == "" {
		t.Skip("cookie.txt not available, skipping live upload test")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	tokenCache := NewTokenCache(cfg, cookieCache, client)
	tokens := tokenCache.Get()

	sampleImg := createTestImage(100, 100)
	ref, err := UploadImage(client, tokens, sampleImg, "image/png", cookieCache, cfg.AuthUser)
	if err != nil {
		t.Logf("UploadImage failed: %v", err)
	} else {
		t.Logf("UploadImage succeeded: ref=%s", ref)
	}

	gemClient := gemini.NewClient(cfg)
	gemClient.Cookies = cookieCache
	err = gemClient.GenerateStream("What is in this image?", 1, 4, []string{ref}, nil, func(delta string) error {
		t.Logf("Delta: %s", delta)
		return nil
	})
	if err != nil {
		t.Logf("GenerateStream with image ref failed: %v", err)
	} else {
		t.Logf("GenerateStream with image ref SUCCEEDED!")
	}
}
