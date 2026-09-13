package auth

import (
	"sync"
	"time"
)

// throttle limits how fast a given key — an account name, or a client address —
// may fail. Both are limited: per-account so one person's password cannot be
// ground down, per-address so a single client cannot sweep many auth.
//
// It is in-process, so it does not survive a restart and does not coordinate
// across replicas. For an application of this size that is the right trade:
// it raises the cost of an online guessing attack without a shared store.
type throttle struct {
	mu       sync.Mutex
	failures map[string]*failureCount
	limit    int
	window   time.Duration
}

type failureCount struct {
	n     int
	until time.Time
}

func newThrottle(limit int, window time.Duration) *throttle {
	return &throttle{failures: map[string]*failureCount{}, limit: limit, window: window}
}

// blocked reports whether a key has failed too often to be allowed another try.
func (t *throttle) blocked(keys ...string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	for _, k := range keys {
		f, ok := t.failures[k]
		if !ok {
			continue
		}
		if now.After(f.until) {
			delete(t.failures, k)
			continue
		}
		if f.n >= t.limit {
			return true
		}
	}
	return false
}

func (t *throttle) fail(keys ...string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	for _, k := range keys {
		f, ok := t.failures[k]
		if !ok || now.After(f.until) {
			f = &failureCount{}
			t.failures[k] = f
		}
		f.n++
		// Each failure extends the window, so a persistent attacker stays
		// blocked while someone who mistypes once is let through shortly.
		f.until = now.Add(t.window)
	}
}

func (t *throttle) succeed(keys ...string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, k := range keys {
		delete(t.failures, k)
	}
}

// sweep drops entries whose window has passed, so the map cannot grow without
// bound from failed attempts against random names.
func (t *throttle) sweep() {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	for k, f := range t.failures {
		if now.After(f.until) {
			delete(t.failures, k)
		}
	}
}
