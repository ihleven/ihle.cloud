package usecase

import (
	"bitbucket.org/hotelplan/webcc-content/cms/content"
	"bitbucket.org/hotelplan/webcc-content/cms/mgmt/auth"
	"bitbucket.org/hotelplan/webcc-content/cms/mgmt/search"
	"bitbucket.org/hotelplan/webcc-content/cms/pkg/errors"
)

// * GetEntryByPath(path string, a *permission.User, resolve content.Resolv) (*content.Entry, error)
// * GetEntrySet(dir string) (*content.Entry, error)

//   - ContentEntries(EntryOptions) ([]content.Entry, error) //
//     (gefilterete) Liste von Entries
func (uc *Service) ListEntries() ([]content.Entry, error) {
	uc.RWMutex.RLock()
	defer uc.RWMutex.RUnlock()

	return uc.repo.ListEntries(false) // LoadMulti("")
}

// * Entry(key string, a *permission.User, resolve content.Resolv) (*content.Entry, error)
// GetEntryByPath liefert den entry, der `path` zugeordnet ist.
func (uc *Service) GetEntry(path string, usr content.User, resolve content.Resolv) (*content.Entry, error) {
	uc.repo.RWMutex.RLock()
	defer uc.repo.RWMutex.RUnlock()

	entry, err := uc.repo.GetEntry(path)
	if err != nil {
		return nil, err
	}

	if err := entry.IsReadableOr(usr); err != nil {
		return nil, errors.WrapWithCode(err, 403, "cannot read")
	}

	if resolve&content.RSLV_REFS != 0 {

		for _, ref := range entry.Refs() {
			ref.Entry, err = uc.repo.GetEntry(path)
			if err != nil {
				return nil, err
			}
		}
	}

	if resolve&content.RSLV_LINKS != 0 {
		for _, link := range entry.Links() {
			if link.EntryID != "" {
				if e, err := uc.repo.GetEntry(path); e != nil && err == nil {
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

func (uc *Service) ContentEntries(opts search.Params) ([]content.Entry, error) {
	// if uc.CMS == nil {
	return nil, errors.New("uc.CMS is nil, not implemented")
	// }
	// return uc.CMS.ContentEntries(opts)
}

func (uc *Service) SaveEntries(commit bool, a *auth.Account, msg string, entries ...content.Entry) ([]content.Entry, error) {
	// return uc.CMS.SaveEntries(commit, a, msg, entries...)
	return nil, errors.New("uc.CMS is nil, not implemented")
}

// vor user updateble: Name, Tags, Notes, Slug, FullSlug, Status
// // nicht updateble: Path, Repo, MIME, Type, Locale, Space, Collection  , ID,
// speziell :Version mit SaveAsDraft/Publish
// Owner: nur von Admin
// Group: von Admin oder Owner
// Permission: Admin oder Owner
// Modified: automatisch
// Created: gar nicht (evtl. von Admin)
// Published: durch Publish
func (uc *Service) UpdateEntry(path string, entry *content.Entry, a *content.User) (content.Changeset, error) {

	if path != "" && path != entry.Path {
		// für aktuelle usecases muss url path identisch mit dem entry path sein. Kann sich für copy/move ändern.
		return nil, errors.NewWithCode(400, "validation error: url path != entry.path")
	}

	if err := entry.CheckContentTypeIntegrity(false); err != nil {
		return nil, errors.WrapWithCode(err, 400, "validation error")
	}

	// TODO: Check if entry is allowed for folder: CMS.GetMatchingCollectionForPath / CheckConstraints

	// prüfen ob es sich wirklich um ein Update handelt
	old, err := uc.repo.GetEntry(entry.Path)
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
			publishedversion, _ := uc.repo.GetEntry(entry.ProdPath())
			cs, err := old.Publish(publishedversion, a)
			if err != nil {
				return nil, err
			}
			changeset.Merge(cs)
		case string(content.DRAFT):
			draftversion, _ := uc.repo.GetEntry(entry.DraftPath())
			cs, err := old.SaveAsDraft(draftversion, a)
			if err != nil {
				return nil, err
			}
			changeset.Merge(cs)
		}
	}

	return changeset, nil
}
