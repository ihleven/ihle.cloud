package main

import (
	"strings"
	"testing"

	"github.com/ihleven/ihlvn/pkg/password"
)

// The wording lives here rather than with the policy, so this is where it is
// checked. What matters is that the numbers the advice carries reach the person
// being asked: "short" without saying how short, or "breached" without saying
// how often, is not advice they can weigh.

func TestWarningsSayTheNumbers(t *testing.T) {
	said := strings.Join(warnings(password.Advice{
		Length: 4, TooShort: true, MinLength: 12, Breaches: 1523,
	}), " ")

	for _, want := range []string{"4 characters", "12 or more", "1523 times"} {
		if !strings.Contains(said, want) {
			t.Errorf("warnings did not mention %q: %s", want, said)
		}
	}
}

// A leak outranks a length, because it is evidence rather than a heuristic —
// and whoever is reading a terminal reads the first line most carefully.
func TestALeakIsSaidFirst(t *testing.T) {
	said := warnings(password.Advice{Length: 4, TooShort: true, MinLength: 12, Breaches: 2})

	if len(said) != 2 {
		t.Fatalf("got %d warnings, want one for the leak and one for the length: %v", len(said), said)
	}
	if !strings.Contains(said[0], "breaches") {
		t.Errorf("the first line is not about the leak: %q", said[0])
	}
}

// Not being able to ask is its own thing, and must not be reported as a clean
// result or hidden.
func TestAnUncheckedPasswordSaysSo(t *testing.T) {
	said := warnings(password.Advice{Length: 30, MinLength: 12, Unchecked: "no network"})

	if len(said) != 1 || !strings.Contains(said[0], "no network") {
		t.Errorf("warnings = %v, want the reason the check did not happen", said)
	}
}

// A password with nothing wrong with it produces nothing to say. An advisory
// that fires on everything is noise.
func TestNothingToSay(t *testing.T) {
	if said := warnings(password.Advice{Length: 30, MinLength: 12}); len(said) != 0 {
		t.Errorf("warnings = %v, want none", said)
	}
}
