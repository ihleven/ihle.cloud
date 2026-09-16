package hidrive

import (
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/ihleven/ihlvn/pkg/blob"
	"github.com/ihleven/ihlvn/pkg/hi"
	"github.com/ihleven/ihlvn/pkg/stream"
	"github.com/interhome-group/cms/pkg/errs"
)

// Library serves one fixed drive: the same tree for everybody.
//
// This is the opposite arrangement to API. There, a path is resolved against
// the account's own drive, so the same URL names different files for different
// people — which is why those responses cannot be cached by URL alone. Here
// there is no account in the resolution at all. One alias, one root, decided at
// startup.
//
// Three things follow from that, and they are the reason the distinction is a
// type rather than a flag:
//
//   - The entitlement is the only thing in front of the bytes. Nothing about
//     the request can widen what is reachable, because nothing about the
//     request decides where it looks.
//   - Responses are cacheable by URL. A cached image cannot be shown to the
//     wrong person, because there is no wrong person — everyone entitled sees
//     the same shelf.
//   - The root is the containment. What is below it is the library; what is not
//     is unreachable, including whatever else the same alias can see.
type Library struct {
	drive  *hi.Drive
	stream http.Handler
	urls   *signedURLs
	tags   *tagCache
}

// NewLibrary builds the library over a drive fixed at startup.
//
// The metadata cache is shared across requests rather than built per request,
// which is the whole reason this is cheaper than a pre-signed URL per call: the
// handler reads an object's size and validators on every request, and against
// HiDrive that is a round-trip unless something remembers it.
func NewLibrary(drive *hi.Drive, log *slog.Logger) *Library {
	if log == nil {
		log = slog.Default()
	}

	return &Library{
		drive: drive,
		urls:  newSignedURLs(0),
		tags:  newTagCache(0),
		stream: &stream.Handler{
			Store: blob.NewStatCache(drive.Blobs(), 0),
			Key:   drivePath,
			Log:   log,
		},
	}
}

// errNoLibrary is what the routes answer where no library is configured.
//
// A nil receiver rather than an absent route: the endpoint stays visible and
// says why it cannot serve, which is how the browser's handlers report the same
// situation. Silently unregistering it would make a deployment that forgot the
// setting look like one that never had the feature.
func (l *Library) configured() error {
	if l == nil {
		return errs.New("no video library is configured", errs.HTTPStatus(http.StatusNotImplemented))
	}

	return nil
}

// Meta lists a directory in the library, or describes one file.
//
// Shares the browser's lookup, so a trailing slash means the same thing here.
func (l *Library) Meta(w http.ResponseWriter, r *http.Request) error {
	if err := l.configured(); err != nil {
		return err
	}

	meta, err := lookupMeta(r.Context(), l.drive, path(r))
	if err != nil {
		return err
	}

	// Private rather than public: the library is the same for everyone entitled
	// to it, but not everyone is. No Vary is needed — unlike the browser's
	// listing, this URL names the same directory whoever asks.
	w.Header().Set("Cache-Control", "private, max-age=60")

	return writeJSON(w, meta)
}

// Stream serves a file's bytes, reading them here rather than handing out a
// pre-signed URL — see (*API).Stream for the measurements behind that choice.
func (l *Library) Stream(w http.ResponseWriter, r *http.Request) error {
	if err := l.configured(); err != nil {
		return err
	}

	// Cacheable without a Vary, unlike the browser's equivalent: the library is
	// one fixed shelf, so this URL names the same file for everyone entitled to
	// it. The handler adds validators of its own, so a stale copy revalidates
	// rather than being refetched.
	w.Header().Set("Cache-Control", "private, max-age=3600")

	// stream.Handler addresses its item by the "id" path value; this route
	// addresses a path, so the value is restated under the name it reads.
	r.SetPathValue("id", path(r))

	// The handler writes its own statuses, including failures, because it
	// decides them against what it has already committed to the response.
	l.stream.ServeHTTP(w, r)

	return nil
}

