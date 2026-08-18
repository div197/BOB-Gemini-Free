package gemini

import (
	"net/url"
	"strings"
	"testing"

	"github.com/div197/bob-gemini-free/internal/config"
)

func TestPayloadBuilding(t *testing.T) {
	// 1. uuidV4 format verification
	u := uuidV4()
	if len(u) != 36 {
		t.Fatalf("Expected UUID length 36, got %d (%s)", len(u), u)
	}
	parts := strings.Split(u, "-")
	if len(parts) != 5 {
		t.Fatalf("Expected 5 UUID sections, got %d", len(parts))
	}

	// 2. BuildBody without fileRefs
	cfg := config.Default()
	cfg.XSRFToken = "test-xsrf-token"
	body := BuildBody("Hello Gemini", 1, 4, nil, nil, cfg)
	vals, err := url.ParseQuery(body)
	if err != nil {
		t.Fatalf("Failed to parse body form-encoding: %v", err)
	}
	if vals.Get("at") != "test-xsrf-token" {
		t.Errorf("Expected 'at' parameter 'test-xsrf-token', got %q", vals.Get("at"))
	}
	if !strings.Contains(vals.Get("f.req"), "Hello Gemini") {
		t.Errorf("Expected f.req to contain prompt")
	}

	// 3. BuildBody with fileRefs and extra indices
	extra := map[int]any{
		88: "custom_flag",
	}
	bodyRefs := BuildBody("Vision Query", 2, 0, []string{"/blob/123", "/blob/456"}, extra, cfg)
	valsRefs, _ := url.ParseQuery(bodyRefs)
	if !strings.Contains(valsRefs.Get("f.req"), "/blob/123") {
		t.Errorf("Expected f.req to contain file reference /blob/123")
	}

	// 4. BuildURL
	uStr := BuildURL(cfg)
	if !strings.Contains(uStr, "https://gemini.google.com") {
		t.Errorf("Expected BuildURL to start with gemini.google.com, got: %s", uStr)
	}
	if !strings.Contains(uStr, "rt=c") {
		t.Errorf("Expected BuildURL to include rt=c streaming parameter")
	}
}
