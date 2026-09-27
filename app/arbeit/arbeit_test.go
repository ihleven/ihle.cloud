package arbeit

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ihleven/ihlvn/pkg/kalender"
)

// noRepo satisfies the storage seam and remembers almost nothing.
//
// Everything worth testing without a database is arithmetic over what the
// caller passed in, so the stub answers empty and records only the two things a
// test needs to see: which spans were stored and which were dropped.
type noRepo struct {
	spannen []Zeitspanne
	saved   []Zeitspanne
	deleted []int
}

func (r *noRepo) ListArbeitstage(int, int, int, int) ([]Arbeitstag, error)  { return nil, nil }
func (r *noRepo) ReadArbeitstag(int, time.Time) (*Arbeitstag, error)        { return &Arbeitstag{}, nil }
func (r *noRepo) ListZeitspannen(int, time.Time) ([]Zeitspanne, error)      { return r.spannen, nil }
func (r *noRepo) Close()                                                    {}
func (r *noRepo) SelectArbeitsmonate(int, int, int) ([]Arbeitsmonat, error) { return nil, nil }
func (r *noRepo) RetrieveArbeitsjahre(int, int) ([]Arbeitsjahr, error)      { return nil, nil }
func (r *noRepo) UpsertKalendertag(kalender.Tag) error                      { return nil }
func (r *noRepo) SetupArbeitsmonat(int, string, int, int) error             { return nil }
func (r *noRepo) ListUrlaube(int, int, int) ([]Urlaub, error)               { return nil, nil }
func (r *noRepo) SaveArbeitstag(int, time.Time, string, Arbeitstag) error   { return nil }

func (r *noRepo) SetupArbeitsjahr(int, string, int, *time.Time, *time.Time) error { return nil }

func (r *noRepo) UpsertZeitspanne(_ int, _ time.Time, z *Zeitspanne) error {
	r.saved = append(r.saved, *z)

	return nil
}

func (r *noRepo) DeleteZeitspanne(_ int, _ time.Time, nr int) error {
	r.deleted = append(r.deleted, nr)

	return nil
}

func at(t *testing.T, clock string) *time.Time {
	t.Helper()

	parsed, err := time.Parse("2006-01-02 15:04", "2026-03-02 "+clock)
	if err != nil {
		t.Fatal(err)
	}

	return &parsed
}

// The rule the whole thing exists for: what a worked day comes to.
//
//	brutto    = ende - start
//	netto     = brutto - pausen + extra
//	differenz = soll - netto
//
// Eight hours at the desk with three quarters of an hour for lunch is seven and
// a quarter worked against a soll of eight, so the day is three quarters of an
// hour short.
//
// Note that the answer comes back through the argument: UpdateArbeitstag writes
// into the Arbeitstag it was handed and returns (nil, nil) on success. A caller
// that reads the return value gets nothing, which is worth knowing before
// anything is built on top of it.
func TestWasEinArbeitstagErgibt(t *testing.T) {
	uc := NewUsecase(&noRepo{})

	tag := &Arbeitstag{
		Soll:  8,
		Start: at(t, "08:00"),
		Ende:  at(t, "16:00"),
		Zeitspannen: []Zeitspanne{
			{Nr: 1, Start: at(t, "12:00"), Ende: at(t, "12:45")},
		},
	}

	got, err := uc.UpdateArbeitstag(2026, 3, 2, 1, tag)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("UpdateArbeitstag returned %+v; it has always returned nil on success", got)
	}

	for _, tc := range []struct {
		name string
		got  float64
		want float64
	}{
		{"brutto", tag.Brutto, 8},
		{"pausen", tag.Pausen, 0.75},
		{"netto", tag.Netto, 7.25},
		// Soll - Netto, although the comment on the field says Netto - Soll.
		{"differenz", tag.Differenz, 0.75},
	} {
		if tc.got != tc.want {
			t.Errorf("%s = %v, want %v", tc.name, tc.got, tc.want)
		}
	}
}

// A day with no start and no end has no numbers — it is a day not worked, not a
// day of zero hours, and the difference matters to the running total.
func TestEinNichtGearbeiteterTagRechnetNicht(t *testing.T) {
	uc := NewUsecase(&noRepo{})

	tag := &Arbeitstag{Soll: 8}
	if _, err := uc.UpdateArbeitstag(2026, 3, 2, 1, tag); err != nil {
		t.Fatal(err)
	}

	if tag.Brutto != 0 || tag.Netto != 0 || tag.Differenz != 0 {
		t.Errorf("an unworked day computed brutto=%v netto=%v diff=%v, want all zero",
			tag.Brutto, tag.Netto, tag.Differenz)
	}
}

