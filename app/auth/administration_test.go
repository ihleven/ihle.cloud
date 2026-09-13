package auth

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/ihleven/ihlvn/app/cmsauth"
	"github.com/ihleven/ihlvn/app/db/dbtest"
	"github.com/interhome-group/cms/modules"
)

// The areas these tests name are defined in app/cmsauth, which a test binary for
// this package does not link. Registering them here is what makes
// "module.filme" a permission that resolves and "*" expand to something.
func init() {
	modules.Define("admin")
	modules.Define("filme")
}

// These run against a real Postgres for the same reason the store's tests do:
// what is being checked is the store's behaviour, including which lookups refuse
// a disabled account.
//
// They call Admin rather than the handlers, because what is under test is what
// may happen, not how a request is parsed. That separation is the point of the
// layer: no recorder, no request, no JSON.
func testAdmin(t *testing.T) (*Admin, *Store) {
	t.Helper()

	pool, ctx := dbtest.Pool(t)
	if _, err := pool.Exec(ctx, `truncate account restart identity cascade`); err != nil {
		t.Fatalf("truncating: %v", err)
	}

	store := NewStore(pool)
	return NewAdmin(store, "https://example.test", 15*time.Minute, 12), store
}

// account creates one and gives it the permissions named.
func account(t *testing.T, store *Store, name string, permissions ...string) *Account {
	t.Helper()

	a, err := store.CreateAccount(context.Background(), name, "Display "+name, name+"@example.test")
	if err != nil {
		t.Fatalf("creating %s: %v", name, err)
	}
	if len(permissions) > 0 {
		if err := store.SetCMSProfile(context.Background(), a.ID,
			CMSProfile{Permissions: permissions}); err != nil {
			t.Fatalf("granting %v to %s: %v", permissions, name, err)
		}
		reloaded, err := store.AccountByName(context.Background(), name)
		if err != nil {
			t.Fatalf("reloading %s: %v", name, err)
		}
		return reloaded
	}
	return a
}

// actingAs is the context an administrator's call carries, which is where the
// lock-out check reads who is acting.
func actingAs(caller *Account) context.Context {
	return WithAccount(context.Background(), caller)
}

