package auth

import (
	"context"
	"github.com/ihleven/ihlvn/app/cmsauth"
	"github.com/interhome-group/cms/pkg/errs"
	"net"
	"net/http"
	"time"

	"github.com/ihleven/ihlvn/pkg/password"
)

// The sign-in surface: password login, the session the frontend polls, and
// logout.
//
// The response shape is the one the frontend already reads. It was written
// against JWT claims, so the fields keep those names even though there is no
// token any more: the mechanism changed, the contract did not.
type sessionResponse struct {
	Issuer      string              `json:"iss"`
	Subject     string              `json:"sub"`
	Audience    []string            `json:"aud"`
	ExpiresAt   int64               `json:"exp"`
	NotBefore   int64               `json:"nbf"`
	IssuedAt    int64               `json:"iat"`
	Permissions map[string]struct{} `json:"permissions"`
	// Modules are the feature areas the frontend may offer. Computed here at
	// serve time, so the client never has to know that an entitlement is really
	// a module.<id> permission.
	Modules []string `json:"modules"`
	Name    string   `json:"name"`
	Email   string   `json:"email"`
}

func (s *Service) sessionResponse(a *Account, expires time.Time) sessionResponse {
	permissions := make(map[string]struct{}, len(a.CMS.Permissions))
	for _, p := range a.CMS.Permissions {
		permissions[p] = struct{}{}
	}
	now := time.Now()
	return sessionResponse{
		Issuer:      s.cfg.Issuer,
		Subject:     a.Name,
		Audience:    []string{s.cfg.Audience},
		ExpiresAt:   expires.Unix(),
		NotBefore:   now.Unix(),
		IssuedAt:    now.Unix(),
		Permissions: permissions,
		Modules:     s.offered(a),
		Name:        a.DisplayName,
		Email:       a.Email,
	}
}

// offered is the areas the frontend may put in front of this account.
//
// Entitlement is nearly all of it, and for every area but one it is all of it.
// Browsing the storage additionally needs storage: an account that names no
// drive and a deployment that configures no shared one leave nothing to browse,
// and an entry that leads to an error is worse than no entry. So the area is
// withheld, which is indistinguishable from not being entitled — which is what
// someone in that position should see.
//
// The server still refuses the endpoints on the entitlement alone. This decides
// what to offer, not what is allowed.
func (s *Service) offered(a *Account) []string {
	entitled := cmsauth.Entitled(a)
	if a.HiDrive.Alias != "" || s.cfg.SharedDrive {
		return entitled
	}

	offered := make([]string, 0, len(entitled))
	for _, id := range entitled {
		if id != cmsauth.HidriveArea {
			offered = append(offered, id)
		}
	}
	return offered
}

// Login verifies a password and starts a session.
//
// Failures are deliberately uniform: an unknown account, a disabled one and a
// wrong password are all "invalid credentials", so the endpoint cannot be used
// to find out which accounts exist.
func (s *Service) Login(w http.ResponseWriter, r *http.Request) error {
	name := formValue(r, "username")
	password := formValue(r, "password")
	client := clientAddr(r)

	if s.throttle.blocked(name, client) {
		s.log.Warn("sign-in throttled", "account", name, "client", client)
		return errs.New("too many attempts, try again later", errs.HTTPStatus(http.StatusTooManyRequests))
	}

	account, err := s.authenticatePassword(r.Context(), name, password)
	if err != nil {
		s.throttle.fail(name, client)
		s.log.Warn("sign-in failed", "account", name, "client", client)
		return errs.New("invalid credentials", errs.HTTPStatus(http.StatusUnauthorized))
	}
	s.throttle.succeed(name, client)

	token, err := s.store.CreateSession(r.Context(), account.ID, s.cfg.SessionTTL)
	if err != nil {
		return err
	}
	s.setSessionCookie(w, token)
	s.log.Info("signed in", "account", account.Name)

	return writeJSON(w, http.StatusOK, s.sessionResponse(account, time.Now().Add(s.cfg.SessionTTL)))
}

// authenticatePassword resolves and verifies in one place so that Login and the
// step-up before a passkey registration cannot drift apart.
func (s *Service) authenticatePassword(ctx context.Context, name, plain string) (*Account, error) {
	if name == "" || plain == "" {
		return nil, ErrNoAccount
	}
	account, err := s.store.AccountByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if !account.HasPassword() {
		return nil, ErrNoAccount
	}
	ok, rehash, err := password.Verify(account.passwordHash, plain)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNoAccount
	}
	// A successful sign-in is the only moment the plaintext is available, so it
	// is when an outdated hash gets replaced.
	if rehash {
		if updated, err := password.Hash(plain); err != nil {
			s.log.Error("re-hashing password", "account", name, "err", err)
		} else if err := s.store.SetPasswordHash(ctx, account.ID, updated); err != nil {
			s.log.Error("storing re-hashed password", "account", name, "err", err)
		} else {
			s.log.Info("password hash upgraded", "account", name)
		}
	}
	return account, nil
}

// Session reports who is signed in. The frontend polls it.
func (s *Service) Session(w http.ResponseWriter, r *http.Request) error {
	account, expires, ok := s.authenticate(r)
	if !ok {
		return errs.New("not signed in", errs.HTTPStatus(http.StatusUnauthorized))
	}
	return writeJSON(w, http.StatusOK, s.sessionResponse(account, expires))
}

// Logout ends the session server-side before clearing the cookie, so a copy of
// the cookie taken beforehand is already useless.
func (s *Service) Logout(w http.ResponseWriter, r *http.Request) error {
	if token := cookieValue(r, s.cfg.CookieName); token != "" {
		if err := s.store.DeleteSession(r.Context(), token); err != nil {
			return err
		}
	}
	s.clearCookie(w, s.cfg.CookieName)
	return writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// clientAddr is the throttling key for "the same caller". Behind a proxy this
// is the proxy, which is why the account name is throttled alongside it.
func clientAddr(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// formValue reads a field from either a urlencoded or a multipart body.
//
// The SPA posts sign-in and the passkey ceremonies as FormData, so those bodies
// are multipart. ParseMultipartForm handles both: for a urlencoded body it
// parses the form on its way and then reports that the body was not multipart,
// which is not a failure here.
//
// Calling ParseForm first would break the multipart case — it leaves PostForm
// non-nil but empty, so the lazy multipart parse never runs and every field
// reads as "".
func formValue(r *http.Request, key string) string {
	if r.PostForm == nil {
		const maxMemory = 1 << 20
		_ = r.ParseMultipartForm(maxMemory)
	}
	return r.PostFormValue(key)
}
