package content_test

import (
	"net/url"
	"path"
	"strings"
	"testing"

	"bitbucket.org/hotelplan/webcc-content/cms/content"

	"github.com/stretchr/testify/assert"
)

func Test_Patch_Diff(t *testing.T) {

	a := content.Meta{
		ID:      "eins",
		Space:   "website",
		Version: "DRAFT",
		Status:  "",
		Access: content.Access{
			Owner: "ihle",
		},
	}
	b := content.Meta{
		ID:      "eins",
		Space:   "",
		Version: "PUBLISHED",
		Status:  "ACTIVE",
		Access: content.Access{
			Owner: "ihlem",
		},
	}

	added, updated, deleted, err := a.Patch(b, content.DIFF)

	assert.NoError(t, err)
	assert.Equal(t, []string{"Space"}, added)
	assert.Equal(t, []string{"Version"}, updated)
	assert.Equal(t, []string{"Status"}, deleted)

}

func Test_Patch_Update(t *testing.T) {

	a := content.Meta{
		ID:      "eins",
		Space:   "website",
		Version: "DRAFT",
		Status:  "",
	}
	b := content.Meta{
		ID:      "eins",
		Space:   "",
		Version: "PUBLISHED",
		Status:  "ACTIVE",
	}

	added, updated, deleted, err := a.Patch(b, content.UPDATE)

	assert.NoError(t, err)
	assert.Equal(t, []string{"Status"}, added)
	assert.Equal(t, []string{"Version"}, updated)
	assert.Equal(t, []string{"Space"}, deleted)

}
func Test_Patch_Patch(t *testing.T) {

	a := content.Meta{
		ID:      "eins",
		Space:   "website",
		Version: "DRAFT",
		Status:  "",
	}
	b := content.Meta{
		ID:      "eins",
		Space:   "",
		Version: "PUBLISHED",
		Status:  "ACTIVE",
	}

	added, updated, deleted, err := a.Patch(b, content.PATCH)

	assert.NoError(t, err)
	assert.Equal(t, []string{"Status"}, added)
	assert.Equal(t, []string{"Version"}, updated)
	assert.Equal(t, []string(nil), deleted)

}

func Test_LogicalKey(t *testing.T) {

	tests := []struct {
		Name, Space, Collection, ID, Locale, Version, Key string
	}{
		{"eins", "website", "pages", "pages/contact", "de-DE", "draft", "website://draft@pages/pages/contact#de-DE"},
		{"zwei", "website", "", "pages/contact", "de-DE", "draft", "website://draft@/pages/contact#de-DE"},
		{"drei", "website", "pages", "contact", "de-DE", "", "website://pages/contact#de-DE"},
		{"vier", "website", "", "contact", "de-DE", "", "website:///contact#de-DE"},
		{"fünf", "website", "contact", "", "de-DE", "", "website://contact#de-DE"},
		{"sechs", "website", "contact", "", "de-DE", "", "website://contact/#de-DE"},
		{"sieben", "website", "pages", "contact", "", "draft:pn13091", "website://draft:pn13091@pages/contact"},
		{"acht", "website", "", "", "foo", "", "website:space::pages/contact?fupp=dot#foo"},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			url, _ := url.Parse(tt.Key)
			if url.Opaque == "" {
				t.Errorf("Opaque is empty")
			}
			if url.Scheme != tt.Space {
				t.Errorf("Space: got %s, want %s", url.Scheme, tt.Space)
			}
			if url.Host != tt.Collection {
				t.Errorf("Collection: got %s, want %s", url.Host, tt.Collection)
			}
			path := strings.TrimPrefix(url.Path, "/")
			if path != tt.ID {
				t.Errorf("ID: got %s, want %s", path, tt.ID)
			}
			if url.Fragment != tt.Locale {
				t.Errorf("Locale: got %s, want %s", url.Fragment, tt.Locale)
			}
			if url.User.String() != tt.Version {
				t.Errorf("Version: got %s, want %s", url.User.String(), tt.Version)
			}
		})
	}
}

func Test_Path(t *testing.T) {

	tests := []struct {
		Name, Repo, Path, Key, Locale, Version, Suffix, URL string
	}{
		{"eins", "content", "pages/contact/de-DE.json", "pages/contact", "de-DE", "draft", ".json", "content:pages/contact/de-DE.json?version=draft#de-DE"},
		{"zwei", "content", "pages/contact/de-DE", "pages/contact", "", "draft", "", "content:pages/contact/de-DE?version=draft"},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			url, _ := url.Parse(tt.URL)
			if url.Scheme != tt.Repo {
				t.Errorf("Scheme/Repo: got %s, want %s", url.Scheme, tt.Repo)
			}
			if url.Opaque != tt.Path {
				t.Errorf("Path: got %s, want %s", url.Opaque, tt.Path)
			}
			if url.Fragment != tt.Locale {
				t.Errorf("Fragment/Locale: got %s, want %s", url.Fragment, tt.Locale)
			}
			if url.Query().Get("version") != tt.Version {
				t.Errorf("Version: got %s, want %s", url.Query().Get("version"), tt.Version)
			}
			if path.Ext(url.Opaque) != tt.Suffix {
				t.Errorf("Suffix: got %s, want %s", path.Ext(url.Opaque), tt.Suffix)
			}
		})
	}
}

