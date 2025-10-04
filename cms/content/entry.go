package content

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"reflect"
	"strings"

	"bitbucket.org/hotelplan/webcc-content/cms/permission"
	"bitbucket.org/hotelplan/webcc-content/cms/pkg/errors"
)

// Entry ist der Wrapper für alle Inhalte.
// Der eigentliche Inhalt ist im interface{}-Typ content als Pointer enthalten.
// Meta.Type muss die Stringrepräsentation des Typs von Content enthalten.
type Entry struct {
	Meta
	Content interface{} `json:"content"`
}

func (e *Entry) Refs() []*EntryRef {
	if r, ok := e.Content.(interface{ Refs() []*EntryRef }); ok {
		return r.Refs()
	}
	return nil
}

func (e *Entry) Links() []*EntryLink {
	if r, ok := e.Content.(interface{ Links() []*EntryLink }); ok {
		return r.Links()
	}
	return nil
}

type cloner interface {
	Clone() interface{}
}

func (e *Entry) Clone() (Entry, error) {
	if cloner, ok := e.Content.(cloner); ok {
		return Entry{
			Meta:    e.Meta,
			Content: cloner.Clone(),
		}, nil
	}

	return Entry{}, errors.New("no cloner")
}

func (e *Entry) MarshalJSON() (text []byte, err error) {
	aux := struct {
		Meta
		Content interface{} `json:"content"`
	}{Meta: e.Meta, Content: e.Content}
	return json.Marshal(aux)
}

