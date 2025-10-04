package ctype

import (
	"bitbucket.org/hotelplan/webcc-content/cms/content"
)

//////////////////////////////
//////////   USPs   //////////
//////////////////////////////

type USPBar struct {
	// content.Component   `name:"USPBar" json:"-"`
	content.ContentType `type:"USPBar" folder:"" json:"-"`
	Headline            string       `json:"headline"`
	Link                Link         `json:"link"`
	LinkPtr             *Link        `json:"linkptr"`
	LinkLabel           string       `json:"linkLabel"`
	Items               []USPBarItem `json:"items"`
}

func (ct *USPBar) Clone() interface{} {
	lp := *ct.LinkPtr
	return &USPBar{
		Headline:  ct.Headline,
		Link:      ct.Link,
		LinkPtr:   &lp,
		LinkLabel: ct.LinkLabel,
		Items:     append(ct.Items[:0:0], ct.Items...),
	}
}

type USPBarItem struct {
	Key   string `json:"key"`
	Icon  string `json:"icon"`
	Label string `json:"label"`
}

type USPCards struct {
	// content.Component   `name:"USPCards" json:"-"`
	content.ContentType `type:"USPCards" folder:"" json:"-"`
	Cards               []USP `json:"cards"`
}

func (ct *USPCards) Clone() interface{} { return USPCards{Cards: append(ct.Cards[:0:0], ct.Cards...)} }

type USPColumns struct {
	// content.Component   `name:"USPColumns" json:"-"`
	content.ContentType `type:"USPColumns" folder:"" json:"-"`
	Left                USP `json:"left"`
	Right               USP `json:"right"`
}

func (ct *USPColumns) Clone() interface{} { return USPColumns{Left: ct.Left, Right: ct.Right} }

type USP struct {
	Key         string `json:"key,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Title       string `json:"title"`
	Subtitle    string `json:"subtitle,omitempty"`
	Description string `json:"description,omitempty"`
}

//////////

type USPBarBlock struct {
	content.Component `name:"USPBarBlock" json:"-"`
	Bar               content.EntryRef `json:"bar,omitempty"`
}

func (b *USPBarBlock) Refs() []*content.EntryRef { return []*content.EntryRef{&b.Bar} }

func (b *USPBarBlock) Clone() content.Bloker {
	copy := *b
	return &copy
}

//////////

type USPCardsBlock struct {
	content.Component `name:"USPCardsBlock" json:"-"`
	Cards             content.EntryRef `json:"cards,omitempty"`
}

func (b *USPCardsBlock) Clone() content.Bloker {
	copy := *b
	return &copy
}

func (b *USPCardsBlock) Refs() []*content.EntryRef { return []*content.EntryRef{&b.Cards} }

//////////

type USPColumnsBlock struct {
	content.Component `name:"USPColumnsBlock" json:"-"`
	Columns           content.EntryRef `json:"columns,omitempty"`
}

func (b *USPColumnsBlock) Clone() content.Bloker {
	copy := *b
	return &copy
}

func (b *USPColumnsBlock) Refs() []*content.EntryRef { return []*content.EntryRef{&b.Columns} }
