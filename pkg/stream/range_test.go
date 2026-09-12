package stream

import (
	"errors"
	"testing"
)

func TestParseRange(t *testing.T) {
	const size = 1000

	tests := []struct {
		name   string
		header string
		size   int64
		want   Range
		ranged bool
		err    error
	}{
		// No usable range -> serve the whole object.
		{name: "absent", header: "", size: size},
		{name: "malformed unit", header: "items=0-10", size: size},
		{name: "no dash", header: "bytes=100", size: size},
		{name: "empty spec", header: "bytes=", size: size},
		{name: "non-numeric start", header: "bytes=abc-10", size: size},
		{name: "non-numeric end", header: "bytes=0-abc", size: size},
		{name: "negative start", header: "bytes=--5", size: size},
		// Multi-range is legal HTTP but answered in full rather than multipart.
		{name: "multi range ignored", header: "bytes=0-99,200-299", size: size},

		// Start-only: what a player sends to begin playback and after a seek.
		{name: "start only", header: "bytes=0-", size: size, want: Range{0, 1000}, ranged: true},
		{name: "start only mid", header: "bytes=500-", size: size, want: Range{500, 500}, ranged: true},
		{name: "last byte", header: "bytes=999-", size: size, want: Range{999, 1}, ranged: true},

		// Closed ranges.
		{name: "closed", header: "bytes=0-99", size: size, want: Range{0, 100}, ranged: true},
		{name: "closed mid", header: "bytes=200-299", size: size, want: Range{200, 100}, ranged: true},
		{name: "single byte", header: "bytes=0-0", size: size, want: Range{0, 1}, ranged: true},
		{name: "whole object", header: "bytes=0-999", size: size, want: Range{0, 1000}, ranged: true},
		{name: "with spaces", header: " bytes= 10 - 19 ", size: size, want: Range{10, 10}, ranged: true},

		// Suffix: how players read a trailing container index.
		{name: "suffix", header: "bytes=-100", size: size, want: Range{900, 100}, ranged: true},
		{name: "suffix whole", header: "bytes=-1000", size: size, want: Range{0, 1000}, ranged: true},
		{name: "suffix oversized clamps", header: "bytes=-5000", size: size, want: Range{0, 1000}, ranged: true},

		// Past EOF is clamped, not refused: players probe beyond the end.
		{name: "past eof clamped", header: "bytes=900-5000", size: size, want: Range{900, 100}, ranged: true},
		{name: "past eof from zero", header: "bytes=0-5000", size: size, want: Range{0, 1000}, ranged: true},

		// Unsatisfiable -> 416.
		{name: "start at size", header: "bytes=1000-", size: size, err: ErrUnsatisfiable},
		{name: "start beyond size", header: "bytes=2000-2999", size: size, err: ErrUnsatisfiable},
		{name: "end before start", header: "bytes=500-100", size: size, err: ErrUnsatisfiable},
		{name: "zero-length suffix", header: "bytes=-0", size: size, err: ErrUnsatisfiable},

		// Empty object: every range is out of bounds, and no range is fine.
		{name: "zero size no range", header: "", size: 0},
		{name: "zero size start only", header: "bytes=0-", size: 0, err: ErrUnsatisfiable},
		{name: "zero size closed", header: "bytes=0-99", size: 0, err: ErrUnsatisfiable},
		{name: "zero size suffix", header: "bytes=-10", size: 0, err: ErrUnsatisfiable},

		// Single-byte object, where off-by-one errors surface immediately.
		{name: "one byte object", header: "bytes=0-", size: 1, want: Range{0, 1}, ranged: true},
		{name: "one byte suffix", header: "bytes=-1", size: 1, want: Range{0, 1}, ranged: true},
		{name: "one byte past eof", header: "bytes=0-99", size: 1, want: Range{0, 1}, ranged: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ranged, err := ParseRange(tc.header, tc.size)

			if tc.err != nil {
				if !errors.Is(err, tc.err) {
					t.Fatalf("ParseRange(%q, %d) err = %v, want %v", tc.header, tc.size, err, tc.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseRange(%q, %d) unexpected err = %v", tc.header, tc.size, err)
			}
			if ranged != tc.ranged {
				t.Fatalf("ParseRange(%q, %d) ranged = %v, want %v", tc.header, tc.size, ranged, tc.ranged)
			}
			if ranged && got != tc.want {
				t.Fatalf("ParseRange(%q, %d) = %+v, want %+v", tc.header, tc.size, got, tc.want)
			}
		})
	}
}

// A resolved range must always stay inside the object: this is the invariant
// that keeps Content-Range honest and stops the store being asked to read past
// the end.
func TestParseRangeStaysInBounds(t *testing.T) {
	sizes := []int64{0, 1, 2, 999, 1000}
	headers := []string{
		"", "bytes=0-", "bytes=1-", "bytes=999-", "bytes=1000-", "bytes=0-0",
		"bytes=0-999", "bytes=500-100", "bytes=-1", "bytes=-1000", "bytes=-5000",
		"bytes=998-5000", "garbage", "bytes=x-y",
	}
	for _, size := range sizes {
		for _, h := range headers {
			r, ranged, err := ParseRange(h, size)
			if err != nil || !ranged {
				continue
			}
			if r.Off < 0 || r.Len <= 0 || r.Off+r.Len > size {
				t.Errorf("ParseRange(%q, %d) = %+v escapes [0,%d)", h, size, r, size)
			}
		}
	}
}

func TestRangeLast(t *testing.T) {
	if got := (Range{Off: 0, Len: 1}).Last(); got != 0 {
		t.Errorf("Last() = %d, want 0", got)
	}
	if got := (Range{Off: 900, Len: 100}).Last(); got != 999 {
		t.Errorf("Last() = %d, want 999", got)
	}
}
