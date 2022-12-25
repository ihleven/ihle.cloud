package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	flag "github.com/spf13/pflag"

	"bitbucket.org/hotelplan/webcc-pkg/web"
	"github.com/ihleven/errors"
	"github.com/ihleven/pkg/hidrive"
)

var (
	VERSION       string
	BUILDTIME     string
	CLIENT_ID     string
	CLIENT_SECRET string
)

// func nuxtHandler() http.Handler {

// 	subFS, err := fs.Sub(Filesystem, "public")
// 	httpFS := http.FS(subFS)

// 	fmt.Println("aubFS", err)
// 	fileServer := http.FileServer(httpFS)
// 	// return http.StripPrefix("public/", fileServer)
// 	return fileServer
// 	// serveIndex := serveFileContents("index.html", httpFS)

// 	// return intercept404(fileServer, serveIndex)
// }

var nuxt *string = flag.String("nuxt", "./public", "path to nuxt .output/public")

func main() {

	fmt.Println("CLIENT_ID", CLIENT_ID)
	fmt.Println("CLIENT_SECRET", CLIENT_SECRET)
	setup()
	// fmt.Println("drive", drive)
	// token := drive.Token("wolfgang")
	// fmt.Println("token", token)
	// meta, err := drive.Meta("/", token)
	// fmt.Println("meta", meta, err)

	// usecase = kunst.NewUsecase(repo, drive)
	// return nil
}

var drive *hidrive.Drive

func setup() {

	fmt.Println(CLIENT_ID)
	fmt.Println(CLIENT_SECRET)
	manager := hidrive.NewAuthManager(CLIENT_ID, CLIENT_SECRET)
	t, e := manager.GetAccessToken("wolfgang")
	drive = hidrive.NewDrive(manager)
	fmt.Println("token:", drive.Token("wolfgang"), e, t)
	srv := web.NewServer(false, web.Addr("", 10815))

	srv.Register("/", asdf(*nuxt))
	srv.Register("/hidrive", handler)
	srv.Register("/serve", serve)
	srv.Register("/wolfgang-ihle", serveWolfgangIhle())
	srv.Register("/media/videos", servePrefix("videos"))

	srv.Run()
}

func handler(rw *web.ResponseWriter, r *http.Request) error {

	token := drive.Token("wolfgang")
	if token == nil {
		return errors.NewWithCode(http.StatusProxyAuthRequired, "Couldn‘t get valid auth token")
	}

	meta, err := drive.GetMeta(r.URL.Path, token)
	if err != nil {
		return errors.Wrap(err, "Couldn‘t list dir %q", r.URL.Path)
	}

	if meta.Filetype == "dir" {
		meta.Members, err = drive.Listdir(r.URL.Path, token)
		if err != nil {
			return errors.Wrap(err, "Couldn‘t list dir %q", r.URL.Path)
		}
	}

	return rw.RespondJSON(meta)
}

func serve(rw http.ResponseWriter, r *http.Request) {

	ctx := context.WithValue(context.Background(), "username", "wolfgang")
	r2 := r.WithContext(ctx)
	fmt.Println("path:", r.URL.Path)
	token := drive.Token("wolfgang")
	if token == nil {
		// return errors.NewWithCode(http.StatusProxyAuthRequired, "Couldn‘t get valid auth token")
	}

	drive.Serve(rw, r2)
}

func serveWolfgangIhle() web.HandlerFunc {

	manager := hidrive.NewAuthManager(CLIENT_ID, CLIENT_SECRET)
	drive := hidrive.NewDrive(manager, hidrive.Prefix("/wolfgang-ihle"))

	return func(rw *web.ResponseWriter, r *http.Request) error {

		ctx := context.WithValue(context.Background(), "username", "wolfgang")

		drive.Serve(rw, r.WithContext(ctx))

		return nil
	}
}

func servePrefix(prefix string) web.HandlerFunc {

	// manager := hidrive.NewAuthManager(CLIENT_ID, CLIENT_SECRET)
	// drive := hidrive.NewDrive(manager, hidrive.Prefix(prefix))

	return func(rw *web.ResponseWriter, r *http.Request) error {

		fd, err := os.Open(prefix + r.URL.Path)
		if err != nil {
			return err
		}
		stat, err := fd.Stat()
		if err != nil {
			return err
		}

		rw.Header().Set("Access-Control-Allow-Origin", "*")
		rw.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")

		http.ServeContent(rw, r, fd.Name(), stat.ModTime(), fd)

		return nil
	}
}

func asdf(path string) http.Handler {
	fs := http.Dir(path)
	fileServer := http.FileServer(fs)
	serveIndex := serveFileContents("index.html", fs)

	return intercept404(fileServer, serveIndex)
}

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
