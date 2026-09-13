// Browsing the family's storage: what a person can see of a HiDrive account,
// as opposed to how the app is allowed to read one at all.
//
// Two questions, deliberately answered by two different things. *Whether*
// someone may browse at all is the hidrive entitlement, checked by the
// middleware in front of these handlers. *Whose* tree they land in is the
// account's HiDrive alias, or the deployment's when the account names none.
//
// Keeping them apart is what stops an alias configured for some other purpose —
// delivering films, say — from quietly handing out a file browser, and what
// stops an entitled account with no alias from landing on an empty page. Either
// rule on its own gets one of those wrong.
//
// Nothing here decides what a path may reach: hi.Drive resolves every path
// against the configured root, and cleaning an absolute path resolves ".."
// against "/" before the join, so a caller cannot address anything above it.
package hidrive

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	"github.com/ihleven/ihlvn/app/auth"
	"github.com/ihleven/ihlvn/pkg/hi"
	"github.com/interhome-group/cms/pkg/errs"
)

// Shared is the drive an entitled account falls back to when it names none of
// its own — the deployment's alias and the root it resolves against.
type Shared struct {
	Alias string
	Root  string
}

// API serves the browser: a listing, a thumbnail, and the bytes.
type API struct {
	tokens AccessTokens
	shared Shared
}

func NewAPI(tokens AccessTokens, shared Shared) *API {
	return &API{tokens: tokens, shared: shared}
}

// drive builds the drive this request browses.
//
// The account's own alias wins; the shared one stands in. Neither configured is
// a deployment that cannot answer, reported per request so the route still
// visibly exists.
func (a *API) drive(r *http.Request) (*hi.Drive, error) {
	account, ok := auth.FromContext(r.Context())
	if !ok {
		// The middleware in front of this requires an account. Arriving without
		// one is a wiring mistake, not an anonymous request.
		return nil, errs.New("not allowed", errs.HTTPStatus(http.StatusUnauthorized))
	}

	cfg := hi.DriveConfig{
		Alias: account.HiDrive.Alias,
		Root:  account.HiDrive.Root,
		Home:  account.HiDrive.Home,
	}
	if cfg.Alias == "" {
		cfg = hi.DriveConfig{Alias: a.shared.Alias, Root: a.shared.Root}
	}
	if cfg.Alias == "" {
		return nil, errs.New("no storage is configured", errs.HTTPStatus(http.StatusNotImplemented))
	}

	return hi.NewDrive(a.tokens, cfg, nil), nil
}

// path is what the caller asked for, from the route or from a query parameter.
func path(r *http.Request) string {
	if p := r.URL.Query().Get("path"); p != "" {
		return p
	}
	return r.PathValue("path")
}

// Meta answers with the entry at a path: a file, or a directory and its members.
//
// Two calls exist upstream and they are not interchangeable. /dir returns the
// members with everything a listing shows — category, size, dimensions — while
// /meta asks only for their id and name, so a directory fetched through /meta
// lists nothing useful. /dir in turn refuses a path that is not a directory.
//
// A browser mostly asks about directories, so /dir goes first and /meta stands
// in when it refuses. That costs a file two round trips and a directory one,
// which is the right way round. It also means a path that does not exist is
// reported by /meta, whose error says more than /dir's does.
func (a *API) Meta(w http.ResponseWriter, r *http.Request) error {
	d, err := a.drive(r)
	if err != nil {
		return err
	}

	p := path(r)

	meta, err := d.Dir(r.Context(), p)
	if err != nil {
		if meta, err = d.Meta(r.Context(), p); err != nil {
			return err
		}
	}

	return writeJSON(w, meta)
}

// Thumbnail serves a preview of an image, rendered by the store.
//
// The size is clamped rather than passed through. Without a ceiling a query
// parameter decides how much work the store does, and a directory listing is
// the easiest place in this app to ask for a thousand of them; a fixed set of
// widths is also what lets the browser cache hit instead of fragmenting across
// arbitrary ones.
func (a *API) Thumbnail(w http.ResponseWriter, r *http.Request) error {
	d, err := a.drive(r)
	if err != nil {
		return err
	}

	resp, err := d.Thumbnail(r.Context(), path(r), thumbSize(r.URL.Query()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// Private because an entitlement and a root decided this response; a shared
	// cache must not hand it to someone those checks would have refused.
	w.Header().Set("Cache-Control", "private, max-age=3600")
	w.WriteHeader(http.StatusOK)

	_, _ = io.Copy(w, resp.Body)
	return nil
}

// thumbWidths are the sizes the browser may ask for: a row in a listing, a
// gallery tile, and a preview. Anything else is rounded up to the next one.
var thumbWidths = []int{100, 200, 400, 800}

func thumbSize(query url.Values) url.Values {
	want, _ := strconv.Atoi(query.Get("width"))
	width := thumbWidths[len(thumbWidths)-1]
	for _, w := range thumbWidths {
		if want <= w {
			width = w
			break
		}
	}
	return url.Values{"width": {strconv.Itoa(width)}}
}

// Media serves the bytes of a file, by asking the store for a pre-signed URL and
// proxying the request to it, so the bytes never pass through this process.
func (a *API) Media(w http.ResponseWriter, r *http.Request) error {
	d, err := a.drive(r)
	if err != nil {
		return err
	}

	url, err := d.URL(r.Context(), path(r))
	if err != nil {
		return err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
	r.URL.Host, r.URL.Scheme, r.Host = url.Host, url.Scheme, url.Host
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	proxy.ServeHTTP(w, r)

	return nil
}

func writeJSON(w http.ResponseWriter, value any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(value)

}
