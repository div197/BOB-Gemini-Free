package updater

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestPreviewReleaseAPIIsBounded(t *testing.T) {
	parsed, err := url.Parse(DesktopPreviewReleaseAPIURL)
	if err != nil {
		t.Fatalf("parse preview release API URL: %v", err)
	}
	if got := parsed.Query().Get("per_page"); got != "30" {
		t.Fatalf("preview release API per_page = %q, want 30", got)
	}
}

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
	}{
		{"v0.1.5", "v0.1.6", true},
		{"v0.1.5", "v0.2.0", true},
		{"v0.1.5", "v1.0.0", true},
		{"v0.1.5", "v0.1.5", false},
		{"v0.1.5", "v0.1.4", false},
		{"v0.2.0", "v0.1.9", false},
		{"dev", "v0.1.6", false},
		{"0.1.5", "0.1.6", true},
		{"v0.1.7-preview.3", "v0.1.7-preview.4", true},
		{"v0.1.7-preview.4", "v0.1.7-preview.3", false},
		{"v0.1.7-preview.4", "v0.1.7", true},
		{"v0.1.7-preview.4", "not-a-version", false},
	}

	for _, tt := range tests {
		got := isNewerVersion(tt.current, tt.latest)
		if got != tt.want {
			t.Errorf("isNewerVersion(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
		}
	}
}

func TestIsDesktopVersionCheckable(t *testing.T) {
	for _, test := range []struct {
		version string
		want    bool
	}{
		{version: "v0.2.0", want: true},
		{version: "v0.1.7-preview.7", want: true},
		{version: "dev", want: false},
		{version: "local-ui", want: false},
		{version: "", want: false},
	} {
		if got := IsDesktopVersionCheckable(test.version); got != test.want {
			t.Errorf("IsDesktopVersionCheckable(%q) = %v, want %v", test.version, got, test.want)
		}
	}
}

func TestDesktopUpdateCheckContextCancellationIsHonored(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := checkLatestDesktopChannelContext(ctx, http.DefaultClient, "http://127.0.0.1:1/releases", "v0.1.7", DesktopChannelStable, "darwin", "arm64"); err == nil {
		t.Fatal("canceled desktop update check unexpectedly succeeded")
	}
}

func TestFindMatchingAsset(t *testing.T) {
	assets := []ReleaseAsset{
		{Name: "bob-gemini-free-darwin-arm64", BrowserDownloadURL: "https://example.com/darwin-arm64"},
		{Name: "bob-gemini-free-darwin-amd64", BrowserDownloadURL: "https://example.com/darwin-amd64"},
		{Name: "bob-gemini-free-linux-amd64", BrowserDownloadURL: "https://example.com/linux-amd64"},
		{Name: "bob-gemini-free-windows-amd64.exe", BrowserDownloadURL: "https://example.com/windows-amd64.exe"},
	}

	match := findMatchingAsset(assets, "darwin", "arm64")
	if match == nil || match.BrowserDownloadURL != "https://example.com/darwin-arm64" {
		t.Errorf("expected darwin-arm64 match, got %v", match)
	}

	matchWin := findMatchingAsset(assets, "windows", "amd64")
	if matchWin == nil || matchWin.BrowserDownloadURL != "https://example.com/windows-amd64.exe" {
		t.Errorf("expected windows-amd64 match, got %v", matchWin)
	}
}

func TestFindMatchingAssetRequiresCanonicalName(t *testing.T) {
	lookalike := []ReleaseAsset{{
		Name:               "mirror-bob-gemini-free-darwin-arm64",
		BrowserDownloadURL: "https://example.com/lookalike",
	}}
	if match := findMatchingAsset(lookalike, "darwin", "arm64"); match != nil {
		t.Fatalf("lookalike asset was accepted: %#v", match)
	}

	assets := append(lookalike, ReleaseAsset{
		Name:               "bob-gemini-free-darwin-arm64",
		BrowserDownloadURL: "https://example.com/canonical",
	})
	match := findMatchingAsset(assets, "darwin", "arm64")
	if match == nil || match.BrowserDownloadURL != "https://example.com/canonical" {
		t.Fatalf("canonical asset was not selected: %#v", match)
	}
}

