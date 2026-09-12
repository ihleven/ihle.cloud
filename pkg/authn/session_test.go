package authn

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func testService(t *testing.T) (*Service, *Store, context.Context) {
	t.Helper()
	store, ctx := testStore(t)
	svc, err := New(store, Config{
		CookieName: "session",
		Issuer:     "ihlvn",
		Audience:   "famihlie",
		SessionTTL: time.Hour,
		SlideAfter: time.Minute,
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return svc, store, ctx
}

// signedInAccount creates an account with a password and returns it with a
// request cookie that authenticates as it.
func signedInAccount(t *testing.T, svc *Service, ctx context.Context, name, password string) (*Account, *http.Cookie) {
	t.Helper()
	a := mustAccount(t, svc.store, ctx, name)
	hash, err := hashPassword(password)
	if err != nil {
		t.Fatalf("hashPassword: %v", err)
	}
	if err := svc.store.SetPasswordHash(ctx, a.ID, hash); err != nil {
		t.Fatalf("SetPasswordHash: %v", err)
	}
	token, err := svc.store.CreateSession(ctx, a.ID, time.Hour)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	return a, &http.Cookie{Name: "session", Value: token}
}

func postLogin(svc *Service, username, password string) *httptest.ResponseRecorder {
	form := url.Values{"username": {username}, "password": {password}}
	r := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.RemoteAddr = "10.0.0.1:1234"
	w := httptest.NewRecorder()
	if err := svc.Login(w, r); err != nil {
		if se, ok := err.(*statusError); ok {
			http.Error(w, se.message, se.status)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	return w
}

// The response keeps the field names the frontend was written against, so the
// mechanism could change without touching the UI.
func TestLoginReturnsTheClaimShapeTheFrontendReads(t *testing.T) {
	svc, _, ctx := testService(t)
	a := mustAccount(t, svc.store, ctx, "matt")
	hash, _ := hashPassword("secret")
	if err := svc.store.SetPasswordHash(ctx, a.ID, hash); err != nil {
		t.Fatalf("SetPasswordHash: %v", err)
	}
	if err := svc.store.SetCMSProfile(ctx, a.ID, CMSProfile{Permissions: []string{"entry.create"}}); err != nil {
		t.Fatalf("SetCMSProfile: %v", err)
	}

	w := postLogin(svc, "matt", "secret")
	if w.Code != http.StatusOK {
		t.Fatalf("login returned %d: %s", w.Code, w.Body.String())
	}

	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decoding the response: %v", err)
	}
	for _, field := range []string{"iss", "sub", "aud", "exp", "nbf", "iat", "permissions", "name", "email"} {
		if _, ok := got[field]; !ok {
			t.Errorf("response is missing %q, which the frontend reads", field)
		}
	}
	if got["sub"] != "matt" {
		t.Errorf("sub = %v, want the login name", got["sub"])
	}
	if exp, _ := got["exp"].(float64); exp <= float64(time.Now().Unix()) {
		t.Errorf("exp = %v, which is not in the future", got["exp"])
	}
	perms, _ := got["permissions"].(map[string]any)
	if _, ok := perms["entry.create"]; !ok {
		t.Errorf("permissions = %v, want the account's cms permissions", got["permissions"])
	}

	var set *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "session" {
			set = c
		}
	}
	if set == nil || set.Value == "" {
		t.Fatal("no session cookie was set")
	}
	if !set.HttpOnly {
		t.Error("the session cookie is not HttpOnly")
	}
}

// An unknown account and a wrong password must be indistinguishable, or the
// endpoint tells an attacker which accounts exist.
func TestLoginFailuresAreIndistinguishable(t *testing.T) {
	svc, _, ctx := testService(t)
	a := mustAccount(t, svc.store, ctx, "matt")
	hash, _ := hashPassword("secret")
	if err := svc.store.SetPasswordHash(ctx, a.ID, hash); err != nil {
		t.Fatalf("SetPasswordHash: %v", err)
	}

	wrong := postLogin(svc, "matt", "not-the-password")
	unknown := postLogin(svc, "nobody", "not-the-password")

	if wrong.Code != http.StatusUnauthorized || unknown.Code != http.StatusUnauthorized {
		t.Fatalf("statuses: wrong=%d unknown=%d, want 401 for both", wrong.Code, unknown.Code)
	}
	if wrong.Body.String() != unknown.Body.String() {
		t.Errorf("responses differ:\n wrong password: %q\n unknown account: %q",
			wrong.Body.String(), unknown.Body.String())
	}
	for _, c := range wrong.Result().Cookies() {
		if c.Name == "session" && c.Value != "" {
			t.Error("a failed sign-in set a session cookie")
		}
	}
}

// An account whose only credential is a passkey cannot be signed into with an
// empty password. The anonymous account is in exactly that position.
func TestLoginRefusesAnAccountWithNoPassword(t *testing.T) {
	svc, _, ctx := testService(t)
	mustAccount(t, svc.store, ctx, "anonymous")

	if w := postLogin(svc, "anonymous", ""); w.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", w.Code)
	}
	if w := postLogin(svc, "anonymous", "guess"); w.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", w.Code)
	}
}

