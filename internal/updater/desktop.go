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
	DesktopChannelStable  = "stable"
	DesktopChannelPreview = "preview"

	// DesktopReleaseAPIURL is intentionally fixed to the project's official
	// repository. A desktop build must not accept an update source from mutable
	// runtime configuration.
	DesktopReleaseAPIURL = "https://api.github.com/repos/div197/bob-gemini-free/releases/latest"
	// Keep the preview listing bounded. GitHub's unauthenticated API can time
	// out on a 100-release page even for a small public repository; the updater
	// only needs the recent preview channel history.
	DesktopPreviewReleaseAPIURL = "https://api.github.com/repos/div197/bob-gemini-free/releases?per_page=30"
	DesktopReleaseURL           = "https://github.com/div197/BOB-Gemini-Free/releases/latest"
	DesktopPreviewReleaseURL    = "https://github.com/div197/BOB-Gemini-Free/releases"
)

// DesktopCheckResult describes a native-package update without downloading or
// replacing anything. Installation is deliberately a separate operation: a
// macOS bundle, Windows installer, and Linux package each have different safe
// replacement semantics.
type DesktopCheckResult struct {
	CurrentVersion    string `json:"current_version"`
	LatestVersion     string `json:"latest_version"`
	HasUpdate         bool   `json:"has_update"`
	AssetAvailable    bool   `json:"asset_available"`
	AssetName         string `json:"asset_name,omitempty"`
	AssetSize         int64  `json:"asset_size,omitempty"`
	DownloadURL       string `json:"download_url,omitempty"`
	ReleaseURL        string `json:"release_url"`
	ChecksumURL       string `json:"checksum_url,omitempty"`
	SignatureURL      string `json:"signature_url,omitempty"`
	ManifestAvailable bool   `json:"manifest_available"`
	Channel           string `json:"channel"`
}

// CheckLatestDesktop checks whether the official release contains a native
// package for this platform. It performs no automatic network request from
// application startup; callers should invoke it only from an explicit user
// action such as “Check for updates”.
func CheckLatestDesktop(currentVersion string) (*DesktopCheckResult, error) {
	return CheckLatestDesktopForChannel(currentVersion, DesktopChannelStable)
}

// CheckLatestDesktopForChannel checks only the fixed official release channel
// selected by the build. Stable builds use GitHub's latest-release endpoint;
// preview builds use the published prerelease list and select the highest
// semver tag whose prerelease identifier is preview.N. The channel is a build
// property, never a runtime URL or user-controlled setting.
func CheckLatestDesktopForChannel(currentVersion, channel string) (*DesktopCheckResult, error) {
	client := &http.Client{Timeout: 8 * time.Second}
	apiURL := DesktopReleaseAPIURL
	if channel == DesktopChannelPreview {
		apiURL = DesktopPreviewReleaseAPIURL
	}
	return checkLatestDesktopChannel(client, apiURL, currentVersion, channel, runtime.GOOS, runtime.GOARCH)
}

func checkLatestDesktop(client *http.Client, apiURL, currentVersion, targetOS, targetArch string) (*DesktopCheckResult, error) {
	return checkLatestDesktopChannel(client, apiURL, currentVersion, DesktopChannelStable, targetOS, targetArch)
}

