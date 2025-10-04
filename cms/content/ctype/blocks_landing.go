package ctype

import (
	"regexp"
	"strings"

	"bitbucket.org/hotelplan/webcc-content/cms/content"
)

// Components:
// *LandingHero
// *LandingH1
// *LandingMarkdown
// *LandingTextWithImage
// *LandingTeasers
// *LandingImageCarouselWithText
// *LandingTextCarousel
// *LandingRecommendations

// *LandingIFrame
// *LandingRaw: -> wird nicht umgesetzt, dafür gibt es IFrame
// *LandingBreadcrumbs: ???

type LandingHero struct {
	content.Component `name:"LandingHero" json:"-"`
	Image             LandingImage `json:"image"`
	ShowSearchBar     bool         `json:"showSearchBar"`
	Filter            []string     `json:"filter"`
	FilterLabel       string       `json:"filterLabel"`
}

func (b *LandingHero) Clone() content.Bloker {
	clone := *b
	clone.Filter = make([]string, len(b.Filter))
	copy(clone.Filter, b.Filter)
	return &clone
}

type LandingH1 struct {
	content.Component `name:"LandingH1" json:"-"`

	H1     string `json:"h1"`
	Styles Style  `json:"styles"`
}

func (b *LandingH1) Clone() content.Bloker {
	clone := *b
	return &clone
}

type Markdown string

var linkmatcher = regexp.MustCompile(`\[[^\]]*\]\([^\)]*\)`)
var urlmatcher = regexp.MustCompile(`^\[[^\]]*\]\(([^\)]*)\)$`)

func (md Markdown) resolveLinksRegex(locale string, resolver func(string, string) string) string {

	urlresolver := func(mdlink string) string {
		if matches := urlmatcher.FindStringSubmatch(mdlink); len(matches) == 2 {
			resolvedurl := resolver(matches[1], locale)
			mdlink = strings.Replace(mdlink, matches[1], resolvedurl, 1)
			// fmt.Println("LINK:", link, matches[1], url, replaced)
		}
		return mdlink
	}

	return linkmatcher.ReplaceAllStringFunc(string(md), urlresolver)
}

type LandingMarkdown struct {
	content.Component `name:"LandingMarkdown" json:"-"`
	Text              Markdown `json:"text"`
	Styles            Style    `json:"styles"`
}

func (b *LandingMarkdown) Clone() content.Bloker {
	clone := *b
	return &clone

}
func (b *LandingMarkdown) ResolveLinks(locale string, f func(string, string) string) {

	md := b.Text.resolveLinksRegex(locale, f)

	b.Text = Markdown(md)
}

// Image link text component: https://jira.migros.net/browse/IHGWEBCC-1193
type LandingTextWithImage struct {
	content.Component `name:"LandingTextWithImage" json:"-"`

	Title      string   `json:"title"`
	Text       Markdown `json:"text"`
	TextExpand string   `json:"textExpand"`

	Image struct {
		LandingImage
		Position string `json:"position"`
	} `json:"image"`

	Link struct {
		Text   string `json:"text"`
		Href   string `json:"href"`
		Target string `json:"target"`
	} `json:"link"`

	Styles Style `json:"styles"`
}

func (b *LandingTextWithImage) Clone() content.Bloker {
	clone := *b
	return &clone

}
func (b *LandingTextWithImage) ResolveLinks(locale string, f func(string, string) string) {

	b.Text = Markdown(b.Text.resolveLinksRegex(locale, f))
	b.Link.Href = f(b.Link.Href, locale)
}

// Static marketing teaser component: https://jira.migros.net/browse/IHGWEBCC-1197
type LandingTeasers struct {
	content.Component `name:"LandingTeasers" json:"-"`

	Title string `json:"title"` // Main title (line of text) optional

	Teasers []struct {
		Image LandingImage `json:"image"`
		Link  LandingLink  `json:"link"`
	} `json:"teasers"`

	Position string `json:"position"`

	Styles Style `json:"styles"`
}

func (b *LandingTeasers) Clone() content.Bloker {
	clone := *b
	clone.Teasers = append(b.Teasers[:0:0], b.Teasers...)
	return &clone

}
func (b *LandingTeasers) ResolveLinks(locale string, f func(string, string) string) {
	for i := range b.Teasers {
		t := &b.Teasers[i]
		t.Link.URL = f(t.Link.URL, locale)
	}
}

// Multi-Image text component: https://jira.migros.net/browse/IHGWEBCC-1195
type LandingImageCarouselWithText struct {
	content.Component `name:"LandingImageCarouselWithText" json:"-"`
	Title             string   `json:"title"`
	Text              Markdown `json:"text"`
	Teasers           []struct {
		Image LandingImage `json:"image"`
		Link  LandingLink  `json:"link"`
	} `json:"teasers"`
	Size   string `json:"size"` // Achtung: diese Größe sollte außerhalb der Liste definiert werden. Wahrscheinlich sollten doch alle Bilder gleich groß sein.
	Styles Style  `json:"styles"`
}

func (b *LandingImageCarouselWithText) ResolveLinks(locale string, f func(string, string) string) {
	b.Text = Markdown(b.Text.resolveLinksRegex(locale, f))
	for i := range b.Teasers {
		t := &b.Teasers[i]
		t.Link.URL = f(t.Link.URL, locale)
	}
}

func (b *LandingImageCarouselWithText) Clone() content.Bloker {
	clone := *b
	clone.Teasers = append(b.Teasers[:0:0], b.Teasers...)
	return &clone
}

