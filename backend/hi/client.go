package hi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type Config struct {
	ClientID     string `arg:"env:CLIENT_ID"      help:"Hidrive client ID"`
	ClientSecret string `arg:"env:CLIENT_SECRET"  help:"Hidrive client secret"`
}

var client = http.Client{
	Timeout: 100 * time.Second,
}

func NewClient(token, prefix string) *hdclient {

	return &hdclient{
		token:  token,
		prefix: path.Clean(prefix),
	}
}

// hdclient
type hdclient struct {
	prefix string
	token  string
}

type Meta struct {
	ID             string `json:"id,omitempty"`
	NameURLEncoded string `json:"name"`
	Path           string `json:"path,omitempty"`
	Type_          string `json:"type,omitempty"`
	Size_          int    `json:"size,omitempty"`
	Category       string `json:"category,omitempty"`
	NMembers       int    `json:"nmembers,omitempty"`
	MTime          int64  `json:"mtime,omitempty"`
	Members        []Meta `json:"members,omitempty"`
	Mimetype       string `json:"mime_type,omitempty"`

	CTime    int    `json:"ctime,omitempty"`
	Readable bool   `json:"readable,omitempty"`
	Writable bool   `json:"writable,omitempty"`
	ParentID string `json:"parent_id,omitempty"`

	Image *Image `json:"image,omitempty"`
}
type Image struct {
	Width  int   `json:"width"`
	Height int   `json:"height"`
	Exif   *Exif `json:"exif"`
}

type Exif struct {
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
}

func (hd *hdclient) resolvepath(p string) string {
	if hd.prefix != "" {
		p = path.Join(hd.prefix, p)
	}
	return path.Clean(p)
}

// GetMeta liefert die Dateien für Pfad p relativ zum client prefix
func (hd *hdclient) GetMeta(p string) (*Meta, error) {
	resolved := hd.resolvepath(p)
	fmt.Println(" GetMeta * resolved path:", p, resolved)

	fields := "id,name,path,category,nmembers,ctime,has_dirs,mtime,readable,size,type,writable,mime_type,members,members.id,members.name,image.exif,image.height,image.width"
	params := url.Values{
		"path":   []string{p},
		"fields": []string{fields},
	}
	resp, err := rq(GET, "/meta", params, bearer(hd.token))
	fmt.Println(resp)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	defer io.Copy(io.Discard, resp.Body)

	var meta Meta
	err = json.NewDecoder(resp.Body).Decode(&meta)
	if err != nil {
		return nil, Error{Message: err.Error()}
	}
	meta.Path = strings.TrimPrefix(meta.Path, hd.prefix)
	return &meta, nil
}

func (hd *hdclient) GetDir(p string) (*Meta, error) {
	p = hd.resolvepath(p)

	fmt.Println(" GetDir * resolved path:", p)

	params := url.Values{
		"path":   []string{p},
		"fields": []string{"id,name,path,category,nmembers,chash,ctime,mtime,has_dirs,readable,rshare,size,type,writable,members,members.category,members.chash,members.ctime,members.has_dirs,members.image.exif,members.image.height,members.image.width,members.mime_type,members.name,members.nmembers,members.size,members.type,mhash,mohash,nhash,parent_id"},
	}
	resp, err := rq(GET, "/dir", params, bearer(hd.token))
	fmt.Println(resp, params)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	defer io.Copy(io.Discard, resp.Body)

	var meta Meta
	err = json.NewDecoder(resp.Body).Decode(&meta)
	if err != nil {
		return nil, Error{Message: err.Error()}
	}
	meta.Path = strings.TrimPrefix(meta.Path, hd.prefix)
	return &meta, nil
}

func (hd *hdclient) GetURL(p string) (*url.URL, error) {

	respath := hd.resolvepath(p)

	fmt.Println("GETURL", respath)

	resp, err := rq(GET, "/file/url", url.Values{"path": []string{respath}}, bearer(hd.token))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response struct {
		URL string `json:"url"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return url.Parse(response.URL)
}

func (hd *hdclient) GetFile(p string, rangeFrom, rangeTo int) (io.ReadCloser, error) {
	respath := hd.resolvepath(p)
	resp, err := rq(GET, "/file", url.Values{"path": []string{respath}},
		bearer(hd.token),
		rangeHeader(rangeFrom, rangeTo),
	)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (hd *hdclient) File(p string, query url.Values, h http.Header) (*http.Response, error) {
	query.Set("path", "/"+hd.resolvepath(p))

	resp, err := rq(GET, "/file", query,
		bearer(hd.token),
		headers(h),
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
