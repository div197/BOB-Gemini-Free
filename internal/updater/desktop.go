package updater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

const (
	officialGitHubOwner      = "div197"
	officialGitHubRepository = "BOB-Gemini-Free"

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

	// Metadata checks are safe to retry because they are read-only GETs. Keep
	// the retry budget deliberately small so a GitHub outage cannot turn a
	// background check into a long-lived connection storm.
	updateMetadataAttempts  = 2
	updateMetadataRetryWait = 150 * time.Millisecond
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
// package for this platform. It never downloads, stages, or installs anything.
// Callers choose whether the check is explicit or part of a low-frequency
// background check.
func CheckLatestDesktop(currentVersion string) (*DesktopCheckResult, error) {
	return CheckLatestDesktopForChannelContext(context.Background(), currentVersion, DesktopChannelStable)
}

// CheckLatestDesktopForChannel checks only fixed official release channels
// selected by the build. Stable builds use GitHub's latest-release endpoint.
// Preview builds first check that same stable endpoint so a preview install can
// make the intentional one-way migration to a newer stable release; when no
// stable update exists, they use the published prerelease list and select the
// highest semver tag whose prerelease identifier is preview.N. The channel is a
// build property, never a runtime URL or user-controlled setting.
func CheckLatestDesktopForChannel(currentVersion, channel string) (*DesktopCheckResult, error) {
	return CheckLatestDesktopForChannelContext(context.Background(), currentVersion, channel)
}

// CheckLatestDesktopForChannelContext is the cancelable form used by the
// native desktop background checker. Cancellation stops an in-flight metadata
// request without changing the signed-manifest installation boundary.
func CheckLatestDesktopForChannelContext(ctx context.Context, currentVersion, channel string) (*DesktopCheckResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	client := newUpdateHTTPClient(8 * time.Second)
	if channel != DesktopChannelStable && channel != DesktopChannelPreview {
		return nil, fmt.Errorf("unsupported desktop update channel: %s", channel)
	}
	if channel == DesktopChannelPreview {
		return checkLatestDesktopPreviewWithStableMigrationContext(ctx, client, DesktopReleaseAPIURL, DesktopPreviewReleaseAPIURL, currentVersion, runtime.GOOS, runtime.GOARCH)
	}
	apiURL := DesktopReleaseAPIURL
	return checkLatestDesktopChannelContext(ctx, client, apiURL, currentVersion, channel, runtime.GOOS, runtime.GOARCH)
}

// checkLatestDesktopPreviewWithStableMigration keeps the two channels
// explicit: a preview app may move forward into stable, but a stable app never
// moves backwards into preview. A failed stable check is not hidden by a
// preview result because doing so would make the update UI report an
// incomplete view of the official release channels.
func checkLatestDesktopPreviewWithStableMigration(client *http.Client, stableURL, previewURL, currentVersion, targetOS, targetArch string) (*DesktopCheckResult, error) {
	return checkLatestDesktopPreviewWithStableMigrationContext(context.Background(), client, stableURL, previewURL, currentVersion, targetOS, targetArch)
}

func checkLatestDesktopPreviewWithStableMigrationContext(ctx context.Context, client *http.Client, stableURL, previewURL, currentVersion, targetOS, targetArch string) (*DesktopCheckResult, error) {
	stable, err := checkLatestDesktopChannelContext(ctx, client, stableURL, currentVersion, DesktopChannelStable, targetOS, targetArch)
	if err != nil {
		return nil, fmt.Errorf("stable desktop update check failed: %w", err)
	}
	if stable.HasUpdate {
		return stable, nil
	}
	return checkLatestDesktopChannelContext(ctx, client, previewURL, currentVersion, DesktopChannelPreview, targetOS, targetArch)
}

func checkLatestDesktop(client *http.Client, apiURL, currentVersion, targetOS, targetArch string) (*DesktopCheckResult, error) {
	return checkLatestDesktopChannel(client, apiURL, currentVersion, DesktopChannelStable, targetOS, targetArch)
}

func checkLatestDesktopChannel(client *http.Client, apiURL, currentVersion, channel, targetOS, targetArch string) (*DesktopCheckResult, error) {
	return checkLatestDesktopChannelContext(context.Background(), client, apiURL, currentVersion, channel, targetOS, targetArch)
}

func checkLatestDesktopChannelContext(ctx context.Context, client *http.Client, apiURL, currentVersion, channel, targetOS, targetArch string) (*DesktopCheckResult, error) {
	if client == nil {
		return nil, fmt.Errorf("desktop update check requires an HTTP client")
	}
	client = withUpdateRedirectPolicy(client)
	if ctx == nil {
		ctx = context.Background()
	}
	if channel != DesktopChannelStable && channel != DesktopChannelPreview {
		return nil, fmt.Errorf("unsupported desktop update channel: %s", channel)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create desktop update request: %w", err)
	}
	req.Header.Set("User-Agent", "BOB-Gemini-Free-Desktop-Updater")
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := doUpdateMetadataRequest(ctx, client, req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach desktop update server: %w", err)
	}
	if resp == nil {
		return nil, fmt.Errorf("desktop update server returned an empty response")
	}
	if resp.StatusCode != http.StatusOK {
		closeUpdateResponse(resp)
		return nil, fmt.Errorf("desktop update server returned HTTP %d", resp.StatusCode)
	}

	data, err := readBoundedUpdateResponse(resp, MaxUpdateMetadataBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to read desktop release metadata: %w", err)
	}
	release, err := decodeDesktopRelease(data, channel)
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

func decodeDesktopRelease(data []byte, channel string) (*GitHubRelease, error) {
	if channel == DesktopChannelStable {
		var release GitHubRelease
		if err := json.Unmarshal(data, &release); err != nil {
			return nil, fmt.Errorf("failed to parse desktop release metadata: %w", err)
		}
		if release.Draft || release.Prerelease {
			return nil, fmt.Errorf("stable desktop release endpoint returned a non-stable release")
		}
		return &release, nil
	}

	var releases []GitHubRelease
	if err := json.Unmarshal(data, &releases); err != nil {
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
	if err != nil || parsed.Scheme != "https" || parsed.User != nil || parsed.Fragment != "" {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	if parsed.Port() != "" && parsed.Port() != "443" {
		return false
	}
	if host == "objects.githubusercontent.com" || host == "release-assets.githubusercontent.com" {
		// GitHub's release CDN uses opaque paths and query parameters. The
		// signed manifest remains the authenticity boundary for those URLs.
		return true
	}
	if host != "github.com" {
		return false
	}
	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	return len(segments) >= 2 &&
		strings.EqualFold(segments[0], officialGitHubOwner) &&
		strings.EqualFold(segments[1], officialGitHubRepository)
}

func isOfficialGitHubAPIURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.User != nil || parsed.Fragment != "" {
		return false
	}
	if parsed.Port() != "" && parsed.Port() != "443" {
		return false
	}
	if !strings.EqualFold(parsed.Hostname(), "api.github.com") {
		return false
	}
	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	return len(segments) >= 3 &&
		strings.EqualFold(segments[0], "repos") &&
		strings.EqualFold(segments[1], officialGitHubOwner) &&
		strings.EqualFold(segments[2], officialGitHubRepository)
}

func isAllowedUpdateRedirect(raw *url.URL) bool {
	if raw == nil {
		return false
	}
	return isOfficialGitHubURL(raw.String()) || isOfficialGitHubAPIURL(raw.String())
}

func newUpdateHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("update server exceeded redirect limit")
			}
			if !isAllowedUpdateRedirect(req.URL) {
				return fmt.Errorf("update server redirected to an untrusted host")
			}
			return nil
		},
	}
}

