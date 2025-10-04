package ctype

import (
	"encoding/json"
	"reflect"

	"net/url"

	"bitbucket.org/hotelplan/webcc-content/cms/content"
	"bitbucket.org/hotelplan/webcc-content/cms/mgmt/search"
	"bitbucket.org/hotelplan/webcc-content/cms/permission"
	"bitbucket.org/hotelplan/webcc-content/cms/pkg/errors"
)

func init() {

	// TODO: Unterscheidung ContentType / Block
	content.Register(Page{})
	content.Register(Header{})
	content.Register(Footer{})
	content.Register(ContentMenu{})
	// content.Register(Slugs{})
	content.Register(Seo{})
	content.Register(Destination{})
	content.Register(DomainRedirects{})
	content.Register(Badge{})
	content.Register(USPBar{})
	content.Register(USPCards{})
	content.Register(USPColumns{})
	content.Register(Announcement{})

	// Blocks / Components
	content.Register(SearchBar{})
	content.Register(Breadcrumbs{})
	content.Register(Heading{})
	content.Register(Menu{})
	content.Register(Columns{})
	content.Register(Hero{})
	content.Register(Richtext{})
	content.Register(TeaserGrid{})
	content.Register(ObjectList{})
	content.Register(USPBarBlock{})
	content.Register(USPCardsBlock{})
	content.Register(USPColumnsBlock{})

	content.Register(StartHeaderTeasers{})
	content.Register(StartInspirations{})
	content.Register(StartTopDestinations{})
	content.Register(StartRecommendations{})
	content.Register(StartSEOLinks{})
	content.Register(StartOwnerTeaser{})
	content.Register(StartSeoTeaser{})
	content.Register(StartUSPBar{})
	content.Register(StartUSPCards{})
	content.Register(StartUSPColumns{})

	content.Register(BadgeBlock{})

	content.Register(Catalogs{})
	content.Register(LSOList{})
	content.Register(LSODetail{})
	content.Register(LSOMedia{})

	content.Register(LandingMarkdown{})
	content.Register(LandingRaw{})
	content.Register(LandingIFrame{})
	content.Register(LandingTextCarousel{})
	content.Register(LandingImageCarouselWithText{})
	content.Register(LandingTextWithImage{})
	content.Register(LandingTeasers{})
	content.Register(LandingRecommendations{})
	content.Register(LandingHero{})
	content.Register(LandingH1{})
	content.Register(LandingBreadcrumbs{})
}

//////////////////////////
////////// PAGE //////////
//////////////////////////

type Page struct {
	content.ContentType `type:"Page" json:"-" folder:"pages" mimetype:"application/json"`

	Title       string   `json:"title"`
	Description string   `json:"description"`
	Robots      string   `json:"robots"`
	Canonical   string   `json:"canonical"`
	Meta        MetaTags `json:"meta"`
	VirtPath    string   `json:"virtpath"`

	Body content.Bloks `json:"body"`

	// TODO Phase 3: AdditionalBody aus ConfApi durch folgendes ersetzen
	// Data map[string]string `json:"data,omitempty"`
}

type MetaTags struct {
	Image      string            `json:"image,omitempty"`
	URL        string            `json:"url,omitempty"`
	Type       string            `json:"type,omitempty"`
	Additional map[string]string `json:"additional,omitempty"`
}

func (p *Page) Clone() interface{} {
	copy := *p
	copy.Meta = MetaTags{Image: p.Meta.Image, URL: p.Meta.URL, Type: p.Meta.Type, Additional: map[string]string{}}
	for k, v := range p.Meta.Additional {
		copy.Meta.Additional[k] = v
	}
	copy.Body = p.Body.Clone() // make(content.Bloks, len(p.Body))
	// for i, blok := range p.Body {
	// 	// copy.Body[i] = blok.Clone()
	// 	newblockptr, err := blok.Clone()
	// 	if err != nil {
	// 		log.Fatalf("fatal error in page block clone %t", blok.Content)
	// 		return nil
	// 	}
	// 	copy.Body[i] = *newblockptr
	// }
	return &copy
}

