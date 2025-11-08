package cmsuc

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"bitbucket.org/hotelplan/webcc-content/cms/content"
	"bitbucket.org/hotelplan/webcc-content/cms/mgmt/gitrepo"
	"bitbucket.org/hotelplan/webcc-content/cms/mgmt/search"
	"bitbucket.org/hotelplan/webcc-content/cms/permission"
	"bitbucket.org/hotelplan/webcc-content/cms/pkg/errors"
	"github.com/gorilla/schema"
	"github.com/ihleven/ihlvn/pkg/auth"
)

func NewCMSApi(conf Config) (*Api, error) {

	cms, err := NewCMS(conf)
	if err != nil {
		return nil, err
	}
	return &Api{CMS: cms}, nil
}

type Api struct {
	*CMS
}

func (a *Api) ReqAuth(r *http.Request) (*content.User, error) {
	account, _, err := auth.GetAccountUnused(r)
	if err != nil {
		return nil, err
	}

	usr := content.User{
		ID:     account.ID,
		Groups: account.CMS.Groups,
		Scope:  map[permission.Type]permission.Permission{},
	}

	usr.Scope.Add(account.CMS.Permissions, []string{}, []string{})

	return &usr, nil
}

func JSON(w http.ResponseWriter, value interface{}) error {
	if value == nil {
		return nil
	}

	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	if err := enc.Encode(value); err != nil {
		return err
	}

	return nil
}

// Returns Entry for given path
// If path is not found as entry, but as dir, a dir entry is returned.
func (a *Api) EntryDetails(w http.ResponseWriter, r *http.Request) error {

	path := strings.TrimSuffix(r.PathValue("path"), "/")

	if slugs := r.URL.Query().Get("slugs"); slugs != "" {

		searchresult, err := a.Engine.Search(search.Params{FullSlug: slugs, PageSize: 1})
		if err == nil {
			switch searchresult.Total {
			case 0:
				return errors.NewWithCode(404, "full_slug not found: %s", slugs)
			case 1:
				path = searchresult.Hits[0].ID
			default:
				return errors.NewWithCode(500, "full_slug not unique: %s", slugs)
			}
		}
	}

	/////////////
	jwt := ""
	if header := r.Header.Get("Authorization"); header != "" {
		jwt = strings.TrimSpace(strings.TrimPrefix(header, "Bearer"))
	}
	if jwt == "" {
		if cookie, err := r.Cookie("jwt"); err == nil {
			jwt = cookie.Value
		}
	}

	claims, err := auth.AuthenticatorPKG.ParseClaims(jwt)
	if err != nil {
		return errors.NewWithCode(401, "invalid token: %v", err)
	}

	usr := content.User{ID: claims.Subject, Groups: claims.Audience}

	// fmt.Println("entryDetails", path, usr)
	entry, err := a.GetEntry(path, usr, content.RSLV_NONE)
	if err != nil {
		fmt.Println("entry not found", errors.Code(err), errors.Cause(err), err)
		return err
	}

	// fmt.Println("entry not found", err)

	// entries := []content.Meta{}
	// for _, tree := range a.Repo.Directories[path].Dirs {

	// 	entries = append(entries, content.Meta{
	// 		Name: tree.Name, Path: tree.Path, Type: "Dir",
	// 	})

	// }
	// for _, name := range a.Repo.Directories[path].Entries {

	// 	entries = append(entries, content.Meta{
	// 		Name: name, Path: path + "/" + name, ID: strings.TrimSuffix(name, gopath.Ext(name)),
	// 	})

	// }
	// e := content.Entry{
	// 	Meta: content.Meta{
	// 		Path: path,
	// 		Name: gopath.Base(path),
	// 		Type: "Dir",
	// 	},
	// 	Content: entries,
	// }
	return JSON(w, entry)

}

// }

