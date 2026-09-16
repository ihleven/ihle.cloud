package hidrive

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ihleven/ihlvn/app/auth"
	"github.com/ihleven/ihlvn/pkg/hi"
	"github.com/interhome-group/cms/pkg/errs"
)

// request carries an account with the given alias and root, as the middleware
// in front of these handlers would.
func request(alias, root string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/api/v1/drive/meta/Bilder", nil)
	a := &auth.Account{Name: "matt"}
	a.HiDrive.Alias = alias
	a.HiDrive.Root = root
	return r.WithContext(auth.WithAccount(r.Context(), a))
}

// Whose tree an entitled account browses: their own alias when they have one,
// the deployment's when they do not. The entitlement itself is the middleware's
// business and is not re-asked here.
func TestDriveResolutionFollowsTheAccountThenTheDeployment(t *testing.T) {
	api := NewAPI(AccessTokens{}, Shared{Alias: "ihleven", Root: "/public"})

	own, err := api.drive(request("matt.ihle", "/users/matt"))
	if err != nil {
		t.Fatal(err)
	}
	if got := own.Config().Alias; got != "matt.ihle" {
		t.Errorf("alias = %q, want the account's own", got)
	}

	shared, err := api.drive(request("", ""))
	if err != nil {
		t.Fatal(err)
	}
	if got := shared.Config().Alias; got != "ihleven" {
		t.Errorf("alias = %q, want the deployment's", got)
	}
	// The root travels with the alias: falling back to the shared drive must not
	// leave the account's root in place, which would address one tree inside
	// another.
	if got := shared.Config().Root; got != "/public" {
		t.Errorf("root = %q, want the deployment's", got)
	}
}

// A deployment that has configured no storage at all answers per request, so
// the route still visibly exists rather than being silently absent.
func TestNoStorageConfiguredIsReportedPerRequest(t *testing.T) {
	api := NewAPI(AccessTokens{}, Shared{})

	_, err := api.drive(request("", ""))
	if err == nil {
		t.Fatal("a request with no alias anywhere was accepted")
	}
	if got := errs.StatusOf(err); got != http.StatusNotImplemented {
		t.Errorf("status = %d, want %d", got, http.StatusNotImplemented)
	}
}

// Arriving without an account is a wiring mistake: these handlers are always
// mounted behind the middleware that requires one.
func TestNoAccountIsRefused(t *testing.T) {
	api := NewAPI(AccessTokens{}, Shared{Alias: "ihleven"})

	r := httptest.NewRequest(http.MethodGet, "/api/v1/drive/meta/x", nil)
	if _, err := api.drive(r); errs.StatusOf(err) != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", errs.StatusOf(err), http.StatusUnauthorized)
	}
}

// The width a caller asks for decides how much work the store does, so it is
// rounded to a fixed set rather than passed on.
func TestThumbnailWidthIsClamped(t *testing.T) {
	for _, tt := range []struct{ asked, want string }{
		{"", "100"},
		{"1", "100"},
		{"100", "100"},
		{"101", "200"},
		{"200", "200"},
		{"201", "400"},
		{"800", "800"},
		{"4000", "800"},
		{"nonsense", "100"},
	} {
		got := thumbSize(url.Values{"width": {tt.asked}}).Get("width")
		if got != tt.want {
			t.Errorf("width %q -> %q, want %q", tt.asked, got, tt.want)
		}
	}
}

// The path may arrive on the route or as a query parameter; the query wins,
// which is how the thumbnail endpoint is addressed.
func TestPathPrefersTheQuery(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/v1/drive/thumb?path=Bilder/a.jpg", nil)
	if got := path(r); got != "Bilder/a.jpg" {
		t.Errorf("path = %q, want the query's", got)
	}
}

// token stands in for the refresher, which needs a database.
type token string

func (t token) AccessToken(string) (string, error) { return string(t), nil }

