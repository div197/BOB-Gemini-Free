package updater

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// DesktopUpdateFeedURL is a small, signed metadata document served from the
	// repository's raw-content CDN. It is deliberately fixed in the binary: the
	// feed can improve availability, but it cannot become a user-controlled
	// update source.
	DesktopUpdateFeedURL          = "https://raw.githubusercontent.com/div197/BOB-Gemini-Free/main/updates/desktop-feed.json"
	DesktopUpdateFeedSignatureURL = "https://raw.githubusercontent.com/div197/BOB-Gemini-Free/main/updates/desktop-feed.json.sig"

	desktopUpdateFeedSchema       = 1
	desktopUpdateFeedMaxLifetime  = 14 * 24 * time.Hour
	desktopUpdateFeedFutureSkew   = 24 * time.Hour
	maxDesktopUpdateFeedSignature = 16 << 10
)

// DesktopUpdateFeed is the signed, low-volume discovery document used by
// native desktop builds. It intentionally carries only release metadata; the
// package bytes and their SHA256SUMS remain on the official GitHub release and
// are verified independently before installation.
type DesktopUpdateFeed struct {
	Schema      int             `json:"schema"`
	GeneratedAt time.Time       `json:"generated_at"`
	ExpiresAt   time.Time       `json:"expires_at"`
	Stable      *GitHubRelease  `json:"stable"`
	Previews    []GitHubRelease `json:"previews"`
}

// SignUpdateFeed returns a base64-encoded detached Ed25519 signature over the
// exact feed bytes. The feed is signed as bytes, rather than re-marshaled by
// the verifier, so whitespace and field ordering cannot silently change the
// signed document.
func SignUpdateFeed(feed []byte, privateKey ed25519.PrivateKey) ([]byte, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid Ed25519 private key length: got %d bytes, want %d", len(privateKey), ed25519.PrivateKeySize)
	}
	if len(feed) == 0 {
		return nil, fmt.Errorf("update feed is empty")
	}
	signature := ed25519.Sign(privateKey, feed)
	return []byte(base64.StdEncoding.EncodeToString(signature) + "\n"), nil
}

func checkLatestDesktopFeedForChannelContext(ctx context.Context, client *http.Client, feedURL, signatureURL, currentVersion, channel, targetOS, targetArch string) (*DesktopCheckResult, error) {
	if client == nil {
		return nil, fmt.Errorf("signed desktop update feed requires an HTTP client")
	}
	if !isOfficialUpdateFeedURL(feedURL) || !isOfficialUpdateFeedURL(signatureURL) {
		return nil, fmt.Errorf("signed desktop update feed URL is not official")
	}
	if channel != DesktopChannelStable && channel != DesktopChannelPreview {
		return nil, fmt.Errorf("unsupported desktop update channel: %s", channel)
	}

	feed, err := fetchSignedDesktopUpdateFeed(ctx, client, feedURL, signatureURL)
	if err != nil {
		return nil, err
	}
	if channel == DesktopChannelStable {
		if feed.Stable == nil {
			return nil, fmt.Errorf("signed desktop update feed has no stable release")
		}
		return desktopCheckResultFromRelease(feed.Stable, currentVersion, DesktopChannelStable, targetOS, targetArch)
	}

	// Keep the same stable-first migration rule as the API path. A stable
	// release that has no native package must not hide a newer native preview.
	if feed.Stable != nil {
		stable, stableErr := desktopCheckResultFromRelease(feed.Stable, currentVersion, DesktopChannelStable, targetOS, targetArch)
		if stableErr != nil {
			return nil, stableErr
		}
		if stable.HasUpdate && stable.AssetAvailable {
			return stable, nil
		}
	}

	preview := highestPublishedPreview(feed.Previews)
	if preview == nil {
		return nil, fmt.Errorf("signed desktop update feed has no published preview release")
	}
	return desktopCheckResultFromRelease(preview, currentVersion, DesktopChannelPreview, targetOS, targetArch)
}

func fetchSignedDesktopUpdateFeed(ctx context.Context, client *http.Client, feedURL, signatureURL string) (*DesktopUpdateFeed, error) {
	publicKey, err := configuredUpdatePublicKey()
	if err != nil {
		return nil, err
	}
	feedBytes, err := fetchDesktopUpdateFeedResource(ctx, client, feedURL, MaxUpdateMetadataBytes)
	if err != nil {
		return nil, fmt.Errorf("download desktop update feed: %w", err)
	}
	signature, err := fetchDesktopUpdateFeedResource(ctx, client, signatureURL, maxDesktopUpdateFeedSignature)
	if err != nil {
		return nil, fmt.Errorf("download desktop update feed signature: %w", err)
	}
	if err := verifySignedUpdateFeed(publicKey, feedBytes, signature); err != nil {
		return nil, err
	}

	var feed DesktopUpdateFeed
	if err := json.Unmarshal(feedBytes, &feed); err != nil {
		return nil, fmt.Errorf("parse signed desktop update feed: %w", err)
	}
	if err := validateDesktopUpdateFeed(&feed, time.Now()); err != nil {
		return nil, err
	}
	return &feed, nil
}

