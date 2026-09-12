package films

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ihleven/ihlvn/pkg/blob"
	"github.com/ihleven/ihlvn/pkg/stream"
	"github.com/interhome-group/cms/mgmt"
	"github.com/interhome-group/cms/pkg/errs"
)

// Thumbnailer renders a preview of a stored object.
//
// Declared here rather than imported so delivery does not depend on HiDrive
// being the store: *hi.Drive satisfies it, and a test can pass a stub.
type Thumbnailer interface {
	Thumbnail(ctx context.Context, key string, params url.Values) (*http.Response, error)
}

// Poster serves the still for a film, addressed the same way its bytes are.
//
// A film that names no poster is previewed from its own footage — the store
// renders a frame — so a listing of silent reels looks like something without
// anyone having had to choose a still for each one.
func Poster(mngr *mgmt.Mngr, thumbs Thumbnailer) func(http.ResponseWriter, *http.Request) error {
	poster := EntryKey(mngr, posterKey)
	film := EntryKey(mngr, mediaKey)

	return func(w http.ResponseWriter, r *http.Request) error {
		if thumbs == nil {
			return notConfigured(w, r)
		}

		id := r.PathValue("id")

		key, err := poster(r.Context(), id)
		if errors.Is(err, stream.ErrNoSuchItem) {
			// No still of its own: fall back to the film, which the store can
			// render a frame from.
			key, err = film(r.Context(), id)
		}
		if err != nil {
			if errors.Is(err, stream.ErrNoSuchItem) {
				return errs.New("no such film", errs.HTTPStatus(http.StatusNotFound))
			}
			return err
		}

		// The key comes from content, so it is untrusted the same way it is on
		// the byte path.
		safe, err := blob.SafeKey(key)
		if err != nil {
			return errs.New("no such film", errs.HTTPStatus(http.StatusNotFound))
		}

		resp, err := thumbs.Thumbnail(r.Context(), safe, posterSize(r.URL.Query()))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// Private because the entry's ACL decided this response; a shared cache
		// must not hand it to someone the check would have refused.
		w.Header().Set("Cache-Control", "private, max-age=3600")
		w.WriteHeader(http.StatusOK)

		_, _ = io.Copy(w, resp.Body)
		return nil
	}
}

// notConfigured answers when no drive is configured. Reported per request so a
// deployment without one still shows the route exists.
func notConfigured(http.ResponseWriter, *http.Request) error {
	return errs.New("film delivery is not configured", errs.HTTPStatus(http.StatusNotImplemented))
}

// Poster sizes. The ceiling is not a policy about taste: without one, a query
// parameter decides how much work the store does, and a listing page is the
// easiest place in the app to ask for a thousand of them.
const (
	defaultPosterWidth  = 320
	defaultPosterHeight = 180
	maxPosterSide       = 1024
)

func posterSize(query url.Values) url.Values {
	return url.Values{
		"width":  {strconv.Itoa(side(query.Get("width"), defaultPosterWidth))},
		"height": {strconv.Itoa(side(query.Get("height"), defaultPosterHeight))},
	}
}

// side reads one dimension, falling back for anything unusable, so a malformed
// query renders a poster rather than an error.
func side(raw string, fallback int) int {
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return fallback
	}

	return min(n, maxPosterSide)
}