func TestFindMatchingDesktopAssetPrefersUniversalMacArchive(t *testing.T) {
	assets := []ReleaseAsset{
		{Name: "bob-gemini-free-macos-arm64.zip", BrowserDownloadURL: "https://example.com/branded-arm64"},
		{Name: "bob-gemini-free-macos-universal.zip", BrowserDownloadURL: "https://example.com/branded-universal"},
		{Name: "bob-gemini-free-wails-macos-arm64.zip", BrowserDownloadURL: "https://example.com/arm64"},
		{Name: "bob-gemini-free-wails-macos-universal.zip", BrowserDownloadURL: "https://example.com/universal"},
	}

	match := findMatchingDesktopAsset(assets, "darwin", "arm64")
	if match == nil || match.Name != "bob-gemini-free-macos-universal.zip" {
		t.Fatalf("desktop match = %v, want branded universal archive", match)
	}
}

func TestFindMatchingDesktopAssetAcceptsLegacyMigrationName(t *testing.T) {
	assets := []ReleaseAsset{{
		Name:               "bob-gemini-free-wails-windows-amd64.exe",
		BrowserDownloadURL: "https://example.com/legacy.exe",
	}}
	match := findMatchingDesktopAsset(assets, "windows", "amd64")
	if match == nil || match.Name != "bob-gemini-free-wails-windows-amd64.exe" {
		t.Fatalf("desktop match = %v, want legacy migration asset", match)
	}
}

func TestIsOfficialGitHubURLPinsOfficialRepository(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{name: "release asset", raw: "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.2.0/app.zip", want: true},
		{name: "case variation", raw: "https://github.com/DIV197/bob-gemini-free/releases/tag/v0.2.0", want: true},
		{name: "release CDN", raw: "https://objects.githubusercontent.com/github-production-release-asset/1/2/3?x=1", want: true},
		{name: "release asset CDN", raw: "https://release-assets.githubusercontent.com/github-production-release-asset/1/2/3?x=1", want: true},
		{name: "other owner", raw: "https://github.com/another/BOB-Gemini-Free/releases/download/v0.2.0/app.zip", want: false},
		{name: "other repository", raw: "https://github.com/div197/other-repository/releases/download/v0.2.0/app.zip", want: false},
		{name: "wrong scheme", raw: "http://github.com/div197/BOB-Gemini-Free/releases/download/v0.2.0/app.zip", want: false},
		{name: "lookalike host", raw: "https://github.com.evil.example/div197/BOB-Gemini-Free/app.zip", want: false},
		{name: "unexpected port", raw: "https://github.com:444/div197/BOB-Gemini-Free/app.zip", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isOfficialGitHubURL(tt.raw); got != tt.want {
				t.Fatalf("isOfficialGitHubURL(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestUpdateClientRejectsUntrustedRedirect(t *testing.T) {
	var destinationRequests atomic.Int32
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		destinationRequests.Add(1)
		_, _ = w.Write([]byte("unexpected"))
	}))
	defer destination.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, destination.URL, http.StatusFound)
	}))
	defer redirect.Close()

	client := newUpdateHTTPClient(time.Second)
	_, err := client.Get(redirect.URL)
	if err == nil || !strings.Contains(err.Error(), "untrusted host") {
		t.Fatalf("untrusted redirect error = %v", err)
	}
	if got := destinationRequests.Load(); got != 0 {
		t.Fatalf("untrusted redirect reached destination %d times", got)
	}
}

func TestReadBoundedUpdateResponseRejectsNilAndOversizedBodies(t *testing.T) {
	if _, err := readBoundedUpdateResponse(nil, 4); err == nil {
		t.Fatal("nil update response was accepted")
	}
	empty := &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}
	if _, err := readBoundedUpdateResponse(empty, 4); err != nil {
		t.Fatalf("empty bounded response: %v", err)
	}
	oversized := &http.Response{StatusCode: http.StatusOK, ContentLength: 5, Body: http.NoBody}
	if _, err := readBoundedUpdateResponse(oversized, 4); err == nil || !strings.Contains(err.Error(), "exceeded") {
		t.Fatalf("oversized content-length response error = %v", err)
	}
}

