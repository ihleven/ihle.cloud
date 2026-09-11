package familie

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/interhome-group/cms/mgmt/search"
	"github.com/interhome-group/cms/pkg/errs"

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

func (a *api) PersonHandler(w http.ResponseWriter, r *http.Request) error {
	path := fmt.Sprintf("familie/%s.md", r.PathValue("person"))
	fmt.Println("person:", path)
	entry, err := a.repo.GetEntry(path)
	if err != nil {
		return err
	}
	fmt.Println("person:", entry)
	return respond(w, entry)
}

func (a *api) ReiseHandler(w http.ResponseWriter, r *http.Request) error {

	entry, err := a.repo.GetEntry(fmt.Sprintf("urlaub/%s.md", r.PathValue("key")))
	if err != nil {
		return err
	}

	return respond(w, entry)
}

func Adapt(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			status := errs.StatusOf(err)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(status)
			enc := json.NewEncoder(w)
			enc.SetIndent("", "    ")
			enc.Encode(err)
		}
	}
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as HTTP handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// [Handler] that calls f.
type ErrorHandlerFunc func(http.ResponseWriter, *http.Request) error

// ServeHTTP calls f(w, r).
func (f ErrorHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := f(w, r)
	if err != nil {
		status := errs.StatusOf(err)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		enc := json.NewEncoder(w)
		enc.SetIndent("", "    ")
		enc.Encode(err)
	}
}

func respond(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	err := enc.Encode(data)
	return err
}
