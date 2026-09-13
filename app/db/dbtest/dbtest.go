// Package dbtest hands a test a database it is allowed to destroy.
//
// It exists because more than one package's tests need the same three things —
// a migrated pool, a guarantee that it is not somebody's real database, and
// exclusive use of it while they run — and because the guarantee is worth
// having in exactly one place.
package dbtest

import (
	"context"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/ihleven/ihlvn/app/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Env names the database these tests may destroy.
const Env = "IHLVN_TEST_DATABASE_URL"

// lockKey serialises the test binaries that truncate. Under `go test ./...` the
// package binaries run at the same time, and each of these tests begins by
// emptying tables. Without a lock one package wipes another's fixtures mid-test,
// which surfaces as an unrelated test failing to find a row it had just created.
// The lock lives in Postgres, so it coordinates across processes without the
// build having to know anything about it.
const lockKey = 1974

// Pool connects, migrates, and holds the database for the duration of the test.
//
// It does not empty anything: what to clear is the caller's business, and a
// package that only reads should not have to say so.
func Pool(t *testing.T) (*pgxpool.Pool, context.Context) {
	t.Helper()

	raw := getenv(t)
	requireDisposable(t, raw)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, raw)
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("pinging: %v", err)
	}
	if err := db.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrating: %v", err)
	}
	hold(t, pool)
	return pool, ctx
}

func getenv(t *testing.T) string {
	t.Helper()
	raw := os.Getenv(Env)
	if raw == "" {
		t.Skipf("%s not set", Env)
	}
	return raw
}

// requireDisposable refuses a database whose name does not mark it as one that
// may be destroyed.
//
// These tests truncate, cascading and irreversibly, and an environment variable
// carries no hint of what it is pointing at. Aiming this at a development
// database once cost four accounts, their passkeys and their permissions — and
// the emptied tables then read as "there was never an account here", which is a
// worse failure than an error, because it looks like a finding.
func requireDisposable(t *testing.T, raw string) {
	t.Helper()

	name := databaseName(raw)
	if !strings.Contains(strings.ToLower(name), "test") {
		t.Fatalf("%s points at %q, which is not named as a test database.\n"+
			"These tests empty tables. Use one that may be destroyed:\n"+
			"    createdb ihlvn_test\n"+
			"    %s=postgres:///ihlvn_test go test ./...", Env, name, Env)
	}
}

// databaseName is the database a connection string names, in either the URL or
// the keyword form. An unreadable string yields "", which the guard refuses.
func databaseName(raw string) string {
	if u, err := url.Parse(raw); err == nil && u.Scheme != "" {
		return strings.TrimPrefix(u.Path, "/")
	}
	for _, field := range strings.Fields(raw) {
		if rest, ok := strings.CutPrefix(field, "dbname="); ok {
			return rest
		}
	}
	return ""
}

// hold takes the lock for the duration of the test.
func hold(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()
	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquiring a connection for the lock: %v", err)
	}
	if _, err := conn.Exec(ctx, `select pg_advisory_lock($1)`, lockKey); err != nil {
		conn.Release()
		t.Fatalf("taking the lock: %v", err)
	}
	t.Cleanup(func() {
		if _, err := conn.Exec(context.Background(), `select pg_advisory_unlock($1)`, lockKey); err != nil {
			t.Errorf("releasing the lock: %v", err)
		}
		conn.Release()
	})
}
