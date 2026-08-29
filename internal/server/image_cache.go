package server

import "sync"

const maxImageCacheEntries = 256

type imageCacheEntry struct {
	ref      string
	lastUsed uint64
}

// imageRefCache stores only small provider references, never the image bytes.
// The zero value is ready for use and the hard entry bound prevents an
// untrusted stream of distinct images from growing the process indefinitely.
type imageRefCache struct {
	mu      sync.Mutex
	entries map[string]imageCacheEntry
	clock   uint64
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
	c.entries[key] = imageCacheEntry{ref: ref, lastUsed: c.clock}
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
