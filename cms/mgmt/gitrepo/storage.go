package gitrepo

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"slices"
	"strings"
	"time"

	"bitbucket.org/hotelplan/webcc-content/cms/content"
	"bitbucket.org/hotelplan/webcc-content/cms/pkg/errors"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/goccy/go-yaml"
)

// GetEntryByPath liefert den entry, der `path` zugeordnet ist.
func (r *Sitory) GetEntry(path string) (*content.Entry, error) {

	r.RWMutex.RLock()
	defer r.RWMutex.RUnlock()

	bytes, err := r.load(path)
	if err != nil {
		return nil, errors.WrapWithCode(err, 404, "not found: %s", path)
	}

	var entry content.Entry
	if strings.HasSuffix(path, ".md") {
		err = content.ParseMarkdownEntryFromStorage(bytes, &entry)
		if err != nil {
			return nil, errors.WrapWithCode(err, 500, "cannot unmarshal entry")
		}
	} else if strings.HasSuffix(path, ".yaml") {
		err = yaml.Unmarshal(bytes, &entry)
		if err != nil {
			return nil, errors.WrapWithCode(err, 500, "cannot unmarshal entry")
		}
	} else {

		err = json.Unmarshal(bytes, &entry)
		if err != nil {
			return nil, errors.WrapWithCode(err, 500, "cannot unmarshal entry")
		}
	}

	if entry.Path != path {
		return nil, errors.New(" entry path (%s) and storage path (%s) differ", entry.Path, path)
	}

	return &entry, nil
}

func (r *Sitory) load(filename string) ([]byte, error) {

	wt, err := r.gitRepo.Worktree()
	if err != nil {
		return nil, err
	}

	f, err := wt.Filesystem.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// fmt.Printf(" => %v %v\n", f, err)
	// all, e := io.ReadAll(f)
	// fmt.Printf("readAll => %s %v\n", all, e)
	return io.ReadAll(f)
}

func (r *Sitory) storeMulti(files map[string][]byte, message, author, email string, now time.Time) error {
	if now.IsZero() {
		now = time.Now()
	}

	wt, err := r.gitRepo.Worktree()
	if err != nil {
		return err
	}
	fs := wt.Filesystem

	filestodelete := []string{}

	for path, content := range files {

		if content == nil {
			filestodelete = append(filestodelete, path)
		} else {
			// if r.fs == nil {
			// 	fmt.Println("r.fs is NULL")
			// }
			f, err := fs.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
			if err != nil {
				return err
			}

			_, err = f.Write(content)
			if err != nil {
				defer f.Close()
				return err
			}

			err = f.Close()
			if err != nil {
				return err
			}
		}
	}
	_, err = wt.Status()
	if err != nil {
		return err
	}
	// for a, b := range status {
	// 	fmt.Println("status:", a, b)
	// }
	slices.Sort(filestodelete)
	slices.Reverse(filestodelete)
	for _, path := range filestodelete {
		err := fs.Remove(path)
		if err != nil {
			return err
		}
	}

	// wt, err := r.GitRepo.Worktree()
	// if err != nil {
	// 	return err
	// }

	err = wt.AddWithOptions(&git.AddOptions{All: true})
	if err != nil {
		return err
	}

	_, err = wt.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  author,
			Email: email,
			When:  now,
		},
		Committer: &object.Signature{
			Name:  r.committer.Name,
			Email: r.committer.Email,
			When:  now,
		},
	})
	if err != nil {
		return errors.Wrap(err, "Couldn't commit worktree")
	}

	err = r.gitRepo.Push(&git.PushOptions{Auth: &http.BasicAuth{Username: r.username, Password: r.password}})
	if err != nil {
		return errors.Wrap(err, "pushing error")
	}

	slog.Info("git repo saving success", "numfiles", len(files), "msg", message, "author", author)

	return nil
}

// tmp for intermal repo
func (r *Sitory) LoadMulti(prefix string) (map[string][]byte, error) {

	// start := time.Now()

	if r.gitRepo == nil {
		return nil, errors.New("gitrepo is nil")
	}

	// branch, err := r.GitRepo.Branch(branch)
	// ref, err = r.GitRepo.Reference(branch.Merge, true)

	var err error
	ref, err := r.gitRepo.Head()
	if err != nil {
		return nil, err
	}

	commit, err := r.gitRepo.CommitObject(ref.Hash())
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't retrieve commit object for %s", ref.Hash())
	}

	// => git ls-tree -r HEAD
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	files := map[string][]byte{}

	tree.Files().ForEach(func(f *object.File) error {
		if strings.HasPrefix(f.Name, prefix) {
			content, err := f.Contents()
			if err != nil {
				return err
			}
			files[f.Name] = []byte(content)
		}
		return nil
	})

	// slog.Info("loaded files from git repo", "numfiles", len(files), "duration", time.Since(start))
	return files, nil
}
