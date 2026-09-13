package auth

import (
	"context"
	"errors"
	"github.com/interhome-group/cms/pkg/errs"
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/ihleven/ihlvn/pkg/passkey"
	"github.com/ihleven/ihlvn/pkg/password"
)

// Passkeys: both the kind held on one device and the kind synced between them.
//
// A synced passkey — iCloud Keychain, Google, a password manager — is only as
// strong as the account it syncs through, where a hardware key keeps the private
// key on one authenticator. Both are accepted, because requiring hardware would
// mean nobody without a security key could enrol at all, and a synced passkey is
// still phishing-resistant in a way a password is not. Which kind a credential
// is gets recorded, so it can be seen rather than guessed.
//
// Registering one always requires the account's password as well, whether the
// ceremony was authorised by an existing session or by an enrollment link.
// Otherwise whoever holds a stolen cookie, or intercepts the link, could attach
// a credential of their own and keep access indefinitely.
//
// So an account with no password cannot enrol at all, and a link is not a way
// around that: an administrator sets the first password — from the admin section
// or the terminal — and it is conveyed separately from the link. Two things have
// to reach the same person before a credential can be attached, and that is the
// point.

const enrollCookie = "ihlvn_enroll"

// errNoPasskeys reports that the service was built without WebAuthn configured.
var errNoPasskeys = errs.New("passkeys are not configured", errs.HTTPStatus(http.StatusNotImplemented))

// subject pairs an account with its loaded credentials, which is all a ceremony
// is told about it. Account itself stays free of the dependency, and credentials
// are only loaded where a ceremony actually needs them.
type subject struct {
	account     *Account
	credentials []webauthn.Credential
}

func (u *subject) Handle() []byte { return u.account.Handle }

// The name shown by the authenticator. The login name is what identifies the
// account everywhere else, so it is what a key should be labelled with.
func (u *subject) Name() string        { return u.account.Name }
func (u *subject) DisplayName() string { return u.account.DisplayName }
func (u *subject) Credentials() []webauthn.Credential {
	return u.credentials
}

func (s *Service) subject(ctx context.Context, a *Account) (*subject, error) {
	creds, err := s.store.Passkeys(ctx, a.ID)
	if err != nil {
		return nil, err
	}
	return &subject{account: a, credentials: creds}, nil
}

// ------------------------------------------------------------------ login --

// LoginBegin starts a usernameless sign-in: the authenticator says which
// credential it holds, so there is no account name to ask for.
func (s *Service) LoginBegin(w http.ResponseWriter, r *http.Request) error {
	if s.ceremonies == nil {
		return errNoPasskeys
	}
	assertion, session, err := s.ceremonies.BeginLogin()
	if err != nil {
		return err
	}

	id, err := s.store.SaveChallenge(r.Context(),
		Challenge{Purpose: PurposeLogin, Data: session}, s.cfg.ChallengeTTL)
	if err != nil {
		return err
	}
	s.setCeremonyCookie(w, id)
	return writeJSON(w, http.StatusOK, ceremonyEnvelope{PublicKey: assertion.Response, Ceremony: id})
}

// LoginFinish verifies the assertion and starts a session.
//
// A failure is never distinguished from an unknown credential: the response
// must not reveal whether a given passkey belongs to this site.
func (s *Service) LoginFinish(w http.ResponseWriter, r *http.Request) error {
	if s.ceremonies == nil {
		return errNoPasskeys
	}
	challenge, err := s.takeCeremony(r, PurposeLogin)
	if err != nil {
		return err
	}

	signedIn, credential, err := s.ceremonies.FinishLogin(*challenge.Data, r.Body, s.discoverableAccount(r.Context()))
	if err != nil {
		s.log.Warn("passkey login failed", "err", err)
		return errs.New("login failed", errs.HTTPStatus(http.StatusUnauthorized))
	}
	account, ok := signedIn.(*subject)
	if !ok {
		return errs.New("login failed", errs.HTTPStatus(http.StatusUnauthorized))
	}

	// The sign counter and the flags advance during a login; losing that write
	// is not worth failing the sign-in over.
	if err := s.store.UpdatePasskey(r.Context(), credential); err != nil {
		s.log.Error("recording passkey use", "err", err)
	}

	token, err := s.store.CreateSession(r.Context(), account.account.ID, s.cfg.SessionTTL)
	if err != nil {
		return err
	}
	s.clearCookie(w, s.cfg.CeremonyCookie)
	s.setSessionCookie(w, token)
	s.log.Info("signed in with a passkey", "account", account.account.Name)

	return writeJSON(w, http.StatusOK, s.sessionResponse(account.account, time.Now().Add(s.cfg.SessionTTL)))
}

// discoverableAccount resolves the opaque handle an authenticator returns.
func (s *Service) discoverableAccount(ctx context.Context) passkey.Resolver {
	return func(rawID, handle []byte) (passkey.Subject, error) {
		account, err := s.store.AccountByHandle(ctx, handle)
		if err != nil {
			return nil, err
		}
		return s.subject(ctx, account)
	}
}

