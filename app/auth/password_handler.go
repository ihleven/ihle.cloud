package auth

import (
	"net/http"

	"github.com/ihleven/ihlvn/pkg/password"
	"github.com/interhome-group/cms/pkg/errs"
)

// ChangePassword replaces the password of whoever is signed in.
//
// It exists because without it a whole class of account cannot change theirs at
// all. A geheimtipp player signs in with the password they set on the pool years
// ago; the only other way to replace it is an administrator, who would then know
// it. Passkey enrollment is no way out either — the step-up before it refuses an
// account whose password it cannot verify.
//
// The current password is required even though the session already proves who
// is asking. A session outlives the moment it was created: a browser left open
// is enough to take over an account permanently if the password can be changed
// without knowing it.
//
// The screening is the same advice-not-verdict the administrator's path gives,
// and it is overridden the same way, by asking again.
func (s *Service) ChangePassword(w http.ResponseWriter, r *http.Request) error {
	account, _, ok := s.authenticate(r)
	if !ok {
		return errs.New("not signed in", errs.HTTPStatus(http.StatusUnauthorized))
	}

	var body struct {
		Current string `json:"current"`
		Next    string `json:"password"`
		Confirm bool   `json:"confirm"`
	}
	if err := decode(r, &body); err != nil {
		return err
	}

	if _, err := s.verifyPassword(r.Context(), account, body.Current); err != nil {
		// Deliberately not the uniform "invalid credentials" of sign-in: who is
		// asking is already known, so there is nothing here to enumerate, and
		// saying which field was wrong is the difference between a usable form
		// and a guessing game.
		return errs.New("the current password is wrong", errs.HTTPStatus(http.StatusUnauthorized))
	}

	if body.Next == "" {
		return errs.New("no new password given", errs.HTTPStatus(http.StatusBadRequest))
	}
	if password.TooLong(body.Next) {
		return errs.New("use at most %d characters", password.MaxLength, errs.HTTPStatus(http.StatusBadRequest))
	}

	advice := password.Check(r.Context(), s.breaches, body.Next, s.cfg.MinPasswordLength)
	if advice.Concerning() && !body.Confirm {
		return writeJSON(w, http.StatusOK, PasswordResult{Set: false, Advice: advice})
	}

	hash, err := password.Hash(body.Next)
	if err != nil {
		return err
	}
	if err := s.store.SetPasswordHash(r.Context(), account.ID, hash); err != nil {
		return err
	}
	s.log.Info("password changed", "account", account.Name)

	return writeJSON(w, http.StatusOK, PasswordResult{Set: true, Advice: advice})
}
