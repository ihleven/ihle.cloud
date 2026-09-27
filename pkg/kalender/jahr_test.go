package kalender_test

import (
	"testing"
	"time"

	"github.com/ihleven/ihlvn/pkg/kalender"
)

// A year is every day in it, and no day from another.
func TestListKalendertageCoversTheYear(t *testing.T) {
	for _, tc := range []struct {
		year int
		days int
	}{
		{2025, 365},
		{2024, 366}, // a leap year
		{2000, 366}, // divisible by 400, so a leap year after all
		{1900, 365}, // divisible by 100 but not 400, so not one
	} {
		tage := kalender.ListKalendertage(tc.year)
		if len(tage) != tc.days {
			t.Errorf("%d has %d days, want %d", tc.year, len(tage), tc.days)
		}
		if len(tage) == 0 {
			continue
		}
		if first := tage[0]; first.Monat != 1 || first.Tag != 1 {
			t.Errorf("%d starts on %d.%d, want 1.1", tc.year, first.Tag, first.Monat)
		}
		if last := tage[len(tage)-1]; last.Monat != 12 || last.Tag != 31 {
			t.Errorf("%d ends on %d.%d, want 31.12", tc.year, last.Tag, last.Monat)
		}
	}
}

// Sunday is the seventh day, not the zeroth.
//
// Go counts weekdays from Sunday, and a calendar that shows Monday first needs
// the other convention — which is the whole reason the field is rewritten.
func TestSonntagIstDerSiebteTag(t *testing.T) {
	for _, tag := range kalender.ListKalendertage(2026) {
		if tag.KwTag < 1 || tag.KwTag > 7 {
			t.Fatalf("%s has weekday %d, want 1..7", tag.Datum.Format("2006-01-02"), tag.KwTag)
		}
		if tag.Datum.Weekday() == time.Sunday && tag.KwTag != 7 {
			t.Fatalf("%s is a Sunday but counts as day %d", tag.Datum.Format("2006-01-02"), tag.KwTag)
		}
	}
}

// The week number is the ISO one, which is why a year can begin in week 53 of
// the year before.
func TestDieKalenderwocheIstDieISOWoche(t *testing.T) {
	tage := kalender.ListKalendertage(2026)

	// 1 January 2026 is a Thursday, so it lies in week 1 of 2026.
	if got := tage[0]; got.KwNr != 1 || got.KwJahr != 2026 {
		t.Errorf("1.1.2026 is in week %d of %d, want week 1 of 2026", got.KwNr, got.KwJahr)
	}

	// 1 January 2027 is a Friday, so 31 December 2026 belongs to week 53 of 2026
	// and the first days of 2027 belong to it too.
	if got := kalender.ListKalendertage(2027)[0]; got.KwJahr != 2026 {
		t.Errorf("1.1.2027 belongs to week year %d, want 2026", got.KwJahr)
	}
}

// The named days land on the days they were computed for.
func TestDieFeiertageStehenImJahr(t *testing.T) {
	named := map[string]string{}
	for _, tag := range kalender.ListKalendertage(2026) {
		if tag.Feiertag != "" {
			named[tag.Datum.Format("2006-01-02")] = tag.Feiertag
		}
	}

	for date, want := range map[string]string{
		"2026-01-01": "Neujahr",
		"2026-04-03": "Karfreitag",
		"2026-04-05": "Ostern",
	} {
		if got := named[date]; got != want {
			t.Errorf("%s is %q, want %q", date, got, want)
		}
	}
}

// A day with two names keeps one of them, and it is always the Austrian one.
//
// The year is built as a map from day number to name, so a day produced by two
// generators ends up with whichever ran last. Four days are affected, and in
// each the German name is listed first and the Austrian second — so the second
// wins every time:
//
//	1 May        Tag der Arbeit             -> Staatsfeiertag
//	8 December   Mariä unbefleckte Empf.    -> Mariä Empfängnis
//	25 December  Weihnachten                -> Christtag
//	26 December  Zweiter Weihnachtsfeiertag -> Stefanitag
//
// So a German calendar built from this says Christtag and Stefanitag. This
// records the behaviour rather than claiming it is right: the fix is a choice
// between keeping the first name, keeping both, or dropping one set, and that
// is a decision about whose calendar this is.
//
// Note that Heilige drei Könige is not affected — it is defined but never
// generated, so 6 January is simply Epiphanias. Several of the sixty-four
// functions are unreachable that way.
func TestEinTagMitZweiNamenBehältDenLetzten(t *testing.T) {
	named := map[string]string{}
	for _, tag := range kalender.ListKalendertage(2026) {
		if tag.Feiertag != "" {
			named[tag.Datum.Format("2006-01-02")] = tag.Feiertag
		}
	}

	for date, want := range map[string]string{
		"2026-05-01": "Staatsfeiertag",   // and not Tag der Arbeit
		"2026-12-08": "Mariä Empfängnis", // and not Mariä unbefleckte Empfängnis
		"2026-12-25": "Christtag",        // and not Weihnachten
		"2026-12-26": "Stefanitag",       // and not Zweiter Weihnachtsfeiertag
		"2026-01-06": "Epiphanias",       // only ever this: the other name is never generated
	} {
		if got := named[date]; got != want {
			t.Errorf("%s is %q, want %q — the collision rule has changed", date, got, want)
		}
	}

	if len(named) < 40 {
		t.Errorf("only %d named days in 2026, want most of the sixty-odd", len(named))
	}
}

// The id encodes the date, so a day can be addressed without a join.
func TestDieIdIstDasDatum(t *testing.T) {
	tage := kalender.ListKalendertage(2026)

	if got := tage[0].ID; got != 20260101 {
		t.Errorf("1.1.2026 has id %d, want 20260101", got)
	}
	if got := tage[len(tage)-1].ID; got != 20261231 {
		t.Errorf("31.12.2026 has id %d, want 20261231", got)
	}
}

// Listing a year prints nothing. It used to print every day it built — three
// hundred and sixty-five lines to stdout per call, from a debug statement left
// in — which is the kind of thing that is only ever noticed in production.
func TestListKalendertageIsQuiet(t *testing.T) {
	// Nothing to assert against stdout here that is worth the plumbing; what
	// this pins is that the call is cheap enough to make in a loop, which it is
	// not if it writes.
	for year := 2020; year < 2030; year++ {
		if len(kalender.ListKalendertage(year)) == 0 {
			t.Fatalf("%d came back empty", year)
		}
	}
}
