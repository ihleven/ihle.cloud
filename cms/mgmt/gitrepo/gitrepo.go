package gitrepo

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"net/url"
	"os"
	"path"
	"sync"
	"time"

	"bitbucket.org/hotelplan/webcc-content/cms/content"
	"bitbucket.org/hotelplan/webcc-content/cms/pkg/errors"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type Repo interface {
	GetEntry(path string) (*content.Entry, error)
	Commit(Entrymap, Signature, msg string) error
	TreeAt(path string) *Tree

	load(filename string) ([]byte, error)
	storeMulti(files map[string][]byte, message, author, email string) error
}

func New(configurl string, dataDir string, clone bool, options ...func(*Sitory) error) (*Sitory, error) {
	fmt.Println(" +++ New git repo", configurl, dataDir, clone)
	parsedURL, err := url.Parse(configurl)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't parse repo config url: %s", configurl)
	}

	repo := Sitory{
		url:         "https://" + parsedURL.Host + parsedURL.Path,
		name:        "content",
		branch:      "main",
		username:    parsedURL.User.Username(),
		committer:   Signature{Name: parsedURL.User.Username(), Email: "cms@ih"},
		Directories: map[string]*Tree{},
	}
	repo.password, _ = parsedURL.User.Password()

	if name := parsedURL.Query().Get("name"); name != "" {
		repo.name = name
	}
	if parsedURL.Fragment != "" {
		repo.branch = parsedURL.Fragment
	}

	for _, option := range options {
		err := option(&repo)
		if err != nil {
			return nil, errors.Wrap(err, "Couldn't apply option %T", option)
		}
	}

	// r.Auth = http.BasicAuth{Username: username, Password: password}

	repopath := path.Join(dataDir, "repo_"+repo.name)

	// * file / inmem
	// * clone / open
	// 1) inmem mode implies clone mode
	// 2) file mode muss clonen, wenn option
	// 3) file mode muss clonen, wenn kein lokales verzeichnis vorhanden

	// if mode==open: try to open else
	// if mode==clone or open failed

	if clone {
		err = os.RemoveAll(repopath)
		if err != nil {
			return nil, errors.Wrap(err, "Couldn't delete %s for fresh clone", repopath)
		}
		err = repo.Clone(repopath)
	} else {
		err = repo.Open(repopath, false)
		fmt.Println("repo.Open done", repo.gitRepo)
	}

	if err != nil {
		return nil, err
	}

	// repo.gitRepo, err = git.PlainOpen(repopath)
	// if err != nil {

	// 	if errors.Is(err, git.ErrRepositoryNotExists) {
	// 		for {
	// 			repo.gitRepo, err = git.PlainClone(repopath, false, &git.CloneOptions{
	// 				URL:           repo.url,
	// 				ReferenceName: plumbing.ReferenceName(repo.branch),
	// 				SingleBranch:  true,
	// 				Auth:          &http.BasicAuth{Username: repo.username, Password: repo.password},
	// 				Progress:      os.Stdout,
	// 			})
	// 			if errors.Is(err, git.ErrRepositoryAlreadyExists) {
	// 				fmt.Printf("Couldn't clone repo because path already exists, trying to delete %s\n", repopath)
	// 				rmerr := os.RemoveAll(repopath)
	// 				if rmerr != nil {
	// 					return nil, errors.Wrap(err, "Couldn't delete %s", repopath)
	// 				}
	// 				continue
	// 			}
	// 			if err != nil {
	// 				return nil, errors.Wrap(err, "Couldn't clone git repo")
	// 			}
	// 			break
	// 		}
	// 	} else {

	// 		fmt.Printf("Opening existing git repo at %s\n", repopath)
	// 		return nil, err

	// 	}
	// }

	// calcDirs()
	repo.RootDir = repo.treeAt("")

	return &repo, nil
}

type Sitory struct {
	sync.RWMutex
	// repo storage.ContentRepository
	name     string
	url      string
	branch   string
	username string
	password string

	gitRepo *git.Repository
	// fs      billy.Filesystem

	// name      string
	// Auth      http.BasicAuth
	committer Signature

	RootDir     *Tree
	Directories map[string]*Tree
}

type Signature struct{ Name, Email string }

func WithCommitter(name, email string) func(*Sitory) error {
	return func(repo *Sitory) error {

		_, err := mail.ParseAddress(name + "<" + email + ">")
		if err != nil {
			return errors.Wrap(err, "invalid committer: %s<%>", name, email)
		}

		repo.committer = Signature{Name: name, Email: email}
		return nil
	}
}

