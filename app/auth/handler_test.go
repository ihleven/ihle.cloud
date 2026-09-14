package auth

import (
	"slices"
	"testing"
)

// The file browser is the one area whose content is a drive, so being entitled
// to it is not by itself a reason to offer it: an account with no drive of its
// own, on a deployment with no shared one, has nothing to browse. Withholding
// the area is indistinguishable from not being entitled, which is what someone
// in that position should see — an entry that leads to an error is worse than
// no entry.
func TestTheFileBrowserIsOnlyOfferedWhenThereIsADriveToBrowse(t *testing.T) {
	entitled := &Account{CMS: CMSProfile{Permissions: []string{"*"}}}

	withOwnDrive := &Account{CMS: CMSProfile{Permissions: []string{"*"}}}
	withOwnDrive.HiDrive.Alias = "matt.ihle"

	tests := []struct {
		name        string
		account     *Account
		sharedDrive bool
		wantOffered bool
	}{
		{"its own drive, no shared one", withOwnDrive, false, true},
		{"its own drive and a shared one", withOwnDrive, true, true},
		{"no drive, but the deployment has one", entitled, true, true},
		{"no drive anywhere", entitled, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{cfg: Config{SharedDrive: tt.sharedDrive}}

			offered := s.offered(tt.account)
			if got := slices.Contains(offered, "hidrive"); got != tt.wantOffered {
				t.Errorf("hidrive offered = %v, want %v (offered: %v)", got, tt.wantOffered, offered)
			}

			// Withholding one area must not disturb the others.
			for _, id := range []string{"filme", "familie", "search"} {
				if !slices.Contains(offered, id) {
					t.Errorf("%q was withheld too", id)
				}
			}
		})
	}
}