// doUpdateMetadataRequest performs the bounded transport retry used by the
// read-only release metadata paths. It intentionally does not retry response
// status codes: in particular, a 429 must remain a visible provider-side
// capacity signal instead of being amplified, and redirect/policy failures
// must fail closed immediately.
func doUpdateMetadataRequest(ctx context.Context, client *http.Client, req *http.Request) (*http.Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if client == nil || req == nil {
		return nil, fmt.Errorf("update metadata request requires an HTTP client and request")
	}

	var lastErr error
	for attempt := 0; attempt < updateMetadataAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if attempt > 0 {
			timer := time.NewTimer(updateMetadataRetryWait)
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				return nil, ctx.Err()
			case <-timer.C:
			}
		}

		attemptReq := req.Clone(ctx)
		resp, err := client.Do(attemptReq)
		if err == nil {
			return resp, nil
		}
		closeUpdateResponse(resp)
		lastErr = err
		if !isRetryableUpdateMetadataError(ctx, err) {
			break
		}
	}
	return nil, lastErr
}

func isRetryableUpdateMetadataError(ctx context.Context, err error) bool {
	if err == nil || (ctx != nil && ctx.Err() != nil) {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		// A client-level timeout is retryable while the caller's context is
		// still alive. A caller deadline is rejected by the ctx check above.
		return true
	}
	var networkErr net.Error
	return errors.As(err, &networkErr) && (networkErr.Timeout() || networkErr.Temporary())
}

func withUpdateRedirectPolicy(client *http.Client) *http.Client {
	if client == nil {
		return nil
	}
	clone := *client
	previous := client.CheckRedirect
	clone.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("update server exceeded redirect limit")
		}
		if !isAllowedUpdateRedirect(req.URL) {
			return fmt.Errorf("update server redirected to an untrusted host")
		}
		if previous != nil {
			return previous(req, via)
		}
		return nil
	}
	return &clone
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
