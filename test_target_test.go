package main

import "testing"

// TestDiagnosticKeyOnlyGoesToHTTPSOrLoopback is the M10 regression: the first
// configured key was attached to any --test-url, including plain HTTP to a LAN
// host.
func TestDiagnosticKeyOnlyGoesToHTTPSOrLoopback(t *testing.T) {
	allowed := []string{
		"http://127.0.0.1:9610",
		"http://localhost:9610",
		"http://[::1]:9610",
		"https://gateway.example.com",
	}
	for _, target := range allowed {
		if !keyMayTravelTo(target) {
			t.Errorf("key should be allowed for %s", target)
		}
	}
	refused := []string{
		"http://gateway.example.com:9610",
		"http://192.168.1.20:9610",
		"http://0.0.0.0:9610",
		"ftp://127.0.0.1",
		"://bad",
	}
	for _, target := range refused {
		if keyMayTravelTo(target) {
			t.Errorf("key must not be sent to %s", target)
		}
	}
}

func TestCheckTestKeyTargetRefusesPlainHTTPWithKey(t *testing.T) {
	if err := checkTestKeyTarget("http://192.168.1.20:9610", "classroom-key"); err == nil {
		t.Fatal("plain HTTP LAN target with a key should be refused")
	}
	if err := checkTestKeyTarget("http://192.168.1.20:9610", ""); err != nil {
		t.Fatalf("no key should never be refused: %v", err)
	}
	if err := checkTestKeyTarget("http://127.0.0.1:9610", "classroom-key"); err != nil {
		t.Fatalf("loopback with a key should be allowed: %v", err)
	}
}
