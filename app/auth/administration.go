package auth

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/ihleven/ihlvn/app/cmsauth"
	"github.com/ihleven/ihlvn/pkg/password"
	"github.com/interhome-group/cms/modules"
)

// Administration: the operations `ihlvn account` performs from a terminal, as
// functions rather than as commands or routes.
//
// Nothing here knows about HTTP. A caller passes a context and values and gets
// values back, so the same rules serve the admin section, the terminal, and a
// test that wants neither. Failures are reported as errors that say what kind of
// failure they are; turning a kind into a status code is the transport's job,
// and only the transport has an opinion about status codes.

// The kinds of failure a caller can do something about. Anything else is a
// fault in the system rather than in the request, and is returned unwrapped.
var (
	// ErrInvalid is a request that cannot be carried out as asked.
	ErrInvalid = errors.New("auth: invalid")
	// ErrConflict is a request that makes sense but not now, or not here: an
	// account that already exists, a change that would lock its author out.
	ErrConflict = errors.New("auth: conflict")
)

// fault carries the sentence to show a person and the kind that classifies it.
// Error() is the sentence alone, so a caller can pass it on without unwrapping
// the plumbing, while errors.Is still answers what kind of failure it was.
type fault struct {
	kind    error
	message string
}

func (f *fault) Error() string { return f.message }
func (f *fault) Unwrap() error { return f.kind }

func invalid(format string, args ...any) error {
	return &fault{kind: ErrInvalid, message: fmt.Sprintf(format, args...)}
}

func conflict(format string, args ...any) error {
	return &fault{kind: ErrConflict, message: fmt.Sprintf(format, args...)}
}

// Admin performs account administration.
type Admin struct {
	store *Store

	// origin is the public base an enrollment link is offered at. It is passed
	// in rather than read off a request, for the same reason the rest of the app
	// derives from PUBLIC_URL: behind a proxy the request describes the local
	// hop, not what the browser will be asked to open.
	origin    string
	enrollTTL time.Duration

	// breaches is the password screening. Held so a test can point it at a stub
	// instead of the public service.
	breaches *password.BreachChecker

	// minPasswordLength is where "this is short" starts. Passed in rather than
	// decided here, so the terminal and the browser nag at the same point.
	minPasswordLength int
}

func NewAdmin(store *Store, origin string, enrollTTL time.Duration, minPasswordLength int) *Admin {
	return &Admin{
		store: store, origin: origin, enrollTTL: enrollTTL,
		breaches:          &password.BreachChecker{},
		minPasswordLength: minPasswordLength,
	}
}

// NewAccount is what creating one needs.
type NewAccount struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

// AccountEdit is everything about an account that can be changed at once.
//
// Being disabled is part of it rather than a call of its own: it is a property
// of the account, and whatever edits an account edits all of it. The login name
// is absent because entries are owned by it.
type AccountEdit struct {
	DisplayName string   `json:"display_name"`
	Email       string   `json:"email"`
	Groups      []string `json:"groups"`
	Permissions []string `json:"permissions"`
	Disabled    bool     `json:"disabled"`
}

// List answers with every account, including disabled ones: an administrator
// needs to see the account they turned off in order to turn it back on.
func (a *Admin) List(ctx context.Context) ([]AccountInfo, error) {
	accounts, err := a.store.ListAccounts(ctx)
	if err != nil {
		return nil, err
	}

	infos := make([]AccountInfo, 0, len(accounts))
	for _, account := range accounts {
		info, err := a.present(ctx, account)
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}
	return infos, nil
}

// Get answers with one account.
func (a *Admin) Get(ctx context.Context, name string) (AccountInfo, error) {
	account, err := a.account(ctx, name)
	if err != nil {
		return AccountInfo{}, err
	}
	return a.present(ctx, account)
}

// Create adds an account. It has no way to sign in yet: a password or a passkey
// is a separate step, which is what the enrollment link is for.
func (a *Admin) Create(ctx context.Context, in NewAccount) (AccountInfo, error) {
	if in.Name == "" {
		return AccountInfo{}, invalid("an account needs a login name")
	}
	if in.Email == "" {
		return AccountInfo{}, invalid("an account needs an email address")
	}
	if in.DisplayName == "" {
		in.DisplayName = in.Name
	}

	account, err := a.store.CreateAccount(ctx, in.Name, in.DisplayName, in.Email)
	if errors.Is(err, ErrExists) {
		return AccountInfo{}, conflict("an account with that name or address already exists")
	}
	if err != nil {
		return AccountInfo{}, err
	}
	return a.present(ctx, account)
}

