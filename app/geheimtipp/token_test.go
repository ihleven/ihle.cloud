package geheimtipp

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"
)

// as is an Identifier that always answers with one login, or never.
type as struct {
	login string
	may   bool
}

func (a as) PoolLogin(*http.Request) (string, bool) { return a.login, a.may }

// The pool reads the name out of a struct field with no json tag, so the wire
// name is the field name — capitalised. A lowercase key would verify and
// identify nobody: the pool would answer as if signed out, and nothing
// anywhere would say why. So the claim set is asserted literally rather than
// round-tripped through a decoder that would accept either spelling.
func TestTokenClaimsAreExactlyExpAndUsername(t *testing.T) {
	token, err := NewMinter("secret").Token("matt", time.Now())
	if err != nil {
		t.Fatal(err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("got %d token segments, want 3", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatal(err)
	}

	keys := make([]string, 0, len(claims))
	for k := range claims {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if got, want := strings.Join(keys, ","), "Username,exp"; got != want {
		t.Errorf("claims = %q, want %q", got, want)
	}
	if claims["Username"] != "matt" {
		t.Errorf("Username = %v, want matt", claims["Username"])
	}
}

// No secret is a deployment without the pool, not a broken one.
func TestNewMinterWithoutASecretIsNil(t *testing.T) {
	if m := NewMinter(""); m != nil {
		t.Errorf("NewMinter(\"\") = %v, want nil", m)
	}
}

// upstreamCookies stands the pool up and reports the token cookies it was sent.
func upstreamCookies(t *testing.T) (*httptest.Server, <-chan []*http.Cookie) {
	t.Helper()
	seen := make(chan []*http.Cookie, 1)
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokens []*http.Cookie
		for _, c := range r.Cookies() {
			if c.Name == tokenName {
				tokens = append(tokens, c)
			}
		}
		seen <- tokens
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(up.Close)
	return up, seen
}

func TestProxyMintsATokenForWhoeverIsSignedIn(t *testing.T) {
	up, seen := upstreamCookies(t)

	proxy, err := Proxy(up.URL, "/ght", as{login: "matt", may: true}, NewMinter("secret"))
	if err != nil {
		t.Fatal(err)
	}
	if err := proxy(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ght/aktuell", nil)); err != nil {
		t.Fatal(err)
	}

	got := <-seen
	if len(got) != 1 {
		t.Fatalf("upstream saw %d token cookies, want 1", len(got))
	}
	if got[0].Value == "" {
		t.Error("token cookie is empty")
	}
}

// The one that matters: a browser carrying a token from the pool's own sign-in,
// with no session here. Forwarding it would be a second way into the pool that
// needs no account, survives signing out and cannot be revoked. Whoever still
// has one signs in again.
func TestProxyDropsAStaleTokenFromAnUnknownCaller(t *testing.T) {
	up, seen := upstreamCookies(t)

	proxy, err := Proxy(up.URL, "/ght", as{may: false}, NewMinter("secret"))
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodGet, "/ght/aktuell", nil)
	r.AddCookie(&http.Cookie{Name: tokenName, Value: "from-the-old-sign-in"})
	if err := proxy(httptest.NewRecorder(), r); err != nil {
		t.Fatal(err)
	}

	if got := <-seen; len(got) != 0 {
		t.Errorf("upstream saw %d token cookies, want none — the old sign-in still works", len(got))
	}
}

// The browser should not be carrying a pool token — this app never gives it one
// — but the pool's own sign-in did, and a browser that used it before this
// existed still has it. Two tokens on one request is a question whose answer
// nobody should have to guess.
func TestProxyReplacesATokenTheBrowserSent(t *testing.T) {
	up, seen := upstreamCookies(t)

	proxy, err := Proxy(up.URL, "/ght", as{login: "matt", may: true}, NewMinter("secret"))
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodGet, "/ght/aktuell", nil)
	r.AddCookie(&http.Cookie{Name: tokenName, Value: "stale"})
	r.AddCookie(&http.Cookie{Name: "unrelated", Value: "kept"})
	if err := proxy(httptest.NewRecorder(), r); err != nil {
		t.Fatal(err)
	}

	got := <-seen
	if len(got) != 1 {
		t.Fatalf("upstream saw %d token cookies, want 1", len(got))
	}
	if got[0].Value == "stale" {
		t.Error("the browser's own token was forwarded instead of a minted one")
	}
}

// The pool has pages a visitor with no account may read. Gating the proxy would
// take those away, so an unidentified request is forwarded as it arrived.
func TestProxyForwardsAnonymouslyWhenNobodyIsSignedIn(t *testing.T) {
	for _, tt := range []struct {
		name string
		id   Identifier
	}{
		{"not signed in", as{may: false}},
		{"no identifier configured", nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			up, seen := upstreamCookies(t)

			proxy, err := Proxy(up.URL, "/ght", tt.id, NewMinter("secret"))
			if err != nil {
				t.Fatal(err)
			}
			if err := proxy(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ght/aktuell", nil)); err != nil {
				t.Fatal(err)
			}

			if got := <-seen; len(got) != 0 {
				t.Errorf("upstream saw %d token cookies, want none", len(got))
			}
		})
	}
}

// A deployment that has not been given the pool's signing key still proxies.
// Taking the whole app down over one area would be the wrong trade.
func TestProxyWithoutAMinterForwardsAnonymously(t *testing.T) {
	up, seen := upstreamCookies(t)

	proxy, err := Proxy(up.URL, "/ght", as{login: "matt", may: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := proxy(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ght/aktuell", nil)); err != nil {
		t.Fatal(err)
	}

	if got := <-seen; len(got) != 0 {
		t.Errorf("upstream saw %d token cookies, want none", len(got))
	}
}
