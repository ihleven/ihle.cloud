package accounts

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/ihleven/ihlvn/pkg/authn"
	"github.com/interhome-group/cms/pkg/errs"
)

// How an account can sign in: a password, and the passkeys registered against
// it. Both are administered here, but neither is administered blindly — the
// screening a password gets from the terminal it gets here too, and revoking a
// credential is separated from editing an account so it cannot happen by
// accident while someone is fixing a display name.

type passwordBody struct {
	Password string `json:"password"`

	// Confirm carries the operator's decision to use a password the screening
	// objected to. The check is advice, not a verdict, so refusing outright
	// would substitute a number and a third party's corpus for their judgement —
	// but it should take a second, deliberate request to override.
	Confirm bool `json:"confirm"`
}

type passwordResult struct {
	Set bool `json:"set"`

	// Advice is what the screening found. Present whether or not the password
	// was set, so a caller that is told "not set" is also told why.
	Advice authn.PasswordAdvice `json:"advice"`
}

// SetPassword sets or replaces an account's password.
//
// It exists for recovery, not for onboarding: a password an administrator
// chooses is one they know and have to convey somehow. The enrollment link is
// the better path for a new account, because the person picks their own and it
// never travels.
func (a *API) SetPassword(w http.ResponseWriter, r *http.Request) error {
	account, err := a.lookup(r)
	if err != nil {
		return err
	}

	var body passwordBody
	if err := decode(r, &body); err != nil {
		return err
	}
	if body.Password == "" {
		return errs.New("no password given", errs.HTTPStatus(http.StatusBadRequest))
	}
	// The one hard bound, and the only refusal that cannot be overridden: it is
	// a technical limit rather than an opinion about the password.
	if authn.TooLong(body.Password) {
		return errs.New("use at most %d characters", authn.MaxPasswordLength,
			errs.HTTPStatus(http.StatusBadRequest))
	}

	advice := authn.CheckPassword(r.Context(), a.breaches, body.Password)
	if advice.Concerning() && !body.Confirm {
		return writeJSON(w, http.StatusOK, passwordResult{Set: false, Advice: advice})
	}

	hash, err := authn.HashPassword(body.Password)
	if err != nil {
		return err
	}
	if err := a.store.SetPasswordHash(r.Context(), account.ID, hash); err != nil {
		return asStatus(err)
	}
	return writeJSON(w, http.StatusOK, passwordResult{Set: true, Advice: advice})
}

type enrollment struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Enroll issues a single-use link for registering a passkey.
//
// The link is returned rather than sent: this app has no reliable way to reach
// someone, and an administrator handing it over in person or in a chat is both
// honest about that and better than pretending an email was delivered.
func (a *API) Enroll(w http.ResponseWriter, r *http.Request) error {
	account, err := a.lookup(r)
	if err != nil {
		return err
	}
	if account.Disabled {
		return errs.New("%s is disabled; enable it before enrolling a device", account.Name,
			errs.HTTPStatus(http.StatusConflict))
	}

	token, err := a.store.CreateEnrollToken(r.Context(), account.ID, a.enrollTTL)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, enrollment{
		URL:       a.origin + "/auth/enroll?t=" + token,
		ExpiresAt: time.Now().Add(a.enrollTTL),
	})
}

type passkey struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Syncable   bool       `json:"syncable"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
}

// Passkeys lists an account's registered devices. The same projection the
// self-service page gets: a credential's public key is of no use to an operator,
// and when it was last used is what tells them whether it is still someone's.
func (a *API) Passkeys(w http.ResponseWriter, r *http.Request) error {
	account, err := a.lookup(r)
	if err != nil {
		return err
	}
	keys, err := a.store.ListPasskeys(r.Context(), account.ID)
	if err != nil {
		return err
	}

	items := make([]passkey, 0, len(keys))
	for _, k := range keys {
		items = append(items, passkey{
			ID:         base64.RawURLEncoding.EncodeToString(k.ID),
			Name:       k.Name,
			Syncable:   k.BackupEligible,
			CreatedAt:  k.CreatedAt,
			LastUsedAt: k.LastUsedAt,
		})
	}
	return writeJSON(w, http.StatusOK, items)
}

// DeletePasskey removes one device, for the ordinary case of a phone that was
// lost or replaced.
func (a *API) DeletePasskey(w http.ResponseWriter, r *http.Request) error {
	account, err := a.lookup(r)
	if err != nil {
		return err
	}
	id, err := base64.RawURLEncoding.DecodeString(r.PathValue("id"))
	if err != nil {
		return errs.New("that is not a passkey id", errs.HTTPStatus(http.StatusBadRequest))
	}
	if err := a.store.DeletePasskey(r.Context(), account.ID, id); err != nil {
		return asStatus(err)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

type revoked struct {
	Passkeys int64 `json:"passkeys"`
	Sessions int64 `json:"sessions"`
}

// Revoke removes every passkey and drops every session: the "this account is
// compromised" action, kept apart from the edit form so it cannot be done by
// mistake.
//
// The password is deliberately left alone. Revoking is about the credentials
// that were taken; clearing the password as well would leave an account that
// cannot be recovered without a second administrator.
func (a *API) Revoke(w http.ResponseWriter, r *http.Request) error {
	account, err := a.lookup(r)
	if err != nil {
		return err
	}
	keys, err := a.store.DeletePasskeys(r.Context(), account.ID)
	if err != nil {
		return err
	}
	sessions, err := a.store.DeleteSessions(r.Context(), account.ID)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, revoked{Passkeys: keys, Sessions: sessions})
}

// DeleteSessions signs an account out everywhere without touching what it can
// sign in with.
func (a *API) DeleteSessions(w http.ResponseWriter, r *http.Request) error {
	account, err := a.lookup(r)
	if err != nil {
		return err
	}
	n, err := a.store.DeleteSessions(r.Context(), account.ID)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, revoked{Sessions: n})
}
