package auth

import (
	"context"
	"log/slog"
	"testing"

	"github.com/ihleven/ihlvn/pkg/password"
)

// These run against a real Postgres for the same reason the other store tests
// do: what is being checked is the database's — a transaction that either
// creates an account and consumes a row or does neither, and a unique
// constraint doing its job.

func geheimtippService(t *testing.T) (*Service, *Store, context.Context) {
	t.Helper()
	store, ctx := testStore(t)
	if _, err := store.pool.Exec(ctx, `truncate geheimtipp_migration`); err != nil {
		t.Fatalf("truncating geheimtipp_migration: %v", err)
	}
	svc := &Service{
		store: store,
		cfg:   Config{AdoptGeheimtipp: true},
		log:   slog.New(slog.DiscardHandler),
	}
	return svc, store, ctx
}

func addPlayer(t *testing.T, s *Store, ctx context.Context, login, passwd, vorname, nachname, email string) {
	t.Helper()
	_, err := s.pool.Exec(ctx, `
		insert into geheimtipp_migration (login, passwd, vorname, nachname, email)
		values ($1, $2, $3, $4, $5)`, login, passwd, vorname, nachname, email)
	if err != nil {
		t.Fatalf("adding player %s: %v", login, err)
	}
}

// The whole point: someone types what they have always typed and is signed in,
// with an account created behind them.
func TestFirstSignInCreatesAConfinedAccount(t *testing.T) {
	svc, store, ctx := geheimtippService(t)
	addPlayer(t, store, ctx, "pauli", "geheim", "Paul", "Auli", "pauli@example.org")

	account, err := svc.authenticatePassword(ctx, "pauli", "geheim")
	if err != nil {
		t.Fatalf("signing in: %v", err)
	}

	if account.Name != "pauli" {
		t.Errorf("name = %q, want pauli — the login is what entries are owned by", account.Name)
	}
	if !account.Confined() {
		t.Error("the account is not confined to the pool")
	}
	if account.DisplayName != "Paul Auli" {
		t.Errorf("display name = %q, want %q", account.DisplayName, "Paul Auli")
	}
	// The password is now this app's, hashed. The pool still holds its own copy
	// and the two are free to diverge from here.
	if !account.HasPassword() {
		t.Error("the account has no password")
	}

	// A confined account is not governed by entitlements, so it gets no
	// permission row at all.
	if len(account.CMS.Permissions) != 0 {
		t.Errorf("permissions = %v, want none", account.CMS.Permissions)
	}

	n, err := store.PendingGeheimtipper(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("%d players still pending, want 0 — the row should be consumed", n)
	}
}

// The second sign-in must take the account path. If it fell through again the
// migration row would be gone and nobody could sign in twice.
func TestSecondSignInUsesTheAccount(t *testing.T) {
	svc, store, ctx := geheimtippService(t)
	addPlayer(t, store, ctx, "pauli", "geheim", "Paul", "Auli", "pauli@example.org")

	first, err := svc.authenticatePassword(ctx, "pauli", "geheim")
	if err != nil {
		t.Fatalf("first sign-in: %v", err)
	}
	second, err := svc.authenticatePassword(ctx, "pauli", "geheim")
	if err != nil {
		t.Fatalf("second sign-in: %v", err)
	}
	if first.ID != second.ID {
		t.Errorf("second sign-in made account %d, want %d", second.ID, first.ID)
	}
}

// The security property this design rests on. Pool passwords are short words
// chosen long ago, and at least one pool login is also an account here with
// wider rights. An existing account must be unreachable with one.
func TestAPoolPasswordCannotOpenAnExistingAccount(t *testing.T) {
	svc, store, ctx := geheimtippService(t)

	hash, err := password.Hash("the-real-one")
	if err != nil {
		t.Fatal(err)
	}
	existing, err := store.CreateAccount(ctx, "matt", "Matt", "matt@example.org")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SetPasswordHash(ctx, existing.ID, hash); err != nil {
		t.Fatal(err)
	}
	// The same login exists upstream with a weaker password.
	addPlayer(t, store, ctx, "matt", "mehmet", "Matt", "Ihle", "other@example.org")

	if _, err := svc.authenticatePassword(ctx, "matt", "mehmet"); err == nil {
		t.Fatal("the pool password opened an existing account")
	}
	if _, err := svc.authenticatePassword(ctx, "matt", "the-real-one"); err != nil {
		t.Errorf("the account's own password stopped working: %v", err)
	}
	// And the fallback was never reached, so the row is untouched.
	n, err := store.PendingGeheimtipper(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("%d players pending, want 1 — the row must not have been consumed", n)
	}
}

// Pool addresses are not verified, so claiming an account because the addresses
// match would be a way into it.
func TestAdoptionRefusesAnAddressThatIsAlreadyAnAccount(t *testing.T) {
	svc, store, ctx := geheimtippService(t)
	if _, err := store.CreateAccount(ctx, "matt", "Matt", "shared@example.org"); err != nil {
		t.Fatal(err)
	}
	addPlayer(t, store, ctx, "pauli", "geheim", "Paul", "Auli", "SHARED@example.org")

	if _, err := svc.authenticatePassword(ctx, "pauli", "geheim"); err == nil {
		t.Error("adoption claimed an address that already belongs to an account")
	}
}

// account.email is unique and is what a commit is signed with, so there is no
// placeholder worth inventing — a player with no address cannot be adopted.
//
// The database refuses the row rather than leaving it to fail at sign-in. It
// used to be the other way round, and the failure surfaced as "invalid
// credentials", which is indistinguishable from a wrong password: whoever added
// the row had no way to tell what they had done. Adoption still refuses such a
// person, for rows added before that constraint existed.
func TestAPlayerWithNoAddressCannotBeAdded(t *testing.T) {
	_, store, ctx := geheimtippService(t)

	_, err := store.pool.Exec(ctx, `
		insert into geheimtipp_migration (login, passwd, vorname, nachname, email)
		values ('sandra', 'geheim', 'Sandra', '', '')`)
	if err == nil {
		t.Fatal("a player with no address was added; adoption would refuse them at sign-in")
	}
}

// The same for a row with no password: it can never produce an account, so it
// is not a row worth having.
func TestAPlayerWithNoPasswordCannotBeAdded(t *testing.T) {
	_, store, ctx := geheimtippService(t)

	_, err := store.pool.Exec(ctx, `
		insert into geheimtipp_migration (login, passwd, email)
		values ('sandra', '', 'sandra@example.org')`)
	if err == nil {
		t.Fatal("a player with no password was added")
	}
}

// Turning the fallback off is the step before dropping the table. After it,
// only people who already came through can sign in.
func TestAdoptionCanBeSwitchedOff(t *testing.T) {
	svc, store, ctx := geheimtippService(t)
	svc.cfg.AdoptGeheimtipp = false
	addPlayer(t, store, ctx, "pauli", "geheim", "Paul", "Auli", "pauli@example.org")

	if _, err := svc.authenticatePassword(ctx, "pauli", "geheim"); err == nil {
		t.Error("the fallback ran with adoption switched off")
	}
}

// A name with no account and no player is the ordinary wrong-username case.
func TestUnknownNameStillFails(t *testing.T) {
	svc, _, ctx := geheimtippService(t)
	if _, err := svc.authenticatePassword(ctx, "nobody", "whatever"); err == nil {
		t.Error("an unknown name signed in")
	}
}
