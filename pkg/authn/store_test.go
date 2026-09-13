package authn

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// These run against a real Postgres because the guarantees being checked are
// the database's: that redeeming an enrollment token is atomic, that a
// challenge cannot be replayed, that a session disappears when it is deleted.
// A stub would only prove that the stub behaves.
//
// Set IHLVN_TEST_DATABASE_URL to a database that may be destroyed: this helper
// truncates, so the name has to say so.
func testStore(t *testing.T) (*Store, context.Context) {
	t.Helper()

	url := os.Getenv("IHLVN_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("IHLVN_TEST_DATABASE_URL not set")
	}
	requireDisposable(t, url)
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("pinging: %v", err)
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("migrating: %v", err)
	}
	if _, err := pool.Exec(ctx, `truncate account restart identity cascade`); err != nil {
		t.Fatalf("truncating: %v", err)
	}
	return NewStore(pool), ctx
}

// requireDisposable refuses a database whose name does not mark it as one that
// may be destroyed.
//
// The truncate below is cascading and irreversible, and an environment variable
// carries no hint of what it is pointing at. Aiming this at a development
// database once cost four accounts, their passkeys and their permissions — and
// the emptied tables then read as "there was never an account here", which is a
// worse failure than an error, because it looks like a finding.
func requireDisposable(t *testing.T, rawURL string) {
	t.Helper()

	name := databaseName(rawURL)
	if !strings.Contains(strings.ToLower(name), "test") {
		t.Fatalf("IHLVN_TEST_DATABASE_URL points at %q, which is not named as a test database.\n"+
			"These tests truncate every table. Use one that may be destroyed:\n"+
			"    createdb ihlvn_test\n"+
			"    IHLVN_TEST_DATABASE_URL=postgres:///ihlvn_test go test ./pkg/authn/", name)
	}
}

// databaseName is the database a connection string names, in either the URL or
// the keyword form. An unreadable string yields "", which the guard refuses.
func databaseName(rawURL string) string {
	if u, err := url.Parse(rawURL); err == nil && u.Scheme != "" {
		return strings.TrimPrefix(u.Path, "/")
	}
	for _, field := range strings.Fields(rawURL) {
		if rest, ok := strings.CutPrefix(field, "dbname="); ok {
			return rest
		}
	}
	return ""
}

func mustAccount(t *testing.T, s *Store, ctx context.Context, name string) *Account {
	t.Helper()
	a, err := s.CreateAccount(ctx, name, "Display "+name, name+"@example.test")
	if err != nil {
		t.Fatalf("creating account %s: %v", name, err)
	}
	return a
}

func TestMigrateIsIdempotent(t *testing.T) {
	s, ctx := testStore(t)
	if err := Migrate(ctx, s.pool); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
}

// An applied migration whose file later changes is drift between environments,
// so it is refused rather than ignored.
func TestMigrateRefusesAnEditedMigration(t *testing.T) {
	s, ctx := testStore(t)

	const version = "0001_baseline.sql"

	var original []byte
	if err := s.pool.QueryRow(ctx,
		`select checksum from schema_migrations where version = $1`, version).Scan(&original); err != nil {
		t.Fatalf("reading the recorded checksum: %v", err)
	}
	// The database outlives this test, so put the real checksum back or every
	// later test fails at Migrate.
	t.Cleanup(func() {
		if _, err := s.pool.Exec(context.Background(),
			`update schema_migrations set checksum = $1 where version = $2`, original, version); err != nil {
			t.Fatalf("restoring the checksum: %v", err)
		}
	})

	if _, err := s.pool.Exec(ctx,
		`update schema_migrations set checksum = $1 where version = $2`,
		[]byte("not the checksum this file has"), version); err != nil {
		t.Fatalf("tampering with the checksum: %v", err)
	}
	if err := Migrate(ctx, s.pool); err == nil {
		t.Fatal("a changed migration was accepted")
	}
}

