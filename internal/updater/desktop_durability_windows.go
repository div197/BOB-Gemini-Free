//go:build windows

package updater

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	moveFileReplaceExisting = 0x1
	moveFileWriteThrough    = 0x8
)

var moveFileEx = syscall.NewLazyDLL("kernel32.dll").NewProc("MoveFileExW")

func replaceDesktopUpdateFile(temporaryPath, path string) error {
	temporaryName, err := syscall.UTF16PtrFromString(temporaryPath)
	if err != nil {
		return fmt.Errorf("encode desktop update temporary path: %w", err)
	}
	destinationName, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return fmt.Errorf("encode desktop update destination path: %w", err)
	}
	result, _, callErr := moveFileEx.Call(
		uintptr(unsafe.Pointer(temporaryName)),
		uintptr(unsafe.Pointer(destinationName)),
		moveFileReplaceExisting|moveFileWriteThrough,
	)
	if result == 0 {
		return fmt.Errorf("move desktop update metadata with write-through replacement: %w", callErr)
	}
	return nil
}
