package spa

import (
	"fmt"
	"net/http"
	"strings"
)

/////////////////////////////////////////////////////

func Serve(path string) http.Handler {
	fs := http.Dir(path)
	fileServer := http.FileServer(fs)
	serveIndex := serveFileContents("index.html", fs)

	return withCaching(intercept404(fileServer, serveIndex))
}

// withCaching splits the build into the two things it actually contains.
//
// Asset filenames carry a content hash, so a given URL never changes and may be
// cached indefinitely. The HTML that names them must not be: it keeps the same
// URL across deploys, and a browser holding an old copy loads the old assets and
// runs a version of the app that is no longer deployed.
//
// Without this the only header is Last-Modified, which leaves browsers free to
// guess a freshness lifetime — and they do.
//
// The value is written when the status is, not before: http.ServeContent strips
// Cache-Control when it errors, and an unknown path reaches the shell precisely
// by way of such an error. Setting it up front would work everywhere except the
// SPA's own routes, which is where it matters most.
func withCaching(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&cachingWriter{ResponseWriter: w, path: r.URL.Path}, r)
	})
}

type cachingWriter struct {
	http.ResponseWriter
	path    string
	written bool
}

func (w *cachingWriter) WriteHeader(status int) {
	w.setCacheControl()
	w.ResponseWriter.WriteHeader(status)
}

// Write covers a handler that never calls WriteHeader explicitly, where the
// status is implied by the first write.
func (w *cachingWriter) Write(p []byte) (int, error) {
	w.setCacheControl()
	return w.ResponseWriter.Write(p)
}

func (w *cachingWriter) setCacheControl() {
	if w.written {
		return
	}
	w.written = true

	// Asset filenames carry a content hash, so the URL never changes contents.
	// Everything else is the shell, which keeps its URL across deploys.
	if strings.HasPrefix(w.path, "/_nuxt/") || strings.HasPrefix(w.path, "/_fonts/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		return
	}
	// no-cache permits storing but requires revalidation, so an unchanged page
	// still costs only a 304.
	w.Header().Set("Cache-Control", "no-cache")
}

// https://hackandsla.sh/posts/2021-11-06-serve-spa-from-go/
type hookedResponseWriter struct {
	http.ResponseWriter
	got404 bool
}

func (hrw *hookedResponseWriter) WriteHeader(status int) {
	if status == http.StatusNotFound {
		// Don't actually write the 404 header, just set a flag.
		hrw.got404 = true
	} else {
		hrw.ResponseWriter.WriteHeader(status)
	}
}

func (hrw *hookedResponseWriter) Write(p []byte) (int, error) {
	if hrw.got404 {
		// No-op, but pretend that we wrote len(p) bytes to the writer.
		return len(p), nil
	}

	return hrw.ResponseWriter.Write(p)
}

func intercept404(handler, on404 http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hookedWriter := &hookedResponseWriter{ResponseWriter: w}
		handler.ServeHTTP(hookedWriter, r)

		if hookedWriter.got404 {
			on404.ServeHTTP(w, r)
		}
	})
}

func serveFileContents(file string, files http.FileSystem) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Restrict only to instances where the browser is looking for an HTML file
		if !strings.Contains(r.Header.Get("Accept"), "text/html") {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404 not found")

			return
		}

		// Open the file and return its contents using http.ServeContent
		index, err := files.Open(file)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "%s not found", file)

			return
		}

		fi, err := index.Stat()
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "%s not found", file)

			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, fi.Name(), fi.ModTime(), index)
	}
}