func (p *Page) Refs() []*content.EntryRef { return p.Body.Refs() }

func (p *Page) Links() []*content.EntryLink { return p.Body.Links() }

func (p *Page) ResolveLinks(locale string, f func(string, string) string) {
	p.Body.ResolveLinks(locale, f)
}

func (p *Page) MarshalJSON() ([]byte, error) {
	foo := struct {
		Title       string        `json:"title"`
		Description string        `json:"description"`
		Robots      string        `json:"robots"`
		Canonical   string        `json:"canonical"`
		Meta        MetaTags      `json:"meta"`
		VirtPath    string        `json:"virtpath"`
		Body        content.Bloks `json:"body"`
	}{Title: p.Title, Description: p.Description, Robots: p.Robots, Canonical: p.Canonical, Meta: p.Meta, VirtPath: p.VirtPath, Body: p.Body}
	return json.Marshal(foo)
}

func (p *Page) MarshalText() ([]byte, error) {

	return p.Body.MarshalText()
}

var (
	PAGE_CREATE    permission.Type = permission.Define("page", "create", true, true)
	PAGE_SEO_EDIT  permission.Type = permission.Define("page", "seo.edit", true, true)
	PAGE_BODY_EDIT permission.Type = permission.Define("page", "body.edit", true, true)
	PAGE_DELETE    permission.Type = permission.Define("page", "delete", true, true)
)

// type Permissions interface {
// 	Denies(string) bool
// 	CheckPermission(string, string, string) bool
// }

func (p *Page) CheckPermissionForWrite(scope permission.Scope, previous *content.Entry) error {

	if previous == nil && scope.Denies(PAGE_CREATE) {
		return errors.New("page create permission missing")
	}

	// // TODO: general page edit permission???
	// if scope.HasPermission(PAGE_EDIT, ... , ... ) {
	// 	return nil
	// }

	permissionMap := map[string]permission.Type{
		"Title":       PAGE_SEO_EDIT,
		"Description": PAGE_SEO_EDIT,
		"Robots":      PAGE_SEO_EDIT,
		"Canonical":   PAGE_SEO_EDIT,
		"Meta":        PAGE_SEO_EDIT,
		"VirtPath":    PAGE_SEO_EDIT,
		"Body":        PAGE_BODY_EDIT,
	}
	for _, field := range p.Diff(previous.Content.(*Page)) {

		permission := permissionMap[field]

		if !scope.HasPermission(permission, previous.Locale, previous.Path) {
			return errors.New("missing permission %s:%s:%s for changed field %s", permissionMap[field].String(), previous.Locale, previous.Path, field)
		}
	}

	return nil
}

func (p *Page) Diff(other *Page) []string {
	fields := []string{}
	if p.Title != other.Title {
		fields = append(fields, "Title")
	}
	if p.Description != other.Description {
		fields = append(fields, "Description")
	}
	if p.Robots != other.Robots {
		fields = append(fields, "Robots")
	}
	if p.Canonical != other.Canonical {
		fields = append(fields, "Canonical")
	}
	if p.VirtPath != other.VirtPath {
		fields = append(fields, "VirtPath")
	}
	if !reflect.DeepEqual(p.Meta, other.Meta) {
		fields = append(fields, "Meta")
	}
	if !reflect.DeepEqual(p.Body, other.Body) {
		fields = append(fields, "Body")
	}
	return fields
}
func (p *Page) AugmentSearchDoc(doc *search.Document, level search.Level) (interface{}, error) {
	doc.DocumentContent = search.DocumentContent{
		Image: "",
		Title: map[string]string{doc.Meta.Locale[:2]: p.Title},
		Text:  map[string]string{doc.Meta.Locale[:2]: p.Description},
	}
	return nil, nil
}

