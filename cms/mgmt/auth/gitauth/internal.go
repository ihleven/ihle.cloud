package gitauth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"bitbucket.org/hotelplan/webcc-content/cms/content"
	"bitbucket.org/hotelplan/webcc-content/cms/mgmt/auth"
	"bitbucket.org/hotelplan/webcc-content/cms/permission"
	"bitbucket.org/hotelplan/webcc-content/cms/pkg/errors"
)

// interface, das das Repo erfüllen muss, damit CMS es benutzen kann
type ContentRepository interface {
	GetBytes(string, bool) ([]byte, error)
	// Store(string, []byte) error
	LoadMulti(string) (map[string][]byte, error)
	// StoreMulti(map[string][]byte, string, string) error
	// Delete([]string, string, string) error
	// History(string, string, time.Duration, int, int) ([]storage.Commit, error)
	// Name() string
}

type authProvider struct {
	repo     ContentRepository
	Accounts map[string]*auth.Account
	Groups   map[string]*auth.Group
	Roles    map[string]*auth.Role
	config   auth.Config
}

func NewAuthorizationProvider(repo ContentRepository) (*authProvider, error) {

	intrnl := authProvider{
		repo:     repo,
		Accounts: map[string]*auth.Account{},
		Groups:   map[string]*auth.Group{},
		Roles:    map[string]*auth.Role{},
	}

	// Settings
	bytemap, err := repo.LoadMulti("config.json")
	if err != nil {
		return nil, err
	}
	settingsentry, err := content.ParseEntry(nil, bytemap["config.json"], "application/json")
	if err != nil {
		return nil, err
	}
	if config, ok := settingsentry.Content.(*auth.Config); ok && config != nil {
		intrnl.config = *config
		for i := range intrnl.config.Roles {
			intrnl.Roles[intrnl.config.Roles[i].ID] = &intrnl.config.Roles[i]
		}
	} else {
		return nil, errors.New("invalid internal cms configuration")
	}

	// GROUPS
	bytemap, err = repo.LoadMulti("groups")
	if err != nil {
		return nil, err
	}
	var groups []string
	for _, bytes := range bytemap {

		entry, err := content.ParseEntry(nil, bytes, "application/json")
		if err != nil {
			return nil, err
		}
		grp := entry.Content.(*auth.Group)

		intrnl.Groups[grp.Name] = grp
		groups = append(groups, grp.Name)
	}
	fmt.Printf("+++ => found %d groups: %s\n", len(intrnl.Groups), strings.Join(groups, ","))

	////////// ACCOUNTS //////////
	bytemap, err = repo.LoadMulti("accounts")
	if err != nil {
		return nil, err
	}

	var accounts []string
	for _, bytes := range bytemap {

		entry, err := content.ParseEntry(nil, bytes, "application/json")
		if err != nil {
			return nil, err
		}
		account := entry.Content.(*auth.Account)

		account.Prepare(intrnl.Roles)

		intrnl.Accounts[account.ID] = account
		accounts = append(accounts, account.ID)
		// fmt.Printf("\033[2K\r * account %s: %s", account.ID, account.Name)
	}
	// fmt.Printf("\n")
	fmt.Printf("+++ => found %d accounts: %s\n", len(intrnl.Accounts), strings.Join(accounts, ","))

	return &intrnl, nil
}

var PERM_ACCOUNT_3RD = permission.Define("account", "3rdparty.read", false, false)
var PERM_ACCOUNT_GROUPS_WRITE = permission.Define("account", "groups.write", false, false)
var PERM_ACCOUNT_AUTH_WRITE = permission.Define("account", "auth.write", false, false)
var PERM_ACCOUNT_CREATE = permission.Define("account", "create", false, false)
var PERM_ACCOUNT_DELETE = permission.Define("account", "delete", false, false)

func (cms *authProvider) commit(a *auth.Account, msg string, reponame string, entrymap map[string]*content.Entry) error {
	return errors.NewWithCode(http.StatusNotImplemented, "")
}

