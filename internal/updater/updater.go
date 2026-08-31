package updater

import (
	"bufio"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ReleaseAsset represents an asset attached to a GitHub release.
type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// GitHubRelease represents release metadata returned by the GitHub API.
type GitHubRelease struct {
	TagName     string         `json:"tag_name"`
	Name        string         `json:"name"`
	Body        string         `json:"body"`
	HTMLURL     string         `json:"html_url"`
	Draft       bool           `json:"draft"`
	Prerelease  bool           `json:"prerelease"`
	PublishedAt time.Time      `json:"published_at"`
	Assets      []ReleaseAsset `json:"assets"`
}

// CheckResult contains the result of an update check.
type CheckResult struct {
	CurrentVersion string         `json:"current_version"`
	LatestVersion  string         `json:"latest_version"`
	HasUpdate      bool           `json:"has_update"`
	Release        *GitHubRelease `json:"release,omitempty"`
	DownloadURL    string         `json:"download_url,omitempty"`
	ChecksumURL    string         `json:"checksum_url,omitempty"`
	SignatureURL   string         `json:"signature_url,omitempty"`
}

const (
	// MaxUpdateMetadataBytes bounds JSON returned by GitHub before it reaches
	// the decoder. The current release metadata is far smaller; this protects
	// a student device from a pathological or intercepted response.
	MaxUpdateMetadataBytes  int64 = 4 << 20
	MaxUpdateManifestBytes  int64 = 1 << 20
	MaxUpdateSignatureBytes int64 = 16 << 10

	// MaxUpdateArtifactBytes bounds the amount of untrusted release data that can
	// be written to the temporary update candidate before checksum verification.
	// It is intentionally much larger than the current binaries while preventing
	// an unexpectedly large response from exhausting local disk space.
	MaxUpdateArtifactBytes int64 = 512 << 20
)

// BuildUpdatePublicKey is injected into release builds with -ldflags. A
// release binary therefore carries its trust anchor instead of depending on a
// mutable runtime environment. Development builds may use the environment
// fallback documented below.
var BuildUpdatePublicKey string

// CheckLatest queries the official GitHub repository for the latest release.
func CheckLatest(currentVersion string) (*CheckResult, error) {
	apiURL := "https://api.github.com/repos/div197/bob-gemini-free/releases/latest"
	client := newUpdateHTTPClient(8 * time.Second)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create update request: %w", err)
	}
	req.Header.Set("User-Agent", "BOB-Gemini-Free-Updater")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := doUpdateMetadataRequest(context.Background(), client, req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach update server: %w", err)
	}
	if resp == nil {
		return nil, fmt.Errorf("update server returned an empty response")
	}

	if resp.StatusCode != http.StatusOK {
		closeUpdateResponse(resp)
		return nil, fmt.Errorf("update server returned HTTP %d", resp.StatusCode)
	}

	data, err := readBoundedUpdateResponse(resp, MaxUpdateMetadataBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to read release metadata: %w", err)
	}
	var release GitHubRelease
	if err := json.Unmarshal(data, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release metadata: %w", err)
	}
	if strings.TrimSpace(release.TagName) == "" {
		return nil, fmt.Errorf("release metadata has no tag name")
	}
	if release.Draft || release.Prerelease {
		return nil, fmt.Errorf("latest release endpoint returned a non-stable release")
	}
	if release.HTMLURL != "" && !isOfficialGitHubURL(release.HTMLURL) {
		return nil, fmt.Errorf("release metadata contains a non-official release URL")
	}

	latestVer := strings.TrimPrefix(release.TagName, "v")
	currVer := strings.TrimPrefix(currentVersion, "v")

	hasUpdate := isNewerVersion(currVer, latestVer)
	targetAsset := findMatchingAsset(release.Assets, runtime.GOOS, runtime.GOARCH)

	downloadURL := ""
	if targetAsset != nil {
		if targetAsset.BrowserDownloadURL == "" || !isOfficialGitHubURL(targetAsset.BrowserDownloadURL) {
			return nil, fmt.Errorf("release asset %q has a non-official download URL", targetAsset.Name)
		}
		downloadURL = targetAsset.BrowserDownloadURL
	}
	checksumURL := ""
	if checksumAsset := findAssetByName(release.Assets, "SHA256SUMS"); checksumAsset != nil {
		if checksumAsset.BrowserDownloadURL == "" || !isOfficialGitHubURL(checksumAsset.BrowserDownloadURL) {
			return nil, fmt.Errorf("release checksum manifest has a non-official download URL")
		}
		checksumURL = checksumAsset.BrowserDownloadURL
	}
	signatureURL := ""
	if signatureAsset := findAssetByName(release.Assets, "SHA256SUMS.sig"); signatureAsset != nil {
		if signatureAsset.BrowserDownloadURL == "" || !isOfficialGitHubURL(signatureAsset.BrowserDownloadURL) {
			return nil, fmt.Errorf("release checksum signature has a non-official download URL")
		}
		signatureURL = signatureAsset.BrowserDownloadURL
	}

	return &CheckResult{
		CurrentVersion: currentVersion,
		LatestVersion:  release.TagName,
		HasUpdate:      hasUpdate,
		Release:        &release,
		DownloadURL:    downloadURL,
		ChecksumURL:    checksumURL,
		SignatureURL:   signatureURL,
	}, nil
}

