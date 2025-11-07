package super8

import (
	"strings"
	"time"

	"bitbucket.org/hotelplan/webcc-content/cms/content"
	"bitbucket.org/hotelplan/webcc-content/cms/mgmt/search"
)

type Super8 struct {
	content.ContentType `type:"Super8" json:"-" yaml:"-" mimetype:"application/yaml"`

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

func (s *Super8) AugmentSearchDoc(doc *search.Document, level search.Level) (interface{}, error) {
	doc.DocumentContent.Image = s.Thumbnail

	if doc.DocumentContent.Image == "" {
		doc.DocumentContent.Image = "http://localhost:8000/hi/media/thumbs?path=" + strings.TrimPrefix(s.Source, "http://localhost:8000/hi/media/proxy/")
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
