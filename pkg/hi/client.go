package hi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// DefaultBaseURL is HiDrive's REST API.
const DefaultBaseURL = "https://api.hidrive.strato.com/2.1"

// Client consumes the HiDrive API: one method per endpoint, taking an access
// token and the parameters the endpoint documents.
//
// It holds no credential and no account. A token is a per-call argument because
// one lives about an hour — a client that captured one at construction would
// outlive it, which is how a long-lived handler ends up authenticating with an
// expired token. Whose drive is being read is likewise not its business; that is
// Drive's.
type Client struct {
	// HTTP is the client used for requests. The zero value is a client with a
	// generous timeout, which suits large files.
	HTTP *http.Client
	// BaseURL is the API root. Empty means DefaultBaseURL; a test points it at
	// a stub.
	BaseURL string
}

func (c *Client) httpClient() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}

	return &http.Client{Timeout: 100 * time.Second}
}

func (c *Client) baseURL() string {
	if c.BaseURL != "" {
		return c.BaseURL
	}

	return DefaultBaseURL
}

// do issues a request against one endpoint, authenticated with token.
func (c *Client) do(ctx context.Context, endpoint, token string, params url.Values, options ...func(*http.Request)) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL()+endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, Error{HTTPStatus: http.StatusInternalServerError, Message: "Couldn't create request: " + err.Error()}
	}

	request.Header.Set("Authorization", "Bearer "+token)

	for _, apply := range options {
		apply(request)
	}

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return nil, Error{HTTPStatus: http.StatusBadGateway, Message: "requesting " + endpoint + ": " + err.Error()}
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		defer io.Copy(io.Discard, resp.Body)

		apiErr := Error{HTTPStatus: resp.StatusCode}
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return nil, Error{HTTPStatus: resp.StatusCode, Message: "HiDrive refused the request and its reason was unreadable"}
		}

		return nil, apiErr
	}

	return resp, nil
}

// metaFields and dirFields are the "need to know" field lists the API asks for:
// requesting everything costs response time.
const (
	metaFields = "id,name,path,category,nmembers,ctime,has_dirs,mtime,readable,size,type,writable,mime_type,members,members.id,members.name,image.exif,image.height,image.width"
	dirFields  = "id,name,path,category,nmembers,chash,ctime,mtime,has_dirs,readable,rshare,size,type,writable,members,members.category,members.chash,members.ctime,members.has_dirs,members.image.exif,members.image.height,members.image.width,members.mime_type,members.name,members.nmembers,members.size,members.type,mhash,mohash,nhash,parent_id"
)

// Meta describes one filesystem object.
func (c *Client) Meta(ctx context.Context, token, path string) (*Meta, error) {
	return c.meta(ctx, "/meta", token, path, metaFields)
}

// Dir describes a directory and its members.
func (c *Client) Dir(ctx context.Context, token, path string) (*Meta, error) {
	return c.meta(ctx, "/dir", token, path, dirFields)
}

func (c *Client) meta(ctx context.Context, endpoint, token, path, fields string) (*Meta, error) {
	resp, err := c.do(ctx, endpoint, token, url.Values{
		"path":   []string{path},
		"fields": []string{fields},
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	defer io.Copy(io.Discard, resp.Body)

	var meta Meta
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return nil, Error{Message: "could not read the " + endpoint + " response: " + err.Error()}
	}

	return &meta, nil
}

// URL returns a pre-signed URL for a file. It carries no credential, so it can
// be handed to a browser — which also means it is shareable until it expires.
func (c *Client) URL(ctx context.Context, token, path string) (*url.URL, error) {
	resp, err := c.do(ctx, "/file/url", token, url.Values{"path": []string{path}})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	defer io.Copy(io.Discard, resp.Body)

	var response struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, Error{Message: "could not read the file/url response: " + err.Error()}
	}

	return url.Parse(response.URL)
}

// File reads a file. off and n select a byte range; n <= 0 reads to the end.
// The caller closes the reader.
func (c *Client) File(ctx context.Context, token, path string, off, n int64) (io.ReadCloser, error) {
	resp, err := c.do(ctx, "/file", token, url.Values{"path": []string{path}}, byteRange(off, n))
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// Thumbnail reads a scaled preview of an image. The caller closes the reader.
func (c *Client) Thumbnail(ctx context.Context, token string, params url.Values) (*http.Response, error) {
	return c.do(ctx, "/file/thumbnail", token, params)
}

// byteRange asks for part of a file, in the inclusive form HTTP uses.
func byteRange(off, n int64) func(*http.Request) {
	return func(req *http.Request) {
		if off <= 0 && n <= 0 {
			return
		}

		spec := "bytes=" + strconv.FormatInt(off, 10) + "-"
		if n > 0 {
			spec += strconv.FormatInt(off+n-1, 10)
		}

		req.Header.Set("Range", spec)
	}
}
