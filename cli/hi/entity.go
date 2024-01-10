package hi

import (
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"time"

	"github.com/ihleven/errors"
)

// Meta implements fs.File interface:
// Stat() (FileInfo, error)
// Read([]byte) (int, error)
// Close() error
type Meta struct {
	accessToken    string
	ID             string `json:"id"`
	NameURLEncoded string `json:"name"`
	Path           string `json:"path"`
	Type_          string `json:"type"`
	Size_          int    `json:"size"`
	Category       string `json:"category"`
	NMembers       int    `json:"nmembers"`
	MTime          int64  `json:"mtime,omitempty"`
	Members        []Meta `json:"members"`
	Mimetype       string `json:"mime_type"`

	CTime    int    `json:"ctime"`
	Readable bool   `json:"readable"`
	Writable bool   `json:"writable"`
	ParentID string `json:"parent_id"`

	Image *struct {
		Width  int `json:"width"`
		Height int `json:"height"`

		Exif *struct {
			DateTimeOriginal string  `json:",omitempty"`
			Make             string  `json:",omitempty"`
			Model            string  `json:",omitempty"`
			ImageWidth       int     `json:",omitempty"`
			ImageHeight      int     `json:",omitempty"`
			ExifImageWidth   int     `json:",omitempty"`
			ExifImageHeight  int     `json:",omitempty"`
			Aperture         float64 `json:",omitempty"`
			ExposureTime     float64 `json:",omitempty"`
			ISO              int     `json:",omitempty"`
			FocalLength      float64 `json:",omitempty"`
			Orientation      int     `json:",omitempty"`
			XResolution      float64 `json:",omitempty"`
			YResolution      float64 `json:",omitempty"`
			ResolutionUnit   int     `json:",omitempty"`
			BitsPerSample    int     `json:",omitempty"`
			GPSLatitude      float64 `json:",omitempty"`
			GPSLongitude     float64 `json:",omitempty"`
			GPSAltitude      float64 `json:",omitempty"`
		} `json:",omitempty"`
	} `json:"image,omitempty"`

	// chash,
	// ctime,
	// has_dirs,
	// mhash,mohash,nhash,nmembers,parent_id,readable,rshare,writable

	// members,
	// members.category,members.chash,members.ctime,members.has_dirs,members.id,
	// members.image.exif,members.image.height,members.image.width,
	// members.mhash,members.mime_type,members.mohash,members.mtime,members.name,members.nmembers,members.nhash,members.parent_id,members.path,members.readable,members.rshare,members.size,members.type,members.writable,
	// *hdclient
	data      []byte
	readIndex int64
}

func (m *Meta) Stat() (fs.FileInfo, error) {
	// var info fs.FileInfo = f.meta
	return m, nil
}

func (m *Meta) Close() error {
	m = nil
	return nil
}

func (m *Meta) Seek(offset int64, whence int) (int64, error) {
	newPos := m.readIndex
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos += offset
	case io.SeekEnd:
		newPos = int64(m.Size_ + int(offset))
	}
	if newPos < 0 {
		return 0, errors.New("negative result pos")
	}
	m.readIndex = newPos
	return newPos, nil

}

func (m *Meta) Read(p []byte) (n int, err error) {
	// buf := make([]byte, len(p))
	// fmt.Println("read", len(p), len(buf))
	body, er := NewHDClient(m.accessToken).GetFile(m.path(), int(m.readIndex), int(m.readIndex)+len(p)-1)
	if er != nil {
		return 0, io.EOF
	}
	defer body.Close()
	buf, err := io.ReadAll(body)
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(p) != len(buf) {
		fmt.Println("len(p) != len(buf)", len(p), len(buf))
		err = io.EOF
	}
	n = copy(p, buf)
	m.readIndex += int64(n)
	return
}

// func (m *meta) Read(p []byte) (n int, err error) {

// 	if m.data == nil {
// 		body, er := NewHDClient(m.accessToken).GetFile(m.path(), 0, 0)
// 		if er != nil {
// 			return 0, io.EOF
// 		}
// 		defer body.Close()

// 		m.data, err = io.ReadAll(body)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 	}

// 	if m.readIndex >= int64(len(m.data)) {
// 		err = io.EOF
// 		return
// 	}

// 	n = copy(p, m.data[m.readIndex:])
// 	m.readIndex += int64(n)
// 	fmt.Println(n, err)
// 	return
// }

func (m *Meta) ReadDir(n int) ([]fs.DirEntry, error) {
	fmt.Println("path:", m.Path)
	meta, err := NewHDClient(m.accessToken).GetDir(m.path())
	if err != nil {
		return nil, errors.Wrap(err, "Couldn‘t list dir %q", m.Path)
	}
	m.Members = meta.Members

	dirEntries := make([]fs.DirEntry, len(meta.Members))
	for i := range meta.Members {
		meta.Members[i].accessToken = m.accessToken
		dirEntries[i] = &meta.Members[i]
	}
	return dirEntries, nil
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
func (e *Meta) Type() fs.FileMode {
	// TODO!!!
	var mode uint32
	return fs.FileMode(mode)
}

// Info is part of fs.DirEntry interface
func (e *Meta) Info() (fs.FileInfo, error) {
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
func (m *Meta) Name() string { // base name of the file
	unescapedName, _ := url.QueryUnescape(m.NameURLEncoded)
	return unescapedName
}

// kein Interface, nur convenience
func (m *Meta) path() string {
	p, _ := url.QueryUnescape(m.Path)
	return p
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
