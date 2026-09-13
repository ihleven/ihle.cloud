package auth

import (
	"context"

	"github.com/interhome-group/cms/content"
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
