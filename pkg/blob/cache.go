package blob

import (
	"context"
	"io"
	"sync"
	"time"
)

// DefaultStatTTL is how long object metadata is trusted. Media metadata changes
// only when a file is replaced, so a minute is generous for correctness and
// long enough to matter: it removes one upstream round-trip from every request
// after the first.
const DefaultStatTTL = time.Minute

// statCacheLimit bounds the map so a long-running process cannot accumulate an
// entry per key it has ever seen.
const statCacheLimit = 1024

// StatCache wraps a Blobstore and memoises Stat for a short period.
//
// The handler needs metadata on every request — to build Content-Range, to
// answer HEAD, and to serve validators for conditional requests — but for a
// remote store each of those is a network round-trip. Measured against HiDrive
// that was ~160 ms added to every ranged read, which is a third of the
// round-trip cost of a seek. Reads of the payload are never cached; only the
// metadata is.
type StatCache struct {
	store Blobstore
	ttl   time.Duration

	mu      sync.Mutex
	entries map[string]statEntry
	hits    int
	misses  int
}

type statEntry struct {
	obj     Object
	expires time.Time
}

// NewStatCache wraps store. A ttl of zero uses DefaultStatTTL.
func NewStatCache(store Blobstore, ttl time.Duration) *StatCache {
	if ttl <= 0 {
		ttl = DefaultStatTTL
	}
	return &StatCache{store: store, ttl: ttl, entries: map[string]statEntry{}}
}

func (c *StatCache) Stat(ctx context.Context, key string) (Object, error) {
	now := time.Now()

	c.mu.Lock()
	if e, ok := c.entries[key]; ok && now.Before(e.expires) {
		c.hits++
		c.mu.Unlock()
		return e.obj, nil
	}
	c.misses++
	c.mu.Unlock()

	// Deliberately not holding the lock across the call: a slow upstream would
	// otherwise block every other key too. Concurrent misses on the same key
	// may both fetch, which costs a duplicate request but never a wrong answer.
	obj, err := c.store.Stat(ctx, key)
	if err != nil {
		// Errors are not cached. Caching a not-found would hide a file that has
		// just been uploaded for the whole TTL, which is a worse failure than
		// an occasional extra lookup.
		return Object{}, err
	}

	c.mu.Lock()
	if len(c.entries) >= statCacheLimit {
		for k, e := range c.entries {
			if !now.Before(e.expires) {
				delete(c.entries, k)
			}
		}
		// Still full of live entries: drop one rather than grow without bound.
		if len(c.entries) >= statCacheLimit {
			for k := range c.entries {
				delete(c.entries, k)
				break
			}
		}
	}
	c.entries[key] = statEntry{obj: obj, expires: now.Add(c.ttl)}
	c.mu.Unlock()

	return obj, nil
}

// OpenRange passes straight through: payload bytes are never cached.
func (c *StatCache) OpenRange(ctx context.Context, key string, off, n int64) (io.ReadCloser, error) {
	return c.store.OpenRange(ctx, key, off, n)
}

// Invalidate drops a key, for when the app learns an object has changed.
func (c *StatCache) Invalidate(key string) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}

// CacheStats reports effectiveness, so /healthz can show whether the cache is
// doing anything.
type CacheStats struct {
	Hits    int
	Misses  int
	Entries int
}

func (c *StatCache) CacheStats() CacheStats {
	c.mu.Lock()
	defer c.mu.Unlock()
	return CacheStats{Hits: c.hits, Misses: c.misses, Entries: len(c.entries)}
}
