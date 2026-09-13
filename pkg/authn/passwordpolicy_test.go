package authn

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
func TestCheckPasswordFindsNothingToSay(t *testing.T) {
	const password = "a passphrase nobody has ever used"

	advice := CheckPassword(context.Background(), breachServer(t, password, 0), password)

	if advice.Concerning() {
		t.Errorf("Concerning() = true for a long unbreached password: %+v", advice)
	}
	if len(advice.Warnings()) != 0 {
		t.Errorf("Warnings() = %v, want none", advice.Warnings())
	}
}

func TestCheckPasswordReportsAShortPassword(t *testing.T) {
	const password = "kurz"

	advice := CheckPassword(context.Background(), breachServer(t, password, 0), password)

	if !advice.TooShort {
		t.Error("TooShort = false for a four-character password")
	}
	if advice.Length != 4 {
		t.Errorf("Length = %d, want 4", advice.Length)
	}
	if !strings.Contains(strings.Join(advice.Warnings(), " "), "4 characters") {
		t.Errorf("Warnings() = %v, want the length said plainly", advice.Warnings())
	}
}

// A leak is evidence, so it is reported however long the password is — the case
// the comment in the policy is about.
func TestCheckPasswordReportsALongButBreachedPassword(t *testing.T) {
	const password = "correct horse battery staple"

	advice := CheckPassword(context.Background(), breachServer(t, password, 1523), password)

	if advice.TooShort {
		t.Error("TooShort = true for a 28-character password")
	}
	if advice.Breaches != 1523 {
		t.Errorf("Breaches = %d, want 1523", advice.Breaches)
	}
	if !advice.Concerning() {
		t.Error("Concerning() = false for a password seen 1523 times")
	}
	if !strings.Contains(strings.Join(advice.Warnings(), " "), "1523 times") {
		t.Errorf("Warnings() = %v, want the count", advice.Warnings())
	}
}

// Not being able to ask is not the same as asking and being told the password is
// clean, so it is reported as unknown rather than silently passing.
func TestCheckPasswordSaysWhenItCouldNotAsk(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	advice := CheckPassword(context.Background(),
		&BreachChecker{BaseURL: srv.URL + "/"}, "a passphrase nobody has ever used")

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
	if TooLong(strings.Repeat("a", MaxPasswordLength)) {
		t.Error("a password of exactly the maximum length was refused")
	}
	if !TooLong(strings.Repeat("a", MaxPasswordLength+1)) {
		t.Error("a password one character over the maximum was accepted")
	}
	if TooLong(strings.Repeat("🙂", MaxPasswordLength)) {
		t.Errorf("%d emoji were counted as more than %d characters",
			MaxPasswordLength, MaxPasswordLength)
	}
}
