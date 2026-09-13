package password

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// breachServer stands in for the range endpoint, reporting the given count for
// the password asked about and nothing for anything else.
// MinLength is the application's number; these tests need some threshold to
// measure against and this is the one the app uses.
const MinLength = 12

func breachServer(t *testing.T, password string, count int) *BreachChecker {
	t.Helper()

	_, suffix := hashOf(password)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if count > 0 {
			fmt.Fprintf(w, "%s:%d\r\n", suffix, count)
		}
		fmt.Fprint(w, "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF:7\r\n")
	}))
	t.Cleanup(srv.Close)

	return &BreachChecker{BaseURL: srv.URL + "/"}
}

// A password that is long and has not leaked is the case where nothing should be
// put to the person at all — an advisory that fires on everything is noise.
func TestCheckFindsNothingToSay(t *testing.T) {
	const password = "a passphrase nobody has ever used"

	advice := Check(context.Background(), breachServer(t, password, 0), password, MinLength)

	if advice.Concerning() {
		t.Errorf("Concerning() = true for a long unbreached password: %+v", advice)
	}
}

func TestCheckReportsAShortPassword(t *testing.T) {
	const password = "kurz"

	advice := Check(context.Background(), breachServer(t, password, 0), password, MinLength)

	if !advice.TooShort {
		t.Error("TooShort = false for a four-character password")
	}
	if advice.Length != 4 {
		t.Errorf("Length = %d, want 4", advice.Length)
	}
	if advice.MinLength != MinLength {
		t.Errorf("MinLength = %d, want the threshold it was measured against", advice.MinLength)
	}
}

// A leak is evidence, so it is reported however long the password is — the case
// the comment in the policy is about.
func TestCheckReportsALongButBreachedPassword(t *testing.T) {
	const password = "correct horse battery staple"

	advice := Check(context.Background(), breachServer(t, password, 1523), password, MinLength)

	if advice.TooShort {
		t.Error("TooShort = true for a 28-character password")
	}
	if advice.Breaches != 1523 {
		t.Errorf("Breaches = %d, want 1523", advice.Breaches)
	}
	if !advice.Concerning() {
		t.Error("Concerning() = false for a password seen 1523 times")
	}
}

// Not being able to ask is not the same as asking and being told the password is
// clean, so it is reported as unknown rather than silently passing.
func TestCheckSaysWhenItCouldNotAsk(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	advice := Check(context.Background(),
		&BreachChecker{BaseURL: srv.URL + "/"}, "a passphrase nobody has ever used", MinLength)

	if advice.Unchecked == "" {
		t.Error("Unchecked is empty after the service refused to answer")
	}
	if advice.Breaches != 0 {
		t.Errorf("Breaches = %d, want 0 when nothing could be established", advice.Breaches)
	}
	if !advice.Concerning() {
		t.Error("Concerning() = false, but the operator should be told the check did not happen")
	}
}

// The ceiling is the one hard bound, and it is counted in characters: eight
// emoji are eight, not thirty-two.
func TestTooLongCountsCharacters(t *testing.T) {
	if TooLong(strings.Repeat("a", MaxLength)) {
		t.Error("a password of exactly the maximum length was refused")
	}
	if !TooLong(strings.Repeat("a", MaxLength+1)) {
		t.Error("a password one character over the maximum was accepted")
	}
	if TooLong(strings.Repeat("🙂", MaxLength)) {
		t.Errorf("%d emoji were counted as more than %d characters",
			MaxLength, MaxLength)
	}
}
