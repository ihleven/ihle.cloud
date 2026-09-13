package db_test

import (
	"context"
	"testing"

	"github.com/ihleven/ihlvn/app/db"
	"github.com/ihleven/ihlvn/app/db/dbtest"
)

func TestMigrateIsIdempotent(t *testing.T) {
	pool, ctx := dbtest.Pool(t)
	if err := db.Migrate(ctx, pool); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
}

// An applied migration whose file later changes is drift between environments,
// so it is refused rather than ignored.
func TestMigrateRefusesAnEditedMigration(t *testing.T) {
	pool, ctx := dbtest.Pool(t)

	const version = "0001_baseline.sql"

	var original []byte
	if err := pool.QueryRow(ctx,
		`select checksum from schema_migrations where version = $1`, version).Scan(&original); err != nil {
		t.Fatalf("reading the recorded checksum: %v", err)
	}
	// The database outlives this test, so put the real checksum back or every
	// later test fails at Migrate.
	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(),
			`update schema_migrations set checksum = $1 where version = $2`, original, version); err != nil {
			t.Fatalf("restoring the checksum: %v", err)
		}
	})

	if _, err := pool.Exec(ctx,
		`update schema_migrations set checksum = $1 where version = $2`,
		[]byte("not the checksum this file has"), version); err != nil {
		t.Fatalf("tampering with the checksum: %v", err)
	}
	if err := db.Migrate(ctx, pool); err == nil {
		t.Fatal("a changed migration was accepted")
	}
}
