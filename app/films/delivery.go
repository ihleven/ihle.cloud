package films

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/ihleven/ihlvn/pkg/authn"
	"github.com/ihleven/ihlvn/pkg/blob"
	"github.com/ihleven/ihlvn/pkg/stream"
	"github.com/interhome-group/cms/content"
	"github.com/interhome-group/cms/mgmt"
	"github.com/interhome-group/cms/mgmt/search"
	"github.com/interhome-group/cms/pkg/errs"
)

// KeyProvider is implemented by a content type whose entry refers to a stored
// object. It is the whole of what delivery needs to know about a type, so a new
// type joins by implementing one method rather than by being named in a list.
type KeyProvider interface {
	// MediaKey returns the storage key of the entry's bytes, or "" for none.
	MediaKey() string
}

// PosterProvider is implemented by a type that names its own still. A type
// without one falls back to a frame rendered from the object itself.
type PosterProvider interface {
	PosterKey() string
}

// Handler serves a film's bytes from store.
//
// A nil store means delivery is not configured, which is reported per request
// rather than at startup: the rest of the app runs perfectly well without a
// media drive, and a route that answers "not configured" is easier to diagnose
// than one that is silently absent.
func Handler(mngr *mgmt.Mngr, store blob.Blobstore, log *slog.Logger) func(http.ResponseWriter, *http.Request) error {
	if store == nil {
		return notConfigured
	}

	h := &stream.Handler{Store: store, Key: EntryKey(mngr, mediaKey), Log: log}

	// stream.Handler writes its own statuses, including failures, because it
	// has to decide them against what it has already committed to the response.
	return func(w http.ResponseWriter, r *http.Request) error {
		h.ServeHTTP(w, r)
		return nil
	}
}

// keyOf reads one storage key off an entry's content.
type keyOf func(*content.Entry) string

// mediaKey is the film itself.
func mediaKey(entry *content.Entry) string {
	provider, ok := entry.Content.(KeyProvider)
	if !ok {
		return ""
	}

	return provider.MediaKey()
}

// posterKey is the still, which may be absent even where the film is not.
func posterKey(entry *content.Entry) string {
	provider, ok := entry.Content.(PosterProvider)
	if !ok {
		return ""
	}

	return provider.PosterKey()
}

// EntryKey resolves a film's id to one of its storage keys.
//
// The address is the entry's id — a single path segment, language-independent,
// and the same one the archive uses for itself. It is deliberately not a
// storage path: a path would put the repository's layout, and the ".yaml" its
// metadata happens to be stored as, into a public URL. Addressing an asset by
// path is a separate route with separate rules.
func EntryKey(mngr *mgmt.Mngr, key keyOf) stream.KeyFunc {
	return func(ctx context.Context, id string) (string, error) {
		entry, err := lookup(mngr, ctx, id)

		return storageKey(entry, err, key)
	}
}

// lookup finds the entry an id names, as the caller it belongs to.
//
// Everything that is not a readable entry comes back as ErrNoSuchItem: whether
// one is missing or merely unreadable by this user is a fact about content the
// caller was not granted.
func lookup(mngr *mgmt.Mngr, ctx context.Context, id string) (*content.Entry, error) {
	path, ok := pathForID(mngr, strings.Trim(id, "/"))
	if !ok {
		return nil, stream.ErrNoSuchItem
	}

	// Delivery wants the entry's own fields, so references are left unresolved:
	// nothing here follows a link.
	entry, err := mngr.GetEntry(path, authn.ContextUser(ctx), content.RSLV_NONE)
	if err != nil {
		switch errs.StatusOf(err) {
		case http.StatusNotFound, http.StatusForbidden, http.StatusUnauthorized:
			return nil, stream.ErrNoSuchItem
		}
		return nil, err
	}
	if entry == nil {
		return nil, stream.ErrNoSuchItem
	}

	return entry, nil
}

// pathForID finds the entry an id names.
//
// An id that names no entry and one that names several are both "no": a
// duplicate id is a content fault, and serving whichever the index happened to
// return first would hide it behind a film that plays.
func pathForID(mngr *mgmt.Mngr, id string) (string, bool) {
	if id == "" || mngr.Engine == nil {
		return "", false
	}

	result, err := mngr.Engine.Search(search.Params{ID: []string{id}, PageSize: 2})
	if err != nil || result.Total != 1 || len(result.Hits) < 1 {
		return "", false
	}

	return result.Hits[0].ID, true
}

// storageKey reads the requested key off an entry.
//
// A type that has no such key is reported as a missing item, not as a different
// failure: that an entry exists but has no poster is a fact about content the
// caller was not granted, and delivery draws no distinction it does not have to.
func storageKey(entry *content.Entry, err error, key keyOf) (string, error) {
	if err != nil {
		return "", err
	}

	if k := key(entry); k != "" {
		return k, nil
	}

	return "", stream.ErrNoSuchItem
}

// PosterURL is where a film's still is served from. It is built here because
// the search index stores it, and an index outlives the page that reads it.
func PosterURL(id string) string {
	if id == "" {
		return ""
	}

	return "/api/v1/films/" + url.PathEscape(id) + "/poster.jpg"
}
