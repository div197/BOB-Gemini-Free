package gemini

import (
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
	if s3.ID != "sapi2" {
		t.Errorf("Expected failover to sapi2, got %s", s3.ID)
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
