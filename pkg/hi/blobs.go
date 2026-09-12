package hi

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ihleven/ihlvn/pkg/blob"
)

// blobs presents a Drive as a blob.Blobstore, so assets can be delivered by the
// standard handler rather than by code that knows about HiDrive.
//
// Keys are opaque and untrusted — they come from content an editor controls — so
// they go through SafeKey before they are resolved. That is stricter than what
// Drive accepts from a person browsing: no leading slash, no "~", no traversal
// to reject in the first place.
type blobs struct {
	drive *Drive
}

// Blobs returns this drive as a blob store.
func (d *Drive) Blobs() blob.Blobstore { return &blobs{drive: d} }

func (b *blobs) resolve(key string) (string, error) {
	clean, err := blob.SafeKey(key)
	if err != nil {
		return "", err
	}

	return b.drive.Resolve(clean), nil
}

func (b *blobs) Stat(ctx context.Context, key string) (blob.Object, error) {
	p, err := b.resolve(key)
	if err != nil {
		return blob.Object{}, err
	}

	token, err := b.drive.token()
	if err != nil {
		return blob.Object{}, err
	}

	meta, err := b.drive.client.Meta(ctx, token, p)
	if err != nil {
		return blob.Object{}, translate(err)
	}

	return blob.Object{
		Key:         key,
		Size:        int64(meta.Size_),
		ContentType: meta.Mimetype,
		// HiDrive's id changes when the file does, which is what an ETag has to
		// promise. Quoted because the header syntax requires it.
		ETag:    quoteETag(meta.ID),
		ModTime: time.Unix(meta.MTime, 0),
	}, nil
}

func (b *blobs) OpenRange(ctx context.Context, key string, off, n int64) (io.ReadCloser, error) {
	p, err := b.resolve(key)
	if err != nil {
		return nil, err
	}

	token, err := b.drive.token()
	if err != nil {
		return nil, err
	}

	body, err := b.drive.client.File(ctx, token, p, off, n)
	if err != nil {
		return nil, translate(err)
	}

	return body, nil
}

func quoteETag(v string) string {
	if v == "" {
		return ""
	}
	if strings.HasPrefix(v, `"`) {
		return v
	}

	return `"` + v + `"`
}

// translate maps a HiDrive failure onto the store's vocabulary, so callers can
// tell a missing object from an outage without matching on text.
func translate(err error) error {
	var hiErr Error
	if errors.As(err, &hiErr) && hiErr.HTTPStatus == http.StatusNotFound {
		return blob.ErrNotFound
	}

	return err
}
