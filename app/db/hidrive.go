package db

import (
	"context"
	"fmt"
	"time"

	"github.com/ihleven/ihlvn/pkg/hiauth"
)

type HiToken struct {
	Alias        string    `db:"alias"`
	Scope        string    `db:"scope"`
	RefreshToken string    `db:"refresh_token"`
	ExpiresAt    time.Time `db:"expires_at"`
	// AccessToken  string `json:"access_token"`
}

func (db *DB) GetToken(alias string) (*HiToken, error) {

	sql := `
        SELECT alias,scope,refresh_token,expires_at
          FROM hitoken
         WHERE alias = $1
    `
	var token HiToken
	err := db.pool.QueryRow(db.ctx, sql, alias).Scan(&token.Alias, &token.Scope, &token.RefreshToken, &token.ExpiresAt)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (db *DB) LoadTokens() ([]map[string]string, error) {

	rows, err := db.pool.Query(context.Background(), "SELECT alias,scope,refresh_token,expires_at FROM hitoken")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]string

	for rows.Next() {

		token := HiToken{}

		err := rows.Scan(&token.Alias, &token.Scope, &token.RefreshToken, &token.ExpiresAt)
		if err != nil {
			return nil, err
		}

		result = append(result, map[string]string{"alias": token.Alias, "scope": token.Scope, "refresh_token": token.RefreshToken, "expires_at": token.ExpiresAt.String()})
	}

	return result, nil
}

func (db *DB) StoreToken(token *hiauth.Token) error {
	sql := `
		UPDATE hitoken
           SET scope=$2,refresh_token=$3,expires_at=$4
         WHERE alias=$1
    `

	expiresAt := time.Now().Add(24 * 30 * 3 * time.Hour) // 3 months

	commandTag, err := db.pool.Exec(db.ctx, sql, token.Alias, token.Scope, token.RefreshToken, expiresAt)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("now row found with alias %s", token.Alias)
	}

	return nil
}
