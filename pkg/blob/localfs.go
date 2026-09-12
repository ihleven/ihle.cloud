package blob

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// LocalFS serves media from a directory on disk.
//
// It exists so the whole delivery path -- ranged reads, conditional requests,
// the player page -- can be built and tested before any HiDrive credential
// exists, and so tests never need the network. The HiDrive store implements the
// same interface, which is what makes it substitutable.
type LocalFS struct {
	root string
}

// NewLocalFS returns a store rooted at dir, which must be an existing directory.
func NewLocalFS(dir string) (*LocalFS, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("blob: resolving media dir: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("blob: media dir %s: %w", abs, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("blob: media dir %s is not a directory", abs)
	}
	return &LocalFS{root: abs}, nil
}

// Root reports the absolute directory this store serves.
func (l *LocalFS) Root() string { return l.root }

// resolve validates the key and maps it to a path inside the root. The
// containment check after Abs is a second line of defence behind SafeKey:
// symlinks and platform-specific path quirks are caught here rather than
// trusted to string handling.
func (l *LocalFS) resolve(key string) (string, error) {
	safe, err := SafeKey(key)
	if err != nil {
		return "", err
	}
	full := filepath.Join(l.root, filepath.FromSlash(safe))
	rel, err := filepath.Rel(l.root, full)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", ErrInvalidKey
	}
	return full, nil
}

func (l *LocalFS) Stat(_ context.Context, key string) (Object, error) {
	full, err := l.resolve(key)
	if err != nil {
		return Object{}, err
	}
	info, err := os.Stat(full)
	if err != nil {
		if os.IsNotExist(err) {
			return Object{}, ErrNotFound
		}
		return Object{}, fmt.Errorf("blob: stat %s: %w", key, err)
	}
	// A directory is not a media object; reporting NotFound avoids leaking the
	// layout of the media root.
	if info.IsDir() {
		return Object{}, ErrNotFound
	}
	return Object{
		Key:         key,
		Size:        info.Size(),
		ContentType: contentType(key),
		ETag:        etag(info),
		ModTime:     info.ModTime(),
	}, nil
}

func (l *LocalFS) OpenRange(_ context.Context, key string, off, n int64) (io.ReadCloser, error) {
	full, err := l.resolve(key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(full)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("blob: open %s: %w", key, err)
	}
	if off > 0 {
		if _, err := f.Seek(off, io.SeekStart); err != nil {
			f.Close()
			return nil, fmt.Errorf("blob: seek %s: %w", key, err)
		}
	}
	return sectionReader{Reader: io.LimitReader(f, n), closer: f}, nil
}

// Keys lists the media files under the root, as slash-separated keys. It is not
// part of Blobstore: only the Phase 1 fixture catalogue needs it, so that
// dropping a file into the media directory is enough to see it in the app.
func (l *LocalFS) Keys() ([]string, error) {
	var keys []string
	err := filepath.WalkDir(l.root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		rel, err := filepath.Rel(l.root, p)
		if err != nil {
			return err
		}
		keys = append(keys, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("blob: listing %s: %w", l.root, err)
	}
	sort.Strings(keys)
	return keys, nil
}

// sectionReader ties a limited reader to the file it reads from, so closing the
// response body closes the descriptor.
type sectionReader struct {
	io.Reader
	closer io.Closer
}

func (s sectionReader) Close() error { return s.closer.Close() }

// contentType derives a MIME type from the key's extension. Guessing from
// content is deliberately avoided: the extension is what the store itself will
// report, and a wrong sniff on media is worse than a generic type.
func contentType(key string) string {
	if ct := mime.TypeByExtension(strings.ToLower(filepath.Ext(key))); ct != "" {
		return ct
	}
	return "application/octet-stream"
}

// etag is a weak-by-nature validator built from size and mtime. It only has to
// change when the file changes, which is enough for If-None-Match on media.
func etag(info os.FileInfo) string {
	return fmt.Sprintf(`"%x-%x"`, info.ModTime().UnixNano(), info.Size())
}
