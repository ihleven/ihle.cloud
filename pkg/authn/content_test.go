package authn

import (
	"slices"
	"testing"

	"github.com/interhome-group/cms/content"
)

// The id the CMS compares against an entry's owner is the login name. Content in
// this repository is owned by names like "matt", not by email addresses — the
// CMS's own deployment uses the email there, which is why this mapping belongs
// to the application.
func TestContentUserIsIdentifiedByLoginName(t *testing.T) {
	a := &Account{
		Name:        "matt",
		DisplayName: "Matthias",
		Email:       "matthias@example.test",
		CMS:         CMSProfile{Groups: []string{"familie"}},
	}

	user := a.ContentUser()

	if user.ID != "matt" {
		t.Errorf("ID = %q, want the login name", user.ID)
	}
	if user.IsAnonymous() {
		t.Error("a real account mapped to the anonymous user")
	}
	if !slices.Contains(user.Groups, "familie") {
		t.Errorf("Groups = %v, want the account's groups", user.Groups)
	}
	// The signature is what a commit is authored with.
	if user.Signature.Name != "Matthias" || user.Signature.Email != "matthias@example.test" {
		t.Errorf("Signature = %+v, want the display name and email", user.Signature)
	}
	if user.Scope == nil {
		t.Error("Scope is nil; adding to it would panic")
	}
}

// An account with no cmsauth row holds no CMS rights, but is still a real,
// non-anonymous identity — it owns what it creates and signs its commits.
func TestContentUserWithoutCMSProfile(t *testing.T) {
	a := &Account{Name: "wolfgang", DisplayName: "Wolfgang", Email: "w@example.test"}

	user := a.ContentUser()

	if user.ID != "wolfgang" || user.IsAnonymous() {
		t.Errorf("expected a real identity, got %+v", user)
	}
	if user.Scope.Denies(content.ENTRY_CREATE) != true {
		t.Error("an account with no permissions should be denied entry.create")
	}
}

// Permissions the CMS does not know are dropped silently when a scope is built,
// so they are reported rather than left to look effective.
//
// What counts as known depends on which packages the binary links: assets.* are
// defined by mgmt/asset, so a binary that does not import it would rightly
// report them here.
func TestUnregisteredPermissionsAreReported(t *testing.T) {
	a := &Account{
		Name: "matt",
		CMS: CMSProfile{Permissions: []string{
			"entry.create",        // defined by the content package
			"entry.create:/pages", // the same, carrying a path constraint
			"meta.name.wrt",       // the CMS commented this one out
			"totally.made.up",
		}},
	}

	got := a.UnregisteredPermissions()

	if slices.Contains(got, "entry.create") {
		t.Errorf("a registered permission was reported as unknown: %v", got)
	}
	if slices.Contains(got, "entry.create:/pages") {
		t.Errorf("a constrained permission was reported as unknown; only the name is registered: %v", got)
	}
	for _, want := range []string{"meta.name.wrt", "totally.made.up"} {
		if !slices.Contains(got, want) {
			t.Errorf("%q was not reported as unregistered; got %v", want, got)
		}
	}
}
