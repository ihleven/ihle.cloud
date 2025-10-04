package search

import (
	"fmt"
	"path"
	"path/filepath"
	"time"

	"bitbucket.org/hotelplan/webcc-content/cms/content"
	"bitbucket.org/hotelplan/webcc-content/cms/pkg/errors"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
)

type Level int

const (
	Off Level = iota
	Basic
	Fulltext
	Extended
)

func (l Level) String() string {
	if l < 0 || l > 3 {
		return "invalid level"
	}
	return []string{"off", "basic", "fulltext", "extended"}[l]
}

func ParseLevel(level string) Level {
	for i, l := range []string{"off", "basic", "fulltext", "extended"} {
		if l == level {
			return Level(i)
		}
	}
	return 0
}

func NewEngine(config Config, mappings map[string]*mapping.DocumentMapping) (engine *Engine, err error) {

	engine = &Engine{
		Config: config,
	}

	fmt.Printf("SEARCH ENGINE config: %+v\n", config)

	if config.Level == Off {
		return engine, nil
	}

	engine.mapping = indexMapping(mappings)

	if config.IndexName != "" {
		path := filepath.Join(config.DataDirPath, config.IndexName)
		engine.BleveIndex, err = bleve.Open(path)
		if err == bleve.ErrorIndexPathDoesNotExist {
			fmt.Printf(" *** %s does not exist, creating new index... ", path)
			engine.BleveIndex, err = bleve.New(path, engine.mapping)
		}
		if err != nil {
			return nil, err
		}
		fmt.Printf(" *** bleve index at %s ***\n", path)

	} else {
		engine.BleveIndex, err = bleve.NewMemOnly(engine.mapping)

		fmt.Println(" *** bleve in mem index ***")
	}
	if err != nil {
		return nil, err
	}

	return engine, nil
}

// Engine
type Engine struct {
	BleveIndex bleve.Index
	mapping    *mapping.IndexMappingImpl
	Config
}

type Config struct {
	Level       Level
	DataDirPath string
	IndexName   string
	// IndexPath    string
	//
	// Languages []string
}

func (e *Engine) Mapping() mapping.IndexMapping {
	return e.mapping
}

func (e *Engine) Index(entries ...content.Entry) error {
	if e.Config.Level == Off {
		return nil
	}

	if len(entries) == 1 {
		entry := &entries[0]

		doc := IndexDoc(entry, e.Config.Level)
		// fmt.Println("doc:", doc)
		return e.BleveIndex.Index(entry.Path, doc)
	}
	err := e.putEntries(entries)
	if err != nil {
		return errors.Wrap(err, "Error in put entries")
	}
	err = e.PutFolders(entries)
	if err != nil {
		return errors.Wrap(err, "Error in put folders")
	}

	return nil
}

func (e *Engine) putEntries(entries []content.Entry) error {

	start := time.Now()

	batch := e.BleveIndex.NewBatch()

	for i := range entries {
		// fmt.Printf("entry %d", i)

		entry := &entries[i]
		fmt.Printf("\033[2K\rENTRY %d: %s", i, entry.Path)
		doc := IndexDoc(entry, e.Config.Level)
		// fmt.Println("TypeField:", e.mapping.TypeField, doc)

		// // if _, ok := doc.(*Document); ok {
		// bytes, _ := json.MarshalIndent(doc, "", "    ")
		// if strings.Contains(string(bytes), "artworks") {
		// 	if bd, ok := doc.(interface{ BleveType() string }); ok {
		// 		fmt.Printf(" ---- %s\n", bytes)
		// 		fmt.Printf(" ---- %s\n", bd.BleveType())
		// 	} else {
		// 		fmt.Printf(" ---- %s\n", bytes)
		// 	}
		// } else {
		// 	continue
		// }
		// // }

		err := batch.Index(entry.Path, doc)
		if err != nil {
			return err
		}
		// if big.NewInt(int64(i)).ProbablyPrime(0) {
		// 	fmt.Printf("\033[2K\rIndexed %s", entry.ID)
		// }
	}
	fmt.Printf("\033[2K\rENGINE: Indexing batch with %d entries in %s", len(entries), time.Since(start))
	intermediate := time.Now()
	err := e.BleveIndex.Batch(batch)
	if err != nil {
		return err
	}
	fmt.Printf("\033[2K\rENGINE: => Indexed %d entries in %s\n", len(entries), time.Since(intermediate))
	return nil
}

func (e *Engine) PutFolders(entries []content.Entry) error {

	start := time.Now()
	foldermap := map[string]struct{}{}
	for i := range entries {
		entry := &entries[i]
		for dir := path.Dir(entry.Path); dir != "."; dir = path.Dir(dir) {

			foldermap[dir] = struct{}{}
		}
	}

	batch := e.BleveIndex.NewBatch()

	for dir := range foldermap {

		d := Document{
			Meta: content.Meta{Path: dir, Name: path.Base(dir), MIME: "application/json", Type: "Dir"},
			Dir:  path.Dir(dir),
		}
		for ancestor := d.Dir; ancestor != "."; ancestor = path.Dir(ancestor) {

			d.Ancestors = append(d.Ancestors, ancestor)
		}

		fmt.Printf("\033[2K\rFOLDER %s", dir)

		err := batch.Index(dir, d)
		if err != nil {
			return err
		}
	}
	fmt.Printf("\033[2K\rENGINE: Indexing batch with %d folders in %s", len(foldermap), time.Since(start))
	intermediate := time.Now()
	err := e.BleveIndex.Batch(batch)
	if err != nil {
		return err
	}

	fmt.Printf("\033[2K\rENGINE: => Indexed %d folders in %s\n", len(foldermap), time.Since(intermediate))
	return nil
}
