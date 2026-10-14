package main

import (
	"strings"
	"testing"
)

// TestDiagnosticKeyOnlyGoesToHTTPSOrLoopback is the M10 regression: the first
// configured key was attached to any --test-url, including plain HTTP to a LAN
// host.
func TestDiagnosticKeyOnlyGoesToHTTPSOrLoopback(t *testing.T) {
	for _, target := range []string{"http://127.0.0.1:9610", "http://localhost:9610", "http://[::1]:9610", "https://gateway.example.com"} {
		if !keyMayTravelTo(target) {
			t.Errorf("key should be allowed for %s", target)
		}
	}
	for _, target := range []string{"http://gateway.example.com:9610", "http://192.168.1.20:9610", "http://0.0.0.0:9610", "ftp://127.0.0.1", "://bad"} {
		if keyMayTravelTo(target) {
			t.Errorf("key must not be sent to %s", target)
		}
	}
}

func TestResolveTestKeyWithholdsConfiguredKeyOverPlainHTTP(t *testing.T) {
	key, note := resolveTestKey("", []string{"classroom-key"}, "http://192.168.1.20:9610")
	if key != "" {
		t.Fatalf("configured key sent over plain HTTP to LAN: %q", key)
	}
	if !strings.Contains(note, "--test-key") {
		t.Fatalf("note should explain how to send a key deliberately: %q", note)
	}
}

func TestResolveTestKeyUsesConfiguredKeyOnLoopbackAndHTTPS(t *testing.T) {
	for _, target := range []string{"http://127.0.0.1:9610", "https://gateway.example.com"} {
		key, note := resolveTestKey("", []string{"classroom-key", "second"}, target)
		if key != "classroom-key" || note != "" {
			t.Fatalf("%s: key=%q note=%q", target, key, note)
		}
	}
}

func TestResolveTestKeyHonoursExplicitFlagWithWarning(t *testing.T) {
	key, note := resolveTestKey("typed-key", []string{"classroom-key"}, "http://192.168.1.20:9610")
	if key != "typed-key" {
		t.Fatalf("explicit --test-key was not used: %q", key)
	}
	if !strings.Contains(note, "plain HTTP") {
		t.Fatalf("explicit key over plain HTTP needs a warning: %q", note)
	}
	if key, note := resolveTestKey("typed-key", nil, "https://gateway.example.com"); key != "typed-key" || note != "" {
		t.Fatalf("explicit key over HTTPS: key=%q note=%q", key, note)
	}
}

func TestResolveTestKeyEmptyWhenNothingConfigured(t *testing.T) {
	if key, note := resolveTestKey("", nil, "http://192.168.1.20:9610"); key != "" || note != "" {
		t.Fatalf("no keys configured: key=%q note=%q", key, note)
	}
}
