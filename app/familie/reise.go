package familie

import (
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/interhome-group/cms/content"
	"github.com/interhome-group/cms/mgmt/search"
)

type Reise struct {
	content.EntryContent `type:"Reise" json:"-" yaml:"-" mimetype:"text/markdown"`

	Ziel string    `json:"ziel"`
	Jahr int       `json:"jahr"`
	Von  string    `json:"von"`
	Bis  time.Time `json:"bis"`

	// See Person.Body: declaring it is what makes the generic markdown codec
	// fill it.
	Body content.Body `json:"body" yaml:"-"`
}

func (p *Reise) Clone() interface{} {

	copy := *p

	return &copy
}

func (p *Reise) AugmentSearchDoc(doc *search.Document, level search.Level) (interface{}, error) {

	d := struct {
		search.Document
		Reise *Reise `json:"reise"`
		// Geburtstag time.Time `json:"geburtstag"`
		// Todestag   time.Time `json:"todestag"`
		// Vater      string    `json:"vater"`
		// Mutter     string    `json:"mutter"`
		// Content    string    `json:"content"`
	}{
		Document: *doc,
		Reise:    p,
		// Geburtstag: p.Geburtstag,
		// Todestag:   p.Todestag,
		// Vater:      p.Vater,
		// Mutter:     p.Mutter,
		// Content:    p.Content,
	}
	d.Text = map[string]string{"de": p.Body.String()}
	// fmt.Printf("AUGMENT: %+v\n", d)
	return d, nil
}

func AdaptMappingReise(entrymap *mapping.DocumentMapping) {

	keyword := bleve.NewKeywordFieldMapping()
	text := bleve.NewTextFieldMapping()
	date := bleve.NewDateTimeFieldMapping()
	// date.DateFormat = sanitized.Name

	reise := bleve.NewDocumentMapping()
	reise.Dynamic = false
	reise.AddFieldMappingsAt("name", text)

	reise.AddFieldMappingsAt("body", text)
	reise.AddFieldMappingsAt("von", date)
	reise.AddFieldMappingsAt("bis", date)
	reise.AddFieldMappingsAt("ziel", keyword)
	reise.AddFieldMappingsAt("jahr", keyword)

	entrymap.AddSubDocumentMapping("reise", reise)
}
