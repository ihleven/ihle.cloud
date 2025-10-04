package content

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"bitbucket.org/hotelplan/webcc-content/cms/pkg/errors"
	"github.com/fatih/structs"
)

// Meta beinhaltet die Metadaten für einen Entry
type Meta struct {

	// WO werden die Entries gespeichert: ein Pfad in einem Storagesystem (git/lokal) als MIMEType (json/md) mit einem Encoding/Charset
	Path string `json:"path"`       // e.g. "posts/my-third-post" | full_slug voller Pfad im Space/Repo zur Datei inklusive Verzeichnis- und Dateiname
	Repo string `json:"-" yaml:"-"` // Repo ist der Name der Storage von außen, d.h. der Name innerhlab des CMS um verschiedene Storages zu unterscheiden, d.h. gespeichert sollte der Name nie werden, sondern beim Einlesen in den Cache hinzugefügt werden
	MIME string `json:"mime"`       // MimeType des Inhalt: application/json oder text/markdown oder ...
	Type string `json:"type" yaml:"type"`

	// redaktionelle Metadaten: frei editierbar
	Name     string   `json:"name"` // e.g."My third post"        |  Name im Editor
	Tags     []string `json:"tags"`
	Notes    string   `json:"notes"`
	Slug     string   `json:"slug"`      // e.g. "my-third-post"       | Basename (Pfad ab Folder)
	FullSlug string   `json:"full_slug"` // e.g. "nov2020/my-third-post"       | Dateiname oder Pfad ab Folder?
	Status   status   `json:"status"`    // inactive entries should be ignored on prod: e.g. ACTIVE, INACTIVE

	// ID & virtuelle Daten: Einträge sind in Spaces organisiert bzw. unterhalb einer folder einsortiert. Dort muss die ID eindeutig ein Element bestimmen, das allerdings in mehereren Locales bzw. Versionen vorliegen kann.
	Locale     string `json:"locale"`          // an id can have several localized occurences
	Version    string `json:"version"`         // published, draft:pn13091, scheduled:2024-04-01T15:30:32
	ID         string `json:"id" yaml:"id"`    // logical id of entry -> TODO: rename to VirtID/virtid
	Space      string `json:"space,omitempty"` // namespace of logical id - Must be within a certain storage - Namespace or virtual folder within which the combination of ID and locale must uniquely determine an entry
	Collection string `json:"collection"`      // an entry can belong to a collection of connected entries: mapping to path within the storage of the space

	Access `yaml:",inline"`
}

type status string

const (
	PUBLISHED status = "published"
	DRAFT     status = "draft"
	ACTIVE    status = "active"
	INACTIVE  status = "inactive"
)

func Status(in string) status {
	return status(in)
}

// Access regelt und protokolliert den Zugriff auf den Entry
type Access struct {
	Created    time.Time `json:"created"`  // Zeitpunkt der erstmaligen Erstellung des Entries
	Modified   time.Time `json:"modified"` // Zeitpunkt der letzten Änderung
	Published  time.Time `json:"published"`
	Owner      string    `json:"owner"`
	Groups     string    `json:"group" yaml:"group"` // Groups    []string  `json:"groups,omitempty"`
	Permission Mode      `json:"permissions"`
}

const ReadUsr Mode = 0x100 // 256 ->  -r--------
const ReadGrp Mode = 0x20  // 32  ->  ----r-----
const ReadOth Mode = 0x4   // 4   ->  -------r--
const WriteUsr Mode = 0x80 // 128 ->  --w-------
const WriteGrp Mode = 0x10 // 16  ->  -----w----
const WriteOth Mode = 0x2  // 2   ->  --------w-

// IsReadableOr checks wether user has _either_ user or group or other permissions for read
func (a *Access) IsReadableOr(user User) error {
	if a.Permission == 0 {
		return nil
	}
	if user.ID == a.Owner && a.Permission&ReadUsr > 0 {
		return nil
	}
	for _, group := range user.Groups {
		if group == a.Groups && a.Permission&ReadGrp > 0 {
			return nil
		}
	}
	if a.Permission&ReadOth > 0 {
		return nil
	}
	return errors.New("user %s does not have owner, group and other permissions for read", user.ID)
}

