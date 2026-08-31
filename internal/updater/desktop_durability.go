package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// writeAtomicDesktopUpdateFile commits a small updater-owned metadata file only
// after its contents are flushed and the containing directory is synchronized.
// The temporary file is created beside the destination so the final rename is
// same-filesystem and readers never observe a partial document.
func writeAtomicDesktopUpdateFile(path string, data []byte, perm os.FileMode) error {
	if path == "" || !filepath.IsAbs(path) {
		return fmt.Errorf("desktop update file path must be absolute")
	}
	if perm.Perm() == 0 {
		return fmt.Errorf("desktop update file mode is invalid")
	}

	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, "."+filepath.Base(path)+".tmp-")
	if err != nil {
		return fmt.Errorf("create desktop update temporary file: %w", err)
	}
	temporaryPath := temporary.Name()
	committed := false
	defer func() {
		_ = temporary.Close()
		if !committed {
			_ = os.Remove(temporaryPath)
		}
	}()

	if err := temporary.Chmod(perm.Perm()); err != nil {
		return fmt.Errorf("set desktop update temporary file permissions: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		return fmt.Errorf("write desktop update temporary file: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return fmt.Errorf("flush desktop update temporary file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close desktop update temporary file: %w", err)
	}
	if err := replaceDesktopUpdateFile(temporaryPath, path); err != nil {
		return fmt.Errorf("commit desktop update file: %w", err)
	}
	committed = true
	if err := syncDesktopUpdateDirectory(directory); err != nil {
		return err
	}
	return nil
}

func replaceDesktopUpdateFile(temporaryPath, path string) error {
	err := os.Rename(temporaryPath, path)
	if err == nil {
		return nil
	}
	if runtime.GOOS != "windows" {
		return err
	}

	// Windows does not replace an existing destination with Rename. These
	// destinations are updater-owned metadata files, never the installed app;
	// remove only the exact existing file before retrying the same-directory
	// rename so failure records and repeated confirmations retain overwrite
	// behavior without weakening executable replacement.
	if _, statErr := os.Lstat(path); statErr != nil {
		return err
	}
	if removeErr := os.Remove(path); removeErr != nil {
		return fmt.Errorf("remove existing desktop update metadata: %w", removeErr)
	}
	return os.Rename(temporaryPath, path)
}

// syncDesktopUpdateDirectory is a seam for fault-injection tests around the
// swap protocol. Windows does not expose a portable directory fsync contract;
// NTFS journaling and the existing rollback transaction remain the platform
// boundary there. macOS and Unix-like hosts flush the directory entry.
var syncDesktopUpdateDirectory = syncDesktopUpdateDirectoryImpl

func syncDesktopUpdateDirectoryImpl(path string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	directory, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open desktop update directory for sync: %w", err)
	}
	syncErr := directory.Sync()
	closeErr := directory.Close()
	if syncErr != nil {
		return fmt.Errorf("flush desktop update directory: %w", syncErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close desktop update directory: %w", closeErr)
	}
	return nil
}
