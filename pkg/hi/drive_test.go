package hi

import (
	"context"
	"errors"
	"testing"

	"github.com/ihleven/ihlvn/pkg/blob"
)

type stubSource struct{ token string }

func (s stubSource) AccessToken(string) (string, error) { return s.token, nil }

func testDrive() *Drive {
	return NewDrive(stubSource{"t"}, DriveConfig{Alias: "matt.ihle", Root: "/users/matt", Home: "/Bilder"}, nil)
}

// Root is a boundary, not a prefix to be helpful about: a path that tries to
// climb out of it lands inside it instead, because ".." is resolved before the
// root is applied.
func TestResolveCannotEscapeTheRoot(t *testing.T) {
	d := testDrive()

	for _, p := range []string{
		"../../etc/passwd",
		"/../../etc/passwd",
		"a/../../../etc/passwd",
		"....//....//etc",
	} {
		t.Run(p, func(t *testing.T) {
			got := d.Resolve(p)
			if !isBelow(got, "/users/matt") {
				t.Errorf("Resolve(%q) = %q, which is outside the root", p, got)
			}
		})
	}
}

// An empty path and "~" mean home, the way a shell treats them.
func TestResolveAppliesHome(t *testing.T) {
	d := testDrive()

	for _, tc := range []struct{ in, want string }{
		{"", "/users/matt/Bilder"},
		{"~", "/users/matt/Bilder"},
		{"~/2024", "/users/matt/Bilder/2024"},
		{"/Super8", "/users/matt/Super8"},
		{"Super8/a.mp4", "/users/matt/Super8/a.mp4"},
	} {
		t.Run(tc.in, func(t *testing.T) {
			if got := d.Resolve(tc.in); got != tc.want {
				t.Errorf("Resolve(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// A drive with no root configured still resolves to something absolute.
func TestResolveDefaultsTheRoot(t *testing.T) {
	d := NewDrive(stubSource{"t"}, DriveConfig{Alias: "a"}, nil)

	if got := d.Resolve("x/y"); got != "/x/y" {
		t.Errorf("Resolve = %q, want /x/y", got)
	}
}

// Blob keys are stricter than browsing paths: they come from content, so a key
// that looks like navigation is refused rather than interpreted.
func TestBlobsRefuseNavigationKeys(t *testing.T) {
	store := testDrive().Blobs()

	for _, key := range []string{"", "/absolute", "../escape", "bad\\path"} {
		t.Run(key, func(t *testing.T) {
			if _, err := store.Stat(context.Background(), key); !errors.Is(err, blob.ErrInvalidKey) {
				t.Errorf("Stat(%q) = %v, want ErrInvalidKey", key, err)
			}
		})
	}
}

func TestBlobsSatisfyBlobstore(t *testing.T) {
	var _ blob.Blobstore = testDrive().Blobs()
}

func isBelow(p, root string) bool {
	return p == root || len(p) > len(root) && p[:len(root)+1] == root+"/"
}
