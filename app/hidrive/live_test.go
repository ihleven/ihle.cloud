package hidrive

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ihleven/ihlvn/app/db"
	"github.com/ihleven/ihlvn/pkg/hi"
)

// Live checks against the real HiDrive, for the questions that cannot be
// answered from the source: what the store actually sends back.
//
// Skipped unless HIDRIVE_LIVE names an alias, because it needs a real token out
// of the database and one round trip per call. Read-only throughout — Meta,
// Thumbnail, URL and ranged reads — so it cannot disturb what it looks at.
//
// Paths are given the way the routes take them: relative to the configured
// root, never with a leading slash — blob.SafeKey refuses one, which is how the
// root stays the boundary.
//
//	HIDRIVE_LIVE=ihleven HIDRIVE_LIVE_ROOT=/public/mediathek \
//	  HIDRIVE_LIVE_FILE=2003-sea-monsters/SEAMONSTERS_DE.mp4 \
//	  go test ./app/hidrive -run Live -v
func liveDrive(t *testing.T, root string) *hi.Drive {
	t.Helper()

	alias := os.Getenv("HIDRIVE_LIVE")
	if alias == "" {
		t.Skip("set HIDRIVE_LIVE=<alias> to run against the real store")
	}

	conn := os.Getenv("DB_CONN")
	if conn == "" {
		conn = "postgres://localhost:5432/authdb"
	}
	pg, err := db.New(conn)
	if err != nil {
		t.Fatalf("connecting to the token store: %v", err)
	}
	t.Cleanup(pg.Close)

	return hi.NewDrive(NewAccessTokens(NewStore(pg.Pool())), hi.DriveConfig{Alias: alias, Root: root}, nil)
}

// What the root actually holds, so the rest of these tests address something
// real rather than a guess.
func TestLiveRootListing(t *testing.T) {
	drive := liveDrive(t, "/")

	meta, err := drive.Dir(context.Background(), os.Getenv("HIDRIVE_LIVE_PATH"))
	if err != nil {
		t.Fatalf("listing: %v", err)
	}

	t.Logf("%s — %d members", meta.Path, len(meta.Members))
	for i, m := range meta.Members {
		if i == 25 {
			t.Logf("  … and %d more", len(meta.Members)-i)
			break
		}
		t.Logf("  %-40s %-10s %10d  %s", m.NameURLEncoded, m.Category, m.Size_, m.Mimetype)
	}
}

// Does a thumbnail carry validators? If it does, passing them through lets a
// browser revalidate instead of refetching after the hour; if it does not,
// there is nothing to pass and the question is closed.
func TestLiveThumbnailHeaders(t *testing.T) {
	path := os.Getenv("HIDRIVE_LIVE_IMAGE")
	if path == "" {
		t.Skip("set HIDRIVE_LIVE_IMAGE=<path to an image> for this one")
	}
	drive := liveDrive(t, "/")

	resp, err := drive.Thumbnail(context.Background(), path, thumbSize(nil))
	if err != nil {
		t.Fatalf("thumbnail: %v", err)
	}
	defer resp.Body.Close()

	t.Logf("status %d", resp.StatusCode)
	for _, h := range []string{"Content-Type", "Content-Length", "ETag", "Last-Modified", "Cache-Control", "Content-Disposition"} {
		t.Logf("  %-20s %q", h, resp.Header.Get(h))
	}
	if resp.Header.Get("ETag") == "" && resp.Header.Get("Last-Modified") == "" {
		t.Log("  -> no validators: there is nothing to pass through, question closed")
	}
}

