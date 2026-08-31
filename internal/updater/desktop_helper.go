package updater

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	// DesktopUpdateHelperFlag is consumed before Wails starts. The helper is a
	// copied executable, so it does not depend on the running app remaining
	// writable or alive while the replacement occurs.
	DesktopUpdateHelperFlag  = "--bob-desktop-update-helper"
	DesktopUpdateConfirmFlag = "--bob-desktop-update-confirm"

	maxDesktopUpdatePlanBytes         int64 = 64 << 10
	maxDesktopUpdateConfirmationBytes int64 = 4 << 10
)

var (
	desktopUpdateRetryWindow         = 30 * time.Second
	desktopUpdatePollInterval        = 250 * time.Millisecond
	desktopUpdateConfirmationTimeout = 45 * time.Second
)

const desktopUpdateLockStaleAfter = 10 * time.Minute

var errDesktopUpdateInProgress = errors.New("another desktop update is already in progress")

type desktopUpdateLock struct {
	path string
	once sync.Once
}

// acquireDesktopUpdateLock coordinates helper processes that may have been
// started by two app instances at the same time. An atomic directory create is
// available on macOS and Windows without a third-party dependency. The lock
// is intentionally beside the install target, not inside a staging directory
// that a competing update could remove during rollback.
func acquireDesktopUpdateLock(installTarget string) (*desktopUpdateLock, error) {
	if strings.TrimSpace(installTarget) == "" || !filepath.IsAbs(installTarget) {
		return nil, fmt.Errorf("desktop update lock requires an absolute install target")
	}
	lockPath := desktopUpdateLockPath(installTarget)
	for attempt := 0; attempt < 2; attempt++ {
		if err := os.Mkdir(lockPath, 0700); err == nil {
			return &desktopUpdateLock{path: lockPath}, nil
		} else if !os.IsExist(err) {
			return nil, fmt.Errorf("create desktop update lock: %w", err)
		}

		info, err := os.Lstat(lockPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("inspect desktop update lock: %w", err)
		}
		age := time.Since(info.ModTime())
		if age < 0 || age <= desktopUpdateLockStaleAfter {
			return nil, fmt.Errorf("%w: %s", errDesktopUpdateInProgress, lockPath)
		}
		// The lock directory is created empty. Remove only the exact lock path;
		// refusing to recursively remove it avoids turning stale-lock recovery
		// into a deletion primitive if a damaged directory contains extra data.
		if err := os.Remove(lockPath); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("remove stale desktop update lock: %w", err)
		}
	}
	return nil, fmt.Errorf("%w: %s", errDesktopUpdateInProgress, lockPath)
}

func desktopUpdateLockPath(installTarget string) string {
	cleanTarget := filepath.Clean(installTarget)
	return filepath.Join(filepath.Dir(cleanTarget), "."+filepath.Base(cleanTarget)+".update-lock")
}

func (lock *desktopUpdateLock) Release() error {
	if lock == nil {
		return nil
	}
	var err error
	lock.once.Do(func() {
		err = os.Remove(lock.path)
		if os.IsNotExist(err) {
			err = nil
		}
	})
	return err
}

// HandleDesktopUpdateCommand handles the private helper command before the
// Wails runtime starts. It returns handled=false for ordinary application
// arguments and handled=true for the helper command.
func HandleDesktopUpdateCommand(args []string) (handled bool, err error) {
	if len(args) == 0 || args[0] != DesktopUpdateHelperFlag {
		return false, nil
	}
	if len(args) != 2 || strings.TrimSpace(args[1]) == "" {
		return true, fmt.Errorf("desktop update helper requires exactly one plan path")
	}
	return true, runDesktopUpdateHelper(args[1])
}

// DesktopUpdateConfirmationPath extracts the confirmation marker passed to a
// newly launched application. It accepts both forms used by platform launch
// tools: --flag value and --flag=value.
func DesktopUpdateConfirmationPath(args []string) string {
	for index, arg := range args {
		if arg == DesktopUpdateConfirmFlag && index+1 < len(args) {
			return args[index+1]
		}
		prefix := DesktopUpdateConfirmFlag + "="
		if strings.HasPrefix(arg, prefix) {
			return strings.TrimPrefix(arg, prefix)
		}
	}
	return ""
}

