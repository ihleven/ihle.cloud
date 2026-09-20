package familie

import (
	"os"
	"testing"

	"github.com/interhome-group/cms/content"
)

// The registry is populated by main.go in the running app, so a package test
// has to do it itself — once, for every type this package's tests parse.
func TestMain(m *testing.M) {
	content.Register(Person{})
	content.Register(Reise{})

	os.Exit(m.Run())
}

// TestPersonStorageRoundTrip covers what was broken: the markdown below the
// frontmatter has to survive a write and a read.
//
// It did not. Person carried its own MarshalMarkdown/UnmarshalMarkdown, and the
// CMS stopped calling those when it moved to a generic codec — so the body was
// written but never read back, and every person came out of the API with an
// empty one. The fix is the content.Body field, which the generic codec finds
// by type; this test is here so it cannot quietly stop being found again.
func TestPersonStorageRoundTrip(t *testing.T) {
	body := "# Sepp\n\nEin Absatz, und noch einer.\n"
	orig := &content.Entry{
		Meta: content.Meta{
			Path: "familie/josef.md", Type: "Person", MIME: "text/markdown",
			ID: "josef", Name: "Josef", Slug: "josef",
		},
		Content: &Person{
			Key:        "josef",
			Geburtstag: "1901-02-03",
			Todestag:   "1975-06-07",
			Vater:      "anton",
			Mutter:     "maria",
			Body:       content.NewBody(body),
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

	p, ok := got.Content.(*Person)
	if !ok {
		t.Fatalf("content came back as %T, want *Person", got.Content)
	}
	if p.Body.String() != body {
		t.Fatalf("body did not round-trip:\n got %q\nwant %q", p.Body.String(), body)
	}
	if p.Key != "josef" || p.Geburtstag != "1901-02-03" || p.Vater != "anton" || p.Mutter != "maria" {
		t.Fatalf("typed fields did not round-trip: %+v", p)
	}
}
