package hiauth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// The refresh response carries a new access token and nothing else. Replacing
// the token with it blanks the refresh token, and every later refresh then asks
// the provider with an empty credential — which is what turned a failed refresh
// into a flood of rejected requests.
func TestMergedKeepsTheRefreshToken(t *testing.T) {
	stored := Token{
		AccessToken:  "old-access",
		RefreshToken: "the-refresh-token",
		Alias:        "ihleven",
		Scope:        "admin,rw",
		TokenType:    "Bearer",
	}

	// What HiDrive actually returns: an access token, a lifetime, an alias.
	got := stored.merged(Token{AccessToken: "new-access", ExpiresIn: 3600, Alias: "ihleven"})

	if got.RefreshToken != "the-refresh-token" {
		t.Errorf("refresh token = %q, want it kept", got.RefreshToken)
	}
	if got.Scope != "admin,rw" {
		t.Errorf("scope = %q, want it kept", got.Scope)
	}
	if got.AccessToken != "new-access" || got.ExpiresIn != 3600 {
		t.Errorf("access token not applied: %+v", got)
	}
}

// A response that does supply a new refresh token replaces the old one.
func TestMergedTakesARotatedRefreshToken(t *testing.T) {
	stored := Token{RefreshToken: "old", Alias: "a"}

	got := stored.merged(Token{AccessToken: "x", RefreshToken: "rotated"})

	if got.RefreshToken != "rotated" {
		t.Errorf("refresh token = %q, want rotated", got.RefreshToken)
	}
}

// The wait before renewing is the token's life less a safety margin — but never
// negative or near-zero, which would schedule the next refresh immediately.
func TestRefreshDelayIsNeverImmediate(t *testing.T) {
	for _, tc := range []struct {
		name      string
		expiresIn time.Duration
		want      time.Duration
	}{
		{"a normal hour-long token", time.Hour, time.Hour - lifeSpanSafetyMargin},
		{"shorter than the safety margin", 30 * time.Second, minRefreshDelay},
		{"already expired", 0, minRefreshDelay},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := refreshDelay(tc.expiresIn); got != tc.want {
				t.Errorf("refreshDelay(%s) = %s, want %s", tc.expiresIn, got, tc.want)
			}
		})
	}
}

// An empty refresh token is refused without a request: there is nothing to ask
// with, so asking only adds load the provider has to reject.
func TestRefreshTokenRefusesAnEmptyCredential(t *testing.T) {
	var calls atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
	}))
	defer srv.Close()

	mngr := &TokenMngr{tokenURL: srv.URL, authclient: authclient{httpclient: srv.Client()}}

	_, err := mngr.RefreshToken("")

	var oauthErr *OAuthError
	if !errors.As(err, &oauthErr) || !oauthErr.Permanent() {
		t.Fatalf("err = %v, want a permanent OAuthError", err)
	}
	if calls.Load() != 0 {
		t.Errorf("made %d requests, want none", calls.Load())
	}
}

// A rejected refresh token is reported as permanent, so the loop stops instead
// of retrying something that cannot succeed without a human.
func TestRefreshTokenReportsRejectionAsPermanent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"refresh token expired"}`))
	}))
	defer srv.Close()

	mngr := &TokenMngr{tokenURL: srv.URL, authclient: authclient{httpclient: srv.Client()}}

	_, err := mngr.RefreshToken("stale")

	var oauthErr *OAuthError
	if !errors.As(err, &oauthErr) {
		t.Fatalf("err = %v, want an OAuthError", err)
	}
	if !oauthErr.Permanent() {
		t.Errorf("invalid_grant should be permanent, got %v", oauthErr)
	}
}

// The loop must not keep asking after a permanent rejection. Before the fix a
// failure scheduled the next attempt at (100ms - 1min), a negative duration that
// fires at once, so this would have been thousands of requests.
func TestRefresherStopsAskingAfterPermanentRejection(t *testing.T) {
	var calls atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"refresh token expired"}`))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mngr := &TokenMngr{ctx: ctx, tokenURL: srv.URL, authclient: authclient{httpclient: srv.Client()}}
	refresher := mngr.newTokenRefresher(Token{RefreshToken: "stale", Alias: "ihleven"})

	// The caller is served the failure rather than blocking forever.
	if _, err := refresher.GetAccessToken(); err == nil {
		t.Fatal("expected the rejection to reach the caller")
	}

	time.Sleep(300 * time.Millisecond)

	if n := calls.Load(); n != 1 {
		t.Errorf("made %d token requests, want exactly 1", n)
	}
}

// A successful refresh serves the new access token and keeps the credential it
// was obtained with.
func TestRefresherServesTheAccessTokenAndKeepsTheCredential(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.FormValue("refresh_token"); got != "the-refresh-token" {
			t.Errorf("asked with refresh_token=%q", got)
		}
		_, _ = w.Write([]byte(`{"access_token":"fresh","expires_in":3600,"alias":"ihleven"}`))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mngr := &TokenMngr{ctx: ctx, tokenURL: srv.URL, authclient: authclient{httpclient: srv.Client()}}
	refresher := mngr.newTokenRefresher(Token{RefreshToken: "the-refresh-token", Alias: "ihleven"})

	token, err := refresher.GetAccessToken()
	if err != nil {
		t.Fatalf("GetAccessToken: %v", err)
	}
	if token != "fresh" {
		t.Errorf("access token = %q, want fresh", token)
	}
}

// A transient failure must back off rather than retry at once. This is the
// defect that reached production: the next attempt was scheduled at
// (retryDelay - lifeSpanSafetyMargin) = 100ms - 1min, a negative duration, so
// the timer fired immediately and the loop asked as fast as the network allowed.
func TestRefresherBacksOffAfterATransientFailure(t *testing.T) {
	var calls atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"server_error","error_description":"try again"}`))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mngr := &TokenMngr{ctx: ctx, tokenURL: srv.URL, authclient: authclient{httpclient: srv.Client()}}
	mngr.newTokenRefresher(Token{RefreshToken: "good", Alias: "ihleven"})

	time.Sleep(500 * time.Millisecond)

	if n := calls.Load(); n != 1 {
		t.Errorf("made %d token requests in 500ms, want 1 — the retry should be seconds away", n)
	}
}
