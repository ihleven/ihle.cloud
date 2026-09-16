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
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync"

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
	// tokens is held as the interface rather than the concrete AccessTokens so
	// a test can stand in for it. Without that seam nothing in this file can be
	// exercised past the first upstream call.
	tokens hi.TokenSource
	shared Shared
	log    *slog.Logger

	// urls remembers pre-signed URLs, so a viewer making many ranged requests
	// pays to mint one once rather than per request.
	urls *signedURLs

	// client is the API this talks to. Nil is HiDrive itself; a test points it
	// at a stub, which is the only way to reach the response-writing half of
	// these handlers without a network.
	client *hi.Client

	// Stream handlers, one per drive configuration, built on first use. See
	// streamer for why they outlive the request.
	mu      sync.Mutex
	streams map[string]http.Handler
}

func NewAPI(tokens hi.TokenSource, shared Shared) *API {
	return &API{tokens: tokens, shared: shared, log: slog.Default(), urls: newSignedURLs(0)}
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

	return hi.NewDrive(a.tokens, cfg, a.client), nil
}

// path is what the caller asked for, from the route or from a query parameter,
// as a path relative to the drive's root.
//
// A leading slash is stripped rather than refused. These routes take a path out
// of a URL, where "/x" and "x" plainly mean the same directory, and every one of
// them resolves against the root afterwards — so tolerating it decides nothing
// about what is reachable.
//
// It has to happen here, once, because the two ways on from this point did not
// agree. A listing goes through hi.Drive.Resolve, which strips the slash itself
// and reads the same file either way; the bytes go through blob.SafeKey, which
// refuses a leading slash outright and logs it as an unsafe key. Both are right
// about their own input — SafeKey guards keys written into CMS content, where a
// malformed one is a bug worth surfacing, and says so — but a request spelled
// with a slash is not that, and it should not get a 404 from one route and a
// listing from the other.
func path(r *http.Request) string {
	asked := r.URL.Query().Get("path")
	if asked == "" {
		asked = r.PathValue("path")
	}

	return strings.TrimLeft(asked, "/")
}

// lookupMeta answers for a path, taking the trailing slash as a hint.
//
// Two calls exist upstream and they are not interchangeable. /dir returns the
// members with everything a listing shows — category, size, dimensions — while
// /meta asks only for their id and name, so a directory fetched through /meta
// reports type="dir" and nmembers, and lists nothing. /dir in turn refuses a
// path that is not a directory.
//
// The caller usually knows which it is asking about, and says so the way a URL
// always has: a trailing slash. Following that hint costs one call either way,
// where always trying /dir first costs a file two.
//
// The hint is not trusted, only believed. A directory addressed without its
// slash still gets a full listing, because /meta having said "dir" is enough to
// ask again — and that matters, since the failure would otherwise be silent: a
// listing with names and nothing else, rather than an error.
func lookupMeta(ctx context.Context, d *hi.Drive, p string) (*hi.Meta, error) {
	if p == "" || strings.HasSuffix(p, "/") {
		if meta, err := d.Dir(ctx, p); err == nil {
			return meta, nil
		}

		// Not a directory after all, or gone. /meta's error says more than
		// /dir's, so let it be the one reported.
		return d.Meta(ctx, p)
	}

	meta, err := d.Meta(ctx, p)
	if err != nil {
		return nil, err
	}
	if meta.Type_ == "dir" {
		if full, err := d.Dir(ctx, p); err == nil {
			return full, nil
		}
	}

	return meta, nil
}

