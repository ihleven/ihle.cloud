package art

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"bitbucket.org/hotelplan/webcc-content/cms/mgmt/search"
	"github.com/blevesearch/bleve/v2"
	blevesearch "github.com/blevesearch/bleve/v2/search"
	"github.com/blevesearch/bleve/v2/search/query"
)

type ExtParams struct {
	search.Params

	Ort        string `category:"term" schema:"ort"   field:"exhibition.ort"`
	Genre      string `category:"term" schema:"genre" field:"artwork.genre"`
	Form       string `category:"term" schema:"form"  field:"artwork.form"`
	Geburtstag string ` schema:"geburtstag"  field:"person.geburtstag"`
	Todestag   string ` schema:"todestag"  field:"person.todestag"`
}

func parseext(p ExtParams) func(*bleve.SearchRequest) {
	return func(req *bleve.SearchRequest) {

		// if p.Genre != "" {
		// 	termquery := bleve.NewTermQuery(p.Genre)
		// 	termquery.SetField("genre")
		// 	req.Query.(*query.BooleanQuery).AddMust(termquery)
		// }

		// if p.Form != "" {
		// 	termquery := bleve.NewTermQuery(p.Form)
		// 	termquery.SetField("artwork.form")
		// 	req.Query.(*query.BooleanQuery).AddMust(termquery)
		// }
		if p.Geburtstag != "" {
			b := true
			q := bleve.NewDateRangeInclusiveStringQuery(p.Geburtstag, p.Geburtstag, &b, &b)
			q.SetField("person.geburtstag")
			req.Query.(*query.BooleanQuery).AddMust(q)
		}

		val := reflect.ValueOf(p) //.Elem()
		for i := range val.NumField() {
			tag := val.Type().Field(i).Tag
			if cat := tag.Get("category"); cat == "term" {
				if v := val.Field(i).String(); v != "" {

					termquery := bleve.NewTermQuery(v)
					termquery.SetField(tag.Get("field"))
					// must = append(must, termquery)
					req.Query.(*query.BooleanQuery).AddMust(termquery)
				}
			}
			if cat := tag.Get("category"); cat == "or" {
				if v := val.Field(i).Interface(); v != nil {
					if strs, ok := v.([]string); ok && len(strs) > 0 {
						var or []query.Query

						for _, term := range strs {
							termquery := bleve.NewTermQuery(term)
							termquery.SetField(tag.Get("field"))
							or = append(or, termquery)
						}

						// must = append(must, query.NewDisjunctionQuery(or))
						req.Query.(*query.BooleanQuery).AddMust(query.NewDisjunctionQuery(or))

					}

				}
			}
		}
	}
}

type SearchParams struct {
	ID string

	Query string

	Title string
	Year  int

	Form    string
	Medium  string
	Support string

	Phase string

	Size      int
	Offset    int
	Fields    []string
	Facets    []string
	Highlight bool
	Explain   bool
}

func Search(params SearchParams) (*SearchResult, error) {
	req, err := newSearchRequest(params)
	if err != nil {
		return nil, err
	}

	// res, err := e.index.Search(req)
	// if err != nil {
	// 	return nil, err
	// }
	res := bleve.SearchResult{}

	result := SearchResult{Total: res.Total, Took: res.Took.String(), Hits: res.Hits, Facets: res.Facets, Aggs: buildaggs(res.Facets)}
	if params.Explain {
		result.Cost = res.Cost
		result.MaxScore = res.MaxScore
		result.Request = req
		result.Status = res.Status
	}
	return &result, nil
}

type SearchResult struct {
	Total    uint64                              `json:"total_hits"`
	Took     string                              `json:"took,omitempty"`
	Cost     uint64                              `json:"cost,omitempty"`
	MaxScore float64                             `json:"max_score,omitempty"`
	Hits     blevesearch.DocumentMatchCollection `json:"hits"`
	Facets   blevesearch.FacetResults            `json:"facets,omitempty"`
	Status   *bleve.SearchStatus                 `json:"status,omitempty"`
	Request  *bleve.SearchRequest                `json:"request,omitempty"`
	Aggs     aggs                                `json:"aggs,omitempty"`
}

func newSearchRequest(params SearchParams) (*bleve.SearchRequest, error) {

	query := bleve.NewConjunctionQuery()

	if params.Medium != "" {
		termquery := bleve.NewTermQuery(params.Medium)
		termquery.SetField("medium")
		query.AddQuery(termquery)
	}

	if params.Support != "" {
		termquery := bleve.NewTermQuery(params.Support)
		termquery.SetField("support")
		query.AddQuery(termquery)
	}

	if params.Form != "" {
		termquery := bleve.NewTermQuery(params.Form)
		termquery.SetField("form")
		query.AddQuery(termquery)
	}

	if params.Phase != "" {
		termquery := bleve.NewTermQuery(params.Phase)
		termquery.SetField("phase")
		query.AddQuery(termquery)
	}

	if params.Title != "" {
		termquery := bleve.NewTermQuery(params.Title)
		termquery.SetField("title")
		query.AddQuery(termquery)
	}

	if params.Year != 0 {
		min, max := float64(params.Year), float64(params.Year+1)
		numrangequery := bleve.NewNumericRangeQuery(&min, &max)
		numrangequery.SetField("year")
		query.AddQuery(numrangequery)
	}

	if params.Query != "" {
		matchquery := bleve.NewMatchQuery(params.Query)
		// termquery.SetField("title")
		query.AddQuery(matchquery)
	}

	if len(query.Conjuncts) == 0 {
		query.AddQuery(bleve.NewMatchAllQuery())
	}
	for _, c := range query.Conjuncts {

		fmt.Printf("conjunct: %#v\n", c)
	}

	searchRequest := bleve.NewSearchRequestOptions(query, params.Size, params.Offset, params.Explain)
	if params.Highlight {
		searchRequest.Highlight = bleve.NewHighlight()
	}

	for _, field := range params.Fields {
		searchRequest.Fields = append(searchRequest.Fields, strings.Split(field, ",")...)
	}

	facets := facets()
	for _, facet := range params.Facets {
		if facet == "*" {
			facet = "forms,support,media,years"
		}
		for _, facet := range strings.Split(facet, ",") {
			if facetRequest, ok := facets[facet]; ok {
				searchRequest.AddFacet(facet, facetRequest)
			}
		}
	}

	return searchRequest, nil
}

func facets() map[string]*bleve.FacetRequest {
	yearsRequest := bleve.NewFacetRequest("year", 50)
	for i := range 50 {
		min, max := float64(1980+i), float64(1980+i+1)
		yearsRequest.AddNumericRange(strconv.Itoa(1980+i), &min, &max)
	}

	return map[string]*bleve.FacetRequest{
		"years":   yearsRequest,
		"media":   bleve.NewFacetRequest("medium", 20),
		"support": bleve.NewFacetRequest("support", 20),
		"forms":   bleve.NewFacetRequest("form", 20),
		"status":  bleve.NewFacetRequest("status", 20),
	}
}

type aggs map[string]counts
type counts map[string]int

func buildaggs(facets blevesearch.FacetResults) aggs {

	a := make(aggs)

	for name, facet := range facets {
		a[name] = make(counts)
		a[name]["_total"] = facet.Total
		a[name]["_missing"] = facet.Missing
		a[name]["_other"] = facet.Other
		for _, term := range facet.Terms.Terms() {
			a[name][term.Term] = term.Count
		}

	}
	return a
}