// ----------------------------------------------------------- registration --

// RegisterBegin starts enrolling a passkey.
//
// Two things must hold: the ceremony has to be authorised — by a session, or by
// an unspent enrollment link — and the account's password must be given. The
// password is recorded on the challenge, so finishing cannot skip it.
func (s *Service) RegisterBegin(w http.ResponseWriter, r *http.Request) error {
	if s.ceremonies == nil {
		return errNoPasskeys
	}

	account, err := s.registrant(r)
	if err != nil {
		return err
	}
	if err := s.reauthenticate(r, account); err != nil {
		return err
	}

	enrolling, err := s.subject(r.Context(), account)
	if err != nil {
		return err
	}

	creation, session, err := s.ceremonies.BeginRegistration(enrolling)
	if err != nil {
		return err
	}

	id, err := s.store.SaveChallenge(r.Context(), Challenge{
		Purpose:   PurposeRegister,
		AccountID: &account.ID,
		Data:      session,
		Reauthed:  true,
	}, s.cfg.ChallengeTTL)
	if err != nil {
		return err
	}
	s.setCeremonyCookie(w, id)
	return writeJSON(w, http.StatusOK, ceremonyEnvelope{PublicKey: creation.Response, Ceremony: id})
}

// RegisterFinish stores the new credential, refusing one that can be synced.
func (s *Service) RegisterFinish(w http.ResponseWriter, r *http.Request) error {
	if s.ceremonies == nil {
		return errNoPasskeys
	}
	challenge, err := s.takeCeremony(r, PurposeRegister)
	if err != nil {
		return err
	}
	if challenge.AccountID == nil {
		return errs.New("this ceremony names no account", errs.HTTPStatus(http.StatusBadRequest))
	}
	// The password is checked when the challenge is created; a challenge without
	// that stamp did not come from RegisterBegin.
	if !challenge.Reauthed {
		return errs.New("this registration was not authorised with a password", errs.HTTPStatus(http.StatusForbidden))
	}

	account, err := s.store.AccountByID(r.Context(), *challenge.AccountID)
	if err != nil {
		return err
	}
	enrolling, err := s.subject(r.Context(), account)
	if err != nil {
		return err
	}

	credential, err := s.ceremonies.FinishRegistration(enrolling, *challenge.Data, r.Body)
	if err != nil {
		s.log.Warn("passkey registration failed", "account", account.Name, "err", err)
		return errs.New("the security key could not be registered", errs.HTTPStatus(http.StatusBadRequest))
	}

	if credential.Flags.BackupEligible {
		// Recorded rather than refused: worth knowing which accounts are behind
		// a key that can leave the device it was created on.
		s.log.Info("enrolled a syncable passkey", "account", account.Name)
	}

	if err := s.store.AddPasskey(r.Context(), account.ID, credential, r.URL.Query().Get("name")); err != nil {
		return err
	}

	// The enrollment link is spent only now that a key actually exists, so an
	// abandoned ceremony does not cost someone their one link.
	if token := cookieValue(r, enrollCookie); token != "" {
		if _, err := s.store.ConsumeEnrollToken(r.Context(), token); err != nil {
			s.log.Error("consuming the enrollment token", "err", err)
		}
		s.clearCookie(w, enrollCookie)
	}
	s.clearCookie(w, s.cfg.CeremonyCookie)
	s.log.Info("passkey enrolled", "account", account.Name)

	return writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Service) registrant(r *http.Request) (*Account, error) {
	if account, ok := s.Authenticate(r); ok {
		return account, nil
	}
	token := cookieValue(r, enrollCookie)
	if token == "" {
		return nil, errs.New("sign in, or open your enrollment link, first", errs.HTTPStatus(http.StatusUnauthorized))
	}
	account, err := s.store.EnrollTokenAccount(r.Context(), token)
	if errors.Is(err, ErrBadToken) {
		return nil, errs.New("this enrollment link is no longer valid", errs.HTTPStatus(http.StatusUnauthorized))
	}
	return account, err
}

// reauthenticate demands the account's password before a credential is added or
// removed, so holding a session is not enough on its own.
func (s *Service) reauthenticate(r *http.Request, account *Account) error {
	client := clientAddr(r)
	if s.throttle.blocked(account.Name, client) {
		return errs.New("too many attempts, try again later", errs.HTTPStatus(http.StatusTooManyRequests))
	}
	if !account.HasPassword() {
		return errs.New("set a password before enrolling a security key", errs.HTTPStatus(http.StatusForbidden))
	}

	ok, _, err := password.Verify(account.passwordHash, formValue(r, "password"))
	if err != nil || !ok {
		s.throttle.fail(account.Name, client)
		return errs.New("invalid credentials", errs.HTTPStatus(http.StatusUnauthorized))
	}
	s.throttle.succeed(account.Name, client)
	return nil
}

