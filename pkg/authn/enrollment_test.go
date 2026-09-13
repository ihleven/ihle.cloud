package authn

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Enrolling always costs a password. What changes with the account is which
// password: one it already has, which it proves, or its first, which it chooses.
//
// These tests are about authorizeRegistration rather than a whole ceremony,
// because the WebAuthn half needs an authenticator and the decision being made
// here is reached before any of that.

// enrollRequest is a registration authorised by a link, carrying a password.
func enrollRequest(token, password string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/auth/passkey/register/begin",
		strings.NewReader("password="+password))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if token != "" {
		r.AddCookie(&http.Cookie{Name: enrollCookie, Value: token})
	}
	return r
}

// An account created in the admin section has no password, so requiring one
// before it may enrol would leave it unable to obtain either credential. The
// link is what stands in: single use, expiring, issued by an administrator.
func TestAnEnrollmentLinkSetsTheFirstPassword(t *testing.T) {
	svc, store := ceremonyService(t)
	ctx := context.Background()

	account := mustAccount(t, store, ctx, "wolfgang")
	if account.HasPassword() {
		t.Fatal("a new account already has a password")
	}
	token, err := store.CreateEnrollToken(ctx, account.ID, time.Minute)
	if err != nil {
		t.Fatalf("CreateEnrollToken: %v", err)
	}

	registrant, viaLink, err := svc.registrant(enrollRequest(token, "a chosen passphrase"))
	if err != nil {
		t.Fatalf("registrant: %v", err)
	}
	if !viaLink {
		t.Fatal("a request carrying only an enrollment link was not recognised as one")
	}
	if err := svc.authorizeRegistration(enrollRequest(token, "a chosen passphrase"), registrant, viaLink); err != nil {
		t.Fatalf("authorizeRegistration: %v", err)
	}

	// Set, and set for real: the account can now sign in with it.
	stored, err := store.AccountByName(ctx, "wolfgang")
	if err != nil {
		t.Fatalf("reloading: %v", err)
	}
	if !stored.HasPassword() {
		t.Fatal("the chosen password was not stored")
	}
	ok, _, err := verifyPassword(stored.passwordHash, "a chosen passphrase")
	if err != nil || !ok {
		t.Errorf("the stored password does not verify: ok=%v err=%v", ok, err)
	}
	// And the copy the rest of the ceremony reads was loaded before it existed.
	if !registrant.HasPassword() {
		t.Error("the in-flight account still reports having no password")
	}
}

// Choosing nothing is not choosing a password.
func TestEnrollmentRefusesAnEmptyFirstPassword(t *testing.T) {
	svc, store := ceremonyService(t)
	ctx := context.Background()

	account := mustAccount(t, store, ctx, "wolfgang")
	token, err := store.CreateEnrollToken(ctx, account.ID, time.Minute)
	if err != nil {
		t.Fatalf("CreateEnrollToken: %v", err)
	}

	r := enrollRequest(token, "")
	registrant, viaLink, err := svc.registrant(r)
	if err != nil {
		t.Fatalf("registrant: %v", err)
	}
	if err := svc.authorizeRegistration(enrollRequest(token, ""), registrant, viaLink); err == nil {
		t.Fatal("an empty password was accepted as an account's first")
	}

	stored, _ := store.AccountByName(ctx, "wolfgang")
	if stored.HasPassword() {
		t.Error("an empty password was stored anyway")
	}
}

// The rule that matters is unchanged: an account that has a password proves it,
// however the request was authorised. Otherwise a stolen link or session would
// be enough to attach a credential and keep access indefinitely.
func TestAnAccountWithAPasswordMustStillProveIt(t *testing.T) {
	svc, store := ceremonyService(t)
	ctx := context.Background()

	account := mustAccount(t, store, ctx, "matt")
	hash, err := hashPassword("the real password")
	if err != nil {
		t.Fatalf("hashPassword: %v", err)
	}
	if err := store.SetPasswordHash(ctx, account.ID, hash); err != nil {
		t.Fatalf("SetPasswordHash: %v", err)
	}
	token, err := store.CreateEnrollToken(ctx, account.ID, time.Minute)
	if err != nil {
		t.Fatalf("CreateEnrollToken: %v", err)
	}

	registrant, viaLink, err := svc.registrant(enrollRequest(token, "the wrong password"))
	if err != nil {
		t.Fatalf("registrant: %v", err)
	}
	err = svc.authorizeRegistration(enrollRequest(token, "the wrong password"), registrant, viaLink)
	if err == nil {
		t.Fatal("a wrong password was accepted because a link was held")
	}

	// And the password was not quietly replaced by the one just offered.
	stored, _ := store.AccountByName(ctx, "matt")
	ok, _, _ := verifyPassword(stored.passwordHash, "the real password")
	if !ok {
		t.Error("the account's password was overwritten by the failed attempt")
	}
}

// A session is not accepted for setting a first password. A session for an
// account with no password can only have come from a passkey, and that path had
// an administrator in it already — so there is no case this refuses that a link
// does not cover.
func TestASessionCannotSetAFirstPassword(t *testing.T) {
	svc, store := ceremonyService(t)
	ctx := context.Background()

	account := mustAccount(t, store, ctx, "wolfgang")

	r := httptest.NewRequest(http.MethodPost, "/auth/passkey/register/begin",
		strings.NewReader("password=a chosen passphrase"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err := svc.authorizeRegistration(r, account, false); err == nil {
		t.Fatal("a first password was set without an enrollment link")
	}

	stored, _ := store.AccountByName(ctx, "wolfgang")
	if stored.HasPassword() {
		t.Error("a password was stored anyway")
	}
}
