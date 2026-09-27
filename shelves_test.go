package main

import (
	"testing"

	"github.com/ihleven/ihlvn/app/cmsauth"
	"github.com/interhome-group/cms/content"
)

// Every address the shelves serve, as they were served before the list replaced
// four copies of the wiring.
//
// Written out rather than derived, because a refactor's whole claim is that
// nothing moved, and a check that computes the answer the same way the code
// does cannot test that claim. If a row changes, this fails and the change has
// to be a deliberate edit here.
var shelfAddresses = []string{
	"GET /api/v1/mediathek/meta/{path...}",
	"GET /api/v1/mediathek/stream/{path...}",

	"GET /api/v1/retro/meta/{path...}",
	"GET /api/v1/retro/thumb/{path...}",
	"GET /api/v1/retro/media/{path...}",

	"GET /api/v1/musik/meta/{path...}",
	"GET /api/v1/musik/search/{path...}",
	"GET /api/v1/musik/tags/{path...}",
	"GET /api/v1/musik/thumb/{path...}",
	"GET /api/v1/musik/stream/{path...}",

	"GET /api/v1/djvet/meta/{path...}",
	"GET /api/v1/djvet/search/{path...}",
	"GET /api/v1/djvet/tags/{path...}",
	"GET /api/v1/djvet/thumb/{path...}",
	"GET /api/v1/djvet/stream/{path...}",

	"GET /api/v1/media/stream/{path...}",
}

func addresses(t *testing.T) []string {
	t.Helper()

	var got []string
	for _, s := range shelves(Flags{}) {
		for _, op := range s.ops {
			got = append(got, s.pattern(op))
		}
	}

	return got
}

func TestShelvesServeTheSameAddressesAsBefore(t *testing.T) {
	got := addresses(t)

	if len(got) != len(shelfAddresses) {
		t.Fatalf("got %d addresses, want %d:\n%v", len(got), len(shelfAddresses), got)
	}
	for i, want := range shelfAddresses {
		if got[i] != want {
			t.Errorf("address %d = %q, want %q", i, got[i], want)
		}
	}
}

// Which right guards which area, stated separately from the addresses.
//
// This is the pairing a hand-written route table gets wrong silently: gating
// the music shelf with the archive's entitlement is a plausible typo and
// nothing else in the program would notice. In a row the two sit together, and
// here they are checked apart.
func TestEachShelfDemandsItsOwnEntitlement(t *testing.T) {
	want := map[string]content.PermissionKey{
		"/api/v1/mediathek": cmsauth.Mediathek,
		"/api/v1/retro":     cmsauth.Retro,
		"/api/v1/musik":     cmsauth.Musik,
		"/api/v1/djvet":     cmsauth.Djvet,
		// The one shelf admitted by a capability rather than by an area, because
		// it is read while writing content rather than visited.
		"/api/v1/media": cmsauth.MediaRead,
	}

	seen := map[string]bool{}
	for _, s := range shelves(Flags{}) {
		expected, known := want[s.prefix]
		if !known {
			t.Errorf("a shelf is served at %q, which this test does not know about", s.prefix)

			continue
		}
		seen[s.prefix] = true
		if s.right != expected {
			t.Errorf("%s is guarded by %v, want %v", s.prefix, s.right, expected)
		}
		if s.refusal == "" {
			t.Errorf("%s refuses without saying why", s.prefix)
		}
	}
	for prefix := range want {
		if !seen[prefix] {
			t.Errorf("%s is no longer served", prefix)
		}
	}
}

// Each shelf is confined to its own directory, which is what keeps a mistake in
// one set of routes from reaching what another serves. Two sharing a root would
// be that mistake made at the table.
func TestNoTwoShelvesShareARoot(t *testing.T) {
	flags := Flags{
		MediathekRoot: "/public/mediathek",
		RetroRoot:     "/public/zeitschriften",
		MusikRoot:     "/public/musik",
		DjvetRoot:     "/public/djvet",
		MediaRoot:     "/public",
	}

	roots := map[string]string{}
	for _, s := range shelves(flags) {
		if s.root == "" {
			t.Errorf("%s has no root even though one was configured", s.prefix)

			continue
		}
		if other, taken := roots[s.root]; taken {
			t.Errorf("%s and %s are both rooted at %q", other, s.prefix, s.root)
		}
		roots[s.root] = s.prefix
	}
}

// Every operation a row asks for has to exist, or the program panics as it
// starts. Cheaper to find out here.
func TestEveryShelfOperationIsImplemented(t *testing.T) {
	for _, s := range shelves(Flags{}) {
		for _, op := range s.ops {
			if _, ok := libraryOps[op]; !ok {
				t.Errorf("%s asks for operation %q, which does not exist", s.prefix, op)
			}
		}
	}
}
