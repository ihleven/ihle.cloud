package hidrive

import (
	"log/slog"
	"net/http"

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