func TestCheckLatestDesktopReportsMissingNativePackage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(GitHubRelease{
			TagName: "v0.2.0",
			HTMLURL: "https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0",
			Assets:  []ReleaseAsset{{Name: "bob-gemini-free-darwin-arm64"}},
		})
	}))
	defer server.Close()

	result, err := checkLatestDesktop(server.Client(), server.URL, "v0.1.7", "darwin", "arm64")
	if err != nil {
		t.Fatalf("checkLatestDesktop: %v", err)
	}
	if !result.HasUpdate {
		t.Fatal("expected newer desktop release")
	}
	if result.AssetAvailable {
		t.Fatal("CLI-only release was incorrectly reported as a desktop package")
	}
	if result.ReleaseURL != "https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0" {
		t.Fatalf("release URL = %q", result.ReleaseURL)
	}
}

func TestCheckLatestDesktopFindsNativePackage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(GitHubRelease{
			TagName: "v0.2.0",
			Assets: []ReleaseAsset{
				{
					Name:               "bob-gemini-free-wails-windows-amd64.exe",
					BrowserDownloadURL: "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.2.0/bob-windows.exe",
					Size:               1234,
				},
				{Name: "SHA256SUMS", BrowserDownloadURL: "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.2.0/SHA256SUMS"},
				{Name: "SHA256SUMS.sig", BrowserDownloadURL: "https://objects.githubusercontent.com/github-production-release-asset-2e65be/1/2/3?x=1"},
			},
		})
	}))
	defer server.Close()

	result, err := checkLatestDesktop(server.Client(), server.URL, "v0.1.7", "windows", "amd64")
	if err != nil {
		t.Fatalf("checkLatestDesktop: %v", err)
	}
	if !result.AssetAvailable || result.AssetName != "bob-gemini-free-wails-windows-amd64.exe" {
		t.Fatalf("desktop asset = %#v", result)
	}
	if result.DownloadURL != "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.2.0/bob-windows.exe" {
		t.Fatalf("download URL = %q", result.DownloadURL)
	}
	if result.AssetSize != 1234 {
		t.Fatalf("asset size = %d, want 1234", result.AssetSize)
	}
	if result.Channel != DesktopChannelStable {
		t.Fatalf("channel = %q, want %q", result.Channel, DesktopChannelStable)
	}
	if !result.ManifestAvailable {
		t.Fatal("signed manifest was not detected")
	}
	if result.ChecksumURL == "" || result.SignatureURL == "" {
		t.Fatalf("manifest URLs = %q / %q", result.ChecksumURL, result.SignatureURL)
	}
}

func TestCheckLatestDesktopPreviewSelectsHighestPublishedPreview(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]GitHubRelease{
			{
				TagName:    "v0.1.7-preview.3",
				Prerelease: true,
				HTMLURL:    "https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.1.7-preview.3",
			},
			{
				TagName:    "v0.1.7-preview.4",
				Prerelease: true,
				HTMLURL:    "https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.1.7-preview.4",
				Assets: []ReleaseAsset{
					{Name: "bob-gemini-free-macos-universal.zip", BrowserDownloadURL: "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.1.7-preview.4/bob-gemini-free-macos-universal.zip", Size: 2048},
					{Name: "SHA256SUMS", BrowserDownloadURL: "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.1.7-preview.4/SHA256SUMS"},
					{Name: "SHA256SUMS.sig", BrowserDownloadURL: "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.1.7-preview.4/SHA256SUMS.sig"},
				},
			},
			{
				TagName:    "v0.1.7-preview.5",
				Prerelease: true,
				Draft:      true,
			},
			{
				TagName: "v0.1.8",
			},
		})
	}))
	defer server.Close()

	result, err := checkLatestDesktopChannel(server.Client(), server.URL, "v0.1.7-preview.3", DesktopChannelPreview, "darwin", "arm64")
	if err != nil {
		t.Fatalf("checkLatestDesktopChannel: %v", err)
	}
	if result.LatestVersion != "v0.1.7-preview.4" || !result.HasUpdate {
		t.Fatalf("preview result = %#v, want v0.1.7-preview.4 update", result)
	}
	if result.Channel != DesktopChannelPreview || !result.AssetAvailable || !result.ManifestAvailable {
		t.Fatalf("preview channel result = %#v", result)
	}
}

