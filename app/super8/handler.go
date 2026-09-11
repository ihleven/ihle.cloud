package super8

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"path"

	"github.com/ihleven/ihlvn/pkg/authn"
	"github.com/ihleven/ihlvn/pkg/hi"
	"github.com/interhome-group/cms/pkg/errs"
)

// TokenSource yields the HiDrive access token for an account's alias.
type TokenSource interface {
	AccessToken(alias string) (string, error)
}

// ServeHiVideo proxies a video out of HiDrive for the signed-in account.
//
// It expects to be wrapped in the middleware that requires a session, so the
// account is already in the request context; reaching it without one is a
// wiring mistake rather than an anonymous request.
func ServeHiVideo(tokens TokenSource) func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {

		account, ok := authn.FromContext(r.Context())
		if !ok {
			return errs.New("not allowed", errs.HTTPStatus(401))
		}

		access_token, err := tokens.AccessToken(account.HiDrive.Alias)
		if err != nil {
			return err
		}

		// prefix, accesstoken, _ := auth.GetPrefixAndToken(req.Request)
		// accesstoken := hicookie(r)
		// prefix := "/"

		pth := r.PathValue("path")
		if p := r.URL.Query().Get("path"); p != "" {
			pth = p
		}
		pth = path.Clean("/" + pth)
		fmt.Println("path", pth)
		url, err := hi.NewClient(access_token, "/").GetURL(pth)
		if err != nil {
			return err
		}
		fmt.Println("url =>", url.String())

		proxy := httputil.NewSingleHostReverseProxy(url)
		r.URL.Host = url.Host
		r.URL.Scheme = url.Scheme
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		r.Host = url.Host
		// w.Header().Set("access-control-allow-origin", "*")
		proxy.ServeHTTP(w, r)

		return nil
	}
}
