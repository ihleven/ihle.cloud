package ctype

import "bitbucket.org/hotelplan/webcc-content/cms/content"

type Columns struct {
	content.Component `name:"Columns" json:"-"`
	Left              content.Bloks `json:"left"`
	Right             content.Bloks `json:"right"`
}

func (b *Columns) Clone() content.Bloker {
	return &Columns{Left: b.Left.Clone(), Right: b.Right.Clone()}
}

type SearchBar struct {
	content.Component `name:"SearchBar" json:"-"`
	Images            []Image `json:"images"`
}

func (b *SearchBar) Clone() content.Bloker {
	clone := SearchBar{Images: make([]Image, len(b.Images))}
	copy(clone.Images, b.Images)
	return &clone
}

type Breadcrumbs struct {
	content.Component `name:"Breadcrumbs" json:"-"`
	Links             []LabelledLink `json:"links"`
}

func (b *Breadcrumbs) Clone() content.Bloker {
	clone := Breadcrumbs{Links: make([]LabelledLink, len(b.Links))}
	copy(clone.Links, b.Links)
	return &clone
}

type Heading struct {
	content.Component `name:"Heading" json:"-"`
	Title             string `json:"title"`
	Level             int    `json:"level"`
	Size              string `json:"size"`
}

func (b *Heading) Clone() content.Bloker {
	clone := *b
	return &clone
}

type Menu struct {
	content.Component `name:"Menu" json:"-"`

	Content content.Bloks    `json:"content"`
	Menu    content.EntryRef `json:"menu"`
}

func (b *Menu) Clone() content.Bloker {
	return &Menu{Content: b.Content.Clone(), Menu: content.EntryRef{Path: b.Menu.Path}}
}

func (b *Menu) Refs() []*content.EntryRef { return []*content.EntryRef{&b.Menu} }

type Richtext struct {
	content.Component `name:"Richtext" json:"-"`
	Text              string `json:"text"`
	Markdown          string `json:"markdown"`
}

func (b *Richtext) Clone() content.Bloker {
	clone := *b
	return &clone
}

type TeaserGrid struct {
	content.Component `name:"TeaserGrid" json:"-"`
	Elements          []Teaser `json:"elements"`
}

func (b *TeaserGrid) Clone() content.Bloker {
	clone := TeaserGrid{Elements: make([]Teaser, len(b.Elements))}
	copy(clone.Elements, b.Elements)
	return &clone
}

type Acc struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

type ObjectList struct {
	content.Component `name:"ObjectList" json:"-"`
	Title             string `json:"title"`
	From              string `json:"from"`
	To                string `json:"to"`
	Objects           []Acc  `json:"objects"`
	Layout            string `json:"layout"`
	LabelsBGColor     string `json:"labelsBGColor"`
	LabelsTextColor   string `json:"labelsTextColor"`
}

func (b *ObjectList) Clone() content.Bloker {
	clone := *b
	clone.Objects = make([]Acc, len(b.Objects))
	copy(clone.Objects, b.Objects)
	return &clone
}

type LSOList struct {
	content.Component `name:"LSOList" json:"-"`
	Offices           map[string]LSO `json:"offices"`
}

func (b *LSOList) Clone() content.Bloker {
	clone := LSOList{Offices: make(map[string]LSO)}
	for key, lso := range b.Offices {
		clone.Offices[key] = lso
	}
	return &clone
}

type LSO struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	CountryCode string `json:"countryCode"`
	CountryName string `json:"countryName"`
}

type LSODetail struct {
	content.Component `name:"LSODetail" json:"-"`
	Country           string     `json:"country"`
	Images            []Image    `json:"images"`
	Text              string     `json:"text"`
	Markdown          string     `json:"markdown"`
	Code              string     `json:"code"`
	Type              string     `json:"type"`
	Name              string     `json:"name"`
	WwwCapability     string     `json:"wwwCapability"`
	Address           LSOAddress `json:"address"`
}

func (b *LSODetail) Clone() content.Bloker {
	clone := *b
	clone.Images = make([]Image, len(b.Images))
	copy(clone.Images, b.Images)
	return &clone
}

type LSOAddress struct {
	Address1  string  `json:"address1"`
	Address2  string  `json:"address2"`
	Street    string  `json:"street"`
	Country   string  `json:"country"`
	Zip       string  `json:"zip"`
	Place     string  `json:"place"`
	Phone     string  `json:"phone"`
	Email     string  `json:"email"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// Used to display a block with an image and text on the LSO pages
type LSOMedia struct {
	content.Component `name:"LSOMedia" json:"-"`
	Picture           Image  `json:"image"`
	Text              string `json:"text"`
	Markdown          string `json:"markdown"`
}

func (b *LSOMedia) Clone() content.Bloker {
	clone := *b
	return &clone
}
