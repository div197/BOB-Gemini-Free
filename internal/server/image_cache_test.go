package server

import (
	"context"
	"errors"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/gemini"
)

func TestImageRefCacheIsBoundedAndRefreshesLeastRecentlyUsedEntry(t *testing.T) {
	var cache imageRefCache
	cache.Store("old", "/old")
	for i := 0; i < maxImageCacheEntries-1; i++ {
		cache.Store("key-"+strconv.Itoa(i), "/ref")
	}

	if got := cache.Len(); got != maxImageCacheEntries {
		t.Fatalf("cache length = %d, want %d", got, maxImageCacheEntries)
	}
	if ref, ok := cache.Load("old"); !ok || ref != "/old" {
		t.Fatalf("recently used entry = (%q, %t), want (/old, true)", ref, ok)
	}

	cache.Store("new", "/new")
	if got := cache.Len(); got != maxImageCacheEntries {
		t.Fatalf("cache length after eviction = %d, want %d", got, maxImageCacheEntries)
	}
	if ref, ok := cache.Load("old"); !ok || ref != "/old" {
		t.Fatalf("recently used entry was evicted: (%q, %t)", ref, ok)
	}
	if _, ok := cache.Load("key-0"); ok {
		t.Fatal("least recently used entry was not evicted")
	}
	if ref, ok := cache.Load("new"); !ok || ref != "/new" {
		t.Fatalf("new entry = (%q, %t), want (/new, true)", ref, ok)
	}
}

func TestImageRefCacheZeroValueAndClear(t *testing.T) {
	var cache imageRefCache
	if _, ok := cache.Load("missing"); ok {
		t.Fatal("zero-value cache returned a missing entry")
	}
	cache.Store("key", "/ref")
	cache.Clear()
	if got := cache.Len(); got != 0 {
		t.Fatalf("cache length after clear = %d, want 0", got)
	}
}

func TestImageRefCacheExpiresReferencesAfterConservativeAge(t *testing.T) {
	now := time.Unix(1_000, 0)
	cache := imageRefCache{nowFn: func() time.Time { return now }}
	cache.Store("session-image", "/ref")

	if ref, ok := cache.Load("session-image"); !ok || ref != "/ref" {
		t.Fatalf("fresh reference = (%q, %t), want (/ref, true)", ref, ok)
	}

	now = now.Add(maxImageCacheAge)
	if ref, ok := cache.Load("session-image"); ok || ref != "" {
		t.Fatalf("expired reference = (%q, %t), want (empty, false)", ref, ok)
	}
}

func TestImageRefCacheRefreshesExpiredReference(t *testing.T) {
	now := time.Unix(2_000, 0)
	cache := imageRefCache{nowFn: func() time.Time { return now }}
	cache.Store("session-image", "/stale")
	now = now.Add(maxImageCacheAge)

	loaderCalls := 0
	ref, shared, err := cache.Do(context.Background(), "session-image", func() (string, error) {
		loaderCalls++
		return "/fresh", nil
	})
	if err != nil {
		t.Fatalf("refresh error = %v", err)
	}
	if shared || ref != "/fresh" {
		t.Fatalf("refresh result = %q shared:%t, want /fresh shared:false", ref, shared)
	}
	if loaderCalls != 1 {
		t.Fatalf("refresh loader calls = %d, want 1", loaderCalls)
	}
	if ref, ok := cache.Load("session-image"); !ok || ref != "/fresh" {
		t.Fatalf("refreshed cache entry = (%q, %t), want (/fresh, true)", ref, ok)
	}
}

func TestImageRefCacheSharesConcurrentLoader(t *testing.T) {
	var cache imageRefCache
	var mu sync.Mutex
	loadCount := 0
	loader := func() (string, error) {
		mu.Lock()
		loadCount++
		mu.Unlock()
		return "/shared-ref", nil
	}

	const callers = 32
	results := make(chan string, callers)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			ref, _, err := cache.Do(context.Background(), "same-session-image", loader)
			if err != nil {
				t.Errorf("Do() error = %v", err)
				return
			}
			results <- ref
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	if loadCount != 1 {
		t.Fatalf("loader call count = %d, want 1", loadCount)
	}
	for ref := range results {
		if ref != "/shared-ref" {
			t.Fatalf("shared reference = %q", ref)
		}
	}
}

func TestImageRefCacheWaiterCanCancelWithoutCancellingLeader(t *testing.T) {
	var cache imageRefCache
	started := make(chan struct{})
	finish := make(chan struct{})
	leaderDone := make(chan struct{})
	go func() {
		_, _, _ = cache.Do(context.Background(), "cancel-image", func() (string, error) {
			close(started)
			<-finish
			close(leaderDone)
			return "/ref", nil
		})
	}()
	<-started

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, shared, err := cache.Do(ctx, "cancel-image", func() (string, error) {
		t.Fatal("cancelled waiter became the leader")
		return "", nil
	}); !shared || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled waiter result = shared:%t err:%v", shared, err)
	}
	close(finish)
	<-leaderDone
	if ref, shared, err := cache.Do(context.Background(), "cancel-image", func() (string, error) {
		t.Fatal("completed leader result was not cached")
		return "", nil
	}); !shared || ref != "/ref" || err != nil {
		t.Fatalf("cached result = %q shared:%t err:%v", ref, shared, err)
	}
}

func TestImageCacheKeySeparatesAuthenticatedScopes(t *testing.T) {
	data := []byte("same image bytes")
	if imageCacheKey(data, "account-a") == imageCacheKey(data, "account-b") {
		t.Fatal("image cache key crossed authenticated scopes")
	}
	if imageCacheKey(data, "account-a") != imageCacheKey(data, "account-a") {
		t.Fatal("image cache key was not deterministic")
	}
}

func TestAuthenticatedImageCacheScopeRequiresSingleCookieSource(t *testing.T) {
	var noGem *App
	if _, ok := noGem.authenticatedImageCacheScope(); ok {
		t.Fatal("nil app unexpectedly enabled image cache")
	}

	cookiePath := filepath.Join(t.TempDir(), "cookie.txt")
	if err := gemini.SaveCookieFile(cookiePath, "SID=session-a; SAPISID=sapisid-a"); err != nil {
		t.Fatalf("save cookie: %v", err)
	}
	app := &App{
		Cfg: config.Config{AuthUser: "0"},
		Gem: &gemini.Client{Cookies: gemini.NewCookieCache(cookiePath)},
	}
	scopeA, ok := app.authenticatedImageCacheScope()
	if !ok || scopeA == "" {
		t.Fatal("single authenticated cookie source did not enable image cache")
	}

	app.Cfg.CookiePool = []string{"another-cookie.txt"}
	if _, ok := app.authenticatedImageCacheScope(); ok {
		t.Fatal("multi-account cookie pool unexpectedly enabled session-bound image cache")
	}

	app.Cfg.CookiePool = nil
	if err := gemini.SaveCookieFile(cookiePath, "SID=session-b; SAPISID=sapisid-b"); err != nil {
		t.Fatalf("rotate cookie: %v", err)
	}
	scopeB, ok := app.authenticatedImageCacheScope()
	if !ok || scopeA == scopeB {
		t.Fatal("cookie rotation did not produce a new image cache scope")
	}
}
