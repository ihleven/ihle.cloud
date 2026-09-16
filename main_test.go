package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/interhome-group/cms/pkg/errs"
)

// Serving files from a directory must not serve files from outside it.
//
// This route is open — no account is required — and it runs with the working
// directory that holds .env, so a path that escapes reads the session key, the
// single-sign-on secret and the database password. It did: ServeMux redirects a
// literal "..", which is what made a Join look safe, but a percent-encoded one
// arrives already decoded in PathValue, and path.Join cleaned it into an escape.
func TestServingLocalFilesCannotLeaveItsDirectory(t *testing.T) {
	serve := serveContentWithPrefix("videos")

	for _, escape := range []string{
		"../go.mod",
		"../.env",
		"..%2f.env", // as PathValue delivers it, already decoded
		"a/../../go.mod",
		"/etc/passwd",
		"../../../../etc/passwd",
	} {
		r := httptest.NewRequest(http.MethodGet, "/media/videos/x", nil)
		r.SetPathValue("path", escape)
		w := httptest.NewRecorder()

		err := serve(w, r)
		if err == nil {
			t.Errorf("%q was served (%d bytes)", escape, w.Body.Len())
			continue
		}
		if got := errs.StatusOf(err); got != http.StatusNotFound {
			t.Errorf("%q gave status %d, want 404 — a probe should not learn the difference", escape, got)
		}
	}
}

// And the ordinary case still works, so the containment is not simply a refusal
// to serve anything.
func TestServingLocalFilesStillServesThem(t *testing.T) {
	serve := serveContentWithPrefix(".")

	r := httptest.NewRequest(http.MethodGet, "/media/videos/x", nil)
	r.SetPathValue("path", "go.mod")
	w := httptest.NewRecorder()

	if err := serve(w, r); err != nil {
		t.Fatalf("a file inside the directory was refused: %v", err)
	}
	if w.Body.Len() == 0 {
		t.Error("served nothing")
	}
}
