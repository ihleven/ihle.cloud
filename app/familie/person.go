package familie

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/interhome-group/cms/content"
	"github.com/interhome-group/cms/mgmt/search"
)

type Person struct {
	content.EntryContent `type:"Person" json:"-" yaml:"-" mimetype:"text/markdown"`

	Key        string `json:"key"`
	Geburtstag string `json:"geburtstag"`
	Todestag   string `json:"todestag"`
	Vater      string `json:"vater"`
	Mutter     string `json:"mutter"`

	// The markdown below the frontmatter. Declaring it is the whole opt-in:
	// the CMS's generic markdown codec finds a content.Body by type and reads
	// the body into it, which is why this type no longer carries a codec of its
	// own. It used to, and the methods stopped being called when the CMS
	// dropped per-content-type marshalling — which is why entries came back
	// with an empty body.
	Body content.Body `json:"body" yaml:"-"`
}

func (p *Person) Clone() interface{} {

	copy := *p

	return &copy
}

func (p *Person) AugmentSearchDoc(doc *search.Document, level search.Level) (interface{}, error) {

	d := struct {
		search.Document
		Person *Person `json:"person"`
		// Geburtstag time.Time `json:"geburtstag"`
		// Todestag   time.Time `json:"todestag"`
		// Vater      string    `json:"vater"`
		// Mutter     string    `json:"mutter"`
		// Content    string    `json:"content"`
	}{
		Document: *doc,
		Person:   p,
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

func AdaptMapping(entrymap *mapping.DocumentMapping) {

	keyword := bleve.NewKeywordFieldMapping()
	text := bleve.NewTextFieldMapping()
	date := bleve.NewDateTimeFieldMapping()
	// date.DateFormat = sanitized.Name

	person := bleve.NewDocumentMapping()
	person.Dynamic = false
	person.AddFieldMappingsAt("name", text)

	person.AddFieldMappingsAt("body", text)
	person.AddFieldMappingsAt("geburtstag", date)
	person.AddFieldMappingsAt("todestag", date)
	person.AddFieldMappingsAt("vater", keyword)
	person.AddFieldMappingsAt("mutter", keyword)

	entrymap.AddSubDocumentMapping("person", person)
}