func TestLegacyPreview7CanDiscoverSameKeyBridgePreview(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/preview" {
			t.Errorf("legacy Preview 7 requested %s; want preview channel only", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]GitHubRelease{{
			TagName:    "v0.2.0-preview.1",
			Prerelease: true,
			HTMLURL:    "https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.1",
			Assets: []ReleaseAsset{
				{Name: "bob-gemini-free-macos-universal.zip", BrowserDownloadURL: "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.2.0-preview.1/bob-gemini-free-macos-universal.zip", Size: 2048},
				{Name: "SHA256SUMS", BrowserDownloadURL: "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.2.0-preview.1/SHA256SUMS"},
				{Name: "SHA256SUMS.sig", BrowserDownloadURL: "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.2.0-preview.1/SHA256SUMS.sig"},
			},
		}})
	}))
	defer server.Close()

	// The published v0.1.7-preview.7 binary used this preview-only lookup.
	// A same-key bridge is therefore the only updater-mediated path it can see.
	result, err := checkLatestDesktopChannel(server.Client(), server.URL+"/preview", "v0.1.7-preview.7", DesktopChannelPreview, "darwin", "arm64")
	if err != nil {
		t.Fatalf("legacy Preview 7 bridge lookup: %v", err)
	}
	if result.LatestVersion != "v0.2.0-preview.1" || !result.HasUpdate || !result.AssetAvailable || !result.ManifestAvailable {
		t.Fatalf("legacy Preview 7 bridge result = %#v", result)
	}
	if requests != 1 {
		t.Fatalf("legacy Preview 7 requests = %d, want 1 preview request", requests)
	}
}

func TestLegacyPreview7CannotDiscoverStableWithoutBridge(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/preview" {
			t.Errorf("legacy Preview 7 requested %s; want preview channel only", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]GitHubRelease{{
			TagName:    "v0.1.7-preview.7",
			Prerelease: true,
			HTMLURL:    "https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.1.7-preview.7",
		}})
	}))
	defer server.Close()

	// Even if a future stable v0.2.0 exists, the released Preview 7 updater did
	// not query the stable endpoint and cannot discover that release directly.
	result, err := checkLatestDesktopChannel(server.Client(), server.URL+"/preview", "v0.1.7-preview.7", DesktopChannelPreview, "darwin", "arm64")
	if err != nil {
		t.Fatalf("legacy Preview 7 stable-only lookup: %v", err)
	}
	if result.HasUpdate {
		t.Fatalf("legacy Preview 7 reported a stable update through preview-only lookup: %#v", result)
	}
	if requests != 1 {
		t.Fatalf("legacy Preview 7 requests = %d, want 1 preview request", requests)
	}
}

