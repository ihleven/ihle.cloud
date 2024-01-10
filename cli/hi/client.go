package hi

import (
	"io"
	"net/http"
	"time"

	"github.com/ihleven/errors"
)

var client = http.Client{
	Timeout: 100 * time.Second,
}

func NewHDClient(token string) *hdclient {

	return &hdclient{
		token: token,
	}
}

type hdclient struct {
	prefix string
	token  string
}

func (hd *hdclient) GetDir(p string) (*Meta, error) {

	req := newRequest(GET, "/dir", path(hd.prefix+p),
		bearer(hd.token),
		fields(metafields+dirfields),
		// name,path,ctime,has_dirs,mtime,readable,size,type,writable,
	)
	// response, err := req.Exec(&client)
	// if err != nil {
	// 	return nil, err
	// }
	// defer response.Body.Close()

	// var meta Meta
	// err = json.NewDecoder(response.Body).Decode(&meta)
	// if err != nil {
	// 	return nil, &Error{Message: err.Error()}
	// }
	// return &meta, nil

	meta, err := req.fetchJSON(&client)
	if err != nil {
		return nil, errors.Wrap(err, "error in GetDir")
	}

	// meta.hdclient = hd
	return meta, nil
}

var metafields = "category,ctime,id,members,mtime,name,nmembers,parent_id,path,readable,size,type,writable,mime_type,image.height,image.width,"
var exiffields = "image.exif.DateTimeOriginal,image.exif.Make,image.exif.Model,image.exif.ImageWidth,image.exif.ImageHeight,image.exif.ExifImageWidth,image.exif.ExifImageHeight,image.exif.Aperture,image.exif.ExposureTime,image.exif.ISO,image.exif.FocalLength,image.exif.Orientation,image.exif.XResolution,image.exif.YResolution,image.exif.ResolutionUnit,image.exif.BitsPerSample,image.exif.GPSLatitude,image.exif.GPSLongitude,image.exif.GPSAltitude"
var dirfields = "members.path,members.category,members.ctime,members.id,members.image.height,members.image.width,members.mime_type,members.mtime,members.name,members.nmembers,members.readable,members.size,members.type,members.writable"

func (hd *hdclient) GetMeta(p string) (*Meta, error) {

	req := newRequest(GET, "/meta",
		path(p),
		bearer(hd.token),
		fields(metafields+dirfields+","+exiffields),
	)

	meta, err := req.fetchJSON(&client)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	// meta.hdclient = hd
	return meta, nil
}

func (c *hdclient) GetFile(p string, rangeFrom, rangeTo int) (io.ReadCloser, error) {

	req := newRequest(GET, "/file",
		path(p),
		bearer(c.token),
		rangeHeader(rangeFrom, rangeTo),
	)

	resp, err := req.Exec(&client)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	return resp.Body, nil
}
