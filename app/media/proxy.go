package media

import (
	"net/http"
	"net/http/httputil"

	"github.com/ihleven/ihlvn/pkg/authn"
	"github.com/ihleven/ihlvn/pkg/hi"
	"github.com/interhome-group/cms/pkg/errs"
)

// TokenSource yields the access token for a drive alias.
type TokenSource interface {
	AccessToken(alias string) (string, error)
}

// Proxy serves a file addressed by its path in the signed-in account's own
// drive, by asking the provider for a pre-signed URL and reverse-proxying the
// request to it.
//
// That is the whole distinction from Handler, and it is a distinction of
// addressing rather than of content: here the caller names a storage path and
// the drive is the viewer's, so it serves whatever that person can reach.
// Handler takes the published address of an entry, reads the storage key off
// it, and serves from one configured drive. This route is therefore the
// browsing path — someone looking at their own files — and Handler is the
// delivery path.
//
// It expects the middleware that requires a session, so the account is already
// in the context; arriving without one is a wiring mistake, not an anonymous
// request.
func Proxy(tokens TokenSource) func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		account, ok := authn.FromContext(r.Context())
		if !ok {
			return errs.New("not allowed", errs.HTTPStatus(http.StatusUnauthorized))
		}

		// Built per request from the account's own configuration, so the path is
		// resolved against that person's root and cannot leave it. Spelled out
		// rather than converted: the two structs agreeing today is not a reason
		// to make one package's field order load-bearing for another's.
		drive := hi.NewDrive(tokens, hi.DriveConfig{
			Alias: account.HiDrive.Alias,
			Root:  account.HiDrive.Root,
			Home:  account.HiDrive.Home,
		}, nil)

		path := r.PathValue("path")
		if p := r.URL.Query().Get("path"); p != "" {
			path = p
		}

		url, err := drive.URL(r.Context(), path)
		if err != nil {
			return err
		}

		proxy := httputil.NewSingleHostReverseProxy(url)
		r.URL.Host, r.URL.Scheme, r.Host = url.Host, url.Scheme, url.Host
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		proxy.ServeHTTP(w, r)

		return nil
	}
}
