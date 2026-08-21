package updater

import (
	"bufio"
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

// MaxUpdateArtifactBytes bounds the amount of untrusted release data that can
// be written to the temporary update candidate before checksum verification.
// It is intentionally much larger than the current binaries while preventing
// an unexpectedly large response from exhausting local disk space.
const MaxUpdateArtifactBytes int64 = 512 << 20

// BuildUpdatePublicKey is injected into release builds with -ldflags. A
// release binary therefore carries its trust anchor instead of depending on a
// mutable runtime environment. Development builds may use the environment
// fallback documented below.
var BuildUpdatePublicKey string

// CheckLatest queries the official GitHub repository for the latest release.
func CheckLatest(currentVersion string) (*CheckResult, error) {
	apiURL := "https://api.github.com/repos/div197/bob-gemini-free/releases/latest"
	client := &http.Client{Timeout: 8 * time.Second}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create update request: %w", err)
	}
	req.Header.Set("User-Agent", "BOB-Gemini-Free-Updater")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach update server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update server returned HTTP %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release metadata: %w", err)
	}

	latestVer := strings.TrimPrefix(release.TagName, "v")
	currVer := strings.TrimPrefix(currentVersion, "v")

	hasUpdate := isNewerVersion(currVer, latestVer)
	targetAsset := findMatchingAsset(release.Assets, runtime.GOOS, runtime.GOARCH)

	downloadURL := ""
	if targetAsset != nil {
		downloadURL = targetAsset.BrowserDownloadURL
	}
	checksumURL := ""
	if checksumAsset := findAssetByName(release.Assets, "SHA256SUMS"); checksumAsset != nil {
		checksumURL = checksumAsset.BrowserDownloadURL
	}
	signatureURL := ""
	if signatureAsset := findAssetByName(release.Assets, "SHA256SUMS.sig"); signatureAsset != nil {
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

	client := &http.Client{Timeout: 90 * time.Second}
	manifest, err := downloadBytes(client, res.ChecksumURL, 1<<20)
	if err != nil {
		return fmt.Errorf("failed to download checksum manifest: %w", err)
	}
	signature, err := downloadBytes(client, res.SignatureURL, 16<<10)
	if err != nil {
		return fmt.Errorf("failed to download checksum signature: %w", err)
	}
	if err := verifyReleaseManifest(publicKey, manifest, signature); err != nil {
		return fmt.Errorf("release authenticity verification failed: %w", err)
	}

	assetName := ""
	if res.Release != nil {
		if asset := findAssetByURL(res.Release.Assets, res.DownloadURL); asset != nil {
			assetName = asset.Name
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

	if n < 5*1024*1024 {
		return fmt.Errorf("downloaded binary payload is suspiciously small (%d bytes), update aborted for safety", n)
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
	var expectedSuffix string
	if targetOS == "windows" {
		expectedSuffix = fmt.Sprintf("windows-%s.exe", targetArch)
	} else {
		expectedSuffix = fmt.Sprintf("%s-%s", targetOS, targetArch)
	}

	for _, asset := range assets {
		if strings.HasSuffix(asset.Name, expectedSuffix) {
			return &asset
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
	resp, err := client.Get(downloadURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download server returned HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("download exceeded %d-byte limit", maxBytes)
	}
	return data, nil
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
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download server returned HTTP %d", resp.StatusCode)
	}

	file, err := os.Create(destination)
	if err != nil {
		return "", err
	}
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

func isNewerVersion(current, latest string) bool {
	if current == "dev" || current == "" {
		return false
	}
	cParts := parseSemVer(current)
	lParts := parseSemVer(latest)

	for i := 0; i < len(cParts) && i < len(lParts); i++ {
		if lParts[i] > cParts[i] {
			return true
		}
		if lParts[i] < cParts[i] {
			return false
		}
	}
	return len(lParts) > len(cParts)
}

func parseSemVer(v string) []int {
	clean := strings.TrimPrefix(v, "v")
	parts := strings.Split(clean, ".")
	res := make([]int, 0, len(parts))
	for _, p := range parts {
		if idx := strings.IndexAny(p, "-+"); idx != -1 {
			p = p[:idx]
		}
		if num, err := strconv.Atoi(p); err == nil {
			res = append(res, num)
		}
	}
	return res
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

	_, err = io.Copy(out, in)
	return err
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
