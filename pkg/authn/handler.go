package authn

import (
	"encoding/json"
	"net"
	"net/http"
	"time"
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
	Name        string              `json:"name"`
	Email       string              `json:"email"`
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
		Name:        a.DisplayName,
		Email:       a.Email,
	}
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
		return errStatus(http.StatusTooManyRequests, "too many attempts, try again later")
	}

	account, err := s.authenticatePassword(r, name, password)
	if err != nil {
		s.throttle.fail(name, client)
		s.log.Warn("sign-in failed", "account", name, "client", client)
		return errStatus(http.StatusUnauthorized, "invalid credentials")
	}
	s.throttle.succeed(name, client)

	token, err := s.store.CreateSession(r.Context(), account.ID, s.cfg.SessionTTL)
	if err != nil {
		return err
	}
	s.setSessionCookie(w, token)
	s.setHiDriveCookie(w, r, account)
	s.log.Info("signed in", "account", account.Name)

	return writeJSON(w, s.sessionResponse(account, time.Now().Add(s.cfg.SessionTTL)))
}

// authenticatePassword resolves and verifies in one place so that Login and the
// step-up before a passkey registration cannot drift apart.
func (s *Service) authenticatePassword(r *http.Request, name, password string) (*Account, error) {
	if name == "" || password == "" {
		return nil, ErrNoAccount
	}
	account, err := s.store.AccountByName(r.Context(), name)
	if err != nil {
		return nil, err
	}
	if !account.HasPassword() {
		return nil, ErrNoAccount
	}
	ok, rehash, err := verifyPassword(account.passwordHash, password)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNoAccount
	}
	// A successful sign-in is the only moment the plaintext is available, so it
	// is when an outdated hash gets replaced.
	if rehash {
		if updated, err := hashPassword(password); err != nil {
			s.log.Error("re-hashing password", "account", name, "err", err)
		} else if err := s.store.SetPasswordHash(r.Context(), account.ID, updated); err != nil {
			s.log.Error("storing re-hashed password", "account", name, "err", err)
		} else {
			s.log.Info("password hash upgraded", "account", name)
		}
	}
	return account, nil
}

// Session reports who is signed in. The frontend polls it, and it is also where
// the browser is handed a fresh HiDrive access token.
func (s *Service) Session(w http.ResponseWriter, r *http.Request) error {
	account, expires, ok := s.authenticate(r)
	if !ok {
		return errStatus(http.StatusUnauthorized, "not signed in")
	}
	s.setHiDriveCookie(w, r, account)
	return writeJSON(w, s.sessionResponse(account, expires))
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
	s.clearCookie(w, hiDriveCookie)
	return writeJSON(w, map[string]bool{"ok": true})
}

const hiDriveCookie = "hitoken"

// setHiDriveCookie hands the browser the access token for the account's HiDrive
// alias, which is what the media routes present upstream. A failure clears the
// cookie rather than leaving a stale token in place.
func (s *Service) setHiDriveCookie(w http.ResponseWriter, r *http.Request, a *Account) {
	if s.tokens == nil || a.HiDrive.Alias == "" {
		return
	}
	token, err := s.tokens.AccessToken(a.HiDrive.Alias)
	if err != nil {
		s.log.Warn("no HiDrive token for alias", "alias", a.HiDrive.Alias, "err", err)
		s.clearCookie(w, hiDriveCookie)
		return
	}
	// The media routes may be fetched cross-site by the player, which needs
	// SameSite=None — but a browser only accepts None together with Secure, so
	// over plain HTTP the cookie falls back to Lax. Same-origin playback, which
	// is what development is, works either way.
	sameSite, secure := http.SameSiteNoneMode, true
	if !s.cfg.CookieSecure {
		sameSite, secure = http.SameSiteLaxMode, false
	}
	http.SetCookie(w, &http.Cookie{
		Name:     hiDriveCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
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

func writeJSON(w http.ResponseWriter, v any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

// statusError carries an HTTP status so the router's error middleware can
// render it without this package importing the router.
type statusError struct {
	status  int
	message string
}

func (e *statusError) Error() string   { return e.message }
func (e *statusError) HTTPStatus() int { return e.status }

func errStatus(status int, message string) error {
	return &statusError{status: status, message: message}
}

func errBadRequest(message string) error {
	return errStatus(http.StatusBadRequest, message)
}
