package blob

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

// countingStore records how often it is asked, which is the whole point of the
// cache: fewer upstream calls.
type countingStore struct {
	mu        sync.Mutex
	statCalls int
	openCalls int
	obj       Object
	statErr   error
}

func (c *countingStore) Stat(context.Context, string) (Object, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.statCalls++
	if c.statErr != nil {
		return Object{}, c.statErr
	}
	return c.obj, nil
}

func (c *countingStore) OpenRange(_ context.Context, _ string, _, n int64) (io.ReadCloser, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.openCalls++
	return io.NopCloser(strings.NewReader(strings.Repeat("x", int(n)))), nil
}

func (c *countingStore) counts() (int, int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.statCalls, c.openCalls
}

func TestStatCacheServesRepeatedStatsFromMemory(t *testing.T) {
	inner := &countingStore{obj: Object{Key: "clip.mp4", Size: 42, ETag: `"abc"`}}
	c := NewStatCache(inner, time.Minute)
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		obj, err := c.Stat(ctx, "clip.mp4")
		if err != nil {
			t.Fatalf("Stat: %v", err)
		}
		if obj.Size != 42 || obj.ETag != `"abc"` {
			t.Fatalf("Stat returned %+v, want the inner object", obj)
		}
	}

	if stats, _ := inner.counts(); stats != 1 {
		t.Errorf("upstream Stat called %d times, want 1", stats)
	}
	if s := c.CacheStats(); s.Hits != 9 || s.Misses != 1 {
		t.Errorf("CacheStats = %+v, want 9 hits and 1 miss", s)
	}
}

func TestStatCacheDistinguishesKeys(t *testing.T) {
	inner := &countingStore{obj: Object{Size: 1}}
	c := NewStatCache(inner, time.Minute)
	ctx := context.Background()

	for _, k := range []string{"a.mp4", "b.mp4", "a.mp4", "b.mp4"} {
		if _, err := c.Stat(ctx, k); err != nil {
			t.Fatal(err)
		}
	}
	if stats, _ := inner.counts(); stats != 2 {
		t.Errorf("upstream Stat called %d times, want 2 (one per key)", stats)
	}
}

func TestStatCacheExpires(t *testing.T) {
	inner := &countingStore{obj: Object{Size: 1}}
	c := NewStatCache(inner, 20*time.Millisecond)
	ctx := context.Background()

	if _, err := c.Stat(ctx, "clip.mp4"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(40 * time.Millisecond)
	if _, err := c.Stat(ctx, "clip.mp4"); err != nil {
		t.Fatal(err)
	}
	if stats, _ := inner.counts(); stats != 2 {
		t.Errorf("upstream Stat called %d times, want 2 (the entry should expire)", stats)
	}
}

// Caching a not-found would hide a freshly uploaded file for the whole TTL,
// which is worse than repeating the lookup.
func TestStatCacheDoesNotCacheErrors(t *testing.T) {
	inner := &countingStore{statErr: ErrNotFound}
	c := NewStatCache(inner, time.Minute)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if _, err := c.Stat(ctx, "missing.mp4"); !errors.Is(err, ErrNotFound) {
			t.Fatalf("Stat = %v, want ErrNotFound", err)
		}
	}
	if stats, _ := inner.counts(); stats != 3 {
		t.Errorf("upstream Stat called %d times, want 3 (errors must not be cached)", stats)
	}
}

// Payload bytes must never be served from cache, only metadata.
func TestStatCachePassesOpenRangeThrough(t *testing.T) {
	inner := &countingStore{obj: Object{Size: 100}}
	c := NewStatCache(inner, time.Minute)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		rc, err := c.OpenRange(ctx, "clip.mp4", 0, 10)
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(rc)
		rc.Close()
		if len(b) != 10 {
			t.Fatalf("read %d bytes, want 10", len(b))
		}
	}
	if _, opens := inner.counts(); opens != 3 {
		t.Errorf("upstream OpenRange called %d times, want 3 (never cached)", opens)
	}
}

func TestStatCacheInvalidate(t *testing.T) {
	inner := &countingStore{obj: Object{Size: 1}}
	c := NewStatCache(inner, time.Minute)
	ctx := context.Background()

	c.Stat(ctx, "clip.mp4")
	c.Invalidate("clip.mp4")
	c.Stat(ctx, "clip.mp4")

	if stats, _ := inner.counts(); stats != 2 {
		t.Errorf("upstream Stat called %d times, want 2 after invalidation", stats)
	}
}

func TestStatCacheIsConcurrencySafe(t *testing.T) {
	inner := &countingStore{obj: Object{Size: 1}}
	c := NewStatCache(inner, time.Minute)
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "clip.mp4"
			if i%2 == 0 {
				key = "other.mp4"
			}
			if _, err := c.Stat(ctx, key); err != nil {
				t.Errorf("Stat: %v", err)
			}
			c.CacheStats()
		}(i)
	}
	wg.Wait()
}

// The map must not grow without bound in a long-running process.
func TestStatCacheIsBounded(t *testing.T) {
	inner := &countingStore{obj: Object{Size: 1}}
	c := NewStatCache(inner, time.Minute)
	ctx := context.Background()

	for i := 0; i < statCacheLimit+200; i++ {
		if _, err := c.Stat(ctx, "key-"+string(rune('a'+i%26))+string(rune(i))); err != nil {
			t.Fatal(err)
		}
	}
	if got := c.CacheStats().Entries; got > statCacheLimit {
		t.Errorf("cache holds %d entries, want at most %d", got, statCacheLimit)
	}
}

func TestNewStatCacheDefaultsTTL(t *testing.T) {
	c := NewStatCache(&countingStore{}, 0)
	if c.ttl != DefaultStatTTL {
		t.Errorf("ttl = %v, want %v", c.ttl, DefaultStatTTL)
	}
}

var _ Blobstore = (*StatCache)(nil)