/////////////////////////////////////
////////// Header & Footer //////////
/////////////////////////////////////

type Header struct {
	content.ContentType `type:"Header" json:"-"`
	LogoCaption         string    `json:"logoCaption"`
	Menu                []NavItem `json:"menu"`
}

func (b *Header) Clone() interface{} {
	return &Header{
		LogoCaption: b.LogoCaption,
		Menu:        append(b.Menu[:0:0], b.Menu...),
	}
}

type NavItem struct {
	Label string `json:"label"`
	Link  Link   `json:"link"`
}

type Footer struct {
	content.ContentType `type:"Footer" json:"-"`
	Menu                []NavItem         `json:"menu"`
	SocialMedia         []SocialMediaItem `json:"socialMedia"`
	Partnerships        []PartnershipItem `json:"partnerships"`
	LocaleURLs          map[string]string `json:"localeURLs"`
}

func (b *Footer) Clone() interface{} {
	clone := Footer{
		Menu:         append(b.Menu[:0:0], b.Menu...),
		SocialMedia:  append(b.SocialMedia[:0:0], b.SocialMedia...),
		Partnerships: append(b.Partnerships[:0:0], b.Partnerships...),
		LocaleURLs:   make(map[string]string),
	}
	for k, v := range b.LocaleURLs {
		clone.LocaleURLs[k] = v
	}

	return &clone
}

type SocialMediaItem struct {
	Icon string `json:"icon"`
	Link Link   `json:"link"`
}
type PartnershipItem struct {
	Label string `json:"label"`
	Image Image  `json:"image" folder:"..."`
}

type Image struct {
	Asset         content.Asset `json:"asset"`
	AlternateText string        `json:"alt"`
	Width         int           `json:"width"`  // TODO: rausfaktorisieren
	Height        int           `json:"height"` // TODO: rausfaktorisieren
}

/////////////////////////////////////
////////// Menu /////////////////////
/////////////////////////////////////

type ContentMenu struct {
	content.ContentType `type:"ContentMenu" folder:"" json:"-"`
	MainMenu            []MenuItem `json:"mainMenu"`
	ContentPagesMenu    []MenuItem `json:"contentPagesMenu"`
}

func (b *ContentMenu) Clone() interface{} {
	return &ContentMenu{
		MainMenu:         append(b.MainMenu[:0:0], b.MainMenu...),
		ContentPagesMenu: append(b.ContentPagesMenu[:0:0], b.ContentPagesMenu...),
	}
}

type MenuItem struct {
	Name     string     `json:"name"`
	URL      string     `json:"url"`
	Children []MenuItem `json:"children"`
}

//////////

// type Slugs struct {
// 	content.ContentType `type:"Slugs" folder:"" json:"-"`
// 	Menu                []MenuItem          `json:"menu"`
// 	BuyingOffices       map[string]string   `json:"buyingoffices"`
// 	SlugsByID           map[string]SlugInfo `json:"slugsByID"`
// }

// type SlugInfo struct {
// 	ID         int    `json:"id"`
// 	EZPageType int    `json:"ezPageType"`
// 	Name       string `json:"name"`
// 	ParentId   int    `json:"parentId"`
// 	UrlPath    string `json:"urlPath"`
// 	VirtPath   string `json:"virtPath"`
// }

////////// ??? hat andere Struktur als im Repo, wird nicht referenziert im conf service

type Destination struct {
	content.ContentType `type:"Destination"  json:"-"`
	Code                string            `json:"code"`
	Image               map[string]string `json:"image"`
	ImgAlt              map[string]string `json:"alt"`
	Name                map[string]string `json:"name"`
	URL                 map[string]string `json:"url"`
}

func (d *Destination) Clone() interface{} {
	clone := *d
	return &clone
}

///////////////////////////////
////////// Redirects //////////
///////////////////////////////

