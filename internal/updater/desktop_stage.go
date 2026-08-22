package updater

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	// These are the public product identifiers. The legacy names remain
	// accepted only so a future signed build can migrate an older Preview 2
	// installation without requiring a delete/reinstall cycle.
	desktopAppBundleName       = "BOB Gemini Free.app"
	legacyDesktopAppBundleName = "bob-gemini-free-wails.app"
	desktopAppBinaryName       = "bob-gemini-free"
	legacyDesktopAppBinaryName = "bob-gemini-free-wails"

	// MaxDesktopArchiveBytes bounds both the downloaded archive and the total
	// uncompressed ZIP payload. The current desktop artifacts are much smaller.
	MaxDesktopArchiveBytes int64 = 512 << 20
)

// DesktopUpdatePlan describes a verified candidate which has not yet replaced
// the installed application. All paths are created by the updater itself and
// are kept on the same filesystem as the install target for transactional
// rename/rollback behavior.
type DesktopUpdatePlan struct {
	PlanPath         string `json:"plan_path"`
	StageDir         string `json:"stage_dir"`
	InstallTarget    string `json:"install_target"`
	CandidatePath    string `json:"candidate_path"`
	BackupPath       string `json:"backup_path"`
	ConfirmationPath string `json:"confirmation_path"`
	HelperPath       string `json:"helper_path,omitempty"`
	CurrentVersion   string `json:"current_version"`
	TargetVersion    string `json:"target_version"`
	AssetName        string `json:"asset_name"`
	Channel          string `json:"channel"`
	TargetOS         string `json:"target_os"`
}

// StageDesktopUpdate downloads and verifies a native package but does not
// replace or execute it. It is intentionally separate from StartDesktopUpdate
// so a caller can present the verified release details before restarting.
func StageDesktopUpdate(result *DesktopCheckResult) (*DesktopUpdatePlan, error) {
	client := &http.Client{Timeout: 90 * time.Second}
	return stageDesktopUpdate(client, result, "", runtime.GOOS, runtime.GOARCH)
}

func stageDesktopUpdate(client *http.Client, result *DesktopCheckResult, targetPath, targetOS, targetArch string) (*DesktopUpdatePlan, error) {
	if client == nil {
		return nil, fmt.Errorf("desktop update staging requires an HTTP client")
	}
	if result == nil {
		return nil, fmt.Errorf("desktop update staging requires a release result")
	}
	if !result.HasUpdate {
		return nil, fmt.Errorf("desktop update staging was requested without a newer release")
	}
	if !result.AssetAvailable || result.AssetName == "" || result.DownloadURL == "" {
		return nil, fmt.Errorf("release %s has no compatible desktop asset", result.LatestVersion)
	}
	if !result.ManifestAvailable || result.ChecksumURL == "" || result.SignatureURL == "" {
		return nil, fmt.Errorf("release %s has no signed desktop manifest; refusing automatic installation", result.LatestVersion)
	}
	if !isOfficialGitHubURL(result.DownloadURL) || !isOfficialGitHubURL(result.ChecksumURL) || !isOfficialGitHubURL(result.SignatureURL) {
		return nil, fmt.Errorf("desktop update sources are not official GitHub URLs")
	}
	if !desktopAssetNameMatches(result.AssetName, targetOS, targetArch) {
		return nil, fmt.Errorf("desktop asset %q does not match %s/%s", result.AssetName, targetOS, targetArch)
	}
	if result.AssetSize > MaxDesktopArchiveBytes {
		return nil, fmt.Errorf("desktop asset %s exceeds the %d-byte safety limit", result.AssetName, MaxDesktopArchiveBytes)
	}
	if strings.TrimSpace(result.Channel) == "" {
		return nil, fmt.Errorf("desktop update release channel is missing; refusing automatic installation")
	}
	publicKey, err := embeddedUpdatePublicKey()
	if err != nil {
		return nil, err
	}

	if targetPath == "" {
		targetPath, err = currentDesktopInstallTarget(targetOS)
		if err != nil {
			return nil, err
		}
	}
	targetPath, err = filepath.Abs(targetPath)
	if err != nil {
		return nil, fmt.Errorf("resolve desktop install target: %w", err)
	}
	if err := validateInstallTarget(targetPath, targetOS); err != nil {
		return nil, err
	}

	stageDir, err := os.MkdirTemp(filepath.Dir(targetPath), ".bob-gemini-free-update-")
	if err != nil {
		return nil, fmt.Errorf("create same-filesystem update staging directory: %w", err)
	}
	keepStage := false
	defer func() {
		if !keepStage {
			_ = os.RemoveAll(stageDir)
		}
	}()

	manifest, err := downloadBytes(client, result.ChecksumURL, 1<<20)
	if err != nil {
		return nil, fmt.Errorf("download desktop checksum manifest: %w", err)
	}
	signature, err := downloadBytes(client, result.SignatureURL, 16<<10)
	if err != nil {
		return nil, fmt.Errorf("download desktop checksum signature: %w", err)
	}
	if err := verifyReleaseManifest(publicKey, manifest, signature); err != nil {
		return nil, fmt.Errorf("desktop release authenticity verification failed: %w", err)
	}

	downloadPath := filepath.Join(stageDir, "verified-download")
	if _, err := downloadVerifiedArtifactLimited(client, result.DownloadURL, downloadPath, result.AssetName, manifest, signature, publicKey, MaxDesktopArchiveBytes); err != nil {
		return nil, fmt.Errorf("verified desktop package download failed: %w", err)
	}
	if result.AssetSize <= 0 {
		return nil, fmt.Errorf("desktop release asset %s has no trusted positive size", result.AssetName)
	}
	if info, statErr := os.Stat(downloadPath); statErr != nil {
		return nil, fmt.Errorf("inspect verified desktop package: %w", statErr)
	} else if info.Size() != result.AssetSize {
		return nil, fmt.Errorf("desktop package size changed during download: got %d bytes, release metadata declares %d", info.Size(), result.AssetSize)
	}

	candidatePath, err := prepareDesktopCandidate(downloadPath, stageDir, result.AssetName, targetOS)
	if err != nil {
		return nil, err
	}

	plan := &DesktopUpdatePlan{
		PlanPath:         filepath.Join(stageDir, "update-plan.json"),
		StageDir:         stageDir,
		InstallTarget:    targetPath,
		CandidatePath:    candidatePath,
		BackupPath:       filepath.Join(stageDir, "rollback-backup"),
		ConfirmationPath: filepath.Join(stageDir, ".bob-gemini-update-confirm"),
		CurrentVersion:   result.CurrentVersion,
		TargetVersion:    result.LatestVersion,
		AssetName:        result.AssetName,
		Channel:          result.Channel,
		TargetOS:         targetOS,
	}
	if err := writeDesktopUpdatePlan(plan); err != nil {
		return nil, err
	}
	keepStage = true
	return plan, nil
}