// func (a *api) EntryUpdate(rw *web.ResponseWriter, r *http.Request) error {
func (a *Api) EntryUpdate(w http.ResponseWriter, r *http.Request) error {

	usr, err := a.ReqAuth(r)
	if err != nil {
		return errors.WrapWithCode(err, 401, "Could not find account")
	}

	now := time.Now()
	mode := r.URL.Query().Get("mode")

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		return errors.NewWithCode(400, "Could not read request body")
	}

	contentType, _, _ := mime.ParseMediaType(r.Header.Get("Content-type"))
	if contentType == "" {
		// contentType = "application/json"
		return errors.NewWithCode(400, "Could not find ContentType")
	}

	entry, err := content.ParseEntry(nil, bytes, contentType)
	if err != nil {
		return errors.Wrap(err, "Parse error")
	}

	// UPDATE
	valid, err := entry.ValidateSyntaxIntegrity(bytes)
	if err != nil {
		return errors.Wrap(err, "ValidateBytes in entryWriteHandler")
	}

	if mode == "validate" {
		return JSON(w, entry)
	}

	if !valid && mode != "force" {
		return errors.NewWithCode(400, "!valid && !force")
	}

	changeset, err := a.CMS.UpdateEntry(r.PathValue("path"), entry, usr)
	if err != nil {
		return err
	}

	signature := gitrepo.Signature{Name: "ihleven", Email: "github@ihleven.de"}
	err = a.Repo.Commit(gitrepo.Entrymap(changeset), signature, r.URL.Query().Get("msg"), now)
	if err != nil {
		return err
	}

	for _, entry := range changeset {
		if entry == nil {
			// TODO: Delete from index
			continue
		}
		err = a.Engine.Index(*entry)
		if err != nil {
			return err
		}
	}
	// if len(emap) == 1 {
	// 	for _, e := range emap {
	// 		return w.RespondJSON(e)
	// 	}
	// } else {
	return JSON(w, changeset)
	// }

}

func ParseWithDefaults[V interface{}](query url.Values, typeWithDefaults V) (V, error) {

	d := schema.NewDecoder()
	d.IgnoreUnknownKeys(true)
	err := d.Decode(&typeWithDefaults, query)
	if err != nil {
		return typeWithDefaults, errors.Wrap(err, "failed to decode search params")
	}
	return typeWithDefaults, nil
}

func SearchEntries(cms *CMS) func(http.ResponseWriter, *http.Request) error {

	return func(w http.ResponseWriter, r *http.Request) error {

		if cms.Engine == nil {
			return errors.NewWithCode(http.StatusNotImplemented, "search not implemented")
		}

		params, err := ParseWithDefaults(r.URL.Query(), search.Params{Sorting: []string{"-_score", "_id"}, PageSize: 20})
		if err != nil {
			return err
		}

		result, err := cms.Engine.Search(params)
		if err != nil {
			return err
		}

		return JSON(w, search.NewResult(result))
	}
}

func EntryLookup(cms *CMS) func(http.ResponseWriter, *http.Request) error {

	return func(w http.ResponseWriter, r *http.Request) error {

		if cms.Engine == nil {
			return errors.NewWithCode(http.StatusNotImplemented, "search not implemented")
		}

		params, err := ParseWithDefaults(r.URL.Query(), search.Params{})
		if err != nil {
			return err
		}

		params.PageSize = 1

		result, err := cms.Engine.Search(params)
		if err != nil {
			return err
		}

		if result.Total != 1 {
			return errors.NewWithCode(404, "not found or not unique")
		}

		// entry, err := cms.Repo.GetBytes(result.Hits[0].ID, false)
		// if err != nil {
		// 	return err
		// }

		// w.Header().Set("ContentType", "application/json")
		// n, err := w.Write(entry)
		// fmt.Print("%d bytes written", n)
		// return err

		entry, err := cms.Repo.GetEntry(result.Hits[0].ID)
		if err != nil {
			return err
		}
		return JSON(w, entry)
	}
}