// Does a pre-signed URL send Content-Disposition: attachment? If it does, a
// video served through the proxying route downloads instead of playing — the
// one behaviour the old handler forced to "inline" and the new one does not.
func TestLiveSignedURLHeaders(t *testing.T) {
	path := os.Getenv("HIDRIVE_LIVE_FILE")
	if path == "" {
		t.Skip("set HIDRIVE_LIVE_FILE=<path to a file> for this one")
	}
	drive := liveDrive(t, "/")

	signed, err := drive.URL(context.Background(), path)
	if err != nil {
		t.Fatalf("signing: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, signed.String(), nil)
	req.Header.Set("Range", "bytes=0-1023")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("fetching the signed URL: %v", err)
	}
	defer resp.Body.Close()

	t.Logf("status %d (want 206 for a ranged request)", resp.StatusCode)
	for _, h := range []string{"Content-Type", "Content-Range", "Accept-Ranges", "ETag", "Last-Modified", "Content-Disposition"} {
		t.Logf("  %-20s %q", h, resp.Header.Get(h))
	}
	if d := resp.Header.Get("Content-Disposition"); strings.HasPrefix(d, "attachment") {
		t.Errorf("attachment: a video served this way downloads instead of playing")
	}
}

// The library end to end: a ranged request through the handler that will serve
// the mediathek, against the real store.
func TestLiveLibraryServesARange(t *testing.T) {
	path := os.Getenv("HIDRIVE_LIVE_FILE")
	if path == "" {
		t.Skip("set HIDRIVE_LIVE_FILE=<path to a file> for this one")
	}
	lib := NewLibrary(liveDrive(t, os.Getenv("HIDRIVE_LIVE_ROOT")), nil)

	r := httptest.NewRequest(http.MethodGet, "/api/v1/mediathek/stream/x", nil)
	r.SetPathValue("path", path)
	r.Header.Set("Range", "bytes=0-1023")

	w := httptest.NewRecorder()
	start := time.Now()
	if err := lib.Stream(w, r); err != nil {
		t.Fatalf("streaming: %v", err)
	}
	elapsed := time.Since(start)

	t.Logf("status %d in %s", w.Code, elapsed.Round(time.Millisecond))
	for _, h := range []string{"Content-Type", "Content-Range", "Content-Length", "Accept-Ranges", "ETag", "Last-Modified", "Cache-Control", "Vary"} {
		t.Logf("  %-20s %q", h, w.Header().Get(h))
	}

	if w.Code != http.StatusPartialContent {
		t.Errorf("status %d, want 206", w.Code)
	}
	if n := w.Body.Len(); n != 1024 {
		t.Errorf("got %d bytes, want 1024", n)
	}
	if got := w.Header().Get("Vary"); got != "" {
		t.Errorf("Vary = %q; the library is the same for everyone", got)
	}
}

// The two routes agreed in the unit tests; this is the same claim against the
// real store, where the slashed spelling used to 404 from one and list from the
// other.
func TestLiveBothSpellingsReachTheSameFile(t *testing.T) {
	file := os.Getenv("HIDRIVE_LIVE_FILE")
	if file == "" {
		t.Skip("set HIDRIVE_LIVE_FILE=<path to a file> for this one")
	}
	lib := NewLibrary(liveDrive(t, os.Getenv("HIDRIVE_LIVE_ROOT")), nil)

	for _, spelling := range []string{file, "/" + file} {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/mediathek/stream/x", nil)
		r.SetPathValue("path", spelling)
		r.Header.Set("Range", "bytes=0-15")

		w := httptest.NewRecorder()
		if err := lib.Stream(w, r); err != nil {
			t.Fatalf("%q: %v", spelling, err)
		}
		if w.Code != http.StatusPartialContent {
			t.Errorf("%q gave status %d, want 206", spelling, w.Code)
		}
		t.Logf("%-60q status %d, %s", spelling, w.Code, w.Header().Get("Content-Range"))
	}
}

// The trailing slash decides which upstream call goes first, and the fallback
// has to work both ways round: a directory asked for without its slash must
// still come back listed, not as a name with nothing in it.
func TestLiveTrailingSlashIsAHintNotAContract(t *testing.T) {
	drive := liveDrive(t, "/")
	dir := "/public/mediathek/2003-sea-monsters"

	slashed, err := lookupMeta(context.Background(), drive, dir+"/")
	if err != nil {
		t.Fatal(err)
	}
	bare, err := lookupMeta(context.Background(), drive, dir)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("with slash: %d members; without: %d", len(slashed.Members), len(bare.Members))
	if len(slashed.Members) == 0 {
		t.Fatal("the slashed form listed nothing")
	}
	if len(bare.Members) != len(slashed.Members) {
		t.Errorf("without the slash the listing was %d members, want %d — the fallback did not fire",
			len(bare.Members), len(slashed.Members))
	}

	// And a file, which is the case the hint is meant to make cheap.
	file, err := lookupMeta(context.Background(), drive, dir+"/cover.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if file.Type_ != "file" {
		t.Errorf("type = %q, want file", file.Type_)
	}
}

// The second read of the same object should not pay for metadata again — that
// cache is the whole reason this route is cheaper than the signed-URL one.
func TestLiveStatCacheSavesTheSecondRead(t *testing.T) {
	path := os.Getenv("HIDRIVE_LIVE_FILE")
	if path == "" {
		t.Skip("set HIDRIVE_LIVE_FILE=<path to a file> for this one")
	}
	lib := NewLibrary(liveDrive(t, os.Getenv("HIDRIVE_LIVE_ROOT")), nil)

	get := func() time.Duration {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/mediathek/stream/x", nil)
		r.SetPathValue("path", path)
		r.Header.Set("Range", "bytes=0-1023")
		start := time.Now()
		_ = lib.Stream(httptest.NewRecorder(), r)
		return time.Since(start)
	}

	first, second := get(), get()
	t.Logf("first %s, second %s", first.Round(time.Millisecond), second.Round(time.Millisecond))
	if second > first {
		t.Logf("  -> the second was not faster; one sample each, so this is a hint rather than a verdict")
	}
}
