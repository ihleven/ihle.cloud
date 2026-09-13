package auth

import (
	"context"
	"github.com/interhome-group/cms/pkg/errs"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/ihleven/ihlvn/pkg/password"
)

// A ceremony cannot be driven from a test — there is no authenticator to sign
// the challenge — so what is covered here is everything around it: who may start
// one, what the server refuses to store, and the invariants that hold whether or
// not the signature ever arrives.

func passkeyService(t *testing.T) (*Service, context.Context) {
	t.Helper()
	store, ctx := testStore(t)
	svc, err := NewService(store, Config{
		CookieName: "session",
		SessionTTL: time.Hour,
		SlideAfter: time.Minute,
		RPID:       "localhost",
		RPOrigins:  []string{"http://localhost:8000"},
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return svc, ctx
}

func status(t *testing.T, err error) int {
	t.Helper()
	if err == nil {
		return http.StatusOK
	}
	return errs.StatusOf(err)
}

func registerBegin(svc *Service, password string, cookies ...*http.Cookie) (*httptest.ResponseRecorder, error) {
	form := url.Values{"password": {password}}
	r := httptest.NewRequest(http.MethodPost, "/auth/passkey/register/begin", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.RemoteAddr = "10.0.0.1:1234"
	for _, c := range cookies {
		r.AddCookie(c)
	}
	w := httptest.NewRecorder()
	return w, svc.RegisterBegin(w, r)
}

// Both kinds of credential are accepted, and which kind a credential is has to
// survive into the row, since that is the only place it can be seen afterwards.
func TestSyncablePasskeyIsRecordedNotRefused(t *testing.T) {
	svc, ctx := passkeyService(t)
	account := mustAccount(t, svc.store, ctx, "matt")

	synced := &webauthn.Credential{
		ID: []byte("synced"), PublicKey: []byte("k"),
		Flags: webauthn.CredentialFlags{BackupEligible: true, BackupState: true},
	}
	bound := &webauthn.Credential{
		ID: []byte("bound"), PublicKey: []byte("k"),
		Flags: webauthn.CredentialFlags{UserVerified: true},
	}
	for _, c := range []*webauthn.Credential{synced, bound} {
		if err := svc.store.AddPasskey(ctx, account.ID, c, string(c.ID)); err != nil {
			t.Fatalf("AddPasskey(%s): %v", c.ID, err)
		}
	}

	keys, err := svc.store.ListPasskeys(ctx, account.ID)
	if err != nil || len(keys) != 2 {
		t.Fatalf("ListPasskeys: %v, %d keys", err, len(keys))
	}
	got := map[string]bool{}
	for _, k := range keys {
		got[k.Name] = k.BackupEligible
	}
	if !got["synced"] {
		t.Error("a syncable credential was not recorded as syncable")
	}
	if got["bound"] {
		t.Error("a device-bound credential was recorded as syncable")
	}
}

func TestRegisterBeginNeedsAuthorisation(t *testing.T) {
	svc, _ := passkeyService(t)

	_, err := registerBegin(svc, "secret")
	if got := status(t, err); got != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", got)
	}
}

// A session alone is not enough: a stolen cookie must not be able to attach a
// credential that outlives it.
func TestRegisterBeginNeedsThePassword(t *testing.T) {
	svc, ctx := passkeyService(t)
	_, cookie := signedInAccount(t, svc, ctx, "matt", "secret")

	_, err := registerBegin(svc, "", cookie)
	if got := status(t, err); got != http.StatusUnauthorized {
		t.Errorf("with no password: got %d, want 401", got)
	}

	_, err = registerBegin(svc, "not-the-password", cookie)
	if got := status(t, err); got != http.StatusUnauthorized {
		t.Errorf("with a wrong password: got %d, want 401", got)
	}
}

func TestRegisterBeginAcceptsSessionAndPassword(t *testing.T) {
	svc, ctx := passkeyService(t)
	_, cookie := signedInAccount(t, svc, ctx, "matt", "secret")

	w, err := registerBegin(svc, "secret", cookie)
	if err != nil {
		t.Fatalf("RegisterBegin: %v", err)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", w.Code)
	}
	// The challenge id travels in a cookie, not in the JSON, so a script cannot
	// move one ceremony's challenge into another's completion.
	var ceremony *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == svc.cfg.CeremonyCookie {
			ceremony = c
		}
	}
	if ceremony == nil || ceremony.Value == "" {
		t.Fatal("no ceremony cookie was set")
	}
	if !ceremony.HttpOnly {
		t.Error("the ceremony cookie is not HttpOnly")
	}
}

// Finishing must refuse a challenge that never carried the password check —
// otherwise the step-up could be skipped by calling finish directly.
func TestRegisterFinishRefusesAChallengeWithoutReauth(t *testing.T) {
	svc, ctx := passkeyService(t)
	account := mustAccount(t, svc.store, ctx, "matt")

	id, err := svc.store.SaveChallenge(ctx, Challenge{
		Purpose:   PurposeRegister,
		AccountID: &account.ID,
		Data:      &webauthn.SessionData{Challenge: "abc"},
		Reauthed:  false,
	}, time.Minute)
	if err != nil {
		t.Fatalf("SaveChallenge: %v", err)
	}

	r := httptest.NewRequest(http.MethodPost, "/auth/passkey/register/finish", strings.NewReader("{}"))
	r.AddCookie(&http.Cookie{Name: svc.cfg.CeremonyCookie, Value: id})
	w := httptest.NewRecorder()

	if got := status(t, svc.RegisterFinish(w, r)); got != http.StatusForbidden {
		t.Errorf("got %d, want 403", got)
	}
}

// A login challenge must not complete a registration.
func TestRegisterFinishRefusesALoginChallenge(t *testing.T) {
	svc, ctx := passkeyService(t)

	id, err := svc.store.SaveChallenge(ctx, Challenge{
		Purpose: PurposeLogin,
		Data:    &webauthn.SessionData{Challenge: "abc"},
	}, time.Minute)
	if err != nil {
		t.Fatalf("SaveChallenge: %v", err)
	}

	r := httptest.NewRequest(http.MethodPost, "/auth/passkey/register/finish", strings.NewReader("{}"))
	r.AddCookie(&http.Cookie{Name: svc.cfg.CeremonyCookie, Value: id})
	w := httptest.NewRecorder()

	if got := status(t, svc.RegisterFinish(w, r)); got != http.StatusBadRequest {
		t.Errorf("got %d, want 400", got)
	}
}

// Opening the link validates it but does not spend it — a mail client
// prefetching the URL must not cost someone their one enrollment.
func TestEnrollDoesNotConsumeTheToken(t *testing.T) {
	svc, ctx := passkeyService(t)
	account := mustAccount(t, svc.store, ctx, "matt")
	token, err := svc.store.CreateEnrollToken(ctx, account.ID, time.Hour)
	if err != nil {
		t.Fatalf("CreateEnrollToken: %v", err)
	}

	r := httptest.NewRequest(http.MethodGet, EnrollPath+"?t="+url.QueryEscape(token), nil)
	w := httptest.NewRecorder()
	if err := svc.Enroll(w, r); err != nil {
		t.Fatalf("Enroll: %v", err)
	}
	if w.Code != http.StatusSeeOther {
		t.Errorf("got %d, want a 303 that takes the token out of the URL", w.Code)
	}

	var enroll *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == enrollCookie {
			enroll = c
		}
	}
	if enroll == nil || enroll.Value != token {
		t.Fatal("the token was not moved into a cookie")
	}
	if _, err := svc.store.ConsumeEnrollToken(ctx, token); err != nil {
		t.Errorf("the token was spent merely by opening the link: %v", err)
	}
}

