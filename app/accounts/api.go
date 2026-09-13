// Package accounts is account administration over HTTP: the operations
// `ihlvn account` performs from a terminal, performed from the app instead.
//
// The CLI remains the bootstrap and recovery path — the first administrator has
// to exist before anyone can sign in to make one, and a broken UI is exactly
// when the terminal is needed. So this is a second front end onto the same
// store, not a replacement: nothing here knows anything the CLI does not.
//
// Every handler assumes it is mounted behind an admin gate. The check lives at
// the route (requireAdmin) rather than in each handler, so there is one place
// that can be read to see what is protected.
package accounts

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/ihleven/ihlvn/pkg/authn"
	"github.com/interhome-group/cms/modules"
	"github.com/interhome-group/cms/pkg/errs"
)

// API holds what the handlers need: the account store, and the two settings an
// enrollment link is built from.
type API struct {
	store *authn.Store

	// origin is the public base an enrollment link is offered at. It is passed
	// in rather than read off the request, for the same reason the rest of the
	// app derives from PUBLIC_URL: behind a proxy the request describes the
	// local hop, not what the browser will be asked to open.
	origin    string
	enrollTTL time.Duration

	// breaches is the password screening. Held so a test can point it at a stub
	// instead of the public service.
	breaches *authn.BreachChecker
}

func New(store *authn.Store, origin string, enrollTTL time.Duration) *API {
	return &API{store: store, origin: origin, enrollTTL: enrollTTL, breaches: &authn.BreachChecker{}}
}

// view is how an account is presented. It is a projection rather than the stored
// struct: the password hash and the WebAuthn handle never leave the server, and
// two derived things are added because the operator needs them to make sense of
// what they are looking at.
type view struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Disabled    bool   `json:"disabled"`

	Groups      []string `json:"groups"`
	Permissions []string `json:"permissions"`

	// Unregistered are permissions this build does not define. They are stored
	// and look like grants, but a scope drops them, so they grant nothing —
	// worth showing rather than leaving as an invisible discrepancy.
	Unregistered []string `json:"unregistered_permissions"`

	// Modules is what the permissions actually add up to: the areas this account
	// may see. Derived here so the UI does not have to reimplement the wildcard.
	Modules []string `json:"modules"`

	// How the account can sign in. Without either it cannot, which is the state
	// a freshly created account is in.
	HasPassword bool `json:"has_password"`
	Passkeys    int  `json:"passkeys"`

	CreatedAt time.Time `json:"created_at"`
}

// present builds the view, counting passkeys with a query per account. That is
// an N+1, and deliberate: this is a family archive with a handful of accounts,
// and the CLI's listing does the same.
func (a *API) present(r *http.Request, account *authn.Account) (view, error) {
	keys, err := a.store.ListPasskeys(r.Context(), account.ID)
	if err != nil {
		return view{}, err
	}
	return view{
		Name:         account.Name,
		DisplayName:  account.DisplayName,
		Email:        account.Email,
		Disabled:     account.Disabled,
		Groups:       orEmpty(account.CMS.Groups),
		Permissions:  orEmpty(account.CMS.Permissions),
		Unregistered: orEmpty(account.UnregisteredPermissions()),
		Modules:      orEmpty(account.EntitledModules()),
		HasPassword:  account.HasPassword(),
		Passkeys:     len(keys),
		CreatedAt:    account.CreatedAt,
	}, nil
}

// List answers with every account, including disabled ones: an administrator
// needs to see the account they turned off in order to turn it back on.
func (a *API) List(w http.ResponseWriter, r *http.Request) error {
	accounts, err := a.store.ListAccounts(r.Context())
	if err != nil {
		return err
	}

	views := make([]view, 0, len(accounts))
	for _, account := range accounts {
		v, err := a.present(r, account)
		if err != nil {
			return err
		}
		views = append(views, v)
	}
	return writeJSON(w, http.StatusOK, views)
}

// Get answers with one account.
func (a *API) Get(w http.ResponseWriter, r *http.Request) error {
	account, err := a.lookup(r)
	if err != nil {
		return err
	}
	v, err := a.present(r, account)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, v)
}

type createBody struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

// Create adds an account. It has no way to sign in yet: a password or a passkey
// is a separate step, which is what the enrollment link is for.
func (a *API) Create(w http.ResponseWriter, r *http.Request) error {
	var body createBody
	if err := decode(r, &body); err != nil {
		return err
	}
	if body.Name == "" {
		return errs.New("an account needs a login name", errs.HTTPStatus(http.StatusBadRequest))
	}
	if body.Email == "" {
		return errs.New("an account needs an email address", errs.HTTPStatus(http.StatusBadRequest))
	}
	if body.DisplayName == "" {
		body.DisplayName = body.Name
	}

	account, err := a.store.CreateAccount(r.Context(), body.Name, body.DisplayName, body.Email)
	if err != nil {
		return asStatus(err)
	}
	v, err := a.present(r, account)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusCreated, v)
}

