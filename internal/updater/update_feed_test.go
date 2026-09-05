package updater

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestSignAndVerifyUpdateFeed(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate feed key: %v", err)
	}
	feed := []byte("{\"schema\":1}\n")
	signature, err := SignUpdateFeed(feed, privateKey)
	if err != nil {
		t.Fatalf("SignUpdateFeed: %v", err)
	}
	if err := verifySignedUpdateFeed(publicKey, feed, signature); err != nil {
		t.Fatalf("verifySignedUpdateFeed: %v", err)
	}

	tampered := append([]byte(nil), feed...)
	tampered[1] = 'x'
	if err := verifySignedUpdateFeed(publicKey, tampered, signature); err == nil {
		t.Fatal("tampered update feed was accepted")
	}
	if err := verifySignedUpdateFeed(publicKey, feed, []byte(base64.StdEncoding.EncodeToString(make([]byte, ed25519.SignatureSize)))); err == nil {
		t.Fatal("invalid update feed signature was accepted")
	}
}

func TestSignedDesktopFeedSelectsLatestPreviewWithoutGitHubAPI(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate feed key: %v", err)
	}
	feed := DesktopUpdateFeed{
		Schema:      desktopUpdateFeedSchema,
		GeneratedAt: time.Date(2026, 9, 5, 4, 45, 0, 0, time.UTC),
		ExpiresAt:   time.Date(2026, 9, 12, 4, 45, 0, 0, time.UTC),
		Stable: &GitHubRelease{
			TagName:     "v0.1.5",
			HTMLURL:     "https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.1.5",
			PublishedAt: time.Date(2026, 8, 20, 3, 11, 36, 0, time.UTC),
		},
		Previews: []GitHubRelease{
			{TagName: "v0.2.0-preview.8", Prerelease: true, HTMLURL: "https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.8"},
			{
				TagName:    "v0.2.0-preview.9",
				Prerelease: true,
				HTMLURL:    "https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.9",
				Assets: []ReleaseAsset{
					{Name: "bob-gemini-free-macos-universal.zip", BrowserDownloadURL: "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.2.0-preview.9/bob-gemini-free-macos-universal.zip", Size: 19017019},
					{Name: "SHA256SUMS", BrowserDownloadURL: "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.2.0-preview.9/SHA256SUMS"},
					{Name: "SHA256SUMS.sig", BrowserDownloadURL: "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.2.0-preview.9/SHA256SUMS.sig"},
				},
			},
		},
	}
	feedBytes, err := json.Marshal(feed)
	if err != nil {
		t.Fatalf("marshal update feed: %v", err)
	}
	signature, err := SignUpdateFeed(feedBytes, privateKey)
	if err != nil {
		t.Fatalf("sign update feed: %v", err)
	}

	var feedRequests, apiRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/div197/BOB-Gemini-Free/main/updates/desktop-feed.json":
			feedRequests.Add(1)
			_, _ = w.Write(feedBytes)
		case "/div197/BOB-Gemini-Free/main/updates/desktop-feed.json.sig":
			feedRequests.Add(1)
			_, _ = w.Write(signature)
		default:
			apiRequests.Add(1)
			t.Errorf("signed feed path fell through to API: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	previousKey := BuildUpdatePublicKey
	BuildUpdatePublicKey = base64.StdEncoding.EncodeToString(publicKey)
	t.Cleanup(func() { BuildUpdatePublicKey = previousKey })
	client := &http.Client{Transport: rewriteOfficialUpdateTransport{base: server.Client().Transport, destination: server.URL}}
	result, err := checkLatestDesktopForChannelWithClientContext(context.Background(), client, "v0.2.0-preview.8", DesktopChannelPreview, "darwin", "arm64")
	if err != nil {
		t.Fatalf("check signed desktop feed: %v", err)
	}
	if result.LatestVersion != "v0.2.0-preview.9" || result.Channel != DesktopChannelPreview || !result.HasUpdate || !result.AssetAvailable || !result.ManifestAvailable {
		t.Fatalf("signed feed result = %#v", result)
	}
	if feedRequests.Load() != 2 {
		t.Fatalf("feed requests = %d, want feed and signature only", feedRequests.Load())
	}
	if apiRequests.Load() != 0 {
		t.Fatalf("API requests = %d, want 0 when signed feed is valid", apiRequests.Load())
	}
}

