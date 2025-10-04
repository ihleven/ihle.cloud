

* A repository has one main worktree (if it’s not a bare repository) and zero or more linked worktrees.
* the "main worktree" prepared by git-init[1] or git-clone[1]. 

type Entrymap map[string]*content.Entry


### ListEntries() ([]content.Entry, error)

### GetEntry(path string) (*content.Entry, error)
### GetBytes(path string, jsonFormat bool) ([]byte, error)
### Status() (git.Status, error)

### load(filename string) ([]byte, error)
### loadMulti(prefix string) (map[string][]byte, error)

###  TreeAt(path string) *Tree

### Folder(path string) *content.Entry
????

### StoreMulti(map[string][]byte, msg, a string) error 

### Commit(Entrymap, usr, msg string) error