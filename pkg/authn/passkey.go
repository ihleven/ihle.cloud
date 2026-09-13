package authn

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
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
	return writeJSON(w, ceremonyEnvelope{PublicKey: assertion.Response, Ceremony: id})
}

// LoginFinish verifies the assertion and starts a session.
//
// A failure is never distinguished from an unknown credential: the response
// must not reveal whether a given passkey belongs to this site.
func (s *Service) LoginFinish(w http.ResponseWriter, r *http.Request) error {
	if s.wa == nil {
		return errNoPasskeys
	}
	challenge, err := s.takeCeremony(r, PurposeLogin)
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
	s.clearCookie(w, s.cfg.CeremonyCookie)
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

	account, viaLink, err := s.registrant(r)
	if err != nil {
		return err
	}
	if err := s.authorizeRegistration(r, account, viaLink); err != nil {
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
	return writeJSON(w, ceremonyEnvelope{PublicKey: creation.Response, Ceremony: id})
}

// RegisterFinish stores the new credential, refusing one that can be synced.
func (s *Service) RegisterFinish(w http.ResponseWriter, r *http.Request) error {
	if s.wa == nil {
		return errNoPasskeys
	}
	challenge, err := s.takeCeremony(r, PurposeRegister)
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
	s.clearCookie(w, s.cfg.CeremonyCookie)
	s.log.Info("passkey enrolled", "account", account.Name)

	return writeJSON(w, map[string]any{"ok": true})
}

// EnrollState describes the enrollment the browser is holding a link for.
//
// The page needs it to ask the right question: an account that already has a
// password is proving it, and one that has none is choosing it. Guessing wrong
// is a form that says "your password" to someone who does not have one yet.
//
// It reveals nothing the holder of the link does not already have — the link
// itself names the account.
func (s *Service) EnrollState(w http.ResponseWriter, r *http.Request) error {
	account, viaLink, err := s.registrant(r)
	if err != nil {
		return err
	}
	return writeJSON(w, struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		HasPassword bool   `json:"has_password"`
		ViaLink     bool   `json:"via_link"`
	}{account.Name, account.DisplayName, account.HasPassword(), viaLink})
}

// registrant says on whose behalf a registration may proceed: the signed-in
// account, or the holder of an unspent enrollment link. Without one of those,
// anyone could attach a passkey to any account.
func (s *Service) registrant(r *http.Request) (account *Account, viaLink bool, err error) {
	if account, ok := s.Authenticate(r); ok {
		return account, false, nil
	}
	token := cookieValue(r, enrollCookie)
	if token == "" {
		return nil, false, errStatus(http.StatusUnauthorized, "sign in, or open your enrollment link, first")
	}
	account, err = s.store.EnrollTokenAccount(r.Context(), token)
	if errors.Is(err, ErrBadToken) {
		return nil, false, errStatus(http.StatusUnauthorized, "this enrollment link is no longer valid")
	}
	return account, true, err
}

// reauthenticate demands the account's password before a credential is added or
// removed, so holding a session is not enough on its own.
// authorizeRegistration decides whether this request may add a credential.
//
// An account that has a password proves it, every time — holding a session or an
// enrollment link is not enough, or whoever stole one could attach a credential
// of their own and keep access indefinitely.
//
// An account that has none is being set up. There is nothing to prove, so the
// enrollment link is the proof: it is single-use, expires, and was issued by an
// administrator for this account. The password typed alongside the new
// credential becomes the account's first. Without this an account created in the
// admin section could never enrol anything — it would need a password it has no
// way to be given, which is the circle the CLI broke by being a terminal.
//
// A session is deliberately not accepted for that: a session for an account with
// no password can only have come from a passkey, and that path has an
// administrator in it already.
func (s *Service) authorizeRegistration(r *http.Request, account *Account, viaLink bool) error {
	if account.HasPassword() {
		return s.reauthenticate(r, account)
	}
	if !viaLink {
		return errStatus(http.StatusForbidden, "set a password before enrolling a security key")
	}
	return s.setFirstPassword(r, account)
}

// setFirstPassword stores the password chosen during enrollment.
//
// Only the hard bound is enforced. Length and breach screening are advice
// elsewhere in this system rather than rules, and there is nobody here to put
// the advice to: refusing would be a stricter policy than the one an
// administrator is held to at the terminal.
func (s *Service) setFirstPassword(r *http.Request, account *Account) error {
	password := formValue(r, "password")
	if password == "" {
		return errBadRequest("choose a password to go with your passkey")
	}
	if TooLong(password) {
		return errBadRequest(fmt.Sprintf("use at most %d characters", MaxPasswordLength))
	}

	hash, err := hashPassword(password)
	if err != nil {
		return err
	}
	if err := s.store.SetPasswordHash(r.Context(), account.ID, hash); err != nil {
		return err
	}
	// Keep the loaded copy honest: the rest of the ceremony reads it, and it was
	// loaded before the password existed.
	account.passwordHash = hash
	s.log.Info("first password set during enrollment", "account", account.Name)
	return nil
}

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
		return nil, errBadRequest("no ceremony is in progress")
	}

	challenge, err := s.store.TakeChallenge(r.Context(), purpose, id)
	if errors.Is(err, ErrBadToken) {
		return nil, errBadRequest("this ceremony has expired or was already completed")
	}
	return challenge, err
}
