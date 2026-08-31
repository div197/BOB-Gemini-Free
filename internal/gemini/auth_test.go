package gemini

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
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

func TestCookieCacheReloadsContentWhenMtimeAndSizeAreUnchanged(t *testing.T) {
	cookieFile := filepath.Join(t.TempDir(), "cookie.txt")
	first := "SID=first_sid; SAPISID=first_sapisid"
	second := "SID=other_sid; SAPISID=other_sapisid"
	if len(first) != len(second) {
		t.Fatalf("fixture sizes differ: %d and %d", len(first), len(second))
	}
	if err := os.WriteFile(cookieFile, []byte(first), 0600); err != nil {
		t.Fatalf("write first cookie: %v", err)
	}
	stat, err := os.Stat(cookieFile)
	if err != nil {
		t.Fatalf("stat first cookie: %v", err)
	}
	cache := NewCookieCache(cookieFile)
	if _, err := cache.Load(); err != nil {
		t.Fatalf("load first cookie: %v", err)
	}
	if err := os.WriteFile(cookieFile, []byte(second), 0600); err != nil {
		t.Fatalf("write second cookie: %v", err)
	}
	if err := os.Chtimes(cookieFile, stat.ModTime(), stat.ModTime()); err != nil {
		t.Fatalf("restore cookie mtime: %v", err)
	}
	info, err := cache.Load()
	if err != nil {
		t.Fatalf("reload changed cookie: %v", err)
	}
	if info.SAPISID != "other_sapisid" {
		t.Fatalf("cached SAPISID = %q, want other_sapisid", info.SAPISID)
	}
}

func TestCookieCacheClearsCredentialsWhenFileDisappears(t *testing.T) {
	cookieFile := filepath.Join(t.TempDir(), "cookie.txt")
	if err := os.WriteFile(cookieFile, []byte("SID=sid; SAPISID=sapisid"), 0600); err != nil {
		t.Fatalf("write cookie: %v", err)
	}
	cache := NewCookieCache(cookieFile)
	if _, err := cache.Load(); err != nil {
		t.Fatalf("load cookie: %v", err)
	}
	if err := os.Remove(cookieFile); err != nil {
		t.Fatalf("remove cookie: %v", err)
	}
	info, err := cache.Load()
	if err == nil {
		t.Fatal("missing cookie file unexpectedly loaded")
	}
	if info.Cookie != "" || info.SAPISID != "" || info.At != "" {
		t.Fatalf("deleted cookie left authenticated state: %+v", info)
	}
}

func TestCookieCacheRejectsOversizedBootstrapDocument(t *testing.T) {
	cache := NewCookieCache("")
	requester := &mockSessionRequester{body: strings.Repeat("x", int(maxSessionHTMLBytes)+1)}
	at, bl, cookie, sapi := cache.GetSessionInfo(t.Context(), requester, "")
	if at != "" || bl != "" || cookie != "" || sapi != "" {
		t.Fatalf("oversized bootstrap returned session data: %q %q %q %q", at, bl, cookie, sapi)
	}
}

func TestReadSecureCookieFileRejectsBroadPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not authoritative on Windows")
	}
	cookieFile := filepath.Join(t.TempDir(), "cookie.txt")
	if err := os.WriteFile(cookieFile, []byte("SID=sid; SAPISID=sapisid"), 0644); err != nil {
		t.Fatalf("write cookie: %v", err)
	}
	if _, _, err := readSecureCookieFile(cookieFile); err == nil || !strings.Contains(err.Error(), "permissions") {
		t.Fatalf("broad cookie permissions were accepted: %v", err)
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

func TestExtractCookiesRejectsOversizedInput(t *testing.T) {
	raw := strings.Repeat("x", int(MaxCookieFileBytes)+1)
	if _, err := ExtractCookies(raw); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("oversized cookie input was accepted: %v", err)
	}
}