// SelfUpdate downloads and replaces the currently running binary.
func SelfUpdate(currentVersion string, logFn func(string, ...any)) error {
	if logFn == nil {
		logFn = func(string, ...any) {}
	}

	logFn("[*] Checking for updates on GitHub...")
	res, err := CheckLatest(currentVersion)
	if err != nil {
		return fmt.Errorf("update check failed: %w", err)
	}

	if !res.HasUpdate {
		logFn("[✔] You are already on the latest version (%s)!", currentVersion)
		return nil
	}

	if res.DownloadURL == "" {
		return fmt.Errorf("no compatible release binary found for %s/%s in release %s", runtime.GOOS, runtime.GOARCH, res.LatestVersion)
	}
	if res.ChecksumURL == "" || res.SignatureURL == "" {
		return fmt.Errorf("release %s has no signed SHA256SUMS manifest; refusing unsigned update", res.LatestVersion)
	}
	publicKey, err := configuredUpdatePublicKey()
	if err != nil {
		return err
	}

	logFn("[*] Found newer release: %s (current: %s)", res.LatestVersion, currentVersion)
	logFn("[*] Downloading from: %s", res.DownloadURL)

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlink: %w", err)
	}

	tempFile, err := os.CreateTemp(filepath.Dir(exePath), "bob-gemini-free-update-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create same-filesystem temp file: %w", err)
	}
	tempPath := tempFile.Name()
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close update temp file: %w", err)
	}
	defer os.Remove(tempPath)

	client := newUpdateHTTPClient(90 * time.Second)
	manifest, err := downloadBytes(client, res.ChecksumURL, MaxUpdateManifestBytes)
	if err != nil {
		return fmt.Errorf("failed to download checksum manifest: %w", err)
	}
	signature, err := downloadBytes(client, res.SignatureURL, MaxUpdateSignatureBytes)
	if err != nil {
		return fmt.Errorf("failed to download checksum signature: %w", err)
	}
	if err := verifyReleaseManifest(publicKey, manifest, signature); err != nil {
		return fmt.Errorf("release authenticity verification failed: %w", err)
	}

	assetName := ""
	assetSize := int64(0)
	if res.Release != nil {
		if asset := findAssetByURL(res.Release.Assets, res.DownloadURL); asset != nil {
			assetName = asset.Name
			assetSize = asset.Size
			if asset.Size <= 0 {
				return fmt.Errorf("release asset %s has no trusted positive size", asset.Name)
			}
			if asset.Size > MaxUpdateArtifactBytes {
				return fmt.Errorf("release asset %s exceeds the %d-byte safety limit", asset.Name, MaxUpdateArtifactBytes)
			}
		}
	}
	if assetName == "" {
		return fmt.Errorf("release metadata did not identify downloaded asset; refusing update")
	}
	shaSum, err := downloadVerifiedArtifact(client, res.DownloadURL, tempPath, assetName, manifest, signature, publicKey)
	if err != nil {
		return fmt.Errorf("verified binary download failed: %w", err)
	}
	info, err := os.Stat(tempPath)
	if err != nil {
		return fmt.Errorf("failed to inspect verified update: %w", err)
	}
	n := info.Size()

	if n != assetSize {
		return fmt.Errorf("downloaded binary size changed during update: got %d bytes, release metadata declares %d", n, assetSize)
	}

	logFn("[+] Downloaded %d bytes successfully (SHA-256: %s).", n, shaSum[:16]+"...")

	// Verify valid executable binary magic bytes
	if err := verifyBinaryMagic(tempPath, runtime.GOOS); err != nil {
		return fmt.Errorf("integrity check failed: %w", err)
	}

	if err := os.Chmod(tempPath, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	if runtime.GOOS == "windows" {
		oldExe := exePath + ".old"
		_ = os.Remove(oldExe)
		if err := os.Rename(exePath, oldExe); err != nil {
			return fmt.Errorf("failed to rename existing binary: %w", err)
		}
		if err := copyFile(tempPath, exePath); err != nil {
			_ = os.Rename(oldExe, exePath)
			return fmt.Errorf("failed to replace binary: %w", err)
		}
	} else {
		if err := os.Rename(tempPath, exePath); err != nil {
			return fmt.Errorf("failed to atomically replace binary at %s: %w", exePath, err)
		}
	}

	logFn("[✔] Successfully updated BOB Gemini Free to %s!", res.LatestVersion)
	logFn("[✔] Updated executable location: %s", exePath)
	return nil
}

