package films

import (
	"strings"
	"testing"

	"github.com/interhome-group/cms/mgmt/search"
)

func valid() *Film {
	return &Film{
		Format: FormatSuper8,
		Key:    "public/Super 8/1975__Weihnachten.mp4",
		Scenes: []Scene{
			{Title: "Wohnzimmer", Start: 0, End: 11.8, Description: "Ein Kameraschwenk"},
			{Title: "Bescherung", Start: 11.82, End: 62},
		},
		Annotations: []Annotation{
			{Kind: KindPerson, Label: "Oma Anna", Start: 63, End: 70, At: &Point{X: 0.4, Y: 0.6}},
		},
	}
}

func TestValidateContentAcceptsAFilm(t *testing.T) {
	if err := valid().ValidateContent(); err != nil {
		t.Fatalf("a well-formed film was refused: %v", err)
	}
}

// Each of these would produce something that cannot be rendered, and would do
// so silently — a chapter list out of order, a name drawn off-screen — so the
// write is where they have to be caught.
func TestValidateContentRefuses(t *testing.T) {
	tests := []struct {
		name string
		mut  func(*Film)
		want string
	}{
		{"no key", func(f *Film) { f.Key = " " }, "needs a key"},
		{"an unknown format", func(f *Film) { f.Format = "betamax" }, "not one this archive knows"},
		{"a scene with no title", func(f *Film) { f.Scenes[0].Title = "" }, "no title"},
		{"a scene ending before it starts", func(f *Film) { f.Scenes[0].End = 0.5; f.Scenes[0].Start = 9 }, "ends before it starts"},
		{"two scenes starting together", func(f *Film) { f.Scenes[1].Start = 0 }, "does not start after"},
		{"scenes out of order", func(f *Film) { f.Scenes[1].Start = -0.5; f.Scenes[1].End = 0 }, "before the film does"},
		{"an unknown annotation kind", func(f *Film) { f.Annotations[0].Kind = "vibe" }, "not one this archive knows"},
		{"an annotation with no label", func(f *Film) { f.Annotations[0].Label = "  " }, "no label"},
		{"a negative start", func(f *Film) { f.Annotations[0].Start = -1 }, "before the film does"},
		// The likeliest cause of this is pixels written into a field that wants
		// fractions, which would place every annotation far off the frame.
		{"a position in pixels", func(f *Film) { f.Annotations[0].At = &Point{X: 480, Y: 270} }, "outside the frame"},
		{"a negative position", func(f *Film) { f.Annotations[0].At = &Point{X: -0.1, Y: 0.5} }, "outside the frame"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := valid()
			tt.mut(f)

			err := f.ValidateContent()
			if err == nil {
				t.Fatal("accepted")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("err = %q, want it to mention %q", err, tt.want)
			}
		})
	}
}

// A scene with no end is the normal case: it runs until the next one starts.
func TestValidateContentAllowsAnOpenEnd(t *testing.T) {
	f := valid()
	f.Scenes[0].End = 0
	f.Annotations[0].End = 0

	if err := f.ValidateContent(); err != nil {
		t.Errorf("an open-ended scene or annotation was refused: %v", err)
	}
}

// The frame's top-left corner is a position like any other, which is why At is
// a pointer: a zero value has to be distinguishable from an absent one.
func TestValidateContentAllowsTheOrigin(t *testing.T) {
	f := valid()
	f.Annotations[0].At = &Point{X: 0, Y: 0}

	if err := f.ValidateContent(); err != nil {
		t.Errorf("the top-left corner was refused: %v", err)
	}
}

// The registry hands clones out to callers who may edit them. Sharing a slice's
// backing array — or a point behind it — would let one edit reach the original.
func TestCloneSharesNothing(t *testing.T) {
	original := valid()
	original.Tags = []string{"weihnachten"}

	clone := original.Clone().(*Film)
	clone.Scenes[0].Title = "changed"
	clone.Annotations[0].Label = "changed"
	clone.Annotations[0].At.X = 0.99
	clone.Tags[0] = "changed"

	switch {
	case original.Scenes[0].Title != "Wohnzimmer":
		t.Error("editing a clone's scene changed the original")
	case original.Annotations[0].Label != "Oma Anna":
		t.Error("editing a clone's annotation changed the original")
	case original.Annotations[0].At.X != 0.4:
		t.Error("editing a clone's position changed the original")
	case original.Tags[0] != "weihnachten":
		t.Error("editing a clone's tags changed the original")
	}
}

// The names in a silent film are in its annotations, so they have to reach the
// index or "which films is Anna in" is a grep.
func TestAugmentSearchDocIndexesTheNames(t *testing.T) {
	doc := &search.Document{}
	doc.ID = "heiligabend75"

	if _, err := valid().AugmentSearchDoc(doc, search.Extended); err != nil {
		t.Fatalf("AugmentSearchDoc: %v", err)
	}

	text := doc.DocumentContent.Text["de"]
	for _, want := range []string{"Oma Anna", "Wohnzimmer", "Ein Kameraschwenk"} {
		if !strings.Contains(text, want) {
			t.Errorf("indexed text is missing %q; got %q", want, text)
		}
	}
	if doc.DocumentContent.Image != "/api/v1/films/heiligabend75/poster.jpg" {
		t.Errorf("Image = %q", doc.DocumentContent.Image)
	}
}

// Below Extended the body is not indexed, and augmenting it anyway would put
// the whole archive's prose in an index that was asked not to hold it.
func TestAugmentSearchDocLeavesTextAloneBelowExtended(t *testing.T) {
	doc := &search.Document{}
	doc.ID = "heiligabend75"

	if _, err := valid().AugmentSearchDoc(doc, search.Fulltext); err != nil {
		t.Fatalf("AugmentSearchDoc: %v", err)
	}
	if doc.DocumentContent.Text != nil {
		t.Errorf("Text = %v, want nothing indexed below Extended", doc.DocumentContent.Text)
	}
}

func TestPosterURL(t *testing.T) {
	for id, want := range map[string]string{
		"heiligabend75": "/api/v1/films/heiligabend75/poster.jpg",
		"a film":        "/api/v1/films/a%20film/poster.jpg",
		"":              "",
	} {
		if got := PosterURL(id); got != want {
			t.Errorf("PosterURL(%q) = %q, want %q", id, got, want)
		}
	}
}