func TestReadSecureCookieFileRejectsOversizedContent(t *testing.T) {
	cookieFile := filepath.Join(t.TempDir(), "cookie.txt")
	if err := os.WriteFile(cookieFile, []byte(strings.Repeat("x", int(MaxCookieFileBytes)+1)), 0600); err != nil {
		t.Fatalf("write oversized cookie: %v", err)
	}
	if _, _, err := readSecureCookieFile(cookieFile); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("oversized cookie file was accepted: %v", err)
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
	body        string
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
		body:        `"SNlM0e":"guest-at-token","cfb2h":"guest-bl-build"`,
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

type rotatingSessionRequester struct {
	responses []string
	calls     int
}

func (r *rotatingSessionRequester) Do(req *http.Request) (*http.Response, error) {
	index := r.calls
	if index >= len(r.responses) {
		index = len(r.responses) - 1
	}
	r.calls++
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(r.responses[index])),
	}, nil
}

func TestCookieCacheInvalidateSessionForcesBootstrapRefresh(t *testing.T) {
	requester := &rotatingSessionRequester{responses: []string{
		`"SNlM0e":"first-at-token","cfb2h":"first-bl-build"`,
		`"SNlM0e":"second-at-token","cfb2h":"second-bl-build"`,
	}}
	cache := NewCookieCache("")

	at, bl, _, _ := cache.GetSessionInfo(t.Context(), requester, "")
	if at != "first-at-token" || bl != "first-bl-build" {
		t.Fatalf("initial session = %q %q, want first bootstrap", at, bl)
	}

	cache.InvalidateSession()
	at, bl, _, _ = cache.GetSessionInfo(t.Context(), requester, "")
	if at != "second-at-token" || bl != "second-bl-build" {
		t.Fatalf("refreshed session = %q %q, want second bootstrap", at, bl)
	}
	if requester.calls != 2 {
		t.Fatalf("bootstrap calls = %d, want 2 after invalidation", requester.calls)
	}
}

type blockingRotatingSessionRequester struct {
	responses []string
	started   chan struct{}
	release   chan struct{}
	calls     int
}

func (r *blockingRotatingSessionRequester) Do(req *http.Request) (*http.Response, error) {
	index := r.calls
	r.calls++
	if index == 0 {
		close(r.started)
		<-r.release
	}
	if index >= len(r.responses) {
		index = len(r.responses) - 1
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(r.responses[index])),
	}, nil
}

func TestCookieCacheInvalidationWinsOverInFlightBootstrap(t *testing.T) {
	requester := &blockingRotatingSessionRequester{
		responses: []string{
			`"SNlM0e":"old-at-token","cfb2h":"old-bl-build"`,
			`"SNlM0e":"fresh-at-token","cfb2h":"fresh-bl-build"`,
		},
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	cache := NewCookieCache("")
	result := make(chan struct {
		at string
		bl string
	}, 1)
	go func() {
		at, bl, _, _ := cache.GetSessionInfo(t.Context(), requester, "")
		result <- struct {
			at string
			bl string
		}{at: at, bl: bl}
	}()

	<-requester.started
	cache.InvalidateSession()
	close(requester.release)

	first := <-result
	if first.at != "" || first.bl != "" {
		t.Fatalf("in-flight bootstrap restored rejected tokens: %q %q", first.at, first.bl)
	}
	at, bl, _, _ := cache.GetSessionInfo(t.Context(), requester, "")
	if at != "fresh-at-token" || bl != "fresh-bl-build" {
		t.Fatalf("next bootstrap = %q %q, want fresh tokens", at, bl)
	}
	if requester.calls != 2 {
		t.Fatalf("bootstrap calls = %d, want 2", requester.calls)
	}
}

func TestCookieCacheInvalidateSessionPreservesConfiguredCookie(t *testing.T) {
	cache := NewCookieCache("")
	cache.mu.Lock()
	cache.info = CookieInfo{
		Cookie:       "SID=configured; SAPISID=secret",
		GuestCookies: "guest=1",
		SAPISID:      "secret",
		At:           "stale-at",
		BL:           "stale-bl",
		AtTime:       time.Now(),
	}
	cache.mu.Unlock()

	cache.InvalidateSession()
	info, err := cache.Load()
	if err != nil {
		t.Fatalf("Load after invalidation: %v", err)
	}
	if info.Cookie != "SID=configured; SAPISID=secret" || info.SAPISID != "secret" {
		t.Fatalf("configured credentials changed during invalidation: %+v", info)
	}
	if info.GuestCookies != "guest=1" || info.At != "" || info.BL != "" || !info.AtTime.IsZero() {
		t.Fatalf("unexpected session state after invalidation: %+v", info)
	}
}