func findMatchingAsset(assets []ReleaseAsset, targetOS, targetArch string) *ReleaseAsset {
	var expectedName string
	if targetOS == "windows" {
		expectedName = fmt.Sprintf("bob-gemini-free-windows-%s.exe", targetArch)
	} else {
		expectedName = fmt.Sprintf("bob-gemini-free-%s-%s", targetOS, targetArch)
	}

	for i := range assets {
		if assets[i].Name == expectedName {
			return &assets[i]
		}
	}
	return nil
}

func findAssetByName(assets []ReleaseAsset, name string) *ReleaseAsset {
	for i := range assets {
		if assets[i].Name == name {
			return &assets[i]
		}
	}
	return nil
}

func findAssetByURL(assets []ReleaseAsset, downloadURL string) *ReleaseAsset {
	for i := range assets {
		if assets[i].BrowserDownloadURL == downloadURL {
			return &assets[i]
		}
	}
	return nil
}

func configuredUpdatePublicKey() (ed25519.PublicKey, error) {
	raw := strings.TrimSpace(BuildUpdatePublicKey)
	if raw == "" {
		raw = strings.TrimSpace(os.Getenv("BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY"))
	}
	if raw == "" {
		return nil, fmt.Errorf("no embedded or configured update public key is available; refusing unsigned update")
	}
	decoded, err := decodeEncodedBytes(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid update public key encoding: %w", err)
	}
	if len(decoded) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid update public key length: got %d bytes, want %d", len(decoded), ed25519.PublicKeySize)
	}
	return ed25519.PublicKey(decoded), nil
}

