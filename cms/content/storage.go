package content

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"bitbucket.org/hotelplan/webcc-content/cms/pkg/errors"
	"github.com/adrg/frontmatter"
)

// SerializeForStorage is used before saving an entry to the storage.
// It fills in missing meta data
// Depending on the mime type it serializes to either json or markdown
// TODO: normalize refs, links and assets
func (e *Entry) SerializeForStorage() ([]byte, error) {

	if e == nil {
		return nil, errors.New("entry is nil")
	}

	ct := e.ContentType()

	e.Meta.Type = ct.Name
	e.Meta.MIME = ct.MIME // "application/json"
	e.Meta.Modified = time.Now()
	if e.Meta.Tags == nil {
		e.Meta.Tags = []string{}
	}

	if ct.MIME == "text/markdown" {
		return e.MarshalToMarkdown() // SerializeMD(e)
	}
	bytes, err := json.MarshalIndent(e, "", "    ")
	return bytes, err
}

func (e *Entry) MarshalToMarkdown() (text []byte, err error) {

	// matter := Frontmatter{
	// 	Additional: make(map[string]interface{}),
	// }

	// delete(matter.Additional, "foo")

	// m, err := yaml.Marshal(e.Meta)
	// if err != nil {
	// 	return nil, err //log.Fatalf("error: %v", err)
	// }
	marshaler, ok := e.Content.(interface {
		MarshalMarkdown(meta Meta) ([]byte, error)
	})
	if !ok {
		return nil, errors.New("MarshalMarkdown iface not implemented")
	}
	md, err := marshaler.MarshalMarkdown(e.Meta)
	if err != nil {
		return nil, err //log.Fatalf("error: %v", err)
	}
	return md, err //bytes.Join([][]byte{[]byte("---\n"), m, []byte("---\n"), c}, nil), nil
}

// func SerializeMD(e *Entry) ([]byte, error) {

// 	var markdown strings.Builder
// 	markdown.WriteString("---\n")

// 	k := []string{"id", "name", "path", "full_slug", "slug"}
// 	for i, v := range []string{e.ID, e.Name, e.Path, e.FullSlug, e.Slug} {
// 		markdown.WriteString(k[i] + ": " + v + "\n")
// 	}
// 	markdown.WriteString("---\n\n")

// 	if markdowner, ok := e.Content.(interface{ GetMarkdown() string }); ok {
// 		markdown.WriteString(markdowner.GetMarkdown())
// 	} else {
// 		return nil, errors.New("entry does not support markdowner interface")
// 	}
// 	markdown.WriteString("\n")

// 	return []byte(markdown.String()), nil
// }

// ParseEntry converts given `bytes` to an Entry and writes the result to where entry points to.
// If `entry` is nil memory is allocated and the pointer is adapted.
// contentType specifies the format, either json or markdown with json as default.
func ParseEntry(entry *Entry, bytes []byte, contentType string) (*Entry, error) {

	if entry == nil {
		var e Entry
		entry = &e
	}

	if strings.HasPrefix(contentType, "text/markdown") {
		err := ParseMarkdownEntryFromStorage(bytes, entry)
		if err != nil {
			return nil, errors.Wrap(err, "Couldn't parse markdown content")
		}
	} else {
		// JSON as default
		err := json.Unmarshal(bytes, entry)
		if err != nil {
			return nil, errors.Wrap(err, "Couldn't parse content with contentType %s -> %s", contentType, bytes)
		}
	}

	return entry, nil
}

// ParseMarkdownEntry parses entries stored in markdown format
func ParseMarkdownEntryFromStorage(data []byte, entry *Entry) error {

	if entry == nil {
		return errors.New("entry is nil")
	}

	matter := Frontmatter{}

	rest, err := frontmatter.Parse(bytes.NewReader(data), &matter)
	if err != nil {
		return err
	}
	entry.Meta = matter.Meta
	entry.Content, err = Instantiate(matter.Meta.Type)
	if err != nil {
		return errors.Wrap(err, "Type not registered: %s", entry.Meta.Type)
	}
	// fmt.Println("frontmatter:", matter.Meta.Type, "---", entry.Content)

	unmarshaler, ok := entry.Content.(markdowner)
	if !ok {
		return errors.Wrap(err, "markdowner interface not implemented: %T", entry.Content)
	}
	// m := Frontmatter{Meta: entry.Meta}
	err = unmarshaler.UnmarshalMarkdown(rest, &matter)
	if err != nil {
		return err
	}

	// entry.Meta = matter

	// if matter.Type == "Article" {

	// } else {

	// 	entry.Content = &MDPage{
	// 		Markdown: Markdown{content: string(rest)},
	// 		Lines:    []string{},
	// 		// Blocks:   Bloks{},
	// 	}
	// }

	// extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock

	// p := parser.NewWithExtensions(extensions)
	// doc := p.Parse(bytes)

	// for _, b := range doc.GetChildren() {
	// 	if b.IsHeading() {
	// 		fmt.Printf("Heading: %s\n", b.GetText())
	// 	} else if b.IsBlockquote() {
	// 		fmt.Printf("Blockquote: %s\n", b.GetText())
	// 	} else if b.IsList() {
	// 		fmt.Printf("List: %s\n", b.GetText())
	// 	}
	// }
	// fmt.Println("doc", doc.ToText())

	return nil
}