// updateBody is everything about an account that can be changed at once.
//
// Being disabled is part of it rather than a resource of its own: it is a
// property of the account, and the form that edits an account edits all of it.
// The login name is absent because entries are owned by it.
type updateBody struct {
	DisplayName string   `json:"display_name"`
	Email       string   `json:"email"`
	Groups      []string `json:"groups"`
	Permissions []string `json:"permissions"`
	Disabled    bool     `json:"disabled"`
}

// Update replaces the mutable half of an account.
//
// The writes are not one transaction — identity and disabled live on account,
// groups and permissions on cmsauth — so a failure part way leaves a partly
// applied edit. It is visible in the response the operator gets back and can be
// repeated, which is the cheaper answer here than threading a transaction
// through the store for a form that one person submits.
func (a *API) Update(w http.ResponseWriter, r *http.Request) error {
	account, err := a.lookup(r)
	if err != nil {
		return err
	}

	var body updateBody
	if err := decode(r, &body); err != nil {
		return err
	}
	if body.DisplayName == "" {
		body.DisplayName = account.Name
	}
	if body.Email == "" {
		return errs.New("an account needs an email address", errs.HTTPStatus(http.StatusBadRequest))
	}
	if err := a.refuseLockout(r, account, body); err != nil {
		return err
	}

	if err := a.store.SetIdentity(r.Context(), account.ID, body.DisplayName, body.Email); err != nil {
		return asStatus(err)
	}
	if err := a.store.SetCMSProfile(r.Context(), account.ID,
		authn.CMSProfile{Groups: body.Groups, Permissions: body.Permissions}); err != nil {
		return err
	}
	if body.Disabled != account.Disabled {
		if _, err := a.store.SetDisabled(r.Context(), account.Name, body.Disabled); err != nil {
			return asStatus(err)
		}
		// Loading refuses a disabled account, but an existing session would go
		// on working until it expired. Dropping them makes the refusal immediate.
		if body.Disabled {
			if _, err := a.store.DeleteSessions(r.Context(), account.ID); err != nil {
				return err
			}
		}
	}

	// Re-read rather than patching the loaded copy, so the response is what the
	// database now holds and a partly applied edit shows as what it is. Through
	// the lookup that sees disabled accounts, or disabling one would come back
	// as "no such account" instead of as the account just disabled.
	updated, err := a.store.DisabledAccountByName(r.Context(), account.Name)
	if err != nil {
		return err
	}
	v, err := a.present(r, updated)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, v)
}

// refuseLockout stops an administrator removing their own admin rights or
// disabling themselves.
//
// Accounts cannot be deleted and admin is granted through this API, so an
// administrator who revokes their own entitlement has no way back except the
// terminal — which is the thing this section exists to avoid needing. Someone
// else's admin rights can still be removed: the constraint is about not locking
// the door from the inside, not about protecting the role.
func (a *API) refuseLockout(r *http.Request, target *authn.Account, body updateBody) error {
	caller, ok := authn.FromContext(r.Context())
	if !ok || caller.ID != target.ID {
		return nil
	}

	after := &authn.Account{CMS: authn.CMSProfile{Groups: body.Groups, Permissions: body.Permissions}}
	if body.Disabled {
		return errs.New("an administrator cannot disable their own account",
			errs.HTTPStatus(http.StatusConflict))
	}
	if !slices.Contains(after.EntitledModules(), "admin") {
		return errs.New("an administrator cannot remove their own admin entitlement",
			errs.HTTPStatus(http.StatusConflict))
	}
	return nil
}

// lookup resolves the {name} in the path.
func (a *API) lookup(r *http.Request) (*authn.Account, error) {
	name := r.PathValue("name")
	if name == "" {
		return nil, errs.New("no account named", errs.HTTPStatus(http.StatusBadRequest))
	}
	// Deliberately the lookup that sees disabled accounts: administration is
	// where one gets turned back on.
	account, err := a.store.DisabledAccountByName(r.Context(), name)
	if err != nil {
		return nil, asStatus(err)
	}
	return account, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func decode(r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return errs.New("could not read the request body: %v", err,
			errs.HTTPStatus(http.StatusBadRequest))
	}
	return nil
}

// asStatus maps the store's sentinel errors onto the status they mean. Anything
// else is left alone, so an unexpected failure stays a 500 rather than being
// dressed up as something the caller did wrong.
func asStatus(err error) error {
	switch {
	case errors.Is(err, authn.ErrNoAccount):
		return errs.New("no such account", errs.HTTPStatus(http.StatusNotFound))
	case errors.Is(err, authn.ErrExists):
		return errs.New("an account with that name or address already exists",
			errs.HTTPStatus(http.StatusConflict))
	}
	return err
}

// orEmpty renders a nil slice as [] rather than null, so the client never has to
// distinguish "none" from "absent".
func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// Areas answers with every feature area this build defines.
//
// The picker that grants them needs the list, and the only source of truth for
// it is the registry in the running binary — a copy in the frontend would drift
// the moment an area is added, and grant names that resolve to nothing.
func (a *API) Areas(w http.ResponseWriter, _ *http.Request) error {
	return writeJSON(w, http.StatusOK, modules.RegisteredModules())
}
