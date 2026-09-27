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

		// Whether a pool player with no account here can sign in with their
		// pool password, which creates one.
		AdoptGeheimtipp: flags.GeheimtippAdopt,

		// The same screening policy the administrator's path uses, so a
		// password someone picks for themselves is held to one standard.
		MinPasswordLength: flags.MinPasswordLength,

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
//
// It also refuses a confined account, which is what keeps a geheimtipp player
// to the pool. Doing it here rather than route by route means the refusal is
// the default: a route added later is closed to them until someone opens it
// deliberately, instead of open until someone remembers to close it. Nothing
// needs opening today — the pool's own routes and the sign-in endpoints are
// bare handlers that never reach this.
func requireAccount(svc authenticator, h func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		account, ok := svc.Authenticate(r)
		if !ok {
			return errs.New("authentication required", errs.HTTPStatus(http.StatusUnauthorized))
		}
		if account.Confined() {
			return errs.New("this account is limited to the Tipprunde", errs.HTTPStatus(http.StatusForbidden))
		}
		return h(w, r.WithContext(auth.WithAccount(r.Context(), account)))
	}
}

// poolIdentity answers the geheimtipp proxy's one question: who is making this
// request, and may they act as themselves on the pool?
//
// It lives here rather than in the geheimtipp package because it is the join
// between the two — what an account is, and what the pool needs to be told —
// and that join belongs where the wiring is.
type poolIdentity struct{ svc authenticator }

func (p poolIdentity) PoolLogin(r *http.Request) (string, bool) {
	account, ok := p.svc.Authenticate(r)
	if !ok {
		return "", false
	}
	// A confined account exists for nothing but the pool, so it needs no
	// entitlement to prove it. A full account needs one, which is how an
	// administrator says that this person plays — and without it the request
	// goes on anonymously rather than as a name the pool has never heard of.
	if account.Confined() || cmsauth.May(account, cmsauth.Geheimtipp) {
		return account.Name, true
	}
	return "", false
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

// requirePermission refuses a request whose account does not hold a right, and
// hands the account to the handler as requireAccount does.
//
// The right is checked here rather than trusted from the client: the navigation
// hides what an account may not use, but that is presentation. Composed from
// requireAccount so there is one place that decides what being signed in means,
// and so the two refusals stay distinct — 401 for nobody, 403 for somebody
// without the right.
//
// It takes any right, because the check does not depend on which kind it is.
// Which kind it is matters to a reader, which is what requireModule below is
// for: the name at the call site says whether an area or a capability is being
// demanded, and those are answers to different questions.
func requirePermission(svc authenticator, key content.PermissionKey, refusal string, h func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) error {
	return requireAccount(svc, func(w http.ResponseWriter, r *http.Request) error {
		account, ok := auth.FromContext(r.Context())
		if !ok || !cmsauth.May(account, key) {
			return errs.New("%s", refusal, errs.HTTPStatus(http.StatusForbidden))
		}
		return h(w, r)
	})
}

// requireModule refuses a request whose account is not entitled to an area.
//
// The same check as requirePermission and a different statement: this one says
// the thing being demanded is a *place in this app* that a visitor may or may
// not be shown, rather than something they may or may not do. Most areas never
// reach here — they are offered in a menu and nothing more — and this is for
// the ones where being in the area is itself the access decision.
func requireModule(svc authenticator, area content.PermissionKey, refusal string, h func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) error {
	return requirePermission(svc, area, refusal, h)
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

// cmsIdentity hands the CMS's own controllers an identity they can read.
//
// They are transport-free and ask for the signed-in person through
// content.RequireUser, which reads one key; this app's middleware writes its own
// Account under another and converts at each call site. So the two have to be
// joined somewhere, and this is the join: authenticate once, and say who it is
// in both vocabularies.
//
// Optional rather than required, like optionalAccount: reading content is open
// in this app — the public site is rendered from it — and every entry carries an
// ACL that decides for itself. What needs an account refuses for itself too: the
// controllers answer 401 when there is nobody to name.
func cmsIdentity(svc authenticator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if account, ok := svc.Authenticate(r); ok {
			ctx := auth.WithAccount(r.Context(), account)
			ctx = content.ContextWithUser(ctx, account)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}