// DeletePasskey removes one credential. It asks for the password too: taking a
// key away is as sensitive as adding one.
func (s *Service) DeletePasskey(w http.ResponseWriter, r *http.Request) error {
	account, ok := s.Authenticate(r)
	if !ok {
		return errs.New("not signed in", errs.HTTPStatus(http.StatusUnauthorized))
	}
	if err := s.reauthenticate(r, account); err != nil {
		return err
	}

	id, err := decodeCredentialID(r.PathValue("id"))
	if err != nil {
		return errs.New("that is not a credential id", errs.HTTPStatus(http.StatusBadRequest))
	}
	if err := s.store.DeletePasskey(r.Context(), account.ID, id); err != nil {
		if errors.Is(err, ErrNoAccount) {
			return errs.New("no such passkey", errs.HTTPStatus(http.StatusNotFound))
		}
		return err
	}
	s.log.Info("passkey removed", "account", account.Name)
	return writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// Passkeys lists the signed-in account's credentials, for the screen that
// manages them. Only ever the caller's own: there is no reason for one account
// to enumerate another's devices.
func (s *Service) Passkeys(w http.ResponseWriter, r *http.Request) error {
	account, ok := s.Authenticate(r)
	if !ok {
		return errs.New("not signed in", errs.HTTPStatus(http.StatusUnauthorized))
	}
	keys, err := s.store.ListPasskeys(r.Context(), account.ID)
	if err != nil {
		return err
	}

	items := make([]PasskeyInfo, 0, len(keys))
	for _, k := range keys {
		items = append(items, k.Info())
	}
	return writeJSON(w, http.StatusOK, items)
}

// ------------------------------------------------------------- enrollment --

// Enroll accepts an enrollment link, moves the token out of the URL into a
// cookie and redirects.
//
// The token is not spent here: the person still has to complete the ceremony,
// and merely opening the link — a mail client prefetching it, say — must not
// burn it. Moving it into a cookie keeps it out of the address bar, the history
// and any Referer header.
func (s *Service) Enroll(w http.ResponseWriter, r *http.Request) error {
	token := r.URL.Query().Get("t")
	if token == "" {
		return errs.New("this link is missing its enrollment token", errs.HTTPStatus(http.StatusBadRequest))
	}
	if _, err := s.store.EnrollTokenAccount(r.Context(), token); err != nil {
		if errors.Is(err, ErrBadToken) {
			return errs.New("this enrollment link is no longer valid", errs.HTTPStatus(http.StatusUnauthorized))
		}
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     enrollCookie,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(s.cfg.EnrollTTL),
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure,
		SameSite: s.cfg.SameSite,
	})
	http.Redirect(w, r, "/enroll", http.StatusSeeOther)
	return nil
}

// ----------------------------------------------------------------- shared --

// CeremonyHeader carries the challenge id back from the page that started the
// ceremony. It exists because a browser can have two ceremonies open at once:
// one offered silently in the username field's autofill, one started by the
// button. A single cookie cannot name both, so the second to finish used to
// find no ceremony at all.
//
// Sending the id is not a weakening: a challenge is single-use and scoped to its
// purpose server-side, and knowing its id proves nothing — completing it still
// needs a signature over that challenge from a registered credential.
const CeremonyHeader = "X-Ceremony"

// ceremonyEnvelope is what a begin returns: the options the browser needs, plus
// the id of the ceremony they belong to.
type ceremonyEnvelope struct {
	PublicKey any    `json:"publicKey"`
	Ceremony  string `json:"ceremony"`
}

// The challenge id also travels in a cookie, so a page that does not send the
// header still works, and so a ceremony is tied to the browser that began it.
func (s *Service) setCeremonyCookie(w http.ResponseWriter, id string) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.CeremonyCookie,
		Value:    id,
		Path:     "/",
		Expires:  time.Now().Add(s.cfg.ChallengeTTL),
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure,
		SameSite: s.cfg.SameSite,
	})
}

// takeCeremony resolves which ceremony is being completed.
//
// The cookie is deliberately not cleared here. It used to be, before the
// assertion had even been checked, so one attempt that failed — or one of two
// concurrent attempts — destroyed the other. The challenge itself is single-use
// in the store, which is what actually stops a replay.
func (s *Service) takeCeremony(r *http.Request, purpose string) (*Challenge, error) {
	id := r.Header.Get(CeremonyHeader)
	if id == "" {
		id = cookieValue(r, s.cfg.CeremonyCookie)
	}
	if id == "" {
		return nil, errs.New("no ceremony is in progress", errs.HTTPStatus(http.StatusBadRequest))
	}

	challenge, err := s.store.TakeChallenge(r.Context(), purpose, id)
	if errors.Is(err, ErrBadToken) {
		return nil, errs.New("this ceremony has expired or was already completed", errs.HTTPStatus(http.StatusBadRequest))
	}
	return challenge, err
}
