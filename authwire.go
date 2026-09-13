package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ihleven/ihlvn/app/db"
	"github.com/ihleven/ihlvn/pkg/authn"
	"github.com/ihleven/ihlvn/pkg/hiauth"
	"github.com/interhome-group/cms/pkg/errs"
)

// openAuth brings up the authentication service against the account database.
//
// Migrations run here rather than as a separate deployment step, so the binary
// and the schema it expects cannot get out of step.
func openAuth(ctx context.Context, pg *db.DB, site *site, flags Flags) (*authn.Service, error) {
	if err := authn.Migrate(ctx, pg.Pool()); err != nil {
		return nil, fmt.Errorf("migrating the account database: %w", err)
	}

	return authn.New(authn.NewStore(pg.Pool()), authn.Config{
		CookieName:    flags.CookieName,
		CookieSecure:  site.CookieSecure,
		SameSite:      flags.CookieSameSite,
		SessionTTL:    time.Duration(flags.JWTDuration) * time.Second,
		Issuer:        flags.JWTIssuer,
		Audience:      "famihlie",
		FailureLimit:  10,
		FailureWindow: 15 * time.Minute,

		RPID:      site.RPID,
		RPOrigins: site.PasskeyOrigins,
		// RPDisplayName is left to authn, which falls back to the relying party
		// id — so the name a browser shows follows the domain automatically.
	}, slog.Default())
}

// hiDriveTokens adapts the HiDrive token manager to what a handler asks for:
// given an account's alias, an access token to use upstream.
type hiDriveTokens struct {
	mngr *hiauth.TokenMngr
}

func (h hiDriveTokens) AccessToken(alias string) (string, error) {
	refresher := h.mngr.GetTokenRefresher(alias)
	if refresher == nil {
		return "", fmt.Errorf("no HiDrive token stored for alias %q", alias)
	}
	return refresher.GetAccessToken()
}

// authenticator is the whole of what these wrappers need from the auth service:
// given a request, who is making it.
//
// Narrowed to an interface so the gate below can be exercised on its own. What
// it decides — whether an entitlement admits a request — is worth a test, and
// tying that to a live session store would mean testing cookie parsing instead.
type authenticator interface {
	Authenticate(*http.Request) (*authn.Account, bool)
}

// requireAccount refuses a request without an account and hands the account to
// the handler in its context.
//
// It reports the refusal as an error rather than writing the response itself, so
// the router's error middleware renders and logs it like any other failure.
// These are API routes, so a browser redirect would be the wrong answer anyway;
// authn.Require is the middleware for routes that a person navigates to.
func requireAccount(svc authenticator, h func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		account, ok := svc.Authenticate(r)
		if !ok {
			return errs.New("authentication required", errs.HTTPStatus(http.StatusUnauthorized))
		}
		return h(w, r.WithContext(authn.WithAccount(r.Context(), account)))
	}
}

// optionalAccount attaches the account when there is one. Public routes use it
// so an anonymous reader is served and a signed-in one is identified.
func optionalAccount(svc authenticator, h func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		if account, ok := svc.Authenticate(r); ok {
			r = r.WithContext(authn.WithAccount(r.Context(), account))
		}
		return h(w, r)
	}
}

// requireAdmin refuses a request whose account is not entitled to the admin
// area, and hands the account to the handler as requireAccount does.
//
// The entitlement is checked here rather than trusted from the client: the
// navigation hides areas an account may not see, but that is presentation, and
// these endpoints can rename an account, grant it rights or set its password.
// Composed from requireAccount so there is one place that decides what being
// signed in means.
func requireAdmin(svc authenticator, h func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) error {
	return requireAccount(svc, func(w http.ResponseWriter, r *http.Request) error {
		account, ok := authn.FromContext(r.Context())
		if !ok || account.ContentUser().Scope.Denies(permAdmin) {
			return errs.New("administration requires the admin entitlement",
				errs.HTTPStatus(http.StatusForbidden))
		}
		return h(w, r)
	})
}
