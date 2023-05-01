package hi

import (
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/ihleven/errors"
)

func New(token string) *Drive {
	return &Drive{token: token}
}

// DRIVE implementiert fs.FS
// custom functionality on top of hidrive client like custom permissions
type Drive struct {
	token      string
	read_perms []string
	write_perm []string
}

func (hd *Drive) Open(name string) (fs.File, error) {

	c := NewHDClient(hd.token)
	meta, err := c.GetMeta(name)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t list dir %q", name)
	}

	// if meta.Filetype == "dir" {
	// 	meta.Members, err = drive.Listdir(r.URL.Path, token)
	// 	if err != nil {
	// 		return errors.Wrap(err, "Couldn‘t list dir %q", r.URL.Path)
	// 	}
	// }

	return &File{meta: meta, fsys: c}, nil

}

// ReadDir reads the named directory
// and returns a list of directory entries sorted by filename.
func (hd *Drive) ReadDir(name string) ([]fs.DirEntry, error) {

	c := NewHDClient(hd.token)
	meta, err := c.GetDir(name)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t list dir %q", name)
	}
	dirEntries := make([]fs.DirEntry, len(meta.Members))
	for i, e := range meta.Members {
		dirEntries[i] = &e
	}
	return dirEntries, nil
}

// type hifile struct {
// 	*meta
// 	token   string
// 	path    string
// 	entries []hifile
// 	name    string
// 	isDir   bool
// }

type File struct {
	meta *meta
	fsys *hdclient
	// reader *bytes.Reader
}

func (f *File) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		IsDir   bool   `json:"isDir"`
		Entries int    `json:"entries"`
	}{
		ID:      f.meta.Path,
		Name:    f.meta.NameURLEncoded,
		IsDir:   f.meta.IsDir(),
		Entries: len(f.meta.Members),
	})
}

func (f *File) Stat() (fs.FileInfo, error) {
	var info fs.FileInfo = f.meta
	return info, nil
}

func (f *File) Read([]byte) (int, error) {
	return 0, nil
}
func (f *File) Close() error {
	return nil
}

func (f *File) ReadDir(n int) ([]fs.DirEntry, error) {

	dirmeta, err := f.fsys.GetDir(f.meta.Path)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t list dir %q", f.meta.Path)
	}

	// https://stackoverflow.com/a/12994852
	var entries = make([]fs.DirEntry, len(dirmeta.Members))
	for i := range dirmeta.Members {
		entries[i] = &dirmeta.Members[i]
		fmt.Println("dirmeta", i, entries[i])
	}
	return entries, nil

	///////
	// f.entries = make([]hifile, len(dirmeta.Members))
	// for i, member := range dirmeta.Members {
	// 	f.entries[i] = hifile{name: member.Name(), isDir: member.IsDir()}
	// }
	// if f == nil {
	// 	return nil, os.ErrInvalid
	// }

}
