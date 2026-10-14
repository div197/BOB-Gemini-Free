package gemini

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestSaveCookieFileDoesNotChmodCurrentDirectory is the M7 regression. Saving
// the local ./cookie.txt used to run os.Chmod on "." (the user's working
// directory), so a login from a shared folder changed that folder's mode.
func TestSaveCookieFileDoesNotChmodCurrentDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits")
	}
	dir := t.TempDir()
	if err := os.Chmod(dir, 0755); err != nil {
		t.Fatalf("chmod fixture dir: %v", err)
	}
	t.Chdir(dir)

	if err := SaveCookieFile("./cookie.txt", "SID=sid; SAPISID=sapisid"); err != nil {
		t.Fatalf("SaveCookieFile: %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0755 {
		t.Fatalf("current directory mode = %o, want 0755 unchanged", got)
	}
	fileInfo, err := os.Stat(filepath.Join(dir, "cookie.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if got := fileInfo.Mode().Perm(); got != 0600 {
		t.Fatalf("cookie file mode = %o, want 0600", got)
	}
}

// A directory we create for the cookie file is still private.
func TestSaveCookieFileCreatesPrivateDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits")
	}
	target := filepath.Join(t.TempDir(), "bob-gemini-free", "cookie.txt")
	if err := SaveCookieFile(target, "SID=sid; SAPISID=sapisid"); err != nil {
		t.Fatalf("SaveCookieFile: %v", err)
	}
	info, err := os.Stat(filepath.Dir(target))
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0700 {
		t.Fatalf("new config directory mode = %o, want 0700", got)
	}
}
