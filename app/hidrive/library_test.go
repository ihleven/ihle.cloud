package hidrive

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ihleven/ihlvn/pkg/hi"
	"github.com/interhome-group/cms/pkg/errs"
)

func library(root string) *Library {
	return NewLibrary(hi.NewDrive(token("t"), hi.DriveConfig{Alias: "ihleven", Root: root}, nil), nil)
}

// The property the whole type exists for: no account takes part in deciding
// what is served. An account in the request must make no difference, because if
// it could, the URL would stop identifying the file and the responses would
// stop being cacheable.
func TestTheLibraryIgnoresWhoIsAsking(t *testing.T) {
	lib := library("/public/mediathek")

	anonymous := httptest.NewRequest(http.MethodGet, "/api/v1/mediathek/stream/a.mp4", nil)
	anonymous.SetPathValue("path", "a.mp4")

	// The same request, but carrying an account with a drive of its own — which
	// is what the browser's handlers would follow instead of the shared one.
	withAccount := request("matt.ihle", "/users/matt")
	withAccount.SetPathValue("path", "a.mp4")

	if got := lib.drive.Resolve(path(anonymous)); got != "/public/mediathek/a.mp4" {
		t.Errorf("resolved to %q", got)
	}
	if got := lib.drive.Resolve(path(withAccount)); got != "/public/mediathek/a.mp4" {
		t.Errorf("an account changed where the library looks: %q", got)
	}
}

// The root is the containment. Nothing outside it is addressable, however the
// caller spells the path — which is what keeps a library route away from the
// films and the browser's trees even though all three share an alias.
func TestTheLibraryRootCannotBeEscaped(t *testing.T) {
	lib := library("/public/mediathek")

	for _, p := range []string{
		"../filme/secret.mp4",
		"../../etc/passwd",
		"/etc/passwd",
		"a/../../b.mp4",
	} {
		got := lib.drive.Resolve(p)
		if len(got) < len("/public/mediathek") || got[:len("/public/mediathek")] != "/public/mediathek" {
			t.Errorf("%q escaped to %q", p, got)
		}
	}
}

// A deployment that has not configured a library says so, rather than the route
// looking absent — the same way the browser reports having no drive.
func TestAnUnconfiguredLibraryReportsItself(t *testing.T) {
	var absent *Library

	for _, serve := range []func(http.ResponseWriter, *http.Request) error{absent.Meta, absent.Stream} {
		err := serve(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/x", nil))
		if err == nil {
			t.Fatal("an unconfigured library served a request")
		}
		if got := errs.StatusOf(err); got != http.StatusNotImplemented {
			t.Errorf("status %d, want %d", got, http.StatusNotImplemented)
		}
	}
}

// One handler, kept, so the metadata cache inside it survives between requests.
// Rebuilding it per request would discard the cache and make this route cost
// what the pre-signed-URL route costs.
func TestTheLibraryKeepsOneStreamHandler(t *testing.T) {
	lib := library("/public/mediathek")

	if lib.stream == nil {
		t.Fatal("no stream handler was built")
	}
	// Built once at construction: there is no per-request path that could
	// replace it, which is the point.
	before := lib.stream
	_ = lib.Stream(httptest.NewRecorder(), func() *http.Request {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/mediathek/stream/", nil)
		r.SetPathValue("path", "")
		return r
	}())
	if lib.stream != before {
		t.Error("serving a request replaced the handler")
	}
}
