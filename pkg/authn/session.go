package authn

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ctxKey is unexported and zero-size, so nothing outside this package can put a
// value under it or read one out by accident.
type ctxKey struct{}

// FromContext returns the account a request was authenticated as.
func FromContext(ctx context.Context) (*Account, bool) {
	a, ok := ctx.Value(ctxKey{}).(*Account)
	return a, ok
}

// The session token is the whole credential, so the cookie is HttpOnly and, off
// localhost, Secure. Lax rather than Strict because sign-in returns from a
// top-level navigation.
func (s *Service) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.CookieName,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(s.cfg.SessionTTL),
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure,
		SameSite: s.cfg.SameSite,
	})
}

func (s *Service) clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure,
		SameSite: s.cfg.SameSite,
	})
}

func cookieValue(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return c.Value
}

// Authenticate resolves a request to its account. It answers who the caller is
// and does not decide whether that is enough — Require and Optional do that.
//
// The account is loaded from the database on every request rather than trusted
// from the cookie, so disabling an account or changing its permissions takes
// effect immediately instead of at the next sign-in.
func (s *Service) Authenticate(r *http.Request) (*Account, bool) {
	account, _, ok := s.authenticate(r)
	return account, ok
}

// authenticate also reports when the session expires, which the session
// endpoint reports to the frontend.
func (s *Service) authenticate(r *http.Request) (*Account, time.Time, bool) {
	token := cookieValue(r, s.cfg.CookieName)
	if token == "" {
		return nil, time.Time{}, false
	}
	account, expires, err := s.store.SessionAccount(r.Context(), token, s.cfg.SessionTTL, s.cfg.SlideAfter)
	switch {
	case errors.Is(err, ErrNoSession), errors.Is(err, ErrNoAccount):
		return nil, time.Time{}, false
	case err != nil:
		s.log.Error("resolving session", "err", err)
		return nil, time.Time{}, false
	}
	return account, expires, true
}

// WithAccount returns a context carrying the account. Handlers that do their own
// gating — because they report failure as an error rather than writing a
// response — use it instead of Require.
func WithAccount(ctx context.Context, a *Account) context.Context {
	return context.WithValue(ctx, ctxKey{}, a)
}

// Require rejects a request that is not signed in.
func (s *Service) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		account, ok := s.Authenticate(r)
		if !ok {
			s.denied(w, r)
			return
		}
		next.ServeHTTP(w, r.WithContext(WithAccount(r.Context(), account)))
	})
}

func (s *Service) RequireFunc(next func(http.ResponseWriter, *http.Request)) http.Handler {
	return s.Require(http.HandlerFunc(next))
}

// Optional attaches the account when there is one and lets the request through
// either way. It is what the public pages use: they are readable by anyone, but
// a signed-in reader may be shown more.
func (s *Service) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if account, ok := s.Authenticate(r); ok {
			r = r.WithContext(WithAccount(r.Context(), account))
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Service) OptionalFunc(next func(http.ResponseWriter, *http.Request)) http.Handler {
	return s.Optional(http.HandlerFunc(next))
}

// denied sends a browser to the sign-in page and everything else a 401.
// Redirecting an XHR or a video request would hand the caller an HTML page
// where it expected data.
func (s *Service) denied(w http.ResponseWriter, r *http.Request) {
	if !wantsHTML(r) {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	target := "/login"
	if next := r.URL.RequestURI(); isLocalPath(next) && next != "/" {
		target += "?next=" + url.QueryEscape(next)
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

// wantsHTML reports whether the caller is a browser navigating, as opposed to
// fetching data.
func wantsHTML(r *http.Request) bool {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return false
	}
	// A ranged request is a media element fetching bytes, never a navigation,
	// whatever it says it accepts.
	if r.Header.Get("Range") != "" {
		return false
	}
	// Where the browser tells us outright, believe it.
	if dest := r.Header.Get("Sec-Fetch-Dest"); dest != "" {
		return dest == "document"
	}
	return strings.Contains(r.Header.Get("Accept"), "text/html")
}

// isLocalPath keeps the post-login redirect on this site.
func isLocalPath(p string) bool {
	return strings.HasPrefix(p, "/") && !strings.HasPrefix(p, "//")
}
