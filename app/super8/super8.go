package super8

import (
	"strings"
	"time"

	"github.com/interhome-group/cms/content"
	"github.com/interhome-group/cms/mgmt/search"
)

type Super8 struct {
	content.EntryContent `type:"Super8" json:"-" yaml:"-" mimetype:"application/yaml"`

	Source    string `yaml:"src"        json:"src"`
	Thumbnail string `yaml:"thumbnail"  json:"thumbnail"`
	Year      int    `yaml:"year"       json:"year"`
	From      string `yaml:"von"        json:"von"`
	To        string `yaml:"bis"        json:"bis"`
	VTT       string `yaml:"vtt"        json:"vtt"`
	Captions  []Cue
	// Token       struct{} `yaml:"-"        json:"token"`
}

type Cue struct {
	ID      string      `yaml:"id"`
	Timings []time.Time `yaml:"ts"`
	Text    string      `yaml:"txt"`
}

func (s *Super8) Clone() interface{} {

	copy := *s
	return &copy
}

// mediaPath reduces a source to the path the media routes take.
//
// Content written before the URLs were made relative still carries an absolute
// host, so both forms are accepted rather than only the current one.
func mediaPath(source string) string {
	if i := strings.Index(source, mediaProxyPrefix); i >= 0 {
		return source[i+len(mediaProxyPrefix):]
	}
	return strings.TrimPrefix(source, "/")
}

const mediaProxyPrefix = "/hi/media/proxy/"

func (s *Super8) AugmentSearchDoc(doc *search.Document, level search.Level) (interface{}, error) {
	doc.DocumentContent.Image = s.Thumbnail

	if doc.DocumentContent.Image == "" {
		// A path rather than a URL: this value is written into the search index
		// and served to the browser, so an absolute host would outlive the
		// deployment that produced it. Relative, it follows whichever domain
		// serves the app.
		doc.DocumentContent.Image = "/hi/media/thumbs?path=" + mediaPath(s.Source)
	}

	if level == search.Extended {
		doc.DocumentContent.Text = map[string]string{"de": s.VTT}
	}

	return doc, nil
}

// func AdaptMapping(entrymap *mapping.DocumentMapping) {

// 	keyword := bleve.NewKeywordFieldMapping()
// 	text := bleve.NewTextFieldMapping()
// 	date := bleve.NewDateTimeFieldMapping()
// 	// date.DateFormat = sanitized.Name

// 	person := bleve.NewDocumentMapping()
// 	person.Dynamic = false
// 	person.AddFieldMappingsAt("name", text)

// 	person.AddFieldMappingsAt("content", text)
// 	person.AddFieldMappingsAt("geburtstag", date)
// 	person.AddFieldMappingsAt("todestag", date)
// 	person.AddFieldMappingsAt("vater", keyword)
// 	person.AddFieldMappingsAt("mutter", keyword)

// 	entrymap.AddSubDocumentMapping("person", person)
// }
