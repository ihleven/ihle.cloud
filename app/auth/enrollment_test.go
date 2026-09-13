package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ihleven/ihlvn/pkg/password"
)

// Attaching a credential costs two things: the link or session that authorises
// the ceremony, and the account's password. Neither substitutes for the other.
//
// That is the property these tests exist to hold on to. It was briefly traded
// away — an account with no password was allowed to choose one while enrolling,
// which made a link a single factor — and these are the tests that would have
// made that visible.
//
// They exercise reauthenticate rather than a whole ceremony, because the WebAuthn
// half needs an authenticator and the decision being made here is reached before
// any of that.

// enrollRequest is a registration authorised by a link, carrying a password.
func enrollRequest(token, plain string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/auth/passkey/register/begin",
		strings.NewReader("password="+plain))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if token != "" {
		r.AddCookie(&http.Cookie{Name: enrollCookie, Value: token})
	}
	return r
}

// An account with no password cannot enrol, however it was authorised. Letting
// it would make the link a single factor: whoever intercepted one could attach
// their own passkey and keep the account.
func TestAnAccountWithoutAPasswordCannotEnrol(t *testing.T) {
	svc, store := ceremonyService(t)
	ctx := context.Background()

	acc := mustAccount(t, store, ctx, "wolfgang")
	if acc.HasPassword() {
		t.Fatal("a new account already has a password")
	}
	token, err := store.CreateEnrollToken(ctx, acc.ID, time.Minute)
	if err != nil {
		t.Fatalf("CreateEnrollToken: %v", err)
	}

	registrant, err := svc.registrant(enrollRequest(token, "a password it does not have"))
	if err != nil {
		t.Fatalf("registrant: %v", err)
	}
	if err := svc.reauthenticate(enrollRequest(token, "a password it does not have"), registrant); err == nil {
		t.Fatal("an account with no password was allowed to enrol")
	}

	// And nothing was set on the way past: the password offered is not adopted.
	stored, err := store.AccountByName(ctx, "wolfgang")
	if err != nil {
		t.Fatalf("reloading: %v", err)
	}
	if stored.HasPassword() {
		t.Error("the offered password became the account's own")
	}
}

// The same refusal, one layer up: a link is not issued for an account that could
// not use it, because issuing one would look like it had been made usable.
func TestNoLinkForAnAccountWithoutAPassword(t *testing.T) {
	admin, store := testAdmin(t)
	ctx := actingAs(account(t, store, "admin", "*"))
	account(t, store, "wolfgang")

	_, err := admin.IssueEnrollment(ctx, "wolfgang")
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("IssueEnrollment = %v, want ErrConflict", err)
	}
	if !strings.Contains(err.Error(), "no password yet") {
		t.Errorf("error = %q, want it to say what is missing", err.Error())
	}
}

// Once an account has a password, a link is issued — and the password is not in
// it, because the two have to travel separately to be two factors.
func TestALinkIsIssuedOnceThereIsAPassword(t *testing.T) {
	admin, store := testAdmin(t)
	ctx := actingAs(account(t, store, "admin", "*"))
	target := account(t, store, "wolfgang")

	hash, err := password.Hash("a chosen passphrase")
	if err != nil {
		t.Fatalf("hashing: %v", err)
	}
	if err := store.SetPasswordHash(ctx, target.ID, hash); err != nil {
		t.Fatalf("SetPasswordHash: %v", err)
	}

	link, err := admin.IssueEnrollment(ctx, "wolfgang")
	if err != nil {
		t.Fatalf("IssueEnrollment: %v", err)
	}
	if !strings.Contains(link.URL, EnrollPath) {
		t.Errorf("URL = %q, want the enrollment route", link.URL)
	}
	if strings.Contains(link.URL, "passphrase") {
		t.Error("the password travelled in the link")
	}
	if link.ExpiresAt.Before(time.Now()) {
		t.Error("the link is already expired")
	}
}

// The rule that matters: an account that has a password proves it, however the
// request was authorised. Otherwise a stolen link or session would be enough to
// attach a credential and keep access indefinitely.
func TestAnAccountWithAPasswordMustProveIt(t *testing.T) {
	svc, store := ceremonyService(t)
	ctx := context.Background()

	acc := mustAccount(t, store, ctx, "matt")
	hash, err := password.Hash("the real password")
	if err != nil {
		t.Fatalf("hashing: %v", err)
	}
	if err := store.SetPasswordHash(ctx, acc.ID, hash); err != nil {
		t.Fatalf("SetPasswordHash: %v", err)
	}
	token, err := store.CreateEnrollToken(ctx, acc.ID, time.Minute)
	if err != nil {
		t.Fatalf("CreateEnrollToken: %v", err)
	}

	registrant, err := svc.registrant(enrollRequest(token, "the wrong password"))
	if err != nil {
		t.Fatalf("registrant: %v", err)
	}
	if err := svc.reauthenticate(enrollRequest(token, "the wrong password"), registrant); err == nil {
		t.Fatal("a wrong password was accepted because a link was held")
	}

	// And the password was not quietly replaced by the one just offered.
	stored, _ := store.AccountByName(ctx, "matt")
	ok, _, _ := password.Verify(stored.passwordHash, "the real password")
	if !ok {
		t.Error("the account's password was overwritten by the failed attempt")
	}
}
