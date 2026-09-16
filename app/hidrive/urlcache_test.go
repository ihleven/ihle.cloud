package hidrive

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestASignedURLIsReusedUntilItExpires(t *testing.T) {
	c := newSignedURLs(time.Minute)
	u, _ := url.Parse("https://stra.to/abc/def")

	if _, ok := c.get("k"); ok {
		t.Error("an empty cache answered")
	}
	c.put("k", u)
	if got, ok := c.get("k"); !ok || got != u {
		t.Error("the URL was not reused")
	}
	c.drop("k")
	if _, ok := c.get("k"); ok {
		t.Error("a dropped URL was still served")
	}
}

func TestAnExpiredURLIsNotReused(t *testing.T) {
	c := newSignedURLs(time.Nanosecond)
	u, _ := url.Parse("https://stra.to/abc/def")

	c.put("k", u)
	time.Sleep(time.Millisecond)

	if _, ok := c.get("k"); ok {
		t.Error("an expired URL was reused; the TTL has to bound it because the store's own lifetime cannot be read")
	}
}

// The statuses that mean "this URL is no longer mine", as opposed to an answer
// worth passing on. A 404 counts: an expired signature and a deleted file look
// the same from here, and re-minting is what tells them apart.
func TestWhichStatusesMeanTheURLIsStale(t *testing.T) {
	for _, code := range []int{401, 403, 404, 410} {
		if !staleStatus(code) {
			t.Errorf("%d should count as stale", code)
		}
	}
	for _, code := range []int{200, 206, 304, 416, 500, 503} {
		if staleStatus(code) {
			t.Errorf("%d should be passed through, not retried", code)
		}
	}
}

// The property the retry depends on: when the store disowns a URL, nothing has
// been written, so the caller can still mint and try again. If ModifyResponse
// ever stopped short-circuiting, this would fail by writing the refusal through.
func TestAStaleResponseWritesNothing(t *testing.T) {
	store := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "gone", http.StatusForbidden)
	}))
	defer store.Close()

	api := NewAPI(token("t"), Shared{Alias: "a", Root: "/"})
	signed, _ := url.Parse(store.URL + "/signed/path")

	w := httptest.NewRecorder()
	if stale := api.proxy(w, httptest.NewRequest(http.MethodGet, "/x", nil), signed); !stale {
		t.Fatal("a 403 from the store was not recognised as a stale URL")
	}
	if w.Body.Len() != 0 {
		t.Errorf("wrote %q; a retry is only possible while nothing has been committed", w.Body)
	}
	if w.Code != http.StatusOK {
		t.Errorf("status %d was committed", w.Code)
	}
}

// A store that answers normally is passed through untouched.
func TestAGoodResponseIsNotTreatedAsStale(t *testing.T) {
	store := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write([]byte("bytes"))
	}))
	defer store.Close()

	api := NewAPI(token("t"), Shared{Alias: "a", Root: "/"})
	signed, _ := url.Parse(store.URL + "/signed/path")

	w := httptest.NewRecorder()
	if stale := api.proxy(w, httptest.NewRequest(http.MethodGet, "/x", nil), signed); stale {
		t.Fatal("a good response was treated as a stale URL")
	}
	if w.Body.String() != "bytes" {
		t.Errorf("body %q", w.Body)
	}
}

// An unreachable store is a bad gateway, not a stale URL: re-minting would not
// help, and swallowing it would answer an empty 200.
func TestAnUnreachableStoreIsABadGateway(t *testing.T) {
	api := NewAPI(token("t"), Shared{Alias: "a", Root: "/"})
	signed, _ := url.Parse("http://127.0.0.1:1/signed/path")

	w := httptest.NewRecorder()
	if stale := api.proxy(w, httptest.NewRequest(http.MethodGet, "/x", nil), signed); stale {
		t.Error("a transport failure was reported as a stale URL")
	}
	if w.Code != http.StatusBadGateway {
		t.Errorf("status %d, want 502", w.Code)
	}
}