func TestCreateAndLoadAccount(t *testing.T) {
	s, ctx := testStore(t)
	created := mustAccount(t, s, ctx, "matt")

	if len(created.Handle) != 32 {
		t.Errorf("handle is %d bytes, want 32", len(created.Handle))
	}

	got, err := s.AccountByName(ctx, "matt")
	if err != nil {
		t.Fatalf("AccountByName: %v", err)
	}
	if got.ID != created.ID || got.Email != created.Email {
		t.Errorf("round trip changed the account: %+v vs %+v", got, created)
	}
	if got.HasPassword() {
		t.Error("a new account should have no password until one is set")
	}

	byHandle, err := s.AccountByHandle(ctx, created.Handle)
	if err != nil || byHandle.ID != created.ID {
		t.Errorf("AccountByHandle: %v, %+v", err, byHandle)
	}

	if _, err := s.AccountByName(ctx, "nobody"); !errors.Is(err, ErrNoAccount) {
		t.Errorf("unknown account: got %v, want ErrNoAccount", err)
	}
}

func TestCreateAccountRejectsDuplicateName(t *testing.T) {
	s, ctx := testStore(t)
	mustAccount(t, s, ctx, "matt")

	_, err := s.CreateAccount(ctx, "matt", "Someone Else", "other@example.test")
	if !errors.Is(err, ErrExists) {
		t.Errorf("duplicate name: got %v, want ErrExists", err)
	}
}

// Disabling is refused at the point of loading, so it takes effect everywhere at
// once rather than needing a check at each call site.
func TestDisabledAccountCannotBeLoaded(t *testing.T) {
	s, ctx := testStore(t)
	mustAccount(t, s, ctx, "matt")

	if _, err := s.SetDisabled(ctx, "matt", true); err != nil {
		t.Fatalf("SetDisabled: %v", err)
	}
	if _, err := s.AccountByName(ctx, "matt"); !errors.Is(err, ErrNoAccount) {
		t.Errorf("disabled account: got %v, want ErrNoAccount", err)
	}

	// ListAccounts is the one reader that still shows it, for the admin commands.
	all, err := s.ListAccounts(ctx)
	if err != nil || len(all) != 1 || !all[0].Disabled {
		t.Errorf("ListAccounts should show the disabled account: %v, %+v", err, all)
	}
}

// An account with no cmsauth row is a valid account holding no CMS rights. The
// old inner join made such an account invisible instead.
func TestAccountWithoutCMSProfile(t *testing.T) {
	s, ctx := testStore(t)
	mustAccount(t, s, ctx, "wolfgang")

	got, err := s.AccountByName(ctx, "wolfgang")
	if err != nil {
		t.Fatalf("an account without a cms profile must still load: %v", err)
	}
	if len(got.CMS.Groups) != 0 || len(got.CMS.Permissions) != 0 {
		t.Errorf("expected no CMS rights, got %+v", got.CMS)
	}
}

func TestSetCMSProfile(t *testing.T) {
	s, ctx := testStore(t)
	a := mustAccount(t, s, ctx, "matt")

	want := CMSProfile{Groups: []string{"familie"}, Permissions: []string{"entry.create"}}
	if err := s.SetCMSProfile(ctx, a.ID, want); err != nil {
		t.Fatalf("SetCMSProfile: %v", err)
	}
	// Writing twice must replace rather than conflict.
	want.Groups = append(want.Groups, "redaktion")
	if err := s.SetCMSProfile(ctx, a.ID, want); err != nil {
		t.Fatalf("SetCMSProfile again: %v", err)
	}

	got, err := s.AccountByName(ctx, "matt")
	if err != nil {
		t.Fatalf("AccountByName: %v", err)
	}
	if len(got.CMS.Groups) != 2 || got.CMS.Groups[1] != "redaktion" {
		t.Errorf("groups = %v, want [familie redaktion]", got.CMS.Groups)
	}
}

