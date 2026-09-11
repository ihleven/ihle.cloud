package main

import (
	"fmt"
	"net/url"
	"strings"
)

// The deployment's public identity: the URL a browser reaches this app at, and
// everything that follows from it.
//
// It is stated rather than observed. Behind a proxy the app sees plain HTTP on a
// local port, so its own listener, r.Host and X-Forwarded-Proto all describe
// something other than what a browser sees — and a forwarded header is
// attacker-controlled unless the proxy is known to strip client copies, which
// the app has no way to know. One setting is both safer and simpler.
type site struct {
	// Origin is scheme://host[:port] with no trailing slash: what a browser
	// calls this app, and what an absolute URL is built from.
	Origin string

	// RPID is the WebAuthn relying party: the domain a passkey is bound to.
	RPID string
	// PasskeyOrigins are the origins a ceremony may come from.
	PasskeyOrigins []string

	CookieSecure bool

	// OAuthRedirect is where HiDrive sends the browser back. It has to match the
	// registration held by the provider exactly.
	OAuthRedirect string
}

// oauthCallbackPath is both the route the callback is served on and the path in
// the redirect_uri handed to HiDrive; they cannot be allowed to drift apart.
const oauthCallbackPath = "/hi/auth/authcode"

// newSite derives the host-dependent settings from the public URL. Each can
// still be set outright, so a deployment that needs something other than the
// obvious value is not stuck with it.
func newSite(flags Flags) (*site, error) {
	raw := strings.TrimSpace(flags.PublicURL)
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("PUBLIC_URL %q is not a URL: %w", raw, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("PUBLIC_URL %q needs an http or https scheme", raw)
	}
	if u.Hostname() == "" {
		return nil, fmt.Errorf("PUBLIC_URL %q needs a host", raw)
	}

	s := &site{Origin: u.Scheme + "://" + u.Host}

	// The relying party is a domain, so it carries no port even when the origin
	// does. Overridable because it may need to be a parent domain, which is what
	// lets one credential serve several subdomains.
	s.RPID = flags.PasskeyRPID
	if s.RPID == "" {
		s.RPID = u.Hostname()
	}

	s.PasskeyOrigins = flags.PasskeyOrigins
	if len(s.PasskeyOrigins) == 0 {
		s.PasskeyOrigins = []string{s.Origin}
	}

	// A browser refuses a Secure cookie over plain http — Safari even on
	// localhost — and refuses to send a non-Secure one nowhere. So this follows
	// the scheme rather than being remembered separately.
	s.CookieSecure = u.Scheme == "https"
	if flags.CookieSecure != nil {
		s.CookieSecure = *flags.CookieSecure
	}

	s.OAuthRedirect = s.Origin + oauthCallbackPath

	return s, nil
}
