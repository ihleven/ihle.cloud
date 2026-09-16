package hidrive

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ihleven/ihlvn/app/auth"
	"github.com/ihleven/ihlvn/pkg/stream"
	"github.com/interhome-group/cms/pkg/errs"
)

// The cache inside a stream handler answers object metadata for a minute, which
// is the whole reason this route is cheaper than Media. Building the handler per
// request would throw that cache away before it ever answered twice — the code
// would look right and measure exactly like the thing it replaced.
func TestTheStreamHandlerOutlivesTheRequest(t *testing.T) {
	api := NewAPI(AccessTokens{}, Shared{Alias: "ihleven", Root: "/public"})

	first, err := api.drive(request("matt.ihle", "/users/matt"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := api.drive(request("matt.ihle", "/users/matt"))
	if err != nil {
		t.Fatal(err)
	}

	if api.streamer(first) != api.streamer(second) {
		t.Error("a second request built a second handler, discarding the first one's cache")
	}
}

// Keyed by the whole drive configuration, not just the alias: the cache inside
// is keyed by the caller's path *before* the root is applied, so sharing a
// handler across roots would serve one account another's metadata.
func TestDrivesWithDifferentRootsDoNotShareAHandler(t *testing.T) {
	api := NewAPI(AccessTokens{}, Shared{Alias: "ihleven", Root: "/public"})

	mine, err := api.drive(request("shared", "/users/matt"))
	if err != nil {
		t.Fatal(err)
	}
	theirs, err := api.drive(request("shared", "/users/wolfgang"))
	if err != nil {
		t.Fatal(err)
	}

	if api.streamer(mine) == api.streamer(theirs) {
		t.Error("two roots under one alias share a metadata cache")
	}
}

// A deployment with no drive configured answers the same way here as on the
// other three routes: the route exists and says why it cannot serve, rather
// than looking absent.
func TestStreamingWithoutAConfiguredDriveIsNotImplemented(t *testing.T) {
	api := NewAPI(AccessTokens{}, Shared{})

	err := api.Stream(httptest.NewRecorder(), request("", ""))
	if err == nil {
		t.Fatal("served without a drive")
	}
	if got := errs.StatusOf(err); got != http.StatusNotImplemented {
		t.Errorf("status %d, want %d", got, http.StatusNotImplemented)
	}
}

// The route addresses storage directly, so the public id is the key. An empty
// one is a missing item rather than a malformed key, which keeps a bare URL out
// of the log as an accusation.
func TestDrivePathIsTheIdentityMappingExceptForEmpty(t *testing.T) {
	if _, err := drivePath(context.Background(), ""); err != stream.ErrNoSuchItem {
		t.Errorf("empty path gave %v, want ErrNoSuchItem", err)
	}

	got, err := drivePath(context.Background(), "mediathek/2003-sea-monsters/1.mp4")
	if err != nil {
		t.Fatal(err)
	}
	if want := "mediathek/2003-sea-monsters/1.mp4"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// The handler reads its item from the "id" path value; this route is declared
// with {path...} to match its siblings, so Stream has to restate it under that
// name. If it ever stops, the handler asks for an empty id and every file is a
// 404 — silently, since an empty id is a legitimate question to ask.
func TestStreamPassesThePathUnderTheNameTheHandlerReads(t *testing.T) {
	api := NewAPI(AccessTokens{}, Shared{Alias: "ihleven", Root: "/public"})

	r := httptest.NewRequest(http.MethodGet, "/api/v1/drive/stream/mediathek/a.mp4", nil)
	account := &auth.Account{Name: "matt"}
	account.HiDrive.Alias = "ihleven"
	r = r.WithContext(auth.WithAccount(r.Context(), account))
	r.SetPathValue("path", "mediathek/a.mp4")

	d, err := api.drive(r)
	if err != nil {
		t.Fatal(err)
	}

	// Stand in for the real handler, so the wiring is observed without a
	// network call to the store behind it.
	var seen string
	api.streams = map[string]http.Handler{
		streamKey(d): http.HandlerFunc(func(_ http.ResponseWriter, got *http.Request) {
			seen = got.PathValue("id")
		}),
	}

	if err := api.Stream(httptest.NewRecorder(), r); err != nil {
		t.Fatal(err)
	}
	if want := "mediathek/a.mp4"; seen != want {
		t.Errorf("the handler was asked for %q, want %q", seen, want)
	}
}