// TODO: deal with list of group memberships
func (a *Access) UsrGrpReadable(usr, grp string) error {
	if a.Permission == 0 {
		return nil
	}
	if usr != "" && usr == a.Owner && a.Permission&ReadUsr == 0 {
		return errors.New("not readable for owner (%s)", a.Owner)
	}
	if grp != "" && grp == a.Groups && a.Permission&ReadGrp == 0 {
		return errors.New("not readable for group (%s)", a.Groups)
	}
	if a.Permission&ReadOth == 0 {
		return errors.New("not readable for non owners or group members")
	}
	return nil
}

func (a *Access) UsrGrpWritable(usr, grp string) error {
	if a.Permission == 0 {
		return nil
	}
	// TODO
	return nil
}

type Mode uint32

func (m Mode) String() string {
	if m == 0 {
		return ""
	}
	var buf [32]byte // Mode is uint32.
	w := 0

	buf[w] = '-'
	w++

	const rwx = "rwxrwxrwx"
	for i, c := range rwx {
		if m&(1<<uint(9-1-i)) != 0 {
			buf[w] = byte(c)
		} else {
			buf[w] = '-'
		}
		w++
	}
	return string(buf[:w])
}
func (m Mode) MarshalText() ([]byte, error) { return []byte(m.String()), nil }

func (m *Mode) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		*m = 0
		return nil
	}
	if len(data) != 10 {
		return errors.New("wrong length")
	}
	var mode Mode
	pattern := []byte("drwxrwxrwx")
	for i, e := range data {
		if e == byte(pattern[i]) {
			mode = mode | 1<<(9-i)
		}
		// fmt.Println(i, e, pattern[i], 10-i, 1<<(9-i))
	}
	*m = mode
	return nil
}

type patchmode int

const (
	DIFF patchmode = iota
	PATCH
	UPDATE
	DEFAULTS
)

// Patch has differing behaviour depending on the given patchmode:
// With patchmode DIFF it calculates the diff between data and m in terms of fields.
// With patchmodes PATCH, UPDATE and DEFAULTS it patches m with data and returns the list of added, changed and deleted fields
// 1) added fields: fields with zero value in m and value in data
// 2) changed fields: fields with differung non zero values in m and data
// 3) deleted fields: fields with non zero value in m and zero value in data
func (m *Meta) Patch(data Meta, mode patchmode) ([]string, []string, []string, error) {

	var added, updated, deleted []string

	src := structs.New(data)
	dst := structs.New(m)
	fields := append(src.Fields(), src.Field("Access").Fields()...)
	for _, field := range fields {

		if field.IsEmbedded() {
			continue
		}
		if !field.IsExported() {
			continue // skip unexported fields
		}
		name := field.Name()
		dstField, ok := dst.FieldOk(name)
		if !ok {
			continue // skip non-existing fields
		}
		// fmt.Println("EMBEDDED:", field.Name(), dstField.Name(), equals(field, dstField))
		switch mode {
		case DIFF:
			if field.IsZero() && !dstField.IsZero() {
				added = append(added, name)
			} else if !field.IsZero() && dstField.IsZero() {
				deleted = append(deleted, name)
			} else if !field.IsZero() && !dstField.IsZero() && !equals(field, dstField) {
				updated = append(updated, name)
			}
			continue

		case UPDATE:
			if field.IsZero() && !dstField.IsZero() {
				deleted = append(deleted, name)
				dstField.Set(field.Value())
			}
			fallthrough
		case PATCH:
			if !field.IsZero() && dstField.IsZero() {
				added = append(added, name)
				dstField.Set(field.Value())
			} else if !field.IsZero() && !dstField.IsZero() && !equals(field, dstField) {
				updated = append(updated, name)
				dstField.Set(field.Value())
			}
			continue

		case DEFAULTS:
			if !field.IsZero() && dstField.IsZero() {
				added = append(added, name)
				dstField.Set(field.Value())
			}
			continue
		}
	}

	return added, updated, deleted, nil
}

func equals(srcField, dstField *structs.Field) bool {
	srcValue, dstValue := srcField.Value(), dstField.Value()
	if skind, dkind := srcField.Kind(), dstField.Kind(); skind != dkind {
		// err = fmt.Errorf("field `%v` types mismatch while patching: %v vs %v", name, dkind, skind)
		return false
	}

	if reflect.DeepEqual(srcValue, dstValue) {
		return true
	}

	return false
}

