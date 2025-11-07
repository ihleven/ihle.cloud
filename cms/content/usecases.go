package content

import (
	"bitbucket.org/hotelplan/webcc-content/cms/permission"
	"bitbucket.org/hotelplan/webcc-content/cms/pkg/errors"
)

type User struct {
	ID     string           `json:"id"`
	Groups []string         `json:"groups"`
	Scope  permission.Scope `json:"-"`
}

func (u *User) Clone() *User {
	clone := User{ID: u.ID, Groups: append(u.Groups[:0:0], u.Groups...), Scope: make(permission.Scope)}
	for k, v := range u.Scope {
		clone.Scope[k] = permission.Permission{
			ID:        v.ID,
			Namespace: v.Namespace,
			Prefixes:  append(v.Prefixes[:0:0], v.Prefixes...),
			Locales:   append(v.Locales[:0:0], v.Locales...),
		}
	}
	return &clone
}

type Changeset map[string]*Entry

func (cs Changeset) Merge(new Changeset) Changeset {
	if cs == nil {
		cs = Changeset{}
	}
	for path, entry := range new {
		cs[path] = entry
	}
	return cs
}

func (u *User) Create(data Entry) (Changeset, error) { // , parents *Folder) (Changeset, error) {
	if u.Scope.Denies(ENTRY_CREATE) {
		return nil, errors.New("missing permission: %s (%s)", ENTRY_CREATE, u.ID)
	}

	err := data.PrepareMetaForCreate(u)
	if err != nil {
		return nil, err
	}

	// Eingabe validieren: war bisher in cms.CheckConstraints(data)
	//  -> folder integrity check  TODO con CheckConstraints lösen
	//     und Defaultwerte je nach Folder setzen wie z.B. Collection
	err = data.CheckContentTypeIntegrity(false)
	if err != nil {
		return nil, err
	}
	// // todo: das bezieht sich auf storage, sollte nicht in content geprueft werden
	// if err := parents.AllowsEntry(&data); err != nil {
	// 	return nil, err
	// }

	cs := map[string]*Entry{data.Path: &data}
	return cs, nil
}

func (e *Entry) Update(data Entry, user User) (Changeset, error) {

	// jetzt der inhaltscheck ob es sich um ein valides Update handelt.
	// hier werden Permissions geprüft
	// ...
	err := data.IsValidUpdateTo(e, &user)
	if err != nil {
		return nil, errors.WrapWithCode(err, 400, "not a valid update")
	}
	e.Content = data.Content
	e.Meta.Name = data.Meta.Name
	e.Meta.Notes = data.Meta.Notes
	e.Meta.Slug = data.Meta.Slug
	e.Meta.FullSlug = data.Meta.FullSlug
	e.Meta.Status = data.Meta.Status
	e.Meta.Tags = make([]string, len(data.Meta.Tags))
	copy(e.Meta.Tags, data.Meta.Tags)
	e.Meta.Name = data.Meta.Name
	e.Meta.Name = data.Meta.Name
	// ... weitere Metadata

	return Changeset{e.Path: e}, nil
}

func (e *Entry) Publish(pubentry *Entry, user *User) (Changeset, error) {
	return nil, nil
}

func (e *Entry) SaveAsDraft(draftentry *Entry, user *User) (Changeset, error) {
	return nil, nil
}

func (u *User) Delete(entry Entry) (Changeset, error) {
	return nil, nil
}

func (u *User) Publish(data, draft, published Entry) (Changeset, error) {
	return nil, nil
}

func (u *User) Read(entry Entry) error {
	return nil
}
