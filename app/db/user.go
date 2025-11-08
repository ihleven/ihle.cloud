package db

import (
	"context"

	"github.com/ihleven/ihlvn/pkg/auth"
)

func (db *DB) GetUser(id string) (*auth.Account, error) {

	row := db.pool.QueryRow(context.Background(), "SELECT email,name,display_name,credentials,hidrive FROM account WHERE name = $1", id)
	// if err != nil {
	// 	return nil, err
	// }
	// defer rows.Close()

	// var email string
	// var credentials auth.Credentials
	var account auth.Account

	// for rows.Next() {

	err := row.Scan(&account.Email, &account.ID, &account.Name, &account.Credentials, &account.Settings.Hidrive)
	if err != nil {
		return nil, err
	}

	// }

	return &account, nil
}
func (db *DB) LoadAccounts() ([]auth.Account, error) {

	sql := `
        SELECT a.email,a.name,a.display_name,a.credentials,a.hidrive, cms.groups,cms.permissions
          FROM account a, cmsauth cms
		 WHERE a.name=cms.id
    `

	rows, err := db.pool.Query(db.ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []auth.Account

	for rows.Next() {

		var a auth.Account
		err := rows.Scan(&a.Email, &a.ID, &a.Name, &a.Credentials, &a.Settings.Hidrive, &a.Settings.CMS.Groups, &a.Settings.CMS.Permissions)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}

	return accounts, nil
}