// embeddedUpdatePublicKey is the native desktop trust boundary. Unlike the
// CLI's local-development path, a packaged desktop app must never allow a
// mutable runtime environment variable to choose its release signer.
func embeddedUpdatePublicKey() (ed25519.PublicKey, error) {
	raw := strings.TrimSpace(BuildUpdatePublicKey)
	if raw == "" {
		return nil, fmt.Errorf("no embedded desktop update public key is available; refusing automatic installation")
	}
	decoded, err := decodeEncodedBytes(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid embedded desktop update public key encoding: %w", err)
	}
	if len(decoded) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid embedded desktop update public key length: got %d bytes, want %d", len(decoded), ed25519.PublicKeySize)
	}
	return ed25519.PublicKey(decoded), nil
}

func decodeEncodedBytes(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if len(value)%2 == 0 {
		if decoded, err := hex.DecodeString(value); err == nil {
			return decoded, nil
		}
	}
	if decoded, err := base64.StdEncoding.DecodeString(value); err == nil {
		return decoded, nil
	}
	decoded, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("expected base64 or hexadecimal data")
	}
	return decoded, nil
}

func downloadBytes(client *http.Client, downloadURL string, maxBytes int64) ([]byte, error) {
	if client == nil {
		return nil, fmt.Errorf("update download requires an HTTP client")
	}
	if maxBytes <= 0 {
		return nil, fmt.Errorf("invalid update response size limit: %d", maxBytes)
	}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("update download returned an empty response")
	}
	if resp.StatusCode != http.StatusOK {
		closeUpdateResponse(resp)
		return nil, fmt.Errorf("download server returned HTTP %d", resp.StatusCode)
	}
	return readBoundedUpdateResponse(resp, maxBytes)
}

func verifyReleaseManifest(publicKey ed25519.PublicKey, manifest, encodedSignature []byte) error {
	if len(publicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid public key length")
	}
	signature, err := decodeEncodedBytes(strings.TrimSpace(string(encodedSignature)))
	if err != nil {
		return fmt.Errorf("invalid signature encoding: %w", err)
	}
	if len(signature) != ed25519.SignatureSize || !ed25519.Verify(publicKey, manifest, signature) {
		return fmt.Errorf("manifest signature is invalid")
	}
	return nil
}

func checksumForAsset(manifest []byte, assetName string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(manifest)))
	var found string
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		name := strings.TrimPrefix(fields[1], "*")
		if name != assetName {
			continue
		}
		digest := strings.ToLower(fields[0])
		if len(digest) != sha256.Size*2 {
			return "", fmt.Errorf("invalid digest length for %s", assetName)
		}
		if _, err := hex.DecodeString(digest); err != nil {
			return "", fmt.Errorf("invalid digest for %s: %w", assetName, err)
		}
		if found != "" && found != digest {
			return "", fmt.Errorf("conflicting digests for %s", assetName)
		}
		found = digest
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("manifest has no entry for %s", assetName)
	}
	return found, nil
}

func downloadVerifiedArtifact(client *http.Client, downloadURL, destination, assetName string, manifest, signature []byte, publicKey ed25519.PublicKey) (string, error) {
	return downloadVerifiedArtifactLimited(client, downloadURL, destination, assetName, manifest, signature, publicKey, MaxUpdateArtifactBytes)
}