func (e *Entry) UnmarshalJSON(data []byte) error {

	if e == nil {
		return nil
	}

	var aux struct {
		Meta
		Content json.RawMessage `json:"content,omitempty"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	e.Meta = aux.Meta

	if t, err := Instantiate(aux.Meta.Type); err == nil {
		// fmt.Printf(" +++++++++++++++++++ %T +++++++++++++++\n", t)
		if aux.Content != nil {
			if err := json.Unmarshal(aux.Content, t); err != nil {
				return err
			}
		}
		e.Content = t
	} else {
		return errors.Wrap(err, "Type not registered: %s", aux.Meta.Type)
	}

	return nil
}

func (e *Entry) Equals(other *Entry) bool {
	if e == nil || other == nil {
		return false
	}

	return reflect.DeepEqual(e, other)
}

// ContentType returns content type info for the embedded ContentType of the entry content
// by reading the struct tags.
// TODO: in addition to return a content type fill the embedded ContentType?
func (e *Entry) ContentType() ContentType {

	var contenttype ContentType

	t := reflect.TypeOf(e.Content)
	if t == nil {
		return contenttype
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	contenttype.Name = t.Name()

	if t.Kind() == reflect.Struct {
		tag := t.Field(0).Tag
		contenttype.Folder = tag.Get("folder")
		contenttype.MIME = tag.Get("mimetype")
		if contenttype.MIME == "" {
			contenttype.MIME = "application/json"
		}
	}

	return contenttype
}

// die Konfiguration für den ContentType wie z.B.
// Name, Ablageorte, ...
type ContentType struct {
	Name   string `json:"name"`
	Folder string `json:"folder"` // Verzeichnis, in dem die Entries des ContentTypes liegen
	MIME   string
}

// type Locale string

// Check ob keine Daten 'verloren' gehen.
// ValidateSyntaxIntegrity compares (deep equal) the parsed bytestream with the serialized and parsed entry.
func (e *Entry) ValidateSyntaxIntegrity(bytestream []byte) (bool, error) {

	// 1. parse bytestream into generic interface
	var dataFromBytestream interface{}
	err := json.Unmarshal(bytestream, &dataFromBytestream)
	if err != nil {
		return false, err
	}

	// 2. serialize our entry ...
	bytestreamFromParsedEntry, _ := json.Marshal(e)
	// ... and parse it back into a generic interface
	var dataFromEntry interface{}
	err = json.Unmarshal(bytestreamFromParsedEntry, &dataFromEntry)
	if err != nil {
		return false, err
	}

	// finally, compare both generic interfaces
	isEqual := reflect.DeepEqual(dataFromBytestream, dataFromEntry)
	if !isEqual {
		fmt.Println(" +++++ DEEP EQUAL differs +++++")
		fmt.Printf("BYTES: %v\n", dataFromBytestream)
		fmt.Printf("ENTRY: %v\n", dataFromEntry)
	}

	return isEqual, nil
}

// TODO: scheduled und Draftversionen mit : unterstützen
func (e *Entry) ProdPath() string {
	// e.g. de-DE.draft.json
	ext := path.Ext(e.Path)
	withoutext := strings.TrimSuffix(e.Path, ext)
	withoutdraft := strings.TrimSuffix(withoutext, "."+string(DRAFT))

	return withoutdraft + ext
}

func (e *Entry) DraftPath() string {

	// de-DE.draft:pn13091:2024-04-01T15:30:04.json
	ext := path.Ext(e.Path)
	withoutext := strings.TrimSuffix(e.Path, ext)
	withoutdraft := strings.TrimSuffix(withoutext, "."+string(DRAFT)) // falls e.Path .draft enthält

	return withoutdraft + "." + string(DRAFT) + ext
}

var metaFieldPermissionMap = map[string]permission.Type{
	"ID":          ENTRY_META_SET_ID,
	"Name":        ENTRY_META_SET_NAME,
	"Notes":       ENTRY_META_SET_NOTES,
	"Path":        ENTRY_META_SET_PATH,
	"Slug":        ENTRY_META_SET_SLUG,
	"Locale":      ENTRY_META_SET_LOCALE,
	"Tags":        ENTRY_META_SET_TAGS,
	"Status":      ENTRY_META_SET_STATUS,
	"Version":     ENTRY_META_SET_VERSION,
	"Owner":       ENTRY_META_SET_OWNER,
	"Group":       ENTRY_META_SET_GROUP,
	"Permissions": ENTRY_META_SET_PERMIS,
}

type checker interface {
	CheckPermissionForWrite(permission.Scope, *Entry) error
}

// IsValidUpdateTo checks which entry meta data and content fields differ in contrast to the old entry and
// checks wether these update are permitted with respect to field level and owner/group level permissions.
// If entry has a non zero modified timestamp, it is compared to the modified timestamp of the old entry.
// If old entry has a newer timestamp it is assumed that a concurrent update is invalidating the checked update.
func (e *Entry) IsValidUpdateTo(old *Entry, account *User) error {

	if !e.Modified.IsZero() && e.Modified.Before(old.Modified) {
		return errors.NewWithCode(http.StatusConflict, "Outdated non-zero timestamp %s ", e.Modified)
	}

	added, changed, deleted, err := e.Meta.Patch(old.Meta, DIFF)
	if err != nil {
		return err
	}

	// check permission
	for _, fieldname := range append(append(added, changed...), deleted...) {
		perm, ok := metaFieldPermissionMap[fieldname]
		if ok && !account.Scope.HasOnePermission(perm, ENTRY_META_SET_ALL) {
			return errors.New("setting field %s not allowed", fieldname)
		}
	}

	contentChanged := !reflect.DeepEqual(e.Content, old.Content)

	if contentChanged {
		// check content permissions
		if checker, ok := e.Content.(checker); ok {
			checkerr := checker.CheckPermissionForWrite(account.Scope, old)
			if checkerr != nil {
				return errors.Wrap(checkerr, "content specific permission check failed")
			}
		}
	}

	if len(added)+len(changed)+len(deleted) == 0 && !contentChanged {
		return errors.New("nothing to update")
	}

	// real update, so check ownership based write access
	if err := old.UsrGrpWritable(account.ID, ""); err != nil {
		return errors.WrapWithCode(err, 403, "cannot write")
	}

	return nil
}
