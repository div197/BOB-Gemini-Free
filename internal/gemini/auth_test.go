package gemini

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnonymousCookieCache(t *testing.T) {
	cache := NewCookieCache("")
	info, err := cache.Load()
	if err != nil {
		t.Fatalf("unexpected error for empty cookie file: %v", err)
	}
	if info.Cookie != "" || info.SAPISID != "" {
		t.Errorf("expected empty info for anonymous cache, got %+v", info)
	}
}

func TestCookieCacheLoadRawText(t *testing.T) {
	tmpDir := t.TempDir()
	cookieFile := filepath.Join(tmpDir, "cookie.txt")
	content := "SID=test_sid; HSID=test_hsid; SAPISID=test_sapisid_value; __Secure-1PSID=test_1psid"
	if err := os.WriteFile(cookieFile, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test cookie file: %v", err)
	}

	cache := NewCookieCache(cookieFile)
	info, err := cache.Load()
	if err != nil {
		t.Fatalf("unexpected error loading cookie file: %v", err)
	}
	if info.Cookie != content {
		t.Errorf("got cookie %q, want %q", info.Cookie, content)
	}
	if info.SAPISID != "test_sapisid_value" {
		t.Errorf("got SAPISID %q, want %q", info.SAPISID, "test_sapisid_value")
	}
}

func TestCookieCacheLoadJSON(t *testing.T) {
	tmpDir := t.TempDir()
	cookieFile := filepath.Join(tmpDir, "cookie.json")
	content := `{"cookie": "SID=json_sid; SAPISID=json_sapisid", "sapisid": "json_sapisid"}`
	if err := os.WriteFile(cookieFile, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test json cookie file: %v", err)
	}

	cache := NewCookieCache(cookieFile)
	info, err := cache.Load()
	if err != nil {
		t.Fatalf("unexpected error loading json cookie file: %v", err)
	}
	if info.Cookie != "SID=json_sid; SAPISID=json_sapisid" {
		t.Errorf("got cookie %q", info.Cookie)
	}
	if info.SAPISID != "json_sapisid" {
		t.Errorf("got SAPISID %q, want json_sapisid", info.SAPISID)
	}
}

func TestSAPISIDHash(t *testing.T) {
	// Empty SAPISID should return empty hash
	if hash := SAPISIDHash(""); hash != "" {
		t.Errorf("expected empty hash for empty SAPISID, got %q", hash)
	}

	// Valid SAPISID should generate SAPISIDHASH timestamp_sha1
	hash := SAPISIDHash("my_secret_sapisid")
	if !strings.HasPrefix(hash, "SAPISIDHASH ") {
		t.Fatalf("expected hash to start with 'SAPISIDHASH ', got %q", hash)
	}
	parts := strings.Split(strings.TrimPrefix(hash, "SAPISIDHASH "), "_")
	if len(parts) != 2 {
		t.Fatalf("expected timestamp_hash format, got %q", hash)
	}
	if len(parts[1]) != 40 { // SHA1 hex is 40 chars
		t.Errorf("expected 40-char hex sha1, got %d chars: %s", len(parts[1]), parts[1])
	}
}

func TestAccountPrefix(t *testing.T) {
	if prefix := AccountPrefix(""); prefix != "" {
		t.Errorf("expected empty prefix for empty authUser, got %q", prefix)
	}
	if prefix := AccountPrefix("0"); prefix != "/u/0" {
		t.Errorf("expected /u/0, got %q", prefix)
	}
	if prefix := AccountPrefix("1"); prefix != "/u/1" {
		t.Errorf("expected /u/1, got %q", prefix)
	}
}

func TestExtractCookies(t *testing.T) {
	// Raw string with header prefix and multiple lines
	raw := "Cookie: SID=test_sid_123; HSID=test_hsid_456; SAPISID=test_sapisid_789; __Secure-1PSID=test_1psid_abc"
	extracted, err := ExtractCookies(raw)
	if err != nil {
		t.Fatalf("ExtractCookies failed: %v", err)
	}
	if extracted.SAPISID != "test_sapisid_789" {
		t.Errorf("expected SAPISID 'test_sapisid_789', got %q", extracted.SAPISID)
	}
	if extracted.Tokens["SID"] != "test_sid_123" {
		t.Errorf("expected SID 'test_sid_123', got %q", extracted.Tokens["SID"])
	}

	// Empty cookie should error
	if _, err := ExtractCookies(""); err == nil {
		t.Errorf("expected error for empty string")
	}
}

func TestSaveCookieFile(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "sub", "cookie.txt")
	cookieContent := "SID=saved_sid; SAPISID=saved_sapisid"

	if err := SaveCookieFile(targetPath, cookieContent); err != nil {
		t.Fatalf("SaveCookieFile failed: %v", err)
	}

	readBytes, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("failed to read saved cookie file: %v", err)
	}
	if string(readBytes) != cookieContent {
		t.Errorf("expected %q, got %q", cookieContent, string(readBytes))
	}

	// Check POSIX permissions
	info, err := os.Stat(targetPath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected 0600 permissions, got %o", perm)
	}
}

type mockSessionRequester struct {
	body string
	respCookies []string
}

func (m *mockSessionRequester) Do(req *http.Request) (*http.Response, error) {
	h := make(http.Header)
	for _, c := range m.respCookies {
		h.Add("Set-Cookie", c)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     h,
		Body:       io.NopCloser(strings.NewReader(m.body)),
	}, nil
}

func TestGetSessionInfoGuestAndCookie(t *testing.T) {
	// 1. Guest anonymous mode
	cache := NewCookieCache("")
	mockReq := &mockSessionRequester{
		body: `"SNlM0e":"guest-at-token","cfb2h":"guest-bl-build"`,
		respCookies: []string{"__Secure-BUCKET=b1; Path=/", "NID=nid1; Path=/"},
	}

	at, bl, cookie, sapi := cache.GetSessionInfo(t.Context(), mockReq, "")
	if at != "guest-at-token" {
		t.Fatalf("expected guest at token 'guest-at-token', got %q", at)
	}
	if bl != "guest-bl-build" {
		t.Fatalf("expected guest bl 'guest-bl-build', got %q", bl)
	}
	if !strings.Contains(cookie, "__Secure-BUCKET=b1") {
		t.Fatalf("expected guest cookies, got %q", cookie)
	}
	if sapi != "" {
		t.Fatalf("expected empty sapi for guest, got %q", sapi)
	}

	// 2. Cached check
	at2, bl2, _, _ := cache.GetSessionInfo(t.Context(), nil, "")
	if at2 != "guest-at-token" || bl2 != "guest-bl-build" {
		t.Fatalf("expected cached tokens, got %q %q", at2, bl2)
	}
}