func downloadVerifiedArtifactLimited(client *http.Client, downloadURL, destination, assetName string, manifest, signature []byte, publicKey ed25519.PublicKey, maxBytes int64) (string, error) {
	if client == nil {
		return "", fmt.Errorf("verified update download requires an HTTP client")
	}
	if maxBytes <= 0 {
		return "", fmt.Errorf("invalid update artifact size limit: %d", maxBytes)
	}
	if err := verifyReleaseManifest(publicKey, manifest, signature); err != nil {
		return "", err
	}
	expected, err := checksumForAsset(manifest, assetName)
	if err != nil {
		return "", err
	}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", fmt.Errorf("verified update download returned an empty response")
	}
	if resp.StatusCode != http.StatusOK {
		closeUpdateResponse(resp)
		return "", fmt.Errorf("download server returned HTTP %d", resp.StatusCode)
	}
	if resp.Body == nil {
		return "", fmt.Errorf("verified update download returned an empty response body")
	}
	if resp.ContentLength > maxBytes {
		closeUpdateResponse(resp)
		return "", fmt.Errorf("download exceeded %d-byte safety limit", maxBytes)
	}

	file, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		closeUpdateResponse(resp)
		return "", err
	}
	defer closeUpdateResponse(resp)
	hasher := sha256.New()
	n, copyErr := io.Copy(io.MultiWriter(file, hasher), io.LimitReader(resp.Body, maxBytes+1))
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(destination)
		return "", copyErr
	}
	if closeErr != nil {
		_ = os.Remove(destination)
		return "", closeErr
	}
	if n > maxBytes {
		_ = os.Remove(destination)
		return "", fmt.Errorf("download exceeded %d-byte safety limit", maxBytes)
	}
	actual := hex.EncodeToString(hasher.Sum(nil))
	if actual != expected {
		_ = os.Remove(destination)
		return "", fmt.Errorf("SHA-256 mismatch for %s", assetName)
	}
	return actual, nil
}

func closeUpdateResponse(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
}

func readBoundedUpdateResponse(resp *http.Response, maxBytes int64) ([]byte, error) {
	if resp == nil {
		return nil, fmt.Errorf("update response is empty")
	}
	if maxBytes <= 0 {
		closeUpdateResponse(resp)
		return nil, fmt.Errorf("invalid update response size limit: %d", maxBytes)
	}
	if resp.Body == nil {
		return nil, fmt.Errorf("update response body is empty")
	}
	if resp.ContentLength > maxBytes {
		closeUpdateResponse(resp)
		return nil, fmt.Errorf("download exceeded %d-byte limit", maxBytes)
	}
	defer closeUpdateResponse(resp)
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("download exceeded %d-byte limit", maxBytes)
	}
	return data, nil
}

func isNewerVersion(current, latest string) bool {
	if current == "dev" || current == "" {
		return false
	}
	comparison, valid := compareSemanticVersions(current, latest)
	return valid && comparison < 0
}

// IsDesktopVersionCheckable reports whether a binary carries a published
// semantic version suitable for release metadata checks. Local and ad-hoc
// development builds must not query the public updater or present themselves
// as an installable release.
func IsDesktopVersionCheckable(version string) bool {
	_, valid := parseSemanticVersion(version)
	return valid
}

type semanticVersion struct {
	core       [3]int
	prerelease []string
}

// compareSemanticVersions returns -1 when current is older than latest, 0
// when equal, and 1 when current is newer. It implements the precedence rules
// needed by the stable and preview release channels without accepting a
// malformed or arbitrary tag as a valid update.
func compareSemanticVersions(current, latest string) (int, bool) {
	currentVersion, currentOK := parseSemanticVersion(current)
	latestVersion, latestOK := parseSemanticVersion(latest)
	if !currentOK || !latestOK {
		return 0, false
	}
	for index := range currentVersion.core {
		if currentVersion.core[index] < latestVersion.core[index] {
			return -1, true
		}
		if currentVersion.core[index] > latestVersion.core[index] {
			return 1, true
		}
	}
	if len(currentVersion.prerelease) == 0 && len(latestVersion.prerelease) > 0 {
		return 1, true
	}
	if len(currentVersion.prerelease) > 0 && len(latestVersion.prerelease) == 0 {
		return -1, true
	}
	for index := 0; index < len(currentVersion.prerelease) && index < len(latestVersion.prerelease); index++ {
		comparison := comparePrereleaseIdentifiers(currentVersion.prerelease[index], latestVersion.prerelease[index])
		if comparison != 0 {
			return comparison, true
		}
	}
	if len(currentVersion.prerelease) < len(latestVersion.prerelease) {
		return -1, true
	}
	if len(currentVersion.prerelease) > len(latestVersion.prerelease) {
		return 1, true
	}
	return 0, true
}

