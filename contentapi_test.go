package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"

	"github.com/ihleven/ihlvn/app/auth"
	"github.com/ihleven/ihlvn/app/films"
	"github.com/interhome-group/cms/content"
	"github.com/interhome-group/cms/mgmt"
)

// nobody is a stand-in for the auth service that never recognises anyone.
//
// The gates have their own tests in app/gate; what these need an authenticator
// for is only that the mount takes one, so the smallest one that satisfies the
// interface is the honest fixture here.
type nobody struct{}

func (nobody) Authenticate(*http.Request) (*auth.Account, bool) { return nil, false }

// This app's content types, registered as main() registers them — a test binary
// does not run main, and without this the CMS's registry holds only its own.
func init() {
	content.Register(films.Film{})
}

// A repository with nothing in it, which the CMS treats as a place to start
// rather than an error. Enough to build a manager, which is all these tests
// need: what they are about is the mount, not the content.
func emptyRepo(t *testing.T) *mgmt.Mngr {
	t.Helper()

	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q", "-b", "main", "."},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "test"},
		{"commit", "-q", "--allow-empty", "-m", "init"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}

	cms, err := mgmt.NewContentManager(dir, func(conf *mgmt.Config) {
		conf.DataDir = t.TempDir()
		conf.RepoClone = false
	})
	if err != nil {
		t.Fatalf("content manager: %v", err)
	}

	return cms
}

// The mount is a subtree handler, and a subtree handler is where a prefix can
// quietly go missing: ServeMux passes the path through unchanged, and the
// patterns registered inside are absolute, so the two have to agree about the
// whole address rather than about a suffix. A request that arrives and is
// answered is the only way to know they do.
//
// The reply is also the thing this whole endpoint exists for: an application's
// own content types, which the CMS's editing UI cannot know and used to leave
// out.
func TestContentAPIServesThisAppsContentTypes(t *testing.T) {
	mounted := http.NewServeMux()
	handler, _ := contentAPI(emptyRepo(t), nobody{}, "test")
	mounted.Handle(contentAPIPrefix, handler)

	rec := httptest.NewRecorder()
	mounted.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/content-api/v1/types", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /content-api/v1/types = %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var types []struct {
		Name string `json:"name"`
		Ext  string `json:"ext"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &types); err != nil {
		t.Fatalf("decoding the answer: %v: %s", err, rec.Body.String())
	}

	for _, ct := range types {
		if ct.Name != "Film" {
			continue
		}
		// Stored as YAML, which is what the editor needs to name the file: a name
		// that disagrees with the storage form is written as one thing and read
		// back as another.
		if ct.Ext != ".yaml" {
			t.Errorf("Film.ext = %q, want %q", ct.Ext, ".yaml")
		}

		return
	}

	t.Fatalf("Film was not among the types offered: %s", rec.Body.String())
}

// Nobody signed in is not a reason to refuse: reading content is open in this
// app, and what needs an account refuses for itself. The middleware's job is to
// say who it is when there is somebody, not to be a gate.
func TestContentAPIAnswersWithoutAnAccount(t *testing.T) {
	handler, _ := contentAPI(emptyRepo(t), nobody{}, "test")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/content-api/v1/types", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("anonymous GET /types = %d, want 200", rec.Code)
	}
}

// Writing does need one, and the controllers enforce that themselves — which is
// worth pinning, because the middleware above deliberately does not.
func TestContentAPIRefusesAnAnonymousWrite(t *testing.T) {
	handler, _ := contentAPI(emptyRepo(t), nobody{}, "test")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/content-api/v1/entries", http.NoBody)
	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusOK {
		t.Errorf("anonymous POST /entries = 200, want a refusal")
	}
}
