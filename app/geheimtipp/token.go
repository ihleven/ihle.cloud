package geheimtipp

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// The pool's own credential, issued from this app's session.
//
// The pool authenticates by a signed token and nothing else — it verifies the
// signature and reads a name out of it, and asks no service anything. That is
// what lets this app sign someone in and then speak to the pool on their behalf
// without the pool being involved, and without the browser ever holding the
// pool's credential.

// tokenName is the cookie the pool reads its token from. It looks nowhere else
// — an Authorization header is ignored — so the name is part of the upstream's
// contract rather than a choice.
const tokenName = "token"

// mintTTL is how long a minted token stays valid.
//
// It is short because the token lives for exactly one proxied request: it is
// created as the request is forwarded and is gone when the response comes back.
// The upstream's own sign-in issues tokens lasting weeks, which it has to,
// because there they are the session. Here they are not.
const mintTTL = 5 * time.Minute

// Minter issues the pool's tokens.
type Minter struct {
	secret []byte
}

// NewMinter returns nil when no secret is configured.
//
// A missing secret is a deployment without the pool, not a broken one: the
// proxy then forwards anonymously and the pool's public pages still work, which
// is better than refusing to start over an area the deployment may not use.
func NewMinter(secret string) *Minter {
	if secret == "" {
		return nil
	}
	return &Minter{secret: []byte(secret)}
}

// Token mints a token naming login.
//
// The claim is "Username", capitalised. That is not a style choice: the pool
// reads the name out of a Go struct field with no json tag, so the wire name is
// the field name, and a lowercase "username" would be read as absent. The token
// would verify and identify nobody — the pool would answer as if signed out,
// with no error anywhere to explain it.
func (m *Minter) Token(login string, now time.Time) (string, error) {
	if m == nil {
		return "", fmt.Errorf("geheimtipp: no signing secret configured")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"Username": login,
		"exp":      now.Add(mintTTL).Unix(),
	})
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("geheimtipp: signing a token for %s: %w", login, err)
	}
	return signed, nil
}

// dropToken removes any pool token the browser sent.
//
// It runs on every forwarded request, whether or not one is minted to replace
// it. The pool's own sign-in used to leave a token in the browser lasting weeks,
// and forwarding it would be a second way into the pool — one that needs no
// account here, survives signing out, and cannot be revoked. Whoever still has
// one signs in again, which is the point: there is one credential and this app
// issues it.
func dropToken(r *http.Request) {
	kept := make([]*http.Cookie, 0, len(r.Cookies()))
	for _, c := range r.Cookies() {
		if c.Name != tokenName {
			kept = append(kept, c)
		}
	}
	r.Header.Del("Cookie")
	for _, c := range kept {
		r.AddCookie(c)
	}
}

// authorize puts a freshly minted token on an outbound request. The caller has
// already dropped whatever the browser sent.
func (m *Minter) authorize(r *http.Request, login string) error {
	token, err := m.Token(login, time.Now())
	if err != nil {
		return err
	}
	r.AddCookie(&http.Cookie{Name: tokenName, Value: token})
	return nil
}
