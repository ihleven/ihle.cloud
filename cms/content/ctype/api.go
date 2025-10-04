package ctype

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"bitbucket.org/hotelplan/webcc-content/cms/api"
	"bitbucket.org/hotelplan/webcc-content/cms/content"
	"bitbucket.org/hotelplan/webcc-content/cms/mgmt/search"
	"bitbucket.org/hotelplan/webcc-content/cms/pkg/errors"
)

func Routes(a *api.Webapi, route func(pattern string, h interface{}, adapters ...api.HAdapter)) {

	grp := routr{a}

	route("  GET  ~/meta       ", grp.PagesMeta, api.Authenticated)

	// erzeugen einer neuen Page auf Basis einer bestehenden mit Anpassung der Inhalte z.b. Breadcrumb, canonical
	route("  POST  ~/createadaptpage      ", grp.PagesMeta)

	// Infos über die zugehörigen Pages in anderen Locales
	// welche IDs identisch/abweichend sind
	// welche Locales kann man spezifisch anlegn
	// welche Locales kann man generisch anlegen
	// infos über die ancestor pages im aktuellen Locale: slugs, locales
	route("  GET  ~/pagectx/{path...}      ", grp.PagesContext, api.Authenticated)
}

type routr struct{ *api.Webapi }

func (a *routr) PagesMeta(w *api.ResponseWriter, r *http.Request) error {

	// if a.CMS == nil {
	return errors.NewWithCode(http.StatusNotImplemented, "CMS not available, probably in NEXT mode")
	// }

	// if w.Authuser == "" {
	// 	return errors.NewWithCode(401, "not authenticated: %s", w.Authuser)
	// }

	// opts := mgmt.EntryOptions{EntryFilter: mgmt.EntryFilter{Type: "Page"}}

	// err := schema.NewDecoder().Decode(&opts, r.URL.Query())
	// if err != nil {
	// 	return errors.WrapWithCode(err, 400, "Couldn't decode search params %v", r.URL.Query())
	// }

	// meta, err := a.CMS.PageMeta(opts)
	// if err != nil {
	// 	return err
	// }

	// return w.RespondJSON(meta, map[api.RWOption]bool{api.PRETTY: r.URL.Query().Has("pretty")})

	// return w.RespondJSON(meta)
}

func (a *routr) PagesContext(w *api.ResponseWriter, r *http.Request) error {
	if w.Authuser == "" {
		return errors.NewWithCode(401, "not authenticated: %s", w.Authuser)
	}

	_, account, err := api.Reqctx[struct{}](a.Webapi, w, r)
	if err != nil {
		return err
	}

	entry, err := a.UC.GetEntry(r.PathValue("path"), account.User, content.RSLV_NONE)
	if err != nil {
		return errors.Wrap(err, "Entry not found")
	}

	// dir := strings.TrimSuffix(path.Dir(entry.Path), "/")
	dir, base := path.Split(entry.Path)
	dir = strings.TrimSuffix(dir, "/")

	// set, err := a.CMS.GetEntrySet(dir)
	// fmt.Println("set.Content", set, err)
	// if err != nil {
	// 	return err
	// }
	set := &content.Entry{}

	metas := []content.Meta{}
	for _, v := range set.Content.(map[string]content.Meta) {
		metas = append(metas, v)
	}

	ancestors := []content.Meta{}
	for ; dir != "."; dir = path.Dir(dir) {
		fmt.Println("dir", dir+"/"+base)
		// entry, err := a.CMS.GetEntryByPath(dir+"/"+base, &account.User, content.RSLV_NONE)
		// if err != nil {
		// 	fmt.Println(err)
		// 	continue
		// }
		// ancestors = append(ancestors, entry.Meta)
	}

	// set.Content.()
	response := struct {
		Anestors []content.Meta `json:"ancestors"`
		Siblings []content.Meta `json:"siblings"`
	}{
		Anestors: ancestors,
		Siblings: metas,
	}
	return w.RespondJSON(response)
}
func (a *routr) PageSlug(w *api.ResponseWriter, r *http.Request) error {

	if w.Authuser == "" {
		return errors.NewWithCode(401, "not authenticated: %s", w.Authuser)
	}

	// _, account, err := api.Reqctx[struct{}](a.Webapi, w, r)
	// if err != nil {
	// 	return err
	// }

	// entry, err := a.CMS.GetEntryByPath(r.PathValue("path"), &account.User, content.RSLV_NONE)
	// if err != nil {
	// 	return errors.Wrap(err, "Entry not found")
	// }

	// newslug := r.PostFormValue("slug")

	// if newslug == "" || newslug == entry.Slug {
	// 	return nil
	// }

	// dir, base := path.Split(entry.Path)
	// dir = strings.TrimSuffix(dir, "/")

	// ancestors := []content.Meta{}
	// for ; dir != "."; dir = path.Dir(dir) {
	// 	fmt.Println("dir", dir+"/"+base)
	// 	entry, err := a.CMS.GetEntryByPath(dir+"/"+base, &account.User, content.RSLV_NONE)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		continue
	// 	}
	// 	ancestors = append(ancestors, entry.Meta)
	// }

	// entry.FullSlug = newslug

	return nil
}

func GetAncestor( //cms *mgmt.CMS,
	entry *content.Entry, engine *search.Engine) (*content.Entry, error) {
	dirWithoutTrailingSlash := path.Dir(entry.Path)

	// filename := fmt.Sprintf("%s%s.%s", entry.Locale, map[string]string{"published": "", "draft": ".draft"}[entry.Version], entry.MIME)

	result, err := engine.Search(search.Params{Dir: dirWithoutTrailingSlash, Type: []string{"Page"}, Field: []string{"locale"}})
	if err != nil {
		return nil, err
	}

	for _, hit := range result.Hits {
		fmt.Println("hit", hit)
		locale := hit.Fields["locale"]
		switch l := locale.(type) {
		case []string:
			for _, loc := range l {
				if loc == entry.Locale {
					// return cms.GetEntryByPath(hit.ID, nil, content.RSLV_NONE)
				}
			}
		case string:
			if l == entry.Locale {
				// return cms.GetEntryByPath(hit.ID, nil, content.RSLV_NONE)
			}
		}

	}

	return nil, errors.New("Ancestor not found")
}
