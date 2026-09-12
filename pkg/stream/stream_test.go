package stream

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/ihleven/ihlvn/pkg/blob"
)

// payload is deliberately not uniform, so a wrong offset produces wrong bytes
// rather than accidentally matching.
func payload(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(i % 251)
	}
	return b
}

func testHandler(t *testing.T, name string, data []byte) (*Handler, string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	store, err := blob.NewLocalFS(dir)
	if err != nil {
		t.Fatalf("NewLocalFS: %v", err)
	}
	h := &Handler{
		Store: store,
		Key: func(_ context.Context, id string) (string, error) {
			if id != "sample" {
				return "", ErrNoSuchItem
			}
			return name, nil
		},
		Log: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	return h, name
}

func do(h *Handler, method, id string, headers map[string]string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, "/stream/"+id, nil)
	r.SetPathValue("id", id)
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestServeWholeObject(t *testing.T) {
	data := payload(1000)
	h, _ := testHandler(t, "sample.mp4", data)

	w := do(h, http.MethodGet, "sample", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !bytes.Equal(w.Body.Bytes(), data) {
		t.Error("body does not match the file")
	}
	if got := w.Header().Get("Content-Length"); got != "1000" {
		t.Errorf("Content-Length = %q, want 1000", got)
	}
	if got := w.Header().Get("Accept-Ranges"); got != "bytes" {
		t.Errorf("Accept-Ranges = %q, want bytes", got)
	}
	// Without this a player cannot seek: it is the advertisement that makes the
	// browser issue ranged requests at all.
	if got := w.Header().Get("Content-Type"); got != "video/mp4" {
		t.Errorf("Content-Type = %q, want video/mp4", got)
	}
	if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
	}
	if w.Header().Get("ETag") == "" {
		t.Error("ETag is missing")
	}
}

// The bytes served for a range must be exactly the bytes at that offset.
func TestServeRangedBytesAreExact(t *testing.T) {
	data := payload(1000)
	h, _ := testHandler(t, "sample.mp4", data)

	cases := []struct {
		header    string
		wantRange string
		wantOff   int
		wantLen   int
	}{
		{"bytes=0-99", "bytes 0-99/1000", 0, 100},
		{"bytes=500-", "bytes 500-999/1000", 500, 500},
		{"bytes=-100", "bytes 900-999/1000", 900, 100},
		{"bytes=999-", "bytes 999-999/1000", 999, 1},
		{"bytes=900-5000", "bytes 900-999/1000", 900, 100},
	}

	for _, tc := range cases {
		t.Run(tc.header, func(t *testing.T) {
			w := do(h, http.MethodGet, "sample", map[string]string{"Range": tc.header})

			if w.Code != http.StatusPartialContent {
				t.Fatalf("status = %d, want 206", w.Code)
			}
			if got := w.Header().Get("Content-Range"); got != tc.wantRange {
				t.Errorf("Content-Range = %q, want %q", got, tc.wantRange)
			}
			want := data[tc.wantOff : tc.wantOff+tc.wantLen]
			if !bytes.Equal(w.Body.Bytes(), want) {
				t.Errorf("body = %d bytes at wrong offset, want %d bytes from %d",
					w.Body.Len(), tc.wantLen, tc.wantOff)
			}
			if got, want := w.Header().Get("Content-Length"), tc.wantLen; got != itoa(want) {
				t.Errorf("Content-Length = %q, want %d", got, want)
			}
		})
	}
}

func TestServeUnsatisfiableRange(t *testing.T) {
	h, _ := testHandler(t, "sample.mp4", payload(100))

	w := do(h, http.MethodGet, "sample", map[string]string{"Range": "bytes=500-"})

	if w.Code != http.StatusRequestedRangeNotSatisfiable {
		t.Fatalf("status = %d, want 416", w.Code)
	}
	// 416 must report the real size, which is how a client corrects itself.
	if got := w.Header().Get("Content-Range"); got != "bytes */100" {
		t.Errorf("Content-Range = %q, want bytes */100", got)
	}
}

func TestServeHeadHasNoBody(t *testing.T) {
	h, _ := testHandler(t, "sample.mp4", payload(1000))

	w := do(h, http.MethodHead, "sample", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Errorf("HEAD returned %d bytes, want 0", w.Body.Len())
	}
	if got := w.Header().Get("Content-Length"); got != "1000" {
		t.Errorf("Content-Length = %q, want 1000", got)
	}
}

func TestServeConditionalNotModified(t *testing.T) {
	h, _ := testHandler(t, "sample.mp4", payload(1000))

	first := do(h, http.MethodGet, "sample", nil)
	etag := first.Header().Get("ETag")

	w := do(h, http.MethodGet, "sample", map[string]string{"If-None-Match": etag})

	if w.Code != http.StatusNotModified {
		t.Fatalf("status = %d, want 304", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Errorf("304 returned %d bytes, want 0", w.Body.Len())
	}

	// A different validator must still serve the object.
	w = do(h, http.MethodGet, "sample", map[string]string{"If-None-Match": `"nope"`})
	if w.Code != http.StatusOK {
		t.Errorf("status = %d for stale ETag, want 200", w.Code)
	}
}

func TestServeUnknownItem(t *testing.T) {
	h, _ := testHandler(t, "sample.mp4", payload(10))

	if w := do(h, http.MethodGet, "missing", nil); w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestServeRejectsWrites(t *testing.T) {
	h, _ := testHandler(t, "sample.mp4", payload(10))

	w := do(h, http.MethodDelete, "sample", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

// An escaping key is content-authored, so it must fail closed rather than read
// outside the media root.
func TestServeRejectsEscapingKey(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "ok.mp4"), payload(10), 0o644); err != nil {
		t.Fatal(err)
	}
	secret := filepath.Join(filepath.Dir(dir), "secret.txt")
	if err := os.WriteFile(secret, []byte("classified"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(secret) })

	store, err := blob.NewLocalFS(dir)
	if err != nil {
		t.Fatal(err)
	}
	h := &Handler{
		Store: store,
		Key:   func(_ context.Context, _ string) (string, error) { return "../secret.txt", nil },
		Log:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	w := do(h, http.MethodGet, "evil", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
	if bytes.Contains(w.Body.Bytes(), []byte("classified")) {
		t.Fatal("escaping key leaked a file outside the media root")
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// cancellingStore reports the error a store returns when the caller's context
// has been cancelled — which is what happens whenever a player abandons a
// range request mid-seek.
type cancellingStore struct{ obj blob.Object }

func (c cancellingStore) Stat(context.Context, string) (blob.Object, error) {
	return c.obj, nil
}

func (c cancellingStore) OpenRange(ctx context.Context, _ string, _, _ int64) (io.ReadCloser, error) {
	return nil, fmt.Errorf("fetching: %w", context.Canceled)
}

// A client that walks away must not be recorded as a server error: seeking
// would otherwise produce a stream of 500s and error logs during normal use.
func TestServeTreatsClientCancellationAsNormal(t *testing.T) {
	h := &Handler{
		Store: cancellingStore{obj: blob.Object{Size: 1000, ContentType: "video/mp4"}},
		Key:   func(context.Context, string) (string, error) { return "clip.mp4", nil },
		Log:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	r := httptest.NewRequest(http.MethodGet, "/stream/clip", nil)
	r.SetPathValue("id", "clip")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code == http.StatusInternalServerError {
		t.Error("an abandoned request must not be reported as a server error")
	}
	if w.Code != StatusClientClosedRequest {
		t.Errorf("status = %d, want %d", w.Code, StatusClientClosedRequest)
	}
}

// A genuine store failure must still be a 500, or real problems get hidden
// behind the cancellation path.
func TestServeStillReportsRealFailures(t *testing.T) {
	h := &Handler{
		Store: failingStore{obj: blob.Object{Size: 1000, ContentType: "video/mp4"}},
		Key:   func(context.Context, string) (string, error) { return "clip.mp4", nil },
		Log:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	r := httptest.NewRequest(http.MethodGet, "/stream/clip", nil)
	r.SetPathValue("id", "clip")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500 for a real store failure", w.Code)
	}
}

type failingStore struct{ obj blob.Object }

func (f failingStore) Stat(context.Context, string) (blob.Object, error) { return f.obj, nil }
func (f failingStore) OpenRange(context.Context, string, int64, int64) (io.ReadCloser, error) {
	return nil, errors.New("disk on fire")
}