// ConfirmDesktopUpdate atomically writes the local health marker watched by
// the helper. The marker is deliberately a short-lived file beside the staged
// update, not a network signal or telemetry event.
func ConfirmDesktopUpdate(path string) error {
	if err := validateConfirmationPath(path); err != nil {
		return err
	}
	if err := writeAtomicDesktopUpdateFile(path, []byte("healthy\n"), 0600); err != nil {
		return fmt.Errorf("commit desktop update confirmation: %w", err)
	}
	return nil
}

// StartDesktopUpdate launches a copied helper and returns. The caller must
// quit the running Wails process immediately after this succeeds. The helper
// waits for the old process to release its files, performs the transaction,
// relaunches the candidate, and rolls back if the candidate never confirms
// healthy startup.
func StartDesktopUpdate(plan *DesktopUpdatePlan) error {
	if err := validateDesktopUpdatePlan(plan); err != nil {
		return err
	}
	currentExecutable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("determine current desktop executable: %w", err)
	}
	if resolved, resolveErr := filepath.EvalSymlinks(currentExecutable); resolveErr == nil {
		currentExecutable = resolved
	}

	helperPath := filepath.Join(plan.StageDir, ".bob-gemini-free-update-helper")
	if err := copyFile(currentExecutable, helperPath); err != nil {
		return fmt.Errorf("prepare desktop update helper: %w", err)
	}
	if err := os.Chmod(helperPath, 0755); err != nil {
		return fmt.Errorf("make desktop update helper executable: %w", err)
	}
	plan.HelperPath = helperPath
	if err := writeDesktopUpdatePlan(plan); err != nil {
		return err
	}

	logPath := filepath.Join(plan.StageDir, "helper.log")
	logFile, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("create desktop update helper log: %w", err)
	}
	command := exec.Command(helperPath, DesktopUpdateHelperFlag, plan.PlanPath)
	command.Dir = plan.StageDir
	command.Stdout = logFile
	command.Stderr = logFile
	if err := command.Start(); err != nil {
		_ = logFile.Close()
		return fmt.Errorf("start desktop update helper: %w", err)
	}
	_ = logFile.Close()
	return nil
}

func runDesktopUpdateHelper(planPath string) error {
	planPath, err := filepath.Abs(planPath)
	if err != nil {
		return fmt.Errorf("resolve desktop update plan path: %w", err)
	}
	data, err := readBoundedDesktopUpdateFile(planPath, maxDesktopUpdatePlanBytes)
	if err != nil {
		return fmt.Errorf("read desktop update plan: %w", err)
	}
	var plan DesktopUpdatePlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return fmt.Errorf("decode desktop update plan: %w", err)
	}
	if plan.PlanPath == "" {
		plan.PlanPath = planPath
	}
	if filepath.Clean(plan.PlanPath) != filepath.Clean(planPath) {
		return fmt.Errorf("desktop update plan path does not match helper argument")
	}
	if err := validateDesktopUpdatePlan(&plan); err != nil {
		return err
	}
	if plan.HelperPath == "" {
		return fmt.Errorf("desktop update plan has no helper path")
	}

	deadline := time.Now().Add(desktopUpdateRetryWindow)
	var lastErr error
	for time.Now().Before(deadline) {
		if _, statErr := os.Stat(plan.CandidatePath); os.IsNotExist(statErr) {
			if lastErr != nil {
				return fmt.Errorf("desktop update candidate is no longer available: %w", lastErr)
			}
			return fmt.Errorf("desktop update candidate is no longer available")
		} else if statErr != nil {
			lastErr = fmt.Errorf("inspect desktop update candidate: %w", statErr)
			break
		}
		if err := replaceAndConfirmDesktopUpdate(&plan); err == nil {
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(desktopUpdatePollInterval)
	}
	return fmt.Errorf("desktop update did not complete within %s: %w", desktopUpdateRetryWindow, lastErr)
}

