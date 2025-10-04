package hi

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"time"
)

// https://www.refurbed.org/posts/golang-filesystem-interfaces/

// type StatFS interface {
// 	FS
// 	Stat(name string) (os.FileInfo, error)
// }

func NewFS(accesstoken string) hifsys {
	return hifsys{client: hdclient{"/", accesstoken}}
}

// DRIVE implementiert fs.FS
// custom functionality on top of hidrive client like custom permissions
type hifsys struct {
	client hdclient
	// token string
	// read_perms []string
	// write_perm []string
}

// type FS interface {
// 	Open(name string) (File, error)
// }

// Open implements fs.FS interface
func (fs *hifsys) Open(name string) (fs.File, error) {
	meta, err := fs.client.GetMeta("/" + name)
	if err != nil {
		return nil, err
	}

	file := File{meta: meta, fsys: *fs}
	// meta.accessToken = fs.client.token
	fmt.Println("OPen", name, file)

	return &file, nil
}
func (fs *hifsys) Stat(name string) (fs.FileInfo, error) {
	return fs.client.GetMeta(name)
}

func (fsys *hifsys) ReadDir(name string) ([]fs.DirEntry, error) {
	meta, err := fsys.client.GetDir(name)
	if err != nil {
		return nil, err
	}
	var direntries []fs.DirEntry
	for _, m := range meta.Members {
		direntries = append(direntries, &m)
	}
	return direntries, nil
}

// ReadFile reads the named file and returns its contents.
// A successful call returns a nil error, not io.EOF.
// (Because ReadFile reads the whole file, the expected EOF
// from the final Read is not treated as an error to be reported.)
//
// The caller is permitted to modify the returned byte slice.
// This method should return a copy of the underlying data.
func (fs *hifsys) ReadFile(name string) ([]byte, error) {
	return nil, nil
}

//////////////////////////////////////////////////
//////////      fs.File interface       //////////
//////////////////////////////////////////////////

type File struct {
	fsys hifsys
	meta *Meta
	// fsys *hdclient
	// reader *bytes.Reader
	// data      []byte
	readIndex int64
}

func (f *File) Stat() (fs.FileInfo, error) {
	// var info fs.FileInfo = f.meta
	fmt.Println("f.Stat()", f.meta)
	return f.meta, nil
}

func (f *File) Read(p []byte) (int, error) {

	fmt.Println("read:", f.meta.Path, len(p))
	if f.readIndex >= int64(f.meta.Size_) {
		return 0, io.EOF
	}

	// io.Reader(p)

	path, _ := url.QueryUnescape(f.meta.Path)

	body, er := f.fsys.client.GetFile(path, int(f.readIndex), int(f.readIndex)+len(p)-1)
	if er != nil {
		return 0, io.EOF
	}
	defer body.Close()

	n, err := body.Read(p)
	// n := copy(p, body)
	f.readIndex += int64(n)
	// return

	// if len(p) != len(buf) {
	// 	fmt.Println("len(p) != len(buf)", len(p), len(buf))
	// 	err = io.EOF
	// }
	// n := copy(p, buf)
	// f.readIndex += int64(n)
	return n, err
}

func (f *File) Close() error {
	fmt.Println("close:", f.meta.Path)
	f.meta = nil
	f.readIndex = 0
	return nil
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	fmt.Println("seek", f.meta.Path, offset, whence)
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = f.readIndex + offset
	case io.SeekEnd:

		abs = int64(f.meta.Size_) + offset
	default:
		return 0, errors.New("Seek: invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("Seek: negative position")
	}
	if abs > int64(f.meta.Size_)-1 {
		return 0, errors.New("Seek: index out of bounds position")
	}
	f.readIndex = abs
	return abs, nil
}

func (f *File) ReadDir(n int) ([]fs.DirEntry, error) {

	return nil, nil
}

//////////////////////////////////////////////////
//////////     fs.FileInfo interface    //////////
//////////////////////////////////////////////////

type FileInfo interface {
	Name() string       // base name of the file
	Size() int64        // length in bytes for regular files; system-dependent for others
	Mode() fs.FileMode  // file mode bits
	ModTime() time.Time // modification time
	IsDir() bool        // abbreviation for Mode().IsDir()
	Sys() any           // underlying data source (can return nil)
}

func (m *Meta) Name() string {
	unescapedName, err := url.QueryUnescape(m.NameURLEncoded)
	if err != nil {
		return m.NameURLEncoded
	}
	return unescapedName
}

// Size is part of fs.FileInfo interface
func (m *Meta) Size() int64 { // length in bytes for regular files; system-dependent for others
	return int64(m.Size_)
}

// Mode is part of fs.FileInfo interface
func (m *Meta) Mode() fs.FileMode { // file mode bits
	var mode uint32
	return fs.FileMode(mode)
}

// ModTime is part of fs.FileInfo interface
func (m *Meta) ModTime() time.Time {
	return time.Unix(0, m.MTime)
}

// IsDir is part of fs.FileInfo and fs.DirEntry interface
func (m *Meta) IsDir() bool { // abbreviation for Mode().IsDir()
	return m.Type_ == "dir"
}

// Sys is part of fs.FileInfo interface
func (m *Meta) Sys() interface{} {
	return m
}

type DirEntry interface {
	// Name returns the name of the file (or subdirectory) described by the entry.
	// This name is only the final element of the path (the base name), not the entire path.
	// For example, Name would return "hello.go" not "home/gopher/hello.go".
	Name() string

	// IsDir reports whether the entry describes a directory.
	IsDir() bool

	// Type returns the type bits for the entry.
	// The type bits are a subset of the usual FileMode bits, those returned by the FileMode.Type method.
	Type() fs.FileMode

	// Info returns the FileInfo for the file or subdirectory described by the entry.
	// The returned FileInfo may be from the time of the original directory read
	// or from the time of the call to Info. If the file has been removed or renamed
	// since the directory read, Info may return an error satisfying errors.Is(err, ErrNotExist).
	// If the entry denotes a symbolic link, Info reports the information about the link itself,
	// not the link's target.
	Info() (FileInfo, error)
}

// Mode is part of fs.FileInfo interface
func (m *Meta) Type() fs.FileMode { // file mode bits
	return m.Mode()
}

// Info is part of fs.DirEntry interface
func (m *Meta) Info() (fs.FileInfo, error) {
	return m, nil
}
