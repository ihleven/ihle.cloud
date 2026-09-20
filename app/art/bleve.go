package art

import (
	"reflect"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/interhome-group/cms/mgmt/search"
)

// ExtParams are the art-specific filters on top of the CMS's own search
// parameters: each field names the indexed field it narrows, and parseext turns
// the ones that were set into term queries.
//
// This file used to carry a second, complete search implementation as well —
// its own Search, SearchResult and `aggs` facet shape. Nothing called it, and
// the /werke page had been written against its `aggs` rather than against what
// the mounted handler actually returns, which is why the page's counts never
// appeared. It is gone; api.go calls the CMS engine directly.
type ExtParams struct {
	search.Params

	Ort     string `category:"term" schema:"ort"     field:"exhibition.ort"`
	Genre   string `category:"term" schema:"genre"   field:"artwork.genre"`
	Form    string `category:"term" schema:"form"    field:"artwork.form"`
	Medium  string `category:"term" schema:"medium"  field:"artwork.medium"`
	Support string `category:"term" schema:"support" field:"artwork.support"`
	// The archive's own status. "DEL" is a work the old database marked as
	// deleted; it is imported only with --deleted, and this is how you find it.
	Status     string `category:"term" schema:"status"  field:"artwork.status"`
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
