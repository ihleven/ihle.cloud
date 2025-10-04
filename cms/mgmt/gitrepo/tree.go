package gitrepo

import (
	"path"
	"strings"
)

type Tree struct {
	Name    string           `json:"name,omitempty"`
	Path    string           `json:"path"`
	Entries []string         `json:"entries,omitempty"`
	Dirs    map[string]*Tree `json:"folders,omitempty"`
}

// treeAt konstruiert den Baum ausgehend von path, d.h. path = "" erzeugt den kompletten Baum
func (rp *Sitory) treeAt(p string) *Tree {

	dir := &Tree{
		Name: path.Base(p),
		Path: p,

		Entries: []string{},
		Dirs:    map[string]*Tree{},
	}
	rp.Directories[dir.Path] = dir

	rp.RLock()
	defer rp.RUnlock()

	wt, err := rp.gitRepo.Worktree()
	if err != nil {
		return nil
	}
	fs := wt.Filesystem

	infos, err := fs.ReadDir(p)
	if err != nil {
		return nil
	}

	for _, info := range infos {
		if strings.HasPrefix(info.Name(), ".") {
			continue
		}
		if info.IsDir() {
			dir.Dirs[info.Name()] = rp.treeAt(path.Join(p, info.Name()))
			// dir.Dirs = append(dir.Dirs, rp.TreeAt(path.Join(path, info.Name())))
		} else {
			dir.Entries = append(dir.Entries, info.Name())
		}
	}

	return dir
}