func TestDesktopDiscoveryFallsBackToAPIWhenFeedSignatureFails(t *testing.T) {
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate feed key: %v", err)
	}
	var feedRequests, apiRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/div197/BOB-Gemini-Free/main/updates/desktop-feed.json":
			feedRequests.Add(1)
			_, _ = w.Write([]byte(`{"schema":1}`))
		case "/div197/BOB-Gemini-Free/main/updates/desktop-feed.json.sig":
			feedRequests.Add(1)
			_, _ = w.Write([]byte("invalid"))
		case "/repos/div197/bob-gemini-free/releases/latest":
			apiRequests.Add(1)
			_ = json.NewEncoder(w).Encode(GitHubRelease{TagName: "v0.1.5", HTMLURL: "https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.1.5"})
		case "/repos/div197/bob-gemini-free/releases":
			apiRequests.Add(1)
			_ = json.NewEncoder(w).Encode([]GitHubRelease{{
				TagName:    "v0.2.0-preview.9",
				Prerelease: true,
				HTMLURL:    "https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.9",
				Assets: []ReleaseAsset{
					{Name: "bob-gemini-free-macos-universal.zip", BrowserDownloadURL: "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.2.0-preview.9/bob-gemini-free-macos-universal.zip", Size: 19017019},
					{Name: "SHA256SUMS", BrowserDownloadURL: "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.2.0-preview.9/SHA256SUMS"},
					{Name: "SHA256SUMS.sig", BrowserDownloadURL: "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.2.0-preview.9/SHA256SUMS.sig"},
				},
			}})
		default:
			t.Errorf("unexpected discovery request: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	previousKey := BuildUpdatePublicKey
	BuildUpdatePublicKey = base64.StdEncoding.EncodeToString(publicKey)
	t.Cleanup(func() { BuildUpdatePublicKey = previousKey })
	client := &http.Client{Transport: rewriteOfficialUpdateTransport{base: server.Client().Transport, destination: server.URL}}
	result, err := checkLatestDesktopForChannelWithClientContext(context.Background(), client, "v0.2.0-preview.8", DesktopChannelPreview, "darwin", "arm64")
	if err != nil {
		t.Fatalf("API fallback discovery: %v", err)
	}
	if result.LatestVersion != "v0.2.0-preview.9" || !result.HasUpdate {
		t.Fatalf("API fallback result = %#v", result)
	}
	if feedRequests.Load() != 2 || apiRequests.Load() != 2 {
		t.Fatalf("requests = feed:%d api:%d, want feed:2 api:2", feedRequests.Load(), apiRequests.Load())
	}
}

