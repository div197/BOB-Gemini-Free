package updater

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestStageDesktopUpdateVerifiesSignedMacArchiveBeforeReplacement(t *testing.T) {
	const assetName = "bob-gemini-free-wails-macos-universal.zip"
	archivePayload := macOSUpdateArchive(t)
	manifest, signature, publicKey := signedManifestFixture(t, assetName, archivePayload)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/manifest"):
			_, _ = w.Write(manifest)
		case strings.HasSuffix(r.URL.Path, "/signature"):
			_, _ = w.Write(signature)
		case strings.HasSuffix(r.URL.Path, ".zip"):
			_, _ = w.Write(archivePayload)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	previousKey := BuildUpdatePublicKey
	BuildUpdatePublicKey = base64.StdEncoding.EncodeToString(publicKey)
	t.Cleanup(func() { BuildUpdatePublicKey = previousKey })
	previousVerifier := verifyDesktopBundleSignature
	verifyDesktopBundleSignature = func(string) error { return nil }
	t.Cleanup(func() { verifyDesktopBundleSignature = previousVerifier })

	client := &http.Client{Transport: rewriteOfficialUpdateTransport{base: server.Client().Transport, destination: server.URL}}
	target := filepath.Join(t.TempDir(), desktopAppBundleName)
	if err := os.MkdirAll(filepath.Join(target, "Contents"), 0755); err != nil {
		t.Fatalf("create target app: %v", err)
	}
	if err := os.WriteFile(filepath.Join(target, "Contents", "old-marker"), []byte("old"), 0644); err != nil {
		t.Fatalf("write target marker: %v", err)
	}

	result := &DesktopCheckResult{
		CurrentVersion:    "v0.1.7",
		LatestVersion:     "v0.1.8",
		HasUpdate:         true,
		AssetAvailable:    true,
		AssetName:         assetName,
		AssetSize:         int64(len(archivePayload)),
		DownloadURL:       "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.1.8/" + assetName,
		ChecksumURL:       "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.1.8/manifest",
		SignatureURL:      "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.1.8/signature",
		ManifestAvailable: true,
		Channel:           DesktopChannelStable,
	}

	plan, err := stageDesktopUpdate(client, result, target, "darwin", "arm64")
	if err != nil {
		t.Fatalf("stageDesktopUpdate: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(plan.StageDir) })
	if _, err := os.Stat(plan.CandidatePath); err != nil {
		t.Fatalf("staged candidate missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "Contents", "old-marker")); err != nil {
		t.Fatalf("staging modified installed app: %v", err)
	}
	if info, err := os.Stat(plan.PlanPath); err != nil {
		t.Fatalf("update plan missing: %v", err)
	} else if info.Mode().Perm() != 0600 {
		t.Fatalf("update plan permissions = %o, want 600", info.Mode().Perm())
	}
}

func TestDesktopStagingDirectoryErrorExplainsReadOnlyMacInstall(t *testing.T) {
	err := desktopStagingDirectoryError(syscall.EROFS)
	if !strings.Contains(err.Error(), "move BOB Gemini Free.app to Applications") {
		t.Fatalf("read-only staging error = %q", err)
	}
}

func TestStageDesktopUpdateRefusesUnsignedManifest(t *testing.T) {
	result := &DesktopCheckResult{
		CurrentVersion: "v0.1.7",
		LatestVersion:  "v0.1.8",
		HasUpdate:      true,
		AssetAvailable: true,
		AssetName:      "bob-gemini-free-wails-macos-universal.zip",
		AssetSize:      1,
		DownloadURL:    "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.1.8/app.zip",
		Channel:        DesktopChannelStable,
	}
	if _, err := stageDesktopUpdate(http.DefaultClient, result, filepath.Join(t.TempDir(), desktopAppBundleName), "darwin", "arm64"); err == nil {
		t.Fatal("unsigned desktop update was accepted")
	}
}

func TestStageDesktopUpdateDoesNotUseMutableEnvironmentKey(t *testing.T) {
	archivePayload := macOSUpdateArchive(t)
	manifest, signature, publicKey := signedManifestFixture(t, "bob-gemini-free-wails-macos-universal.zip", archivePayload)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/manifest"):
			_, _ = w.Write(manifest)
		case strings.HasSuffix(r.URL.Path, "/signature"):
			_, _ = w.Write(signature)
		default:
			_, _ = w.Write(archivePayload)
		}
	}))
	defer server.Close()

	previousBuildKey := BuildUpdatePublicKey
	previousEnvironmentKey := os.Getenv("BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY")
	BuildUpdatePublicKey = ""
	_ = os.Setenv("BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY", base64.StdEncoding.EncodeToString(publicKey))
	t.Cleanup(func() {
		BuildUpdatePublicKey = previousBuildKey
		_ = os.Setenv("BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY", previousEnvironmentKey)
	})

	result := &DesktopCheckResult{
		CurrentVersion:    "v0.1.7",
		LatestVersion:     "v0.1.8",
		HasUpdate:         true,
		AssetAvailable:    true,
		AssetName:         "bob-gemini-free-wails-macos-universal.zip",
		AssetSize:         int64(len(archivePayload)),
		DownloadURL:       "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.1.8/app.zip",
		ChecksumURL:       "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.1.8/manifest",
		SignatureURL:      "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.1.8/signature",
		ManifestAvailable: true,
		Channel:           DesktopChannelStable,
	}
	client := &http.Client{Transport: rewriteOfficialUpdateTransport{base: server.Client().Transport, destination: server.URL}}
	if _, err := stageDesktopUpdate(client, result, filepath.Join(t.TempDir(), desktopAppBundleName), "darwin", "arm64"); err == nil || !strings.Contains(err.Error(), "embedded desktop update public key") {
		t.Fatalf("mutable environment key was accepted: %v", err)
	}
}

