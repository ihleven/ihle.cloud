package authn

import (
	"bytes"
	"context"
	"crypto/sha256"
	"embed"
	"fmt"
	"io/fs"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// migrateLockKey identifies the advisory lock guarding the migration run. Any
// constant will do as long as nothing else in the database picks the same one.
const migrateLockKey int64 = 0x6968_6C76_6E01 // "ihlvn" in ASCII, plus a byte

// Migrate applies any migrations the database has not seen yet.
//
// A hand-rolled runner rather than a migration tool: there are a handful of
// files, they only ever go forward, and embedding them means the binary is the
// whole deployment. Each file runs inside a transaction together with its
// bookkeeping row, so a failure leaves no half-applied version behind.
//
// Two properties the runner depends on the database for. A session-level
// advisory lock serialises concurrent runs, so two processes starting against
// the same database cannot both apply the same file. And every applied file's
// SHA-256 is recorded and re-checked, so editing a migration that has already
// run is an error rather than silent drift between environments.
//
// Because a file is executed as a single batch inside a transaction, statements
// that cannot run in one — CREATE INDEX CONCURRENTLY, most notably — are not
// available. There are no down migrations: a dump taken before the run is the
// way back.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	// The lock is held by a session, so everything below runs on one connection.
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("authn: acquiring a connection to migrate: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, `select pg_advisory_lock($1)`, migrateLockKey); err != nil {
		return fmt.Errorf("authn: taking the migration lock: %w", err)
	}
	// Releasing a pooled connection does not reset its session, so the lock has
	// to go back explicitly. Detached from ctx so a cancelled run still unlocks.
	defer conn.Exec(context.WithoutCancel(ctx), `select pg_advisory_unlock($1)`, migrateLockKey)

	if _, err := conn.Exec(ctx, `
		create table if not exists schema_migrations (
			version    text        primary key,
			checksum   bytea       not null,
			applied_at timestamptz not null default now()
		)`); err != nil {
		return fmt.Errorf("authn: creating schema_migrations: %w", err)
	}

	names, err := migrationNames()
	if err != nil {
		return err
	}

	applied, err := appliedMigrations(ctx, conn)
	if err != nil {
		return err
	}

	for _, name := range names {
		body, err := migrationFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("authn: reading migration %s: %w", name, err)
		}
		sum := sha256.Sum256(body)

		if was, ok := applied[name]; ok {
			if !bytes.Equal(was, sum[:]) {
				return fmt.Errorf(
					"authn: migration %s changed after it was applied; add a new migration instead of editing an applied one",
					name)
			}
			continue
		}
		if err := applyMigration(ctx, conn, name, string(body), sum[:]); err != nil {
			return err
		}
	}
	return nil
}

// appliedMigrations reads the versions already recorded, with the checksum each
// had when it ran.
func appliedMigrations(ctx context.Context, conn *pgxpool.Conn) (map[string][]byte, error) {
	rows, err := conn.Query(ctx, `select version, checksum from schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("authn: reading applied migrations: %w", err)
	}
	defer rows.Close()

	applied := map[string][]byte{}
	for rows.Next() {
		var version string
		var checksum []byte
		if err := rows.Scan(&version, &checksum); err != nil {
			return nil, fmt.Errorf("authn: reading applied migrations: %w", err)
		}
		applied[version] = checksum
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("authn: reading applied migrations: %w", err)
	}
	return applied, nil
}

func applyMigration(ctx context.Context, conn *pgxpool.Conn, name, body string, checksum []byte) error {
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("authn: starting migration %s: %w", name, err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, body); err != nil {
		return fmt.Errorf("authn: applying migration %s: %w", name, err)
	}
	if _, err := tx.Exec(ctx,
		`insert into schema_migrations (version, checksum) values ($1, $2)`, name, checksum); err != nil {
		return fmt.Errorf("authn: recording migration %s: %w", name, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("authn: committing migration %s: %w", name, err)
	}
	return nil
}

// migrationNames returns the migration filenames in lexical order, which is why
// they are numbered.
func migrationNames() ([]string, error) {
	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("authn: listing migrations: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}
