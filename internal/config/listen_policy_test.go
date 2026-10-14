package config

import (
	"errors"
	"strings"
	"testing"
)

// TestCheckListenPolicyRefusesPublicBindWithoutAPIKey is the H3 regression: the
// shipped default and Docker image bind 0.0.0.0, and nothing stopped a
// keyless gateway from starting there.
func TestCheckListenPolicyRefusesPublicBindWithoutAPIKey(t *testing.T) {
	for _, host := range []string{"0.0.0.0", "::", "192.168.1.20", "gateway.lan"} {
		cfg := Default()
		cfg.Host = host
		cfg.APIKeys = nil
		err := CheckListenPolicy(cfg)
		if err == nil {
			t.Fatalf("host %q without api_keys: want refusal, got nil", host)
		}
		if !strings.Contains(err.Error(), "api_keys") {
			t.Fatalf("host %q: refusal does not tell the operator what to set: %v", host, err)
		}
	}
}

func TestCheckListenPolicyAllowsPublicBindWithAPIKey(t *testing.T) {
	cfg := Default()
	cfg.Host = "0.0.0.0"
	cfg.APIKeys = []string{"classroom-key"}
	if err := CheckListenPolicy(cfg); err != nil {
		t.Fatalf("public bind with an api key should start: %v", err)
	}
}

func TestCheckListenPolicyAllowsLoopbackWithoutAPIKey(t *testing.T) {
	for _, host := range []string{"127.0.0.1", "localhost", "::1", "[::1]", "127.8.9.10"} {
		cfg := Default()
		cfg.Host = host
		cfg.APIKeys = nil
		if err := CheckListenPolicy(cfg); err != nil {
			t.Fatalf("loopback host %q without api_keys should start: %v", host, err)
		}
	}
}

func TestCheckListenPolicyRejectsEmptyHost(t *testing.T) {
	cfg := Default()
	cfg.Host = "  "
	cfg.APIKeys = []string{"classroom-key"}
	if err := CheckListenPolicy(cfg); !errors.Is(err, ErrEmptyListenHost) {
		t.Fatalf("empty host error = %v, want ErrEmptyListenHost", err)
	}
}

// The example config is copied by install scripts. It must not ship a key that
// looks usable, because operators who copy it verbatim would expose the gateway
// with a known credential.
func TestExampleConfigShipsNoAPIKey(t *testing.T) {
	cfg, err := Load("../../config.example.json")
	if err != nil {
		t.Fatalf("load example config: %v", err)
	}
	if len(cfg.APIKeys) != 0 {
		t.Fatalf("config.example.json must not ship an API key: %v", cfg.APIKeys)
	}
}
