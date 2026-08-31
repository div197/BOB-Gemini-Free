package server

import (
	"context"
	"errors"
	"sync"
	"time"
)

const (
	maxImageCacheEntries = 256
	// Scotty reference lifetime is provider-dependent and is not advertised as
	// a durable contract. Expire local references conservatively so a stale
	// reference cannot be reused indefinitely if the provider invalidates it.
	maxImageCacheAge = 15 * time.Minute
)

type imageCacheEntry struct {
	ref      string
	lastUsed uint64
	storedAt time.Time
}

type imageCacheFlight struct {
	done chan struct{}
	ref  string
	err  error
}

// imageRefCache stores only small provider references, never the image bytes.
// The zero value is ready for use and the hard entry bound prevents an
// untrusted stream of distinct images from growing the process indefinitely.
type imageRefCache struct {
	mu      sync.Mutex
	entries map[string]imageCacheEntry
	flights map[string]*imageCacheFlight
	clock   uint64
	nowFn   func() time.Time // test hook; nil uses time.Now
}

func (c *imageRefCache) now() time.Time {
	if c != nil && c.nowFn != nil {
		return c.nowFn()
	}
	return time.Now()
}

func imageCacheEntryExpired(entry imageCacheEntry, now time.Time) bool {
	return !entry.storedAt.IsZero() && !now.Before(entry.storedAt.Add(maxImageCacheAge))
}

func (c *imageRefCache) Load(key string) (string, bool) {
	if c == nil || key == "" {
		return "", false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[key]
	if !ok || entry.ref == "" {
		return "", false
	}
	if imageCacheEntryExpired(entry, c.now()) {
		delete(c.entries, key)
		return "", false
	}
	c.clock++
	entry.lastUsed = c.clock
	c.entries[key] = entry
	return entry.ref, true
}

func (c *imageRefCache) Store(key, ref string) {
	if c == nil || key == "" || ref == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.storeLocked(key, ref)
}

func (c *imageRefCache) storeLocked(key, ref string) {
	if c.entries == nil {
		c.entries = make(map[string]imageCacheEntry, maxImageCacheEntries)
	}
	c.clock++
	if _, exists := c.entries[key]; !exists && len(c.entries) >= maxImageCacheEntries {
		oldestKey := ""
		oldestUse := ^uint64(0)
		for candidate, entry := range c.entries {
			if entry.lastUsed < oldestUse {
				oldestKey = candidate
				oldestUse = entry.lastUsed
			}
		}
		if oldestKey != "" {
			delete(c.entries, oldestKey)
		}
	}
	c.entries[key] = imageCacheEntry{ref: ref, lastUsed: c.clock, storedAt: c.now()}
}

// Do returns a cached reference or shares one in-flight upload for the same
// session-scoped key. The boolean is true when the result came from an
// existing cache entry or another caller's flight. A waiting caller may stop
// waiting through ctx without cancelling the leader's upload; the leader's
// request context still controls the actual network operation.
func (c *imageRefCache) Do(ctx context.Context, key string, fn func() (string, error)) (string, bool, error) {
	if c == nil {
		return "", false, errors.New("image cache is unavailable")
	}
	if key == "" {
		return "", false, errors.New("image cache key is empty")
	}
	if fn == nil {
		return "", false, errors.New("image cache loader is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	c.mu.Lock()
	if entry, ok := c.entries[key]; ok && entry.ref != "" {
		if imageCacheEntryExpired(entry, c.now()) {
			delete(c.entries, key)
		} else {
			c.clock++
			entry.lastUsed = c.clock
			c.entries[key] = entry
			c.mu.Unlock()
			return entry.ref, true, nil
		}
	}
	if flight, ok := c.flights[key]; ok {
		done := flight.done
		c.mu.Unlock()
		select {
		case <-done:
			return flight.ref, true, flight.err
		case <-ctx.Done():
			return "", true, ctx.Err()
		}
	}
	if c.flights == nil {
		c.flights = make(map[string]*imageCacheFlight)
	}
	flight := &imageCacheFlight{done: make(chan struct{})}
	c.flights[key] = flight
	c.mu.Unlock()

	ref, err := fn()
	c.mu.Lock()
	delete(c.flights, key)
	flight.ref = ref
	flight.err = err
	if err == nil && ref != "" {
		c.storeLocked(key, ref)
	}
	close(flight.done)
	c.mu.Unlock()
	return ref, false, err
}

func (c *imageRefCache) Len() int {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.entries)
}

func (c *imageRefCache) Clear() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = nil
}