func TestLoginThrottlesRepeatedFailures(t *testing.T) {
	store, ctx := testStore(t)
	svc, err := New(store, Config{
		CookieName: "session", SessionTTL: time.Hour, SlideAfter: time.Minute,
		FailureLimit: 3, FailureWindow: time.Minute,
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	a := mustAccount(t, store, ctx, "matt")
	hash, _ := hashPassword("secret")
	if err := store.SetPasswordHash(ctx, a.ID, hash); err != nil {
		t.Fatalf("SetPasswordHash: %v", err)
	}

	for i := range 3 {
		if w := postLogin(svc, "matt", "wrong"); w.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d returned %d, want 401", i, w.Code)
		}
	}
	// Even the right password is refused once the limit is reached, or the
	// throttle would only slow down the attacker's wrong guesses.
	if w := postLogin(svc, "matt", "secret"); w.Code != http.StatusTooManyRequests {
		t.Errorf("got %d after the limit, want 429", w.Code)
	}
}

// The bcrypt hashes carried over from the old schema are replaced the first
// time their owner signs in.
func TestLoginUpgradesALegacyBcryptHash(t *testing.T) {
	svc, store, ctx := testService(t)
	a := mustAccount(t, store, ctx, "matt")

	legacy, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("generating a bcrypt hash: %v", err)
	}
	if err := store.SetPasswordHash(ctx, a.ID, string(legacy)); err != nil {
		t.Fatalf("SetPasswordHash: %v", err)
	}

	if w := postLogin(svc, "matt", "secret"); w.Code != http.StatusOK {
		t.Fatalf("a bcrypt account could not sign in: %d %s", w.Code, w.Body.String())
	}

	var stored string
	if err := store.pool.QueryRow(ctx, `select password_hash from account where id = $1`, a.ID).Scan(&stored); err != nil {
		t.Fatalf("reading the stored hash: %v", err)
	}
	if !strings.HasPrefix(stored, "$argon2id$") {
		t.Errorf("hash was not upgraded, still %q", stored[:min(10, len(stored))])
	}
	// And the upgraded hash must still accept the same password.
	if w := postLogin(svc, "matt", "secret"); w.Code != http.StatusOK {
		t.Errorf("sign-in broke after the hash was upgraded: %d", w.Code)
	}
}

// Logout ends the session in the database, so a cookie copied beforehand is
// already useless — which a JWT could not offer.
func TestLogoutRevokesServerSide(t *testing.T) {
	svc, _, ctx := testService(t)
	_, cookie := signedInAccount(t, svc, ctx, "matt", "secret")

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(cookie)
	if _, ok := svc.Authenticate(r); !ok {
		t.Fatal("the cookie did not authenticate before logout")
	}

	w := httptest.NewRecorder()
	lr := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	lr.AddCookie(cookie)
	if err := svc.Logout(w, lr); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	after := httptest.NewRequest(http.MethodGet, "/", nil)
	after.AddCookie(cookie)
	if _, ok := svc.Authenticate(after); ok {
		t.Error("the cookie still authenticates after logout")
	}
}

func TestRequireGatesAndRedirects(t *testing.T) {
	svc, _, ctx := testService(t)
	account, cookie := signedInAccount(t, svc, ctx, "matt", "secret")

	var seen *Account
	handler := svc.RequireFunc(func(w http.ResponseWriter, r *http.Request) {
		seen, _ = FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	t.Run("a browser is sent to sign in", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/entries/x", nil)
		r.Header.Set("Sec-Fetch-Dest", "document")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusSeeOther {
			t.Fatalf("got %d, want 303", w.Code)
		}
		if got := w.Header().Get("Location"); got != "/login?next=%2Fentries%2Fx" {
			t.Errorf("Location = %q", got)
		}
	})

	t.Run("a ranged request gets 401, not a redirect", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/super8/clip.mp4", nil)
		r.Header.Set("Range", "bytes=0-1023")
		r.Header.Set("Accept", "text/html")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want 401 — a player must not be handed an HTML page", w.Code)
		}
	})

	t.Run("an API call gets 401", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, "/api/v1/entries/x", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want 401", w.Code)
		}
	})

	t.Run("a signed-in request reaches the handler with its account", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/entries/x", nil)
		r.AddCookie(cookie)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("got %d, want 200", w.Code)
		}
		if seen == nil || seen.ID != account.ID {
			t.Errorf("the handler did not receive the account: %+v", seen)
		}
	})
}

