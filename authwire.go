package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ihleven/ihlvn/app/auth"
	"github.com/ihleven/ihlvn/app/cmsauth"
	"github.com/ihleven/ihlvn/app/db"
	"github.com/interhome-group/cms/content"
	"github.com/interhome-group/cms/pkg/errs"
)

// openAuth brings up the authentication service against the account database.
//
// Migrations run here rather than as a separate deployment step, so the binary
// and the schema it expects cannot get out of step.
func openAuth(ctx context.Context, pg *db.DB, site *site, flags Flags) (*auth.Service, error) {
	if err := db.Migrate(ctx, pg.Pool()); err != nil {
		return nil, fmt.Errorf("migrating the account database: %w", err)
	}

	return auth.NewService(auth.NewStore(pg.Pool()), auth.Config{
		CookieName:    flags.CookieName,
		CookieSecure:  site.CookieSecure,
		SameSite:      flags.CookieSameSite,
		SessionTTL:    time.Duration(flags.JWTDuration) * time.Second,
		Issuer:        flags.JWTIssuer,
		Audience:      "famihlie",
		FailureLimit:  10,
		FailureWindow: 15 * time.Minute,

		// Whether an account with no drive of its own still has something to
		// browse, which is what decides if the file browser is offered at all.
		SharedDrive: flags.MediaAlias != "",

		RPID:      site.RPID,
		RPOrigins: site.PasskeyOrigins,
		// RPDisplayName is left to the passkey package, which falls back to the relying party
		// id — so the name a browser shows follows the domain automatically.
	}, slog.Default())
}

// authenticator is the whole of what these wrappers need from the auth service:
// given a request, who is making it.
//
// Narrowed to an interface so the gate below can be exercised on its own. What
// it decides — whether an entitlement admits a request — is worth a test, and
// tying that to a live session store would mean testing cookie parsing instead.
type authenticator interface {
	Authenticate(*http.Request) (*auth.Account, bool)
}

// requireAccount refuses a request without an account and hands the account to
// the handler in its context.
//
// It reports the refusal as an error rather than writing the response itself, so
// the router's error middleware renders and logs it like any other failure.
// These are API routes, so a browser redirect would be the wrong answer anyway;
// auth.Require is the middleware for routes that a person navigates to.
func requireAccount(svc authenticator, h func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		account, ok := svc.Authenticate(r)
		if !ok {
			return errs.New("authentication required", errs.HTTPStatus(http.StatusUnauthorized))
		}
		return h(w, r.WithContext(auth.WithAccount(r.Context(), account)))
	}
}

// optionalAccount attaches the account when there is one. Public routes use it
// so an anonymous reader is served and a signed-in one is identified.
func optionalAccount(svc authenticator, h func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		if account, ok := svc.Authenticate(r); ok {
			r = r.WithContext(auth.WithAccount(r.Context(), account))
		}
		return h(w, r)
	}
}

// requireModule refuses a request whose account is not entitled to an area, and
// hands the account to the handler as requireAccount does.
//
// The entitlement is checked here rather than trusted from the client: the
// navigation hides areas an account may not see, but that is presentation.
// Composed from requireAccount so there is one place that decides what being
// signed in means.
//
// Most areas never reach this — they are offered in a menu and nothing more.
// It is for the ones where the area is the access decision itself.
func requireModule(svc authenticator, key content.PermissionKey, refusal string, h func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) error {
	return requireAccount(svc, func(w http.ResponseWriter, r *http.Request) error {
		account, ok := auth.FromContext(r.Context())
		if !ok || !cmsauth.May(account, key) {
			return errs.New("%s", refusal, errs.HTTPStatus(http.StatusForbidden))
		}
		return h(w, r)
	})
}

// requireAdmin gates account administration, which can rename an account, grant
// it rights or set its password.
func requireAdmin(svc authenticator, h func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) error {
	return requireModule(svc, cmsauth.Admin, "administration requires the admin entitlement", h)
}

// requireHidrive gates browsing the family's storage.
func requireHidrive(svc authenticator, h func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) error {
	return requireModule(svc, cmsauth.Hidrive, "browsing the files requires the hidrive entitlement", h)
}
