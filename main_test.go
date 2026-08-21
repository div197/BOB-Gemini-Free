package main

import (
	"context"
	"errors"
	"testing"
)

func TestLaunchStudioOrFallbackRunsFallbackOnce(t *testing.T) {
	var fallbackURLs []string
	launch := func(context.Context, int, func(string, ...any)) error {
		return errors.New("no app-mode browser")
	}

	launchStudioOrFallback(context.Background(), 9610, launch, func(url string) {
		fallbackURLs = append(fallbackURLs, url)
	})

	if len(fallbackURLs) != 1 {
		t.Fatalf("fallback calls = %d, want 1", len(fallbackURLs))
	}
	if fallbackURLs[0] != "http://localhost:9610/playground" {
		t.Fatalf("fallback URL = %q", fallbackURLs[0])
	}
}
