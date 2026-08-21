//go:build !darwin && !linux && !freebsd && !openbsd && !netbsd

package diag

func currentRSSBytes() (uint64, bool) {
	return 0, false
}
