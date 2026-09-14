package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ihleven/ihlvn/pkg/password"
	"github.com/interhome-group/cms/pkg/errs"
)

// Changing a password needs a screening that does not reach the public breach
// service: a test that depends on the network fails for reasons that have
// nothing to do with it. An empty range response means "not in the corpus",
// which is the uninteresting answer and therefore the right default.
func cleanScreening(t *testing.T, svc *Service) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	svc.breaches = &password.BreachChecker{BaseURL: server.URL + "/"}
	svc.cfg.MinPasswordLength = 4
}

func changePassword(svc *Service, cookie *http.Cookie, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, "/auth/password", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	if err := svc.ChangePassword(w, r); err != nil {
		// Handlers report failure as an error and the router's middleware turns
		// it into a status. The test wants the status, so it does the same.
		w.Code = errs.StatusOf(err)
	}
	return w
}

// The point of the endpoint: someone can replace the password they arrived
// with. Without it an adopted pool player is stuck with one chosen years ago
// and only an administrator could change it.
func TestChangingAPasswordReplacesIt(t *testing.T) {
	svc, _, ctx := testService(t)
	cleanScreening(t, svc)
	_, cookie := signedInAccount(t, svc, ctx, "paul", "spacedog")

	w := changePassword(svc, cookie, `{"current":"spacedog","password":"a-better-one"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200: %s", w.Code, w.Body)
	}
	var result PasswordResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !result.Set {
		t.Fatalf("password was not set: %+v", result.Advice)
	}

	if _, err := svc.authenticatePassword(ctx, "paul", "a-better-one"); err != nil {
		t.Errorf("the new password does not work: %v", err)
	}
	if _, err := svc.authenticatePassword(ctx, "paul", "spacedog"); err == nil {
		t.Error("the old password still works")
	}
}

// The session proves who is asking, not that they are still at the keyboard. A
// browser left open would otherwise be enough to take an account over for good.
func TestChangingAPasswordNeedsTheCurrentOne(t *testing.T) {
	svc, _, ctx := testService(t)
	cleanScreening(t, svc)
	_, cookie := signedInAccount(t, svc, ctx, "paul", "spacedog")

	w := changePassword(svc, cookie, `{"current":"guessing","password":"a-better-one"}`)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", w.Code)
	}
	if _, err := svc.authenticatePassword(ctx, "paul", "spacedog"); err != nil {
		t.Error("the password was changed anyway")
	}
}

func TestChangingAPasswordNeedsASession(t *testing.T) {
	svc, _, _ := testService(t)
	cleanScreening(t, svc)

	if w := changePassword(svc, nil, `{"current":"x","password":"y"}`); w.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", w.Code)
	}
}

func TestChangingAPasswordRefusesAnEmptyOne(t *testing.T) {
	svc, _, ctx := testService(t)
	cleanScreening(t, svc)
	_, cookie := signedInAccount(t, svc, ctx, "paul", "spacedog")

	if w := changePassword(svc, cookie, `{"current":"spacedog","password":""}`); w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", w.Code)
	}
	if _, err := svc.authenticatePassword(ctx, "paul", "spacedog"); err != nil {
		t.Error("the password was cleared")
	}
}

// The screening is advice, not a verdict — the person choosing knows things the
// check does not — so a concerning password is put to them rather than refused,
// and going ahead is a second, deliberate act.
func TestAConcerningPasswordIsPutToThePersonFirst(t *testing.T) {
	svc, _, ctx := testService(t)
	cleanScreening(t, svc)
	svc.cfg.MinPasswordLength = 12
	_, cookie := signedInAccount(t, svc, ctx, "paul", "spacedog")

	w := changePassword(svc, cookie, `{"current":"spacedog","password":"short"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200: %s", w.Code, w.Body)
	}
	var result PasswordResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Set {
		t.Fatal("a short password was set without being put to the person")
	}
	if !result.Advice.TooShort {
		t.Errorf("advice does not say why: %+v", result.Advice)
	}
	if _, err := svc.authenticatePassword(ctx, "paul", "spacedog"); err != nil {
		t.Error("the password changed despite not being set")
	}

	// Asked again, with the answer.
	w = changePassword(svc, cookie, `{"current":"spacedog","password":"short","confirm":true}`)
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !result.Set {
		t.Fatal("confirming did not set it")
	}
	if _, err := svc.authenticatePassword(ctx, "paul", "short"); err != nil {
		t.Errorf("the confirmed password does not work: %v", err)
	}
}

// An account confined to the pool must reach this, which is why the route is a
// bare handler rather than one behind requireAccount — that gate refuses them.
// Being able to replace the password they arrived with is the whole point.
func TestAConfinedAccountCanChangeItsPassword(t *testing.T) {
	svc, store, ctx := testService(t)
	cleanScreening(t, svc)
	_, cookie := signedInAccount(t, svc, ctx, "paul", "spacedog")
	if _, err := store.pool.Exec(ctx,
		`update account set type = $1 where name = 'paul'`, TypeGeheimtipp); err != nil {
		t.Fatal(err)
	}

	w := changePassword(svc, cookie, `{"current":"spacedog","password":"a-better-one"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200: %s", w.Code, w.Body)
	}
	if _, err := svc.authenticatePassword(ctx, "paul", "a-better-one"); err != nil {
		t.Errorf("a confined account could not change its password: %v", err)
	}
}