// Optional is what the public pages use: anyone may read, and a signed-in
// reader is identified.
func TestOptionalAllowsAnonymous(t *testing.T) {
	svc, _, ctx := testService(t)
	_, cookie := signedInAccount(t, svc, ctx, "matt", "secret")

	var seen *Account
	var had bool
	handler := svc.OptionalFunc(func(w http.ResponseWriter, r *http.Request) {
		seen, had = FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusOK {
		t.Errorf("an anonymous request got %d, want 200", w.Code)
	}
	if had {
		t.Error("an anonymous request carried an account")
	}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(cookie)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if !had || seen == nil {
		t.Error("a signed-in request did not carry its account")
	}
}

// The redirect target must stay on this site.
func TestDeniedRedirectStaysOnSite(t *testing.T) {
	svc, _, _ := testService(t)
	handler := svc.RequireFunc(func(w http.ResponseWriter, r *http.Request) {})

	r := httptest.NewRequest(http.MethodGet, "//evil.example/", nil)
	r.Header.Set("Sec-Fetch-Dest", "document")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if got := w.Header().Get("Location"); got != "/login" {
		t.Errorf("Location = %q, want a bare /login", got)
	}
}

// A session belonging to a disabled account stops working at once, without
// waiting for the cookie to expire.
func TestDisablingAnAccountEndsItsSession(t *testing.T) {
	svc, store, ctx := testService(t)
	_, cookie := signedInAccount(t, svc, ctx, "matt", "secret")

	if _, err := store.SetDisabled(ctx, "matt", true); err != nil {
		t.Fatalf("SetDisabled: %v", err)
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(cookie)
	if _, ok := svc.Authenticate(r); ok {
		t.Error("a disabled account still authenticates")
	}
}

// The SPA posts the sign-in form as FormData, so the body is multipart. Parsing
// it with ParseForm leaves PostForm non-nil but empty, which silently blanks
// every field — the credentials then look wrong rather than unread.
func TestLoginAcceptsAMultipartForm(t *testing.T) {
	svc, _, ctx := testService(t)
	a := mustAccount(t, svc.store, ctx, "matt")
	// A bcrypt hash and a multipart body together: the exact shape of a real
	// sign-in against accounts carried over from the old schema.
	legacy, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("generating a bcrypt hash: %v", err)
	}
	if err := svc.store.SetPasswordHash(ctx, a.ID, string(legacy)); err != nil {
		t.Fatalf("SetPasswordHash: %v", err)
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	if err := mw.WriteField("username", "matt"); err != nil {
		t.Fatalf("WriteField: %v", err)
	}
	if err := mw.WriteField("password", "secret"); err != nil {
		t.Fatalf("WriteField: %v", err)
	}
	mw.Close()

	r := httptest.NewRequest(http.MethodPost, "/auth/login", &body)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	r.RemoteAddr = "10.0.0.1:1234"
	w := httptest.NewRecorder()

	if err := svc.Login(w, r); err != nil {
		t.Fatalf("a multipart sign-in was refused: %v", err)
	}
	if w.Code != http.StatusOK {
		t.Errorf("got %d, want 200", w.Code)
	}
}

// Secure is configurable because Safari drops a Secure cookie over plain http,
// including on localhost — sign-in then fails with no cookie and no explanation.
func TestCookieSecureFollowsConfig(t *testing.T) {
	store, _ := testStore(t)

	for _, secure := range []bool{true, false} {
		svc, err := New(store, Config{
			CookieName: "session", CookieSecure: secure,
			SessionTTL: time.Hour, SlideAfter: time.Minute,
		}, slog.New(slog.NewTextHandler(io.Discard, nil)))
		if err != nil {
			t.Fatalf("New: %v", err)
		}

		w := httptest.NewRecorder()
		svc.setSessionCookie(w, "token")

		cookies := w.Result().Cookies()
		if len(cookies) != 1 {
			t.Fatalf("expected one cookie, got %d", len(cookies))
		}
		if cookies[0].Secure != secure {
			t.Errorf("CookieSecure=%v produced Secure=%v", secure, cookies[0].Secure)
		}
		// HttpOnly is not a preference: the token is the whole credential.
		if !cookies[0].HttpOnly {
			t.Error("the session cookie is not HttpOnly")
		}
	}
}
