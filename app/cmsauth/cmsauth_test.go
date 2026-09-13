package cmsauth

import (
	"slices"
	"testing"
)

// Permissions the CMS does not know are dropped silently when a scope is built,
// so they are reported rather than left to look effective.
//
// What counts as known depends on which packages the binary links: assets.* are
// defined by mgmt/asset, so a binary that does not import it would rightly
// report them here.
func TestUnregisteredPermissionsAreReported(t *testing.T) {
	got := Unregistered([]string{
		"entry.create",        // defined by the content package
		"entry.create:/pages", // the same, carrying a path constraint
		"meta.name.wrt",       // the CMS commented this one out
		"totally.made.up",
	})

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

// "*" is a grant-all that content.Scope expands at build time, so reporting it
// as a name this build does not define told the operator the opposite of what
// it does: it grants everything, including every area.
func TestTheWildcardIsNotReportedAsUnregistered(t *testing.T) {
	permissions := []string{"*"}

	if unknown := Unregistered(permissions); len(unknown) != 0 {
		t.Errorf("Unregistered() = %v, want none: \"*\" grants everything", unknown)
	}
}

// The wildcard is matched as the entire permission string. A constrained form
// is not expanded and really does grant nothing, so it must still be reported —
// that is the case a laxer check would hide.
func TestAConstrainedWildcardIsStillReported(t *testing.T) {
	permissions := []string{"*:de"}

	if unknown := Unregistered(permissions); !slices.Contains(unknown, "*:de") {
		t.Errorf("Unregistered() = %v, want it to report \"*:de\"", unknown)
	}
}

// A genuinely unknown name is still reported; the wildcard exemption must not
// widen into "anything goes".
func TestAnUnknownPermissionIsStillReported(t *testing.T) {
	permissions := []string{"module.nosuchthing"}

	if unknown := Unregistered(permissions); !slices.Contains(unknown, "module.nosuchthing") {
		t.Errorf("Unregistered() = %v, want it to report the unknown name", unknown)
	}
}
