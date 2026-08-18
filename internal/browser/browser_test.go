package browser

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/div197/bob-gemini-free/internal/gemini"
)

func TestFindBrowser(t *testing.T) {
	info, err := FindBrowser()
	if err != nil {
		t.Logf("Browser not detected on this environment (expected on headless CI): %v", err)
		return
	}

	if info.Name == "" || info.Path == "" {
		t.Errorf("Expected valid browser name and path, got %+v", info)
	}
	t.Logf("Detected browser: %s at %s", info.Name, info.Path)
}

func TestGetFreePort(t *testing.T) {
	port, err := getFreePort()
	if err != nil {
		t.Fatalf("Failed to get free port: %v", err)
	}
	if port <= 1024 || port > 65535 {
		t.Errorf("Expected valid ephemeral port number, got %d", port)
	}
}

func TestCDPCookieParsing(t *testing.T) {
	rawJSON := `{
		"id": 2,
		"result": {
			"cookies": [
				{"name": "SAPISID", "value": "mock_sapisid_value_1234567890abcdef", "domain": ".google.com", "path": "/"},
				{"name": "__Secure-1PSID", "value": "mock_1psid_value_abcdef1234567890", "domain": ".google.com", "path": "/"},
				{"name": "SID", "value": "mock_sid_value_1234567890abcdef", "domain": ".google.com", "path": "/"},
				{"name": "HSID", "value": "mock_hsid_value_1234567890", "domain": ".google.com", "path": "/"},
				{"name": "SSID", "value": "mock_ssid_value_1234567890", "domain": ".google.com", "path": "/"}
			]
		}
	}`

	var resp cdpResponse
	if err := json.Unmarshal([]byte(rawJSON), &resp); err != nil {
		t.Fatalf("Failed to parse mock CDP response: %v", err)
	}

	var pairs []string
	for _, c := range resp.Result.Cookies {
		if strings.Contains(c.Domain, "google.com") {
			pairs = append(pairs, c.Name+"="+c.Value)
		}
	}

	if len(pairs) != 5 {
		t.Fatalf("Expected 5 cookie pairs, got %d", len(pairs))
	}

	rawCookie := strings.Join(pairs, "; ")
	extracted, err := gemini.ExtractCookies(rawCookie)
	if err != nil {
		t.Fatalf("ExtractCookies failed on CDP parsed cookies: %v", err)
	}

	if extracted.SAPISID != "mock_sapisid_value_1234567890abcdef" {
		t.Errorf("Unexpected SAPISID: %s", extracted.SAPISID)
	}
	if len(extracted.Tokens) != 5 {
		t.Errorf("Expected 5 extracted tokens, got %d", len(extracted.Tokens))
	}
}