// Thumbnail serves a small rendering of a file — the cover of a magazine, the
// first page of a document — made by the store rather than here.
//
// The same picture for everybody, so unlike the browser's thumbnails this one
// carries no Vary: the URL names one file, because the root and not the asking
// account decides what it resolves to.
func (l *Library) Thumbnail(w http.ResponseWriter, r *http.Request) error {
	if err := l.configured(); err != nil {
		return err
	}

	resp, err := l.drive.Thumbnail(r.Context(), path(r), thumbSize(r.URL.Query()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	head := w.Header()
	head.Set("Content-Type", resp.Header.Get("Content-Type"))
	head.Set("X-Content-Type-Options", "nosniff")
	if n := resp.Header.Get("Content-Length"); n != "" {
		head.Set("Content-Length", n)
	}
	if lm := resp.Header.Get("Last-Modified"); lm != "" {
		head.Set("Last-Modified", lm)
	}
	head.Set("Cache-Control", "private, max-age=3600")

	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)

	return nil
}

// Media serves a file by handing the request to the store at a signed URL,
// rather than reading the bytes here as Stream does.
//
// The difference that matters is multi-range. A reader opening a large document
// asks for several pieces of it at once, and only the store can answer that:
// Stream serves one range per request and answers anything else with the whole
// file, which for a ninety-megabyte scan is the difference between showing a
// page and downloading a magazine. Video players ask one range at a time, so
// they keep the cheaper route.
func (l *Library) Media(w http.ResponseWriter, r *http.Request) error {
	if err := l.configured(); err != nil {
		return err
	}

	// Cacheable without a Vary: the library is one fixed shelf, so this URL
	// names the same file for everyone entitled to it.
	w.Header().Set("Cache-Control", "private, max-age=3600")

	return serveSigned(w, r, l.drive, l.urls, path(r))
}

// Search lists what is below a path in the library, at any depth, in one call.
//
// The one thing the library offers that a listing cannot: Meta reaches exactly
// one level, so a shelf arranged in folders costs a request per folder, while
// this costs one request whatever its shape. It is not simply the better of the
// two — a search takes seconds where a listing takes tens of milliseconds, and
// the cost does not follow the size of the answer — so which one a page should
// use depends on how deep the shelf is and how many folders that means. Both
// are served so the two can be compared against a real shelf rather than
// guessed at.
//
// What selects the hits is the caller's: "category=dir" for every folder below
// the path, needing no pattern, or "pattern=" for a substring of a filename.
// See (*hi.Client).Search for why files have no category of their own.
func (l *Library) Search(w http.ResponseWriter, r *http.Request) error {
	if err := l.configured(); err != nil {
		return err
	}

	query := r.URL.Query()
	hits, err := l.drive.Search(r.Context(), path(r), query.Get("pattern"), query.Get("category"))
	if err != nil {
		return err
	}

	// A hit names itself by its path and nothing else, and names it absolutely
	// — so unlike a listing, whose members are relative to the directory asked
	// for, these arrive spelled against the drive's own layout with the root in
	// front. Put back below the root they are addressable by the same routes
	// that served them; left alone they would name a file the library cannot
	// reach, and say where the shelf sits besides.
	//
	// Decoding is part of the rewrite: paths arrive percent-encoded, and a
	// caller that encodes a path to put it in a URL would otherwise encode it
	// twice. Names are left exactly as the provider sent them, because every
	// reader of a listing already decodes those.
	for i := range hits {
		plain := hits[i].Path
		if decoded, err := url.PathUnescape(plain); err == nil {
			plain = decoded
		}
		hits[i].Path = l.drive.Below(plain)
	}

	// Private and briefly cached, on the same reasoning as Meta: one shelf, the
	// same for everyone entitled to it, so no Vary is needed.
	w.Header().Set("Cache-Control", "private, max-age=60")

	return writeJSON(w, struct {
		Result []hi.Meta `json:"result"`
	}{hits})
}