// A drive URL does not identify a file: the path is resolved against the asking
// account's drive. A browser cache keyed on the URL alone would therefore show
// one account's image to the next person signed in on the same machine, which
// is what Vary prevents.
//
// Driven against a stub store, because these headers are set on the way out of
// a successful response — deliberately, since setting them earlier would let an
// error be cached for the hour the success was meant to be.
func TestThumbnailsVaryOnTheSession(t *testing.T) {
	store := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write([]byte("\xff\xd8\xff"))
	}))
	defer store.Close()

	api := NewAPI(token("t"), Shared{Alias: "ihleven", Root: "/public"})
	api.client = &hi.Client{BaseURL: store.URL}

	w := httptest.NewRecorder()
	if err := api.Thumbnail(w, request("matt.ihle", "/users/matt")); err != nil {
		t.Fatal(err)
	}

	if got := w.Header().Get("Vary"); got != "Cookie" {
		t.Errorf("Vary = %q, want Cookie — this response is account-specific", got)
	}
	if got := w.Header().Get("Cache-Control"); !strings.Contains(got, "private") {
		t.Errorf("Cache-Control = %q, want it private", got)
	}
	if w.Code != http.StatusOK {
		t.Errorf("status %d", w.Code)
	}
}

// The streaming route is account-specific too, and its headers are set before
// the store is reached, so the same check needs no stub.
func TestStreamingVariesOnTheSession(t *testing.T) {
	api := NewAPI(AccessTokens{}, Shared{Alias: "ihleven", Root: "/public"})

	w := httptest.NewRecorder()
	// Reaching the store needs a network call; the headers that matter are set
	// before it, so a failure here is expected and not the subject.
	_ = api.Stream(w, request("matt.ihle", "/users/matt"))

	if got := w.Header().Get("Vary"); got != "Cookie" {
		t.Errorf("Vary = %q, want Cookie", got)
	}
}

// The library is the opposite case: one fixed shelf, identical for everybody,
// so its URLs do identify their files and must not be split per session.
func TestTheLibraryDoesNotVary(t *testing.T) {
	lib := library("/public/mediathek")

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/mediathek/stream/a.mp4", nil)
	r.SetPathValue("path", "a.mp4")
	_ = lib.Stream(w, r)

	if got := w.Header().Get("Vary"); got != "" {
		t.Errorf("Vary = %q; the library looks the same to everyone, so keying a cache per session only wastes it", got)
	}
}

// A listing and the bytes have to agree about what a path means. They reach the
// store by different routes — Resolve for one, SafeKey for the other — and those
// two disagreed about a leading slash: one read the file, the other refused it
// as an unsafe key. Normalising once, here, is what keeps them saying the same
// thing.
func TestALeadingSlashMeansTheSameThingEverywhere(t *testing.T) {
	for _, spelling := range []string{"a/b.mp4", "/a/b.mp4", "//a/b.mp4"} {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/drive/meta/x", nil)
		r.SetPathValue("path", spelling)

		if got := path(r); got != "a/b.mp4" {
			t.Errorf("%q became %q, want a/b.mp4", spelling, got)
		}
	}

	// And through the query parameter, which is the other way in.
	r := httptest.NewRequest(http.MethodGet, "/api/v1/drive/thumb?path=/a/b.jpg", nil)
	if got := path(r); got != "a/b.jpg" {
		t.Errorf("query path became %q, want a/b.jpg", got)
	}
}

// Stripping the slash must not change where a path lands, only whether it is
// accepted: the root is still what contains it.
func TestNormalisingAPathDoesNotMoveIt(t *testing.T) {
	drive := hi.NewDrive(token("t"), hi.DriveConfig{Alias: "a", Root: "/public/mediathek"}, nil)

	if with, without := drive.Resolve("/x/y.mp4"), drive.Resolve("x/y.mp4"); with != without {
		t.Errorf("%q and %q resolve differently", with, without)
	}
	if got := drive.Resolve("x/y.mp4"); got != "/public/mediathek/x/y.mp4" {
		t.Errorf("resolved to %q", got)
	}
}