// Update replaces the mutable half of an account.
//
// The writes are not one transaction — identity and disabled live on account,
// groups and permissions on cmsauth — so a failure part way leaves a partly
// applied edit. It is visible in what comes back and can be repeated, which is
// the cheaper answer here than threading a transaction through the store for a
// form that one person submits.
func (a *Admin) Update(ctx context.Context, name string, edit AccountEdit) (AccountInfo, error) {
	account, err := a.account(ctx, name)
	if err != nil {
		return AccountInfo{}, err
	}
	if edit.DisplayName == "" {
		edit.DisplayName = account.Name
	}
	if edit.Email == "" {
		return AccountInfo{}, invalid("an account needs an email address")
	}
	if err := a.refuseLockout(ctx, account, edit); err != nil {
		return AccountInfo{}, err
	}

	if err := a.store.SetIdentity(ctx, account.ID, edit.DisplayName, edit.Email); err != nil {
		return AccountInfo{}, err
	}
	if err := a.store.SetCMSProfile(ctx, account.ID,
		CMSProfile{Groups: edit.Groups, Permissions: edit.Permissions}); err != nil {
		return AccountInfo{}, err
	}
	if edit.Disabled != account.Disabled {
		if _, err := a.store.SetDisabled(ctx, account.Name, edit.Disabled); err != nil {
			return AccountInfo{}, err
		}
		// Loading refuses a disabled account, but an existing session would go
		// on working until it expired. Dropping them makes the refusal immediate.
		if edit.Disabled {
			if _, err := a.store.DeleteSessions(ctx, account.ID); err != nil {
				return AccountInfo{}, err
			}
		}
	}

	// Re-read rather than patching the loaded copy, so what comes back is what
	// the database now holds and a partly applied edit shows as what it is.
	updated, err := a.store.DisabledAccountByName(ctx, account.Name)
	if err != nil {
		return AccountInfo{}, err
	}
	return a.present(ctx, updated)
}

// refuseLockout stops an administrator removing their own admin rights or
// disabling themselves.
//
// Accounts cannot be deleted and admin is granted through here, so an
// administrator who revokes their own entitlement has no way back except the
// terminal — which is the thing this exists to avoid needing. Someone else's
// admin rights can still be removed: the constraint is about not locking the
// door from the inside, not about protecting the role.
//
// The caller is read from the context, which is where whoever is acting is
// carried. A call with nobody in the context — the CLI — is nobody's own
// account and passes.
func (a *Admin) refuseLockout(ctx context.Context, target *Account, edit AccountEdit) error {
	caller, ok := FromContext(ctx)
	if !ok || caller.ID != target.ID {
		return nil
	}

	if edit.Disabled {
		return conflict("an administrator cannot disable their own account")
	}
	after := &Account{CMS: CMSProfile{Groups: edit.Groups, Permissions: edit.Permissions}}
	if !cmsauth.May(after, cmsauth.Admin) {
		return conflict("an administrator cannot remove their own admin entitlement")
	}
	return nil
}

// SetPassword sets or replaces an account's password.
//
// It exists for recovery, not for onboarding: a password an administrator
// chooses is one they know and have to convey somehow. The enrollment link is
// the better path for a new account, because the person picks their own and it
// never travels.
//
// A password the screening objects to is not set unless confirm says to. The
// check is advice, not a verdict — refusing outright would substitute a number
// and a third party's corpus for the operator's judgement — but overriding it
// should be a second, deliberate act.
func (a *Admin) SetPassword(ctx context.Context, name, plain string, confirm bool) (PasswordResult, error) {
	account, err := a.account(ctx, name)
	if err != nil {
		return PasswordResult{}, err
	}
	if plain == "" {
		return PasswordResult{}, invalid("no password given")
	}
	// The one hard bound, and the only refusal that cannot be overridden: it is
	// a technical limit rather than an opinion about the password.
	if password.TooLong(plain) {
		return PasswordResult{}, invalid("use at most %d characters", password.MaxLength)
	}

	advice := password.Check(ctx, a.breaches, plain, a.minPasswordLength)
	if advice.Concerning() && !confirm {
		return PasswordResult{Set: false, Advice: advice}, nil
	}

	hash, err := password.Hash(plain)
	if err != nil {
		return PasswordResult{}, err
	}
	if err := a.store.SetPasswordHash(ctx, account.ID, hash); err != nil {
		return PasswordResult{}, err
	}
	return PasswordResult{Set: true, Advice: advice}, nil
}

// IssueEnrollment creates a single-use link for registering a passkey.
//
// The link is returned rather than sent: this app has no reliable way to reach
// someone, and an administrator handing it over in person or in a chat is both
// honest about that and better than pretending an email was delivered.
//
// Named for the direction it goes: Service.Enroll *redeems* a link, and two
// operations called Enroll that undo each other is a trap.
func (a *Admin) IssueEnrollment(ctx context.Context, name string) (Enrollment, error) {
	account, err := a.account(ctx, name)
	if err != nil {
		return Enrollment{}, err
	}
	if account.Disabled {
		return Enrollment{}, conflict("%s is disabled; enable it before enrolling a device", account.Name)
	}
	// A link is possession; the password is knowledge. Registering a credential
	// needs both, so a link for an account that cannot prove itself would be a
	// single factor that hands the account to whoever intercepts it.
	if !account.HasPassword() {
		return Enrollment{}, conflict("%s has no password yet; set one first, and convey it separately from the link", account.Name)
	}

	token, err := a.store.CreateEnrollToken(ctx, account.ID, a.enrollTTL)
	if err != nil {
		return Enrollment{}, err
	}
	return Enrollment{
		URL:       EnrollURL(a.origin, token),
		ExpiresAt: time.Now().Add(a.enrollTTL),
	}, nil
}