func checkLatestDesktopChannel(client *http.Client, apiURL, currentVersion, channel, targetOS, targetArch string) (*DesktopCheckResult, error) {
	if client == nil {
		return nil, fmt.Errorf("desktop update check requires an HTTP client")
	}
	if channel != DesktopChannelStable && channel != DesktopChannelPreview {
		return nil, fmt.Errorf("unsupported desktop update channel: %s", channel)
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

	release, err := decodeDesktopRelease(resp, channel)
	if err != nil {
		return nil, err
	}
	latestVersion := release.TagName
	if latestVersion == "" {
		return nil, fmt.Errorf("desktop release metadata has no tag name")
	}
	asset := findMatchingDesktopAsset(release.Assets, targetOS, targetArch)
	releaseURL := release.HTMLURL
	if releaseURL == "" {
		releaseURL = DesktopReleaseURL
		if channel == DesktopChannelPreview {
			releaseURL = DesktopPreviewReleaseURL
		}
	}
	if !isOfficialGitHubURL(releaseURL) {
		return nil, fmt.Errorf("desktop release contains a non-official release URL; refusing to offer it")
	}
	result := &DesktopCheckResult{
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		HasUpdate:      isNewerVersion(currentVersion, latestVersion),
		ReleaseURL:     releaseURL,
		Channel:        channel,
	}
	if asset != nil {
		if asset.BrowserDownloadURL == "" || !isOfficialGitHubURL(asset.BrowserDownloadURL) {
			return nil, fmt.Errorf("desktop release asset %q has a non-official download URL; refusing to offer it", asset.Name)
		}
		result.AssetAvailable = true
		result.AssetName = asset.Name
		result.AssetSize = asset.Size
		result.DownloadURL = asset.BrowserDownloadURL
	}
	checksumAsset := findAssetByName(release.Assets, "SHA256SUMS")
	signatureAsset := findAssetByName(release.Assets, "SHA256SUMS.sig")
	if checksumAsset != nil && signatureAsset != nil {
		if !isOfficialGitHubURL(checksumAsset.BrowserDownloadURL) || !isOfficialGitHubURL(signatureAsset.BrowserDownloadURL) {
			return nil, fmt.Errorf("desktop release manifest URLs are not official GitHub downloads; refusing to offer automatic installation")
		}
		result.ChecksumURL = checksumAsset.BrowserDownloadURL
		result.SignatureURL = signatureAsset.BrowserDownloadURL
		result.ManifestAvailable = true
	}
	return result, nil
}

func decodeDesktopRelease(resp *http.Response, channel string) (*GitHubRelease, error) {
	if channel == DesktopChannelStable {
		var release GitHubRelease
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return nil, fmt.Errorf("failed to parse desktop release metadata: %w", err)
		}
		if release.Draft || release.Prerelease {
			return nil, fmt.Errorf("stable desktop release endpoint returned a non-stable release")
		}
		return &release, nil
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to parse desktop preview release metadata: %w", err)
	}
	var selected *GitHubRelease
	for index := range releases {
		release := &releases[index]
		if release.Draft || !release.Prerelease || !isPreviewReleaseTag(release.TagName) {
			continue
		}
		if selected == nil {
			selected = release
			continue
		}
		comparison, valid := compareSemanticVersions(selected.TagName, release.TagName)
		if valid && comparison < 0 {
			selected = release
		}
	}
	if selected == nil {
		return nil, fmt.Errorf("no published preview desktop release was found")
	}
	return selected, nil
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
		// The branded universal archive is preferred so one Mac release serves both
		// Apple Silicon and Intel users. Architecture-specific archives remain
		// valid fallbacks for future size-optimized releases.
		return []string{
			"bob-gemini-free-macos-universal.zip",
			fmt.Sprintf("bob-gemini-free-macos-%s.zip", targetArch),
			// Legacy Preview 2 names are accepted only for migration/recovery.
			"bob-gemini-free-wails-macos-universal.zip",
			fmt.Sprintf("bob-gemini-free-wails-macos-%s.zip", targetArch),
		}
	case "windows":
		return []string{
			fmt.Sprintf("bob-gemini-free-windows-%s.exe", targetArch),
			fmt.Sprintf("bob-gemini-free-wails-windows-%s.exe", targetArch),
		}
	case "linux":
		return []string{
			fmt.Sprintf("bob-gemini-free-linux-%s.tar.gz", targetArch),
			fmt.Sprintf("bob-gemini-free-wails-linux-%s.tar.gz", targetArch),
		}
	default:
		return nil
	}
}