func TestStageDesktopUpdateRejectsDeclaredSizeMismatch(t *testing.T) {
	archivePayload := macOSUpdateArchive(t)
	manifest, signature, publicKey := signedManifestFixture(t, "bob-gemini-free-wails-macos-universal.zip", archivePayload)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/manifest"):
			_, _ = w.Write(manifest)
		case strings.HasSuffix(r.URL.Path, "/signature"):
			_, _ = w.Write(signature)
		default:
			_, _ = w.Write(archivePayload)
		}
	}))
	defer server.Close()

	previousKey := BuildUpdatePublicKey
	BuildUpdatePublicKey = base64.StdEncoding.EncodeToString(publicKey)
	t.Cleanup(func() { BuildUpdatePublicKey = previousKey })
	previousVerifier := verifyDesktopBundleSignature
	verifyDesktopBundleSignature = func(string) error { return nil }
	t.Cleanup(func() { verifyDesktopBundleSignature = previousVerifier })
	target := filepath.Join(t.TempDir(), desktopAppBundleName)
	if err := os.MkdirAll(filepath.Join(target, "Contents"), 0755); err != nil {
		t.Fatalf("create target app: %v", err)
	}

	result := &DesktopCheckResult{
		CurrentVersion:    "v0.1.7",
		LatestVersion:     "v0.1.8",
		HasUpdate:         true,
		AssetAvailable:    true,
		AssetName:         "bob-gemini-free-wails-macos-universal.zip",
		AssetSize:         int64(len(archivePayload)) + 1,
		DownloadURL:       "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.1.8/app.zip",
		ChecksumURL:       "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.1.8/manifest",
		SignatureURL:      "https://github.com/div197/BOB-Gemini-Free/releases/download/v0.1.8/signature",
		ManifestAvailable: true,
		Channel:           DesktopChannelStable,
	}
	client := &http.Client{Transport: rewriteOfficialUpdateTransport{base: server.Client().Transport, destination: server.URL}}
	if _, err := stageDesktopUpdate(client, result, target, "darwin", "arm64"); err == nil || !strings.Contains(err.Error(), "size changed") {
		t.Fatalf("declared size mismatch was accepted: %v", err)
	}
}

func TestSafeArchiveNameRejectsTraversal(t *testing.T) {
	for _, name := range []string{"../escape", "/absolute/path", "bob-gemini-free-wails.app/../../escape"} {
		if _, err := safeArchiveName(name); err == nil {
			t.Errorf("safeArchiveName(%q) accepted traversal", name)
		}
	}
}

func TestExtractMacOSAppAcceptsLegacyBundleForMigration(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "legacy.zip")
	if err := os.WriteFile(archivePath, macOSBundleArchive(t, legacyDesktopAppBundleName, legacyDesktopAppBinaryName), 0600); err != nil {
		t.Fatalf("write legacy archive: %v", err)
	}
	destination := t.TempDir()
	candidate, err := extractMacOSApp(archivePath, destination)
	if err != nil {
		t.Fatalf("extractMacOSApp: %v", err)
	}
	if filepath.Base(candidate) != legacyDesktopAppBundleName {
		t.Fatalf("candidate = %q, want legacy bundle for migration", candidate)
	}
	if _, err := findDesktopBundleBinary(candidate); err != nil {
		t.Fatalf("legacy binary not found: %v", err)
	}
}

