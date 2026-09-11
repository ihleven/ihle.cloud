package main

import (
	"slices"
	"strings"
	"testing"
)

// The point of deriving is that one setting is enough, so these pin what each
// public URL produces. Getting a relying party id or an origin subtly wrong does
// not fail at startup — it fails later, in a ceremony, as an error about a
// credential rather than about configuration.
func TestNewSiteDerives(t *testing.T) {
	tests := []struct {
		name          string
		publicURL     string
		wantOrigin    string
		wantRPID      string
		wantOrigins   []string
		wantSecure    bool
		wantOAuthPath string
	}{{
		name: "a plain https domain", publicURL: "https://tschabrun.de",
		wantOrigin: "https://tschabrun.de", wantRPID: "tschabrun.de",
		wantOrigins: []string{"https://tschabrun.de"}, wantSecure: true,
		wantOAuthPath: "https://tschabrun.de/hi/auth/authcode",
	}, {
		name: "a subdomain", publicURL: "https://app.tschabrun.de",
		wantOrigin: "https://app.tschabrun.de", wantRPID: "app.tschabrun.de",
		wantOrigins: []string{"https://app.tschabrun.de"}, wantSecure: true,
		wantOAuthPath: "https://app.tschabrun.de/hi/auth/authcode",
	}, {
		// The origin keeps the port; the relying party id is a domain and never
		// carries one.
		name: "development, with a port", publicURL: "http://localhost:8000",
		wantOrigin: "http://localhost:8000", wantRPID: "localhost",
		wantOrigins: []string{"http://localhost:8000"}, wantSecure: false,
		wantOAuthPath: "http://localhost:8000/hi/auth/authcode",
	}, {
		name: "a trailing slash is not part of the origin", publicURL: "https://tschabrun.de/",
		wantOrigin: "https://tschabrun.de", wantRPID: "tschabrun.de",
		wantOrigins: []string{"https://tschabrun.de"}, wantSecure: true,
		wantOAuthPath: "https://tschabrun.de/hi/auth/authcode",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := newSite(Flags{PublicURL: tt.publicURL})
			if err != nil {
				t.Fatalf("newSite: %v", err)
			}
			if s.Origin != tt.wantOrigin {
				t.Errorf("Origin = %q, want %q", s.Origin, tt.wantOrigin)
			}
			if s.RPID != tt.wantRPID {
				t.Errorf("RPID = %q, want %q", s.RPID, tt.wantRPID)
			}
			if !slices.Equal(s.PasskeyOrigins, tt.wantOrigins) {
				t.Errorf("PasskeyOrigins = %v, want %v", s.PasskeyOrigins, tt.wantOrigins)
			}
			if s.CookieSecure != tt.wantSecure {
				t.Errorf("CookieSecure = %v, want %v", s.CookieSecure, tt.wantSecure)
			}
			if s.OAuthRedirect != tt.wantOAuthPath {
				t.Errorf("OAuthRedirect = %q, want %q", s.OAuthRedirect, tt.wantOAuthPath)
			}
		})
	}
}

// Deriving is a default, not a rule: a deployment that needs something else must
// be able to say so.
func TestNewSiteOverrides(t *testing.T) {
	no := false

	s, err := newSite(Flags{
		PublicURL: "https://app.tschabrun.de",
		// A parent domain, which is what lets one credential serve several
		// subdomains.
		PasskeyRPID:    "tschabrun.de",
		PasskeyOrigins: []string{"https://app.tschabrun.de", "https://alt.tschabrun.de"},
		CookieSecure:   &no,
	})
	if err != nil {
		t.Fatalf("newSite: %v", err)
	}

	if s.RPID != "tschabrun.de" {
		t.Errorf("RPID = %q, want the explicit parent domain", s.RPID)
	}
	if len(s.PasskeyOrigins) != 2 {
		t.Errorf("PasskeyOrigins = %v, want both explicit origins", s.PasskeyOrigins)
	}
	if s.CookieSecure {
		t.Error("an explicit CookieSecure=false was overridden by the scheme")
	}
}

// An unusable value stops the process, rather than leaving a deployment that
// looks up and fails only when someone tries to sign in.
func TestNewSiteRejectsUnusableURLs(t *testing.T) {
	for _, publicURL := range []string{
		"",                   // nothing at all
		"tschabrun.de",       // no scheme, so no origin
		"ftp://tschabrun.de", // not a scheme a browser speaks here
		"https://",           // no host
		"://tschabrun.de",    // unparseable
	} {
		if _, err := newSite(Flags{PublicURL: publicURL}); err == nil {
			t.Errorf("newSite(%q) was accepted", publicURL)
		}
	}
}

// The callback path is shared between the route and the redirect_uri handed to
// the provider; if they drift the provider rejects the exchange.
func TestOAuthRedirectUsesTheRoutePath(t *testing.T) {
	s, err := newSite(Flags{PublicURL: "https://tschabrun.de"})
	if err != nil {
		t.Fatalf("newSite: %v", err)
	}
	if !strings.HasSuffix(s.OAuthRedirect, oauthCallbackPath) {
		t.Errorf("OAuthRedirect = %q, want it to end in %q", s.OAuthRedirect, oauthCallbackPath)
	}
}
