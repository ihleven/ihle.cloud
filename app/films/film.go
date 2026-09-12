// Package films is the film archive: the content type for a digitised reel and
// the endpoints that deliver one.
//
// Delivery is addressed by the entry's id, never by where the bytes are stored,
// so nothing about the storage backend reaches the browser. Addressing an asset
// by its path is a different job and lives in app/media.
package films

import (
	"fmt"
	"strings"

	"github.com/interhome-group/cms/content"
	"github.com/interhome-group/cms/mgmt/search"
)

// Formats a film may have been shot on. Only Super 8 exists today; the field
// is there so the archive can take an 8mm reel or a VHS tape without a schema
// change, and so "Super 8" stops being structural.
const (
	FormatSuper8 = "super8"
)

var formats = map[string]bool{FormatSuper8: true}

// Film is one digitised reel.
//
// Storage is YAML: this is structured archival metadata with a little prose in
// it, not a document.
type Film struct {
	content.EntryContent `type:"Film" mimetype:"application/yaml" json:"-" yaml:"-"`

	Format string `json:"format" yaml:"format"`

	// Key is the object in media storage holding the digitised film. It is a
	// storage key, not a URL: what resolves it is the server's business.
	Key string `json:"key" yaml:"key"`

	// Poster is a still representing the film, as a storage key. Empty means
	// one is rendered from the film itself.
	Poster string `json:"poster,omitempty" yaml:"poster,omitempty"`

	FilmedBy   string `json:"filmed_by,omitempty" yaml:"filmed_by,omitempty"`
	FilmedFrom string `json:"filmed_from,omitempty" yaml:"filmed_from,omitempty"`
	FilmedTo   string `json:"filmed_to,omitempty" yaml:"filmed_to,omitempty"`

	// Location is where it was shot, as free text: an archive of family films
	// has places no gazetteer will resolve.
	Location string `json:"location,omitempty" yaml:"location,omitempty"`

	// Duration in seconds. A restatement of what the file says, kept so a
	// listing can be rendered without reading every film.
	Duration float64 `json:"duration,omitempty" yaml:"duration,omitempty"`

	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Tags        []string `json:"tags,omitempty" yaml:"tags,omitempty"`

	Scenes      []Scene      `json:"scenes,omitempty" yaml:"scenes,omitempty"`
	Annotations []Annotation `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

// Scene is one run of the camera: a named, continuous stretch of the film.
//
// Scenes segment the whole film rather than marking parts of it, which is what
// separates them from annotations — and what makes them a chapter list.
type Scene struct {
	Title string  `json:"title" yaml:"title"`
	Start float64 `json:"start" yaml:"start"`

	// End is optional. A scene runs until the next one begins, or to the end of
	// the film, so stating it is only necessary where a gap is meant.
	End float64 `json:"end,omitempty" yaml:"end,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// Kinds of annotation. The list is open by design: a new kind is a new constant
// and a validation entry, not a new field.
const (
	KindPerson   = "person"
	KindLocation = "location"
	KindObject   = "object"
	KindNote     = "note"
)

var kinds = map[string]bool{KindPerson: true, KindLocation: true, KindObject: true, KindNote: true}

// Annotation marks something visible at a point in the film — most often a
// person, identified by name.
//
// The name is free text rather than a reference to a Person entry. That is a
// deliberate limit: a contributor can write a name that has no entry, or an
// uncertain one, and neither is expressible as a reference. The cost is that
// renaming a person leaves the label behind, which is why labels are indexed.
type Annotation struct {
	Kind  string  `json:"kind" yaml:"kind"`
	Label string  `json:"label" yaml:"label"`
	Start float64 `json:"start" yaml:"start"`
	End   float64 `json:"end,omitempty" yaml:"end,omitempty"`

	// At places the annotation on the frame, so a name can be shown next to the
	// person it names. Optional: an annotation without one is about the film at
	// that moment rather than about a spot in it.
	At *Point `json:"at,omitempty" yaml:"at,omitempty"`

	// Author is the account that added this, by name. Not enforced — an
	// annotation may outlive the account, and may arrive from someone without
	// one.
	Author string `json:"author,omitempty" yaml:"author,omitempty"`
}

// Point is a position on the frame, as fractions of its width and height.
//
// Fractions rather than pixels because the frame is whatever the scan produced
// and the video is displayed at whatever size the page gives it; a pixel
// coordinate would be wrong on both counts, and a fraction drops straight into
// a CSS percentage.
//
// It is a position at a moment. People move, so an annotation whose subject
// crosses the frame wants either a short span or a second annotation.
type Point struct {
	X float64 `json:"x" yaml:"x"`
	Y float64 `json:"y" yaml:"y"`
}

// Clone implements the registry's cloner interface. Slices and the points
// behind them are copied, so a clone can never share memory with the original.
func (f *Film) Clone() interface{} {
	c := *f
	c.Tags = append(f.Tags[:0:0], f.Tags...)
	c.Scenes = append(f.Scenes[:0:0], f.Scenes...)
	c.Annotations = append(f.Annotations[:0:0], f.Annotations...)

	for i, annotation := range c.Annotations {
		if annotation.At != nil {
			at := *annotation.At
			c.Annotations[i].At = &at
		}
	}

	return &c
}

// MediaKey implements the delivery seam: it is where the film's bytes are.
func (f *Film) MediaKey() string { return f.Key }

// PosterKey is the still to show for the film, or "" to render one from the
// film itself.
func (f *Film) PosterKey() string { return f.Poster }

// ValidateContent implements content.ContentValidator, so the CMS runs these
// checks on every write.
func (f *Film) ValidateContent() error {
	if strings.TrimSpace(f.Key) == "" {
		return fmt.Errorf("a film needs a key: where its bytes are stored")
	}
	if !formats[f.Format] {
		return fmt.Errorf("format %q is not one this archive knows", f.Format)
	}
	if err := f.validateScenes(); err != nil {
		return err
	}

	return f.validateAnnotations()
}

func (f *Film) validateScenes() error {
	for i, scene := range f.Scenes {
		switch {
		case strings.TrimSpace(scene.Title) == "":
			return fmt.Errorf("scene %d has no title", i+1)
		case scene.Start < 0:
			return fmt.Errorf("scene %q starts before the film does", scene.Title)
		case scene.End != 0 && scene.End <= scene.Start:
			return fmt.Errorf("scene %q ends before it starts", scene.Title)
		}
		// Scenes segment the film, so they are a sequence rather than a set:
		// each must start strictly after the last. Out of order they cannot be
		// rendered as chapters, and two starting at the same instant leaves no
		// way to say which one a moment belongs to.
		if i > 0 && scene.Start <= f.Scenes[i-1].Start {
			return fmt.Errorf("scene %q does not start after the one before it", scene.Title)
		}
	}

	return nil
}

func (f *Film) validateAnnotations() error {
	for i, annotation := range f.Annotations {
		switch {
		case !kinds[annotation.Kind]:
			return fmt.Errorf("annotation %d has kind %q, which is not one this archive knows", i+1, annotation.Kind)
		case strings.TrimSpace(annotation.Label) == "":
			return fmt.Errorf("annotation %d has no label", i+1)
		case annotation.Start < 0:
			return fmt.Errorf("annotation %q starts before the film does", annotation.Label)
		case annotation.End != 0 && annotation.End <= annotation.Start:
			return fmt.Errorf("annotation %q ends before it starts", annotation.Label)
		}
		// A position outside the frame cannot be drawn, and is far more likely
		// to be pixels written into a field that wants fractions.
		if at := annotation.At; at != nil {
			if at.X < 0 || at.X > 1 || at.Y < 0 || at.Y > 1 {
				return fmt.Errorf("annotation %q sits outside the frame: x and y are fractions between 0 and 1", annotation.Label)
			}
		}
	}

	return nil
}

// AugmentSearchDoc puts the film's own words into the index.
//
// Scene descriptions and annotation labels are the searchable substance of a
// silent film — they are where the names are — so "which films is Anna in" is a
// query rather than a grep.
func (f *Film) AugmentSearchDoc(doc *search.Document, level search.Level) (interface{}, error) {
	doc.DocumentContent.Image = PosterURL(doc.ID)

	if level == search.Extended {
		var text strings.Builder
		for _, scene := range f.Scenes {
			fmt.Fprintln(&text, scene.Title)
			if scene.Description != "" {
				fmt.Fprintln(&text, scene.Description)
			}
		}
		for _, annotation := range f.Annotations {
			fmt.Fprintln(&text, annotation.Label)
		}
		if f.Description != "" {
			fmt.Fprintln(&text, f.Description)
		}
		doc.DocumentContent.Text = map[string]string{"de": text.String()}
	}

	return doc, nil
}