// Passkeys lists an account's registered devices — the same projection the
// self-service page gets, because it is the same question asked about someone
// else's account.
func (a *Admin) Passkeys(ctx context.Context, name string) ([]PasskeyInfo, error) {
	account, err := a.account(ctx, name)
	if err != nil {
		return nil, err
	}
	keys, err := a.store.ListPasskeys(ctx, account.ID)
	if err != nil {
		return nil, err
	}

	infos := make([]PasskeyInfo, 0, len(keys))
	for _, k := range keys {
		infos = append(infos, k.Info())
	}
	return infos, nil
}

// RemovePasskey removes one device, for the ordinary case of a phone that was
// lost or replaced.
func (a *Admin) RemovePasskey(ctx context.Context, name, id string) error {
	account, err := a.account(ctx, name)
	if err != nil {
		return err
	}
	raw, err := base64.RawURLEncoding.DecodeString(id)
	if err != nil {
		return invalid("that is not a passkey id")
	}
	return a.store.DeletePasskey(ctx, account.ID, raw)
}

// Revoked counts what was taken away.
type Revoked struct {
	Passkeys int64 `json:"passkeys"`
	Sessions int64 `json:"sessions"`
}

// Revoke removes every passkey and drops every session: the "this account is
// compromised" action, kept apart from an ordinary edit so it cannot happen by
// mistake.
//
// The password is deliberately left alone. Revoking is about the credentials
// that were taken; clearing the password as well would leave an account that
// cannot be recovered without a second administrator.
func (a *Admin) Revoke(ctx context.Context, name string) (Revoked, error) {
	account, err := a.account(ctx, name)
	if err != nil {
		return Revoked{}, err
	}
	keys, err := a.store.DeletePasskeys(ctx, account.ID)
	if err != nil {
		return Revoked{}, err
	}
	sessions, err := a.store.DeleteSessions(ctx, account.ID)
	if err != nil {
		return Revoked{}, err
	}
	return Revoked{Passkeys: keys, Sessions: sessions}, nil
}

// SignOutEverywhere ends every session without touching what the account can
// sign in with.
func (a *Admin) SignOutEverywhere(ctx context.Context, name string) (Revoked, error) {
	account, err := a.account(ctx, name)
	if err != nil {
		return Revoked{}, err
	}
	n, err := a.store.DeleteSessions(ctx, account.ID)
	if err != nil {
		return Revoked{}, err
	}
	return Revoked{Sessions: n}, nil
}

// Areas are the feature areas this build defines, which is what may be granted.
//
// Read from the registry rather than listed by whoever is granting, so a picker
// cannot offer a name that resolves to nothing.
func (a *Admin) Areas() []string {
	return modules.RegisteredModules()
}

// account resolves a name, deliberately through the lookup that sees disabled
// accounts: administration is where one gets turned back on.
func (a *Admin) account(ctx context.Context, name string) (*Account, error) {
	if name == "" {
		return nil, invalid("no account named")
	}
	account, err := a.store.DisabledAccountByName(ctx, name)
	if errors.Is(err, ErrNoAccount) {
		return nil, ErrNoAccount
	}
	return account, err
}

// present projects an account, counting passkeys with a query per account. That
// is an N+1, and deliberate: this is a family archive with a handful of
// accounts, and the CLI's listing does the same.
func (a *Admin) present(ctx context.Context, account *Account) (AccountInfo, error) {
	keys, err := a.store.ListPasskeys(ctx, account.ID)
	if err != nil {
		return AccountInfo{}, err
	}
	return AccountInfo{
		Name:         account.Name,
		DisplayName:  account.DisplayName,
		Email:        account.Email,
		Disabled:     account.Disabled,
		Groups:       orEmpty(account.CMS.Groups),
		Permissions:  orEmpty(account.CMS.Permissions),
		Unregistered: orEmpty(cmsauth.Unregistered(account.CMS.Permissions)),
		Modules:      orEmpty(cmsauth.Entitled(account)),
		HasPassword:  account.HasPassword(),
		Passkeys:     len(keys),
		CreatedAt:    account.CreatedAt,
	}, nil
}

// orEmpty renders a nil slice as [] rather than null, so a client never has to
// distinguish "none" from "absent".
func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