// Meta answers with the entry at a path: a file, or a directory and its members.
func (a *API) Meta(w http.ResponseWriter, r *http.Request) error {
	d, err := a.drive(r)
	if err != nil {
		return err
	}

	meta, err := lookupMeta(r.Context(), d, path(r))
	if err != nil {
		return err
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

	head := w.Header()
	head.Set("Content-Type", resp.Header.Get("Content-Type"))
	head.Set("X-Content-Type-Options", "nosniff")
	if n := resp.Header.Get("Content-Length"); n != "" {
		head.Set("Content-Length", n)
	}
	// The store sends Last-Modified but no ETag (measured). Passed through
	// because it costs nothing and lets a browser judge freshness for itself.
	//
	// Deliberately *not* forwarding the caller's conditional headers upstream to
	// turn that into a 304: the store takes none, so it would mean threading
	// them through Client and Drive, and a thumbnail measures 4.3 KB. The
	// plumbing would cost more than every 304 it could ever save.
	if lm := resp.Header.Get("Last-Modified"); lm != "" {
		head.Set("Last-Modified", lm)
	}

	// Private because an entitlement and a root decided this response; a shared
	// cache must not hand it to someone those checks would have refused.
	//
	// Vary because this URL does not identify the image. The path is resolved
	// against the asking account's drive, so the same URL names different files
	// for different people, and a browser cache keyed on the URL alone would
	// show one account's picture to the next person signed in on that machine.
	// The session cookie is issued at sign-in and never rotated, so keying on it
	// costs nothing within a session and misses exactly when it should.
	head.Set("Cache-Control", "private, max-age=3600")
	head.Set("Vary", "Cookie")

	// Overriding the store's own Cache-Control, which is max-age=5: a thumbnail
	// of a file that has not changed does not go stale in five seconds, and the
	// validator above covers the case where it does.
	//
	// The store's status rather than an assumed 200. Nothing arrives here with
	// anything else today — conditional headers are not forwarded upstream, so
	// the store never answers 304, and anything from 400 up already became an
	// error. That is true by accident rather than by design, and writing 200
	// regardless would turn the day either changes into an empty body the
	// browser caches as the image.
	w.WriteHeader(resp.StatusCode)

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

// Media serves the bytes of a file by asking the store for a pre-signed URL and
// proxying the request to it, which leaves the protocol — ranges, validators,
// 206 and 416 — to whatever answers that URL.
//
// The bytes do pass through this process. A reverse proxy makes the outbound
// request and copies the response back; only a redirect would keep them out.
// An earlier version of this comment claimed the opposite, which is worth
// correcting out loud because it is the sort of thing that gets reasoned from.
//
// It costs a /file/url round-trip on every request — measured at ~104 ms to
// first byte against ~44 ms for Stream, which pays for metadata once a minute
// instead. Fetching a file once, that does not show; seeking through one, it
// does, which is what Stream is for.
func (a *API) Media(w http.ResponseWriter, r *http.Request) error {
	d, err := a.drive(r)
	if err != nil {
		return err
	}

	asked := path(r)
	key := d.Resolve(asked)

	mint := func() (*url.URL, error) {
		signed, err := d.URL(r.Context(), asked)
		if err != nil {
			return nil, err
		}
		a.urls.put(key, signed)

		return signed, nil
	}

	signed, cached := a.urls.get(key)
	if !cached {
		if signed, err = mint(); err != nil {
			return err
		}
	}

	if !a.proxy(w, r, signed) {
		return nil
	}

	// The URL was reused and the store has stopped honouring it. Nothing has
	// been written yet — that is what makes this recoverable — so mint a fresh
	// one and go again, once.
	a.urls.drop(key)
	if !cached {
		// Freshly minted and already refused: retrying would ask the same
		// question twice.
		return errs.New("the store refused a newly signed URL", errs.HTTPStatus(http.StatusBadGateway))
	}
	if signed, err = mint(); err != nil {
		return err
	}
	if a.proxy(w, r, signed) {
		return errs.New("the store refused a newly signed URL", errs.HTTPStatus(http.StatusBadGateway))
	}

	return nil
}

// staleStatus is the store disowning a URL, as opposed to answering with it.
// A 404 counts: an expired signature and a deleted file are reported the same
// way, and re-minting settles which it was.
func staleStatus(code int) bool {
	switch code {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusGone:
		return true
	}

	return false
}

// proxy forwards the request to a signed URL and reports whether the store
// disowned it, in which case nothing has been written to w.
//
// The response is inspected before any of it is committed: ModifyResponse
// returning an error means ReverseProxy hands off to ErrorHandler instead of
// writing, which is the hook that makes a retry possible at all. Once a status
// line is out, a stale URL can only be reported as a broken download.
//
// The request is cloned per attempt because the proxy rewrites what it is
// given, and a second attempt has to start from the caller's original.
func (a *API) proxy(w http.ResponseWriter, r *http.Request, signed *url.URL) (stale bool) {
	attempt := r.Clone(r.Context())
	attempt.URL.Host, attempt.URL.Scheme, attempt.Host = signed.Host, signed.Scheme, signed.Host
	attempt.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))

	proxy := httputil.NewSingleHostReverseProxy(signed)
	proxy.ModifyResponse = func(resp *http.Response) error {
		if staleStatus(resp.StatusCode) {
			stale = true

			return errStaleURL
		}

		return nil
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, err error) {
		if errors.Is(err, errStaleURL) {
			// Caller's to retry; writing anything here would take that away.
			return
		}
		http.Error(w, "the store could not be reached", http.StatusBadGateway)
	}
	proxy.ServeHTTP(w, attempt)

	return stale
}

var errStaleURL = errors.New("the signed URL is no longer honoured")

func writeJSON(w http.ResponseWriter, value any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(value)

}
