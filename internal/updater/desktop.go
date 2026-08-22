package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

const (
	// DesktopReleaseAPIURL is intentionally fixed to the project's official
	// repository. A desktop build must not accept an update source from mutable
	// runtime configuration.
	DesktopReleaseAPIURL = "https://api.github.com/repos/div197/bob-gemini-free/releases/latest"
	DesktopReleaseURL    = "https://github.com/div197/BOB-Gemini-Free/releases/latest"
)

// DesktopCheckResult describes a native-package update without downloading or
// replacing anything. Installation is deliberately a separate operation: a
// Wails bundle, Windows installer, and Linux package each have different safe
// replacement semantics.
type DesktopCheckResult struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	HasUpdate      bool   `json:"has_update"`
	AssetAvailable bool   `json:"asset_available"`
	AssetName      string `json:"asset_name,omitempty"`
	DownloadURL    string `json:"download_url,omitempty"`
	ReleaseURL     string `json:"release_url"`
}

// CheckLatestDesktop checks whether the official release contains a native
// package for this platform. It performs no automatic network request from
// application startup; callers should invoke it only from an explicit user
// action such as “Check for updates”.
func CheckLatestDesktop(currentVersion string) (*DesktopCheckResult, error) {
	client := &http.Client{Timeout: 8 * time.Second}
	return checkLatestDesktop(client, DesktopReleaseAPIURL, currentVersion, runtime.GOOS, runtime.GOARCH)
}

func checkLatestDesktop(client *http.Client, apiURL, currentVersion, targetOS, targetArch string) (*DesktopCheckResult, error) {
	if client == nil {
		return nil, fmt.Errorf("desktop update check requires an HTTP client")
	}
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create desktop update request: %w", err)
	}
	req.Header.Set("User-Agent", "BOB-Gemini-Free-Desktop-Updater")
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach desktop update server: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("desktop update server returned HTTP %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse desktop release metadata: %w", err)
	}
	latestVersion := release.TagName
	asset := findMatchingDesktopAsset(release.Assets, targetOS, targetArch)
	releaseURL := release.HTMLURL
	if releaseURL == "" {
		releaseURL = DesktopReleaseURL
	}
	if !isOfficialGitHubURL(releaseURL) {
		return nil, fmt.Errorf("desktop release contains a non-official release URL; refusing to offer it")
	}
	result := &DesktopCheckResult{
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		HasUpdate:      isNewerVersion(strings.TrimPrefix(currentVersion, "v"), strings.TrimPrefix(latestVersion, "v")),
		ReleaseURL:     releaseURL,
	}
	if asset != nil {
		if asset.BrowserDownloadURL == "" || !isOfficialGitHubURL(asset.BrowserDownloadURL) {
			return nil, fmt.Errorf("desktop release asset %q has a non-official download URL; refusing to offer it", asset.Name)
		}
		result.AssetAvailable = true
		result.AssetName = asset.Name
		result.DownloadURL = asset.BrowserDownloadURL
	}
	return result, nil
}

func isOfficialGitHubURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.User != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	return host == "github.com" || host == "objects.githubusercontent.com"
}

func findMatchingDesktopAsset(assets []ReleaseAsset, targetOS, targetArch string) *ReleaseAsset {
	candidates := desktopAssetNames(targetOS, targetArch)
	for _, candidate := range candidates {
		if asset := findAssetByName(assets, candidate); asset != nil {
			return asset
		}
	}
	return nil
}

func desktopAssetNames(targetOS, targetArch string) []string {
	switch targetOS {
	case "darwin":
		// The universal archive is preferred so one Mac release serves both
		// Apple Silicon and Intel users. Architecture-specific archives remain
		// valid fallbacks for future size-optimized releases.
		return []string{
			"bob-gemini-free-wails-macos-universal.zip",
			fmt.Sprintf("bob-gemini-free-wails-macos-%s.zip", targetArch),
		}
	case "windows":
		return []string{fmt.Sprintf("bob-gemini-free-wails-windows-%s.exe", targetArch)}
	case "linux":
		return []string{fmt.Sprintf("bob-gemini-free-wails-linux-%s.tar.gz", targetArch)}
	default:
		return nil
	}
}
