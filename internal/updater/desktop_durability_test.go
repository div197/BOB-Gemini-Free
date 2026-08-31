package updater

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteAtomicDesktopUpdateFileFlushesAndLeavesNoTemporaryFile(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "update-state.json")
	if err := os.WriteFile(target, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}

	previousSync := syncDesktopUpdateDirectory
	var syncedPath string
	syncDesktopUpdateDirectory = func(path string) error {
		syncedPath = path
		return nil
	}
	t.Cleanup(func() { syncDesktopUpdateDirectory = previousSync })

	if err := writeAtomicDesktopUpdateFile(target, []byte("new\n"), 0600); err != nil {
		t.Fatalf("writeAtomicDesktopUpdateFile: %v", err)
	}
	if syncedPath != root {
		t.Fatalf("synced directory = %q, want %q", syncedPath, root)
	}
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "new\n" {
		t.Fatalf("committed content = %q, want new content", content)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("committed mode = %o, want 600", info.Mode().Perm())
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".update-state.json.tmp-") {
			t.Fatalf("temporary updater file remains: %s", entry.Name())
		}
	}
}

func TestReplaceAndConfirmDesktopUpdateRestoresAfterActivationSyncFailure(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "bob-gemini-free-wails.exe")
	stage := filepath.Join(root, ".bob-gemini-free-update-sync-failure")
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
		TargetVersion:    "v0.2.0",
		AssetName:        "bob-gemini-free-wails-windows-amd64.exe",
		Channel:          DesktopChannelStable,
		TargetOS:         "windows",
	}

	previousSync := syncDesktopUpdateDirectory
	calls := 0
	syncDesktopUpdateDirectory = func(string) error {
		calls++
		if calls == 2 {
			return errors.New("injected directory sync failure")
		}
		return nil
	}
	t.Cleanup(func() { syncDesktopUpdateDirectory = previousSync })

	err := replaceAndConfirmDesktopUpdate(plan)
	if err == nil || !strings.Contains(err.Error(), "persist verified desktop candidate activation") {
		t.Fatalf("sync failure result = %v", err)
	}
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "old" {
		t.Fatalf("restored content = %q, want old", content)
	}
	if _, err := os.Stat(stage); !os.IsNotExist(err) {
		t.Fatalf("staging directory remains after rollback: %v", err)
	}
	if calls < 4 {
		t.Fatalf("directory sync calls = %d, want activation failure and rollback flushes", calls)
	}
}