func TestReplaceAndConfirmDesktopUpdateCommitsCandidate(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "bob-gemini-free-wails.exe")
	stage := filepath.Join(root, ".bob-gemini-free-update-test")
	if err := os.Mkdir(stage, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}
	candidate := filepath.Join(stage, "candidate.exe")
	if err := os.WriteFile(candidate, []byte("new"), 0755); err != nil {
		t.Fatal(err)
	}
	plan := &DesktopUpdatePlan{
		PlanPath:         filepath.Join(stage, "update-plan.json"),
		StageDir:         stage,
		InstallTarget:    target,
		CandidatePath:    candidate,
		BackupPath:       filepath.Join(stage, "rollback-backup"),
		ConfirmationPath: filepath.Join(stage, ".bob-gemini-update-confirm"),
		TargetVersion:    "v0.1.8",
		AssetName:        "bob-gemini-free-wails-windows-amd64.exe",
		Channel:          DesktopChannelStable,
		TargetOS:         "windows",
	}

	previousLauncher := launchUpdatedDesktop
	previousPoll := desktopUpdatePollInterval
	previousTimeout := desktopUpdateConfirmationTimeout
	launchUpdatedDesktop = func(_, confirmation string) (*os.Process, error) {
		return nil, ConfirmDesktopUpdate(confirmation)
	}
	desktopUpdatePollInterval = time.Millisecond
	desktopUpdateConfirmationTimeout = 100 * time.Millisecond
	t.Cleanup(func() {
		launchUpdatedDesktop = previousLauncher
		desktopUpdatePollInterval = previousPoll
		desktopUpdateConfirmationTimeout = previousTimeout
		_ = os.RemoveAll(stage)
	})

	if err := replaceAndConfirmDesktopUpdate(plan); err != nil {
		t.Fatalf("replaceAndConfirmDesktopUpdate: %v", err)
	}
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "new" {
		t.Fatalf("installed content = %q, want new", content)
	}
	if _, err := os.Stat(stage); !os.IsNotExist(err) {
		t.Fatalf("staging directory still exists after successful update: %v", err)
	}
}

func TestReplaceAndConfirmDesktopUpdateRollsBackWhenCandidateDoesNotConfirm(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "bob-gemini-free-wails.exe")
	stage := filepath.Join(root, ".bob-gemini-free-update-rollback-test")
	if err := os.Mkdir(stage, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}
	candidate := filepath.Join(stage, "candidate.exe")
	if err := os.WriteFile(candidate, []byte("bad-new"), 0755); err != nil {
		t.Fatal(err)
	}
	plan := &DesktopUpdatePlan{
		PlanPath:         filepath.Join(stage, "update-plan.json"),
		StageDir:         stage,
		InstallTarget:    target,
		CandidatePath:    candidate,
		BackupPath:       filepath.Join(stage, "rollback-backup"),
		ConfirmationPath: filepath.Join(stage, ".bob-gemini-update-confirm"),
		TargetVersion:    "v0.1.8",
		AssetName:        "bob-gemini-free-wails-windows-amd64.exe",
		Channel:          DesktopChannelStable,
		TargetOS:         "windows",
	}

	previousLauncher := launchUpdatedDesktop
	previousPoll := desktopUpdatePollInterval
	previousTimeout := desktopUpdateConfirmationTimeout
	launchUpdatedDesktop = func(_, _ string) (*os.Process, error) { return nil, nil }
	desktopUpdatePollInterval = time.Millisecond
	desktopUpdateConfirmationTimeout = 5 * time.Millisecond
	t.Cleanup(func() {
		launchUpdatedDesktop = previousLauncher
		desktopUpdatePollInterval = previousPoll
		desktopUpdateConfirmationTimeout = previousTimeout
		_ = os.RemoveAll(stage)
	})

	if err := replaceAndConfirmDesktopUpdate(plan); err == nil {
		t.Fatal("unconfirmed update was accepted")
	}
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "old" {
		t.Fatalf("rolled-back content = %q, want old", content)
	}
}

func TestConfirmDesktopUpdateRejectsUnexpectedPath(t *testing.T) {
	if err := ConfirmDesktopUpdate(filepath.Join(t.TempDir(), "confirmation")); err == nil {
		t.Fatal("confirmation was written to an unexpected path")
	}
}

func macOSUpdateArchive(t *testing.T) []byte {
	t.Helper()
	return macOSBundleArchive(t, desktopAppBundleName, desktopAppBinaryName)
}

func macOSBundleArchive(t *testing.T, bundleName, binaryName string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	entries := []struct {
		name string
		data []byte
		mode os.FileMode
	}{
		{name: bundleName + "/Contents/Info.plist", data: []byte("plist"), mode: 0644},
		{name: bundleName + "/Contents/MacOS/" + binaryName, data: []byte{0xcf, 0xfa, 0xed, 0xfe, 0x01}, mode: 0755},
	}
	for _, entry := range entries {
		header := &zip.FileHeader{Name: entry.name, Method: zip.Store}
		header.SetMode(entry.mode)
		writer, err := archive.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := writer.Write(entry.data); err != nil {
			t.Fatal(err)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

type rewriteOfficialUpdateTransport struct {
	base        http.RoundTripper
	destination string
}

func (t rewriteOfficialUpdateTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	clone := request.Clone(request.Context())
	destination, err := http.NewRequest(http.MethodGet, t.destination, nil)
	if err != nil {
		return nil, err
	}
	clone.URL.Scheme = destination.URL.Scheme
	clone.URL.Host = destination.URL.Host
	if t.base == nil {
		t.base = http.DefaultTransport
	}
	return t.base.RoundTrip(clone)
}
