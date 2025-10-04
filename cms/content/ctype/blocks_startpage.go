package ctype

import (
	"bitbucket.org/hotelplan/webcc-content/cms/content"
)

////////// HEADER TEASERS ///////////

type StartHeaderTeasers struct {
	content.Component `name:"StartHeaderTeasers" json:"-"`
	Teasers           []Teaser `json:"teasers"`
}

func (b *StartHeaderTeasers) Clone() content.Bloker {
	dst := StartHeaderTeasers{Teasers: make([]Teaser, len(b.Teasers))}
	copy(dst.Teasers, b.Teasers)
	return &dst
}

func (b *StartHeaderTeasers) Links() []*content.EntryLink {
	var links []*content.EntryLink
	for i := range b.Teasers {
		links = append(links, (&b.Teasers[i]).Links()...)
	}
	return links
}

func (b *StartHeaderTeasers) ResolveLinks(locale string, f func(string, string) string) {
	for i := range b.Teasers {
		teaser := &b.Teasers[i]
		teaser.ResolveLinks(locale, f)
	}
}

////////// INSPIRATION TEASERS ///////////

type StartInspirations struct {
	content.Component `name:"StartInspirations" json:"-"`
	Teasers           []Teaser `json:"teasers"`
}

func (b *StartInspirations) Clone() content.Bloker {
	var copy StartInspirations
	copy.Teasers = append(copy.Teasers, b.Teasers...)
	return &copy
}

func (b *StartInspirations) Links() []*content.EntryLink {
	var links []*content.EntryLink
	for i := range b.Teasers {
		links = append(links, (&b.Teasers[i]).Links()...)
	}
	return links
}

func (b *StartInspirations) ResolveLinks(locale string, f func(string, string) string) {
	for i := range b.Teasers {
		teaser := &b.Teasers[i]
		teaser.ResolveLinks(locale, f)
	}
}

////////// TOP DEST TEASERS ///////////

type StartTopDestinations struct {
	content.Component `name:"StartTopDestinations" json:"-"`
	Countries         []Teaser           `json:"countries"`
	Regions           []Teaser           `json:"regions"`
	AllCountries      []IconLabelledLink `json:"allCountries"`
}

func (b *StartTopDestinations) Clone() content.Bloker {
	return &StartTopDestinations{
		Countries:    append(b.Countries[:0:0], b.Countries...),
		Regions:      append(b.Regions[:0:0], b.Regions...),
		AllCountries: append(b.AllCountries[:0:0], b.AllCountries...),
	}
}

func (b *StartTopDestinations) Links() []*content.EntryLink {
	var links []*content.EntryLink
	for i := range b.Countries {
		links = append(links, (&b.Countries[i]).Links()...)
	}
	for i := range b.Regions {
		links = append(links, (&b.Regions[i]).Links()...)
	}
	return links
}

////////// RECOMMENDATIONS ///////////

type StartRecommendations struct {
	content.Component `name:"StartRecommendations" json:"-"`
}

func (b *StartRecommendations) Clone() content.Bloker {
	return &StartRecommendations{}
}

////////// OWNER TEASER ///////////

type StartOwnerTeaser struct {
	content.Component `name:"StartOwnerTeaser" json:"-"`
	Teaser
	ButtonText string `json:"buttonText"`
}

func (b *StartOwnerTeaser) Clone() content.Bloker {
	return &StartOwnerTeaser{Teaser: b.Teaser, ButtonText: b.ButtonText}
}

func (b *StartOwnerTeaser) Links() []*content.EntryLink { return b.Teaser.Links() }

////////// SEO TEASERS ///////////

type StartSeoTeaser struct {
	content.Component `name:"StartSeoTeaser" json:"-"`
	Title             string `json:"title"`
	Text              string `json:"text"`
	Markdown          string `json:"markdown"`
}

func (b *StartSeoTeaser) Clone() content.Bloker {
	clone := *b
	return &clone
}

////////// SEO LINKS ///////////

type StartSEOLinks struct {
	content.Component `name:"StartSEOLinks" json:"-"`
	Title             string         `json:"title"`
	Links             []LabelledLink `json:"links"`
}

func (b *StartSEOLinks) Clone() content.Bloker {
	return &StartSEOLinks{
		Title: b.Title,
		Links: append(b.Links[0:0], b.Links...),
	}
}

////////// Start USPS ///////////

type StartUSPBar struct {
	content.Component `name:"StartUSPBar" json:"-"`
	Bar               content.EntryRef `json:"bar,omitempty"`
}

func (b *StartUSPBar) Refs() []*content.EntryRef { return []*content.EntryRef{&b.Bar} }

func (b *StartUSPBar) Clone() content.Bloker {
	clone := *b
	return &clone
}

//////////

type StartUSPCards struct {
	content.Component `name:"StartUSPCards" json:"-"`
	Cards             content.EntryRef `json:"cards,omitempty"`
}

func (b *StartUSPCards) Clone() content.Bloker {
	clone := *b
	return &clone
}

func (b *StartUSPCards) Refs() []*content.EntryRef { return []*content.EntryRef{&b.Cards} }

//////////

type StartUSPColumns struct {
	content.Component `name:"StartUSPColumns" json:"-"`
	Columns           content.EntryRef `json:"columns,omitempty"`
}

func (b *StartUSPColumns) Clone() content.Bloker {
	clone := *b
	return &clone
}

func (b *StartUSPColumns) Refs() []*content.EntryRef { return []*content.EntryRef{&b.Columns} }
