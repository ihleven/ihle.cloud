package hidrive

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/ihleven/ihlvn/pkg/hiauth"
	"github.com/interhome-group/cms/pkg/errs"
)

// The consent flow, which is how a refresh token is obtained for an alias in the
// first place. The session system authenticates people; this authorises storage
// access, and the two have nothing to do with each other.
//
// It is an operator action, not something a visitor can start: the routes are
// mounted behind an account, and what comes back is stored for the whole app
// rather than for whoever happened to click.

// consentURL is where HiDrive asks the account holder to agree.
const consentURL = "https://my.hidrive.com/client/authorize"

// OAuth runs the consent flow against one registered client.
type OAuth struct {
	clientID     string
	clientSecret string

	// redirectURI has to match the registration held by HiDrive exactly, so
	// changing the domain means updating it there too. It is derived from the
	// app's public URL rather than from the request, which describes the local
	// hop when there is a proxy in front.
	redirectURI string

	tokens *Store
}

func NewOAuth(clientID, clientSecret, redirectURI string, tokens *Store) *OAuth {
	return &OAuth{
		clientID: clientID, clientSecret: clientSecret,
		redirectURI: redirectURI, tokens: tokens,
	}
}

// Authorize sends the operator to HiDrive to agree.
func (o *OAuth) Authorize(w http.ResponseWriter, r *http.Request) error {
	// state comes back untouched and is where the browser is sent afterwards,
	// so it is a path on this site and is validated as one.
	next := r.URL.Query().Get("state")
	if next == "" {
		next = "/"
	}
	if !isLocalPath(next) {
		return errs.New("state must be a path on this site", errs.HTTPStatus(http.StatusBadRequest))
	}

	params := url.Values{
		"client_id":     {o.clientID},
		"response_type": {"code"},
		"scope":         {"admin,rw"},
		"state":         {next},
		"redirect_uri":  {o.redirectURI},
	}
	http.Redirect(w, r, consentURL+"?"+params.Encode(), http.StatusSeeOther)
	return nil
}

// Callback exchanges the authorization code for a refresh token and stores it.
//
// Nothing is handed to the browser. An access token used to be set as a cookie
// here, which put a credential for the storage account into whichever browser
// happened to complete the flow; nothing ever read it. The token belongs to the
// app, and the app keeps it.
func (o *OAuth) Callback(w http.ResponseWriter, r *http.Request) error {
	token, err := hiauth.RefreshTokenWithAuthCode(
		o.clientID, o.clientSecret, r.URL.Query().Get("code"))
	if err != nil {
		return errs.Wrap(err, "exchanging the authorization code", errs.HTTPStatus(http.StatusUnauthorized))
	}
	if err := o.tokens.StoreToken(r.Context(), token); err != nil {
		return err
	}

	next := r.URL.Query().Get("state")
	if !isLocalPath(next) {
		next = "/"
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
	return nil
}

// isLocalPath keeps a redirect on this site: "//elsewhere" is a URL, not a path.
func isLocalPath(p string) bool {
	return strings.HasPrefix(p, "/") && !strings.HasPrefix(p, "//")
}

// AccessTokens answers what a handler actually asks for: given an alias, a token
// to use upstream. It wraps the refresher so callers do not have to know that a
// refresh token exists.
type AccessTokens struct {
	mngr *hiauth.TokenMngr
}

// NewAccessTokens brings up the refresher over the stored tokens.
func NewAccessTokens(tokens *Store) AccessTokens {
	return AccessTokens{mngr: hiauth.NewTokenMngr(tokens)}
}

func (a AccessTokens) AccessToken(alias string) (string, error) {
	refresher := a.mngr.GetTokenRefresher(alias)
	if refresher == nil {
		return "", errs.New("no HiDrive token stored for alias %q", alias)
	}
	return refresher.GetAccessToken()
}