type DomainRedirects struct {
	content.ContentType `type:"DomainRedirects" folder:"" json:"-"`

	Domain       string                      `json:"domain"`
	RedirectsNew map[string][]RedirectTarget `json:"redirects_new"`
	Redirects    []Redirect                  `json:"redirects"`
}

func (b *DomainRedirects) Clone() interface{} {
	clone := DomainRedirects{
		Domain:       b.Domain,
		RedirectsNew: make(map[string][]RedirectTarget),
		Redirects:    append(b.Redirects[:0:0], b.Redirects...),
	}
	for k, v := range b.RedirectsNew {
		clone.RedirectsNew[k] = append(v[:0:0], v...)
	}

	return &clone
}

type RedirectTarget struct {
	MatchParams  map[string]string `json:"matchQueryParams"` // zu matchende Query-Parameter
	TargetPath   string            `json:"path"`             // Zielpfad der URL
	TargetTokens url.Values        `json:"queryparams"`      // evtl. zusätzliche Query-Parameter
}
type RedirectOption int

type Redirect struct {
	SourceURL     url.URL
	SourceOptions RedirectOption
	TargetURL     url.URL
	TargetOptions RedirectOption
}

func (r *Redirect) MarshalJSON() ([]byte, error) {
	aux := struct {
		SourceURL string
		TargetURL string
	}{SourceURL: r.SourceURL.String(), TargetURL: r.TargetURL.String()}

	return json.Marshal(aux)
}

func (r *Redirect) UnmarshalJSON(data []byte) error {
	var aux struct {
		SourceURL string
		TargetURL string
	}
	err := json.Unmarshal(data, &aux)
	if err != nil {
		return err
	}

	sourceURL, err := url.Parse(aux.SourceURL)
	r.SourceURL = *sourceURL
	if err != nil {
		return err
	}
	targetURL, err := url.Parse(aux.TargetURL)
	if err != nil {
		return err
	}
	r.TargetURL = *targetURL

	return nil
}

// ///////////////////////////
// ////// Badge //////////////
// ///////////////////////////
type Badge struct {
	content.ContentType `type:"Badge" folder:"" json:"-"`

	Code     string `json:"code"`
	Title    string `json:"title"`
	Text     string `json:"text"`
	Markdown string `json:"markdown"`
}

func (b *Badge) Clone() interface{} {
	clone := *b
	return &clone
}

// //////////////////////////////////
// ////// Announcements ////////////////////
// /////////////////////////////////
type Announcement struct {
	content.ContentType `type:"Announcement" folder:"" json:"-"`
	Name                string                           `json:"name"`
	Description         string                           `json:"description"`
	Icon                string                           `json:"icon"`
	Locales             map[string]LocalizedAnnouncement `json:"locales"`
	From                string                           `json:"from"`
	To                  string                           `json:"to"`
	Variant             string                           `json:"variant"`
}
type LocalizedAnnouncement struct {
	Active bool   `json:"active"`
	Title  string `json:"title"`
	Text   string `json:"text"` // Markdown string `json:"markdown"` TODO: check ob markdown besser past
	Link   Link   `json:"link"`
}

func (b *Announcement) Clone() interface{} {
	clone := *b
	return &clone
}

////////// ???

type Catalogs struct {
	content.Component `name:"Catalogs" json:"-"`
	Cataloges         []Catalog `json:"catalogs"`
}

func (b *Catalogs) Clone() content.Bloker {
	clone := Catalogs{Cataloges: make([]Catalog, len(b.Cataloges))}

	for i, c := range b.Cataloges {
		catalog := c
		for k, v := range c.Countries {
			catalog.Countries[k] = v
		}
		clone.Cataloges[i] = catalog
	}
	return &clone
}

type Catalog struct {
	ItemNumber     string            `json:"itemNumber"`
	Title          string            `json:"title"`
	SubTitle       string            `json:"subTitle"`
	Image          Image             `json:"image"`
	Description    string            `json:"description"`
	FlippingBookID string            `json:"flippingBookID"`
	Countries      map[string]string `json:"countries"`
}
