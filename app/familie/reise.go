package familie

import (
	"fmt"
	"time"

	"bitbucket.org/hotelplan/webcc-content/cms/content"
	"bitbucket.org/hotelplan/webcc-content/cms/mgmt/search"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"gopkg.in/yaml.v2"
)

type Reise struct {
	content.ContentType `type:"Reise" json:"-" yaml:"-" mimetype:"text/markdown"`

	Ziel string    `json:"ziel"`
	Jahr int       `json:"jahr"`
	Von  string    `json:"von"`
	Bis  time.Time `json:"bis"`

	Content string `json:"markdown" yaml:"-"`
}

func (p *Reise) MarshalMarkdown(meta content.Meta) ([]byte, error) {

	type matter struct {
		Meta    content.Meta `yaml:",inline"`
		Content *Reise       `yaml:"content"`
	}

	matterbytes, err := yaml.Marshal(matter{Meta: meta, Content: p})
	if err != nil {
		return nil, err
	}

	return []byte(fmt.Sprintf("---\n%s---\n%s\n", matterbytes, p.Content)), nil
}

func (p *Reise) UnmarshalMarkdown(data []byte, matter *content.Frontmatter) error {

	if t, ok := matter.Content["von"].(string); ok {
		// p.Geburtstag, _ = time.Parse(time.RFC3339, t)
		p.Von = t
	}
	if t, ok := matter.Content["bis"].(string); ok {

		p.Bis, _ = time.Parse("2006-01-02", t)
	}
	if vater, ok := matter.Content["ziel"].(string); ok {
		p.Ziel = vater
	}
	if mutter, ok := matter.Content["jahr"].(int); ok {

		p.Jahr = mutter
	}
	p.Content = string(data)
	// fmt.Printf("unmarschal Person: %+v\n", p)
	return nil
}

func (p *Reise) Clone() interface{} {

	copy := *p

	return &copy
}

func (p *Reise) AugmentSearchDoc(doc *search.Document, level search.Level) (interface{}, error) {

	d := struct {
		search.Document
		Reise *Reise `json:"person"`
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
	d.Text = map[string]string{"de": p.Content}
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

	reise.AddFieldMappingsAt("content", text)
	reise.AddFieldMappingsAt("von", date)
	reise.AddFieldMappingsAt("bis", date)
	reise.AddFieldMappingsAt("ziel", keyword)
	reise.AddFieldMappingsAt("jahr", keyword)

	entrymap.AddSubDocumentMapping("reise", reise)
}
