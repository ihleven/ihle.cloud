package content

import (
	"reflect"
	"strings"

	"bitbucket.org/hotelplan/webcc-content/cms/pkg/errors"
)

// A folder is a collection of stories that can be used to group your entries of specific content types.
// Do you want to have all your Posts in one place? Creating a folder will allow that.
// You can also use our API to query on all entries from a specific folder to build out overviews in your project.
type Folder struct {
	ContentType `type:"Folder" json:"-"`

	Name string `json:"name,omitempty"`
	Path string `json:"path"`

	// Folder content type restriction
	// With the content-type restriction, you can restrict specific content types within a folder.
	// This allows you to tailor your folders to hold only the content types you need.
	ContentTypesRestriction []string `json:"allowedContentTypes,omitempty"` // []reflect.Type `json:",omitempty"`
	Settings                struct {
		// e.g. Disable visual Editor
		// or   Lock sub folders content type change
		// stored in a settings.json file in the folder?
	} `json:"-"`
	Entries []string `json:"entries,omitempty"`
	Folders []string `json:"folders,omitempty"`
}

func (f *Folder) Clone() interface{} {
	return &Folder{
		Name: f.Name, Path: f.Path,
		ContentTypesRestriction: append(f.ContentTypesRestriction[:0:0], f.ContentTypesRestriction...),
		Entries:                 append(f.Entries[:0:0], f.Entries...),
		Folders:                 append(f.Folders[:0:0], f.Folders...),
	}
}

// prüfen, ob e.Path zur folder passt und der ContentType zur einer evtl. enthaltenen Restriction passt
func (f *Folder) AllowsEntry(e *Entry) error {

	if len(f.ContentTypesRestriction) == 0 {
		return nil
	}

	entrytype := reflect.TypeOf(e.Content)
	if entrytype != nil && entrytype.Kind() == reflect.Pointer {
		entrytype = entrytype.Elem()
	}

	for _, restname := range f.ContentTypesRestriction {

		contenttypeInstance, _ := Instantiate(restname)
		resttype := reflect.TypeOf(contenttypeInstance)
		if resttype != nil && resttype.Kind() == reflect.Pointer {
			resttype = resttype.Elem()
		}

		if entrytype == resttype {
			return nil
		}
	}

	return errors.NewWithCode(400, "Entry ContentType (%s) doesn't match folder restriction (%v)", e.Type, f.ContentTypesRestriction)
}

type Tree struct {
	*Folder
	Folders map[string]Tree `json:"folders"`
}

func (t *Tree) Delete(path string) error {
	head, tail := SplitPath(path)
	tree := t.Folders
	for ; tail != ""; head, tail = SplitPath(tail) {
		tree = tree[head].Folders
	}
	delete(tree, head)
	return nil
}

// SplitPath is like ShiftPath but without the leading slash in the result
func SplitPath(p string) (head, tail string) {

	p = strings.TrimLeft(p, "/")

	if i := strings.Index(p, "/"); i >= 0 {

		return p[0:i], p[i+1:]
	}
	return p, ""
}