func (c *authProvider) CreateAccount(data auth.Account, a *auth.Account) (*auth.Account, error) {

	if a.User.Scope.Denies(PERM_ACCOUNT_CREATE) {
		return nil, errors.New("missing permission %s", PERM_ACCOUNT_CREATE)
	}

	f, err := c.repo.LoadMulti("accounts/" + data.ID + ".json")
	if len(f) > 0 {
		// account exists -> 409 Conflict
		return nil, errors.WrapWithCode(err, http.StatusConflict, "account %s already exists repo", data.ID)
	}
	_, ok := c.Accounts[data.ID]
	if ok {
		// account exists -> 409 Conflict
		return nil, errors.NewWithCode(http.StatusConflict, "account %s already exists cache", data.ID)
	}

	data.Prepare(c.Roles)

	now := time.Now()
	entry := content.Entry{
		Meta: content.Meta{
			ID:    "accounts/" + data.ID,
			Space: "internal",
			Repo:  "internal",
			Path:  "accounts/" + data.ID + ".json",
			Type:  "Account",
			MIME:  "application/json",
			// Suffix:     ".json",
			Version:    "pub",
			Status:     "active",
			Collection: "accounts",
			Access: content.Access{
				Owner:     a.ID,
				Created:   now,
				Modified:  now,
				Published: now,
			},
		},
		Content: &data,
	}
	err = c.commit(a, "create account "+data.ID, "internal", map[string]*content.Entry{entry.Path: &entry})
	if err != nil {
		return nil, err
	}

	c.Accounts[data.ID] = &data

	return entry.Content.(*auth.Account), nil
}

func (c *authProvider) ListAccounts(user *content.User) ([]*auth.Account, error) {

	if user.Scope.Denies(PERM_ACCOUNT_3RD) {
		return nil, errors.New("missing permission %s", PERM_ACCOUNT_3RD)
	}

	var accounts []*auth.Account
	for _, account := range c.Accounts {
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (c *authProvider) GetAccount(id string, user content.User) (*auth.Account, error) {
	if user.ID != id && user.Scope.Denies(PERM_ACCOUNT_3RD) {
		return nil, errors.New("missing permission %s", PERM_ACCOUNT_3RD)
	}
	account, ok := c.Accounts[id]
	if !ok {
		return nil, errors.NewWithCode(404, "account not found: %s", id)
	}
	return account, nil
}

func (c *authProvider) AuthAccount(id string) (*auth.Account, bool) {
	a, ok := c.Accounts[id]
	return a, ok
}

func (c *authProvider) UpdateAccount(data auth.Account, a *auth.Account) (*content.Entry, error) {

	bytes, err := c.repo.GetBytes("accounts/"+data.ID+".json", false)
	if err != nil {
		return nil, errors.NewWithCode(404, "account %s not found in internal repo", data.ID)
	}
	entry, err := content.ParseEntry(nil, bytes, "application/json")
	if err != nil {
		return nil, err
	}
	storedAccount := entry.Content.(*auth.Account)
	cachedAccount, ok := c.Accounts[data.ID]
	if !ok {
		return nil, errors.NewWithCode(404, "account %s not found in cms cache", data.ID)
	}
	// ein frische Kopie der aktuellen Accountdaten, die gefahrlos geändert werden können
	account := cachedAccount.Clone().(*auth.Account)

	// wenn sich beim Patch ein Unterschied zu den repo-Daten ergibt, stimmt etwas nicht
	fields := account.Patch(storedAccount)
	if len(fields) > 0 {
		return nil, errors.NewWithCode(404, "cloned cached account data changed after patch with repo data (%s) => %v %v %v", data.ID, fields, cachedAccount, storedAccount)
	}

	msg := "update account " + data.ID + " on fields:"

	// eigentliches Update durchführen und Permission für geänderte Fields evaluieren
	fields = account.Patch(&data)
	for _, field := range fields {
		if field == "groups" && a.Scope.Denies(PERM_ACCOUNT_GROUPS_WRITE) {
			return nil, errors.NewWithCode(http.StatusForbidden, "missing permission %s", PERM_ACCOUNT_GROUPS_WRITE)
		} else if a.Scope.Denies(PERM_ACCOUNT_AUTH_WRITE) {
			return nil, errors.NewWithCode(http.StatusForbidden, "missing permission %s", PERM_ACCOUNT_AUTH_WRITE)
		}
		msg += " " + field
	}

	// we do not want to write clear names and e-mail-addresses into repository
	// only applies when logged in user changes his own account
	account.Name = account.ID
	account.Email = ""

	entry.Content = account

	err = c.commit(a, msg, "internal", map[string]*content.Entry{entry.Path: entry})
	if err != nil {
		return nil, err
	}
	account.Prepare(c.Roles)
	c.Accounts[data.ID] = account

	return entry, nil
}

func (c *authProvider) DeleteAccount(id string, a *auth.Account) error {

	if a.User.Scope.Denies(PERM_ACCOUNT_DELETE) {
		return errors.New("missing permission %s", PERM_ACCOUNT_DELETE)
	}

	path := fmt.Sprintf("accounts/%s.json", id)
	err := c.commit(a, "delete account "+id, "internal", map[string]*content.Entry{path: nil})
	if err != nil {
		return err
	}

	delete(c.Accounts, id)
	if _, ok := c.Accounts[id]; ok {
		return errors.New("Couldn't delete account %s in cache", id)
	}

	return nil
}
