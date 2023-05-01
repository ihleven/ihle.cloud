package hi

import (
	"net/http"
	"time"

	"github.com/ihleven/errors"
)

var client = http.Client{
	Timeout: 10 * time.Second,
}

func NewHDClient(token string) *hdclient {

	var HiDriveClient = hdclient{
		// Client: &client,
		token: token,
	}
	return &HiDriveClient
}

type hdclient struct {
	// *http.Client
	token string
}

func (hd *hdclient) GetDir(p string) (*meta, error) {

	req := newRequest(GET, "/dir", path(p),
		bearer(hd.token),
		fields("name,path,ctime,has_dirs,mtime,readable,size,type,writable,members.name"),
	)

	meta, err := req.fetchJSON(&client)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	// meta.hdclient = hd
	return meta, nil
}

func (hd *hdclient) GetMeta(p string) (*meta, error) {

	req := newRequest(GET, "/meta",
		path(p),
		bearer(hd.token),
		fields("name,path,category,ctime,has_dirs,mtime,readable,size,type,writable"),
	)

	meta, err := req.fetchJSON(&client)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	// meta.hdclient = hd
	return meta, nil
}