// An abandoned ceremony must not burn the link either: it is spent only once a
// key actually exists.
func TestAbandonedRegistrationKeepsTheEnrollmentToken(t *testing.T) {
	svc, ctx := passkeyService(t)
	account := mustAccount(t, svc.store, ctx, "matt")
	hash, _ := password.Hash("secret")
	if err := svc.store.SetPasswordHash(ctx, account.ID, hash); err != nil {
		t.Fatalf("SetPasswordHash: %v", err)
	}
	token, err := svc.store.CreateEnrollToken(ctx, account.ID, time.Hour)
	if err != nil {
		t.Fatalf("CreateEnrollToken: %v", err)
	}

	// Start a ceremony with the enrollment cookie, then walk away.
	if _, err := registerBegin(svc, "secret", &http.Cookie{Name: enrollCookie, Value: token}); err != nil {
		t.Fatalf("RegisterBegin: %v", err)
	}

	if _, err := svc.store.ConsumeEnrollToken(ctx, token); err != nil {
		t.Errorf("an abandoned ceremony consumed the enrollment link: %v", err)
	}
}

// Removing a key is as sensitive as adding one, so it asks for the password too.
func TestDeletePasskeyNeedsThePassword(t *testing.T) {
	svc, ctx := passkeyService(t)
	account, cookie := signedInAccount(t, svc, ctx, "matt", "secret")
	if err := svc.store.AddPasskey(ctx, account.ID,
		&webauthn.Credential{ID: []byte("cred"), PublicKey: []byte("k")}, "key"); err != nil {
		t.Fatalf("AddPasskey: %v", err)
	}

	form := url.Values{"password": {"wrong"}}
	r := httptest.NewRequest(http.MethodDelete, "/auth/passkey/x", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.RemoteAddr = "10.0.0.1:1234"
	r.AddCookie(cookie)
	w := httptest.NewRecorder()

	if got := status(t, svc.DeletePasskey(w, r)); got != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", got)
	}
	if keys, _ := svc.store.ListPasskeys(ctx, account.ID); len(keys) != 1 {
		t.Error("the passkey was removed without the password")
	}
}

// With no relying party configured the ceremonies are unavailable, and say so,
// rather than failing in some other way. Password sign-in is unaffected.
func TestPasskeyEndpointsAreUnavailableWithoutAnRPID(t *testing.T) {
	svc, _, _ := testService(t) // built without RPID

	for name, h := range map[string]func(http.ResponseWriter, *http.Request) error{
		"login/begin":     svc.LoginBegin,
		"login/finish":    svc.LoginFinish,
		"register/begin":  svc.RegisterBegin,
		"register/finish": svc.RegisterFinish,
	} {
		r := httptest.NewRequest(http.MethodPost, "/auth/passkey/"+name, nil)
		w := httptest.NewRecorder()
		if got := status(t, h(w, r)); got != http.StatusNotImplemented {
			t.Errorf("%s: got %d, want 501", name, got)
		}
	}
}
