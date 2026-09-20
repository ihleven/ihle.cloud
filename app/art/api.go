package art

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/interhome-group/cms/mgmt/search"
	"github.com/interhome-group/cms/pkg/errs"

	"github.com/gorilla/schema"
	"github.com/interhome-group/cms/mgmt/gitrepo"
)

func NewApi(repo *gitrepo.Sitory, eng *search.Engine) *api {

	api := &api{repo: repo, engine: eng}

	return api
}

type api struct {
	repo   *gitrepo.Sitory
	engine *search.Engine
}

func (a *api) ArtworkHandler(w http.ResponseWriter, r *http.Request) error {

	entry, err := a.repo.GetEntry(fmt.Sprintf("artworks/%s.json", r.PathValue("id")))
	if err != nil {
		return respond(w, r, errs.StatusOf(err), err)
	}

	return respond(w, r, 200, entry)
}

func (a *api) Search(w http.ResponseWriter, r *http.Request) error {

	params := ExtParams{Params: search.Params{PageSize: 20, Fields: "*"}}
	enc := schema.NewDecoder()
	// Unknown query keys are the client's business, not an error: the page
	// sends every filter it has and leaves the empty ones out.
	enc.IgnoreUnknownKeys(true)
	enc.Decode(&params, r.URL.Query())

	// Refused rather than answered emptily. Everything this searches on — the
	// artwork sub-document and its facets — is written by AugmentSearchDoc,
	// which the indexer calls only at fulltext and above; at basic it returns
	// the plain entry document first. So below fulltext every filter matches
	// nothing and every facet comes back with no terms, and /werke renders a
	// working page with no content and no reason given.
	//
	// 501 because it is the deployment that cannot answer, not the request that
	// was wrong — the same answer the CMS gives for reads that need an index it
	// was not started with.
	if a.engine.Level < search.Fulltext {
		return respond(w, r, http.StatusNotImplemented, errs.New(
			"art search needs SEARCH_LEVEL fulltext or extended, this one runs at %s",
			a.engine.Level, errs.HTTPStatus(501)))
	}

	// Facet names are the page's, not the field's: /werke reads them back by
	// name to put a count beside each filter.
	searchResult, err := a.engine.Search(params.Params,
		search.Facet("forms", "artwork.form", 10),
		search.Facet("media", "artwork.medium", 20),
		search.Facet("support", "artwork.support", 20),
		search.Facet("status", "artwork.status", 10),
		search.Facet("ort", "exhibition.ort", 100),
		search.Facet("type", "type", 100),

		parseext(params),
	)
	if err != nil {
		return respond(w, r, 500, err)
	}

	return respond(w, r, 200, searchResult)
}

func (a *api) MappingHandler(w http.ResponseWriter, r *http.Request) error {

	return respond(w, r, 200, a.engine.Mapping())
}

func respond(w http.ResponseWriter, r *http.Request, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	err := enc.Encode(data)
	return err
}
