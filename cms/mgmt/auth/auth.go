package auth

import (
	"bitbucket.org/hotelplan/webcc-content/cms/content"
)

func New() Authenticator {
	return nil
}

type Authenticator interface {
	AuthAccount(string) (*Account, bool)
	GetAccount(string, content.User) (*Account, error)
	GetGroup(grp string) (*Group, error)
}

type ReadAuthenticator interface {
	Authenticator
	ListAccounts(*content.User) ([]*Account, error)
	ListGroups() []*Group
}

type WriteAuthenticator interface {
	Authenticator
	CreateAccount(Account, *Account) (*Account, error)
	UpdateAccount(Account, *Account) (*content.Entry, error)
	DeleteAccount(string, *Account) error
	CreateGroup(data Group, a *Account) (*Group, error)
	UpdateGroup(data Group, a *Account) (*content.Entry, error)
	DeleteGroup(id string, auth *Account) error
}
