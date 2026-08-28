package gemini

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// AccountSession represents an authenticated Google account session with health state.
type AccountSession struct {
	ID           string
	SourceFile   string
	Cookie       string
	SAPISID      string
	AuthUser     string
	FailureCount int64
	LastFailure  time.Time
	Active       bool
}

// CookiePool manages a multi-account pool of Google session cookies with auto-rotation and failover.
type CookiePool struct {
	mu        sync.RWMutex
	sessions  []*AccountSession
	sources   []string
	sourceDir string
	cursor    atomic.Uint64
}

// NewCookiePool initializes an empty or populated cookie pool.
func NewCookiePool() *CookiePool {
	return &CookiePool{
		sessions: make([]*AccountSession, 0),
		sources:  make([]string, 0),
	}
}

// AddSession registers an account session into the pool.
func (p *CookiePool) AddSession(source, cookieStr, sapisid, authUser string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	id := sapisid
	if id == "" {
		id = source
	}
	if id == "" {
		id = "session_" + cookieStr[:min(8, len(cookieStr))]
	}

	session := &AccountSession{
		ID:         id,
		SourceFile: source,
		Cookie:     cookieStr,
		SAPISID:    sapisid,
		AuthUser:   authUser,
		Active:     true,
	}
	p.sessions = append(p.sessions, session)
}

// LoadFromFiles loads multiple cookie files into the pool.
func (p *CookiePool) LoadFromFiles(files []string) int {
	p.mu.Lock()
	p.sources = append(p.sources, files...)
	p.mu.Unlock()

	loaded := 0
	for _, f := range files {
		if strings.TrimSpace(f) == "" {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		raw := strings.TrimSpace(string(data))
		if raw == "" {
			continue
		}
		extracted, err := ExtractCookies(raw)
		if err == nil && extracted.RawCookie != "" {
			p.AddSession(f, extracted.RawCookie, extracted.SAPISID, "")
			loaded++
		}
	}
	return loaded
}

// LoadFromDirectory discovers all .txt cookie files in a directory.
func (p *CookiePool) LoadFromDirectory(dir string) int {
	p.mu.Lock()
	p.sourceDir = dir
	p.mu.Unlock()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".txt") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	return p.LoadFromFiles(files)
}

// Reload refreshes all session tokens from registered sources on disk without restarting the server.
func (p *CookiePool) Reload() int {
	p.mu.RLock()
	sources := append([]string{}, p.sources...)
	sourceDir := p.sourceDir
	p.mu.RUnlock()

	// Build a deduplicated set to avoid loading directory files twice
	// (they were already appended to p.sources in LoadFromDirectory)
	seen := make(map[string]bool, len(sources))
	for _, s := range sources {
		seen[s] = true
	}

	if sourceDir != "" {
		entries, err := os.ReadDir(sourceDir)
		if err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".txt") {
					p := filepath.Join(sourceDir, e.Name())
					if !seen[p] {
						sources = append(sources, p)
						seen[p] = true
					}
				}
			}
		}
	}

	loaded := 0
	var newSessions []*AccountSession

	for _, f := range sources {
		if strings.TrimSpace(f) == "" {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		raw := strings.TrimSpace(string(data))
		if raw == "" {
			continue
		}
		extracted, err := ExtractCookies(raw)
		if err == nil && extracted.RawCookie != "" {
			id := extracted.SAPISID
			if id == "" {
				id = f
			}
			newSessions = append(newSessions, &AccountSession{
				ID:         id,
				SourceFile: f,
				Cookie:     extracted.RawCookie,
				SAPISID:    extracted.SAPISID,
				Active:     true,
			})
			loaded++
		}
	}

	if loaded > 0 {
		p.mu.Lock()
		p.sessions = newSessions
		p.mu.Unlock()
	}
	return loaded
}

// StartAutoReload launches a lightweight background goroutine that periodically syncs session tokens.
func (p *CookiePool) StartAutoReload(interval time.Duration) {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			p.Reload()
		}
	}()
}

// Count returns the total number of sessions in the pool.
func (p *CookiePool) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.sessions)
}

// CountHealthy returns the number of active sessions not currently in a 60s failure cooldown.
func (p *CookiePool) CountHealthy() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	now := time.Now()
	healthy := 0
	for _, s := range p.sessions {
		if s.Active && now.Sub(s.LastFailure) > 60*time.Second {
			healthy++
		}
	}
	return healthy
}

// GetHealthySession selects the next healthy session using round-robin.
func (p *CookiePool) GetHealthySession() *AccountSession {
	p.mu.RLock()
	defer p.mu.RUnlock()

	total := len(p.sessions)
	if total == 0 {
		return nil
	}

	now := time.Now()
	if total == 1 {
		s := p.sessions[0]
		if s.Active && now.Sub(s.LastFailure) > 60*time.Second {
			return s
		}
		return nil
	}

	startIdx := p.cursor.Add(1) % uint64(total)

	// 1. First pass: find active session with no recent failure cooldown (60s)
	for i := 0; i < total; i++ {
		idx := (int(startIdx) + i) % total
		s := p.sessions[idx]
		if s.Active && now.Sub(s.LastFailure) > 60*time.Second {
			return s
		}
	}

	// 2. Fallback: return session with oldest failure
	var best *AccountSession
	for _, s := range p.sessions {
		if s.Active {
			if best == nil || s.LastFailure.Before(best.LastFailure) {
				best = s
			}
		}
	}

	return best
}

// MarkFailure records an upstream error and triggers failover cooldown for this session.
func (p *CookiePool) MarkFailure(sessionID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, s := range p.sessions {
		if s.ID == sessionID || s.SAPISID == sessionID {
			atomic.AddInt64(&s.FailureCount, 1)
			s.LastFailure = time.Now()
			log.Printf("[CookiePool] Session %s marked with failure (cooldown 60s)", s.ID[:min(10, len(s.ID))])
			break
		}
	}
}

// MarkSuccess resets consecutive failure penalties for this session.
func (p *CookiePool) MarkSuccess(sessionID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, s := range p.sessions {
		if s.ID == sessionID || s.SAPISID == sessionID {
			atomic.StoreInt64(&s.FailureCount, 0)
			s.LastFailure = time.Time{}
			break
		}
	}
}
