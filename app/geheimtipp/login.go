package geheimtipp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/interhome-group/cms/pkg/errs"
)

// userAgent identifies this app to the pool's host, which answers Go's default
// with a 403.
const userAgent = "ihlvn"

// TokenCookie is the name the pool's backend looks for. It reads the JWT from
// this cookie and from nowhere else — an Authorization header is ignored — so
// the name is part of the upstream's contract, not a choice.
const TokenCookie = "token"

// tokenTTL bounds how long a signed-in browser keeps the cookie. The JWT carries
// its own expiry and the backend enforces it; this only keeps a stale cookie
// from being offered long after it stopped working.
const tokenTTL = 30 * 24 * time.Hour

// Login signs in against the pool and turns its answer into a cookie.
//
// The pool's backend returns the JWT in the response body and sets no cookie —
// minting one was the job of the frontend's own server layer, which this app is
// replacing. So it happens here.
//
// The JWT is deliberately not passed on to the browser. Their frontend put it in
// a readable cookie and decoded it client-side to learn who was signed in; here
// GET /aktuell answers that, so the credential can stay HttpOnly and out of
// reach of any script on the page.
func Login(upstream string, secure bool, prefix string) (func(http.ResponseWriter, *http.Request) error, error) {
	target, err := url.Parse(strings.TrimSuffix(upstream, "/"))
	if err != nil || target.Scheme == "" || target.Host == "" {
		return nil, fmt.Errorf("geheimtipp: upstream %q needs a scheme and a host", upstream)
	}
	endpoint := target.JoinPath("login").String()

	return func(w http.ResponseWriter, r *http.Request) error {
		if err := r.ParseForm(); err != nil {
			return errs.Wrap(err, "could not read the sign-in form", errs.HTTPStatus(http.StatusBadRequest))
		}

		// The upstream takes a form, not JSON, and the field is `username`.
		// Both are easy to get wrong and both fail the same way — a 401 that
		// looks like a wrong password.
		form := url.Values{
			"username": {r.PostForm.Get("username")},
			"password": {r.PostForm.Get("password")},
		}

		request, err := http.NewRequestWithContext(r.Context(), http.MethodPost, endpoint,
			strings.NewReader(form.Encode()))
		if err != nil {
			return errs.Wrap(err, "building the sign-in request", errs.HTTPStatus(http.StatusInternalServerError))
		}
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		// Named, because the pool's host refuses Go's default User-Agent with a
		// 403 — which arrives here as "wrong login or password" and sends
		// whoever is signing in to check a password that was never the problem.
		request.Header.Set("User-Agent", userAgent)

		resp, err := http.DefaultClient.Do(request)
		if err != nil {
			return errs.Wrap(err, "the geheimtipp backend did not answer", errs.HTTPStatus(http.StatusBadGateway))
		}
		defer resp.Body.Close()

		// Only 401 is reported as a refusal of the credentials. Anything else is
		// the pool being unreachable or unhappy with the request — 403 is what
		// its host answers a request it will not serve at all — and saying
		// "wrong password" to that sends someone to fix the one thing that was
		// never broken.
		if resp.StatusCode == http.StatusUnauthorized {
			return errs.New("wrong login or password", errs.HTTPStatus(http.StatusUnauthorized))
		}
		if resp.StatusCode != http.StatusOK {
			return errs.New("the geheimtipp backend refused the sign-in (%d)", resp.StatusCode,
				errs.HTTPStatus(http.StatusBadGateway))
		}

		var answer struct {
			Username string
			JWT      string
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
		if err := json.Unmarshal(body, &answer); err != nil || answer.JWT == "" {
			return errs.New("the geheimtipp backend answered with no token", errs.HTTPStatus(http.StatusBadGateway))
		}

		setToken(w, answer.JWT, secure, prefix)

		// Only the name goes back. The browser has no use for the token.
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(map[string]string{"username": answer.Username})
	}, nil
}

// Logout drops the cookie. The pool's backend has no session to end — the JWT is
// self-contained — so forgetting it here is the whole of signing out.
func Logout(secure bool, prefix string) func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		setToken(w, "", secure, prefix)
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}

// setToken writes the cookie, or clears it when the value is empty.
//
// Path is the proxy's mount point, so the pool's credential is sent to the pool
// and to nothing else this app serves. The two sites share an origin but not an
// identity, and the cookie is where that would blur first.
func setToken(w http.ResponseWriter, jwt string, secure bool, prefix string) {
	cookie := &http.Cookie{
		Name:     TokenCookie,
		Value:    jwt,
		Path:     prefix,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
	if jwt == "" {
		cookie.MaxAge = -1
	} else {
		cookie.Expires = time.Now().Add(tokenTTL)
	}
	http.SetCookie(w, cookie)
}