func replaceAndConfirmDesktopUpdate(plan *DesktopUpdatePlan) error {
	if err := validateInstallTarget(plan.InstallTarget, plan.TargetOS); err != nil {
		return err
	}
	if err := validateCandidateTarget(plan.CandidatePath, plan.TargetOS); err != nil {
		return err
	}
	lock, err := acquireDesktopUpdateLock(plan.InstallTarget)
	if err != nil {
		return err
	}
	defer func() { _ = lock.Release() }()
	// Revalidate after taking the cross-process lock. The first checks protect
	// normal input errors; these checks narrow the target-swap window between
	// validation and the destructive rename.
	if err := validateInstallTarget(plan.InstallTarget, plan.TargetOS); err != nil {
		return err
	}
	if err := validateCandidateTarget(plan.CandidatePath, plan.TargetOS); err != nil {
		return err
	}
	if _, err := os.Stat(plan.BackupPath); err == nil {
		return fmt.Errorf("rollback path already exists: %s", plan.BackupPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect rollback path: %w", err)
	}

	if err := os.Rename(plan.InstallTarget, plan.BackupPath); err != nil {
		return fmt.Errorf("move current desktop install to rollback path: %w", err)
	}
	if err := syncDesktopUpdateDirectory(filepath.Dir(plan.InstallTarget)); err != nil {
		restoreErr := os.Rename(plan.BackupPath, plan.InstallTarget)
		restoreSyncErr := syncDesktopUpdateDirectory(filepath.Dir(plan.InstallTarget))
		if restoreErr != nil || restoreSyncErr != nil {
			failure := fmt.Errorf("persist current desktop rollback move: %w; restore previous desktop install: rename=%v, sync=%v", err, restoreErr, restoreSyncErr)
			_ = writeDesktopUpdateFailure(plan, failure)
			return failure
		}
		return fmt.Errorf("persist current desktop rollback move: %w", err)
	}
	if err := os.Rename(plan.CandidatePath, plan.InstallTarget); err != nil {
		restoreErr := os.Rename(plan.BackupPath, plan.InstallTarget)
		restoreSyncErr := syncDesktopUpdateDirectory(filepath.Dir(plan.InstallTarget))
		if restoreErr != nil || restoreSyncErr != nil {
			failure := fmt.Errorf("activate verified desktop candidate: %w; restore previous desktop install: rename=%v, sync=%v", err, restoreErr, restoreSyncErr)
			_ = writeDesktopUpdateFailure(plan, failure)
			return failure
		}
		return fmt.Errorf("activate verified desktop candidate: %w", err)
	}
	if err := syncDesktopUpdateDirectory(filepath.Dir(plan.InstallTarget)); err != nil {
		return rollbackDesktopUpdate(plan, nil, fmt.Errorf("persist verified desktop candidate activation: %w", err))
	}

	process, err := launchUpdatedDesktop(plan.InstallTarget, plan.ConfirmationPath)
	if err != nil {
		return rollbackDesktopUpdate(plan, nil, fmt.Errorf("launch updated desktop app: %w", err))
	}
	if err := waitForDesktopConfirmation(plan.ConfirmationPath); err != nil {
		return rollbackDesktopUpdate(plan, process, err)
	}

	if err := os.RemoveAll(plan.BackupPath); err != nil {
		// Keeping the backup is safer than replacing an otherwise healthy update
		// again. Leave a local warning for the operator; do not retry the
		// replacement now that the new application is healthy.
		_ = writeDesktopUpdateWarning(plan, fmt.Errorf("updated to %s but could not remove rollback backup %s: %w", plan.TargetVersion, plan.BackupPath, err))
		return nil
	}
	if err := syncDesktopUpdateDirectory(filepath.Dir(plan.BackupPath)); err != nil {
		_ = writeDesktopUpdateWarning(plan, fmt.Errorf("updated to %s but could not persist rollback cleanup %s: %w", plan.TargetVersion, plan.BackupPath, err))
		return nil
	}
	_ = os.Remove(failureRecordPath(plan))
	if err := os.RemoveAll(plan.StageDir); err != nil {
		_ = writeDesktopUpdateWarning(plan, fmt.Errorf("updated to %s but could not remove staging directory %s: %w", plan.TargetVersion, plan.StageDir, err))
		return nil
	}
	if err := syncDesktopUpdateDirectory(filepath.Dir(plan.StageDir)); err != nil {
		_ = writeDesktopUpdateWarning(plan, fmt.Errorf("updated to %s but could not persist staging cleanup %s: %w", plan.TargetVersion, plan.StageDir, err))
	}
	return nil
}

func rollbackDesktopUpdate(plan *DesktopUpdatePlan, process *os.Process, cause error) error {
	_ = writeDesktopUpdateFailure(plan, cause)
	if process != nil {
		_ = process.Kill()
		_, _ = process.Wait()
	}
	if err := os.RemoveAll(plan.InstallTarget); err != nil {
		return fmt.Errorf("%v; remove failed desktop candidate for rollback: %w", cause, err)
	}
	if err := os.Rename(plan.BackupPath, plan.InstallTarget); err != nil {
		return fmt.Errorf("%v; restore desktop rollback backup: %w", cause, err)
	}
	if err := syncDesktopUpdateDirectory(filepath.Dir(plan.InstallTarget)); err != nil {
		return fmt.Errorf("%v; persist restored desktop install: %w", cause, err)
	}
	if err := os.RemoveAll(plan.StageDir); err != nil {
		return fmt.Errorf("%v; remove failed desktop staging directory: %w", cause, err)
	}
	if err := syncDesktopUpdateDirectory(filepath.Dir(plan.StageDir)); err != nil {
		return fmt.Errorf("%v; persist failed desktop staging cleanup: %w", cause, err)
	}
	return cause
}

func waitForDesktopConfirmation(path string) error {
	deadline := time.Now().Add(desktopUpdateConfirmationTimeout)
	for time.Now().Before(deadline) {
		data, err := readBoundedDesktopUpdateFile(path, maxDesktopUpdateConfirmationBytes)
		if err == nil {
			if strings.TrimSpace(string(data)) == "healthy" {
				return nil
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("read desktop update confirmation: %w", err)
		}
		time.Sleep(desktopUpdatePollInterval)
	}
	return fmt.Errorf("updated desktop app did not confirm healthy startup within %s", desktopUpdateConfirmationTimeout)
}

var launchUpdatedDesktop = launchUpdatedDesktopProcess

func launchUpdatedDesktopProcess(installTarget, confirmationPath string) (*os.Process, error) {
	executable := installTarget
	if runtime.GOOS == "darwin" {
		var err error
		executable, err = findDesktopBundleBinary(installTarget)
		if err != nil {
			return nil, err
		}
	}
	command := exec.Command(executable, DesktopUpdateConfirmFlag, confirmationPath)
	command.Dir = filepath.Dir(executable)
	if err := command.Start(); err != nil {
		return nil, err
	}
	return command.Process, nil
}

func validateDesktopUpdatePlan(plan *DesktopUpdatePlan) error {
	if plan == nil {
		return fmt.Errorf("desktop update plan is nil")
	}
	for name, path := range map[string]string{
		"plan":         plan.PlanPath,
		"stage":        plan.StageDir,
		"install":      plan.InstallTarget,
		"candidate":    plan.CandidatePath,
		"rollback":     plan.BackupPath,
		"confirmation": plan.ConfirmationPath,
	} {
		if path == "" || !filepath.IsAbs(path) {
			return fmt.Errorf("desktop update %s path must be absolute", name)
		}
	}
	if plan.TargetVersion == "" || plan.AssetName == "" || plan.Channel == "" {
		return fmt.Errorf("desktop update plan is missing release identity")
	}
	if plan.TargetOS != "darwin" && plan.TargetOS != "windows" {
		return fmt.Errorf("desktop update plan has unsupported target OS: %s", plan.TargetOS)
	}
	stagePath := filepath.Clean(plan.StageDir)
	if !strings.HasPrefix(filepath.Base(stagePath), desktopStagingPrefix) {
		return fmt.Errorf("desktop update staging directory has an unexpected name")
	}
	if filepath.Clean(plan.PlanPath) != filepath.Join(stagePath, "update-plan.json") {
		return fmt.Errorf("desktop update plan path has an unexpected location")
	}
	if filepath.Clean(plan.BackupPath) != filepath.Join(stagePath, "rollback-backup") {
		return fmt.Errorf("desktop update rollback path has an unexpected location")
	}
	if filepath.Clean(plan.ConfirmationPath) != filepath.Join(stagePath, ".bob-gemini-update-confirm") {
		return fmt.Errorf("desktop update confirmation path has an unexpected location")
	}
	if filepath.Clean(plan.StageDir) == filepath.Clean(filepath.Dir(plan.InstallTarget)) {
		return fmt.Errorf("desktop update staging directory cannot be the install directory")
	}
	if err := ensurePathWithin(filepath.Dir(plan.InstallTarget), plan.StageDir); err != nil {
		return fmt.Errorf("desktop update staging directory is not beside install target: %w", err)
	}
	if err := ensurePathWithin(plan.StageDir, plan.CandidatePath); err != nil {
		return fmt.Errorf("desktop candidate is outside staging directory: %w", err)
	}
	if err := ensurePathWithin(plan.StageDir, plan.BackupPath); err != nil {
		return fmt.Errorf("desktop rollback path is outside staging directory: %w", err)
	}
	if err := ensurePathWithin(plan.StageDir, plan.ConfirmationPath); err != nil {
		return fmt.Errorf("desktop confirmation path is outside staging directory: %w", err)
	}
	if err := ensurePathWithin(plan.StageDir, plan.PlanPath); err != nil {
		return fmt.Errorf("desktop plan path is outside staging directory: %w", err)
	}
	if plan.HelperPath != "" {
		if err := ensurePathWithin(plan.StageDir, plan.HelperPath); err != nil {
			return fmt.Errorf("desktop helper path is outside staging directory: %w", err)
		}
		if filepath.Base(filepath.Clean(plan.HelperPath)) != ".bob-gemini-free-update-helper" {
			return fmt.Errorf("desktop helper path has an unexpected name")
		}
	}
	return nil
}

func readBoundedDesktopUpdateFile(path string, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, fmt.Errorf("desktop update file limit is invalid")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("desktop update file exceeds %d bytes", maxBytes)
	}
	return data, nil
}

func validateConfirmationPath(path string) error {
	if path == "" || !filepath.IsAbs(path) {
		return fmt.Errorf("desktop update confirmation path must be absolute")
	}
	base := filepath.Base(path)
	if base != ".bob-gemini-update-confirm" {
		return fmt.Errorf("desktop update confirmation path has an unexpected name")
	}
	parent := filepath.Base(filepath.Dir(filepath.Clean(path)))
	if !strings.HasPrefix(parent, desktopStagingPrefix) {
		return fmt.Errorf("desktop update confirmation path is outside updater staging")
	}
	return nil
}

func validateCandidateTarget(path, targetOS string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("inspect staged desktop candidate %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to activate a symlinked desktop candidate: %s", path)
	}
	if targetOS == "darwin" && !info.IsDir() {
		return fmt.Errorf("staged macOS desktop candidate is not an app directory: %s", path)
	}
	if targetOS == "windows" && info.IsDir() {
		return fmt.Errorf("staged Windows desktop candidate is a directory: %s", path)
	}
	return nil
}

