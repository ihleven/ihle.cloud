// Package blob is the storage seam of the media app: everything the app needs
// from a media store, and nothing more.
//
// Two methods suffice because the CMS entry carries the storage key, so the app
// never has to list or search a directory. That keeps the HiDrive
// implementation small and makes a local-filesystem implementation enough to
// develop and test the whole delivery path without credentials.
package blob

import (
	"context"
	"errors"
	"io"
	"path"
	"strings"
	"time"
)

var (
	// ErrNotFound reports that a key does not exist in the store.
	ErrNotFound = errors.New("blob: not found")

	// ErrInvalidKey reports a key that is unsafe or malformed. Keys reaching a
	// store come from CMS content and are therefore untrusted input.
	ErrInvalidKey = errors.New("blob: invalid key")
)

// Object is the metadata needed to answer a media request without reading the
// payload: enough for HEAD, for conditional requests, and for the Content-Range
// arithmetic of a ranged response.
type Object struct {
	Key         string
	Size        int64
	ContentType string
	ETag        string
	ModTime     time.Time
}

// Blobstore is a read-only media store addressed by opaque keys.
type Blobstore interface {
	// Stat returns the object's metadata, or ErrNotFound.
	Stat(ctx context.Context, key string) (Object, error)

	// OpenRange returns n bytes starting at off. The caller must Close the
	// reader. Callers are expected to have clamped off and n against the size
	// reported by Stat.
	OpenRange(ctx context.Context, key string, off, n int64) (io.ReadCloser, error)
}

// SafeKey validates and normalises an untrusted storage key.
//
// Keys originate in CMS content, so an editor ultimately controls them. Without
// this check, a key like "../../etc/passwd" would read outside the configured
// media root -- on the local filesystem now, and in the whole HiDrive account
// once that store is wired up. Rejecting rather than sanitising is deliberate:
// a key that needed fixing is a content bug worth surfacing.
func SafeKey(key string) (string, error) {
	if key == "" {
		return "", ErrInvalidKey
	}
	// Backslashes would be path separators on some systems, and control bytes
	// have no business in a key.
	if strings.ContainsAny(key, `\`) {
		return "", ErrInvalidKey
	}
	for _, r := range key {
		if r < 0x20 || r == 0x7f {
			return "", ErrInvalidKey
		}
	}
	if strings.HasPrefix(key, "/") {
		return "", ErrInvalidKey
	}

	// path.Clean resolves "." and ".." segments; anything that still escapes
	// upwards afterwards was an attempt to leave the root.
	cleaned := path.Clean(key)
	switch {
	case cleaned == "." || cleaned == "/":
		return "", ErrInvalidKey
	case cleaned == "..", strings.HasPrefix(cleaned, "../"):
		return "", ErrInvalidKey
	case strings.HasPrefix(cleaned, "/"):
		return "", ErrInvalidKey
	}
	return cleaned, nil
}
