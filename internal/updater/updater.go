package updater

import (
	"crypto/sha256"
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
}

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

	return &CheckResult{
		CurrentVersion: currentVersion,
		LatestVersion:  release.TagName,
		HasUpdate:      hasUpdate,
		Release:        &release,
		DownloadURL:    downloadURL,
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
		tempFile, err = os.CreateTemp("", "bob-gemini-free-update-*.tmp")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	client := &http.Client{Timeout: 90 * time.Second}
	dlResp, err := client.Get(res.DownloadURL)
	if err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer dlResp.Body.Close()

	if dlResp.StatusCode != http.StatusOK {
		tempFile.Close()
		return fmt.Errorf("download server returned HTTP %d", dlResp.StatusCode)
	}

	hasher := sha256.New()
	multiWriter := io.MultiWriter(tempFile, hasher)
	n, err := io.Copy(multiWriter, dlResp.Body)
	tempFile.Close()
	if err != nil {
		return fmt.Errorf("failed to save update payload: %w", err)
	}

	if n < 5*1024*1024 {
		return fmt.Errorf("downloaded binary payload is suspiciously small (%d bytes), update aborted for safety", n)
	}

	shaSum := hex.EncodeToString(hasher.Sum(nil))
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
			if err := copyFile(tempPath, exePath); err != nil {
				return fmt.Errorf("failed to replace binary at %s: %w", exePath, err)
			}
			_ = os.Chmod(exePath, 0755)
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
