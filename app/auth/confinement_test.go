package auth

import (
	"slices"
	"testing"

	"github.com/ihleven/ihlvn/app/cmsauth"
)

// A confined account is not governed by entitlements. What it is offered comes
// from its type, so a permission it picked up — or was granted by mistake —
// cannot widen it. That matters because what an account is offered and what it
// may reach are the same thing for these accounts: requireAccount refuses them
// everywhere else.
func TestOfferedIgnoresPermissionsForAConfinedAccount(t *testing.T) {
	s := &Service{cfg: Config{SharedDrive: true}}

	confined := &Account{
		Type: TypeGeheimtipp,
		// "*" expands to every registered permission, which is the strongest
		// thing an account can hold. It must still change nothing here.
		CMS: CMSProfile{Permissions: []string{"*"}},
	}

	got := s.offered(confined)
	if len(got) != 1 || got[0] != cmsauth.GeheimtippArea {
		t.Errorf("offered = %v, want exactly [%s]", got, cmsauth.GeheimtippArea)
	}
}

// The same permissions on a full account do what they always did.
func TestOfferedFollowsPermissionsForAFullAccount(t *testing.T) {
	s := &Service{cfg: Config{SharedDrive: true}}

	full := &Account{Type: TypeFull, CMS: CMSProfile{Permissions: []string{"*"}}}

	got := s.offered(full)
	if len(got) <= 1 {
		t.Fatalf("offered = %v, want every area", got)
	}
	if !slices.Contains(got, cmsauth.HidriveArea) {
		t.Errorf("offered = %v, want it to include %s", got, cmsauth.HidriveArea)
	}
}

// An account created the ordinary way is full, so nothing that exists today
// becomes confined by adding the column.
func TestAccountsAreFullUnlessSaidOtherwise(t *testing.T) {
	if (&Account{Type: TypeFull}).Confined() {
		t.Error("a full account reports as confined")
	}
	if !(&Account{Type: TypeGeheimtipp}).Confined() {
		t.Error("a geheimtipp account does not report as confined")
	}
}