func TestPasskeyRoundTrip(t *testing.T) {
	s, ctx := testStore(t)
	a := mustAccount(t, s, ctx, "matt")

	cred := &webauthn.Credential{
		ID:        []byte("credential-id"),
		PublicKey: []byte("public-key"),
		Flags:     webauthn.CredentialFlags{UserPresent: true, UserVerified: true},
	}
	if err := s.AddPasskey(ctx, a.ID, cred, "YubiKey"); err != nil {
		t.Fatalf("AddPasskey: %v", err)
	}
	// Re-registering the same authenticator is success, not a conflict.
	if err := s.AddPasskey(ctx, a.ID, cred, "YubiKey"); err != nil {
		t.Fatalf("AddPasskey twice: %v", err)
	}

	keys, err := s.ListPasskeys(ctx, a.ID)
	if err != nil || len(keys) != 1 {
		t.Fatalf("ListPasskeys: %v, %d keys", err, len(keys))
	}
	if keys[0].Name != "YubiKey" || keys[0].BackupEligible {
		t.Errorf("unexpected stored passkey: %+v", keys[0])
	}
	if string(keys[0].Credential.PublicKey) != "public-key" {
		t.Error("the credential did not survive the round trip intact")
	}

	creds, err := s.Passkeys(ctx, a.ID)
	if err != nil || len(creds) != 1 {
		t.Fatalf("Passkeys: %v, %d", err, len(creds))
	}

	if err := s.DeletePasskey(ctx, a.ID, cred.ID); err != nil {
		t.Fatalf("DeletePasskey: %v", err)
	}
	if keys, _ := s.ListPasskeys(ctx, a.ID); len(keys) != 0 {
		t.Errorf("passkey survived deletion: %+v", keys)
	}
}

// backup_eligible is denormalised out of the credential so the device-bound
// policy is queryable; it has to reflect what was registered.
func TestPasskeyRecordsBackupEligibility(t *testing.T) {
	s, ctx := testStore(t)
	a := mustAccount(t, s, ctx, "matt")

	synced := &webauthn.Credential{
		ID:        []byte("synced"),
		PublicKey: []byte("k"),
		Flags:     webauthn.CredentialFlags{BackupEligible: true, BackupState: true},
	}
	if err := s.AddPasskey(ctx, a.ID, synced, "iCloud"); err != nil {
		t.Fatalf("AddPasskey: %v", err)
	}
	keys, err := s.ListPasskeys(ctx, a.ID)
	if err != nil || len(keys) != 1 {
		t.Fatalf("ListPasskeys: %v", err)
	}
	if !keys[0].BackupEligible {
		t.Error("backup_eligible was not recorded from the credential flags")
	}
}

// The update is the check, so two concurrent redemptions cannot both succeed.
func TestEnrollTokenIsSingleUseUnderConcurrency(t *testing.T) {
	s, ctx := testStore(t)
	a := mustAccount(t, s, ctx, "matt")

	token, err := s.CreateEnrollToken(ctx, a.ID, time.Hour)
	if err != nil {
		t.Fatalf("CreateEnrollToken: %v", err)
	}

	// Validating must not spend it: a mail client prefetching the link must not
	// burn someone's one-time enrollment.
	if _, err := s.EnrollTokenAccount(ctx, token); err != nil {
		t.Fatalf("EnrollTokenAccount: %v", err)
	}

	const racers = 8
	var wg sync.WaitGroup
	results := make([]error, racers)
	wg.Add(racers)
	for i := range racers {
		go func() {
			defer wg.Done()
			_, results[i] = s.ConsumeEnrollToken(ctx, token)
		}()
	}
	wg.Wait()

	won := 0
	for _, err := range results {
		if err == nil {
			won++
		} else if !errors.Is(err, ErrBadToken) {
			t.Errorf("unexpected error from a losing racer: %v", err)
		}
	}
	if won != 1 {
		t.Errorf("%d of %d redemptions succeeded, want exactly 1", won, racers)
	}
}

func TestEnrollTokenExpiryAndUnknown(t *testing.T) {
	s, ctx := testStore(t)
	a := mustAccount(t, s, ctx, "matt")

	expired, err := s.CreateEnrollToken(ctx, a.ID, -time.Second)
	if err != nil {
		t.Fatalf("CreateEnrollToken: %v", err)
	}
	if _, err := s.ConsumeEnrollToken(ctx, expired); !errors.Is(err, ErrBadToken) {
		t.Errorf("expired token: got %v, want ErrBadToken", err)
	}
	if _, err := s.ConsumeEnrollToken(ctx, "not-a-token"); !errors.Is(err, ErrBadToken) {
		t.Errorf("unknown token: got %v, want ErrBadToken", err)
	}
}

func TestSessionLifecycle(t *testing.T) {
	s, ctx := testStore(t)
	a := mustAccount(t, s, ctx, "matt")

	token, err := s.CreateSession(ctx, a.ID, time.Hour)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	got, _, err := s.SessionAccount(ctx, token, time.Hour, time.Minute)
	if err != nil || got.ID != a.ID {
		t.Fatalf("SessionAccount: %v, %+v", err, got)
	}

	if err := s.DeleteSession(ctx, token); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
	if _, _, err := s.SessionAccount(ctx, token, time.Hour, time.Minute); !errors.Is(err, ErrNoSession) {
		t.Errorf("deleted session: got %v, want ErrNoSession", err)
	}
}

