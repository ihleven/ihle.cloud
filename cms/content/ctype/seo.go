package ctype

import (
	"bitbucket.org/hotelplan/webcc-content/cms/content"
	"bitbucket.org/hotelplan/webcc-content/cms/permission"
	"bitbucket.org/hotelplan/webcc-content/cms/pkg/errors"
)

type Seo struct {
	content.ContentType `type:"Seo" json:"-" folder:"entries" mimetype:"application/json"`

	DetailPage  SeoPageData `json:"detailPage"`
	ReviewsPage SeoPageData `json:"reviewsPage"`

	Search []SeoSearchData `json:"search"`
}

func (s *Seo) Clone() interface{} {
	clone := *s
	clone.Search = make([]SeoSearchData, len(s.Search))
	for i, v := range s.Search {
		searchdata := v
		searchdata.SeoSearchPage.Filters = make(map[string]string)
		for k, w := range v.SeoSearchPage.Filters {

			searchdata.SeoSearchPage.Filters[k] = w
		}
		searchdata.H1 = append(v.H1[:0:0], v.H1...)
		clone.Search[i] = searchdata
	}

	return &clone
}

type SeoSearchData struct {
	Link string `json:"link"`
	SeoSearchPage
	SeoPageData
	H1 []string `json:"h1"`
	H2 string   `json:"h2"`
}
type SeoSearchPage struct {
	CountryCode string            `json:"country"`
	RegionCode  string            `json:"region"`
	PlaceCode   string            `json:"place"`
	Filters     map[string]string `json:"filters"`
}
type SeoPageData struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Robots      string `json:"robots"`
}

// func PageURL(c, r, p string, filters map[string]interface{}) url.URL {

// 	query := url.Values{}
// 	if c != "" {
// 		query.Add("country", c)
// 	}
// 	if r != "" {
// 		query.Add("region", r)
// 	}
// 	if p != "" {
// 		query.Add("place", p)
// 	}
// 	for k, v := range filters {
// 		query.Add(k, fmt.Sprintf("%v", v))
// 	}

// 	link := url.URL{
// 		Scheme:   "page",
// 		Path:     "search",
// 		RawQuery: query.Encode(),
// 	}
// 	fmt.Println("PAGEURL:", link.String())
// 	return link
// }

var (
	SEO_EDIT      permission.Type = permission.Define("seo", "edit", false, false)
	SEO_TREE_EDIT permission.Type = permission.Define("seo", "tree.edit", false, false) // currently used only in frontend setting json edtor mode
)

func (s *Seo) CheckPermissionForWrite(scope permission.Scope, previous *content.Entry) error {

	if scope.Denies(SEO_EDIT) {
		return errors.New("missing seo edit permission")
	}
	return nil
}
