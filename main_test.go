package main

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestLaunchStudioOrFallbackRunsFallbackOnce(t *testing.T) {
	var fallbackURLs []string
	launch := func(context.Context, int, func(string, ...any)) error {
		return errors.New("no app-mode browser")
	}

	if err := launchStudioOrFallback(context.Background(), 9610, launch, func(url string) error {
		fallbackURLs = append(fallbackURLs, url)
		return nil
	}); err != nil {
		t.Fatalf("launchStudioOrFallback: %v", err)
	}

	if len(fallbackURLs) != 1 {
		t.Fatalf("fallback calls = %d, want 1", len(fallbackURLs))
	}
	if fallbackURLs[0] != "http://localhost:9610/playground" {
		t.Fatalf("fallback URL = %q", fallbackURLs[0])
	}
}

func TestLaunchStudioOrFallbackReportsFallbackFailure(t *testing.T) {
	launch := func(context.Context, int, func(string, ...any)) error {
		return errors.New("no app-mode browser")
	}
	fallbackErr := errors.New("browser command unavailable")
	err := launchStudioOrFallback(context.Background(), 9610, launch, func(string) error {
		return fallbackErr
	})
	if err == nil || !strings.Contains(err.Error(), "browser fallback failed") || !strings.Contains(err.Error(), fallbackErr.Error()) {
		t.Fatalf("error = %v, want fallback failure", err)
	}
}

func TestNewStatusRequestValidatesURLAndAuthHeader(t *testing.T) {
	req, err := newStatusRequest("http://127.0.0.1:9610", "status-key")
	if err != nil {
		t.Fatalf("newStatusRequest: %v", err)
	}
	if req.URL.String() != "http://127.0.0.1:9610/" {
		t.Fatalf("status URL = %q", req.URL)
	}
	if req.Header.Get("Authorization") != "Bearer status-key" {
		t.Fatalf("status authorization = %q", req.Header.Get("Authorization"))
	}
	if _, err := newStatusRequest("://not-a-url", ""); err == nil {
		t.Fatal("invalid status URL was accepted")
	}
}
