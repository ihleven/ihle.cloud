package hidrive

import (
	"context"
	"net/http"

	"github.com/ihleven/ihlvn/pkg/blob"
	"github.com/ihleven/ihlvn/pkg/hi"
	"github.com/ihleven/ihlvn/pkg/stream"
)

// Stream serves a file's bytes by reading them, where Media hands the browser a
// pre-signed URL and proxies to it.
//
// The difference is one round-trip per request. Media asks for a fresh signed
// URL every time; this asks for the object's metadata, which a cache answers for
// a minute at a time. Measured on a 45 MB film: ~44 ms to first byte against
// ~104 ms, with the protocol and throughput otherwise identical. A file fetched
// once will not notice. A video being scrubbed issues a request per seek, and
// does.
//
// It also keeps the store credential on this side. A pre-signed URL is a
// shareable, if short-lived, grant on the storage backend; these bytes arrive
// gated by the same session as everything else.
//
// Nothing here decides what a path may reach. The store applies blob.SafeKey and
// resolves against the drive's root on every read, so the identity mapping below
// hands on an unvalidated path safely.
func (a *API) Stream(w http.ResponseWriter, r *http.Request) error {
	d, err := a.drive(r)
	if err != nil {
		return err
	}

	// stream.Handler addresses its item by the "id" path value. This route
	// addresses a path, and takes ?path= like its siblings do, so the value is
	// restated under the name the handler reads — which keeps the route's shape
	// consistent with the rest of the drive API rather than bending a generic
	// handler to suit one caller.
	r.SetPathValue("id", path(r))

	// Same reasoning as Thumbnail: this path resolves against the asking
	// account's drive, so a cache keyed on the URL alone could serve one
	// account's file to another. stream.Handler sends validators but no
	// Cache-Control, which leaves a browser free to cache on its own judgement.
	w.Header().Set("Cache-Control", "private")
	w.Header().Set("Vary", "Cookie")

	// stream.Handler writes its own statuses, including failures, because it
	// decides them against what it has already committed to the response.
	a.streamer(d).ServeHTTP(w, r)

	return nil
}

// streamer is the handler for one drive, built once and kept.
//
// Keeping it is the point rather than an economy: stream.Handler reads object
// metadata on every request — to build Content-Range, to answer HEAD, and to
// send validators — and a cache built per request would be discarded before it
// ever answered twice. Against HiDrive that metadata call measured ~160 ms on
// every ranged read, a third of the cost of a seek.
//
// Keyed by the whole drive configuration, not just the alias. The cache inside
// is keyed by the caller's path, *before* the root is applied, so two accounts
// sharing an alias under different roots would otherwise be served each other's
// metadata.
func (a *API) streamer(d *hi.Drive) http.Handler {
	key := streamKey(d)

	a.mu.Lock()
	defer a.mu.Unlock()

	if h, ok := a.streams[key]; ok {
		return h
	}

	h := &stream.Handler{
		Store: blob.NewStatCache(d.Blobs(), 0),
		Key:   drivePath,
		Log:   a.log,
	}
	if a.streams == nil {
		a.streams = map[string]http.Handler{}
	}
	a.streams[key] = h

	return h
}

// streamKey identifies the drive a handler was built for. The whole
// configuration, because Home decides what an empty path and "~" resolve to,
// and Root decides everything else.
func streamKey(d *hi.Drive) string {
	cfg := d.Config()

	return cfg.Alias + "\x00" + cfg.Root + "\x00" + cfg.Home
}

// drivePath is the identity mapping from public id to storage key: this route
// addresses storage directly, so there is no catalogue to look anything up in.
//
// An empty path is reported as a missing item rather than left to the store,
// which would refuse it as a malformed key — the same 404 either way, but one
// of them logs an accusation.
func drivePath(_ context.Context, p string) (string, error) {
	if p == "" {
		return "", stream.ErrNoSuchItem
	}

	return p, nil
}
