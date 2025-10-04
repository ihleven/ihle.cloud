package art

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"bitbucket.org/hotelplan/webcc-content/cms/mgmt/search"
	"bitbucket.org/hotelplan/webcc-content/cms/pkg/errors"

	"bitbucket.org/hotelplan/webcc-content/cms/mgmt/gitrepo"
	"github.com/blevesearch/bleve/v2"
	"github.com/gorilla/schema"
)

func NewApi(repo *gitrepo.Sitory, eng *search.Engine) *api {

	api := &api{repo: repo, engine: eng}

	return api
}

type api struct {
	repo   *gitrepo.Sitory
	engine *search.Engine
}

func (a *api) Handler(origins ...string) http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc("GET /art-api/hello/{id}", a.HelloHandler)
	mux.HandleFunc("GET /art-api/artworks/{id}", a.ArtworkHandler)
	mux.HandleFunc("GET /art-api/search", a.Search)
	mux.HandleFunc("GET /art-api/search/mapping", a.MappingHandler)

	// corsoptions := &cors.Options{
	// 	AllowedOrigins:   origins,
	// 	AllowCredentials: true,
	// 	AllowedHeaders:   []string{"JWT", "authorization", "cmsauth", "Cookie"},
	// 	AllowedMethods:   []string{"PUT", "POST", "GET", "DELETE"},
	// 	// Enable Debugging for testing, consider disabling in production
	// 	Debug: false,
	// }

	// return cors.New(corsoptions).Handler(a)

	// if corsoptions != nil {
	return mux //cors.New(*corsoptions).Handler(mux)
	// }
	// return http.Handler(a.mux)
}

func (a *api) HelloHandler(w http.ResponseWriter, r *http.Request) {
	what := r.PathValue("id")
	if what == "" {
		what = "World"
	}

	respond(w, r, 200, fmt.Sprintf("Hello %s!", what))
}

func (a *api) ArtworkHandler(w http.ResponseWriter, r *http.Request) {

	entry, err := a.repo.GetEntry(fmt.Sprintf("artworks/%s.json", r.PathValue("id")))
	if err != nil {
		respond(w, r, errors.Code(err), err)
		return
	}

	respond(w, r, 200, entry)
}

func (a *api) Search(w http.ResponseWriter, r *http.Request) {

	params := ExtParams{Params: search.Params{PageSize: 20, Fields: "*"}}
	enc := schema.NewDecoder()
	enc.Decode(&params, r.URL.Query())

	searchResult, err := a.engine.Search(params.Params,
		search.Facet("form", "artwork.form", 10),
		search.Facet("ort", "exhibition.ort", 100),
		// search.Facet("filter", "filter", 100),
		search.Facet("type", "type", 100),

		parseext(params),
	)
	if err != nil {
		respond(w, r, 500, err)
		return
	}

	respond(w, r, 200, searchResult)
}

func (a *api) MappingHandler(w http.ResponseWriter, r *http.Request) {

	respond(w, r, 200, a.engine.Mapping())
}

func (a *api) SearchHandler(w http.ResponseWriter, r *http.Request) {

	highlight := r.URL.Query().Has("highlight")

	query := bleve.NewConjunctionQuery()

	for _, param := range []string{"id", "medium", "support", "title", "form", "phase"} {
		if p := r.URL.Query().Get(param); p != "" {
			termquery := bleve.NewTermQuery(p)
			termquery.SetField(param)
			query.AddQuery(termquery)
		}
	}

	if year, err := strconv.Atoi(r.URL.Query().Get("year")); err == nil {
		min, max := float64(year), float64(year+1)
		numrangequery := bleve.NewNumericRangeQuery(&min, &max)
		numrangequery.SetField("year")
		query.AddQuery(numrangequery)
	}

	if q := r.URL.Query().Get("q"); q != "" {

		mq := bleve.NewMatchQuery(q)
		query.AddQuery(mq)
	}

	if len(r.URL.Query()) == 0 {
		query.AddQuery(bleve.NewMatchAllQuery())
	}
	for _, c := range query.Conjuncts {

		fmt.Printf("query: %#v\n", c)
	}

	searchRequest := bleve.NewSearchRequestOptions(query, 1000, 0, false)
	if highlight {
		searchRequest.Highlight = bleve.NewHighlight()
	}

	searchRequest.Fields = []string{"*"}

	f := bleve.NewFacetRequest("year", 50)
	for i := range 50 {
		min := float64(1980 + i)
		max := float64(1980 + i + 1)
		// fmt.Println(min, max)
		f.AddNumericRange(strconv.Itoa(1980+i), &min, &max)
	}
	searchRequest.AddFacet("years", f)
	searchRequest.AddFacet("media", bleve.NewFacetRequest("medium", 20))
	searchRequest.AddFacet("support", bleve.NewFacetRequest("support", 20))
	searchRequest.AddFacet("forms", bleve.NewFacetRequest("form", 20))
	searchRequest.AddFacet("status", bleve.NewFacetRequest("status", 20))

	// res, err := a.index.Search(searchRequest)
	// if err != nil {
	// 	respond(w, r, 500, err)
	// 	return
	// }
	// respond(w, r, 200, res)
	return //bunrouter.JSON(rw, searchResult)
	// for _, hit := range searchResult.Hits {
	// 	fmt.Fprintf(w, "ID: %s, Score: %f\n", hit.ID, hit.Score)
	// }
}

func respond(w http.ResponseWriter, r *http.Request, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	err := enc.Encode(data)
	return err
}