// type landingImageCarouselWithTextTeaser

type LandingTextCarousel struct {
	content.Component `name:"LandingTextCarousel" json:"-"`

	Title string `json:"title"`

	// Multiple Main texts - supporting bold, italic, headings (h3, h4, h5), generic list, numbered list, links) mandatory
	// limited on number of number of characters (tbd by business) mandatory min 2 texts
	// created mainly needed for unique draggable-id in frontend
	Texts []struct {
		Text    Markdown `json:"text" format:"bold,italic,h2,h3,h4,h5,ul,ol,link"`
		Created string   `json:"created"`
	} `json:"texts" min:"2"`
	Image struct {
		LandingImage
		Position string   `json:"position"`
		Display  []string `json:"display"`
	} `json:"image"`

	Styles Style `json:"styles"`
}

func (b *LandingTextCarousel) ResolveLinks(locale string, f func(string, string) string) {
	for i := range b.Texts {
		t := &b.Texts[i]
		t.Text = Markdown(t.Text.resolveLinksRegex(locale, f))
	}
}

// Die Idee hinter der Raw-Componente ist die Einbettung von Inhalten wie iFrames, JS-Script-Tags, Web-Components,...
// Das hat natürlich Security-Implikationen. Wahrscheinlich ist es deshalb sinnvoll, die Inhalte mehr zu kontrollieren und
// z.B. für iframes eine eigene Komponente zu haben
func (b *LandingTextCarousel) Clone() content.Bloker {
	clone := *b
	clone.Texts = append(b.Texts[:0:0], b.Texts...)
	clone.Image.Display = append(b.Image.Display, b.Image.Display...)
	return &clone
}

// Die Idee hinter der Raw-Componente war die Einbettung von Inhalten wie iFrames, JS-Script-Tags, Web-Components,...
// Das hat Security-Implikationen. Wahrscheinlich ist es deshalb sinnvoll, die Inhalte mehr zu kontrollieren und
// für Iframes eine eigene Komponente zu haben
type LandingIFrame struct {
	content.Component `name:"LandingIFrame" json:"-"`

	Attributes map[string]string `json:"attributes"`
	Src        string            `json:"src"`
	SrcDoc     string            `json:"srcdoc"`
	Width      int               `json:"width"`
	Height     int               `json:"height"`
	Autoheight bool              `json:"autoheight"`
}

func (b *LandingIFrame) Clone() content.Bloker {
	clone := *b
	clone.Attributes = make(map[string]string)
	for k, v := range b.Attributes {
		clone.Attributes[k] = v
	}
	return &clone
}

// https://jira.migros.net/browse/IHGWEBCC-1255
type LandingRecommendations struct {
	content.Component `name:"LandingRecommendations" json:"-"`
	Title             string            `json:"title"`
	Query             string            `json:"query"`
	QueryParams       map[string]string `json:"params"`
	Layout            string            `json:"layout"`
	Styles            Style             `json:"styles"`
}

func (b *LandingRecommendations) Clone() content.Bloker {
	clone := *b
	clone.QueryParams = make(map[string]string)
	for k, v := range b.QueryParams {
		clone.QueryParams[k] = v
	}
	return &clone
}

//////// HELPERS ////////
////////////////////////

// LandingImage bescheibt ein Bild aus der IK-Media-Library mit zusätzlichen Daten
type LandingImage struct {
	Source        content.Asset `json:"src"`
	AlternateText string        `json:"alt"`
	Copyright     string        `json:"copyright"`
}

type LandingLink struct {
	Text      string `json:"text"`
	URL       string `json:"href"`
	Target    string `json:"target"`
	Alignment string `json:"alignment"`
}

type Style struct {
	BackgroundColor string `json:"styleBGColor"`
	TextColor       string `json:"styleTextColor"`
	PaddingTop      string `json:"stylePaddingTop"`
	PaddingBottom   string `json:"stylePaddingBottom"`

	// als Alternative könnte man hier auch einfach eine Liste von (tailwind)-Classes spezifizieren,
	// die dann auf dem Root-Element der Component angewendet werden.
	Classes string `json:"classes"`
}

////// under discussion:            ////
///// Potential future components //////
///////////////////////////////////////

// war in der urspruenglichen Anforderung enthalten, deshalb noch drin
type LandingBreadcrumbs struct {
	content.Component `name:"LandingBreadcrumbs" json:"-"`
	Links             []LabelledLink `json:"links"`
}

func (b *LandingBreadcrumbs) Clone() content.Bloker {
	clone := *b
	clone.Links = append(b.Links[:0:0], b.Links...)
	return &clone
}

// https://jira.migros.net/browse/IHGWEBCC-1184
type LandingRaw struct {
	content.Component `name:"LandingRaw" json:"-"`
	Title             string
	Rawtext           string
	BGColor           string
}

func (b *LandingRaw) Clone() content.Bloker {
	clone := *b
	return &clone
}

type landingMedia struct {

	// type unterscheiden zwischen Bild, Youtube-Video, Vimeo-Video oder....
	// Type string
	// Ein Link zu Vimeo oder Youtube oder ...
	VideoURL string `json:"video_url"`

	// Ein Bild aus IK
	// Im Falle von Videos dient das Bild als Preview bevor das Video gestartet wird.
	Asset content.Asset `json:"asset"`

	AlternateText string `json:"alt"`
	Copyright     string `json:"copyright"`
}
