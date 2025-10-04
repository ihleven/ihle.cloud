package ctype

import (
	"encoding/json"

	"bitbucket.org/hotelplan/webcc-content/cms/content"
)

// Teaser ist ein gemeinsamer Datentyp, der in mehreren Blöcken auf Start und Nebenseiten verwendet wird
type Teaser struct {
	Asset         content.Asset `assetType:"image" json:"image"`
	AssetWinter   content.Asset `assetType:"image" json:"imageWinter"`
	AlternateText string        `json:"alt"`
	Title         string        `json:"title"`
	Subtitle      string        `json:"subtitle"`
	Link          Link          `json:"link"`
	Virtpath      string        `json:"virtpath"`
}

func (t *Teaser) Links() []*content.EntryLink {
	return []*content.EntryLink{&t.Link.Link.Href}
}

func (t *Teaser) ResolveLinks(locale string, f func(string, string) string) {

	if t.Link.Link.Href.EntryID != "" {
		t.Link.Link.Href.ResolveLinks(locale, f)
	}
}

///////////////////////////
////////// LINK: //////////
///////////////////////////

type Link struct {
	content.Link

	// spezifiziert, ob Link als <a> oder als <NuxtLink> gerendert werden soll
	Anchor bool `json:"anchor"`
}

func (l Link) MarshalJSON() ([]byte, error) {

	type aux struct {
		content.Link
		Anchor bool `json:"anchor"`
	}

	return json.Marshal(aux{
		Link: l.Link, Anchor: l.Anchor,
	})
}

func (l *Link) UnmarshalJSON(data []byte) error {

	var aux struct {
		Href   content.EntryLink `json:"href"`
		Target string            `json:"target"`
		Anchor bool              `json:"anchor"`
	}
	err := json.Unmarshal(data, &aux)
	if err != nil {
		return err
	}

	l.Link = content.Link{Href: aux.Href, Target: aux.Target}
	l.Anchor = aux.Anchor
	return nil
}

type LabelledLink struct {
	Label string `json:"label"`
	URL   string `json:"href"`
}

type IconLabelledLink struct {
	Label string `json:"label"`
	URL   string `json:"href"`
	Icon  string `json:"icon"`
}
