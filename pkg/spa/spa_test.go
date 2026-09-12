package spa

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// The SPA is two kinds of file with opposite caching needs, and getting it wrong
// is invisible until someone is silently running a version that is no longer
// deployed.
func TestCacheControl(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "index.html"), "<html>shell</html>")
	if err := os.MkdirAll(filepath.Join(dir, "_nuxt"), 0o755); err != nil {
		t.Fatalf("creating _nuxt: %v", err)
	}
	write(t, filepath.Join(dir, "_nuxt", "abc123.js"), "console.log(1)")

	handler := Serve(dir)

	tests := []struct {
		what, path, want string
	}{
		// Hashed filenames never change contents, so they may be kept forever.
		{"a hashed asset", "/_nuxt/abc123.js", "public, max-age=31536000, immutable"},
		// The shell keeps its URL across deploys and names the assets, so it has
		// to be revalidated or an old copy pins an old build.
		{"the shell", "/index.html", "no-cache"},
		{"an unknown route falling through to the shell", "/passkeys", "no-cache"},
	}
	for _, tt := range tests {
		r := httptest.NewRequest(http.MethodGet, tt.path, nil)
		r.Header.Set("Accept", "text/html")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if got := w.Header().Get("Cache-Control"); got != tt.want {
			t.Errorf("%s: Cache-Control = %q, want %q", tt.what, got, tt.want)
		}
	}
}

// An unknown path is the SPA's own routing, so it gets the shell rather than a
// 404 — otherwise a deep link into the app would not load.
func TestUnknownPathServesTheShell(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "index.html"), "<html>shell</html>")

	r := httptest.NewRequest(http.MethodGet, "/passkeys", nil)
	r.Header.Set("Accept", "text/html")
	w := httptest.NewRecorder()
	Serve(dir).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", w.Code)
	}
	if body := w.Body.String(); body != "<html>shell</html>" {
		t.Errorf("body = %q, want the shell", body)
	}
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

// A manifest served as text/plain is the difference between an installed app
// with a name and icon and one without; Go's own table does not cover it.
func TestWebmanifestContentType(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "manifest.webmanifest"), `{"name":"x"}`)

	rec := httptest.NewRecorder()
	Serve(dir).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/manifest.webmanifest", nil))

	if got := rec.Header().Get("Content-Type"); got != "application/manifest+json" {
		t.Errorf("Content-Type = %q, want application/manifest+json", got)
	}
}
