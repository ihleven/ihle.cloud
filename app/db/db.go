package db

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/pkg/errors"
)

func New(conn string) (*DB, error) {
	if conn == "" {
		return nil, fmt.Errorf("no db connection string provided")
	}
	dbpool, err := pgxpool.New(context.Background(), conn)
	if err != nil {
		return nil, err
	}

	return &DB{ctx: context.Background(), pool: dbpool}, nil
}

type DB struct {
	ctx  context.Context
	pool *pgxpool.Pool
}

func (db *DB) Close() {
	db.pool.Close()
}

// Pool exposes the connection pool so other packages can run their own queries
// against the same database without opening a second pool.
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

// type KunstDB interface {
// 	// SaveTask(title, description string) error
// 	GetAusstellungen() ([]Ausstellung, error)
// 	LoadBilder(where map[string]interface{}, serienbilder bool, deleted bool, orderBy string) ([]Bild2, error)
// }

func (r *DB) Select(dst interface{}, query string, args ...interface{}) error {

	err := pgxscan.Select(r.ctx, r.pool, dst, query, args...)
	if err != nil {
		// if errors.As(err, &pgx.ErrNoRows) {
		// 	return errors.NewWithCode(errors.NotFound, "Not found: %s", query)
		// }
		return errors.Wrap(err, "Error")
	}
	return nil
}

// func QueryGeneric[K comparable](db *DB, stmt string, params ...interface{}) (K, error) {

// }