func TestExplicitDesktopDiscoveryBypassesSignedFeedForFreshness(t *testing.T) {
	var apiRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/div197/BOB-Gemini-Free/main/updates/") {
			t.Errorf("fresh explicit check unexpectedly requested signed feed: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		apiRequests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/repos/div197/bob-gemini-free/releases/latest":
			_ = json.NewEncoder(w).Encode(GitHubRelease{TagName: "v0.1.5", HTMLURL: "https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.1.5"})
		case "/repos/div197/bob-gemini-free/releases":
			_ = json.NewEncoder(w).Encode([]GitHubRelease{{TagName: "v0.2.0-preview.9", Prerelease: true, HTMLURL: "https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.9"}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := &http.Client{Transport: rewriteOfficialUpdateTransport{base: server.Client().Transport, destination: server.URL}}
	result, err := checkLatestDesktopForChannelFreshWithClientContext(context.Background(), client, "v0.2.0-preview.8", DesktopChannelPreview, "darwin", "arm64")
	if err != nil {
		t.Fatalf("fresh desktop discovery: %v", err)
	}
	if result.LatestVersion != "v0.2.0-preview.9" || !result.HasUpdate {
		t.Fatalf("fresh discovery result = %#v", result)
	}
	if apiRequests.Load() != 2 {
		t.Fatalf("fresh discovery API requests = %d, want stable and preview", apiRequests.Load())
	}
}

func TestValidateDesktopUpdateFeedRejectsExpiredAndLongValidity(t *testing.T) {
	base := DesktopUpdateFeed{
		Schema:      desktopUpdateFeedSchema,
		GeneratedAt: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		ExpiresAt:   time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC),
		Stable:      &GitHubRelease{TagName: "v0.1.5"},
		Previews:    []GitHubRelease{{TagName: "v0.2.0-preview.9", Prerelease: true}},
	}
	now := time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC)
	if err := validateDesktopUpdateFeed(&base, now); err != nil {
		t.Fatalf("valid update feed rejected: %v", err)
	}
	expired := base
	expired.ExpiresAt = now.Add(-time.Second)
	if err := validateDesktopUpdateFeed(&expired, now); err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expired feed result = %v, want expiry rejection", err)
	}
	longLived := base
	longLived.ExpiresAt = longLived.GeneratedAt.Add(desktopUpdateFeedMaxLifetime + time.Second)
	if err := validateDesktopUpdateFeed(&longLived, now); err == nil || !strings.Contains(err.Error(), "too long") {
		t.Fatalf("long-lived feed result = %v, want validity-window rejection", err)
	}
}

func TestOfficialUpdateFeedURLIsExactAndPinned(t *testing.T) {
	for _, test := range []struct {
		name string
		url  string
		want bool
	}{
		{name: "feed", url: DesktopUpdateFeedURL, want: true},
		{name: "signature", url: DesktopUpdateFeedSignatureURL, want: true},
		{name: "query", url: DesktopUpdateFeedURL + "?raw=1", want: false},
		{name: "other repository", url: "https://raw.githubusercontent.com/other/repo/main/updates/desktop-feed.json", want: false},
		{name: "github api", url: DesktopReleaseAPIURL, want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := isOfficialUpdateFeedURL(test.url); got != test.want {
				t.Fatalf("isOfficialUpdateFeedURL(%q) = %v, want %v", test.url, got, test.want)
			}
		})
	}
}

func TestCheckedInDesktopUpdateFeedHasDetachedSignature(t *testing.T) {
	feedPath := filepath.Join("..", "..", "updates", "desktop-feed.json")
	signaturePath := feedPath + ".sig"
	feedBytes, err := os.ReadFile(feedPath)
	if err != nil {
		t.Fatalf("read checked-in update feed: %v", err)
	}
	signature, err := os.ReadFile(signaturePath)
	if err != nil {
		t.Fatalf("read checked-in update feed signature: %v", err)
	}
	publicKey, err := decodeEncodedBytes("ba7854781bca2a14da4f1ec5e931ff45f458ac9377c42ac127c349e5ecad2dff")
	if err != nil {
		t.Fatalf("decode checked-in update public key: %v", err)
	}
	if err := verifySignedUpdateFeed(ed25519.PublicKey(publicKey), feedBytes, signature); err != nil {
		t.Fatalf("checked-in update feed signature: %v", err)
	}
	var feed DesktopUpdateFeed
	if err := json.Unmarshal(feedBytes, &feed); err != nil {
		t.Fatalf("decode checked-in update feed: %v", err)
	}
	if err := validateDesktopUpdateFeed(&feed, feed.GeneratedAt.Add(time.Hour)); err != nil {
		t.Fatalf("validate checked-in update feed: %v", err)
	}
}
