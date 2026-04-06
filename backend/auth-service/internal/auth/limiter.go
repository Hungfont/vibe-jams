package auth

import (
	"sync"
	"time"
)

// FixedWindowLimiter is a deterministic in-memory fixed-window rate limiter.
type FixedWindowLimiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	now      func() time.Time
	counters map[string]fixedWindowCounter
}

type fixedWindowCounter struct {
	WindowStart time.Time
	Count       int
}

type LimitDecision struct {
	Allowed    bool
	RetryAfter time.Duration
}

func NewFixedWindowLimiter(limit int, window time.Duration, now func() time.Time) *FixedWindowLimiter {
	if now == nil {
		now = time.Now
	}
	if limit <= 0 {
		limit = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	return &FixedWindowLimiter{
		limit:    limit,
		window:   window,
		now:      now,
		counters: make(map[string]fixedWindowCounter),
	}
}

func (l *FixedWindowLimiter) Allow(key string) LimitDecision {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	counter := l.counters[key]
	if counter.WindowStart.IsZero() || now.Sub(counter.WindowStart) >= l.window {
		counter.WindowStart = now
		counter.Count = 0
	}
	if counter.Count >= l.limit {
		retryAfter := l.window - now.Sub(counter.WindowStart)
		if retryAfter < 0 {
			retryAfter = 0
		}
		return LimitDecision{Allowed: false, RetryAfter: retryAfter}
	}

	counter.Count++
	l.counters[key] = counter
	return LimitDecision{Allowed: true}
}

// LockoutTracker tracks repeated failures and applies deterministic backoff.
type LockoutTracker struct {
	mu        sync.Mutex
	threshold int
	backoff   time.Duration
	now       func() time.Time
	entries   map[string]lockoutEntry
}

type lockoutEntry struct {
	Failures  int
	LockedTil time.Time
}

func NewLockoutTracker(threshold int, backoff time.Duration, now func() time.Time) *LockoutTracker {
	if now == nil {
		now = time.Now
	}
	if threshold <= 0 {
		threshold = 1
	}
	if backoff <= 0 {
		backoff = time.Minute
	}
	return &LockoutTracker{
		threshold: threshold,
		backoff:   backoff,
		now:       now,
		entries:   make(map[string]lockoutEntry),
	}
}

func (t *LockoutTracker) IsLocked(identity string) (bool, time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	entry, ok := t.entries[identity]
	if !ok {
		return false, 0
	}
	now := t.now()
	if now.Before(entry.LockedTil) {
		return true, entry.LockedTil.Sub(now)
	}
	if !entry.LockedTil.IsZero() {
		entry.LockedTil = time.Time{}
		entry.Failures = 0
		t.entries[identity] = entry
	}
	return false, 0
}

func (t *LockoutTracker) RegisterFailure(identity string) (bool, time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	entry := t.entries[identity]
	if now.Before(entry.LockedTil) {
		return true, entry.LockedTil.Sub(now)
	}

	entry.Failures++
	if entry.Failures >= t.threshold {
		entry.Failures = 0
		entry.LockedTil = now.Add(t.backoff)
		t.entries[identity] = entry
		return true, t.backoff
	}
	t.entries[identity] = entry
	return false, 0
}

func (t *LockoutTracker) Reset(identity string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.entries, identity)
}
