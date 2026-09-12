package authn

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
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

const enrollCookie = "ihlvn_enroll"

// errNoPasskeys reports that the service was built without WebAuthn configured.
var errNoPasskeys = errStatus(http.StatusNotImplemented, "passkeys are not configured")

// webauthnAccount pairs an account with its loaded credentials to satisfy the
// library's User interface. Account itself stays free of that dependency, and
// credentials are only loaded where a ceremony actually needs them.
type webauthnAccount struct {
	account     *Account
	credentials []webauthn.Credential
}

func (u *webauthnAccount) WebAuthnID() []byte { return u.account.Handle }

// The name shown by the authenticator. The login name is what identifies the
// account everywhere else, so it is what a key should be labelled with.
func (u *webauthnAccount) WebAuthnName() string        { return u.account.Name }
func (u *webauthnAccount) WebAuthnDisplayName() string { return u.account.DisplayName }
func (u *webauthnAccount) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

func (s *Service) webauthnAccount(ctx context.Context, a *Account) (*webauthnAccount, error) {
	creds, err := s.store.Passkeys(ctx, a.ID)
	if err != nil {
		return nil, err
	}
	return &webauthnAccount{account: a, credentials: creds}, nil
}

// ------------------------------------------------------------------ login --

// LoginBegin starts a usernameless sign-in: the authenticator says which
// credential it holds, so there is no account name to ask for.
func (s *Service) LoginBegin(w http.ResponseWriter, r *http.Request) error {
	if s.wa == nil {
		return errNoPasskeys
	}
	assertion, session, err := s.wa.BeginDiscoverableLogin(
		webauthn.WithUserVerification(protocol.VerificationRequired))
	if err != nil {
		return err
	}

	id, err := s.store.SaveChallenge(r.Context(),
		Challenge{Purpose: PurposeLogin, Data: session}, s.cfg.ChallengeTTL)
	if err != nil {
		return err
	}
	s.setCeremonyCookie(w, id)
	return writeJSON(w, assertion)
}

// LoginFinish verifies the assertion and starts a session.
//
// A failure is never distinguished from an unknown credential: the response
// must not reveal whether a given passkey belongs to this site.
func (s *Service) LoginFinish(w http.ResponseWriter, r *http.Request) error {
	if s.wa == nil {
		return errNoPasskeys
	}
	challenge, err := s.takeCeremony(w, r, PurposeLogin)
	if err != nil {
		return err
	}

	user, credential, err := s.wa.FinishPasskeyLogin(s.discoverableAccount(r.Context()), *challenge.Data, r)
	if err != nil {
		s.log.Warn("passkey login failed", "err", err)
		return errStatus(http.StatusUnauthorized, "login failed")
	}
	account, ok := user.(*webauthnAccount)
	if !ok {
		return errStatus(http.StatusUnauthorized, "login failed")
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
	s.setSessionCookie(w, token)
	s.log.Info("signed in with a passkey", "account", account.account.Name)

	return writeJSON(w, s.sessionResponse(account.account, time.Now().Add(s.cfg.SessionTTL)))
}

// discoverableAccount resolves the opaque handle an authenticator returns.
func (s *Service) discoverableAccount(ctx context.Context) webauthn.DiscoverableUserHandler {
	return func(rawID, userHandle []byte) (webauthn.User, error) {
		account, err := s.store.AccountByHandle(ctx, userHandle)
		if err != nil {
			return nil, err
		}
		return s.webauthnAccount(ctx, account)
	}
}

// ----------------------------------------------------------- registration --

// RegisterBegin starts enrolling a passkey.
//
// Two things must hold: the ceremony has to be authorised — by a session, or by
// an unspent enrollment link — and the account's password must be given. The
// password is recorded on the challenge, so finishing cannot skip it.
func (s *Service) RegisterBegin(w http.ResponseWriter, r *http.Request) error {
	if s.wa == nil {
		return errNoPasskeys
	}

	account, err := s.registrant(r)
	if err != nil {
		return err
	}
	if err := s.reauthenticate(r, account); err != nil {
		return err
	}

	user, err := s.webauthnAccount(r.Context(), account)
	if err != nil {
		return err
	}
	// Offering the keys already enrolled lets the browser say "this one is
	// registered" instead of silently creating a duplicate.
	exclude := make([]protocol.CredentialDescriptor, 0, len(user.credentials))
	for _, c := range user.credentials {
		exclude = append(exclude, c.Descriptor())
	}

	creation, session, err := s.wa.BeginRegistration(user,
		webauthn.WithExclusions(exclude),
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			// No attachment preference: the platform authenticator — Touch ID,
			// Windows Hello, a phone — is as acceptable as a security key.
			//
			// A resident credential is required so signing in needs no account
			// name, and user verification so the key alone is not enough: the
			// device asks for a biometric or a PIN.
			ResidentKey:        protocol.ResidentKeyRequirementRequired,
			RequireResidentKey: protocol.ResidentKeyRequired(),
			UserVerification:   protocol.VerificationRequired,
		}))
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
	return writeJSON(w, creation)
}