func fetchDesktopUpdateFeedResource(ctx context.Context, client *http.Client, resourceURL string, maxBytes int64) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, err := url.ParseRequestURI(resourceURL); err != nil {
		return nil, fmt.Errorf("invalid feed resource URL: %w", err)
	}
	client = withUpdateRedirectPolicy(client)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resourceURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create feed resource request: %w", err)
	}
	req.Header.Set("User-Agent", "BOB-Gemini-Free-Desktop-Updater")
	req.Header.Set("Accept", "application/json, text/plain;q=0.9")
	resp, err := doUpdateMetadataRequest(ctx, client, req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("feed resource returned an empty response")
	}
	if resp.StatusCode != http.StatusOK {
		closeUpdateResponse(resp)
		return nil, fmt.Errorf("feed resource returned HTTP %d", resp.StatusCode)
	}
	return readBoundedUpdateResponse(resp, maxBytes)
}

func verifySignedUpdateFeed(publicKey ed25519.PublicKey, feed, encodedSignature []byte) error {
	if len(publicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid update feed public key length")
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(encodedSignature)))
	if err != nil || len(decoded) != ed25519.SignatureSize || !ed25519.Verify(publicKey, feed, decoded) {
		return fmt.Errorf("signed desktop update feed signature is invalid")
	}
	return nil
}

func validateDesktopUpdateFeed(feed *DesktopUpdateFeed, now time.Time) error {
	if feed == nil {
		return fmt.Errorf("signed desktop update feed is empty")
	}
	if feed.Schema != desktopUpdateFeedSchema {
		return fmt.Errorf("unsupported signed desktop update feed schema: %d", feed.Schema)
	}
	if feed.Stable == nil {
		return fmt.Errorf("signed desktop update feed has no stable release")
	}
	if feed.GeneratedAt.IsZero() || feed.ExpiresAt.IsZero() || !feed.ExpiresAt.After(feed.GeneratedAt) {
		return fmt.Errorf("signed desktop update feed has invalid validity timestamps")
	}
	if feed.ExpiresAt.Sub(feed.GeneratedAt) > desktopUpdateFeedMaxLifetime {
		return fmt.Errorf("signed desktop update feed validity window is too long")
	}
	if now.IsZero() {
		now = time.Now()
	}
	if now.After(feed.ExpiresAt) {
		return fmt.Errorf("signed desktop update feed has expired")
	}
	if feed.GeneratedAt.After(now.Add(desktopUpdateFeedFutureSkew)) {
		return fmt.Errorf("signed desktop update feed was generated too far in the future")
	}
	if feed.Stable.Draft || feed.Stable.Prerelease || !isStableReleaseTag(feed.Stable.TagName) {
		return fmt.Errorf("signed desktop update feed contains an invalid stable release")
	}
	if feed.Stable.HTMLURL != "" && !isOfficialGitHubURL(feed.Stable.HTMLURL) {
		return fmt.Errorf("signed desktop update feed contains a non-official stable release URL")
	}
	for index := range feed.Previews {
		preview := &feed.Previews[index]
		if preview.Draft || !preview.Prerelease || !isPreviewReleaseTag(preview.TagName) {
			return fmt.Errorf("signed desktop update feed contains an invalid preview release at index %d", index)
		}
		if preview.HTMLURL != "" && !isOfficialGitHubURL(preview.HTMLURL) {
			return fmt.Errorf("signed desktop update feed contains a non-official preview release URL")
		}
	}
	if highestPublishedPreview(feed.Previews) == nil {
		return fmt.Errorf("signed desktop update feed has no published preview release")
	}
	return nil
}

func highestPublishedPreview(releases []GitHubRelease) *GitHubRelease {
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
	return selected
}

func isOfficialUpdateFeedURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.User != nil || parsed.Fragment != "" || parsed.RawQuery != "" {
		return false
	}
	if parsed.Port() != "" && parsed.Port() != "443" {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	if host != "raw.githubusercontent.com" {
		return false
	}
	return parsed.Path == "/div197/BOB-Gemini-Free/main/updates/desktop-feed.json" ||
		parsed.Path == "/div197/BOB-Gemini-Free/main/updates/desktop-feed.json.sig"
}
