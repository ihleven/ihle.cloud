package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"

	"bitbucket.org/hotelplan/webcc-pkg/web"
)

func FileServer(fsys fs.FS) web.HandlerFunc {

	return func(rw *web.ResponseWriter, r *http.Request) error {

		file, err := fsys.Open(r.URL.Path)
		if err != nil {
			return err
		}

		stat, err := file.Stat()
		if err != nil {
			return err
		}
		fmt.Printf("%T\n", stat.Sys())
		if stat.IsDir() {
			readDirFile, _ := file.(fs.ReadDirFile)

			entries, err := readDirFile.ReadDir(-1)
			if err != nil {
				return err
			}
			sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

			switch r.Header.Get("Accept") {

			case "application/json":
				// m := stat.Sys().(*hi.Meta)
				// f := struct {
				// 	*hi.Meta
				// 	Entries []fs.DirEntry `json:"members"`
				// }{Meta: m, Entries: entries}

				rw.RespondJSON(readDirFile)

			default:
				// dirList(rw, stat.Name(), entries)
				rw.Header().Set("Content-Type", "text/html; charset=utf-8")
				fmt.Fprintf(rw, "<pre>\n")
				for _, e := range entries {
					name := e.Name()
					if e.IsDir() {
						name += "/"
					}
					url := url.URL{Path: path.Join(rw.Route, r.URL.Path, name)}
					fmt.Fprintf(rw, "<a href=\"%s\">%s</a>\n", url.String(), name)
				}
				fmt.Fprintf(rw, "</pre>\n")
			}

			return nil
		}

		switch r.Header.Get("Accept") {
		case "application/json":
			// bytes, err := json.MarshalIndent(stat.Sys(), "", "    ")
			// if err != nil {
			// 	return err
			// }
			rw.RespondJSON(stat.Sys())
		default:

			if seeker, ok := file.(io.ReadSeeker); ok {
				fmt.Println("using servecontent-> ", stat.Name(), stat.ModTime())
				http.ServeContent(rw, r, stat.Name(), stat.ModTime(), seeker)
			} else {
				written, err := io.Copy(rw, file)
				fmt.Println("io.Copy -> ", written, err)
			}
		}
		return nil
	}
}

type file struct {
	*os.File
	name    string
	isDir   bool
	meta    fs.FileInfo
	entries []file
	path    string
}

func (f *file) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		IsDir   bool   `json:"isDir"`
		Entries []file `json:"entries,omitempty"`
	}{
		ID:      f.path,
		Name:    f.name,
		IsDir:   f.isDir,
		Entries: f.entries,
	})
}

func (f *file) ReadDir(n int) ([]fs.DirEntry, error) {
	entries, _ := f.File.ReadDir(n)
	f.entries = make([]file, len(entries))
	for i, e := range entries {
		f.entries[i] = file{name: e.Name(), isDir: e.IsDir()}
	}
	if f == nil {
		return nil, os.ErrInvalid
	}
	return entries, nil
}

// Stat returns the FileInfo structure describing file.
// If there is an error, it will be of type *PathError.
func (f *file) Stat() (fs.FileInfo, error) {
	f.meta, _ = f.File.Stat()
	if f == nil {
		return nil, os.ErrInvalid
	}
	f.name = f.meta.Name()
	f.isDir = f.meta.IsDir()

	return f.meta, nil
}

// //////////////////////////////////////////////////////////////////////////
type dirFS string

func (dir dirFS) Open(name string) (fs.File, error) {
	fullname := filepath.Join(string(dir), name)

	f, err := os.Open(fullname)
	if err != nil {
		return nil, err // nil fs.File
	}

	return &file{File: f, path: name}, nil
	// return f, nil
}

// Stat makes dirFS implement StatFS
func (dir dirFS) Stat(name string) (fs.FileInfo, error) {
	fullname := filepath.Join(string(dir), name)
	f, err := os.Stat(fullname)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// ReadDir makes dirFS implement ReadDirFS
// func (dir dirFS) ReadDir(name string) ([]fs.DirEntry, error) {
// 	fullname := filepath.Join(string(dir), name)
// 	f, err := os.ReadDir(fullname)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return f, nil
// }
