package hidrive

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ihleven/ihlvn/app/auth"
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
