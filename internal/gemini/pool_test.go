package gemini

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCookiePool(t *testing.T) {
	pool := NewCookiePool()
	if count := pool.Count(); count != 0 {
		t.Errorf("Expected 0 sessions, got %d", count)
	}

	pool.AddSession("acc1.txt", "SAPISID=sapi1; SID=sid1", "sapi1", "0")
	pool.AddSession("acc2.txt", "SAPISID=sapi2; SID=sid2", "sapi2", "1")

	if count := pool.Count(); count != 2 {
		t.Fatalf("Expected 2 sessions, got %d", count)
	}

	s1 := pool.GetHealthySession()
	s2 := pool.GetHealthySession()

	if s1 == nil || s2 == nil {
		t.Fatalf("Expected non-nil sessions")
	}

	// Verify round-robin alternating
	if s1.ID == s2.ID {
		t.Errorf("Expected round-robin rotation, got same session %s", s1.ID)
	}

	// Test Failover
	pool.MarkFailure("sapi1")
	// Now sapi1 should be in cooldown, so next requests should prefer sapi2
	s3 := pool.GetHealthySession()
	if s3.SAPISID != "sapi2" {
		t.Errorf("Expected failover to sapi2, got %s", s3.SAPISID)
	}

	// Reset success
	pool.MarkSuccess("sapi1")
}

func TestCookiePoolSuccessClearsFailureCooldown(t *testing.T) {
	pool := NewCookiePool()
	pool.AddSession("acc.txt", "SAPISID=sapi; SID=sid", "sapi", "0")

	pool.MarkFailure("sapi")
	if healthy := pool.CountHealthy(); healthy != 0 {
		t.Fatalf("healthy sessions after failure = %d, want 0", healthy)
	}

	pool.MarkSuccess("sapi")
	if healthy := pool.CountHealthy(); healthy != 1 {
		t.Fatalf("healthy sessions after success = %d, want 1", healthy)
	}
}

func TestCookiePoolDeduplicatesSessionIdentity(t *testing.T) {
	pool := NewCookiePool()
	pool.AddSession("one.txt", "SAPISID=same; SID=one", "same", "0")
	pool.AddSession("two.txt", "SAPISID=same; SID=two", "same", "1")
	if got := pool.Count(); got != 1 {
		t.Fatalf("duplicate session count = %d, want 1", got)
	}
	session := pool.GetHealthySession()
	if session == nil || session.SAPISID != "same" || session.AuthUser != "1" {
		t.Fatalf("deduplicated session = %#v", session)
	}
	if session.ID == "same" {
		t.Fatal("session ID retained raw SAPISID")
	}
}

func TestCookiePoolReloadRemovesDeletedSourcesAndPreservesCooldown(t *testing.T) {
	directory := t.TempDir()
	cookieFile := filepath.Join(directory, "student.txt")
	if err := os.WriteFile(cookieFile, []byte("SID=sid; SAPISID=sapisid"), 0600); err != nil {
		t.Fatalf("write cookie: %v", err)
	}
	pool := NewCookiePool()
	if got := pool.LoadFromFiles([]string{cookieFile}); got != 1 {
		t.Fatalf("loaded cookies = %d, want 1", got)
	}
	pool.MarkFailure(pool.GetHealthySession().ID)
	if got := pool.CountHealthy(); got != 0 {
		t.Fatalf("healthy sessions after failure = %d, want 0", got)
	}
	if got := pool.Reload(); got != 1 {
		t.Fatalf("reloaded cookies = %d, want 1", got)
	}
	if got := pool.CountHealthy(); got != 0 {
		t.Fatalf("reload cleared cooldown; healthy sessions = %d", got)
	}
	if err := os.Remove(cookieFile); err != nil {
		t.Fatalf("remove cookie: %v", err)
	}
	if got := pool.Reload(); got != 0 {
		t.Fatalf("reload after deletion = %d, want 0", got)
	}
	if got := pool.Count(); got != 0 {
		t.Fatalf("deleted source left %d sessions", got)
	}
}

func TestCookiePoolDoesNotBypassAllSessionCooldowns(t *testing.T) {
	pool := NewCookiePool()
	pool.AddSession("one.txt", "SID=one; SAPISID=one", "one", "")
	session := pool.GetHealthySession()
	if session == nil {
		t.Fatal("initial session is nil")
	}
	pool.MarkFailure(session.ID)
	if got := pool.GetHealthySession(); got != nil {
		t.Fatalf("unhealthy session was returned during cooldown: %#v", got)
	}
}
