package ctype

import "bitbucket.org/hotelplan/webcc-content/cms/content"

type Hero struct {
	content.Component `name:"Hero" json:"-"`
	Headline          string `json:"headline"`
	Image             Image  `json:"image"`
}

func (b *Hero) Clone() content.Bloker {
	clone := *b
	return &clone
}

type BadgeBlock struct {
	content.Component `name:"BadgeBlock" json:"-"`
	Badge             content.EntryRef `json:"badge,omitempty"`
}

func (b *BadgeBlock) Refs() []*content.EntryRef {
	return []*content.EntryRef{&b.Badge}
}

func (b *BadgeBlock) Clone() content.Bloker {
	clone := *b
	return &clone
}
