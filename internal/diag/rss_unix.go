//go:build darwin || linux || freebsd || openbsd || netbsd

package diag

import (
	"runtime"
	"syscall"
)

func currentRSSBytes() (uint64, bool) {
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
		return 0, false
	}
	rss := uint64(usage.Maxrss)
	if runtime.GOOS == "linux" {
		rss *= 1024
	}
	return rss, true
}