func desktopAssetNameMatches(assetName, targetOS, targetArch string) bool {
	for _, candidate := range desktopAssetNames(targetOS, targetArch) {
		if assetName == candidate {
			return true
		}
	}
	return false
}

func currentDesktopInstallTarget(targetOS string) (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("determine current desktop executable: %w", err)
	}
	if resolved, resolveErr := filepath.EvalSymlinks(executable); resolveErr == nil {
		executable = resolved
	}

	if targetOS == "darwin" {
		const marker = ".app/Contents/MacOS/"
		index := strings.Index(filepath.ToSlash(executable), marker)
		if index < 0 {
			return "", fmt.Errorf("current executable is not inside a macOS app bundle: %s", executable)
		}
		return filepath.FromSlash(filepath.ToSlash(executable)[:index+len(".app")]), nil
	}
	if targetOS == "windows" {
		return executable, nil
	}
	return "", fmt.Errorf("automatic native updates are not implemented for %s", targetOS)
}

func validateInstallTarget(targetPath, targetOS string) error {
	info, err := os.Lstat(targetPath)
	if err != nil {
		return fmt.Errorf("inspect desktop install target %s: %w", targetPath, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to update a symlinked desktop install target: %s", targetPath)
	}
	if targetOS == "darwin" && !info.IsDir() {
		return fmt.Errorf("macOS desktop install target is not an app directory: %s", targetPath)
	}
	if targetOS == "windows" && info.IsDir() {
		return fmt.Errorf("Windows desktop install target is a directory: %s", targetPath)
	}
	return nil
}

func prepareDesktopCandidate(downloadPath, stageDir, assetName, targetOS string) (string, error) {
	switch targetOS {
	case "darwin":
		if !strings.HasSuffix(assetName, ".zip") {
			return "", fmt.Errorf("unsupported macOS desktop update asset: %s", assetName)
		}
		candidateRoot := filepath.Join(stageDir, "candidate")
		if err := os.Mkdir(candidateRoot, 0700); err != nil {
			return "", fmt.Errorf("create macOS candidate directory: %w", err)
		}
		candidatePath, err := extractMacOSApp(downloadPath, candidateRoot)
		if err != nil {
			return "", err
		}
		binaryPath, err := findDesktopBundleBinary(candidatePath)
		if err != nil {
			return "", fmt.Errorf("locate staged macOS executable: %w", err)
		}
		if err := verifyBinaryMagic(binaryPath, "darwin"); err != nil {
			return "", fmt.Errorf("verify staged macOS executable: %w", err)
		}
		if err := verifyDesktopBundleSignature(candidatePath); err != nil {
			return "", fmt.Errorf("verify staged macOS code signature: %w", err)
		}
		return candidatePath, nil
	case "windows":
		if !strings.HasSuffix(strings.ToLower(assetName), ".exe") {
			return "", fmt.Errorf("unsupported Windows desktop update asset: %s", assetName)
		}
		if err := verifyBinaryMagic(downloadPath, "windows"); err != nil {
			return "", fmt.Errorf("verify staged Windows executable: %w", err)
		}
		candidatePath := filepath.Join(stageDir, assetName)
		if err := os.Rename(downloadPath, candidatePath); err != nil {
			return "", fmt.Errorf("move verified Windows candidate into staging: %w", err)
		}
		return candidatePath, nil
	default:
		return "", fmt.Errorf("automatic native updates are not implemented for %s", targetOS)
	}
}

func extractMacOSApp(zipPath, destination string) (string, error) {
	archive, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("open verified macOS archive: %w", err)
	}
	defer archive.Close()

	var totalUncompressed uint64
	var candidatePath string
	for _, entry := range archive.File {
		if entry.FileInfo().Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("macOS update archive contains a symlink: %s", entry.Name)
		}
		if entry.UncompressedSize64 > uint64(MaxDesktopArchiveBytes) {
			return "", fmt.Errorf("macOS update archive entry is too large: %s", entry.Name)
		}
		totalUncompressed += entry.UncompressedSize64
		if totalUncompressed > uint64(MaxDesktopArchiveBytes) {
			return "", fmt.Errorf("macOS update archive expands beyond the safety limit")
		}

		cleanName, err := safeArchiveName(entry.Name)
		if err != nil {
			return "", err
		}
		parts := strings.Split(filepath.ToSlash(cleanName), "/")
		if len(parts) == 0 || !isDesktopBundleRoot(parts[0]) {
			return "", fmt.Errorf("macOS update archive contains unexpected root path: %s", entry.Name)
		}
		if candidatePath == "" {
			candidatePath = filepath.Join(destination, filepath.FromSlash(parts[0]))
		} else if candidatePath != filepath.Join(destination, filepath.FromSlash(parts[0])) {
			return "", fmt.Errorf("macOS update archive contains multiple application bundles")
		}
		target := filepath.Join(destination, filepath.FromSlash(cleanName))
		if err := ensurePathWithin(destination, target); err != nil {
			return "", err
		}
		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return "", fmt.Errorf("create macOS archive directory %s: %w", entry.Name, err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return "", fmt.Errorf("create macOS archive parent %s: %w", entry.Name, err)
		}
		reader, err := entry.Open()
		if err != nil {
			return "", fmt.Errorf("open macOS archive entry %s: %w", entry.Name, err)
		}
		mode := entry.Mode().Perm()
		if mode == 0 {
			mode = 0644
		}
		file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
		if err == nil {
			_, err = io.Copy(file, io.LimitReader(reader, int64(entry.UncompressedSize64)+1))
			closeErr := file.Close()
			if err == nil {
				err = closeErr
			}
		}
		_ = reader.Close()
		if err != nil {
			return "", fmt.Errorf("extract macOS archive entry %s: %w", entry.Name, err)
		}
	}

	if candidatePath == "" {
		return "", fmt.Errorf("macOS update archive did not contain a supported BOB Gemini Free app bundle")
	}
	if info, err := os.Stat(candidatePath); err != nil || !info.IsDir() {
		return "", fmt.Errorf("macOS update archive did not contain an application bundle")
	}
	return candidatePath, nil
}

