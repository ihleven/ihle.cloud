package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ihleven/ihlvn/app/db"
	"github.com/ihleven/ihlvn/pkg/authn"
	"github.com/ihleven/ihlvn/pkg/hi"
	"github.com/interhome-group/cms/pkg/errs"
)

// openAuth brings up the authentication service against the account database.
//
// Migrations run here rather than as a separate deployment step, so the binary
// and the schema it expects cannot get out of step.
func openAuth(ctx context.Context, pg *db.DB, site *site, flags Flags, tokens authn.TokenSource) (*authn.Service, error) {
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
	}, slog.Default(), tokens)
}

// hiDriveTokens adapts the HiDrive token manager to what authn asks for: given
// an account's alias, the access token to hand the browser.
type hiDriveTokens struct {
	mngr *hi.TokenMngr
}

func (h hiDriveTokens) AccessToken(alias string) (string, error) {
	refresher := h.mngr.GetTokenRefresher(alias)
	if refresher == nil {
		return "", fmt.Errorf("no HiDrive token stored for alias %q", alias)
	}
	return refresher.GetAccessToken()
}

// requireAccount refuses a request without an account and hands the account to
// the handler in its context.
//
// It reports the refusal as an error rather than writing the response itself, so
// the router's error middleware renders and logs it like any other failure.
// These are API routes, so a browser redirect would be the wrong answer anyway;
// authn.Require is the middleware for routes that a person navigates to.
func requireAccount(svc *authn.Service, h func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) error {
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
func optionalAccount(svc *authn.Service, h func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		if account, ok := svc.Authenticate(r); ok {
			r = r.WithContext(authn.WithAccount(r.Context(), account))
		}
		return h(w, r)
	}
}
