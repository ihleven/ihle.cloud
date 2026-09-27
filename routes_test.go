package main

import (
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/ihleven/ihlvn/pkg/srvr"
	"github.com/interhome-group/cms/mgmt"
)

// The familie and art APIs are the only dependencies a route *declaration*
// touches: the rest are pointers, and a method value on a nil pointer is taken
// without being called, while one on a nil interface is not. So these two are
// stood in for and nothing else needs to be.
type noAPI struct{}

func (noAPI) PersonHandler(http.ResponseWriter, *http.Request) error  { return nil }
func (noAPI) ReiseHandler(http.ResponseWriter, *http.Request) error   { return nil }
func (noAPI) ArtworkHandler(http.ResponseWriter, *http.Request) error { return nil }
func (noAPI) MappingHandler(http.ResponseWriter, *http.Request) error { return nil }
func (noAPI) Search(http.ResponseWriter, *http.Request) error         { return nil }

// declared is a deps holding only what a declaration reads.
//
// An empty manager rather than none: one route reads a field off it while being
// declared, and reading a field is a dereference where taking a method value is
// not. The CMS mount is a handler that answers nothing: mounting requires one,
// and what it answers is not what is being checked.
func declared() *deps {
	return &deps{
		fapi: noAPI{}, aapi: noAPI{},
		cms:        &mgmt.Mngr{},
		contentAPI: http.NotFoundHandler(),
	}
}

// Every route this application declares, registered together.
//
// Go's router refuses two patterns when a request would match both and neither
// is the more specific, and it refuses them by panicking as the server starts.
// Nothing before this point catches that: it is not a compile error, not a vet
// finding, and not something any test could reach while the table was built
// inside a function that needed a database first.
//
// It cost an afternoon once — a media route written under /films/ collided with
// /films/{id}/poster.jpg, because "/films/media/poster.jpg" matches both and
// neither wins. That is the whole reason routes() takes its dependencies as an
// argument rather than building them.
//
// The dependencies are zero here. Nothing is served, so nothing is called; what
// is being checked is the shape of the addresses, which no handler contributes
// to.
func TestEveryRouteCanBeRegistered(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("the routes cannot be registered together: %v", r)
		}
	}()

	srvr.New(routes(declared())...)
}

// A count, so that a route silently vanishing is a failure rather than a quiet
// improvement in the number above.
//
// Deliberately not a list of every address: they are declared a few feet away
// in a form that reads better than any restatement of them, and a second copy
// would be maintained by copying. What this pins is that the table is whole —
// if the number moves, the change was either meant or it was not, and either
// way somebody looks.
func TestTheRouteTableIsWhole(t *testing.T) {
	const want = 52

	if got := len(routes(declared())); got != want {
		t.Errorf("the application declares %d routes, want %d", got, want)
	}
}

// What the application serves is more than what its table declares.
//
// The CMS's API is mounted whole: its operations, its OpenAPI document and its
// documentation UI are registered by its own code on the mux it is handed. They
// are reachable addresses of this service, and nothing here wrote them down.
//
// Recording the registrations is what makes the answer true by construction —
// the alternative is a second list somebody keeps in step, which is the same as
// not knowing.
func TestTheListingIncludesWhatTheCMSMounted(t *testing.T) {
	d := declared()
	handler, mux := contentAPI(emptyRepo(t), nobody{}, "test")
	d.contentAPI, d.contentMux = handler, mux

	served := srvr.New(routes(d)...).Patterns()
	if len(served) == 0 {
		t.Fatal("nothing was recorded")
	}

	// The table's own routes are there.
	if !slices.ContainsFunc(served, func(p string) bool { return strings.Contains(p, "/api/v1/media/stream/") }) {
		t.Error("the media shelf is not in the listing")
	}

	// And so is the prefix the CMS was mounted under, which is all the main mux
	// can see of it.
	if !slices.Contains(served, strings.TrimSpace(contentAPIPrefix)) {
		t.Error("the CMS mount point is not in the listing")
	}

	// What it mounted underneath is the other mux's to report.
	whole := d.contentMux.Mounted()
	for _, want := range []string{
		"GET /content-api/v1/entries/{path...}",
		"GET /content-api/v1/types",
	} {
		if !slices.Contains(whole, want) {
			t.Errorf("%q was mounted but is not in the CMS listing", want)
		}
	}

	// Including what huma registered for itself, which no table anywhere
	// declares and which a hand-kept list would always have missed.
	if !slices.ContainsFunc(whole, func(p string) bool { return strings.Contains(p, "/openapi") }) {
		t.Error("huma's own endpoints were not recorded")
	}
}
