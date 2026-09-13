// Package passkey runs WebAuthn ceremonies.
//
// It is a helper and nothing more: it builds the options a browser is given,
// and it verifies what the browser sends back. It stores nothing, reads no
// cookies and serves no routes — the application holds the challenge between
// the two halves of a ceremony, decides who may start one, and decides what a
// successful one means.
//
// That division is what keeps the WebAuthn part small. Everything else the word
// "authentication" suggests — accounts, sessions, passwords, enrollment — is the
// application's, and lives with the application.
package passkey

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-webauthn/webauthn/webauthn"
)

// Config is the relying party: the domain credentials are bound to, and the
// origins a ceremony may come from.
type Config struct {
	// RPID is the domain a passkey is bound to. A credential created under one
	// cannot be used under another, so this is effectively permanent.
	RPID string

	// RPDisplayName is what an authenticator shows when asking. Defaults to RPID.
	RPDisplayName string

	// Origins a ceremony may come from. Checked here rather than trusted,
	// because an origin that does not match the relying party produces a
	// credential that can never be used, and the failure surfaces much later as
	// something that reads like a broken key.
	Origins []string
}

// Ceremonies runs registrations and logins for one relying party.
type Ceremonies struct {
	wa *webauthn.WebAuthn
}

// New validates the relying party and prepares its ceremonies.
func New(cfg Config) (*Ceremonies, error) {
	if cfg.RPID == "" {
		return nil, fmt.Errorf("passkey: a relying party id is required")
	}
	if len(cfg.Origins) == 0 {
		return nil, fmt.Errorf("passkey: at least one origin is required")
	}
	for _, origin := range cfg.Origins {
		if err := checkOrigin(origin, cfg.RPID); err != nil {
			return nil, err
		}
	}

	name := cfg.RPDisplayName
	if name == "" {
		name = cfg.RPID
	}
	wa, err := webauthn.New(&webauthn.Config{
		RPID:          cfg.RPID,
		RPDisplayName: name,
		RPOrigins:     cfg.Origins,
	})
	if err != nil {
		return nil, fmt.Errorf("passkey: configuring webauthn: %w", err)
	}
	return &Ceremonies{wa: wa}, nil
}

// checkOrigin rejects a misconfigured origin at startup rather than leaving it
// to fail in a ceremony, where it reads as a problem with the credential.
func checkOrigin(origin, rpID string) error {
	u, err := url.Parse(origin)
	if err != nil {
		return fmt.Errorf("passkey: origin %q is not a URL: %w", origin, err)
	}
	if u.Scheme == "" || u.Hostname() == "" {
		return fmt.Errorf("passkey: origin %q needs a scheme and a host", origin)
	}
	host := u.Hostname()
	// The "."-prefixed suffix check keeps evilihle.cloud from matching ihle.cloud.
	if host != rpID && !strings.HasSuffix(host, "."+rpID) {
		return fmt.Errorf("passkey: origin %q is not %s or a subdomain of it", origin, rpID)
	}
	if u.Scheme != "https" && !isLoopback(host) {
		return fmt.Errorf("passkey: origin %q must be https", origin)
	}
	return nil
}

// isLoopback reports whether a host is one a browser treats as a secure context
// without TLS, which is what makes development over plain http possible.
func isLoopback(host string) bool {
	return host == "localhost" || strings.HasSuffix(host, ".localhost") ||
		host == "127.0.0.1" || host == "::1"
}
