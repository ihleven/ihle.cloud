// Package authn is ihlvn's own authentication: local accounts in Postgres,
// passwords and passkeys as credentials, and server-side sessions.
//
// The account database is the authority for identity. An account carries the
// two things the rest of the application needs from it: the groups and
// permissions the CMS is handed as a content.User, and the HiDrive alias that
// says which stored credential serves that person's media.
//
// Two credentials are supported on purpose. A password is what everyone has and
// what a step-up re-authentication asks for; a passkey is the stronger option,
// restricted here to keys that cannot be synced between devices. Neither is a
// fallback for the other: adding a passkey requires the password, so the
// stronger credential cannot be attached by whoever merely holds a session.
package authn

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

// Config is the deployment's half of the service: names, lifetimes and the
// cookie's security posture, none of which the package should decide for itself.
type Config struct {
	// CookieName is the session cookie. Kept configurable because the old JWT
	// cookie name is in browsers already.
	CookieName string
	// CookieSecure should be false only when serving plain HTTP in development.
	CookieSecure bool
	SameSite     http.SameSite

	SessionTTL time.Duration
	// SlideAfter is how stale a session's last_seen must be before its expiry is
	// extended. Without it every request — including every range request of a
	// video — would be a database write.
	SlideAfter time.Duration

	// Issuer and Audience are carried in the session response for the frontend,
	// which was written against the JWT claims this replaced.
	Issuer   string
	Audience string

	// FailureLimit and FailureWindow throttle password attempts, per account and
	// per client address.
	FailureLimit  int
	FailureWindow time.Duration

	// WebAuthn. Leaving RPID empty leaves passkeys unconfigured, and their
	// endpoints answer 501 — the password login still works.
	RPID          string
	RPDisplayName string
	// RPOrigins are the exact origins a ceremony may come from. A credential is
	// bound to these, so they are checked at startup rather than guessed.
	RPOrigins      []string
	ChallengeTTL   time.Duration
	EnrollTTL      time.Duration
	CeremonyCookie string
}

// TokenSource yields a HiDrive access token for an alias. The session endpoint
// hands one to the browser so the media routes can use it; authn only needs to
// be able to ask.
type TokenSource interface {
	AccessToken(alias string) (string, error)
}

// Service is the use-case and HTTP half. Store is the persistence half; nothing
// here writes SQL.
type Service struct {
	store    *Store
	cfg      Config
	log      *slog.Logger
	tokens   TokenSource
	throttle *throttle
	// wa is nil when RPID is unset: passkeys are optional, passwords are not.
	wa *webauthn.WebAuthn
}

// New builds the service, filling in defaults for anything the caller left zero.
func New(store *Store, cfg Config, log *slog.Logger, tokens TokenSource) (*Service, error) {
	if store == nil {
		return nil, errors.New("authn: a store is required")
	}
	if log == nil {
		log = slog.Default()
	}
	if cfg.CookieName == "" {
		cfg.CookieName = "session"
	}
	if cfg.SessionTTL <= 0 {
		cfg.SessionTTL = 30 * 24 * time.Hour
	}
	if cfg.SlideAfter <= 0 {
		cfg.SlideAfter = 15 * time.Minute
	}
	if cfg.SameSite == 0 {
		cfg.SameSite = http.SameSiteLaxMode
	}
	if cfg.FailureLimit <= 0 {
		cfg.FailureLimit = 10
	}
	if cfg.FailureWindow <= 0 {
		cfg.FailureWindow = 15 * time.Minute
	}
	if cfg.SlideAfter >= cfg.SessionTTL {
		return nil, fmt.Errorf("authn: SlideAfter (%s) must be shorter than SessionTTL (%s), "+
			"or a session can expire before it is ever extended", cfg.SlideAfter, cfg.SessionTTL)
	}

	if cfg.ChallengeTTL <= 0 {
		cfg.ChallengeTTL = 5 * time.Minute
	}
	if cfg.EnrollTTL <= 0 {
		cfg.EnrollTTL = 15 * time.Minute
	}
	if cfg.CeremonyCookie == "" {
		cfg.CeremonyCookie = "ihlvn_ceremony"
	}

	svc := &Service{
		store:    store,
		cfg:      cfg,
		log:      log,
		tokens:   tokens,
		throttle: newThrottle(cfg.FailureLimit, cfg.FailureWindow),
	}

	if cfg.RPID != "" {
		for _, origin := range cfg.RPOrigins {
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
			RPOrigins:     cfg.RPOrigins,
		})
		if err != nil {
			return nil, fmt.Errorf("authn: configuring webauthn: %w", err)
		}
		svc.wa = wa
	}
	return svc, nil
}

// checkOrigin rejects a misconfigured origin at startup rather than leaving
// every ceremony to fail at run time. An origin must be the relying party's own
// host or a subdomain of it, and must be https unless it is localhost.
func checkOrigin(origin, rpID string) error {
	u, err := url.Parse(origin)
	if err != nil {
		return fmt.Errorf("authn: origin %q is not a URL: %w", origin, err)
	}
	if u.Scheme == "" || u.Hostname() == "" {
		return fmt.Errorf("authn: origin %q needs a scheme and a host", origin)
	}
	host := u.Hostname()
	// The "."-prefixed suffix check keeps evilihle.cloud from matching ihle.cloud.
	if host != rpID && !strings.HasSuffix(host, "."+rpID) {
		return fmt.Errorf("authn: origin %q is not %s or a subdomain of it", origin, rpID)
	}
	if u.Scheme != "https" && !isLoopback(host) {
		return fmt.Errorf("authn: origin %q must be https", origin)
	}
	return nil
}

// isLoopback reports whether a host always resolves to this machine, which the
// URL standard treats as trustworthy even over plain http.
//
// The *.localhost form matters in practice: two applications developed on plain
// localhost share one relying-party id, so a browser offers each one's passkeys
// to the other. Giving them ihlvn.localhost and media.localhost separates the
// credentials without inventing certificates.
func isLoopback(host string) bool {
	return host == "localhost" || strings.HasSuffix(host, ".localhost") ||
		host == "127.0.0.1" || host == "::1"
}

// decodeCredentialID reads the base64url form a credential id is shown in.
func decodeCredentialID(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

// Store exposes persistence for the admin commands, which need it without a
// second constructor.
func (s *Service) Store() *Store { return s.store }

// StartCleanup drops expired rows periodically. Correctness does not depend on
// it — every query filters on expiry — so a failure is logged, not fatal.
func (s *Service) StartCleanup(ctx context.Context, every time.Duration) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.throttle.sweep()
			runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			if err := s.store.Cleanup(runCtx); err != nil {
				s.log.Error("authn cleanup failed", "err", err)
			}
			cancel()
		}
	}
}
