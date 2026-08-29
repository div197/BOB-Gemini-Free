package updater

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecoverDesktopUpdateRestoresInstallWhenPowerLossRemovedTarget(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "bob-gemini-free-wails.exe")
	stage := filepath.Join(root, ".bob-gemini-free-update-recover-missing")
	plan := recoveryPlan(target, stage)
	if err := os.Mkdir(stage, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(plan.BackupPath, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(plan.CandidatePath, []byte("new"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := writeDesktopUpdatePlan(plan); err != nil {
		t.Fatal(err)
	}

	if err := recoverDesktopUpdate(target, "windows", ""); err != nil {
		t.Fatalf("recoverDesktopUpdate: %v", err)
	}
	content, err := os.ReadFile(target)
	if err != nil || string(content) != "old" {
		t.Fatalf("restored install = %q, error = %v", content, err)
	}
	if _, err := os.Stat(stage); !os.IsNotExist(err) {
		t.Fatalf("staging directory remains after recovery: %v", err)
	}
	if _, err := os.Stat(failureRecordPath(plan)); err != nil {
		t.Fatalf("recovery failure record missing: %v", err)
	}
}

func TestRecoverDesktopUpdateRollsBackUnconfirmedCandidate(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "bob-gemini-free-wails.exe")
	stage := filepath.Join(root, ".bob-gemini-free-update-recover-unconfirmed")
	plan := recoveryPlan(target, stage)
	if err := os.Mkdir(stage, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("new"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(plan.BackupPath, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := writeDesktopUpdatePlan(plan); err != nil {
		t.Fatal(err)
	}

	if err := recoverDesktopUpdate(target, "windows", ""); err != nil {
		t.Fatalf("recoverDesktopUpdate: %v", err)
	}
	content, err := os.ReadFile(target)
	if err != nil || string(content) != "old" {
		t.Fatalf("rolled-back install = %q, error = %v", content, err)
	}
	if _, err := os.Stat(stage); !os.IsNotExist(err) {
		t.Fatalf("staging directory remains after rollback: %v", err)
	}
}

func TestRecoverDesktopUpdateFinalizesConfirmedCandidate(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "bob-gemini-free-wails.exe")
	stage := filepath.Join(root, ".bob-gemini-free-update-recover-confirmed")
	plan := recoveryPlan(target, stage)
	if err := os.Mkdir(stage, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("new"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(plan.BackupPath, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := writeDesktopUpdatePlan(plan); err != nil {
		t.Fatal(err)
	}
	if err := ConfirmDesktopUpdate(plan.ConfirmationPath); err != nil {
		t.Fatal(err)
	}

	if err := recoverDesktopUpdate(target, "windows", ""); err != nil {
		t.Fatalf("recoverDesktopUpdate: %v", err)
	}
	content, err := os.ReadFile(target)
	if err != nil || string(content) != "new" {
		t.Fatalf("confirmed install = %q, error = %v", content, err)
	}
	if _, err := os.Stat(stage); !os.IsNotExist(err) {
		t.Fatalf("confirmed staging directory remains: %v", err)
	}
}

func TestRecoverDesktopUpdateLeavesCandidateStartingTransactionAlone(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "bob-gemini-free-wails.exe")
	stage := filepath.Join(root, ".bob-gemini-free-update-recover-starting")
	plan := recoveryPlan(target, stage)
	if err := os.Mkdir(stage, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("new"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(plan.BackupPath, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := writeDesktopUpdatePlan(plan); err != nil {
		t.Fatal(err)
	}

	if err := recoverDesktopUpdate(target, "windows", plan.ConfirmationPath); err != nil {
		t.Fatalf("recoverDesktopUpdate: %v", err)
	}
	content, err := os.ReadFile(target)
	if err != nil || string(content) != "new" {
		t.Fatalf("starting candidate changed = %q, error = %v", content, err)
	}
	if _, err := os.Stat(plan.BackupPath); err != nil {
		t.Fatalf("starting candidate rollback backup disappeared: %v", err)
	}
}

func TestRecoverDesktopUpdateRefusesAmbiguousCandidateState(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "bob-gemini-free-wails.exe")
	stage := filepath.Join(root, ".bob-gemini-free-update-recover-ambiguous")
	plan := recoveryPlan(target, stage)
	if err := os.Mkdir(stage, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("new"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(plan.BackupPath, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(plan.CandidatePath, []byte("also-new"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := writeDesktopUpdatePlan(plan); err != nil {
		t.Fatal(err)
	}

	if err := recoverDesktopUpdate(target, "windows", ""); err == nil || !strings.Contains(err.Error(), "both an active install") {
		t.Fatalf("ambiguous recovery error = %v", err)
	}
	content, err := os.ReadFile(target)
	if err != nil || string(content) != "new" {
		t.Fatalf("ambiguous state was modified = %q, error = %v", content, err)
	}
}

func recoveryPlan(target, stage string) *DesktopUpdatePlan {
	return &DesktopUpdatePlan{
		PlanPath:         filepath.Join(stage, "update-plan.json"),
		StageDir:         stage,
		InstallTarget:    target,
		CandidatePath:    filepath.Join(stage, "candidate.exe"),
		BackupPath:       filepath.Join(stage, "rollback-backup"),
		ConfirmationPath: filepath.Join(stage, ".bob-gemini-update-confirm"),
		TargetVersion:    "v0.2.0",
		AssetName:        "bob-gemini-free-wails-windows-amd64.exe",
		Channel:          DesktopChannelStable,
		TargetOS:         "windows",
	}
}
