package cmsuc

import (
	"fmt"
	"log"
	gopath "path"
	"strings"

	"bitbucket.org/hotelplan/webcc-content/cms/content"
	"bitbucket.org/hotelplan/webcc-content/cms/mgmt/gitrepo"
	"bitbucket.org/hotelplan/webcc-content/cms/mgmt/search"
	"bitbucket.org/hotelplan/webcc-content/cms/pkg/errors"
)

type Config struct {
	RepoContent string
	DataDir     string
	Search      search.Config
}

func NewCMS(conf Config) (*CMS, error) {

	repo, err := gitrepo.New(conf.RepoContent, conf.DataDir, false, gitrepo.WithCommitter("content api", "cms@interhome.group"))
	if err != nil {
		return nil, err
	}

	// mappings := map[string]*mapping.DocumentMapping{
	// 	"Work":       art.Work{}.BleveMapping(),
	// 	"exhibition": art.Ausstellung{}.BleveMapping(),
	// }

	engine, err := search.NewEngine(conf.Search, nil) // mappings)
	if err != nil {
		log.Fatalf("engine %v %s", conf, err)
	}

	entries, err := repo.ListEntries(true)
	if err != nil {
		log.Fatal(err)
	}

	err = engine.Index(entries...)
	if err != nil {
		log.Fatal(err)
	}
	cms := CMS{Repo: repo, Engine: engine}

	return &cms, nil
}

type CMS struct {
	// sync.RWMutex
	Repo   *gitrepo.Sitory
	Engine *search.Engine
}

func (cms *CMS) GetEntry(path string, usr content.User, resolve content.Resolv) (*content.Entry, error) {
	cms.Repo.RWMutex.RLock()
	defer cms.Repo.RWMutex.RUnlock()

	entry, err := cms.Repo.GetEntry(path)
	if err != nil {
		fmt.Println(err)
		if errors.Code(err) == 404 {
			return cms.GetDirEntrySearch(path)
			// return cms.GetDirEntries(path), nil
		}
		return nil, err
	}
	// fmt.Println("GetEntry", entry)

	if err := entry.IsReadableOr(usr); err != nil {
		return nil, errors.WrapWithCode(err, 403, "cannot read")
	}

	if resolve&content.RSLV_REFS != 0 {

		for _, ref := range entry.Refs() {
			ref.Entry, err = cms.Repo.GetEntry(path)
			if err != nil {
				return nil, err
			}
		}
	}

	if resolve&content.RSLV_LINKS != 0 {
		for _, link := range entry.Links() {
			if link.EntryID != "" {
				if e, err := cms.Repo.GetEntry(path); e != nil && err == nil {
					link.Slug = e.FullSlug
				}
			}
		}
	}

	// TODO: resolving links => check for and use search

	linkResolver := func(input string, locale string) string {
		// if m, ok := mgmt.cache.byID[input]; ok {
		// 	return m[locale].FullSlug
		// }
		return ""
	}
	if resolve&content.RSLV_CB != 0 {
		if resolver, ok := entry.Content.(interface {
			ResolveLinks(string, func(string, string) string)
		}); ok {
			resolver.ResolveLinks(entry.Locale, linkResolver)
		}
	}

	return entry, err

}

type DirEntry struct {
	content.Meta
	Dir string `json:"dir"`
	Img string `json:"img"`
}

// GetDirEntry listet die Inhalte eines Verzeichnisses auf, indem es die Suche verwendet.
func (cms *CMS) GetDirEntrySearch(path string) (*content.Entry, error) {

	entry := content.Entry{
		Meta: content.Meta{
			Path: path,
			Name: gopath.Base(path),
			Type: "Dir",
		},
		Content: []DirEntry{},
	}

	if path == "" {
		path = "."
	}
	result, err := cms.Engine.Search(search.Params{Dir: path, Fields: "*", PageSize: 100000})
	if err != nil {
		return nil, err
	}

	if result.Total > 0 {
		var entries []DirEntry
		for _, hit := range result.Hits {

			var direntry DirEntry

			if path, ok := hit.Fields["path"].(string); ok {
				direntry.Path = path
			}
			if name, ok := hit.Fields["name"].(string); ok {
				direntry.Name = name
			}
			if typ, ok := hit.Fields["type"].(string); ok {
				direntry.Type = typ
			}
			if locale, ok := hit.Fields["locale"].(string); ok {
				direntry.Locale = locale
			}
			if version, ok := hit.Fields["version"].(string); ok {
				direntry.Version = version
			}
			if slug, ok := hit.Fields["slug"].(string); ok {
				direntry.Slug = slug
			}
			if fullslug, ok := hit.Fields["full_slug"].(string); ok {
				direntry.FullSlug = fullslug
			}
			if dir, ok := hit.Fields["dir"].(string); ok {
				direntry.Dir = dir
			}

			entries = append(entries, direntry)
		}
		entry.Content = entries
	}

	return &entry, nil

}

// GetDirEntries listet die Inhalte eines Verzeichnisses auf, ohne die Einträge zu laden.
// Und ohne auf die Suche zuzugreifen.
func (cms *CMS) GetDirEntries(path string) *content.Entry {

	entries := []content.Meta{}
	for _, tree := range cms.Repo.Directories[path].Dirs {

		entries = append(entries, content.Meta{
			Name: tree.Name, Path: tree.Path, Type: "Dir",
		})

	}
	for _, name := range cms.Repo.Directories[path].Entries {

		entries = append(entries, content.Meta{
			Name: name,
			Path: path + "/" + name,
			ID:   strings.TrimSuffix(name, gopath.Ext(name)),
		})

	}
	e := content.Entry{
		Meta: content.Meta{
			Path: path,
			Name: gopath.Base(path),
			Type: "Dir",
		},
		Content: entries,
	}
	return &e
}

func (cms *CMS) UpdateEntry(path string, entry *content.Entry, a *content.User) (content.Changeset, error) {

	if path != "" && path != entry.Path {
		// für aktuelle usecases muss url path identisch mit dem entry path sein. Kann sich für copy/move ändern.
		return nil, errors.NewWithCode(400, "validation error: url path != entry.path")
	}

	if err := entry.CheckContentTypeIntegrity(false); err != nil {
		return nil, errors.WrapWithCode(err, 400, "validation error")
	}

	// TODO: Check if entry is allowed for folder: CMS.GetMatchingCollectionForPath / CheckConstraints

	// prüfen ob es sich wirklich um ein Update handelt
	old, err := cms.Repo.GetEntry(entry.Path)
	if err != nil {
		// TODO: check for 404
		return nil, errors.NewWithCode(400, "validation error: entry %s already exists", entry.Path)
	}

	// update von content und den user editable meta daten
	changeset, err := old.Update(*entry, *a)
	if err != nil {
		return nil, err
	}

	if old.Version != entry.Version {
		switch entry.Version {
		case string(content.PUBLISHED):
			publishedversion, _ := cms.Repo.GetEntry(entry.ProdPath())
			cs, err := old.Publish(publishedversion, a)
			if err != nil {
				return nil, err
			}
			changeset.Merge(cs)
		case string(content.DRAFT):
			draftversion, _ := cms.Repo.GetEntry(entry.DraftPath())
			cs, err := old.SaveAsDraft(draftversion, a)
			if err != nil {
				return nil, err
			}
			changeset.Merge(cs)
		}
	}

	return changeset, nil
}
