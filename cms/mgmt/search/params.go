package search

import (
	"reflect"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
)

type Params struct {
	Query       string `schema:"q" json:"q,omitempty"`
	PhraseQuery string `schema:"pq" json:"pq,omitempty"`
	Suggest     string `schema:"suggest" json:"suggest,omitempty"`

	Facets []string `json:"facets,omitempty" schema:"facet"`

	// Meta-Parameter, die für die Aufbereitung/Darstellung des Suchergebnisses relevant sind
	Language  string `json:"language,omitempty"` // SearchAccom in dieser Sprache
	Field     []string
	Fields    string
	Highlight bool
	Verbose   bool

	// Suchparameter
	Page     int      `json:"page,omitempty"`               // kein Filter -> SearchParam vs SearchFilter
	PageSize int      `json:"size,omitempty" schema:"size"` // kein Filter
	From     int      `json:"from,omitempty"`               // kein Filter
	Sorting  []string `json:"sorting,omitempty"`            // kein Filter -> feldname bzw. -feldname

	Type    []string `json:"type,omitempty"    category:"or"   schema:"type"     field:"type"`
	Version string   `json:"version,omitempty" category:"term" schema:"version"  field:"version"`
	Status  string   `json:"status,omitempty"  category:"term" schema:"status"   field:"status"`
	Tag     string   `json:"tag,omitempty"     category:"term" schema:"tag"      field:"tags"`

	Path     string   `json:"path,omitempty"     category:"term" schema:"path"     field:"path"`
	FullSlug string   `json:"slugs,omitempty"    category:"term" schema:"slugs"    field:"full_slug"`
	Dir      string   `json:"dir,omitempty"`
	Ancestor string   `json:"ancestor,omitempty" category:"term" schema:"ancestor" field:"ancestors"`
	ID       []string `json:"id,omitempty"       category:"or" schema:"id"       field:"id"`
	Locale   string   `json:"locale,omitempty"   category:"term" schema:"locale"   field:"locale"`
}

func (p *Params) createQuery() (query.Query, error) {
	// fmt.Printf("create Query params: %#v\n", p)
	var must, should, mustnot []query.Query

	var language string = "de"
	for _, lang := range []string{"de", "en", "fr", "nl"} {
		if strings.ToLower(p.Language) == lang {
			language = lang
		}
	}

	if p.Query != "" {
		query := query.NewMatchQuery(p.Query)
		query.SetField("text." + language)
		query.Analyzer = language
		must = append(must, query)
	}
	if p.PhraseQuery != "" {
		query := query.NewMatchPhraseQuery(p.PhraseQuery)
		query.SetField("text." + language)
		query.Analyzer = language
		should = append(should, query)
	}
	if p.Suggest != "" {
		query := query.NewPrefixQuery(p.Suggest)
		query.SetField("text." + language)
		// query.Analyzer = bleve.NewKeywordFieldMapping().Analyzer
		must = append(must, query)
	}

	if p.Dir != "" {
		query := bleve.NewTermQuery(p.Dir) //NewPrefixQuery(p.Dir) // PrefixQuery
		query.SetField("dir")
		must = append(must, query)
	}

	val := reflect.ValueOf(p).Elem()
	for i := range val.NumField() {
		tag := val.Type().Field(i).Tag
		if cat := tag.Get("category"); cat == "term" {
			if v := val.Field(i).String(); v != "" {

				termquery := bleve.NewTermQuery(v)
				termquery.SetField(tag.Get("field"))
				must = append(must, termquery)
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

					must = append(must, query.NewDisjunctionQuery(or))
				}

			}
		}
	}
	if len(must) == 0 && len(should) == 0 && len(mustnot) == 0 {
		must = append(must, bleve.NewMatchAllQuery())
	}
	return query.NewBooleanQuery(must, should, mustnot), nil
}

func (p *Params) createSearchRequest() (*bleve.SearchRequest, error) {

	// start := time.Now()

	query, err := p.createQuery()
	if err != nil {
		return nil, err
	}

	// if size == 0 {
	// 	size = 20
	// }
	size := p.PageSize
	from := p.From
	if p.Page > 0 {
		from += (p.Page - 1) * size
	}
	searchreq := bleve.NewSearchRequestOptions(query, size, from, p.Verbose)

	// Fields
	searchreq.SortBy(p.Sorting)
	if p.Fields != "" {
		searchreq.Fields = strings.Split(p.Fields, ",")
	}
	if len(p.Field) > 0 {
		searchreq.Fields = append(searchreq.Fields, p.Field...)
	}

	// Highlighting
	if p.Highlight {
		searchreq.Highlight = bleve.NewHighlight()
		searchreq.Highlight.AddField("text.de")
	}
	if p.Verbose {
		//
	}

	// bytes, _ := json.MarshalIndent(searchreq, "", "    ")
	// fmt.Printf(" * created bleve.SearchRequest %s in %s", bytes, time.Since(start))

	// return searchreq, nil

	for _, f := range p.Facets {
		Facet(f, f, 20)(searchreq)
	}
	// searchreq.AddFacet("attrs", bleve.NewFacetRequest("content.attrs", 100))
	// searchreq.AddFacet("type", bleve.NewFacetRequest("type", 20))
	// searchreq.AddFacet("dest", bleve.NewFacetRequest("dest", 20))
	// searchreq.AddFacet("filter", bleve.NewFacetRequest("filter", 100))
	// searchreq.AddFacet("attrs", bleve.NewFacetRequest("attrs", 20))

	// bytes, _ := json.MarshalIndent(searchreq, "", "    ")
	// fmt.Printf("query: %s\n", bytes)
	// fmt.Printf(" * created bleve.SearchRequest %s in %s", bytes, time.Since(start))

	return searchreq, nil
}
