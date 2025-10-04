package auth

import (
	"encoding/json"
	"slices"

	"bitbucket.org/hotelplan/webcc-content/cms/content"
	"bitbucket.org/hotelplan/webcc-content/cms/permission"
)

func init() {
	content.Register(Config{})
	content.Register(Account{})
	content.Register(Group{})
}

type Account struct {
	content.ContentType `type:"Account" folder:"pages" mimetype:"application/json" json:"-"`

	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Groups      []string `json:"groups"`
	Active      bool     `json:"active"`
	Paths       []string `json:"paths"`
	Locales     []string `json:"locales"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`

	content.User `json:"-"`
}

func (a *Account) MarshalJSON() ([]byte, error) {
	data := struct {
		ID          string   `json:"id"`
		Name        string   `json:"name"`
		Email       string   `json:"email"`
		Groups      []string `json:"groups"`
		Active      bool     `json:"active"`
		Paths       []string `json:"paths"`
		Locales     []string `json:"locales"`
		Roles       []string `json:"roles"`
		Permissions []string `json:"permissions"`
		Scope       []string `json:"scope"`
	}{ID: a.ID, Name: a.Name, Email: a.Email, Groups: a.Groups, Active: a.Active, Paths: a.Paths, Locales: a.Locales, Roles: a.Roles, Permissions: a.Permissions}
	//, Scope: a.Scope.String()}

	for _, perm := range a.Scope {
		data.Scope = append(data.Scope, perm.String())
	}
	slices.Sort(data.Scope)

	return json.Marshal(data)
}

func (a *Account) Prepare(rolemap map[string]*Role) {
	if a.Groups == nil {
		a.Groups = []string{}
	}
	if a.Roles == nil {
		a.Roles = []string{}
	}
	if a.Permissions == nil {
		a.Permissions = []string{}
	}
	if a.Locales == nil {
		a.Locales = []string{}
	}
	if a.Paths == nil {
		a.Paths = []string{}
	}
	a.User.ID = a.ID
	a.User.Groups = a.Groups
	a.User.Scope = map[permission.Type]permission.Permission{}

	for _, rolename := range a.Roles {
		if role, ok := rolemap[rolename]; ok {
			a.User.Scope.Add(role.Permissions, a.Locales, a.Paths)
		}
	}

	a.User.Scope.Add(a.Permissions, a.Locales, a.Paths)
}

func (a *Account) Clone() interface{} {
	clone := Account{
		ID:          a.ID,
		Name:        a.Name,
		Email:       a.Email,
		Groups:      make([]string, len(a.Groups)),
		Active:      a.Active,
		Paths:       make([]string, len(a.Paths)),
		Locales:     make([]string, len(a.Locales)),
		Roles:       make([]string, len(a.Roles)),
		Permissions: make([]string, len(a.Permissions)),
		User:        *a.User.Clone(),
	}
	copy(clone.Groups, a.Groups)
	copy(clone.Roles, a.Roles)
	copy(clone.Permissions, a.Permissions)
	copy(clone.Locales, a.Locales)
	copy(clone.Paths, a.Paths)

	return &clone
}
func (a *Account) Patch(data *Account) []string {
	var changedfields []string
	if different(a.Groups, data.Groups) {
		a.Groups = data.Groups
		changedfields = append(changedfields, "groups")
	}
	if different(a.Roles, data.Roles) {
		a.Roles = data.Roles
		changedfields = append(changedfields, "roles")
	}
	if different(a.Permissions, data.Permissions) {
		a.Permissions = data.Permissions
		changedfields = append(changedfields, "permissions")
	}
	if different(a.Paths, data.Paths) {
		a.Paths = data.Paths
		changedfields = append(changedfields, "paths")
	}
	if different(a.Locales, data.Locales) {
		a.Locales = data.Locales
		changedfields = append(changedfields, "locales")
	}
	if a.Active != data.Active {
		a.Active = data.Active
		changedfields = append(changedfields, "active")
	}
	return changedfields
}

func different(target []string, source []string) (different bool) {
	if len(target) != len(source) {
		return true
	}
	slices.Sort(target)
	slices.Sort(source)
	for i := range target {
		if target[i] != source[i] {
			return true
		}
	}
	return false
}

type Group struct {
	content.ContentType `type:"Group" repo:"internal" folder:"groups" mimetype:"application/json" json:"-"`
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	Accounts            []string `json:"accounts"`
}

func (g *Group) Clone() interface{} {
	return &Group{Name: g.Name, Description: g.Description, Accounts: append(g.Accounts[:0:0], g.Accounts...)}
}