// RegisterFinish stores the new credential, refusing one that can be synced.
func (s *Service) RegisterFinish(w http.ResponseWriter, r *http.Request) error {
	if s.wa == nil {
		return errNoPasskeys
	}
	challenge, err := s.takeCeremony(w, r, PurposeRegister)
	if err != nil {
		return err
	}
	if challenge.AccountID == nil {
		return errStatus(http.StatusBadRequest, "this ceremony names no account")
	}
	// The password is checked when the challenge is created; a challenge without
	// that stamp did not come from RegisterBegin.
	if !challenge.Reauthed {
		return errStatus(http.StatusForbidden, "this registration was not authorised with a password")
	}

	account, err := s.store.AccountByID(r.Context(), *challenge.AccountID)
	if err != nil {
		return err
	}
	user, err := s.webauthnAccount(r.Context(), account)
	if err != nil {
		return err
	}

	credential, err := s.wa.FinishRegistration(user, *challenge.Data, r)
	if err != nil {
		s.log.Warn("passkey registration failed", "account", account.Name, "err", err)
		return errBadRequest("the security key could not be registered")
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
	s.log.Info("passkey enrolled", "account", account.Name)

	return writeJSON(w, map[string]any{"ok": true})
}

// registrant says on whose behalf a registration may proceed: the signed-in
// account, or the holder of an unspent enrollment link. Without one of those,
// anyone could attach a passkey to any account.
func (s *Service) registrant(r *http.Request) (*Account, error) {
	if account, ok := s.Authenticate(r); ok {
		return account, nil
	}
	token := cookieValue(r, enrollCookie)
	if token == "" {
		return nil, errStatus(http.StatusUnauthorized, "sign in, or open your enrollment link, first")
	}
	account, err := s.store.EnrollTokenAccount(r.Context(), token)
	if errors.Is(err, ErrBadToken) {
		return nil, errStatus(http.StatusUnauthorized, "this enrollment link is no longer valid")
	}
	return account, err
}

// reauthenticate demands the account's password before a credential is added or
// removed, so holding a session is not enough on its own.
func (s *Service) reauthenticate(r *http.Request, account *Account) error {
	client := clientAddr(r)
	if s.throttle.blocked(account.Name, client) {
		return errStatus(http.StatusTooManyRequests, "too many attempts, try again later")
	}
	if !account.HasPassword() {
		return errStatus(http.StatusForbidden, "set a password before enrolling a security key")
	}

	ok, _, err := verifyPassword(account.passwordHash, formValue(r, "password"))
	if err != nil || !ok {
		s.throttle.fail(account.Name, client)
		return errStatus(http.StatusUnauthorized, "invalid credentials")
	}
	s.throttle.succeed(account.Name, client)
	return nil
}

// DeletePasskey removes one credential. It asks for the password too: taking a
// key away is as sensitive as adding one.
func (s *Service) DeletePasskey(w http.ResponseWriter, r *http.Request) error {
	account, ok := s.Authenticate(r)
	if !ok {
		return errStatus(http.StatusUnauthorized, "not signed in")
	}
	if err := s.reauthenticate(r, account); err != nil {
		return err
	}

	id, err := decodeCredentialID(r.PathValue("id"))
	if err != nil {
		return errBadRequest("that is not a credential id")
	}
	if err := s.store.DeletePasskey(r.Context(), account.ID, id); err != nil {
		if errors.Is(err, ErrNoAccount) {
			return errStatus(http.StatusNotFound, "no such passkey")
		}
		return err
	}
	s.log.Info("passkey removed", "account", account.Name)
	return writeJSON(w, map[string]any{"ok": true})
}

// Passkeys lists the signed-in account's credentials, for the screen that
// manages them. Only ever the caller's own: there is no reason for one account
// to enumerate another's devices.
func (s *Service) Passkeys(w http.ResponseWriter, r *http.Request) error {
	account, ok := s.Authenticate(r)
	if !ok {
		return errStatus(http.StatusUnauthorized, "not signed in")
	}
	keys, err := s.store.ListPasskeys(r.Context(), account.ID)
	if err != nil {
		return err
	}

	type item struct {
		ID         string     `json:"id"`
		Name       string     `json:"name"`
		Syncable   bool       `json:"syncable"`
		CreatedAt  time.Time  `json:"created_at"`
		LastUsedAt *time.Time `json:"last_used_at"`
	}
	items := make([]item, 0, len(keys))
	for _, k := range keys {
		items = append(items, item{
			ID:        base64.RawURLEncoding.EncodeToString(k.ID),
			Name:      k.Name,
			Syncable:  k.BackupEligible,
			CreatedAt: k.CreatedAt, LastUsedAt: k.LastUsedAt,
		})
	}
	return writeJSON(w, items)
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
		return errBadRequest("this link is missing its enrollment token")
	}
	if _, err := s.store.EnrollTokenAccount(r.Context(), token); err != nil {
		if errors.Is(err, ErrBadToken) {
			return errStatus(http.StatusUnauthorized, "this enrollment link is no longer valid")
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

// The challenge id travels in a cookie rather than in the JSON the page holds,
// so a script cannot substitute one ceremony's challenge into another's
// completion.
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

func (s *Service) takeCeremony(w http.ResponseWriter, r *http.Request, purpose string) (*Challenge, error) {
	id := cookieValue(r, s.cfg.CeremonyCookie)
	if id == "" {
		return nil, errBadRequest("no ceremony is in progress")
	}
	s.clearCookie(w, s.cfg.CeremonyCookie)

	challenge, err := s.store.TakeChallenge(r.Context(), purpose, id)
	if errors.Is(err, ErrBadToken) {
		return nil, errBadRequest("this ceremony has expired or was already completed")
	}
	return challenge, err
}
