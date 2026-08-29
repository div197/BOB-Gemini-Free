package updater

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// RecoverDesktopUpdate repairs an updater transaction left behind by an
// interrupted helper or power loss. It is intentionally a no-op for ordinary
// development binaries that do not live inside a native install target.
// currentConfirmationPath is supplied only to the newly launched candidate;
// that process must be allowed to finish startup before recovery judges the
// transaction as unconfirmed.
func RecoverDesktopUpdate(currentConfirmationPath string) error {
	if runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
		return nil
	}
	targetPath, err := currentDesktopInstallTarget(runtime.GOOS)
	if err != nil {
		return nil
	}
	return recoverDesktopUpdate(targetPath, runtime.GOOS, currentConfirmationPath)
}

func recoverDesktopUpdate(targetPath, targetOS, currentConfirmationPath string) error {
	targetPath = filepath.Clean(targetPath)
	parent := filepath.Dir(targetPath)
	entries, err := os.ReadDir(parent)
	if err != nil {
		return fmt.Errorf("inspect desktop update staging for recovery: %w", err)
	}
	currentConfirmationPath = filepath.Clean(currentConfirmationPath)
	var recovered bool
	for _, entry := range entries {
		if !entry.IsDir() || entry.Type()&os.ModeSymlink != 0 || !strings.HasPrefix(entry.Name(), desktopStagingPrefix) {
			continue
		}
		stagePath := filepath.Join(parent, entry.Name())
		planPath := filepath.Join(stagePath, "update-plan.json")
		data, readErr := os.ReadFile(planPath)
		if readErr != nil {
			continue
		}
		var plan DesktopUpdatePlan
		if json.Unmarshal(data, &plan) != nil || filepath.Clean(plan.StageDir) != stagePath || filepath.Clean(plan.InstallTarget) != targetPath || plan.TargetOS != targetOS {
			continue
		}
		if err := validateDesktopUpdatePlan(&plan); err != nil {
			continue
		}
		if filepath.Clean(plan.ConfirmationPath) == currentConfirmationPath && currentConfirmationPath != "" {
			// The verified candidate is currently starting. Its OnStartup hook
			// owns the confirmation marker; do not roll it back underneath it.
			continue
		}
		if recovered {
			return fmt.Errorf("multiple recoverable desktop update plans target %s", targetPath)
		}
		lock, lockErr := acquireDesktopUpdateLock(targetPath)
		if lockErr != nil {
			return fmt.Errorf("lock desktop update recovery: %w", lockErr)
		}
		recoveryErr := recoverDesktopUpdatePlan(&plan)
		_ = lock.Release()
		if recoveryErr != nil {
			return recoveryErr
		}
		if _, statErr := os.Stat(plan.StageDir); os.IsNotExist(statErr) {
			recovered = true
		}
	}
	return nil
}

func recoverDesktopUpdatePlan(plan *DesktopUpdatePlan) error {
	backupInfo, err := os.Lstat(plan.BackupPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect desktop rollback backup: %w", err)
	}
	if backupInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to recover a symlinked desktop rollback backup")
	}

	_, installErr := os.Lstat(plan.InstallTarget)
	installExists := installErr == nil
	if installErr != nil && !os.IsNotExist(installErr) {
		return fmt.Errorf("inspect desktop install during recovery: %w", installErr)
	}
	candidateInfo, candidateErr := os.Lstat(plan.CandidatePath)
	candidateExists := candidateErr == nil
	if candidateErr != nil && !os.IsNotExist(candidateErr) {
		return fmt.Errorf("inspect desktop candidate during recovery: %w", candidateErr)
	}
	if candidateExists && candidateInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to recover with a symlinked desktop candidate")
	}

	confirmed, confirmationErr := desktopUpdateConfirmed(plan.ConfirmationPath)
	if confirmationErr != nil {
		return confirmationErr
	}
	if confirmed {
		if !installExists || candidateExists {
			return fmt.Errorf("desktop update recovery found an inconsistent confirmed transaction")
		}
		if err := validateInstallTarget(plan.InstallTarget, plan.TargetOS); err != nil {
			return fmt.Errorf("validate confirmed desktop install: %w", err)
		}
		if err := os.RemoveAll(plan.BackupPath); err != nil {
			return fmt.Errorf("remove confirmed desktop rollback backup: %w", err)
		}
		_ = os.Remove(failureRecordPath(plan))
		if err := os.RemoveAll(plan.StageDir); err != nil {
			return fmt.Errorf("remove confirmed desktop staging directory: %w", err)
		}
		return nil
	}

	if installExists {
		if candidateExists {
			return fmt.Errorf("desktop update recovery found both an active install and an unactivated candidate")
		}
		if err := validateInstallTarget(plan.InstallTarget, plan.TargetOS); err != nil {
			return fmt.Errorf("validate unconfirmed desktop install: %w", err)
		}
	}

	cause := fmt.Errorf("desktop update was interrupted before healthy startup confirmation")
	_ = writeDesktopUpdateFailure(plan, cause)
	if installExists {
		if err := os.RemoveAll(plan.InstallTarget); err != nil {
			return fmt.Errorf("remove interrupted desktop candidate: %w", err)
		}
	}
	if err := os.Rename(plan.BackupPath, plan.InstallTarget); err != nil {
		return fmt.Errorf("restore desktop install after interruption: %w", err)
	}
	if err := os.RemoveAll(plan.StageDir); err != nil {
		return fmt.Errorf("remove recovered desktop staging directory: %w", err)
	}
	return nil
}

func desktopUpdateConfirmed(path string) (bool, error) {
	if err := validateConfirmationPath(path); err != nil {
		return false, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read desktop update confirmation: %w", err)
	}
	return strings.TrimSpace(string(data)) == "healthy", nil
}
