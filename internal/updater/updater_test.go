package updater

import (
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
