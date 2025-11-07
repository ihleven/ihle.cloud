package db

func (db *DB) GetCMSAccount(id string) (*HiToken, error) {

	sql := `
        SELECT id,groups,permissions
          FROM cmsauth
         WHERE id = $1
    `
	var token HiToken
	err := db.pool.QueryRow(db.ctx, sql, id).Scan(&token.Alias, &token.Scope, &token.RefreshToken, &token.ExpiresAt)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

type Uuser struct {
	ID                  string
	Groups, Permissions []string
}

func (db *DB) LoadCMSUsers() ([]Uuser, error) {

	sql := `
        SELECT id,groups,permissions
          FROM cmsauth
    `
	rows, err := db.pool.Query(db.ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []Uuser

	for rows.Next() {

		var user Uuser
		err := rows.Scan(&user.ID, &user.Groups, &user.Permissions)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
