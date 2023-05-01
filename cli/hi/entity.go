package hi

import (
	"io/fs"
	"net/url"
	"time"
)

type meta struct {
	ID             string `json:"id"`
	NameURLEncoded string `json:"name"`
	Path           string `json:"path"`
	Type_          string `json:"type"`
	Size_          int    `json:"size"`
	Category       string `json:"category"`
	NMembers       int    `json:"nmembers"`
	MTime          int64  `json:"mtime,omitempty"`
	Members        []meta `json:"members"`

	// chash,
	// ctime,
	// has_dirs,
	// mhash,mohash,nhash,nmembers,parent_id,readable,rshare,writable

	// members,
	// members.category,members.chash,members.ctime,members.has_dirs,members.id,
	// members.image.exif,members.image.height,members.image.width,
	// members.mhash,members.mime_type,members.mohash,members.mtime,members.name,members.nmembers,members.nhash,members.parent_id,members.path,members.readable,members.rshare,members.size,members.type,members.writable,
	// *hdclient
}

type entry struct {
	ID             string `json:"id"`
	NameURLEncoded string `json:"name"`
	Type_          string `json:"type"`
}

// type DirEntry interface {
// 	// Name returns the name of the file (or subdirectory) described by the entry.
// 	// This name is only the final element of the path (the base name), not the entire path.
// 	// For example, Name would return "hello.go" not "home/gopher/hello.go".
// 	Name() string

// 	// IsDir reports whether the entry describes a directory.
// 	IsDir() bool

// 	// Type returns the type bits for the entry.
// 	// The type bits are a subset of the usual FileMode bits, those returned by the FileMode.Type method.
// 	Type() FileMode

// 	// Info returns the FileInfo for the file or subdirectory described by the entry.
// 	// The returned FileInfo may be from the time of the original directory read
// 	// or from the time of the call to Info. If the file has been removed or renamed
// 	// since the directory read, Info may return an error satisfying errors.Is(err, ErrNotExist).
// 	// If the entry denotes a symbolic link, Info reports the information about the link itself,
// 	// not the link's target.
// 	Info() (FileInfo, error)
// }

var t fs.DirEntry

//    Name() string -> see above
//    IsDir() bool -> see above
//    Type() FileMode
//    Info() (FileInfo, error)

// Type is part of fs.DirEntry interface
func (e *meta) Type() fs.FileMode {
	// TODO!!!
	var mode uint32
	return fs.FileMode(mode)
}

// Info is part of fs.DirEntry interface
func (e *meta) Info() (fs.FileInfo, error) {
	return e, nil
}

// type FileInfo interface {
//     Name() string       // base name of the file
//     Size() int64        // length in bytes for regular files; system-dependent for others
//     Mode() FileMode     // file mode bits
//     ModTime() time.Time // modification time
//     IsDir() bool        // abbreviation for Mode().IsDir()
//     Sys() any           // underlying data source (can return nil)
// }

// Name is part of fs.FileInfo and fs.DirEntry interface
func (m *meta) Name() string { // base name of the file
	unescapedName, _ := url.QueryUnescape(m.NameURLEncoded)
	return unescapedName
}

// Size is part of fs.FileInfo interface
func (m *meta) Size() int64 { // length in bytes for regular files; system-dependent for others
	return int64(m.Size_)
}

// Mode is part of fs.FileInfo interface
func (m *meta) Mode() fs.FileMode { // file mode bits
	var mode uint32
	return fs.FileMode(mode)
}

// ModTime is part of fs.FileInfo interface
func (m *meta) ModTime() time.Time {
	return time.Unix(0, m.MTime)
}

// IsDir is part of fs.FileInfo and fs.DirEntry interface
func (m *meta) IsDir() bool { // abbreviation for Mode().IsDir()
	return m.Type_ == "dir"
}

// Sys is part of fs.FileInfo interface
func (m *meta) Sys() interface{} {
	return m
}