func TestCreateThenGet(t *testing.T) {
	admin, store := testAdmin(t)
	ctx := actingAs(account(t, store, "admin", "*"))

	created, err := admin.Create(ctx, NewAccount{Name: "wolfgang", Email: "w@example.test"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// A new account can do nothing and sign in with nothing. Saying so is the
	// point of AccountInfo: it is the state the enrollment link then resolves.
	if created.HasPassword || created.Passkeys != 0 {
		t.Errorf("a new account already has credentials: %+v", created)
	}
	if len(created.Modules) != 0 {
		t.Errorf("Modules = %v, want none for an account granted nothing", created.Modules)
	}
	// Absent rather than null, so a client never distinguishes none from missing.
	if created.Groups == nil || created.Permissions == nil {
		t.Errorf("empty slices came back as nil: %+v", created)
	}

	got, err := admin.Get(ctx, "wolfgang")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != created.Name {
		t.Errorf("Get returned %q, want the account just created", got.Name)
	}
}

func TestCreateNeedsANameAndAnAddress(t *testing.T) {
	admin, store := testAdmin(t)
	ctx := actingAs(account(t, store, "admin", "*"))

	for _, in := range []NewAccount{
		{Email: "w@example.test"},
		{Name: "wolfgang"},
	} {
		if _, err := admin.Create(ctx, in); !errors.Is(err, ErrInvalid) {
			t.Errorf("Create(%+v) = %v, want ErrInvalid", in, err)
		}
	}
}

// AccountInfo's job is to say what permissions add up to. "*" is the case a
// client would otherwise have to reimplement.
func TestTheWildcardIsResolved(t *testing.T) {
	admin, store := testAdmin(t)
	ctx := actingAs(account(t, store, "admin", "*"))

	got, err := admin.Get(ctx, "admin")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(got.Modules) == 0 {
		t.Error("Modules is empty for an account holding \"*\"")
	}
	if len(got.Unregistered) != 0 {
		t.Errorf("Unregistered = %v; \"*\" is a grant-all, not an unknown name", got.Unregistered)
	}
}

// A disabled account is listed but, through every ordinary lookup, cannot be
// loaded — so without the administration lookup it could be seen and never
// turned back on.
func TestADisabledAccountCanBeReadAndReEnabled(t *testing.T) {
	admin, store := testAdmin(t)
	ctx := actingAs(account(t, store, "admin", "*"))
	target := account(t, store, "wolfgang")

	if _, err := store.SetDisabled(ctx, target.Name, true); err != nil {
		t.Fatalf("disabling: %v", err)
	}

	got, err := admin.Get(ctx, "wolfgang")
	if err != nil {
		t.Fatalf("a disabled account could not be read: %v", err)
	}
	if !got.Disabled {
		t.Error("the account does not report itself as disabled")
	}

	updated, err := admin.Update(ctx, "wolfgang", AccountEdit{
		Email:       "w@example.test",
		Permissions: []string{"module.filme"},
	})
	if err != nil {
		t.Fatalf("re-enabling: %v", err)
	}
	if updated.Disabled {
		t.Error("the account is still disabled after being enabled")
	}
	if !slices.Equal(updated.Modules, []string{"filme"}) {
		t.Errorf("Modules = %v, want [filme]", updated.Modules)
	}
}

// Disabling is the closest thing to deleting an account, so what comes back has
// to show the state that was actually reached rather than failing to re-read it.
func TestDisablingAnAccountAnswersWithIt(t *testing.T) {
	admin, store := testAdmin(t)
	ctx := actingAs(account(t, store, "admin", "*"))
	account(t, store, "wolfgang")

	got, err := admin.Update(ctx, "wolfgang", AccountEdit{Email: "w@example.test", Disabled: true})
	if err != nil {
		t.Fatalf("disabling: %v", err)
	}
	if !got.Disabled {
		t.Error("the result does not report the account as disabled")
	}
}

// Accounts cannot be deleted and admin is granted through here, so an
// administrator who removes their own entitlement has no way back except the
// terminal — which is what this exists to avoid needing.
func TestAnAdminCannotLockThemselvesOut(t *testing.T) {
	admin, store := testAdmin(t)
	caller := account(t, store, "admin", "*")
	ctx := actingAs(caller)

	for _, tc := range []struct {
		name string
		edit AccountEdit
	}{
		{"removing their own admin", AccountEdit{Email: "a@example.test", Permissions: []string{"module.filme"}}},
		{"disabling themselves", AccountEdit{Email: "a@example.test", Permissions: []string{"*"}, Disabled: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := admin.Update(ctx, "admin", tc.edit); !errors.Is(err, ErrConflict) {
				t.Fatalf("Update = %v, want ErrConflict", err)
			}

			// And nothing was written: the refusal happens before the writes.
			after, err := store.AccountByName(ctx, "admin")
			if err != nil {
				t.Fatalf("reloading: %v", err)
			}
			if !slices.Contains(cmsauth.Entitled(after), "admin") {
				t.Error("the admin entitlement was lost anyway")
			}
		})
	}
}

// Someone else's admin rights can still be taken away: the constraint is about
// not locking the door from the inside, not about protecting the role.
func TestAnotherAdminsRightsCanBeRemoved(t *testing.T) {
	admin, store := testAdmin(t)
	ctx := actingAs(account(t, store, "admin", "*"))
	account(t, store, "second", "module.admin")

	if _, err := admin.Update(ctx, "second", AccountEdit{Email: "s@example.test"}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	after, err := store.AccountByName(ctx, "second")
	if err != nil {
		t.Fatalf("reloading: %v", err)
	}
	if slices.Contains(cmsauth.Entitled(after), "admin") {
		t.Error("the second admin kept their entitlement")
	}
}

// The login name is what entries are owned by, so it is not editable — but the
// name shown and the address commits are signed with have to be fixable, since
// an account can never be deleted and recreated.
func TestIdentityCanBeCorrected(t *testing.T) {
	admin, store := testAdmin(t)
	ctx := actingAs(account(t, store, "admin", "*"))
	account(t, store, "wolfgang")

	if _, err := admin.Update(ctx, "wolfgang", AccountEdit{
		DisplayName: "Wolfgang Ihle",
		Email:       "wolfgang@ihle.cloud",
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	after, err := store.AccountByName(ctx, "wolfgang")
	if err != nil {
		t.Fatalf("reloading: %v", err)
	}
	if after.DisplayName != "Wolfgang Ihle" || after.Email != "wolfgang@ihle.cloud" {
		t.Errorf("identity = %q <%s>, want the corrected one", after.DisplayName, after.Email)
	}
	if after.Name != "wolfgang" {
		t.Errorf("the login name changed to %q", after.Name)
	}
}

func TestAnUnknownAccountIsNotFound(t *testing.T) {
	admin, store := testAdmin(t)
	ctx := actingAs(account(t, store, "admin", "*"))

	if _, err := admin.Get(ctx, "nobody"); !errors.Is(err, ErrNoAccount) {
		t.Errorf("Get(nobody) = %v, want ErrNoAccount", err)
	}
}

// An enrollment link cannot be issued for an account that is turned off: the
// person could not use it, and issuing one would look like it re-enabled them.
func TestADisabledAccountCannotBeEnrolled(t *testing.T) {
	admin, store := testAdmin(t)
	ctx := actingAs(account(t, store, "admin", "*"))
	target := account(t, store, "wolfgang")
	if _, err := store.SetDisabled(ctx, target.Name, true); err != nil {
		t.Fatalf("disabling: %v", err)
	}

	if _, err := admin.IssueEnrollment(ctx, "wolfgang"); !errors.Is(err, ErrConflict) {
		t.Errorf("IssueEnrollment = %v, want ErrConflict", err)
	}
}
