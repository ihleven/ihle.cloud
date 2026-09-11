package authn

import (
	"testing"

	"github.com/interhome-group/cms/content"
)

// The CMS reads the request's account through content.UserProvider. An account
// here satisfies it without an adapter, which is what makes the CMS's handlers
// usable from this app.
func TestAccountSatisfiesContentUserProvider(t *testing.T) {
	var _ content.UserProvider = (*Account)(nil)
}