// Open will open local repopath, e.g. "file://Users/ih/src/webcc-content/cms/data/repo_content"
func (r *Sitory) Open(repopath string, verbose bool) error {

	repo, err := git.PlainOpen(repopath)
	if err != nil {
		if errors.Is(err, git.ErrRepositoryNotExists) {
			return r.Clone(repopath)
		}
		return err
	}

	r.gitRepo = repo

	if verbose {
		fmt.Println("open", repopath)
		wt, _ := repo.Worktree()
		remotes, _ := repo.Remotes()
		fmt.Println("opened repo:", repo.Storer, wt, remotes)
	}

	return nil
}
func (r *Sitory) Clone(repopath string) error {
	fmt.Println("clone")

	repo, err := git.PlainClone(repopath, false, &git.CloneOptions{
		URL:           r.url,
		ReferenceName: plumbing.ReferenceName(r.branch),
		SingleBranch:  true,
		Auth:          &http.BasicAuth{Username: r.username, Password: r.password},
		Progress:      os.Stdout,
	})
	if err != nil {
		return errors.Wrap(err, "Couldn't clone git repo")
	}

	r.gitRepo = repo
	return nil
}

type Entrymap map[string]*content.Entry

func (r *Sitory) Commit(entrymap Entrymap, author Signature, msg string, ts time.Time) error {

	bytesmap := map[string][]byte{}
	for path, entry := range entrymap {
		if entry == nil {
			bytesmap[path] = nil
			continue
		}
		// fmt.Println("serializing for storage", path)
		bytes, err := entry.SerializeForStorage()
		if err != nil {
			return err
		}

		bytesmap[path] = bytes
	}

	err := r.storeMulti(bytesmap, msg, author.Name, author.Email, ts)
	if err != nil {
		return err
	}
	return nil
}

func (r *Sitory) Worktree() (*git.Worktree, error) {
	return r.gitRepo.Worktree()
}

func (r *Sitory) Status() (git.Status, error) {

	// fmt.Println("Status:")
	wt, err := r.gitRepo.Worktree()
	if err != nil {
		return nil, err
	}
	// fs := wt.Filesystem
	// fd, err := fs.OpenFile("bar.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	// if err != nil {
	// 	return nil, err
	// }
	// defer fd.Close()
	// _, err = fd.Write([]byte("{'foo': 'bar'}"))
	// if err != nil {
	// 	return nil, err
	// }
	status, err := wt.Status()
	if err != nil {
		return nil, err
	}

	fmt.Printf(" status => %t %s\n", status.IsClean(), status.String())
	return status, nil
}

func (repo *Sitory) ListEntries(breakonerror bool) ([]content.Entry, error) {
	repo.RWMutex.RLock()
	defer repo.RWMutex.RUnlock()

	entries := make([]content.Entry, 0)

	for pth, dir := range repo.Directories {
		for _, filename := range dir.Entries {

			fpath := path.Join(pth, filename)

			e, err := repo.GetEntry(fpath)
			if err != nil && breakonerror {
				return nil, errors.Wrap(err, "cannot get entry %s", fpath)
			}

			if err == nil {
				entries = append(entries, *e)
			}

			// entries = append(entries, *e)
		}
	}

	// bytesmap, err := repo.loadMulti("")
	// if err != nil {
	// 	return nil, errors.Wrap(err, "load all entries")
	// }

	// i := 0
	// entries := make([]content.Entry, len(bytesmap))
	// // for path := range bytesmap {
	// // }
	// for path, bytes := range bytesmap {
	// 	fmt.Println("+++++++++++++++++++", path)

	// 	if strings.HasSuffix(path, ".json") {
	// 		err := json.Unmarshal(bytes, &entries[i])
	// 		if err != nil {
	// 			// return nil, errors.WrapWithCode(err, 500, "cannot unmarshal entry")
	// 			fmt.Println(err)
	// 			fmt.Printf("%s\n", bytes)
	// 		}
	// 	} else if strings.HasSuffix(path, ".md") {
	// 		err := content.ParseMarkdownEntryFromStorage(bytes, &entries[i])
	// 		fmt.Println("ParseMarkdownEntryFromStorage", err, path, entries[i].Path)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 	} else {
	// 		continue
	// 	}

	// 	if path != entries[i].Path {
	// 		fmt.Printf("path != entry.path => %s != %s\n", path, entries[i].Path)
	// 		return nil, errors.New("path != entry.path => %s != %s", path, entries[i].Path)
	// 	}

	// 	i++
	// }

	return entries, nil
}

func (r *Sitory) GetBytes(path string, jsonFormat bool) ([]byte, error) {
	r.RWMutex.RLock()
	defer r.RWMutex.RUnlock()

	if jsonFormat {
		e, err := r.GetEntry(path)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(e, "", "    ")
	}
	bytes, err := r.load(path)
	if err != nil {
		return nil, errors.WrapWithCode(err, 404, "not found: %s", path)
	}
	return bytes, nil
}
