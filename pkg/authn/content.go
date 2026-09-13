package authn

import (
	"context"

	"github.com/interhome-group/cms/content"
	"github.com/interhome-group/cms/modules"
)

// ContentUser maps an account onto the identity the CMS evaluates.
//
// Name is used as the id because that is what content in this repository is
// owned by — an entry's Access.Owner holds a login name, not an email. The CMS's
// own deployment uses the email there instead, which is why this mapping is the
// application's to make rather than something the CMS can do for it.
//
// The signature is what a commit is authored with, so an edit is attributable to
// a person rather than to the service.
func (a *Account) ContentUser() content.User {
	return content.NewUser(a.Name,
		content.Signature{Name: a.DisplayName, Email: a.Email},
		a.CMS.Groups, a.CMS.Permissions)
}

// EntitledModules returns the feature areas this account may see, derived from
// its scope rather than stored on it. An account granting nothing is entitled to
// nothing, which is what keeps the anonymous account out of every area.
func (a *Account) EntitledModules() []string {
	return modules.Entitled(a.ContentUser().Scope)
}

// wildcardPermission is content.Scope's grant-all, matched there as the entire
// permission string.
const wildcardPermission = "*"

// UnregisteredPermissions returns the account's permission strings that no
// package in this binary has defined. The CMS drops them silently when building
// a scope, so an account can appear to hold rights that grant nothing; the admin
// commands report them rather than leaving that invisible.
func (a *Account) UnregisteredPermissions() []string {
	known := map[string]struct{}{}
	for _, p := range content.RegisteredPermissions() {
		known[p] = struct{}{}
	}

	var unknown []string
	for _, p := range a.CMS.Permissions {
		// "*" is not a permission name but a grant-all the scope expands to every
		// registered permission when it is built. It is the whole string or
		// nothing: a constrained form like "*:de" is not expanded, and so really
		// does grant nothing.
		if p == wildcardPermission {
			continue
		}

		// A permission may carry locale and path constraints after a colon;
		// only the name is registered.
		name := p
		for i := range len(p) {
			if p[i] == ':' {
				name = p[:i]
				break
			}
		}
		if _, ok := known[name]; !ok {
			unknown = append(unknown, p)
		}
	}
	return unknown
}

// ContextUser is the identity the CMS evaluates a request against.
//
// A request without an account is the anonymous user rather than a refusal:
// entry ACLs are applied to it too, and it is granted whatever an entry grants
// to others. That is what lets the public site read content without the write
// path being open.
func ContextUser(ctx context.Context) content.User {
	if account, ok := FromContext(ctx); ok {
		return account.ContentUser()
	}
	return content.Anonymous()
}
