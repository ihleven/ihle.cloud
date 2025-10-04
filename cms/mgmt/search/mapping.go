package search

import (
	"fmt"
	"path"

	"bitbucket.org/hotelplan/webcc-content/cms/content"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/lang/de"
	"github.com/blevesearch/bleve/v2/analysis/lang/en"
	"github.com/blevesearch/bleve/v2/analysis/lang/fr"
	"github.com/blevesearch/bleve/v2/analysis/lang/nl"
	"github.com/blevesearch/bleve/v2/mapping"
)

type augmenter interface {
	AugmentSearchDoc(*Document, Level) (interface{}, error)
}

// IndexDoc creates a document for indexing from a content.Entry.
// If the level is not Off, it will create a basic search.Document with the entry's metadata and path information.
// If the level is Fulltext or Extended, it will call the AugmentSearchDoc method on the entry's content if it implements the augmenter interface.
func IndexDoc(entry *content.Entry, level Level) interface{} {

	if level == Off {
		fmt.Println(" +++ Search indexing is disabled +++")
		return nil
	}

	doc := Document{
		Meta: entry.Meta,
		Dir:  path.Dir(entry.Path),
	}
	for ancestor := path.Dir(entry.Path); ancestor != "."; ancestor = path.Dir(ancestor) {

		doc.Ancestors = append(doc.Ancestors, ancestor)
	}

	if level == Basic {
		return doc
	}

	// level == Fulltext || level == Extended

	if augmenter, ok := entry.Content.(augmenter); ok {
		augmentedDoc, err := augmenter.AugmentSearchDoc(&doc, level)
		if augmentedDoc != nil {
			// if entry.Type == "Destination" {

			// 	bytes, _ := json.MarshalIndent(augmentedDoc, "", "    ")
			// 	fmt.Printf("\n%s\n", bytes)
			// }
			if err != nil {
				fmt.Println(" +++ Error augmenting search doc:", err)
			} else {
				return augmentedDoc
			}
		}
	}

	return doc
}

type Document struct {
	content.Meta
	Dir       string   `json:"dir"`
	Ancestors []string `json:"ancestors"`
	DocumentContent
	// Image string            `json:"img"`
	// Name  map[string]string `json:"name"`
	// Text  map[string]string `json:"text"`

	// Ext struct {
	// 	Image  string `json:"img"`
	// 	Title  string `json:"title"`
	// 	TextDE string `json:"text"`
	// 	TextFR string `json:"text.fr"`
	// 	TextNL string `json:"text.nl"`
	// }
	// Content interface{} `json:"content"`
}

type DocumentContent struct {
	Image string            `json:"img"`
	Title map[string]string `json:"title"`
	Text  map[string]string `json:"text"`
}

var mappingAdapters []func(*mapping.DocumentMapping)

func RegisterMappingAdapter(name func(mapping *mapping.DocumentMapping)) {

	mappingAdapters = append(mappingAdapters, name)
}

func indexMapping(docmappings map[string]*mapping.DocumentMapping) *mapping.IndexMappingImpl {

	/////////////////////////////////////////
	mapping := bleve.NewIndexMapping()
	mapping.TypeField = "type"
	// mapping.DefaultAnalyzer = "en"
	for i, m := range docmappings {
		mapping.AddDocumentMapping(i, m)
	}

	entryMapping := CreateEntryDocMapping()
	for _, adpt := range mappingAdapters {
		// fmt.Println("+++++++++++++++++++++++++name", adpt)
		adpt(entryMapping)
	}
	mapping.AddDocumentMapping("entry", entryMapping)

	mapping.DefaultType = "entry"
	mapping.DefaultMapping = entryMapping

	// a generic reusable mapping for english text
	// englishTextFieldMapping := bleve.NewTextFieldMapping()
	// englishTextFieldMapping.Analyzer = en.AnalyzerName

	// a generic reusable mapping for keyword text
	// keywordFieldMapping := bleve.NewKeywordFieldMapping() // bleve.NewTextFieldMapping()
	// keywordFieldMapping.Analyzer = keyword.Name
	// pointFieldMapping := bleve.NewGeoPointFieldMapping()

	// docMapping.AddFieldMappingsAt("id", keywordFieldMapping)
	// docMapping.AddFieldMappingsAt("type", keywordFieldMapping)
	// docMapping.AddFieldMappingsAt("dest", keywordFieldMapping)
	// docMapping.AddFieldMappingsAt("filter", keywordFieldMapping)
	// docMapping.AddFieldMappingsAt("attrs", keywordFieldMapping)
	// docMapping.AddFieldMappingsAt("image", keywordFieldMapping)
	// docMapping.AddFieldMappingsAt("coords", pointFieldMapping)

	// authorEmailFieldMapping := bleve.NewTextFieldMapping()
	// authorEmailFieldMapping.IncludeInAll = false
	// author.AddFieldMappingsAt("email", authorEmailFieldMapping)
	// docMapping.AddSubDocumentMapping("name", langMapMapping)
	// docMapping.AddSubDocumentMapping("txt_short", langMapMapping)
	// docMapping.AddSubDocumentMapping("txt_teaser", langMapMapping)
	// docMapping.AddSubDocumentMapping("txt_long", langMapMapping)

	// beerMapping := bleve.NewDocumentMapping()

	// // name
	// beerMapping.AddFieldMappingsAt("name", englishTextFieldMapping)

	// // description
	// beerMapping.AddFieldMappingsAt("description",
	// 	englishTextFieldMapping)

	// beerMapping.AddFieldMappingsAt("type", keywordFieldMapping)
	// beerMapping.AddFieldMappingsAt("style", keywordFieldMapping)
	// beerMapping.AddFieldMappingsAt("category", keywordFieldMapping)

	// breweryMapping := bleve.NewDocumentMapping()
	// breweryMapping.AddFieldMappingsAt("name", englishTextFieldMapping)
	// breweryMapping.AddFieldMappingsAt("description", englishTextFieldMapping)

	// indexMapping := bleve.NewIndexMapping()
	// indexMapping.AddDocumentMapping("beer", beerMapping)
	// indexMapping.AddDocumentMapping("brewery", breweryMapping)

	// indexMapping.TypeField = "type"
	// indexMapping.DefaultAnalyzer = "en"

	// bytes, _ := json.MarshalIndent(mapping, "", "    ")
	// fmt.Printf("%s\n", bytes)

	return mapping

}

