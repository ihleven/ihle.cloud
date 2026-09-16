package hidrive

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

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

// store stands in for HiDrive: it signs a URL back to itself, serves the bytes
// behind that URL with the standard library's own range handling, and renders a
// thumbnail with whatever status the test asks for.
func store(t *testing.T, content []byte, thumbStatus int) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/file/url", func(w http.ResponseWriter, r *http.Request) {
		// The signed URL points back here, which is what the handler under test
		// then proxies to.
		signed := "http://" + r.Host + "/signed?path=" + url.QueryEscape(r.URL.Query().Get("path"))
		_ = json.NewEncoder(w).Encode(map[string]string{"url": signed})
	})
	mux.HandleFunc("/signed", func(w http.ResponseWriter, r *http.Request) {
		// ServeContent answers ranges the way a real store does, several at a
		// time included.
		http.ServeContent(w, r, "issue.pdf", time.Time{}, bytes.NewReader(content))
	})
	mux.HandleFunc("/file/thumbnail", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Last-Modified", "Tue, 01 Oct 2024 00:00:00 GMT")
		w.Header().Set("Cache-Control", "max-age=5")
		w.WriteHeader(thumbStatus)
		_, _ = w.Write([]byte("jpeg-bytes"))
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return server
}

func libraryOf(server *httptest.Server, root string) *Library {
	return NewLibrary(
		hi.NewDrive(token("t"), hi.DriveConfig{Alias: "ihleven", Root: root}, &hi.Client{BaseURL: server.URL}),
		nil,
	)
}

// The reason Media exists at all.
//
// A reader opening a large document asks for several pieces of it at once, and
// only the store can answer that: Stream serves one range per request and gives
// the whole file for anything else, which for a ninety-megabyte scan is the
// difference between a page and a magazine. The archive's routes were on the
// wrong one of the two until this was noticed.
func TestMediaPassesSeveralRangesToTheStore(t *testing.T) {
	content := bytes.Repeat([]byte("0123456789"), 500)
	lib := libraryOf(store(t, content, http.StatusOK), "/public/zeitschriften")

	r := httptest.NewRequest(http.MethodGet, "/api/v1/retro/media/x.pdf", nil)
	r.SetPathValue("path", "HappyComputer/x.pdf")
	r.Header.Set("Range", "bytes=0-99,200-299")

	w := httptest.NewRecorder()
	if err := lib.Media(w, r); err != nil {
		t.Fatalf("serving: %v", err)
	}

	if w.Code != http.StatusPartialContent {
		t.Errorf("status %d, want 206", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "multipart/byteranges") {
		t.Errorf("content type %q, want multipart/byteranges — the ranges were not answered separately", ct)
	}
	if n := w.Body.Len(); n >= len(content) {
		t.Errorf("answered with %d bytes of a %d byte file: the whole thing rather than the parts asked for", n, len(content))
	}
}

// One range is the ordinary case and must still work, as a player's seek does.
func TestMediaServesASingleRange(t *testing.T) {
	content := bytes.Repeat([]byte("abcdefghij"), 100)
	lib := libraryOf(store(t, content, http.StatusOK), "/public/zeitschriften")

	r := httptest.NewRequest(http.MethodGet, "/api/v1/retro/media/x.pdf", nil)
	r.SetPathValue("path", "x.pdf")
	r.Header.Set("Range", "bytes=10-19")

	w := httptest.NewRecorder()
	if err := lib.Media(w, r); err != nil {
		t.Fatal(err)
	}

	if w.Code != http.StatusPartialContent {
		t.Fatalf("status %d, want 206", w.Code)
	}
	if got := w.Body.String(); got != "abcdefghij" {
		t.Errorf("body %q", got)
	}
	if got := w.Header().Get("Content-Range"); got != "bytes 10-19/1000" {
		t.Errorf("content range %q", got)
	}
}

// A thumbnail is the store's answer, passed on: its status, because writing 200
// over anything else would cache an error as if it were the picture, and its
// Last-Modified, so a browser can ask whether it still holds.
func TestThumbnailPassesTheStoresAnswerOn(t *testing.T) {
	lib := libraryOf(store(t, nil, http.StatusOK), "/public/zeitschriften")

	r := httptest.NewRequest(http.MethodGet, "/api/v1/retro/thumb/x.pdf?width=400", nil)
	r.SetPathValue("path", "HappyComputer/x.pdf")

	w := httptest.NewRecorder()
	if err := lib.Thumbnail(w, r); err != nil {
		t.Fatal(err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("status %d", w.Code)
	}
	if got := w.Header().Get("Content-Type"); got != "image/jpeg" {
		t.Errorf("content type %q", got)
	}
	if got := w.Header().Get("Last-Modified"); got == "" {
		t.Error("the store's validator was dropped, so a browser must refetch rather than revalidate")
	}
	// The store says five seconds, which is no use for a picture of a file that
	// has not changed; ours is the one that should reach the browser.
	if got := w.Header().Get("Cache-Control"); got != "private, max-age=3600" {
		t.Errorf("cache-control %q", got)
	}
	// Nothing about who is asking may vary the answer: the library is one shelf.
	if got := w.Header().Get("Vary"); got != "" {
		t.Errorf("Vary %q, want none", got)
	}
}

// A thumbnail the store refuses is an error, not a picture.
//
// It never reaches the passing-on above: the client treats a refusal as a
// failure, so the handler returns one and the route reports it. What matters is
// that nothing is written — an empty 200 would be cached for an hour as though
// it were the cover.
func TestARefusedThumbnailIsNotServedAsAnImage(t *testing.T) {
	lib := libraryOf(store(t, nil, http.StatusNotFound), "/public/zeitschriften")

	r := httptest.NewRequest(http.MethodGet, "/api/v1/retro/thumb/gone.pdf", nil)
	r.SetPathValue("path", "gone.pdf")

	w := httptest.NewRecorder()
	err := lib.Thumbnail(w, r)
	if err == nil {
		t.Fatal("a refused thumbnail was reported as success")
	}
	if w.Body.Len() != 0 {
		t.Errorf("wrote %d bytes of a picture that does not exist", w.Body.Len())
	}
}

// searchStore stands in for the endpoint, answering with hits spelled the way
// HiDrive spells them: absolute, and percent-encoded a segment at a time.
func searchStore(t *testing.T, hits string) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"result":` + hits + `}`))
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return server
}

// What the handler is for, beyond passing the call along.
//
// A hit names itself absolutely, so it arrives with the library's root in front
// of it — and a caller that fed that back to any of these routes would be
// asking for a path below the root twice over. It also arrives encoded, which a
// caller building a URL would encode again.
func TestSearchAddressesHitsTheWayItsOwnRoutesDo(t *testing.T) {
	lib := libraryOf(searchStore(t, `[
		{"path":"/public/alben","type":"dir","category":"directory"},
		{"path":"/public/alben/1987%20-%20Appetite%20For%20Destruction%20%28lameV3A%29","type":"dir","category":"directory","nmembers":14},
		{"path":"/public/alben/Trash/01.%20Poison.mp3","type":"file","category":"file","size":5617680}
	]`), "/public/alben")

	r := httptest.NewRequest(http.MethodGet, "/api/v1/musik/search/?category=dir", nil)
	r.SetPathValue("path", "")

	w := httptest.NewRecorder()
	if err := lib.Search(w, r); err != nil {
		t.Fatalf("searching: %v", err)
	}

	var got struct {
		Result []hi.Meta `json:"result"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("reading the answer: %v", err)
	}

	want := []string{
		// The shelf itself, which is the root: below it lies nothing, so it
		// keeps the only name it has.
		"/public/alben",
		"1987 - Appetite For Destruction (lameV3A)",
		"Trash/01. Poison.mp3",
	}
	if len(got.Result) != len(want) {
		t.Fatalf("got %d hits, want %d", len(got.Result), len(want))
	}
	for i, path := range want {
		if got.Result[i].Path != path {
			t.Errorf("hit %d is %q, want %q", i, got.Result[i].Path, path)
		}
	}
}

