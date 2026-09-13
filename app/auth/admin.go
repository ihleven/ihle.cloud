package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/ihleven/ihlvn/pkg/password"
	"github.com/interhome-group/cms/pkg/errs"
)

// AdminAPI is administration over HTTP: read the request, call Admin, write the
// answer. Every exported method is a handler, and none of them decides anything —
// what may happen is Admin's to say, and this only translates.
//
// Every handler assumes it is mounted behind an admin gate. The check lives at
// the route (requireAdmin) rather than in each handler, so there is one place
// that can be read to see what is protected.
type AdminAPI struct {
	admin *Admin
}

func NewAdminAPI(admin *Admin) *AdminAPI {
	return &AdminAPI{admin: admin}
}

// AccountInfo is how an account is presented. It is a projection rather than the
// stored struct: the password hash and the WebAuthn handle never leave the
// server, and two derived things are added because whoever is looking needs them
// to make sense of what they see.
type AccountInfo struct {
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
	// may see. Derived here so a client does not have to reimplement the wildcard.
	Modules []string `json:"modules"`

	// How the account can sign in. Without either it cannot, which is the state
	// a freshly created account is in.
	HasPassword bool `json:"has_password"`
	Passkeys    int  `json:"passkeys"`

	CreatedAt time.Time `json:"created_at"`
}

// Enrollment is a single-use link and when it stops working.
type Enrollment struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// PasswordResult says whether a password was set, and what the screening found
// either way — a caller told "not set" is also told why.
type PasswordResult struct {
	Set    bool            `json:"set"`
	Advice password.Advice `json:"advice"`
}

func (h *AdminAPI) List(w http.ResponseWriter, r *http.Request) error {
	accounts, err := h.admin.List(r.Context())
	if err != nil {
		return asStatus(err)
	}
	return writeJSON(w, http.StatusOK, accounts)
}

func (h *AdminAPI) Get(w http.ResponseWriter, r *http.Request) error {
	account, err := h.admin.Get(r.Context(), r.PathValue("name"))
	if err != nil {
		return asStatus(err)
	}
	return writeJSON(w, http.StatusOK, account)
}

func (h *AdminAPI) Create(w http.ResponseWriter, r *http.Request) error {
	var in NewAccount
	if err := decode(r, &in); err != nil {
		return err
	}
	account, err := h.admin.Create(r.Context(), in)
	if err != nil {
		return asStatus(err)
	}
	return writeJSON(w, http.StatusCreated, account)
}

func (h *AdminAPI) Update(w http.ResponseWriter, r *http.Request) error {
	var edit AccountEdit
	if err := decode(r, &edit); err != nil {
		return err
	}
	account, err := h.admin.Update(r.Context(), r.PathValue("name"), edit)
	if err != nil {
		return asStatus(err)
	}
	return writeJSON(w, http.StatusOK, account)
}

func (h *AdminAPI) SetPassword(w http.ResponseWriter, r *http.Request) error {

	var body struct {
		Password string `json:"password"`
		// Confirm carries the decision to use a password the screening objected to.
		Confirm bool `json:"confirm"`
	}
	if err := decode(r, &body); err != nil {
		return err
	}

	result, err := h.admin.SetPassword(r.Context(), r.PathValue("name"), body.Password, body.Confirm)
	if err != nil {
		return asStatus(err)
	}

	return writeJSON(w, http.StatusOK, result)
}

func (h *AdminAPI) IssueEnrollment(w http.ResponseWriter, r *http.Request) error {
	link, err := h.admin.IssueEnrollment(r.Context(), r.PathValue("name"))
	if err != nil {
		return asStatus(err)
	}
	return writeJSON(w, http.StatusOK, link)
}

func (h *AdminAPI) Passkeys(w http.ResponseWriter, r *http.Request) error {
	keys, err := h.admin.Passkeys(r.Context(), r.PathValue("name"))
	if err != nil {
		return asStatus(err)
	}
	return writeJSON(w, http.StatusOK, keys)
}

func (h *AdminAPI) DeletePasskey(w http.ResponseWriter, r *http.Request) error {
	if err := h.admin.RemovePasskey(r.Context(), r.PathValue("name"), r.PathValue("id")); err != nil {
		return asStatus(err)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *AdminAPI) Revoke(w http.ResponseWriter, r *http.Request) error {
	revoked, err := h.admin.Revoke(r.Context(), r.PathValue("name"))
	if err != nil {
		return asStatus(err)
	}
	return writeJSON(w, http.StatusOK, revoked)
}

func (h *AdminAPI) DeleteSessions(w http.ResponseWriter, r *http.Request) error {
	revoked, err := h.admin.SignOutEverywhere(r.Context(), r.PathValue("name"))
	if err != nil {
		return asStatus(err)
	}
	return writeJSON(w, http.StatusOK, revoked)
}

func (h *AdminAPI) Areas(w http.ResponseWriter, _ *http.Request) error {
	return writeJSON(w, http.StatusOK, h.admin.Areas())
}

// asStatus gives a failure its HTTP meaning.
//
// This is the whole of what the transport adds, and the only place here that
// knows about status codes: Admin says what kind of failure it was, and turning
// a kind into a number is a fact about HTTP rather than about accounts. Anything
// unrecognised is passed through and renders as a 500, which is the right answer
// for a failure nobody anticipated.
func asStatus(err error) error {
	switch {
	case errors.Is(err, ErrNoAccount):
		return errs.New("no such account", errs.HTTPStatus(http.StatusNotFound))
	case errors.Is(err, ErrInvalid):
		return errs.New(err.Error(), errs.HTTPStatus(http.StatusBadRequest))
	case errors.Is(err, ErrConflict), errors.Is(err, ErrExists):
		return errs.New(err.Error(), errs.HTTPStatus(http.StatusConflict))
	}
	return err
}

func writeJSON(w http.ResponseWriter, code int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(v)
}

func decode(r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return errs.New("could not read the request body: %v", err,
			errs.HTTPStatus(http.StatusBadRequest))
	}
	return nil
}