func CreateEntryDocMapping() *mapping.DocumentMapping {

	keyword := bleve.NewKeywordFieldMapping()
	keyword.IncludeInAll = false
	date := bleve.NewDateTimeFieldMapping()

	mapping := bleve.NewDocumentMapping()
	mapping.Dynamic = false
	mapping.AddFieldMappingsAt("path", keyword)
	mapping.AddFieldMappingsAt("dir", keyword)
	mapping.AddFieldMappingsAt("ancestors", keyword)
	mapping.AddFieldMappingsAt("locale", keyword)
	mapping.AddFieldMappingsAt("version", keyword)
	mapping.AddFieldMappingsAt("status", keyword)
	mapping.AddFieldMappingsAt("slug", keyword)
	mapping.AddFieldMappingsAt("full_slug", keyword)

	// mapping.AddFieldMappingsAt("owner", keyword)
	// mapping.AddFieldMappingsAt("group", keyword)
	mapping.AddFieldMappingsAt("created", date)
	mapping.AddFieldMappingsAt("modified", date)
	mapping.AddFieldMappingsAt("published", date)
	// mapping.AddFieldMappingsAt("permissions", keyword)

	mapping.AddFieldMappingsAt("id", keyword)
	// mapping.AddFieldMappingsAt("collection", keyword)
	// mapping.AddFieldMappingsAt("space", keyword)

	// mapping.AddFieldMappingsAt("repo", keyword)
	mapping.AddFieldMappingsAt("name", keyword)
	mapping.AddFieldMappingsAt("notes", bleve.NewTextFieldMapping())
	mapping.AddFieldMappingsAt("tags", keyword)

	// mapping.AddFieldMappingsAt("mime", keyword)
	mapping.AddFieldMappingsAt("type", keyword)

	mapping.AddFieldMappingsAt("img", keyword)
	mapping.AddSubDocumentMapping("title", Langmappping())
	mapping.AddSubDocumentMapping("text", Langmappping())

	content := bleve.NewDocumentMapping()
	mapping.AddSubDocumentMapping("content", content)

	return mapping
}

func Langmappping() *mapping.DocumentMapping {

	deFieldMapping := bleve.NewTextFieldMapping()
	deFieldMapping.Analyzer = de.AnalyzerName
	deFieldMapping.Store = true

	enFieldMapping := bleve.NewTextFieldMapping()
	enFieldMapping.Analyzer = en.AnalyzerName

	frFieldMapping := bleve.NewTextFieldMapping()
	frFieldMapping.Analyzer = fr.AnalyzerName

	nlFieldMapping := bleve.NewTextFieldMapping()
	nlFieldMapping.Analyzer = nl.AnalyzerName

	langMapMapping := bleve.NewDocumentMapping()
	langMapMapping.AddFieldMappingsAt("de", deFieldMapping)
	langMapMapping.AddFieldMappingsAt("en", enFieldMapping)
	langMapMapping.AddFieldMappingsAt("fr", frFieldMapping)
	langMapMapping.AddFieldMappingsAt("nl", nlFieldMapping)

	return langMapMapping
}
