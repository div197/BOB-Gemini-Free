package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/div197/bob-gemini-free/internal/config"
)

// cspDirective returns the value of one CSP directive, or "" when absent.
func cspDirective(policy, name string) string {
	for _, part := range strings.Split(policy, ";") {
		fields := strings.Fields(part)
		if len(fields) > 0 && fields[0] == name {
			return strings.Join(fields[1:], " ")
		}
	}
	return ""
}

// TestStudioImageSourcesAreRestricted is the H2 regression. The studio CSP
// used default-src with https:, so a model-generated markdown image could load
// from any HTTPS host. Image loads must be limited to local data, blob, and
// the Google image host that the gateway itself recognises.
func TestStudioImageSourcesAreRestricted(t *testing.T) {
	app := New(config.Default(), "csp-test")
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/playground", nil))

	policy := rec.Header().Get("Content-Security-Policy")
	imgSrc := cspDirective(policy, "img-src")
	if imgSrc == "" {
		t.Fatalf("studio CSP has no img-src directive: %q", policy)
	}
	want := "'self' data: blob: https://lh3.googleusercontent.com"
	if imgSrc != want {
		t.Fatalf("img-src = %q, want %q", imgSrc, want)
	}
	for _, source := range strings.Fields(imgSrc) {
		if source == "https:" || source == "*" {
			t.Fatalf("img-src must not allow every HTTPS host: %q", imgSrc)
		}
	}
}

// TestStudioMarkdownImagesUseTheAllowedHostList ties the markdown image
// renderer to the same host list, so a model-chosen image host cannot render
// as an <img> that the CSP would then have to block silently.
func TestStudioMarkdownImagesUseTheAllowedHostList(t *testing.T) {
	app := New(config.Default(), "renderer-test")
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/playground", nil))
	body := rec.Body.String()

	for _, marker := range []string{
		"const MARKDOWN_IMAGE_HOSTS = ['lh3.googleusercontent.com'];",
		"function isAllowedMarkdownImageUrl(url)",
		"if (!isAllowedMarkdownImageUrl(rawHref)) {",
	} {
		if !strings.Contains(body, marker) {
			t.Fatalf("studio markdown image renderer missing %q", marker)
		}
	}
}
