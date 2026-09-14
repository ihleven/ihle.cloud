// Package auth is ihlvn's own authentication: local accounts in Postgres,
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
package auth

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ihleven/ihlvn/pkg/passkey"
	"github.com/ihleven/ihlvn/pkg/password"
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

	// SharedDrive says whether the deployment configured storage that an account
	// with none of its own falls back to.
	//
	// It is here because the session reports which areas the frontend may offer,
	// and an area whose whole content is a drive has nothing to offer without
	// one. Held as a bool rather than as the alias: what this decides is whether
	// there is anything to browse, not what.
	SharedDrive bool

	// AdoptGeheimtipp enables the migration fallback: a name with no account
	// here is looked up among the pool's players, and a matching password
	// creates the account.
	//
	// It is a setting rather than a constant because it is meant to be turned
	// off. Once everyone who plays has signed in once, geheimtipp_migration is
	// empty and the fallback is dead weight holding a table of plaintext
	// passwords open. Turning it off is the step before dropping both.
	AdoptGeheimtipp bool

	// MinPasswordLength is where the screening starts calling a password short.
	// Passed in rather than fixed, for the same reason the administrator's path
	// takes it: it is a policy, not a fact.
	MinPasswordLength int
}

// Service is the use-case and HTTP half. Store is the persistence half; nothing
// here writes SQL.
type Service struct {
	store    *Store
	cfg      Config
	log      *slog.Logger
	throttle *throttle
	// ceremonies is nil when RPID is unset: passkeys are optional, passwords are
	// not. The WebAuthn protocol itself lives in pkg/passkey; what is here is who
	// may start a ceremony and what finishing one means.
	ceremonies *passkey.Ceremonies
	// breaches screens a password someone chooses for themselves, exactly as the
	// administrator's path screens one chosen for them. Held so a test can point
	// it at a stub rather than at a third party over the network.
	breaches *password.BreachChecker
}

// New builds the service, filling in defaults for anything the caller left zero.
func NewService(store *Store, cfg Config, log *slog.Logger) (*Service, error) {
	if store == nil {
		return nil, errors.New("auth: a store is required")
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
		return nil, fmt.Errorf("auth: SlideAfter (%s) must be shorter than SessionTTL (%s), "+
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
		throttle: newThrottle(cfg.FailureLimit, cfg.FailureWindow),
		breaches: &password.BreachChecker{},
	}

	if cfg.RPID != "" {
		ceremonies, err := passkey.New(passkey.Config{
			RPID:          cfg.RPID,
			RPDisplayName: cfg.RPDisplayName,
			Origins:       cfg.RPOrigins,
		})
		if err != nil {
			return nil, err
		}
		svc.ceremonies = ceremonies
	}
	return svc, nil
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
				s.log.Error("auth cleanup failed", "err", err)
			}
			cancel()
		}
	}
}
