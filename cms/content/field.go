package content

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type Bloks []Blok

func (b Bloks) Clone() Bloks {
	clone := make(Bloks, len(b))
	for i, blok := range b {
		clone[i] = blok.Clone()
	}
	return clone
}

func (b Bloks) Refs() []*EntryRef {
	refs := []*EntryRef{}
	for i := range b {
		block := &b[i]
		if r, ok := block.Content.(interface{ Refs() []*EntryRef }); ok {
			refs = append(refs, r.Refs()...)
		}
	}
	return refs
}

func (b Bloks) Links() []*EntryLink {
	links := []*EntryLink{}
	for i := range b {
		block := &b[i]
		if r, ok := block.Content.(interface{ Links() []*EntryLink }); ok {
			links = append(links, r.Links()...)
		}
	}
	return links
}

func (b Bloks) ResolveLinks(locale string, f func(string, string) string) {
	for i := range b {
		block := &b[i]
		if r, ok := block.Content.(interface {
			ResolveLinks(string, func(string, string) string)
		}); ok {
			r.ResolveLinks(locale, f)
		}
	}
}
func (b Bloks) MarshalJSON() ([]byte, error) {
	return json.Marshal([]Blok(b))
}

func (b Bloks) MarshalText() ([]byte, error) {

	var marshaledBlocks [][]byte
	for i := range b {
		block := &b[i]
		if r, ok := block.Content.(interface {
			MarshalText() ([]byte, error)
		}); ok {
			bytes, err := r.MarshalText()
			if err != nil {
				return nil, err
			}
			marshaledBlocks = append(marshaledBlocks, bytes)
		} else {
			name := componentName(block.Content)
			bytes := fmt.Sprintf("<%s id=\"%d\" ... ><%s>", name, block.ID, name)
			marshaledBlocks = append(marshaledBlocks, []byte(bytes))
		}
	}
	return bytes.Join(marshaledBlocks, []byte("\n\n")), nil
}

//////////////////////////////
//////////  ASSETS  //////////
//////////////////////////////

type Asset string

//////////////////////////////
//////////   LINK   //////////
//////////////////////////////

// Link ist FieldType und kann über StructTags konfiguriert werden:
// * Beschränkung auf intern / speziellen ContentType / Ordner / ...
// * Option für target (in neuem Fenster) im Editor anzeigen
// * Im Editor angezeigter Name
// * Möglichkeit, die AnchorID festzulegen

type Link struct {
	Href EntryLink `json:"href"` // Pfad des Entries

	// _blank    Opens the linked document in a new window or tab
	// _self     Opens the linked document in the same frame as it was clicked (this is default)
	// _parent   Opens the linked document in the parent frame
	// _top      Opens the linked document in the full body of the window
	// framename Opens the linked document in the named iframe
	Target string `json:"target,omitempty"`
}

// evtl. wäre href oder uri der bessere Name, da es hier eigentlich um das Ziel eines Links geht, das entweder eine interne EntryID oder eine URL ist.
// TODO: Thema in Phase 3 angehen
type EntryLink struct {
	URL     string // speichert Externe URL oder leer
	EntryID string // ID des verlinkten Entries (bei internem Link) sonst leer
	Slug    string // wird nur on the fly aufgelöst, wenn angefordert
}

func (l *EntryLink) ResolveLinks(locale string, f func(string, string) string) {

	l.Slug = f(l.EntryID, locale)
	fmt.Printf(" +++ link to entry %s resolved to %s | %p +++\n", l.EntryID, l.Slug, l)
}

func (l EntryLink) MarshalJSON() ([]byte, error) {
	switch {
	case l.Slug != "":
		return json.Marshal(l.Slug)
	case l.EntryID != "":
		return json.Marshal("entry://" + l.EntryID)
	}
	return json.Marshal(l.URL)
}

// Unmarshal erwartet einen Link in Stringrepräsentation.
// Diese wird als URI geparst und sollte das Scheme `entry` sein ins Feld EntryID und sonst in das Feld URL geschrieben.
// Das Feld SLug wird weder gelesen noch geschrieben.
func (l *EntryLink) UnmarshalJSON(data []byte) error {
	var aux string
	err := json.Unmarshal(data, &aux)
	if err != nil {
		return err
	}

	// Nachfolgend ist die Linkstruktur der website aufgelistet. Das hat aber keinen Bezug zum pkg content, sondern wird im ConfService aufgelöst
	// search://search?pool=true
	// search://DE
	// search://DE12?pool=true
	// search://DE12345
	// accom://DE5420.100.1
	// page://topics
	// page://lsos
	// page://startpage
	// page://bookmarks
	// travelguide://start
	// travelguide://topics/hiking
	url, err := url.Parse(aux)
	switch {
	case url.Scheme == "entry":
		l.EntryID = url.Host + url.Path
	case err != nil:
		l.EntryID = aux
	default:
		l.URL = url.String()
	}

	return nil
}

/////////////////////////////////
//////////  MARKDOWN  //////////
///////////////////////////////

type Markdown struct {
	content string
}

var linkmatcher = regexp.MustCompile(`\[[^\]]*\]\([^\)]*\)`)
var urlmatcher = regexp.MustCompile(`^\[[^\]]*\]\(([^\)]*)\)$`)

func (md *Markdown) ResolveLinks(resolverfunc func(string) string) {

	urlresolver := func(mdlink string) string {
		if matches := urlmatcher.FindStringSubmatch(mdlink); len(matches) == 2 {

			mdlink = strings.Replace(mdlink, matches[1], resolverfunc(matches[1]), 1)
		}
		return mdlink
	}

	md.content = linkmatcher.ReplaceAllStringFunc(md.content, urlresolver)
}

func (md *Markdown) ResolveLinksLocale(locale string, resolverfunc func(string, string) string) {

	urlresolver := func(mdlink string) string {
		if matches := urlmatcher.FindStringSubmatch(mdlink); len(matches) == 2 {

			mdlink = strings.Replace(mdlink, matches[1], resolverfunc(matches[1], locale), 1)
		}
		return mdlink
	}

	md.content = linkmatcher.ReplaceAllStringFunc(md.content, urlresolver)
}

func (md *Markdown) MarshalText() (text []byte, err error) { return []byte(md.content), nil }

func (md *Markdown) UnmarshalText(text []byte) error {

	md.content = string(text)

	return nil
}