func parseSemanticVersion(raw string) (semanticVersion, bool) {
	clean := strings.TrimSpace(strings.TrimPrefix(raw, "v"))
	if clean == "" {
		return semanticVersion{}, false
	}
	if plus := strings.IndexByte(clean, '+'); plus >= 0 {
		clean = clean[:plus]
	}
	coreText := clean
	prereleaseText := ""
	if dash := strings.IndexByte(clean, '-'); dash >= 0 {
		coreText = clean[:dash]
		prereleaseText = clean[dash+1:]
	}
	coreParts := strings.Split(coreText, ".")
	if len(coreParts) == 0 || len(coreParts) > 3 {
		return semanticVersion{}, false
	}
	version := semanticVersion{}
	for index, part := range coreParts {
		if part == "" {
			return semanticVersion{}, false
		}
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 {
			return semanticVersion{}, false
		}
		version.core[index] = number
	}
	if prereleaseText == "" {
		return version, true
	}
	for _, identifier := range strings.Split(prereleaseText, ".") {
		if identifier == "" {
			return semanticVersion{}, false
		}
		for _, character := range identifier {
			if !(character == '-' || character >= '0' && character <= '9' || character >= 'A' && character <= 'Z' || character >= 'a' && character <= 'z') {
				return semanticVersion{}, false
			}
		}
		if len(identifier) > 1 && identifier[0] == '0' && isNumericIdentifier(identifier) {
			return semanticVersion{}, false
		}
		version.prerelease = append(version.prerelease, identifier)
	}
	return version, true
}

func comparePrereleaseIdentifiers(current, latest string) int {
	currentNumeric := isNumericIdentifier(current)
	latestNumeric := isNumericIdentifier(latest)
	if currentNumeric && latestNumeric {
		if len(current) < len(latest) {
			return -1
		}
		if len(current) > len(latest) {
			return 1
		}
		return strings.Compare(current, latest)
	}
	if currentNumeric && !latestNumeric {
		return -1
	}
	if !currentNumeric && latestNumeric {
		return 1
	}
	return strings.Compare(current, latest)
}

func isNumericIdentifier(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func isPreviewReleaseTag(tag string) bool {
	version, valid := parseSemanticVersion(tag)
	return valid && len(version.prerelease) == 2 && version.prerelease[0] == "preview" && isNumericIdentifier(version.prerelease[1])
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func verifyBinaryMagic(filePath string, targetOS string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	header := make([]byte, 4)
	n, err := f.Read(header)
	if err != nil || n < 4 {
		return fmt.Errorf("failed to read binary header")
	}

	switch targetOS {
	case "linux":
		// ELF header: 0x7F 'E' 'L' 'F'
		if header[0] != 0x7f || header[1] != 'E' || header[2] != 'L' || header[3] != 'F' {
			return fmt.Errorf("invalid Linux ELF header (got %02x%02x%02x%02x)", header[0], header[1], header[2], header[3])
		}
	case "darwin":
		// Mach-O headers: 0xfeedfacf (64-bit), 0xfeedface (32-bit), 0xcafebabe (Universal)
		isMachO := (header[0] == 0xcf && header[1] == 0xfa && header[2] == 0xed && header[3] == 0xfe) ||
			(header[0] == 0xfe && header[1] == 0xed && header[2] == 0xfa && header[3] == 0xcf) ||
			(header[0] == 0xca && header[1] == 0xfe && header[2] == 0xba && header[3] == 0xbe)
		if !isMachO {
			return fmt.Errorf("invalid macOS Mach-O header (got %02x%02x%02x%02x)", header[0], header[1], header[2], header[3])
		}
	case "windows":
		// PE header begins with MZ: 'M' 'Z'
		if header[0] != 'M' || header[1] != 'Z' {
			return fmt.Errorf("invalid Windows PE header (got %02x%02x%02x%02x)", header[0], header[1], header[2], header[3])
		}
	}
	return nil
}
