package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/div197/bob-gemini-free/internal/config"
)

func TestSecurityBoundaryRejectsUntrustedOrigin(t *testing.T) {
	cfg := config.Default()
	app := New(cfg, "security-baseline")
	handler := app.Handler()

	preflight := httptest.NewRequest(http.MethodOptions, "/v1/chat/completions", nil)
	preflight.Header.Set("Origin", "https://evil.example")
	preflight.Header.Set("Access-Control-Request-Method", http.MethodPost)
	preflight.Header.Set("Access-Control-Request-Headers", "content-type")
	preflightRec := httptest.NewRecorder()
	handler.ServeHTTP(preflightRec, preflight)
	if preflightRec.Code != http.StatusForbidden {
		t.Fatalf("untrusted preflight status = %d, want 403", preflightRec.Code)
	}
	if got := preflightRec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("untrusted allow-origin = %q, want empty", got)
	}

	actual := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	actual.Header.Set("Origin", "https://evil.example")
	actualRec := httptest.NewRecorder()
	handler.ServeHTTP(actualRec, actual)
	if actualRec.Code != http.StatusForbidden {
		t.Fatalf("untrusted API status = %d, want 403", actualRec.Code)
	}
}

func TestSecurityBoundaryAllowsExactLocalGatewayOrigin(t *testing.T) {
	app := New(config.Default(), "security-loopback")
	preflight := httptest.NewRequest(http.MethodOptions, "/v1/chat/completions", nil)
	preflight.Host = "127.0.0.1:39123"
	preflight.Header.Set("Origin", "http://127.0.0.1:39123")
	preflight.Header.Set("Access-Control-Request-Method", http.MethodPost)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, preflight)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("loopback preflight status = %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://127.0.0.1:39123" {
		t.Fatalf("loopback allow-origin = %q", got)
	}
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Fatalf("loopback Vary = %q, want Origin", got)
	}
}

func TestSecurityBoundaryRejectsDifferentLoopbackOrigin(t *testing.T) {
	app := New(config.Default(), "security-loopback-reject")
	preflight := httptest.NewRequest(http.MethodOptions, "/v1/chat/completions", nil)
	preflight.Host = "127.0.0.1:39123"
	preflight.Header.Set("Origin", "http://127.0.0.1:39124")
	preflight.Header.Set("Access-Control-Request-Method", http.MethodPost)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, preflight)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("different loopback preflight status = %d, want 403", rec.Code)
	}
}

func TestSecurityBoundaryRejectsNonLoopbackHostOriginConfusion(t *testing.T) {
	app := New(config.Default(), "security-host-confusion")
	preflight := httptest.NewRequest(http.MethodOptions, "/v1/chat/completions", nil)
	preflight.Host = "evil.example:39123"
	preflight.Header.Set("Origin", "http://evil.example:39123")
	preflight.Header.Set("Access-Control-Request-Method", http.MethodPost)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, preflight)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("host-header origin confusion status = %d, want 403", rec.Code)
	}
}

func TestSecurityBoundaryAllowsExplicitRemoteOrigin(t *testing.T) {
	cfg := config.Default()
	cfg.AllowedOrigins = []string{"https://studio.example.test"}
	app := New(cfg, "security-remote")
	preflight := httptest.NewRequest(http.MethodOptions, "/v1/chat/completions", nil)
	preflight.Header.Set("Origin", "https://studio.example.test")
	preflight.Header.Set("Access-Control-Request-Method", http.MethodPost)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, preflight)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("configured remote preflight status = %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://studio.example.test" {
		t.Fatalf("configured remote allow-origin = %q", got)
	}
}

func TestHealthzIsUnauthenticatedAndStable(t *testing.T) {
	cfg := config.Default()
	cfg.APIKeys = []string{"health-secret"}
	app := New(cfg, "healthz-test")
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("healthz status = %d, want 200", rec.Code)
	}
	if got := rec.Body.String(); got != "{\"status\":\"ok\"}\n" {
		t.Fatalf("healthz body = %q", got)
	}
	if got := rec.Header().Get("X-BOB-Gateway"); got != "bob-gemini-free" {
		t.Fatalf("healthz gateway identity = %q", got)
	}
	if got := rec.Header().Get("X-BOB-Protocol"); got != HealthzProtocolVersion {
		t.Fatalf("healthz protocol = %q", got)
	}
	if got := rec.Header().Get("X-BOB-Auth-Required"); got != "true" {
		t.Fatalf("healthz auth marker = %q", got)
	}
}

func TestProtectedRoutePrefixDoesNotBypassAPIKey(t *testing.T) {
	cfg := config.Default()
	cfg.APIKeys = []string{"route-secret"}
	app := New(cfg, "route-test")
	req := httptest.NewRequest(http.MethodGet, "/healthz-not-a-route", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("near-match health route status = %d, want 401", rec.Code)
	}
}

func TestMetricsEndpointRequiresAPIKeyAndReturnsAggregates(t *testing.T) {
	cfg := config.Default()
	cfg.APIKeys = []string{"metrics-secret"}
	app := New(cfg, "metrics-test")
	app.Metrics.RequestsTotal.Add(3)
	handler := app.Handler()

	unauthorized := httptest.NewRequest(http.MethodGet, "/v1/metrics", nil)
	unauthorizedRec := httptest.NewRecorder()
	handler.ServeHTTP(unauthorizedRec, unauthorized)
	if unauthorizedRec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized metrics status = %d", unauthorizedRec.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/metrics", nil)
	req.Header.Set("Authorization", "Bearer metrics-secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"requests_total":5`) {
		t.Fatalf("metrics response = status %d body %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(strings.ToLower(rec.Body.String()), "cookie") || strings.Contains(strings.ToLower(rec.Body.String()), "prompt") {
		t.Fatalf("metrics response contains sensitive field: %s", rec.Body.String())
	}
}
