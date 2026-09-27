package kalender_test

import (
	"testing"
	"time"

	"github.com/ihleven/ihlvn/pkg/kalender"
)

func date(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatal(err)
	}

	return d
}

// Easter, against the years an almanac agrees on.
//
// Everything movable in the year hangs off this one date, so it is the
// calculation worth pinning: Gauss's formula is right or sixty days are wrong
// together. The years span a full-moon correction and both centuries' century
// terms, which is where a transcription slip would show.
func TestOstern(t *testing.T) {
	for _, tc := range []struct {
		year int
		want string
	}{
		{2024, "2024-03-31"},
		{2025, "2025-04-20"},
		{2026, "2026-04-05"},
		{2027, "2027-03-28"},
		{2030, "2030-04-21"},
		{1999, "1999-04-04"},
		{2038, "2038-04-25"}, // the latest Easter can fall
	} {
		if got := kalender.Ostern(tc.year); !got.Equal(date(t, tc.want)) {
			t.Errorf("Ostern(%d) = %s, want %s", tc.year, got.Format("2006-01-02"), tc.want)
		}
	}
}

// The days that are counted from Easter. If one of these is out, the offset is
// wrong rather than the formula.
func TestDieBeweglichenFeiertage(t *testing.T) {
	const year = 2026 // Easter is 5 April

	for _, tc := range []struct {
		name string
		got  kalender.Feiertag
		want string
	}{
		{"Karfreitag", kalender.Karfreitag(year), "2026-04-03"},
		{"Ostermontag", kalender.Ostermontag(year), "2026-04-06"},
		{"ChristiHimmelfahrt", kalender.ChristiHimmelfahrt(year), "2026-05-14"},
		{"Pfingsten", kalender.Pfingsten(year), "2026-05-24"},
		{"Pfingstmontag", kalender.Pfingstmontag(year), "2026-05-25"},
		{"Fronleichnam", kalender.Fronleichnam(year), "2026-06-04"},
		{"Aschermittwoch", kalender.Aschermittwoch(year), "2026-02-18"},
		{"Rosenmontag", kalender.Rosenmontag(year), "2026-02-16"},
	} {
		if !tc.got.Equal(date(t, tc.want)) {
			t.Errorf("%s(%d) = %s, want %s", tc.name, year, tc.got.Format("2006-01-02"), tc.want)
		}
	}
}

// The days that are fixed, and the ones fixed by a rule rather than a date —
// the fourth Thursday, the last Sunday, the Wednesday before a Sunday.
func TestDieFestenUndDieGeregelten(t *testing.T) {
	for _, tc := range []struct {
		name string
		got  kalender.Feiertag
		want string
	}{
		{"Neujahr", kalender.Neujahr(2026), "2026-01-01"},
		{"TagDerArbeit", kalender.TagDerArbeit(2026), "2026-05-01"},
		{"TagDerDeutschenEinheit", kalender.TagDerDeutschenEinheit(2026), "2026-10-03"},
		{"Weihnachten", kalender.Weihnachten(2026), "2026-12-25"},
		{"Silvester", kalender.Silvester(2026), "2026-12-31"},

		// Fourth Thursday in November.
		{"Thanksgiving", kalender.Thanksgiving(2026), "2026-11-26"},
		// Last Sunday in March and October.
		{"BeginnSommerzeit", kalender.BeginnSommerzeit(2026), "2026-03-29"},
		{"BeginnWinterzeit", kalender.BeginnWinterzeit(2026), "2026-10-25"},
		// Fourth Sunday before 25 December.
		{"ErsterAdvent", kalender.ErsterAdvent(2026), "2026-11-29"},
	} {
		if !tc.got.Equal(date(t, tc.want)) {
			t.Errorf("%s = %s, want %s", tc.name, tc.got.Format("2006-01-02"), tc.want)
		}
	}
}

// A Feiertag is a time.Time that also knows what it is called.
func TestFeiertagPrintsItselfWithItsName(t *testing.T) {
	if got := kalender.Neujahr(2026).String(); got != "01.01.2026 Neujahr" {
		t.Errorf("String() = %q, want %q", got, "01.01.2026 Neujahr")
	}
}
