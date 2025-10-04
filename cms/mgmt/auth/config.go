package auth

import "bitbucket.org/hotelplan/webcc-content/cms/content"

type Config struct {
	content.ContentType `type:"Config" folder:"" json:"-"`
	Folders             map[string]*content.Folder
	Roles               []Role `json:"roles"`
}

func (c *Config) Clone() interface{} {
	clone := Config{Folders: make(map[string]*content.Folder)} //, Roles: append(c.Roles[:0:0], c.Roles...)}
	for k, v := range c.Folders {
		folderclone := v.Clone()
		clone.Folders[k] = folderclone.(*content.Folder)
	}
	return &clone
}

// Roles sind eine konfigurative Zusammenfassung von Permissions
type Role struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}
