package gemini

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const cookieSessionCooldown = 60 * time.Second

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
	source = strings.TrimSpace(source)
	if source != "" {
		source = filepath.Clean(source)
	}
	cookieStr = strings.TrimSpace(cookieStr)
	if cookieStr == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	id := sessionIdentity(source, cookieStr, sapisid)
	for _, existing := range p.sessions {
		if existing.ID == id || (source != "" && existing.SourceFile == source) {
			existing.ID = id
			existing.SourceFile = source
			existing.Cookie = cookieStr
			existing.SAPISID = sapisid
			existing.AuthUser = authUser
			existing.Active = true
			return
		}
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
	for _, file := range files {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}
		file = filepath.Clean(file)
		if !containsString(p.sources, file) {
			p.sources = append(p.sources, file)
		}
	}
	p.mu.Unlock()

	loaded := 0
	for _, f := range files {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		f = filepath.Clean(f)
		_, data, err := readSecureCookieFile(f)
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
	dir = strings.TrimSpace(dir)
	if dir != "" {
		dir = filepath.Clean(dir)
	}
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
	oldByID := make(map[string]sessionState, len(p.sessions))
	oldBySource := make(map[string]sessionState, len(p.sessions))
	for _, session := range p.sessions {
		if session == nil {
			continue
		}
		state := sessionState{
			failureCount: atomic.LoadInt64(&session.FailureCount),
			lastFailure:  session.LastFailure,
			active:       session.Active,
		}
		oldByID[session.ID] = state
		if session.SourceFile != "" {
			oldBySource[session.SourceFile] = state
		}
	}
	p.mu.RUnlock()

	// Build a deduplicated set to avoid loading directory files twice
	// (they were already appended to p.sources in LoadFromDirectory)
	seen := make(map[string]bool, len(sources))
	for _, s := range sources {
		s = filepath.Clean(strings.TrimSpace(s))
		if s == "." {
			continue
		}
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
	seenIDs := make(map[string]struct{})

	for _, f := range sources {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		f = filepath.Clean(f)
		_, data, err := readSecureCookieFile(f)
		if err != nil {
			continue
		}
		raw := strings.TrimSpace(string(data))
		if raw == "" {
			continue
		}
		extracted, err := ExtractCookies(raw)
		if err == nil && extracted.RawCookie != "" {
			id := sessionIdentity(f, extracted.RawCookie, extracted.SAPISID)
			if _, exists := seenIDs[id]; exists {
				continue
			}
			seenIDs[id] = struct{}{}
			session := &AccountSession{
				ID:         id,
				SourceFile: f,
				Cookie:     extracted.RawCookie,
				SAPISID:    extracted.SAPISID,
				Active:     true,
			}
			if state, ok := oldByID[id]; ok {
				session.FailureCount = state.failureCount
				session.LastFailure = state.lastFailure
				session.Active = state.active
			} else if state, ok := oldBySource[f]; ok {
				session.FailureCount = state.failureCount
				session.LastFailure = state.lastFailure
				session.Active = state.active
			}
			newSessions = append(newSessions, session)
			loaded++
		}
	}

	if len(sources) > 0 || sourceDir != "" {
		p.mu.Lock()
		p.sessions = newSessions
		p.mu.Unlock()
	}
	return loaded
}

// StartAutoReload launches a lightweight background goroutine that periodically syncs session tokens.
func (p *CookiePool) StartAutoReload(interval time.Duration) {
	p.StartAutoReloadContext(context.Background(), interval)
}

// StartAutoReloadContext launches a reload loop that can be stopped with the
// returned function. The context-aware form prevents a long-lived gateway
// shutdown from leaking a ticker goroutine.
func (p *CookiePool) StartAutoReloadContext(ctx context.Context, interval time.Duration) (stop func()) {
	if ctx == nil {
		ctx = context.Background()
	}
	if interval <= 0 {
		interval = 30 * time.Second
	}
	stopCh := make(chan struct{})
	var stopOnce sync.Once
	stop = func() { stopOnce.Do(func() { close(stopCh) }) }
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				p.Reload()
			case <-stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
	return stop
}

// Count returns the total number of sessions in the pool.
func (p *CookiePool) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.sessions)
}

// CountHealthy returns the number of active sessions not currently in a failure cooldown.
func (p *CookiePool) CountHealthy() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	now := time.Now()
	healthy := 0
	for _, s := range p.sessions {
		if s.Active && now.Sub(s.LastFailure) > cookieSessionCooldown {
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
		if s.Active && now.Sub(s.LastFailure) > cookieSessionCooldown {
			return cloneAccountSession(s)
		}
		return nil
	}

	startIdx := p.cursor.Add(1) % uint64(total)

	// 1. First pass: find active session with no recent failure cooldown (60s)
	for i := 0; i < total; i++ {
		idx := (int(startIdx) + i) % total
		s := p.sessions[idx]
		if s.Active && now.Sub(s.LastFailure) > cookieSessionCooldown {
			return cloneAccountSession(s)
		}
	}

	// Never bypass the cooldown by returning an unhealthy session. The caller
	// can surface a bounded unavailable state instead of amplifying a provider
	// policy/rate-limit decision.
	return nil
}

func cloneAccountSession(s *AccountSession) *AccountSession {
	if s == nil {
		return nil
	}
	copy := *s
	copy.FailureCount = atomic.LoadInt64(&s.FailureCount)
	return &copy
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

type sessionState struct {
	failureCount int64
	lastFailure  time.Time
	active       bool
}

func sessionIdentity(source, cookieStr, sapisid string) string {
	material := strings.TrimSpace(sapisid)
	if material == "" {
		material = strings.TrimSpace(cookieStr)
	}
	if material == "" {
		material = strings.TrimSpace(source)
	}
	digest := sha256.Sum256([]byte(material))
	return "session_" + hex.EncodeToString(digest[:8])
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