func isDesktopBundleRoot(root string) bool {
	return root == desktopAppBundleName || root == legacyDesktopAppBundleName
}

func findDesktopBundleBinary(appPath string) (string, error) {
	for _, name := range []string{desktopAppBinaryName, legacyDesktopAppBinaryName} {
		candidate := filepath.Join(appPath, "Contents", "MacOS", name)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("bundle contains neither %s nor %s", desktopAppBinaryName, legacyDesktopAppBinaryName)
}

func safeArchiveName(name string) (string, error) {
	name = filepath.ToSlash(strings.TrimSpace(name))
	if name == "" || strings.HasPrefix(name, "/") || strings.ContainsRune(name, '\x00') {
		return "", fmt.Errorf("unsafe macOS archive path: %q", name)
	}
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(name)))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../") {
		return "", fmt.Errorf("macOS archive path escapes candidate directory: %q", name)
	}
	return clean, nil
}

func ensurePathWithin(root, candidate string) error {
	relative, err := filepath.Rel(root, candidate)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) || filepath.IsAbs(relative) {
		return fmt.Errorf("archive path escapes staging directory: %s", candidate)
	}
	return nil
}

var verifyDesktopBundleSignature = verifyMacOSBundleSignature

func verifyMacOSBundleSignature(appPath string) error {
	if runtime.GOOS != "darwin" {
		return nil
	}
	command := exec.Command("codesign", "--verify", "--deep", "--strict", appPath)
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("codesign verification failed: %w (%s)", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func writeDesktopUpdatePlan(plan *DesktopUpdatePlan) error {
	if plan == nil || plan.PlanPath == "" {
		return fmt.Errorf("desktop update plan is incomplete")
	}
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("encode desktop update plan: %w", err)
	}
	temporary := plan.PlanPath + ".tmp"
	if err := os.WriteFile(temporary, append(data, '\n'), 0600); err != nil {
		return fmt.Errorf("write desktop update plan: %w", err)
	}
	if err := os.Rename(temporary, plan.PlanPath); err != nil {
		_ = os.Remove(temporary)
		return fmt.Errorf("commit desktop update plan: %w", err)
	}
	return nil
}
