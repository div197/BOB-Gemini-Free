//go:build !windows

package updater

import "os"

func replaceDesktopUpdateFile(temporaryPath, path string) error {
	return os.Rename(temporaryPath, path)
}
