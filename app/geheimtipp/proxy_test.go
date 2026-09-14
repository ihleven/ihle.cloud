package geheimtipp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The pool's backend sets its session cookie for "/" — right on its own site,
// wrong here, where "/" is the family app.
func TestProxyScopesCookiesToItsMountPoint(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Set-Cookie", "token=abc; Path=/; Domain=ihleven.de; HttpOnly")
		w.WriteHeader(http.StatusOK)
	}))
	defer up.Close()

	proxy, err := Proxy(up.URL, "/ght", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	if err := proxy(w, httptest.NewRequest(http.MethodGet, "/ght/aktuell", nil)); err != nil {
		t.Fatal(err)
	}

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("got %d cookies, want 1", len(cookies))
	}
	if cookies[0].Path != "/ght" {
		t.Errorf("cookie path = %q, want /ght", cookies[0].Path)
	}
	// A cookie for the upstream's domain is simply dropped by a browser that
	// believes it is talking to us.
	if cookies[0].Domain != "" {
		t.Errorf("cookie domain = %q, want none", cookies[0].Domain)
	}
	if !cookies[0].HttpOnly {
		t.Error("HttpOnly was lost in rewriting")
	}
	if raw := w.Header().Get("Set-Cookie"); strings.Count(strings.ToLower(raw), "path=") != 1 {
		t.Errorf("Set-Cookie = %q, want exactly one Path", raw)
	}
}
