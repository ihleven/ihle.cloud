package films

import (
	"strings"
	"testing"
)

// The output has to be a file a browser will parse; these are the ways a
// hand-rolled writer usually fails to be one.
func TestRenderVTT(t *testing.T) {
	got := renderVTT([]Scene{
		{Title: "Wohnzimmer", Start: 0, End: 11.8},
		{Title: "Bescherung", Start: 11.82, End: 62},
		{Title: "Großeltern", Start: 62.58, End: 81},
	}, 0)

	want := "WEBVTT\n" +
		"\n00:00:00.000 --> 00:00:11.800\nWohnzimmer\n" +
		"\n00:00:11.820 --> 00:01:02.000\nBescherung\n" +
		"\n00:01:02.580 --> 00:01:21.000\nGroßeltern\n"

	if got != want {
		t.Errorf("renderVTT =\n%q\nwant\n%q", got, want)
	}
}

// A scene runs until the next one starts, so an absent end is the normal case
// rather than an error.
func TestRenderVTTClosesAnOpenSceneWithTheNext(t *testing.T) {
	got := renderVTT([]Scene{
		{Title: "Erste", Start: 0},
		{Title: "Zweite", Start: 30},
	}, 90)

	for _, want := range []string{"00:00:00.000 --> 00:00:30.000", "00:00:30.000 --> 00:01:30.000"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
}

// With nothing to close the last scene there is no cue to write. Emitting one
// with a zero or equal end would produce a file the browser rejects, taking the
// whole track with it.
func TestRenderVTTDropsAnUnclosableLastScene(t *testing.T) {
	got := renderVTT([]Scene{
		{Title: "Erste", Start: 0, End: 30},
		{Title: "Letzte", Start: 30},
	}, 0)

	if strings.Contains(got, "Letzte") {
		t.Errorf("wrote a cue with no end:\n%s", got)
	}
	if !strings.Contains(got, "Erste") {
		t.Errorf("dropped a closable scene too:\n%s", got)
	}
}

func TestRenderVTTOfNothingIsStillAValidFile(t *testing.T) {
	if got := renderVTT(nil, 0); got != "WEBVTT\n" {
		t.Errorf("renderVTT(nil) = %q", got)
	}
}

func TestVTTTime(t *testing.T) {
	for seconds, want := range map[float64]string{
		0:        "00:00:00.000",
		11.8:     "00:00:11.800",
		62.58:    "00:01:02.580",
		308.16:   "00:05:08.160",
		3661.001: "01:01:01.001",
		-5:       "00:00:00.000",
		// Binary floating point cannot hold 106.59 exactly; rounding rather
		// than truncating is what keeps it from becoming 106.589.
		106.59: "00:01:46.590",
	} {
		if got := vttTime(seconds); got != want {
			t.Errorf("vttTime(%v) = %q, want %q", seconds, got, want)
		}
	}
}
