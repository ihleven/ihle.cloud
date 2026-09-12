package hi

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// TokenSource yields an access token for a HiDrive alias.
//
// Declared here rather than imported so this package stays independent of how
// credentials are stored: hiauth satisfies it, and a test can pass a stub.
type TokenSource interface {
	AccessToken(alias string) (string, error)
}

// DriveConfig is what an account records about its HiDrive: which credentials
// to use, what may be seen, and where browsing starts. It is the shape stored in
// the account row, so it can be unmarshalled straight from there.
type DriveConfig struct {
	// Alias selects the refresh token held for this account.
	Alias string `json:"alias"`
	// Root confines the drive. Nothing outside it can be addressed.
	Root string `json:"root"`
	// Home is where a path of "" or "~" resolves to — the directory browsing
	// opens at. It is a convenience, not a boundary; Root is the boundary.
	Home string `json:"home"`
}

// Drive is one HiDrive account, seen through its root.
//
// It turns the paths a caller uses into the absolute paths the API wants, gets a
// current access token for each call, and refuses anything that would leave the
// root. The API itself is Client's business.
type Drive struct {
	client *Client
	src    TokenSource
	cfg    DriveConfig
}

// NewDrive returns the drive for one account.
func NewDrive(src TokenSource, cfg DriveConfig, client *Client) *Drive {
	if client == nil {
		client = &Client{}
	}
	if cfg.Root == "" {
		cfg.Root = "/"
	}

	cfg.Root = path.Clean("/" + strings.TrimPrefix(cfg.Root, "/"))

	return &Drive{client: client, src: src, cfg: cfg}
}

// Config reports what this drive is scoped to. Home is here for a caller that
// wants to open browsing at it; Resolve already applies it.
func (d *Drive) Config() DriveConfig { return d.cfg }

// Resolve turns a caller's path into an absolute one below the root.
//
// An empty path and "~" both mean home, the way a shell treats them. Everything
// is then cleaned and placed below the root: ".." is resolved first, so a path
// that tried to climb out simply cannot, and the result is always inside.
func (d *Drive) Resolve(p string) string {
	p = strings.TrimSpace(p)

	switch {
	case p == "" || p == "~":
		p = d.cfg.Home
	case strings.HasPrefix(p, "~/"):
		p = path.Join(d.cfg.Home, strings.TrimPrefix(p, "~/"))
	}

	// Cleaning an absolute path resolves every ".." against "/", so nothing can
	// escape; joining the remainder below the root is then safe by construction.
	inside := path.Clean("/" + strings.TrimPrefix(p, "/"))

	return path.Join(d.cfg.Root, inside)
}

func (d *Drive) token() (string, error) {
	return d.src.AccessToken(d.cfg.Alias)
}

// Meta describes one object on this drive.
func (d *Drive) Meta(ctx context.Context, p string) (*Meta, error) {
	token, err := d.token()
	if err != nil {
		return nil, err
	}

	return d.client.Meta(ctx, token, d.Resolve(p))
}

// Dir lists a directory on this drive. An empty path lists home.
func (d *Drive) Dir(ctx context.Context, p string) (*Meta, error) {
	token, err := d.token()
	if err != nil {
		return nil, err
	}

	return d.client.Dir(ctx, token, d.Resolve(p))
}

// URL returns a pre-signed URL for a file on this drive.
func (d *Drive) URL(ctx context.Context, p string) (*url.URL, error) {
	token, err := d.token()
	if err != nil {
		return nil, err
	}

	return d.client.URL(ctx, token, d.Resolve(p))
}

// File reads part of a file on this drive; n <= 0 reads to the end.
func (d *Drive) File(ctx context.Context, p string, off, n int64) (io.ReadCloser, error) {
	token, err := d.token()
	if err != nil {
		return nil, err
	}

	return d.client.File(ctx, token, d.Resolve(p), off, n)
}

// Thumbnail reads a scaled preview of an image on this drive.
func (d *Drive) Thumbnail(ctx context.Context, p string, params url.Values) (*http.Response, error) {
	token, err := d.token()
	if err != nil {
		return nil, err
	}

	params.Set("path", d.Resolve(p))

	return d.client.Thumbnail(ctx, token, params)
}
