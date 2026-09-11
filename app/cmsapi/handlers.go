// Package cmsapi provides the entry and search HTTP handlers this app needs
// from the CMS.
//
// They used to come from the CMS itself, as mgmt/handler.EntryDetails,
// EntryUpdate, EntryLookup and SearchHandler. The CMS has since migrated its
// HTTP layer to typed (huma) controllers registered on its own router, and the
// old plain http.HandlerFunc-style constructors were removed. Rather than adopt
// that router — which would change the response envelope this app's UI reads —
// the four handlers are reimplemented here against the current mgmt.Mngr API,
// preserving the JSON shapes the frontend already consumes: an entry is
// returned as the marshalled content.Entry, a search as the engine's result.
package cmsapi

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/ihleven/ihlvn/pkg/authn"
	"github.com/interhome-group/cms/content"
	"github.com/interhome-group/cms/mgmt"
	"github.com/interhome-group/cms/mgmt/handler"
	"github.com/interhome-group/cms/mgmt/search"
	"github.com/interhome-group/cms/pkg/errs"
)

// contentUser is the user the CMS evaluates for a request.
//
// An unauthenticated request is the anonymous user rather than a bypass: entry
// ACLs are still applied to it, and it is granted whatever an entry grants to
// others. That is what lets the public site read content without the write path
// being open to everyone.
func contentUser(r *http.Request) content.User {
	if account, ok := authn.FromContext(r.Context()); ok {
		return account.ContentUser()
	}
	return content.Anonymous()
}

// EntryDetails returns a single entry by storage path.
//
// A `slugs` query parameter resolves the path through the search index first,
// which is how the public viewer addresses entries by their full slug.
func EntryDetails(mngr *mgmt.Mngr) func(http.ResponseWriter, *http.Request) error {
	type params struct{ Resolve []string }

	return func(w http.ResponseWriter, r *http.Request) error {
		p, err := handler.ParseWithDefaults(r.URL.Query(), params{})
		if err != nil {
			return err
		}

		path := strings.TrimSuffix(r.PathValue("path"), "/")

		if slugs := r.URL.Query().Get("slugs"); slugs != "" {
			if mngr.Engine == nil {
				return errs.New("search not available: cannot resolve slugs", errs.HTTPStatus(http.StatusNotImplemented))
			}
			result, err := mngr.Engine.Search(search.Params{FullSlug: slugs, PageSize: 1})
			if err == nil {
				switch result.Total {
				case 0:
					return errs.New("full_slug not found: %s", slugs, errs.HTTPStatus(http.StatusNotFound))
				case 1:
					path = result.Hits[0].ID
				default:
					return errs.New("full_slug not unique: %s", slugs, errs.HTTPStatus(http.StatusInternalServerError))
				}
			}
		}

		entry, err := mngr.GetEntry(path, contentUser(r), content.ParseResolv(p.Resolve))
		if err != nil {
			return errs.Wrap(err, "GetEntry for path %q failed", path)
		}
		return JSON(w, entry)
	}
}

// EntryLookup finds a single entry by identity rather than by path — by id or
// full slug — and therefore needs the search index.
func EntryLookup(mngr *mgmt.Mngr) func(http.ResponseWriter, *http.Request) error {
	type params struct {
		Resolve []string
		Path    string
		Slugs   string
		ID      []string
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		if mngr.Engine == nil {
			return errs.New("search not implemented", errs.HTTPStatus(http.StatusNotImplemented))
		}

		p, err := handler.ParseWithDefaults(r.URL.Query(), params{Path: r.PathValue("path")})
		if err != nil {
			return err
		}

		result, err := mngr.Engine.Search(search.Params{
			ID:       p.ID,
			Path:     strings.TrimSuffix(p.Path, "/"),
			FullSlug: p.Slugs,
			PageSize: 1,
		})
		if err != nil {
			return err
		}
		switch {
		case result.Total < 1, len(result.Hits) < 1:
			return errs.New("not found: %v", p, errs.HTTPStatus(http.StatusNotFound))
		case result.Total > 1:
			return errs.New("not unique: %d hits", result.Total, errs.HTTPStatus(http.StatusNotFound))
		}

		entry, err := mngr.GetEntry(result.Hits[0].ID, contentUser(r), content.ParseResolv(p.Resolve))
		if err != nil {
			return err
		}
		return JSON(w, entry)
	}
}

// EntryUpdate writes an entry back.
//
// `mode=validate` parses and checks without writing; `mode=force` accepts an
// entry whose bytes do not round-trip exactly. Committing and reindexing are now
// one call — mgmt.SaveChangeset — which also means an indexing failure after a
// successful commit is reported as a warning rather than failing the write.
func EntryUpdate(mngr *mgmt.Mngr) func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		now := time.Now()
		mode := r.URL.Query().Get("mode")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			return errs.New("could not read request body", errs.HTTPStatus(http.StatusBadRequest))
		}

		contentType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if contentType == "" {
			return errs.New("missing Content-Type", errs.HTTPStatus(http.StatusBadRequest))
		}

		entry, err := content.ParseEntry(nil, body, contentType)
		if err != nil {
			return errs.Wrap(err, "parse error", errs.HTTPStatus(http.StatusBadRequest))
		}

		valid, err := entry.ValidateSyntaxIntegrity(body)
		if err != nil {
			return errs.Wrap(err, "validating the submitted bytes failed")
		}
		if mode == "validate" {
			return JSON(w, entry)
		}
		if !valid && mode != "force" {
			return errs.New("the entry does not round-trip byte-for-byte; resubmit with mode=force to save anyway",
				errs.HTTPStatus(http.StatusBadRequest))
		}

		usr := contentUser(r)
		changeset, err := mngr.UpdateEntry(r.PathValue("path"), entry, usr, mgmt.UpdateParams{})
		if err != nil {
			return err
		}

		warnings, err := mngr.SaveChangeset(changeset, usr.Signature, r.URL.Query().Get("msg"), now)
		if err != nil {
			return err
		}
		if len(warnings) > 0 {
			w.Header().Set("Warning", strings.Join(warnings, "; "))
		}
		return JSON(w, changeset)
	}
}

// SearchHandler queries the index. The engine's result is returned as-is, so
// hits keep the `fields` shape the frontend reads.
func SearchHandler(engine *search.Engine) func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		if engine == nil {
			return errs.New("search not implemented", errs.HTTPStatus(http.StatusNotImplemented))
		}

		params, err := handler.ParseWithDefaults(r.URL.Query(), search.Params{PageSize: 20})
		if err != nil {
			return err
		}

		result, err := engine.Search(params)
		if err != nil {
			return errs.Wrap(err, "search failed")
		}
		return JSON(w, result)
	}
}

// JSON writes v as the response body.
func JSON(w http.ResponseWriter, v any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}
