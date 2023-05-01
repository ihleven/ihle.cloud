package main

import (
	"io/fs"
	"net/http"
	"os"

	"github.com/ihleven/errors"
	"github.com/ihleven/pkg/hidrive"
)

type hifile struct {
	*hidrive.Meta
	path    string
	entries []hifile
	name    string
	isDir   bool
}

func (f *hifile) Stat() (fs.FileInfo, error) {
	var info fs.FileInfo = f.Meta
	return info, nil
}

func (f *hifile) Read([]byte) (int, error) {
	return 0, nil
}
func (f *hifile) Close() error {
	return nil
}

func (f *hifile) ReadDir(n int) ([]fs.DirEntry, error) {

	token := drive.Token("wolfgang")
	if token == nil {
		return nil, errors.NewWithCode(http.StatusProxyAuthRequired, "Couldn‘t get valid auth token")
	}
	entries, err := drive.Listdir(f.path, token)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t list dir %q", f.path)
	}

	f.entries = make([]hifile, len(entries))
	for i, e := range entries {
		f.entries[i] = hifile{name: e.Name(), isDir: e.IsDir()}
	}
	if f == nil {
		return nil, os.ErrInvalid
	}

	ret := make([]fs.DirEntry, len(entries))
	for i, e := range entries {
		ret[i] = &e
	}
	return ret, nil
}

// //////////////////////////////////////////////////////////////////////////
type hidriveFS hidrive.Drive

func (fs *hidriveFS) Open(name string) (fs.File, error) {

	token := drive.Token("wolfgang")
	if token == nil {
		return nil, errors.NewWithCode(http.StatusProxyAuthRequired, "Couldn‘t get valid auth token")
	}

	meta, err := drive.GetMeta(name, token)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t list dir %q", name)
	}

	// if meta.Filetype == "dir" {
	// 	meta.Members, err = drive.Listdir(r.URL.Path, token)
	// 	if err != nil {
	// 		return errors.Wrap(err, "Couldn‘t list dir %q", r.URL.Path)
	// 	}
	// }

	return &hifile{Meta: meta, path: name}, nil

}
