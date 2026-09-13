package hidrive

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// The consent redirect is built from configuration rather than from the request,
// because HiDrive matches redirect_uri against a registration: a value taken
// from a proxied request would describe the local hop and be rejected.
func TestConsentRedirectCarriesTheRegistration(t *testing.T) {
	o := NewOAuth("demo-client", "s3cret", "https://example.test/hi/auth/authcode", nil)

	w := httptest.NewRecorder()
	if err := o.Authorize(w, httptest.NewRequest(http.MethodGet, "/?state=/filme", nil)); err != nil {
		t.Fatalf("Authorize: %v", err)
	}
	if w.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want a redirect", w.Code)
	}

	loc := w.Header().Get("Location")
	for _, want := range []string{
		consentURL,
		"client_id=demo-client",
		"state=%2Ffilme",
		"redirect_uri=https%3A%2F%2Fexample.test%2Fhi%2Fauth%2Fauthcode",
	} {
		if !strings.Contains(loc, want) {
			t.Errorf("redirect %q does not carry %q", loc, want)
		}
	}
	// The secret authorises the exchange afterwards; it has no business in a URL
	// the browser is handed.
	if strings.Contains(loc, "s3cret") {
		t.Errorf("the client secret travelled in the redirect: %q", loc)
	}
}

// state comes back from HiDrive untouched and is where the browser is sent, so
// it is the one value in this flow an attacker can choose. A path only.
func TestStateMustBeAPathOnThisSite(t *testing.T) {
	o := NewOAuth("demo-client", "s3cret", "https://example.test/hi/auth/authcode", nil)

	for _, state := range []string{
		"//evil.example",          // scheme-relative: a URL wearing a path's clothes
		"https://evil.example",    // plainly elsewhere
		"http:/\\/\\evil.example", // and a mangling of the same idea
	} {
		err := o.Authorize(httptest.NewRecorder(),
			httptest.NewRequest(http.MethodGet, "/?state="+state, nil))
		if err == nil {
			t.Errorf("state %q was accepted", state)
		}
	}
}

// Missing state is not an error: the flow is started from a link an operator
// follows, and landing back on the front page is the sensible default.
func TestNoStateLandsOnTheFrontPage(t *testing.T) {
	o := NewOAuth("demo-client", "s3cret", "https://example.test/hi/auth/authcode", nil)

	w := httptest.NewRecorder()
	if err := o.Authorize(w, httptest.NewRequest(http.MethodGet, "/", nil)); err != nil {
		t.Fatalf("Authorize: %v", err)
	}
	loc, err := url.Parse(w.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parsing the redirect: %v", err)
	}
	if got := loc.Query().Get("state"); got != "/" {
		t.Errorf("state = %q, want it to default to /", got)
	}
}
