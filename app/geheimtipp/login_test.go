package geheimtipp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// upstream stands in for the pool's backend: form in, JSON out, no cookie.
func upstream(t *testing.T, jwt string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login" {
			t.Errorf("upstream path = %q, want /login", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("upstream could not read the form: %v", err)
		}
		if r.PostForm.Get("username") != "paul" || r.PostForm.Get("password") != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"Username": "paul", "JWT": jwt})
	}))
}

func post(t *testing.T, h func(http.ResponseWriter, *http.Request) error, form url.Values) (*httptest.ResponseRecorder, error) {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, "/ght/login", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	return w, h(w, r)
}

func TestLoginSetsTheTokenCookie(t *testing.T) {
	up := upstream(t, "jwt.value.here")
	defer up.Close()

	login, err := Login(up.URL, true, "/ght")
	if err != nil {
		t.Fatal(err)
	}

	w, err := post(t, login, url.Values{"username": {"paul"}, "password": {"secret"}})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("got %d cookies, want 1", len(cookies))
	}
	c := cookies[0]
	if c.Name != TokenCookie || c.Value != "jwt.value.here" {
		t.Errorf("cookie = %s=%q, want %s=jwt.value.here", c.Name, c.Value, TokenCookie)
	}
	// Scoped to the proxy: the pool's credential is not offered to the rest of
	// this app, which is a different site sharing an origin.
	if c.Path != "/ght" {
		t.Errorf("cookie path = %q, want /ght", c.Path)
	}
	// Their frontend needed a readable cookie to learn the login name; /aktuell
	// answers that here, so nothing on the page needs to read the credential.
	if !c.HttpOnly || !c.Secure {
		t.Errorf("cookie HttpOnly=%v Secure=%v, want both true", c.HttpOnly, c.Secure)
	}
}

func TestLoginDoesNotHandTheTokenToTheBrowser(t *testing.T) {
	up := upstream(t, "jwt.value.here")
	defer up.Close()

	login, _ := Login(up.URL, false, "/ght")
	w, err := post(t, login, url.Values{"username": {"paul"}, "password": {"secret"}})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	body := w.Body.String()
	if strings.Contains(body, "jwt.value.here") {
		t.Errorf("the response body carries the token: %s", body)
	}
	var answer struct{ Username string }
	if err := json.Unmarshal([]byte(body), &answer); err != nil || answer.Username != "paul" {
		t.Errorf("body = %s, want {\"username\":\"paul\"}", body)
	}
}

func TestLoginRefusesWrongCredentials(t *testing.T) {
	up := upstream(t, "jwt.value.here")
	defer up.Close()

	login, _ := Login(up.URL, false, "/ght")
	w, err := post(t, login, url.Values{"username": {"paul"}, "password": {"wrong"}})
	if err == nil {
		t.Fatal("a wrong password was accepted")
	}
	if len(w.Result().Cookies()) != 0 {
		t.Error("a cookie was set for a refused sign-in")
	}
}

func TestLogoutClearsTheCookie(t *testing.T) {
	w, err := post(t, Logout(false, "/ght"), nil)
	if err != nil {
		t.Fatalf("logout: %v", err)
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("got %d cookies, want 1", len(cookies))
	}
	if cookies[0].Value != "" || cookies[0].MaxAge >= 0 {
		t.Errorf("cookie = %q MaxAge=%d, want empty and expired", cookies[0].Value, cookies[0].MaxAge)
	}
}

func TestLoginNeedsAUsableUpstream(t *testing.T) {
	if _, err := Login("ihleven.de/api/v1", false, "/ght"); err == nil {
		t.Error("an upstream with no scheme was accepted")
	}
}

// A request the pool's host will not serve — it answers 403 to some clients
// regardless of credentials — must not be reported as a wrong password. Saying
// so sends whoever is signing in to check the one thing that is not broken.
func TestARefusedRequestIsNotReportedAsAWrongPassword(t *testing.T) {
	for _, status := range []int{http.StatusForbidden, http.StatusBadGateway, http.StatusNotFound} {
		up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(status)
		}))

		login, _ := Login(up.URL, false, "/ght")
		_, err := post(t, login, url.Values{"username": {"paul"}, "password": {"secret"}})
		if err == nil {
			t.Fatalf("upstream %d was accepted", status)
		}
		if strings.Contains(err.Error(), "wrong login or password") {
			t.Errorf("upstream %d was reported as a wrong password: %v", status, err)
		}
		up.Close()
	}
}

// The pool's host answers Go's default User-Agent with a 403, so the request
// has to name itself.
func TestTheRequestNamesItself(t *testing.T) {
	seen := make(chan string, 1)
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen <- r.Header.Get("User-Agent")
		_ = json.NewEncoder(w).Encode(map[string]string{"Username": "paul", "JWT": "j"})
	}))
	defer up.Close()

	login, _ := Login(up.URL, false, "/ght")
	if _, err := post(t, login, url.Values{"username": {"paul"}, "password": {"secret"}}); err != nil {
		t.Fatal(err)
	}

	if got := <-seen; got != userAgent {
		t.Errorf("User-Agent = %q, want %q", got, userAgent)
	}
}
