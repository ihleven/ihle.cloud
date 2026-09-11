package authn

import (
	"context"
	"crypto/sha1"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// hashOf is what the checker computes locally; the tests use it to build a
// plausible range response.
func hashOf(password string) (prefix, suffix string) {
	sum := sha1.Sum([]byte(password))
	h := strings.ToUpper(fmt.Sprintf("%x", sum))
	return h[:5], h[5:]
}

func TestBreachCountFindsAKnownPassword(t *testing.T) {
	const password = "Sommer2024!!"
	wantPrefix, suffix := hashOf(password)

	var gotPath, gotPadding string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = strings.TrimPrefix(r.URL.Path, "/")
		gotPadding = r.Header.Get("Add-Padding")
		fmt.Fprintf(w, "0000000000000000000000000000000000000:3\r\n")
		fmt.Fprintf(w, "%s:1523\r\n", suffix)
		fmt.Fprintf(w, "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF:7\r\n")
	}))
	defer srv.Close()

	count, err := (&BreachChecker{BaseURL: srv.URL + "/"}).Count(context.Background(), password)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if count != 1523 {
		t.Errorf("count = %d, want 1523", count)
	}

	// The whole point of the range API: only a five-character prefix is sent.
	if gotPath != wantPrefix {
		t.Errorf("requested %q, want the 5-character prefix %q", gotPath, wantPrefix)
	}
	if len(gotPath) != 5 {
		t.Errorf("sent %d characters of the hash; only 5 may leave the machine", len(gotPath))
	}
	if strings.Contains(gotPath, suffix) {
		t.Error("the request carried the hash suffix")
	}
	if gotPadding != "true" {
		t.Error("padding was not requested, so the response size leaks information")
	}
}

func TestBreachCountIsZeroForAnUnknownPassword(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "0000000000000000000000000000000000000:3\r\n")
	}))
	defer srv.Close()

	count, err := (&BreachChecker{BaseURL: srv.URL + "/"}).Count(context.Background(), "a passphrase nobody has used")
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
}

// Padding entries come back with a count of zero and must not be read as a hit.
func TestBreachCountIgnoresPadding(t *testing.T) {
	const password = "padded"
	_, suffix := hashOf(password)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s:0\r\n", suffix)
	}))
	defer srv.Close()

	count, err := (&BreachChecker{BaseURL: srv.URL + "/"}).Count(context.Background(), password)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d; a zero-count padding entry was read as a breach", count)
	}
}

// The service being unreachable or unhappy is reported, so the caller can decide
// whether to continue rather than being told the password is clean.
func TestBreachCountReportsAFailedCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	if _, err := (&BreachChecker{BaseURL: srv.URL + "/"}).Count(context.Background(), "x"); err == nil {
		t.Error("a failing service was reported as a clean password")
	}

	unreachable := &BreachChecker{BaseURL: "http://127.0.0.1:1/"}
	if _, err := unreachable.Count(context.Background(), "x"); err == nil {
		t.Error("an unreachable service was reported as a clean password")
	}
}

// The suffix comparison is case-insensitive, since the corpus is uppercase but
// nothing guarantees the response is.
func TestBreachCountMatchesCaseInsensitively(t *testing.T) {
	const password = "mixed case"
	_, suffix := hashOf(password)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s:42\r\n", strings.ToLower(suffix))
	}))
	defer srv.Close()

	count, err := (&BreachChecker{BaseURL: srv.URL + "/"}).Count(context.Background(), password)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if count != 42 {
		t.Errorf("count = %d, want 42", count)
	}
}
