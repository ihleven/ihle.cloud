package blob

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestSafeKey(t *testing.T) {
	valid := map[string]string{
		"sample.mp4":          "sample.mp4",
		"videos/sample.mp4":   "videos/sample.mp4",
		"a/b/c/d.webm":        "a/b/c/d.webm",
		"videos/./sample.mp4": "videos/sample.mp4",
		"videos/sub/../s.mp4": "videos/s.mp4",
		"with space.mp4":      "with space.mp4",
		"Ümlaut.mp4":          "Ümlaut.mp4",
	}
	for in, want := range valid {
		got, err := SafeKey(in)
		if err != nil {
			t.Errorf("SafeKey(%q) unexpected error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("SafeKey(%q) = %q, want %q", in, got, want)
		}
	}

	// Each of these would read outside the media root if it were honoured.
	invalid := []string{
		"",
		"..",
		"../secret",
		"../../etc/passwd",
		"videos/../../secret",
		"/etc/passwd",
		"/",
		".",
		`videos\sample.mp4`,
		"videos/\x00sample.mp4",
		"videos/\nsample.mp4",
	}
	for _, in := range invalid {
		if got, err := SafeKey(in); !errors.Is(err, ErrInvalidKey) {
			t.Errorf("SafeKey(%q) = %q, %v; want ErrInvalidKey", in, got, err)
		}
	}
}

func TestLocalFSStatAndRead(t *testing.T) {
	dir := t.TempDir()
	data := []byte("0123456789")
	if err := os.WriteFile(filepath.Join(dir, "clip.mp4"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	store, err := NewLocalFS(dir)
	if err != nil {
		t.Fatalf("NewLocalFS: %v", err)
	}

	obj, err := store.Stat(context.Background(), "clip.mp4")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if obj.Size != int64(len(data)) {
		t.Errorf("Size = %d, want %d", obj.Size, len(data))
	}
	if obj.ContentType != "video/mp4" {
		t.Errorf("ContentType = %q, want video/mp4", obj.ContentType)
	}
	if obj.ETag == "" || obj.ModTime.IsZero() {
		t.Error("ETag and ModTime must be set so conditional requests work")
	}

	rc, err := store.OpenRange(context.Background(), "clip.mp4", 3, 4)
	if err != nil {
		t.Fatalf("OpenRange: %v", err)
	}
	defer rc.Close()
	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("reading range: %v", err)
	}
	if string(got) != "3456" {
		t.Errorf("OpenRange(3,4) = %q, want %q", got, "3456")
	}
}

func TestLocalFSMissingAndDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	store, err := NewLocalFS(dir)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.Stat(context.Background(), "nope.mp4"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Stat(missing) err = %v, want ErrNotFound", err)
	}
	// A directory must not present itself as a zero-byte media object.
	if _, err := store.Stat(context.Background(), "sub"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Stat(dir) err = %v, want ErrNotFound", err)
	}
	if _, err := store.Stat(context.Background(), "../outside"); !errors.Is(err, ErrInvalidKey) {
		t.Errorf("Stat(escaping) err = %v, want ErrInvalidKey", err)
	}
}

func TestLocalFSKeys(t *testing.T) {
	dir := t.TempDir()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(dir, "clips"), 0o755))
	must(os.WriteFile(filepath.Join(dir, "b.mp4"), []byte("x"), 0o644))
	must(os.WriteFile(filepath.Join(dir, "a.mp4"), []byte("x"), 0o644))
	must(os.WriteFile(filepath.Join(dir, "clips", "c.webm"), []byte("x"), 0o644))
	must(os.WriteFile(filepath.Join(dir, ".hidden.mp4"), []byte("x"), 0o644))

	store, err := NewLocalFS(dir)
	must(err)

	keys, err := store.Keys()
	must(err)

	want := []string{"a.mp4", "b.mp4", "clips/c.webm"}
	if len(keys) != len(want) {
		t.Fatalf("Keys() = %v, want %v", keys, want)
	}
	for i := range want {
		if keys[i] != want[i] {
			t.Fatalf("Keys() = %v, want %v", keys, want)
		}
	}
}

func TestNewLocalFSRejectsNonDirectory(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f.mp4")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := NewLocalFS(file); err == nil {
		t.Error("NewLocalFS(file) should fail")
	}
	if _, err := NewLocalFS(filepath.Join(dir, "missing")); err == nil {
		t.Error("NewLocalFS(missing) should fail")
	}
}
