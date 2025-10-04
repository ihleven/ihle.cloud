package auth

import "golang.org/x/crypto/bcrypt"

type Account struct {
	ID string `json:"id"`
	Credentials
	Settings `json:"settings"`
}

// Create a struct that models the structure of a user, both in the request body, and in the DB
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Settings struct {
	Hidrive struct {
		Alias string `json:"alias"`
		Home  string `json:"home"`
		Root  string `json:"root"`
	} `json:"hidrive"`
}

// HashPassword takes a password and returns the bcrypt hash in a string format.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash compares a password to a hash and returns if it is valid or not.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
