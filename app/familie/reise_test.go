package familie

import (
	"testing"
	"time"

	"github.com/interhome-group/cms/content"
)

// TestReiseStorageRoundTrip is TestPersonStorageRoundTrip for a journey: the
// diary is the body, so losing it loses the entry's whole point.
func TestReiseStorageRoundTrip(t *testing.T) {
	body := "# Reisetagebuch\n\n## 04.07. Orleans\n\n::ausgaben\n42 €\n::\n"
	orig := &content.Entry{
		Meta: content.Meta{
			Path: "urlaub/2025-bretagne.md", Type: "Reise", MIME: "text/markdown",
			ID: "2025-bretagne", Name: "Bretagne 2025", Slug: "2025-bretagne",
		},
		Content: &Reise{
			Ziel: "bretagne",
			Jahr: 2025,
			Von:  "2025-07-04",
			Bis:  time.Date(2025, 7, 12, 0, 0, 0, 0, time.UTC),
			Body: content.NewBody(body),
		},
	}

	data, err := orig.SerializeForStorage()
	if err != nil {
		t.Fatalf("SerializeForStorage: %v", err)
	}

	var got content.Entry
	if err := content.ParseMarkdownEntryFromStorage(data, &got); err != nil {
		t.Fatalf("ParseMarkdownEntryFromStorage: %v\n--- produced ---\n%s", err, data)
	}

	r, ok := got.Content.(*Reise)
	if !ok {
		t.Fatalf("content came back as %T, want *Reise", got.Content)
	}
	if r.Body.String() != body {
		t.Fatalf("body did not round-trip:\n got %q\nwant %q", r.Body.String(), body)
	}
	if r.Ziel != "bretagne" || r.Jahr != 2025 || r.Von != "2025-07-04" {
		t.Fatalf("typed fields did not round-trip: %+v", r)
	}
	if !r.Bis.Equal(orig.Content.(*Reise).Bis) {
		t.Fatalf("bis did not round-trip: got %v", r.Bis)
	}
}
