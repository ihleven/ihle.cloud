package db

func NewAuthDB(db *DB) *authDB {

	return &authDB{DB: db}
}

type authDB struct {
	*DB
}

// type AuthDB interface {
// 	GetUser(name string) (*webauthn.User, error)
// }

type User struct {
	ID          uint64
	Name        string
	DisplayName string
	// Credentials []webauthn.Credential
}
