package gitrepo

import (
	"net/url"
	"os"
	"path"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type RepositoryDep struct {
	//   ///
	//  ////
	// // //
	////  //
	///   //
	sync.RWMutex
	// repo storage.ContentRepository

	GitRepo *git.Repository
	// fs      billy.Filesystem

	name string
	Auth http.BasicAuth

	RootDir     *Tree
	Directories map[string]*Tree
}

func NewRepoDep(rawurl string, dataDir string) (*RepositoryDep, error) {

	r := RepositoryDep{name: "content", Directories: map[string]*Tree{}}

	parsedURL, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	if name := parsedURL.Query().Get("name"); name != "" {
		r.name = name
	}

	giturl := "https://" + parsedURL.Host + parsedURL.Path
	branch := parsedURL.Fragment
	username := parsedURL.User.Username()
	password, _ := parsedURL.User.Password()
	r.Auth = http.BasicAuth{Username: username, Password: password}

	repopath := path.Join(dataDir, "repo")

	r.GitRepo, err = git.PlainClone(repopath, false, &git.CloneOptions{
		URL:           giturl,
		ReferenceName: plumbing.ReferenceName(branch),
		SingleBranch:  true,
		Auth:          &http.BasicAuth{Username: username, Password: password},
		Progress:      os.Stdout,
	})
	if err != nil {
		return nil, err
	}

	// calcDirs()
	// r.RootDir = r.TreeAt("")

	return &r, nil
}

// func (r *Repository) Raw() storage.ContentRepository {
// 	return r.repo
// }

// type Entrymap map[string]*content.Entry

// func (r *Repository) Commit(entrymap Entrymap, usr, msg string) error {

// 	bytesmap := map[string][]byte{}
// 	for path, entry := range entrymap {
// 		if entry == nil {
// 			bytesmap[path] = nil
// 			continue
// 		}
// 		bytes, err := entry.SerializeForStorage()
// 		if err != nil {
// 			return err
// 		}
// 		bytesmap[path] = bytes
// 	}

// 	err := r.StoreMulti(bytesmap, msg, usr)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (r *Repository) StoreMulti(files map[string][]byte, message, author string) error {

// 	wt, err := r.GitRepo.Worktree()
// 	if err != nil {
// 		return err
// 	}
// 	fs := wt.Filesystem

// 	filestodelete := []string{}

// 	for path, content := range files {

// 		if content == nil {
// 			filestodelete = append(filestodelete, path)
// 		} else {
// 			// if r.fs == nil {
// 			// 	fmt.Println("r.fs is NULL")
// 			// }
// 			f, err := fs.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
// 			if err != nil {
// 				return err
// 			}

// 			_, err = f.Write(content)
// 			if err != nil {
// 				defer f.Close()
// 				return err
// 			}

// 			err = f.Close()
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	status, err := wt.Status()
// 	if err != nil {
// 		return err
// 	}
// 	for a, b := range status {
// 		fmt.Println("status:", a, b)
// 	}
// 	slices.Sort(filestodelete)
// 	slices.Reverse(filestodelete)
// 	for _, path := range filestodelete {
// 		err := fs.Remove(path)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	// wt, err := r.GitRepo.Worktree()
// 	// if err != nil {
// 	// 	return err
// 	// }

// 	err = wt.AddWithOptions(&git.AddOptions{All: true})
// 	if err != nil {
// 		return err
// 	}

// 	addr, err := mail.ParseAddress(author)
// 	if err != nil {
// 		addr = &mail.Address{Name: author, Address: "invalidauthor@interhome"}
// 	}

// 	_, err = wt.Commit(message, &git.CommitOptions{
// 		Author: &object.Signature{
// 			Name:  addr.Name,
// 			Email: addr.Address,
// 			When:  time.Now(),
// 		},
// 		Committer: &object.Signature{
// 			Name:  r.name,
// 			Email: r.name + "@ih",
// 			When:  time.Now(),
// 		},
// 	})
// 	if err != nil {
// 		return errors.Wrap(err, "Couldn't commit worktree")
// 	}

// 	err = r.GitRepo.Push(&git.PushOptions{Auth: &r.Auth})
// 	if err != nil {
// 		return errors.Wrap(err, "pushing error")
// 	}

// 	slog.Info("git repo saving success", "numfiles", len(files), "msg", message, "author", author)

// 	return nil
// }

// //   - ContentEntries(EntryOptions) ([]content.Entry, error) //
// //     (gefilterete) Liste von Entries
// func (repo *Repository) ListEntries() ([]content.Entry, error) {

// 	repo.RWMutex.RLock()
// 	defer repo.RWMutex.RUnlock()

// 	bytesmap, err := repo.loadMulti("")
// 	if err != nil {
// 		return nil, errors.Wrap(err, "load all entries")
// 	}

// 	i := 0
// 	entries := make([]content.Entry, len(bytesmap))
// 	// for path := range bytesmap {
// 	// }
// 	for path, bytes := range bytesmap {
// 		// fmt.Println("+++++++++++++++++++", path)

// 		if strings.HasSuffix(path, ".json") {
// 			err := json.Unmarshal(bytes, &entries[i])
// 			if err != nil {
// 				// return nil, errors.WrapWithCode(err, 500, "cannot unmarshal entry")
// 				fmt.Println(err)
// 				fmt.Printf("%s\n", bytes)
// 			}
// 		} else if strings.HasSuffix(path, ".md") {
// 			err := content.ParseMarkdownEntryFromStorage(bytes, &entries[i])
// 			fmt.Println("ParseMarkdownEntryFromStorage", err, path, entries[i].Path)
// 			if err != nil {
// 				return nil, err
// 			}
// 		} else {
// 			continue
// 		}

// 		if path != entries[i].Path {
// 			fmt.Printf("path != entry.path => %s != %s\n", path, entries[i].Path)
// 			return nil, errors.New("path != entry.path => %s != %s", path, entries[i].Path)
// 		}

// 		i++
// 	}

// 	return entries, nil
// }

// func (rp *Repository) Folder(path string) *content.Entry {
// 	// return rp.Directories[path]
// 	return nil
// }

// type Tree struct {
// 	Name    string           `json:"name,omitempty"`
// 	Path    string           `json:"path"`
// 	Entries []string         `json:"entries,omitempty"`
// 	Dirs    map[string]*Tree `json:"folders,omitempty"`
// }

// func (rp *Repository) TreeAt(p string) *Tree {

// 	dir := &Tree{
// 		Name: path.Base(p),
// 		Path: p,

// 		Entries: []string{},
// 		Dirs:    map[string]*Tree{},
// 	}
// 	rp.Directories[dir.Path] = dir

// 	rp.RLock()
// 	defer rp.RUnlock()

// 	wt, err := rp.GitRepo.Worktree()
// 	if err != nil {
// 		return nil
// 	}
// 	fs := wt.Filesystem

// 	infos, err := fs.ReadDir(p)
// 	if err != nil {
// 		return nil
// 	}

// 	for _, info := range infos {
// 		if strings.HasPrefix(info.Name(), ".") {
// 			continue
// 		}
// 		if info.IsDir() {
// 			dir.Dirs[info.Name()] = rp.TreeAt(path.Join(p, info.Name()))
// 			// dir.Dirs = append(dir.Dirs, rp.TreeAt(path.Join(path, info.Name())))
// 		} else {
// 			dir.Entries = append(dir.Entries, info.Name())
// 		}
// 	}

// 	return dir
// }

// // * Entry(key string, a *permission.User, resolve content.Resolv) (*content.Entry, error)
// // GetEntryByPath liefert den entry, der `path` zugeordnet ist.
// func (r *Repository) GetEntry(path string) (*content.Entry, error) {

// 	r.RWMutex.RLock()
// 	defer r.RWMutex.RUnlock()

// 	bytes, err := r.load(path)
// 	if err != nil {
// 		return nil, errors.WrapWithCode(err, 404, "not found: %s", path)
// 	}

// 	var entry content.Entry
// 	if strings.HasSuffix(path, ".md") {
// 		err = content.ParseMarkdownEntryFromStorage(bytes, &entry)
// 		if err != nil {
// 			return nil, errors.WrapWithCode(err, 500, "cannot unmarshal entry")
// 		}
// 	} else {

// 		err = json.Unmarshal(bytes, &entry)
// 		if err != nil {
// 			return nil, errors.WrapWithCode(err, 500, "cannot unmarshal entry")
// 		}
// 	}

// 	return &entry, nil
// }
// func (r *Repository) GetBytes(path string, jsonFormat bool) ([]byte, error) {
// 	r.RWMutex.RLock()
// 	defer r.RWMutex.RUnlock()

// 	if jsonFormat {
// 		e, err := r.GetEntry(path)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return json.MarshalIndent(e, "", "    ")
// 	}
// 	bytes, err := r.load(path)
// 	if err != nil {
// 		return nil, errors.WrapWithCode(err, 404, "not found: %s", path)
// 	}
// 	return bytes, nil
// }

// func (r *Repository) Status() (git.Status, error) {

// 	fmt.Println("Status:")
// 	wt, err := r.GitRepo.Worktree()
// 	if err != nil {
// 		return nil, err
// 	}
// 	fs := wt.Filesystem
// 	fd, err := fs.OpenFile("bar.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer fd.Close()
// 	_, err = fd.Write([]byte("{'foo': 'bar'}"))
// 	if err != nil {
// 		return nil, err
// 	}
// 	status, err := wt.Status()
// 	if err != nil {
// 		return nil, err
// 	}

// 	fmt.Printf(" status => %t %s\n", status.IsClean(), status.String())
// 	return status, nil
// }

// func (r *Repository) load(filename string) ([]byte, error) {
// 	w, err := r.GitRepo.Worktree()
// 	if err != nil {
// 		return nil, err
// 	}

// 	f, err := w.Filesystem.OpenFile(filename, os.O_RDWR, 0666)
// 	if err != nil {
// 		return nil, err
// 	}

// 	fmt.Printf(" => %v %v\n", f, err)
// 	all, e := io.ReadAll(f)
// 	fmt.Printf("readAll => %s %v\n", all, e)
// 	return all, e
// }

// func (r *Repository) loadMulti(prefix string) (map[string][]byte, error) {

// 	start := time.Now()

// 	if r.GitRepo == nil {
// 		return nil, errors.New("gitrepo is nil")
// 	}

// 	// branch, err := r.GitRepo.Branch(branch)
// 	// ref, err = r.GitRepo.Reference(branch.Merge, true)

// 	var err error
// 	ref, err := r.GitRepo.Head()
// 	if err != nil {
// 		return nil, err
// 	}

// 	commit, err := r.GitRepo.CommitObject(ref.Hash())
// 	if err != nil {
// 		return nil, errors.Wrap(err, "Couldn't retrieve commit object for %s", ref.Hash())
// 	}

// 	// => git ls-tree -r HEAD
// 	tree, err := commit.Tree()
// 	if err != nil {
// 		return nil, err
// 	}

// 	files := map[string][]byte{}

// 	tree.Files().ForEach(func(f *object.File) error {
// 		if strings.HasPrefix(f.Name, prefix) {
// 			content, err := f.Contents()
// 			if err != nil {
// 				return err
// 			}
// 			files[f.Name] = []byte(content)
// 		}
// 		return nil
// 	})

// 	// slog.Info("loaded files from git repo", "numfiles", len(files), "duration", time.Since(start))
// 	return files, nil
// }
