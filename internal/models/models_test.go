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
			input:       "unknown-model",
			defaultName: "gemini-3.6-flash",
			wantName:    "gemini-3.6-flash",
			wantMode:    1,
			wantThink:   4,
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