// PatchZeroFieldsDep overwrite only fields in m with zero value where new has non zero values in the corresponding field
func (m *Meta) PatchZeroFieldsDep(new Meta) (changed bool, err error) {
	var dst = structs.New(m)
	var fields = structs.New(new).Fields() // work stack

	for N := len(fields); N > 0; N = len(fields) {
		var srcField = fields[N-1] // pop the top
		fields = fields[:N-1]

		if !srcField.IsExported() {
			continue // skip unexported fields
		}
		if srcField.IsEmbedded() {
			// add the embedded fields into the work stack
			fields = append(fields, srcField.Fields()...)
			continue
		}
		if srcField.IsZero() {
			continue // skip zero-value fields
		}

		var name = srcField.Name()

		var dstField, ok = dst.FieldOk(name)
		if !ok {
			continue // skip non-existing fields
		}
		fmt.Println("field", srcField.Name(), name)
		if !dstField.IsZero() {
			continue // skip zero-value fields
		}
		var srcValue = reflect.ValueOf(srcField.Value())
		srcValue = reflect.Indirect(srcValue)
		if skind, dkind := srcValue.Kind(), dstField.Kind(); skind != dkind {
			err = fmt.Errorf("field `%v` types mismatch while patching: %v vs %v", name, dkind, skind)
			return
		}

		if !reflect.DeepEqual(srcValue.Interface(), dstField.Value()) {
			changed = true
		}

		err = dstField.Set(srcValue.Interface())
		if err != nil {
			return
		}
	}
	fmt.Println("meta:", m)
	return
}

// Check (or sets in case of fix==true)
// Replacement for first half of CMS.CheckConstraints(*content.Entry) error
func (e *Entry) CheckContentTypeIntegrity(fix bool) error {

	// content type integrity check
	ct := e.ContentType()
	if fix {
		e.Meta.Type = ct.Name
		e.Meta.MIME = ct.MIME
		if ct.Folder != "" {
			e.Meta.Collection = ct.Folder
		}
		return nil
	}
	if ct.MIME != "" && ct.MIME != e.MIME {
		return errors.NewWithCode(400, "validation error: %s != %s", e.MIME, ct.MIME)
	}
	if ct.Folder != "" && !strings.HasPrefix(e.Path, ct.Folder) {
		return errors.NewWithCode(400, "validation error: %s must be stored in folder %s", ct.Name, ct.Folder)
	}
	return nil

	// TODO: second half of CMS.CheckConstraints(*content.Entry) error
	// // folder integrity check  TODO con CheckConstraints lösen und Defaultwerte je nach Folder setzen wie z.B. Collection
	// if folder := cms.GetMatchingCollectionForPath(entry.Path); folder != nil {
	// 	fmt.Println("FOLDER = ", folder, entry.Path)
	// 	if err := folder.AllowsEntry(entry); err != nil {
	// 		return err
	// 	}
	// }
}

type Defaulter interface{ MetaDefaults(*Meta) error }

func (e *Entry) PrepareMetaForCreate(account *User) error {

	i, err := Instantiate(e.Meta.Type)
	if err != nil {
		return errors.New("no valid content type: %s", e.Meta.Type)
	}
	if reflect.TypeOf(i).Kind() != reflect.TypeOf(e.Content).Kind() {
		return errors.New("inconsistent content type %s - %T", e.Meta.Type, e.Content)
	}

	if defaulter, ok := e.Content.(Defaulter); ok {
		err = defaulter.MetaDefaults(&e.Meta)
	}

	ct := e.ContentType()
	e.Meta.Space = "website"
	// Version:    string(PUBLISHED)
	if e.Meta.Status != "" {
		e.Meta.Status = "active"
	}
	e.Meta.MIME = ct.MIME
	e.Meta.Collection = ct.Folder
	if e.Meta.Tags == nil {
		e.Meta.Tags = []string{}
	}

	now := time.Now()
	e.Meta.Access = Access{
		Owner:      account.ID,
		Groups:     "",
		Created:    now,
		Modified:   now,
		Published:  time.Time{},
		Permission: 0,
	}
	err = (&e.Meta.Access.Permission).UnmarshalText([]byte("drwxrwxrwx"))
	if err != nil {
		return errors.New("error creating permissions in PrepareMetaForCreate")
	}
	return nil
}

// func (m *Meta) Key(logical bool) string {

// 	scheme := "content"
// 	if m.Repo != "" {
// 		scheme = m.Repo
// 	}
// 	return fmt.Sprintf("%s:%s", scheme, m.Path)
// }
