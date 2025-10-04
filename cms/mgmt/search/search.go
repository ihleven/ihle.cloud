package search

import (
	"encoding/json"
	"fmt"
	"time"

	"bitbucket.org/hotelplan/webcc-content/cms/content"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
)

// func (e *Engine) SearchString(q string) (*bleve.SearchResult, error) {
// 	query := bleve.NewQueryStringQuery(q)
// 	searchRequest := bleve.NewSearchRequest(query)
// 	searchResult, err := e.BleveIndex.Search(searchRequest)
// 	if err != nil {
// 		fmt.Println(err)
// 		return nil, err
// 	}
// 	return searchResult, nil
// }

func Facet(name string, field string, size int) func(*bleve.SearchRequest) {
	return func(req *bleve.SearchRequest) {
		req.AddFacet(name, bleve.NewFacetRequest(field, size))
	}
}

func Verbose() func(*bleve.SearchRequest) {
	return func(req *bleve.SearchRequest) {
		req.Explain = true
	}
}

func (e *Engine) Search(p Params, options ...func(*bleve.SearchRequest)) (*bleve.SearchResult, error) {

	searchreq, err := p.createSearchRequest()
	if err != nil {
		return nil, err
	}

	for _, applyopt := range options {
		applyopt(searchreq)
	}

	if searchreq.Explain {
		bytes, _ := json.MarshalIndent(searchreq, "", "    ")
		fmt.Printf("searchrequest: %s - engine: %v\n", bytes, e)
	}

	result, err := e.BleveIndex.Search(searchreq)
	if err != nil {
		return nil, err
	}

	result.Request = searchreq
	return result, nil
}

func NewResult(blevesearchres *bleve.SearchResult) *Result {

	res := Result{
		SearchResult: blevesearchres,
		Total:        blevesearchres.Total,
		Cost:         blevesearchres.Cost,
		MaxScore:     blevesearchres.MaxScore,
		Took:         blevesearchres.Took,
		TookMilli:    blevesearchres.Took.String(),
		Hits:         make([]DocumentMatch, len(blevesearchres.Hits)),
		Facets:       blevesearchres.Facets,
		// Status:       blevesearchres.Status,
		// Request:      blevesearchres.Request,
	}
	if blevesearchres.Request.Explain {
		res.Status = blevesearchres.Status
		res.Request = blevesearchres.Request
	}
	for i, hit := range blevesearchres.Hits {
		dm := &res.Hits[i]
		// dm.Index = hit.Index
		dm.ID = hit.ID
		dm.Score = hit.Score
		// dm.Sort = hit.Sort
		dm.Expl = hit.Expl
		dm.Fragments = hit.Fragments
		dm.Locations = hit.Locations
		dm.Fields = hit.Fields
		// dm.Doc = Document{
		// 	// Meta: content.Meta{
		// 	// 	Path: hit.Fields["path"].(string),
		// 	// },
		// 	// DocumentContent: DocumentContent{
		// 	// 	Image: hit.Fields["content.image"].(string),
		// 	// 	Title: hit.Fields["content.Title"].(map[string]string),
		// 	// 	Text:  hit.Fields["content.text"].(map[string]string),
		// 	// },
		// }
		// if img, ok := hit.Fields["content.image"].(string); ok {
		// 	d.Doc.DocumentContent.Image = img
		// }
		// if title, ok := hit.Fields["content.title"].(map[string]string); ok {
		// 	d.Doc.DocumentContent.Title = title
		// }
		// if text, ok := hit.Fields["content.text"].(map[string]string); ok {
		// 	d.Doc.DocumentContent.Text = text
		// }
	}
	return &res
}

type Result struct {
	*bleve.SearchResult
	Total     uint64        `json:"total_hits"`
	Cost      uint64        `json:"cost"`
	MaxScore  float64       `json:"max_score"`
	Took      time.Duration `json:"took"`
	TookMilli string        `json:"tookms"`
	// Hits     search.DocumentMatchCollection `json:"hits"`
	Hits   []DocumentMatch     `json:"hits"`
	Facets search.FacetResults `json:"facets"`

	Status *bleve.SearchStatus `json:"status"`

	Request *bleve.SearchRequest `json:"request,omitempty"`
}

type DocumentMatch struct {
	// Index string `json:"index,omitempty"`
	ID        string                      `json:"id"`
	Score     float64                     `json:"score"`
	Expl      *search.Explanation         `json:"explanation,omitempty"`
	Locations search.FieldTermLocationMap `json:"locations,omitempty"`
	Fragments search.FieldFragmentMap     `json:"fragments,omitempty"`
	Sort      []string                    `json:"sort,omitempty"`
	Meta      *content.Meta               `json:"entry,omitempty"`
	Entry     *content.Entry              `json:"entry,omitempty"`
	Fields    map[string]interface{}      `json:"fields,omitempty"`
}

func Fields2Meta(fields map[string]interface{}) *content.Meta {
	meta := &content.Meta{}
	if path, ok := fields["path"].(string); ok {
		meta.Path = path
	}
	if name, ok := fields["name"].(string); ok {
		meta.Name = name
	}
	if typ, ok := fields["type"].(string); ok {
		meta.Type = typ
	}
	if locale, ok := fields["locale"].(string); ok {
		meta.Locale = locale
	}
	if version, ok := fields["version"].(string); ok {
		meta.Version = version
	}
	if slug, ok := fields["slug"].(string); ok {
		meta.Slug = slug
	}
	if fullslug, ok := fields["full_slug"].(string); ok {
		meta.FullSlug = fullslug
	}
	return meta
}
