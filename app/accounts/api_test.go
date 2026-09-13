package accounts

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ihleven/ihlvn/pkg/authn"
	"github.com/interhome-group/cms/modules"
	"github.com/interhome-group/cms/pkg/errs"
	"github.com/jackc/pgx/v5/pgxpool"
)

// The areas these tests name are defined in package main, which a test binary
// for this package does not link. Registering them here is what makes
// "module.filme" a permission that resolves and "*" expand to something.
func init() {
	modules.Define("admin")
	modules.Define("filme")
}

// These run against a real Postgres for the same reason pkg/authn's do: what is
// being checked is the store's behaviour, including which lookups refuse a
// disabled account.
//
// The database is truncated between tests, so it must be one that may be
// destroyed — see requireDisposable, which enforces the same rule there.
func testAPI(t *testing.T) (*API, *authn.Store) {
	t.Helper()

	url := os.Getenv("IHLVN_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("IHLVN_TEST_DATABASE_URL not set")
	}
	if !strings.Contains(strings.ToLower(url), "test") {
		t.Fatalf("IHLVN_TEST_DATABASE_URL points at %q, which is not named as a test "+
			"database; these tests truncate every table", url)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := authn.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrating: %v", err)
	}
	holdTheDatabase(t, pool)
	if _, err := pool.Exec(ctx, `truncate account restart identity cascade`); err != nil {
		t.Fatalf("truncating: %v", err)
	}

	store := authn.NewStore(pool)
	return New(store, "https://example.test", 15*time.Minute), store
}

