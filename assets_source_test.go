package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/interhome-group/cms/pkg/godoc"
)

func docs() *godoc.Handler {
	return godoc.New(appSource, ".", "github.com/ihleven/ihlvn", "/godoc")
}

// The embed and the renderer have to agree: a documentation server pointed at a
// source tree that does not contain this application's packages renders an
// empty page and no error, which looks like it works.
func TestGodocFindsThisApplicationsPackages(t *testing.T) {
	w := httptest.NewRecorder()
	if err := docs().Index(w, httptest.NewRequest(http.MethodGet, "/godoc", nil)); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("got %d", w.Code)
	}

	body := w.Body.String()
	for _, want := range []string{
		"github.com/ihleven/ihlvn",          // the module root, this package
		"github.com/ihleven/ihlvn/app/auth", // an app package
		"github.com/ihleven/ihlvn/pkg/api",  // a pkg package
	} {
		if !strings.Contains(body, want) {
			t.Errorf("%s is missing from the index — check the embed patterns", want)
		}
	}
}

// One package, rendered. This is the path the "pkg" value drives, so it also
// confirms the route pattern the server mounts is the one the handler expects.
func TestGodocRendersOnePackage(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/godoc/app/auth", nil)
	r.SetPathValue("pkg", "app/auth")

	w := httptest.NewRecorder()
	if err := docs().Package(w, r); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "github.com/ihleven/ihlvn/app/auth") {
		t.Error("the page does not name the package it rendered")
	}
}

// A package that does not exist is a 404, not a read of whatever the URL says:
// the handler serves only what it discovered, and the source tree is in the
// binary rather than behind a path someone could aim elsewhere.
func TestGodocRefusesAnUnknownPackage(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/godoc/../../etc", nil)
	r.SetPathValue("pkg", "../../etc")

	if err := docs().Package(httptest.NewRecorder(), r); err == nil {
		t.Error("a package outside the discovered set was served")
	}
}
