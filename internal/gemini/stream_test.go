package gemini

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStreamParserAndParsing(t *testing.T) {
	parser := NewStreamParser()
	parser.ResetBuffer()

	// Construct mock Google BoQ line
	// Structure: [["wrb.fr", null, "[null, null, null, null, [[\"chunk1\", null, null]], null, [\"resp_id\"]]\n"]]
	innerParts := []any{"Hello world"}
	innerPayload := []any{nil, nil, nil, nil, innerParts}
	innerBytes, _ := json.Marshal(innerPayload)

	outerItem := []any{"wrb.fr", nil, string(innerBytes) + "\n"}
	outerPayload := []any{outerItem}
	outerBytes, _ := json.Marshal(outerPayload)

	mockLine := string(outerBytes) + "\n"

	// Feed chunk into parser
	_, err := parser.Feed(mockLine)
	if err != nil {
		t.Fatalf("StreamParser Feed failed: %v", err)
	}

	// Test IsBardError detection
	bardErrStr := "Some response BardErrorInfo [42901] happened"
	code, ok := IsBardError(bardErrStr)
	if !ok || code != "42901" {
		t.Errorf("Expected BardErrorInfo 42901, got %s, %v", code, ok)
	}

	// Feed Bard error into parser
	parserErr := NewStreamParser()
	_, feedErr := parserErr.Feed("Error: BardErrorInfo [50001]\n")
	if feedErr == nil || !strings.Contains(feedErr.Error(), "50001") {
		t.Errorf("Expected Feed to return BardErrorInfo error, got: %v", feedErr)
	}
}

func TestFingerprintProfiles(t *testing.T) {
	profiles := []string{
		"chrome_120", "chrome_124", "chrome_131", "chrome_133", "chrome_144", "chrome_146",
		"firefox_120", "firefox_123", "firefox_147", "safari_16_0", "safari_ios_17_0", "unknown_profile",
	}

	for _, p := range profiles {
		_ = ResolveFingerprint(p)
	}

	client, err := getTLSClient("chrome_133", 10)
	if err != nil {
		t.Fatalf("getTLSClient failed: %v", err)
	}
	if client == nil {
		t.Fatalf("Expected non-nil client adapter")
	}

	// Repeated call should retrieve cached adapter
	clientCached, err := getTLSClient("chrome_133", 10)
	if err != nil || clientCached != client {
		t.Errorf("Expected cached client adapter")
	}
}
