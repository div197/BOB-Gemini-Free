package multimodal

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/gemini"
)

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

func TestFetchImageBytes(t *testing.T) {
	_, err := FetchImageBytes(nil, "file:///etc/passwd")
	if err == nil {
		t.Error("Expected error for file:// scheme, got nil")
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("fake-png-data"))
	}))
	defer ts.Close()

	b, err := FetchImageBytes(ts.Client(), ts.URL)
	if err != nil || string(b) != "fake-png-data" {
		t.Errorf("Expected fake-png-data, got err=%v, data=%q", err, string(b))
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

	// Without network, should return default fallback tokens
	tokens := tc.Get()
	if tokens.PushID != DefaultPushID || tokens.Pctx != DefaultPctx {
		t.Errorf("Expected fallback default tokens, got: %+v", tokens)
	}
}
