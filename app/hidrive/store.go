// Package hidrive is the storage provider's half of the app: the OAuth tokens
// that let it read a HiDrive account.
//
// These lived with the account tables, which put a third party's credentials in
// the authentication schema and made the auth package responsible for something that has
// nothing to do with who anyone is. The rows stay where they are — an applied
// migration is pinned by a checksum of its text and cannot be rewritten — but
// the code that owns them is here.
package hidrive

import (
	"context"
	"fmt"
	"time"

	"github.com/ihleven/ihlvn/pkg/hiauth"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Token struct {
	Alias        string    `db:"alias"`
	Scope        string    `db:"scope"`
	RefreshToken string    `db:"refresh_token"`
	ExpiresAt    time.Time `db:"expires_at"`
	// AccessToken  string `json:"access_token"`
}

func (s *Store) LoadTokens(ctx context.Context) ([]map[string]string, error) {

	rows, err := s.pool.Query(context.Background(), "SELECT alias,scope,refresh_token,expires_at FROM hitoken")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]string

	for rows.Next() {

		token := Token{}

		err := rows.Scan(&token.Alias, &token.Scope, &token.RefreshToken, &token.ExpiresAt)
		if err != nil {
			return nil, err
		}

		result = append(result, map[string]string{"alias": token.Alias, "scope": token.Scope, "refresh_token": token.RefreshToken, "expires_at": token.ExpiresAt.String()})
	}

	return result, nil
}

func (s *Store) StoreToken(ctx context.Context, token *hiauth.Token) error {
	sql := `
		UPDATE hitoken
           SET scope=$2,refresh_token=$3,expires_at=$4
         WHERE alias=$1
    `

	expiresAt := time.Now().Add(24 * 30 * 3 * time.Hour) // 3 months

	commandTag, err := s.pool.Exec(ctx, sql, token.Alias, token.Scope, token.RefreshToken, expiresAt)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("now row found with alias %s", token.Alias)
	}

	return nil
}

// Store reads and writes the tokens. A thin type over the pool, like the account
// store: the queries are few and the rows are simple.
type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}
