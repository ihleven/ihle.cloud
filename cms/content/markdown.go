package content

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func init() {
	Register(Article{})
	Register(MDPage{})
}

type markdowner interface {
	MarshalMarkdown(meta Meta) ([]byte, error)
	UnmarshalMarkdown([]byte, *Frontmatter) error
}

type Frontmatter struct {
	Meta    Meta                   `yaml:",inline"`
	Content map[string]interface{} `yaml:"content"`
}

////////////////////////////////////////////////////////////////////
////////// Article /////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////

// MarkdownPage
type Article struct {
	ContentType `type:"MDArticle" json:"-" yaml:"-" mimetype:"text/markdown"`
	ID          string `yaml:"-"`
	Image       string `yaml:"-"`
	// Title       string `yaml:"-"`
	Destination string `yaml:"-"`

	Canonical   string            `json:"canonical"`
	Alternates  map[string]string `json:"alternates,omitempty"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Robots      string            `json:"robots"`

	Author string    `json:"author"`
	Date   time.Time `json:"date"`

	// OGTitle       string `json:"og:title"`
	// OGDescription string `json:"og:description"`
	// OGImage       string `json:"og:image"`
	// OGURL         string `json:"og:url"`
	// OGType        string `json:"og:type"`

	Matter map[string]interface{} `json:"matter" yaml:"-"`

	Markdown []string `json:"markdown"`
}

func (a *Article) MarshalMarkdown(meta Meta) ([]byte, error) {
	// matter := struct {
	// 	Meta        Meta
	// 	Title       string
	// 	Destination string
	// }{
	// 	Meta:        meta,
	// 	Title:       a.Title,
	// 	Destination: a.Destination,
	// }
	m, err := yaml.Marshal(&meta)
	if err != nil {
		fmt.Println("Article", m)
		return nil, err //log.Fatalf("error: %v", err)
	}

	md, err := a.MarshalText()
	if err != nil {
		return nil, err //log.Fatalf("error: %v", err)
	}
	response := [][]byte{
		[]byte("---\n"), m, []byte("---\n"), md,
	}
	// for _, line := range a.Markdown {
	// 	response = append(response, []byte(line))
	// }

	return bytes.Join(response, nil), nil
}

func (a *Article) MarshalText() ([]byte, error) {
	content := strings.Join(a.Markdown, "\n")
	return []byte(content), nil
}

func (a *Article) MarshalJSON() ([]byte, error) {
	return json.MarshalIndent(&struct {
		Matter   map[string]interface{} `json:"matter"`
		Markdown []string               `json:"markdown"`
	}{
		Matter:   a.Matter,
		Markdown: a.Markdown,
	}, "", "    ")

}

func (a *Article) UnmarshalMarkdown(data []byte, matter *Frontmatter) error {
	a.Markdown = strings.Split(string(data), "\n")
	// fmt.Printf("+++++++++++++++++%+v => %v\n", matter, a.Markdown)
	a.Matter = matter.Content
	// a.Title = matter.Fields["title"].(string)
	// a.Destination = matter.Fields["destination"].(string)
	return nil
}

func (a *Article) Clone() interface{} {

	copy := *a

	copy.Matter = map[string]interface{}{}
	for k, v := range a.Matter {
		copy.Matter[k] = v
	}

	copy.Markdown = make([]string, len(a.Markdown))
	for i, line := range a.Markdown {
		copy.Markdown[i] = line
	}

	return &copy
}

////////////////////////////////////////////////////////////////////
////////// MDPage //////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////

// MarkdownPage is a Page that ist stored in md format instead of json.
// This has two implications:
// 1) Meta data is stored in front matter
// 2) Content is stored as Markdown with embedded components.
//
// content is either delivered as is (markdown with Accept header `text/markdown`) or
// or json (with Accept header `application/json`, default)
type MDPage struct {
	ContentType `type:"MDPage" json:"-" mimetype:"text/markdown"`

	Matter   map[string]interface{} `json:"matter"`
	Markdown Markdown               `json:"markdown"`
	Lines    []string               `json:"lines"` // mit Leerzeile getrennt oder jede einzelne Zeile? -> Lines oder Paragraphs?
	Blocks   Bloks                  `json:"blocks"`
}

func (p *MDPage) MarshalText() ([]byte, error) {
	return p.Markdown.MarshalText()
}

func (a *MDPage) MarshalJSON() ([]byte, error) {
	fmt.Println(" +++++++++++ MarshalJSON ++++++++++++++")
	fmt.Println(" +++++++++++ MarshalJSON ++++++++++++++", a.Markdown.content)
	return json.MarshalIndent(&struct {
		Matter   map[string]interface{} `json:"matter"`
		Markdown Markdown               `json:"markdown"`
		Lines    []string               `json:"lines"`
	}{
		Matter:   a.Matter,
		Markdown: a.Markdown,
		Lines:    a.Lines,
	}, "", "    ")

}

func (p *MDPage) Clone() interface{} {
	clone := MDPage{Matter: make(map[string]interface{}), Markdown: p.Markdown, Lines: append(p.Lines[:0:0], p.Lines...), Blocks: p.Blocks.Clone()}
	for k, v := range p.Matter {
		clone.Matter[k] = v
	}
	return &clone
}

func (p *MDPage) GetMarkdown() string {
	b, _ := p.Markdown.MarshalText()
	return string(b)
}
func (a *MDPage) MarshalMarkdown(meta Meta) ([]byte, error) {
	return nil, nil
}

func (p *MDPage) UnmarshalMarkdown(data []byte, matter *Frontmatter) error {
	// fmt.Println(" +++++++++++ UnmarshalMarkdown ++++++++++++++", string(data))
	// rest, err := frontmatter.Parse(bytes.NewReader(data), &e.Meta)
	// if err != nil {
	// 	return err
	// }

	// page := MDPage{Markdown: Markdown{content: string(rest)}}
	// e.Content = &page

	// fmt.Printf("%+v\n", e.Meta)
	// fmt.Println(string(rest))
	p.Matter = matter.Content
	p.Markdown = Markdown{content: string(data)}

	fileScanner := bufio.NewScanner(bytes.NewReader(data))
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		p.Lines = append(p.Lines, fileScanner.Text())
	}

	return nil
}