// A path that is not below the root is left absolute rather than trimmed to
// something that looks relative and is not — there is no relative name for a
// file the library cannot reach.
func TestSearchLeavesAPathOutsideTheRootAlone(t *testing.T) {
	lib := libraryOf(searchStore(t, `[
		{"path":"/public/alben-archiv/x.mp3","type":"file"},
		{"path":"/public/albendings","type":"dir"}
	]`), "/public/alben")

	r := httptest.NewRequest(http.MethodGet, "/api/v1/musik/search/", nil)
	r.SetPathValue("path", "")

	w := httptest.NewRecorder()
	if err := lib.Search(w, r); err != nil {
		t.Fatalf("searching: %v", err)
	}

	var got struct {
		Result []hi.Meta `json:"result"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("reading the answer: %v", err)
	}
	for _, hit := range got.Result {
		if !strings.HasPrefix(hit.Path, "/") {
			t.Errorf("%q was made to look like a path inside the library", hit.Path)
		}
	}
}

// The selection is the caller's, and it is the whole of what the endpoint does:
// passing neither through would search for nothing, since the API enumerates
// only directories and only when asked by category.
func TestSearchPassesTheSelectionThrough(t *testing.T) {
	var asked url.Values

	mux := http.NewServeMux()
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		asked = r.URL.Query()
		_, _ = w.Write([]byte(`{"result":[]}`))
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	lib := libraryOf(server, "/public/alben")

	r := httptest.NewRequest(http.MethodGet, "/api/v1/musik/search/Trash?pattern=mp3&category=dir", nil)
	r.SetPathValue("path", "Trash")

	w := httptest.NewRecorder()
	if err := lib.Search(w, r); err != nil {
		t.Fatalf("searching: %v", err)
	}

	if got := asked.Get("pattern"); got != "mp3" {
		t.Errorf("pattern %q, want mp3", got)
	}
	if got := asked.Get("category"); got != "dir" {
		t.Errorf("category %q, want dir", got)
	}
	if got := asked.Get("path"); got != "/public/alben/Trash" {
		t.Errorf("searched %q, want the path resolved below the root", got)
	}
}

// An unconfigured shelf answers rather than disappearing, the way the rest of
// the library's routes do.
func TestSearchSaysWhenNoShelfIsConfigured(t *testing.T) {
	var absent *Library

	err := absent.Search(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/v1/musik/search/", nil))
	if err == nil {
		t.Fatal("an unconfigured library searched something")
	}
	if code := errs.StatusOf(err); code != http.StatusNotImplemented {
		t.Errorf("status %d, want 501", code)
	}
}