func TestPreviewDesktopBuildCanMigrateToNewerStableRelease(t *testing.T) {
	var stableRequests, previewRequests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/stable":
			stableRequests++
			_ = json.NewEncoder(w).Encode(GitHubRelease{
				TagName: "v0.2.0",
				HTMLURL: "https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0",
				Assets: []ReleaseAsset{
					{Name: "bob-gemini-free-macos-universal.zip", BrowserDownloadURL: "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.2.0/bob-gemini-free-macos-universal.zip", Size: 2048},
					{Name: "SHA256SUMS", BrowserDownloadURL: "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.2.0/SHA256SUMS"},
					{Name: "SHA256SUMS.sig", BrowserDownloadURL: "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.2.0/SHA256SUMS.sig"},
				},
			})
		case "/preview":
			previewRequests++
			_ = json.NewEncoder(w).Encode([]GitHubRelease{{
				TagName:    "v0.1.7-preview.8",
				Prerelease: true,
				HTMLURL:    "https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.1.7-preview.8",
			}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	result, err := checkLatestDesktopPreviewWithStableMigration(server.Client(), server.URL+"/stable", server.URL+"/preview", "v0.1.7-preview.7", "darwin", "arm64")
	if err != nil {
		t.Fatalf("checkLatestDesktopPreviewWithStableMigration: %v", err)
	}
	if result.LatestVersion != "v0.2.0" || result.Channel != DesktopChannelStable {
		t.Fatalf("migration result = %#v, want stable v0.2.0", result)
	}
	if !result.HasUpdate || !result.AssetAvailable || !result.ManifestAvailable {
		t.Fatalf("migration result is not installable: %#v", result)
	}
	if stableRequests != 1 || previewRequests != 0 {
		t.Fatalf("requests = stable:%d preview:%d, want stable:1 preview:0", stableRequests, previewRequests)
	}
}

func TestPreviewDesktopBuildChecksPreviewWhenStableHasNoUpdate(t *testing.T) {
	var stableRequests, previewRequests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/stable":
			stableRequests++
			_ = json.NewEncoder(w).Encode(GitHubRelease{TagName: "v0.1.5", HTMLURL: "https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.1.5"})
		case "/preview":
			previewRequests++
			_ = json.NewEncoder(w).Encode([]GitHubRelease{{
				TagName:    "v0.1.7-preview.8",
				Prerelease: true,
				HTMLURL:    "https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.1.7-preview.8",
			}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	result, err := checkLatestDesktopPreviewWithStableMigration(server.Client(), server.URL+"/stable", server.URL+"/preview", "v0.1.7-preview.7", "darwin", "arm64")
	if err != nil {
		t.Fatalf("checkLatestDesktopPreviewWithStableMigration: %v", err)
	}
	if result.LatestVersion != "v0.1.7-preview.8" || result.Channel != DesktopChannelPreview || !result.HasUpdate {
		t.Fatalf("preview fallback result = %#v", result)
	}
	if stableRequests != 1 || previewRequests != 1 {
		t.Fatalf("requests = stable:%d preview:%d, want one request to each channel", stableRequests, previewRequests)
	}
}

func TestPreviewDesktopBuildDoesNotHideStableCheckFailure(t *testing.T) {
	var previewRequests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/stable":
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"message":"upstream unavailable"}`))
		case "/preview":
			previewRequests++
			_ = json.NewEncoder(w).Encode([]GitHubRelease{{
				TagName:    "v0.2.0-preview.1",
				Prerelease: true,
				HTMLURL:    "https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.1",
			}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	if _, err := checkLatestDesktopPreviewWithStableMigration(server.Client(), server.URL+"/stable", server.URL+"/preview", "v0.1.7-preview.7", "darwin", "arm64"); err == nil {
		t.Fatal("stable metadata failure was hidden by the preview path")
	}
	if previewRequests != 0 {
		t.Fatalf("preview requests = %d, want 0 after stable failure", previewRequests)
	}
}

func TestCheckLatestDesktopRejectsUnsupportedChannel(t *testing.T) {
	if _, err := checkLatestDesktopChannel(http.DefaultClient, "https://example.invalid/releases", "v0.1.7", "nightly", "darwin", "arm64"); err == nil {
		t.Fatal("unsupported update channel was accepted")
	}
}

func TestCheckLatestDesktopRejectsNonOfficialManifestURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(GitHubRelease{
			TagName: "v0.2.0",
			Assets: []ReleaseAsset{
				{Name: "bob-gemini-free-wails-windows-amd64.exe", BrowserDownloadURL: "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.2.0/bob.exe"},
				{Name: "SHA256SUMS", BrowserDownloadURL: "https://example.invalid/SHA256SUMS"},
				{Name: "SHA256SUMS.sig", BrowserDownloadURL: "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.2.0/SHA256SUMS.sig"},
			},
		})
	}))
	defer server.Close()

	if _, err := checkLatestDesktop(server.Client(), server.URL, "v0.1.7", "windows", "amd64"); err == nil {
		t.Fatal("non-official manifest URL was accepted")
	}
}

func TestCheckLatestDesktopRejectsNonOfficialPackageURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(GitHubRelease{
			TagName: "v0.2.0",
			Assets: []ReleaseAsset{{
				Name:               "bob-gemini-free-wails-windows-amd64.exe",
				BrowserDownloadURL: "https://example.invalid/bob-windows.exe",
			}},
		})
	}))
	defer server.Close()

	if _, err := checkLatestDesktop(server.Client(), server.URL, "v0.1.7", "windows", "amd64"); err == nil {
		t.Fatal("non-official desktop download URL was accepted")
	}
}
