package updater

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
	}

	for _, tt := range tests {
		got := isNewerVersion(tt.current, tt.latest)
		if got != tt.want {
			t.Errorf("isNewerVersion(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
		}
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
