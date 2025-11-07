package super8

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"path"

	"bitbucket.org/hotelplan/webcc-content/cms/pkg/errors"
	"github.com/ihleven/ihle.cloud/pkg/auth"
	"github.com/ihleven/ihle.cloud/pkg/hi"
)

func ServeHiVideo(w http.ResponseWriter, r *http.Request) error {

	account, claims, err := auth.GetAccountUnused(r)
	if err != nil {
		return err
	}
	if claims.Audience == nil { // []string{"familie"} {
		return errors.NewWithCode(401, "not allowed")
	}

	access_token, err := auth.AuthenticatorPKG.GetToken(account)
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
