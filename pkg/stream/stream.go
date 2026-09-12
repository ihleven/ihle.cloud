package stream

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ihleven/ihlvn/pkg/blob"
)

// KeyFunc maps a public media id to a storage key. It keeps this package
// independent of the catalogue: streaming only needs to know where the bytes
// are, not what the metadata says about them.
type KeyFunc func(ctx context.Context, id string) (string, error)

// ErrNoSuchItem reports an unknown media id.
var ErrNoSuchItem = errors.New("stream: no such item")

// Handler proxies media bytes from a Blobstore to the client, translating the
// client's range and conditional requests into store reads.
//
// Bytes are proxied rather than redirected on purpose: the store credential
// stays on the server, and access stays gated by this app's session (from
// Phase 3). A redirect or a pre-signed URL would hand the storage backend
// straight to the browser.
type Handler struct {
	Store blob.Blobstore
	Key   KeyFunc
	Log   *slog.Logger
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	id := r.PathValue("id")

	key, err := h.Key(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNoSuchItem) {
			http.NotFound(w, r)
			return
		}
		h.fail(w, r, "resolving id", id, err)
		return
	}

	obj, err := h.Store.Stat(ctx, key)
	switch {
	case errors.Is(err, blob.ErrNotFound):
		http.NotFound(w, r)
		return
	case errors.Is(err, blob.ErrInvalidKey):
		// The key came from content, so this is a data problem, not a client
		// one: report 404 outward and make it visible in the log.
		h.Log.Error("unsafe media key rejected", "id", id, "key", key)
		http.NotFound(w, r)
		return
	case err != nil:
		h.fail(w, r, "stat", key, err)
		return
	}

	// Validators go out before the conditional check so a 304 carries them too.
	head := w.Header()
	head.Set("Accept-Ranges", "bytes")
	if obj.ETag != "" {
		head.Set("ETag", obj.ETag)
	}
	if !obj.ModTime.IsZero() {
		head.Set("Last-Modified", obj.ModTime.UTC().Format(http.TimeFormat))
	}

	if notModified(r, obj) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// Never let a content-supplied MIME type decide how the browser treats the
	// body; the store's own type wins, and sniffing stays off.
	head.Set("Content-Type", obj.ContentType)
	head.Set("X-Content-Type-Options", "nosniff")

	rng, ranged, err := ParseRange(r.Header.Get("Range"), obj.Size)
	if errors.Is(err, ErrUnsatisfiable) {
		head.Set("Content-Range", fmt.Sprintf("bytes */%d", obj.Size))
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		return
	}

	off, length, status := int64(0), obj.Size, http.StatusOK
	if ranged {
		off, length, status = rng.Off, rng.Len, http.StatusPartialContent
		head.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", rng.Off, rng.Last(), obj.Size))
	}
	head.Set("Content-Length", strconv.FormatInt(length, 10))

	if r.Method == http.MethodHead {
		w.WriteHeader(status)
		return
	}

	// Open before writing the status line: once the header is committed, a
	// store error can no longer be reported as one.
	body, err := h.Store.OpenRange(ctx, key, off, length)
	if err != nil {
		if errors.Is(err, blob.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		h.fail(w, r, "open", key, err)
		return
	}
	defer body.Close()

	w.WriteHeader(status)
	if _, err := io.Copy(w, body); err != nil {
		// The client seeking or closing the tab aborts the copy; that is normal
		// for media and not worth an error line.
		h.Log.Debug("media copy ended early", "id", id, "key", key, "err", err)
	}
}

// StatusClientClosedRequest is nginx's non-standard 499. Nothing can be
// delivered to a client that has already gone, but recording it distinctly
// keeps abandoned requests out of the server-error count.
const StatusClientClosedRequest = 499

// fail reports a request that could not be served.
//
// A player abandons range requests as a matter of course: it opens one, then
// cancels it the moment the user seeks again or it has buffered enough. That
// surfaces as a cancelled context, which is normal traffic rather than a
// server fault — so it must not become a 500 or an error-level log line, or
// ordinary seeking fills the log with alarms.
func (h *Handler) fail(w http.ResponseWriter, r *http.Request, what, subject string, err error) {
	if clientGone(r, err) {
		h.Log.Debug("client went away", "op", what, "subject", subject)
		w.WriteHeader(StatusClientClosedRequest)
		return
	}
	h.Log.Error("media stream failed", "op", what, "subject", subject, "err", err)
	http.Error(w, "internal error", http.StatusInternalServerError)
}

// clientGone reports whether an error is the consequence of the client
// disconnecting. The request context is checked as well as the error, so a
// cancellation reported by a store wrapper is recognised too.
func clientGone(r *http.Request, err error) bool {
	return errors.Is(err, context.Canceled) || r.Context().Err() != nil
}

// notModified implements the conditional-request rules a media player relies on
// for cheap re-watches: If-None-Match wins over If-Modified-Since, per RFC 9110.
func notModified(r *http.Request, obj blob.Object) bool {
	if inm := r.Header.Get("If-None-Match"); inm != "" {
		return etagMatches(inm, obj.ETag)
	}
	ims := r.Header.Get("If-Modified-Since")
	if ims == "" || obj.ModTime.IsZero() {
		return false
	}
	since, err := http.ParseTime(ims)
	if err != nil {
		return false
	}
	// HTTP dates have second resolution, so compare truncated.
	return !obj.ModTime.Truncate(time.Second).After(since)
}

func etagMatches(header, etag string) bool {
	if etag == "" {
		return false
	}
	for _, candidate := range strings.Split(header, ",") {
		candidate = strings.TrimSpace(candidate)
		if candidate == "*" {
			return true
		}
		// A weak validator still identifies the same bytes for our purposes.
		if strings.TrimPrefix(candidate, "W/") == strings.TrimPrefix(etag, "W/") {
			return true
		}
	}
	return false
}
