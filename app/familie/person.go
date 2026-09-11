package familie

import (
	"fmt"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/goccy/go-yaml"
	"github.com/interhome-group/cms/content"
	"github.com/interhome-group/cms/mgmt/search"
)

// Frontmatter is the YAML envelope of a markdown entry: the CMS metadata
// inline, plus the content type's own fields under `content`.
//
// This used to be content.Frontmatter. The CMS no longer has a per-content-type
// markdown codec — frontmatter is now decoded typed in a single pass — so the
// MarshalMarkdown/UnmarshalMarkdown methods below are no longer called by
// anything in the CMS. The type is kept here so the code still compiles and
// still says what it meant; the methods are dead and could be deleted.
type Frontmatter struct {
	Meta    content.Meta   `yaml:",inline"`
	Content map[string]any `yaml:"content"`
}

type Person struct {
	content.EntryContent `type:"Person" json:"-" yaml:"-" mimetype:"text/markdown"`

	Key        string `json:"key"`
	Geburtstag string `json:"geburtstag"`
	Todestag   string `json:"todestag"`
	Vater      string `json:"vater"`
	Mutter     string `json:"mutter"`

	Content string `json:"markdown" yaml:"-"`
}

func (p *Person) MarshalMarkdown(meta content.Meta) ([]byte, error) {

	type matter struct {
		Meta    content.Meta `yaml:",inline"`
		Content *Person      `yaml:"content"`
	}

	matterbytes, err := yaml.Marshal(matter{Meta: meta, Content: p})
	if err != nil {
		return nil, err
	}

	return []byte(fmt.Sprintf("---\n%s---\n%s\n", matterbytes, p.Content)), nil
}

func (p *Person) UnmarshalMarkdown(data []byte, matter *Frontmatter) error {

	bytes, err := yaml.Marshal(matter.Content)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(bytes, p)
	if err != nil {
		return err
	}

	p.Content = string(data)

	return nil

	// if t, ok := matter.Content["geburtstag"].(string); ok {
	// 	// p.Geburtstag, _ = time.Parse(time.RFC3339, t)
	// 	p.Geburtstag = t
	// }
	// if t, ok := matter.Content["todestag"].(string); ok {

	// 	p.Todestag, _ = time.Parse("2006-01-02", t)
	// }
	// if vater, ok := matter.Content["vater"].(string); ok {
	// 	p.Vater = vater
	// }
	// if mutter, ok := matter.Content["mutter"].(string); ok {

	// 	p.Mutter = mutter
	// }
	// p.Content = string(data)
	// // fmt.Printf("unmarschal Person: %+v\n", p)
	// return nil
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
	d.Text = map[string]string{"de": p.Content}
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

	person.AddFieldMappingsAt("content", text)
	person.AddFieldMappingsAt("geburtstag", date)
	person.AddFieldMappingsAt("todestag", date)
	person.AddFieldMappingsAt("vater", keyword)
	person.AddFieldMappingsAt("mutter", keyword)

	entrymap.AddSubDocumentMapping("person", person)
}
