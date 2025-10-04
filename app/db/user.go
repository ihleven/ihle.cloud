package db

import (
	"context"

	"github.com/ihleven/ihle.cloud/backend/auth"
)

func (db *DB) GetUser(id string) (*auth.Account, error) {

	row := db.pool.QueryRow(context.Background(), "SELECT email,name,credentials,hidrive FROM account WHERE name = $1", id)
	// if err != nil {
	// 	return nil, err
	// }
	// defer rows.Close()

	var email string
	// var credentials auth.Credentials
	var account auth.Account

	// for rows.Next() {

	err := row.Scan(&email, &account.ID, &account.Credentials, &account.Settings.Hidrive)
	if err != nil {
		return nil, err
	}

	// }

	return &account, nil
}
func (db *DB) LoadAccounts() ([]auth.Account, error) {

	rows, err := db.pool.Query(context.Background(), "SELECT email,name,credentials,hidrive FROM account")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []auth.Account

	for rows.Next() {
		var email string
		// var credentials auth.Credentials
		var account auth.Account
		err := rows.Scan(&email, &account.ID, &account.Credentials, &account.Settings.Hidrive)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}