func Test_Permissions(t *testing.T) {
	tests := []struct {
		Permissions string
		Mode        content.Mode
	}{
		{"-rwxrwxrwx", 256 + 128 + 64 + 32 + 16 + 8 + 4 + 2 + 1},
		{"---------x", 1},
		{"-------r--", content.ReadOth},
	}

	for _, tt := range tests {
		// fmt.Println("mode:", tt.Mode, "permssions:", tt.Permissions)

		if tt.Mode.String() != tt.Permissions {
			t.Errorf("got %s, want %s", tt.Mode.String(), tt.Permissions)
		}

	}
}

// func Test_UsrGrpPermissions(t *testing.T) {

// 	tests := []struct {
// 		Access content.Access
// 	}{
// 		{Access: content.Access{"", "", time.time{}, time.time{}, time.time{}, "rwxtwxrwx"}},
// 	}

// 	for _, tt := range tests {

// 	}
// }

func Test_PathNew(t *testing.T) {

	tests := []struct {
		URL, Repo, Path string
	}{
		{"/path/to/entry.json", "", "path/to/entry.json"},
		{"content/pages/contact/de-DE.json", "", ""},
		{"entry://content/pages/contact/de-DE.json", "content", "pages/contact/de-DE.json"},
	}

	for _, tt := range tests {
		t.Run(tt.URL, func(t *testing.T) {
			url, _ := url.Parse(tt.URL)
			if url.Scheme != "" && url.Scheme != "entry" {
				t.Errorf("invalid Scheme: %q", url.Scheme)
			}
			if url.Opaque != "" {
				t.Errorf("URL is opaque but should not be: %s", url.Opaque)
			}

		})
	}
}

func Test_URL(t *testing.T) {

	tests := []struct {
		Name, URL, Scheme, Opaque, Username, Password, Host, Path, Query, Fragment string
	}{
		{"eins", "website://draft@pages/pages/contact#de-DE", "website", "", "draft", "", "pages", "/pages/contact", "", "de-DE"},
		{"eins", "website:draft@pages/pages/contact#de-DE", "website", "draft@pages/pages/contact", "", "", "", "", "", "de-DE"},
		{"eins", "website:draft@pages/pages/contact?foo&fr#de-DE", "website", "draft@pages/pages/contact", "", "", "", "", "foo&fr", "de-DE"},
		{"eins", "content:/pages/contact/de-DE-draft.json", "content", "", "", "", "", "/pages/contact/de-DE-draft.json", "", ""},
		{"eins", "entry:/pages/contact/de-DE-draft.json", "entry", "", "", "", "", "/pages/contact/de-DE-draft.json", "", ""},
		{"eins", "entry:///pages/contact/de-DE-draft.json", "entry", "", "", "", "", "/pages/contact/de-DE-draft.json", "", ""},
		{"eins", "entry://content/pages/contact/de-DE-draft.json", "entry", "", "", "", "content", "/pages/contact/de-DE-draft.json", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			url, err := url.Parse(tt.URL)
			if err != nil {
				t.Errorf("Parse error: %s", err.Error())
			}
			if url.Opaque != tt.Opaque {
				t.Errorf("Opaque: want %s, got %s", tt.Opaque, url.Opaque)
			}
			if url.Scheme != tt.Scheme {
				t.Errorf("Scheme: want %s, got %s", tt.Scheme, url.Scheme)
			}
			if url.User.Username() != tt.Username {
				t.Errorf("Username: want %s, got %s", tt.Username, url.User.Username())
			}
			pwd, _ := url.User.Password()
			if pwd != tt.Password {
				t.Errorf("Password: want %s, got %s", tt.Password, pwd)
			}
			if url.Host != tt.Host {
				t.Errorf("Host: want %s, got %s", tt.Host, url.Host)
			}
			if url.Path != tt.Path {
				t.Errorf("Path: want %s, got %s", tt.Path, url.Path)
			}
			if url.RawQuery != tt.Query {
				t.Errorf("Query: want %s, got %s", tt.Query, url.RawQuery)
			}
			if url.Fragment != tt.Fragment {
				t.Errorf("Fragment: want %s, got %s", tt.Fragment, url.Fragment)
			}
		})
	}
}