func TestExpiredSessionIsRefused(t *testing.T) {
	s, ctx := testStore(t)
	a := mustAccount(t, s, ctx, "matt")

	token, err := s.CreateSession(ctx, a.ID, -time.Second)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if _, _, err := s.SessionAccount(ctx, token, time.Hour, time.Minute); !errors.Is(err, ErrNoSession) {
		t.Errorf("expired session: got %v, want ErrNoSession", err)
	}
}

// Disabling an account must kill its live sessions, not just its next sign-in.
func TestSessionOfDisabledAccountIsRefused(t *testing.T) {
	s, ctx := testStore(t)
	a := mustAccount(t, s, ctx, "matt")
	token, _ := s.CreateSession(ctx, a.ID, time.Hour)

	if _, err := s.SetDisabled(ctx, "matt", true); err != nil {
		t.Fatalf("SetDisabled: %v", err)
	}
	if _, _, err := s.SessionAccount(ctx, token, time.Hour, time.Minute); !errors.Is(err, ErrNoAccount) {
		t.Errorf("session of a disabled account: got %v, want ErrNoAccount", err)
	}
}

// The expiry slides so an active viewer is not signed out mid-film, but only
// once last_seen is stale — otherwise every range request of a video is a write.
func TestSessionSlidesOnlyWhenStale(t *testing.T) {
	s, ctx := testStore(t)
	a := mustAccount(t, s, ctx, "matt")
	token, _ := s.CreateSession(ctx, a.ID, time.Hour)

	var before time.Time
	if err := s.pool.QueryRow(ctx, `select last_seen from session limit 1`).Scan(&before); err != nil {
		t.Fatalf("reading last_seen: %v", err)
	}

	// Fresh session, wide window: nothing should move.
	if _, _, err := s.SessionAccount(ctx, token, time.Hour, time.Hour); err != nil {
		t.Fatalf("SessionAccount: %v", err)
	}
	var after time.Time
	if err := s.pool.QueryRow(ctx, `select last_seen from session limit 1`).Scan(&after); err != nil {
		t.Fatalf("reading last_seen: %v", err)
	}
	if !after.Equal(before) {
		t.Error("last_seen moved on a fresh session; every request would be a write")
	}

	// Now age the row past the window.
	if _, err := s.pool.Exec(ctx, `update session set last_seen = now() - interval '10 minutes'`); err != nil {
		t.Fatalf("ageing the session: %v", err)
	}
	if _, _, err := s.SessionAccount(ctx, token, time.Hour, time.Minute); err != nil {
		t.Fatalf("SessionAccount: %v", err)
	}
	if err := s.pool.QueryRow(ctx, `select last_seen from session limit 1`).Scan(&after); err != nil {
		t.Fatalf("reading last_seen: %v", err)
	}
	if time.Since(after) > time.Minute {
		t.Error("last_seen did not slide once it was stale")
	}
}

func TestChallengeIsSingleUse(t *testing.T) {
	s, ctx := testStore(t)
	a := mustAccount(t, s, ctx, "matt")

	id, err := s.SaveChallenge(ctx, Challenge{
		Purpose:   PurposeRegister,
		AccountID: &a.ID,
		Data:      &webauthn.SessionData{Challenge: "abc"},
	}, time.Minute)
	if err != nil {
		t.Fatalf("SaveChallenge: %v", err)
	}

	got, err := s.TakeChallenge(ctx, PurposeRegister, id)
	if err != nil {
		t.Fatalf("TakeChallenge: %v", err)
	}
	if got.Data.Challenge != "abc" || got.AccountID == nil || *got.AccountID != a.ID {
		t.Errorf("challenge did not round trip: %+v", got)
	}
	if got.Reauthed {
		t.Error("a challenge saved without re-authentication reported Reauthed")
	}

	if _, err := s.TakeChallenge(ctx, PurposeRegister, id); !errors.Is(err, ErrBadToken) {
		t.Errorf("replayed challenge: got %v, want ErrBadToken", err)
	}
}