// account creates one and gives it the permissions named.
func account(t *testing.T, store *authn.Store, name string, permissions ...string) *authn.Account {
	t.Helper()

	a, err := store.CreateAccount(context.Background(), name, "Display "+name, name+"@example.test")
	if err != nil {
		t.Fatalf("creating %s: %v", name, err)
	}
	if len(permissions) > 0 {
		if err := store.SetCMSProfile(context.Background(), a.ID,
			authn.CMSProfile{Permissions: permissions}); err != nil {
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

// as builds a request carrying the account, which is what requireAdmin puts in
// the context before a handler runs.
func as(caller *authn.Account, method, target string, body string) *http.Request {
	r := httptest.NewRequest(method, target, strings.NewReader(body))
	if caller != nil {
		r = r.WithContext(authn.WithAccount(r.Context(), caller))
	}
	return r
}

func TestCreateThenGet(t *testing.T) {
	api, store := testAPI(t)
	admin := account(t, store, "admin", "*")

	w := httptest.NewRecorder()
	r := as(admin, http.MethodPost, "/", `{"name":"wolfgang","email":"w@example.test"}`)
	if err := api.Create(w, r); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", w.Code)
	}

	var created view
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	// A new account can do nothing and sign in with nothing. Saying so is the
	// point of the view: it is the state the enrollment link then resolves.
	if created.HasPassword || created.Passkeys != 0 {
		t.Errorf("a new account already has credentials: %+v", created)
	}
	if len(created.Modules) != 0 {
		t.Errorf("Modules = %v, want none for an account granted nothing", created.Modules)
	}
	// Absent rather than null, so the client never distinguishes none from
	// missing.
	if created.Groups == nil || created.Permissions == nil {
		t.Errorf("empty slices came back as null: %+v", created)
	}
}

// The view's job is to say what permissions add up to. "*" is the case a client
// would otherwise have to reimplement.
func TestTheViewResolvesTheWildcard(t *testing.T) {
	api, store := testAPI(t)
	admin := account(t, store, "admin", "*")

	w := httptest.NewRecorder()
	r := as(admin, http.MethodGet, "/", "")
	r.SetPathValue("name", "admin")
	if err := api.Get(w, r); err != nil {
		t.Fatalf("Get: %v", err)
	}

	var got view
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decoding: %v", err)
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
	api, store := testAPI(t)
	admin := account(t, store, "admin", "*")
	target := account(t, store, "wolfgang")

	if _, err := store.SetDisabled(context.Background(), target.Name, true); err != nil {
		t.Fatalf("disabling: %v", err)
	}

	w := httptest.NewRecorder()
	r := as(admin, http.MethodGet, "/", "")
	r.SetPathValue("name", "wolfgang")
	if err := api.Get(w, r); err != nil {
		t.Fatalf("a disabled account could not be read: %v", err)
	}
	var got view
	json.Unmarshal(w.Body.Bytes(), &got)
	if !got.Disabled {
		t.Error("the account does not report itself as disabled")
	}

	w = httptest.NewRecorder()
	r = as(admin, http.MethodPut, "/",
		`{"email":"w@example.test","disabled":false,"permissions":["module.filme"]}`)
	r.SetPathValue("name", "wolfgang")
	if err := api.Update(w, r); err != nil {
		t.Fatalf("re-enabling: %v", err)
	}

	var updated view
	if err := json.Unmarshal(w.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if updated.Disabled {
		t.Error("the account is still disabled after being enabled")
	}
	if len(updated.Modules) != 1 || updated.Modules[0] != "filme" {
		t.Errorf("Modules = %v, want [filme]", updated.Modules)
	}
}

// Disabling is the closest thing to deleting an account, so the response has to
// show the state that was actually reached rather than failing to re-read it.
func TestDisablingAnAccountAnswersWithIt(t *testing.T) {
	api, store := testAPI(t)
	admin := account(t, store, "admin", "*")
	account(t, store, "wolfgang")

	w := httptest.NewRecorder()
	r := as(admin, http.MethodPut, "/", `{"email":"w@example.test","disabled":true}`)
	r.SetPathValue("name", "wolfgang")
	if err := api.Update(w, r); err != nil {
		t.Fatalf("disabling: %v", err)
	}

	var got view
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if !got.Disabled {
		t.Error("the response does not report the account as disabled")
	}
}

// Accounts cannot be deleted and admin is granted through this API, so an
// administrator who removes their own entitlement has no way back except the
// terminal — which is what this section exists to avoid needing.
func TestAnAdminCannotLockThemselvesOut(t *testing.T) {
	api, store := testAPI(t)
	admin := account(t, store, "admin", "*")

	for _, tc := range []struct{ name, body string }{
		{"removing their own admin", `{"email":"a@example.test","permissions":["module.filme"]}`},
		{"disabling themselves", `{"email":"a@example.test","permissions":["*"],"disabled":true}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := as(admin, http.MethodPut, "/", tc.body)
			r.SetPathValue("name", "admin")
			if err := api.Update(w, r); err == nil {
				t.Fatal("the edit was accepted")
			}

			// And nothing was written: the refusal happens before the writes.
			after, err := store.AccountByName(context.Background(), "admin")
			if err != nil {
				t.Fatalf("reloading: %v", err)
			}
			if !slices.Contains(after.EntitledModules(), "admin") {
				t.Error("the admin entitlement was lost anyway")
			}
		})
	}
}

// Someone else's admin rights can still be taken away: the constraint is about
// not locking the door from the inside, not about protecting the role.
func TestAnotherAdminsRightsCanBeRemoved(t *testing.T) {
	api, store := testAPI(t)
	admin := account(t, store, "admin", "*")
	account(t, store, "second", "module.admin")

	w := httptest.NewRecorder()
	r := as(admin, http.MethodPut, "/", `{"email":"s@example.test","permissions":[]}`)
	r.SetPathValue("name", "second")
	if err := api.Update(w, r); err != nil {
		t.Fatalf("Update: %v", err)
	}

	after, err := store.AccountByName(context.Background(), "second")
	if err != nil {
		t.Fatalf("reloading: %v", err)
	}
	if slices.Contains(after.EntitledModules(), "admin") {
		t.Error("the second admin kept their entitlement")
	}
}

// The login name is what entries are owned by, so it is not editable — but the
// name shown and the address commits are signed with have to be fixable, since
// an account can never be deleted and recreated.
func TestIdentityCanBeCorrected(t *testing.T) {
	api, store := testAPI(t)
	admin := account(t, store, "admin", "*")
	account(t, store, "wolfgang")

	w := httptest.NewRecorder()
	r := as(admin, http.MethodPut, "/",
		`{"display_name":"Wolfgang Ihle","email":"wolfgang@ihle.cloud"}`)
	r.SetPathValue("name", "wolfgang")
	if err := api.Update(w, r); err != nil {
		t.Fatalf("Update: %v", err)
	}

	after, err := store.AccountByName(context.Background(), "wolfgang")
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
	api, store := testAPI(t)
	admin := account(t, store, "admin", "*")

	w := httptest.NewRecorder()
	r := as(admin, http.MethodGet, "/", "")
	r.SetPathValue("name", "nobody")

	err := api.Get(w, r)
	if err == nil {
		t.Fatal("an unknown account was found")
	}
	if status := errs.StatusOf(err); status != http.StatusNotFound {
		t.Errorf("status = %d, want 404", status)
	}
}

// dbLockKey serialises the test binaries that truncate this database. Every
// package that does so takes the same lock; pkg/authn holds the other copy of
// this constant.
const dbLockKey = 1974

// holdTheDatabase takes the truncation lock for the duration of the test.
//
// Under `go test ./...` the package binaries run at the same time, and each of
// these tests begins by truncating. Without a lock one package wipes another's
// fixtures mid-test, which surfaces as an unrelated test failing to find an
// account it had just created. The lock lives in Postgres, so it coordinates
// across processes without the build having to know anything about it.
func holdTheDatabase(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()
	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquiring a connection for the lock: %v", err)
	}
	if _, err := conn.Exec(ctx, `select pg_advisory_lock($1)`, dbLockKey); err != nil {
		conn.Release()
		t.Fatalf("taking the lock: %v", err)
	}
	t.Cleanup(func() {
		if _, err := conn.Exec(context.Background(), `select pg_advisory_unlock($1)`, dbLockKey); err != nil {
			t.Errorf("releasing the lock: %v", err)
		}
		conn.Release()
	})
}