func failureRecordPath(plan *DesktopUpdatePlan) string {
	return filepath.Join(filepath.Dir(plan.InstallTarget), "."+filepath.Base(plan.InstallTarget)+".update-failure.txt")
}

func warningRecordPath(plan *DesktopUpdatePlan) string {
	return filepath.Join(filepath.Dir(plan.InstallTarget), "."+filepath.Base(plan.InstallTarget)+".update-warning.txt")
}

func writeDesktopUpdateFailure(plan *DesktopUpdatePlan, cause error) error {
	if plan == nil || cause == nil {
		return nil
	}
	content := fmt.Sprintf("BOB Gemini Free desktop update failed at %s\nTarget: %s\nRelease: %s\nReason: %s\n", time.Now().UTC().Format(time.RFC3339), plan.InstallTarget, plan.TargetVersion, cause)
	return writeAtomicDesktopUpdateFile(failureRecordPath(plan), []byte(content), 0600)
}

func writeDesktopUpdateWarning(plan *DesktopUpdatePlan, cause error) error {
	if plan == nil || cause == nil {
		return nil
	}
	content := fmt.Sprintf("BOB Gemini Free desktop update completed with cleanup warning at %s\nTarget: %s\nRelease: %s\nReason: %s\n", time.Now().UTC().Format(time.RFC3339), plan.InstallTarget, plan.TargetVersion, cause)
	return writeAtomicDesktopUpdateFile(warningRecordPath(plan), []byte(content), 0600)
}