// A login challenge must not be usable to complete a registration.
func TestChallengePurposeIsEnforced(t *testing.T) {
	s, ctx := testStore(t)

	id, err := s.SaveChallenge(ctx, Challenge{
		Purpose: PurposeLogin,
		Data:    &webauthn.SessionData{Challenge: "abc"},
	}, time.Minute)
	if err != nil {
		t.Fatalf("SaveChallenge: %v", err)
	}
	if _, err := s.TakeChallenge(ctx, PurposeRegister, id); !errors.Is(err, ErrBadToken) {
		t.Errorf("crossed purpose: got %v, want ErrBadToken", err)
	}
}

// The password step-up is recorded on the challenge, so the finish endpoint can
// refuse a registration that never re-authenticated.
func TestChallengeCarriesReauth(t *testing.T) {
	s, ctx := testStore(t)
	a := mustAccount(t, s, ctx, "matt")

	id, err := s.SaveChallenge(ctx, Challenge{
		Purpose:   PurposeRegister,
		AccountID: &a.ID,
		Data:      &webauthn.SessionData{Challenge: "abc"},
		Reauthed:  true,
	}, time.Minute)
	if err != nil {
		t.Fatalf("SaveChallenge: %v", err)
	}
	got, err := s.TakeChallenge(ctx, PurposeRegister, id)
	if err != nil {
		t.Fatalf("TakeChallenge: %v", err)
	}
	if !got.Reauthed {
		t.Error("the re-authentication stamp was lost")
	}
}

func TestCleanupRemovesExpiredRows(t *testing.T) {
	s, ctx := testStore(t)
	a := mustAccount(t, s, ctx, "matt")

	if _, err := s.CreateSession(ctx, a.ID, -time.Second); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if _, err := s.CreateEnrollToken(ctx, a.ID, -time.Second); err != nil {
		t.Fatalf("CreateEnrollToken: %v", err)
	}
	if _, err := s.SaveChallenge(ctx, Challenge{
		Purpose: PurposeLogin, Data: &webauthn.SessionData{},
	}, -time.Second); err != nil {
		t.Fatalf("SaveChallenge: %v", err)
	}
	live, err := s.CreateSession(ctx, a.ID, time.Hour)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	if err := s.Cleanup(ctx); err != nil {
		t.Fatalf("Cleanup: %v", err)
	}

	for _, table := range []string{"session", "enroll_token", "webauthn_challenge"} {
		var n int
		if err := s.pool.QueryRow(ctx, `select count(*) from `+table).Scan(&n); err != nil {
			t.Fatalf("counting %s: %v", table, err)
		}
		want := 0
		if table == "session" {
			want = 1 // the live one survives
		}
		if n != want {
			t.Errorf("%s has %d rows after cleanup, want %d", table, n, want)
		}
	}
	if _, _, err := s.SessionAccount(ctx, live, time.Hour, time.Minute); err != nil {
		t.Errorf("cleanup removed a live session: %v", err)
	}
}

// Everything hanging off an account goes with it.
func TestDeletingAccountCascades(t *testing.T) {
	s, ctx := testStore(t)
	a := mustAccount(t, s, ctx, "matt")

	if err := s.SetCMSProfile(ctx, a.ID, CMSProfile{Groups: []string{"familie"}}); err != nil {
		t.Fatalf("SetCMSProfile: %v", err)
	}
	if _, err := s.CreateSession(ctx, a.ID, time.Hour); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if _, err := s.CreateEnrollToken(ctx, a.ID, time.Hour); err != nil {
		t.Fatalf("CreateEnrollToken: %v", err)
	}
	if err := s.AddPasskey(ctx, a.ID, &webauthn.Credential{ID: []byte("c"), PublicKey: []byte("k")}, ""); err != nil {
		t.Fatalf("AddPasskey: %v", err)
	}

	if _, err := s.pool.Exec(ctx, `delete from account where id = $1`, a.ID); err != nil {
		t.Fatalf("deleting the account: %v", err)
	}

	for _, table := range []string{"cmsauth", "session", "enroll_token", "passkey"} {
		var n int
		if err := s.pool.QueryRow(ctx, `select count(*) from `+table).Scan(&n); err != nil {
			t.Fatalf("counting %s: %v", table, err)
		}
		if n != 0 {
			t.Errorf("%s still has %d rows after the account was deleted", table, n)
		}
	}
}
