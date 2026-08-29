package models

import (
	"testing"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		input       string
		defaultName string
		wantName    string
		wantMode    int
		wantThink   int
		wantErr     bool
	}{
		{
			input:       "gemini-3.7-flash",
			defaultName: "gemini-3.6-flash",
			wantName:    "gemini-3.7-flash",
			wantMode:    1,
			wantThink:   4,
			wantErr:     false,
		},
		{
			input:       "gemini-3.7-flash-thinking",
			defaultName: "gemini-3.6-flash",
			wantName:    "gemini-3.7-flash-thinking",
			wantMode:    2,
			wantThink:   0,
			wantErr:     false,
		},
		{
			input:       "gemini-3.6-flash",
			defaultName: "gemini-3.6-flash",
			wantName:    "gemini-3.6-flash",
			wantMode:    1,
			wantThink:   4,
			wantErr:     false,
		},
		{
			input:       "gemini-3.5-flash-thinking",
			defaultName: "gemini-3.6-flash",
			wantName:    "gemini-3.5-flash-thinking",
			wantMode:    2,
			wantThink:   0,
			wantErr:     false,
		},
		{
			input:       "gemini-3.6-flash@think=0",
			defaultName: "gemini-3.6-flash",
			wantName:    "gemini-3.6-flash",
			wantMode:    1,
			wantThink:   0,
			wantErr:     false,
		},
		{
			input:       "models/gemini-3.7-flash",
			defaultName: "gemini-3.6-flash",
			wantName:    "gemini-3.7-flash",
			wantMode:    1,
			wantThink:   4,
			wantErr:     false,
		},
		{
			input:       "gemini-pro",
			defaultName: "gemini-3.6-flash",
			wantName:    "gemini-pro",
			wantMode:    3,
			wantThink:   4,
			wantErr:     false,
		},
		{
			input:       "gemini-thinking",
			defaultName: "gemini-3.6-flash",
			wantName:    "gemini-thinking",
			wantMode:    2,
			wantThink:   0,
			wantErr:     false,
		},
		{
			input:       "unknown-model",
			defaultName: "gemini-3.6-flash",
			wantName:    "gemini-3.6-flash",
			wantMode:    1,
			wantThink:   4,
			wantErr:     false,
		},
		{
			input:       "gpt-5.6",
			defaultName: "gemini-3.6-flash",
			wantName:    "gpt-5.6",
			wantMode:    1,
			wantThink:   4,
			wantErr:     false,
		},
		{
			input:       "o3",
			defaultName: "gemini-3.6-flash",
			wantName:    "o3",
			wantMode:    2,
			wantThink:   0,
			wantErr:     false,
		},
		{
			input:       "claude-3-7-sonnet",
			defaultName: "gemini-3.6-flash",
			wantName:    "claude-3-7-sonnet",
			wantMode:    2,
			wantThink:   0,
			wantErr:     false,
		},
		{
			input:       "claude-code",
			defaultName: "gemini-3.6-flash",
			wantName:    "claude-code",
			wantMode:    2,
			wantThink:   0,
			wantErr:     false,
		},
		{
			input:       "gemini-3.6-flash@think=abc",
			defaultName: "gemini-3.6-flash",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			res, err := Resolve(tt.input, tt.defaultName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Resolve(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if res.Name != tt.wantName {
				t.Errorf("res.Name = %q, want %q", res.Name, tt.wantName)
			}
			if res.Mode != tt.wantMode {
				t.Errorf("res.Mode = %d, want %d", res.Mode, tt.wantMode)
			}
			if res.Think != tt.wantThink {
				t.Errorf("res.Think = %d, want %d", res.Think, tt.wantThink)
			}
		})
	}
}

func TestResolveEnhanced(t *testing.T) {
	res, err := Resolve("gemini-3.1-pro-enhanced", "gemini-3.6-flash")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if res.Extra == nil || res.Extra[31] != 2 || res.Extra[80] != 3 {
		t.Errorf("Expected extra fields {31:2, 80:3}, got %v", res.Extra)
	}
}

func TestResolveStrictRejectsUnknownModelWithoutFallback(t *testing.T) {
	if _, err := ResolveStrict("not-a-gemini-model", "gemini-3.6-flash"); err == nil {
		t.Fatal("ResolveStrict unexpectedly accepted an unknown model")
	}
	resolved, err := ResolveStrict("gemini-3.7-flash@think=0", "gemini-3.6-flash")
	if err != nil {
		t.Fatalf("ResolveStrict known model: %v", err)
	}
	if resolved.Name != "gemini-3.7-flash" || resolved.Think != 0 {
		t.Fatalf("strict resolution = %#v", resolved)
	}
}

func TestPricing(t *testing.T) {
	pClaude := GetModelPricing("claude-5-sonnet")
	if pClaude.BlendedPer1M != 9.00 {
		t.Errorf("expected claude-5-sonnet blended price 9.00, got %f", pClaude.BlendedPer1M)
	}

	pOpus := GetModelPricing("claude-5-opus")
	if pOpus.BlendedPer1M != 45.00 {
		t.Errorf("expected claude-5-opus blended price 45.00, got %f", pOpus.BlendedPer1M)
	}

	pGPT := GetModelPricing("gpt-5.6-sol")
	if pGPT.BlendedPer1M != 12.50 {
		t.Errorf("expected gpt-5.6-sol blended price 12.50, got %f", pGPT.BlendedPer1M)
	}

	pGemini := GetModelPricing("gemini-3.7-flash")
	if pGemini.BlendedPer1M != 0.375 {
		t.Errorf("expected gemini-3.7-flash blended price 0.375, got %f", pGemini.BlendedPer1M)
	}

	// Test exact savings calculation for 100k tokens
	savings := CalculateSavingsUSD("claude-5-sonnet", 100_000)
	expectedSavings := 0.90 // 100k / 1M * $9.00 = $0.90
	if savings != expectedSavings {
		t.Errorf("expected savings $0.90, got %f", savings)
	}
}
