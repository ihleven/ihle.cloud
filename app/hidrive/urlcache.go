package hidrive

import (
	"net/url"
	"sync"
	"time"
)

// signedURLTTL is how long a minted URL is reused.
//
// Deliberately short, and deliberately not the store's own lifetime — which is
// not knowable. A pre-signed URL comes back as https://stra.to/<three path
// segments> with no query string, so there is no expiry to read and no header
// that reports one. Measured, minting costs ~48 ms against a ~110 ms fetch, so
// a couple of minutes already removes that from everything a document viewer
// does in a sitting, and guessing longer would buy little for the risk.
//
// The guess cannot be wrong in a way that breaks a request: a URL the store no
// longer honours is discarded and re-minted, so this is a performance setting
// and not a correctness one. That is the only safe way to cache something whose
// lifetime cannot be inspected.
const signedURLTTL = 2 * time.Minute

// signedURLs remembers pre-signed URLs for the paths they address.
//
// Keyed by the *resolved* store path rather than the path as asked for, so two
// accounts whose roots put them at the same file share one entry — and so a
// path spelled differently does not mint twice.
//
// The URLs never leave this process: they are what the proxy fetches from, not
// what the browser is handed. Holding a bearer credential in memory is the same
// trade the access token itself already makes.
type signedURLs struct {
	ttl time.Duration

	mu      sync.Mutex
	entries map[string]signedURL
}

type signedURL struct {
	url     *url.URL
	expires time.Time
}

func newSignedURLs(ttl time.Duration) *signedURLs {
	if ttl <= 0 {
		ttl = signedURLTTL
	}

	return &signedURLs{ttl: ttl, entries: map[string]signedURL{}}
}

func (c *signedURLs) get(key string) (*url.URL, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]
	if !ok || time.Now().After(entry.expires) {
		return nil, false
	}

	return entry.url, true
}

func (c *signedURLs) put(key string, u *url.URL) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = signedURL{url: u, expires: time.Now().Add(c.ttl)}
}

// drop forgets an entry the store has stopped honouring, so the next caller
// mints rather than repeating the same refusal.
func (c *signedURLs) drop(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}