// Saving the spans of a day stores what was sent and removes what was dropped.
//
// Identity is the number, so a span that moved is the same span; one that is no
// longer in the list was deleted by the person editing and has to go.
func TestGeaenderteZeitspannenErsetzenDieGespeicherten(t *testing.T) {
	repo := &noRepo{spannen: []Zeitspanne{{Nr: 1}, {Nr: 2}, {Nr: 3}}}
	uc := NewUsecase(repo)

	// Span 2 is gone, 1 has been moved, 4 is new.
	pausen, err := uc.UpdateZeitspannen(1, time.Now(), []Zeitspanne{
		{Nr: 1, Start: at(t, "12:00"), Ende: at(t, "12:30")},
		{Nr: 3, Start: at(t, "15:00"), Ende: at(t, "15:15")},
		{Nr: 4, Start: at(t, "17:00"), Ende: at(t, "17:15")},
	})
	if err != nil {
		t.Fatal(err)
	}

	if pausen != 1.0 {
		t.Errorf("pausen = %v hours, want 1", pausen)
	}
	if len(repo.saved) != 3 {
		t.Errorf("stored %d spans, want 3", len(repo.saved))
	}
	if len(repo.deleted) != 1 || repo.deleted[0] != 2 {
		t.Errorf("deleted %v, want span 2 and nothing else", repo.deleted)
	}
	// Each stored span carries the length it was worked out to have.
	for _, z := range repo.saved {
		if z.Dauer == 0 {
			t.Errorf("span %d was stored without a duration", z.Nr)
		}
	}
}

// Identity is the number and not the times, which is what the deletion above
// relies on.
func TestIsContained(t *testing.T) {
	list := []Zeitspanne{{Nr: 1}, {Nr: 7}}

	if !IsContained(list, Zeitspanne{Nr: 7, Start: at(t, "09:00")}) {
		t.Error("a span that moved was treated as a different span")
	}
	if IsContained(list, Zeitspanne{Nr: 2}) {
		t.Error("a span that is not in the list was found in it")
	}
}

// A date crosses JSON as a plain day, because that is what it is: no time of
// day in it and no zone to argue about.
func TestDateRoundTrip(t *testing.T) {
	var d Date
	if err := json.Unmarshal([]byte(`"2026-03-02"`), &d); err != nil {
		t.Fatal(err)
	}
	if got := time.Time(d).Format(ctLayout); got != "2026-03-02" {
		t.Errorf("parsed %s, want 2026-03-02", got)
	}

	out, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `"2026-03-02"` {
		t.Errorf("marshalled %s, want %q", out, "2026-03-02")
	}
}

// Setting up a year writes a day for every day of it, and what each day is
// worth follows from the calendar.
//
// This is the one place the two packages meet, and it is where three things
// about the port are visible, so they are pinned here rather than left to be
// discovered:
//
//  1. Every *named* day becomes status F, and kalender names fifty-six days in
//     2026 — among them Valentinstag, Halloween, Black Friday, Muttertag and
//     the start of summer time. None of those is a day off anywhere.
//  2. A day marked F keeps its soll. So the thirty-two named weekdays are
//     flagged as holidays and still demand eight hours, which is the worst of
//     both readings.
//  3. The account and the job are hardcoded to 1 and "IC" in the written day,
//     whatever account was asked for. A second person setting up a year would
//     write over the first one's.
//
// All three are one-line fixes in SetupArbeitsjahr, and all three are decisions
// about what a holiday is and whose year this is — not something to settle
// while moving the code. When they are fixed, this test is what says so.
func TestSetupArbeitsjahrLegtDasGanzeJahrAn(t *testing.T) {
	var tage []Arbeitstag
	repo := &recordingRepo{onSave: func(a Arbeitstag) { tage = append(tage, a) }}

	if _, err := NewUsecase(repo).SetupArbeitsjahr(2026, 7); err != nil {
		t.Fatal(err)
	}

	if len(tage) != 365 {
		t.Fatalf("wrote %d days for 2026, want 365", len(tage))
	}

	byStatus := map[ArbeitstagStatus]int{}
	sollOnHolidays := map[float64]int{}
	for _, a := range tage {
		byStatus[a.Status]++

		datum := time.Time(a.Datum).Format(ctLayout)
		switch a.Status {
		case "A":
			if a.Soll != 8 {
				t.Errorf("%s is a working day with soll %v, want 8", datum, a.Soll)
			}
		case "-":
			if a.Soll != 0 {
				t.Errorf("%s is a weekend day with soll %v, want 0", datum, a.Soll)
			}
		case "F":
			sollOnHolidays[a.Soll]++
		}

		if a.Account != 1 || a.Job != "IC" {
			t.Fatalf("%s was written for account %d job %q; the setup hardcodes 1/IC",
				datum, a.Account, a.Job)
		}
	}

	// 56 named days, 80 weekend days that are not named, 229 ordinary ones.
	for status, want := range map[ArbeitstagStatus]int{"F": 56, "-": 80, "A": 229} {
		if byStatus[status] != want {
			t.Errorf("%d days with status %q, want %d", byStatus[status], status, want)
		}
	}

	// And of the fifty-six, the thirty-two on weekdays still owe eight hours.
	if sollOnHolidays[8] != 32 || sollOnHolidays[0] != 24 {
		t.Errorf("holidays owe: %v; want 32 days owing 8 hours and 24 owing none",
			sollOnHolidays)
	}
}

// recordingRepo watches the days a year setup writes.
type recordingRepo struct {
	noRepo

	onSave func(Arbeitstag)
}

func (r *recordingRepo) SaveArbeitstag(_ int, _ time.Time, _ string, a Arbeitstag) error {
	r.onSave(a)

	return nil
}

// The setup reads the year back after writing it, and answers not-found if it
// cannot — so a stub that stores nothing has to at least admit the year exists.
func (r *recordingRepo) RetrieveArbeitsjahre(_, jahr int) ([]Arbeitsjahr, error) {
	return []Arbeitsjahr{{Jahr: jahr}}, nil
}
