package art

import (
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/lang/de"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/interhome-group/cms/content"
	"github.com/interhome-group/cms/mgmt/search"
)

// Complete Works
type Work struct { // => ARTWORK
	content.EntryContent `type:"Work"  json:"-" mimetype:"application/json"`

	ID          int `json:"id"` //
	Dir         string
	Foto        int    // foto_id |
	Serie       string // serie_id
	SerieNummer int    // serie_nr
	Status      string `json:"status"` // status |

	Title       string  `json:"title"`                                  // titel |
	Description string  `json:"desc"`                                   //
	Year        int     `json:"year"`                                   // jahr    |
	Form        string  `json:"form" enum:"painting,drawing,sculpture"` // gattung | drawing, painting, sculpture -> art form
	Genre       string  `json:"genre"`                                  // in der bildenden Kunst ein thematisch-inhaltliches Gebiet, wie Historie, Landschaft, Porträt oder Stillleben
	Style       string  `json:"kunststil"`                              // dt: Kunststil, en: art movement
	Medium      string  `json:"medium"`                                 // technik | Oel, Aquarell, Pastell, Radierung, Buntstifte, Tinte, Siebdruck
	Support     string  `json:"support" `                               // träger  | Leinwand, Papier, Holz,
	Height      int     `json:"height" `                                // höhe    |
	Width       int     `json:"width"`                                  // breite  |
	Depth       int     `json:"depth"`                                  // tiefe   |
	Area        float64 `json:"area"`                                   // fläche  | in Quadratmeter
	Remark      string  `json:"remark"`                                 // anmerkungen | Anmerkungen des Künstlers
	Comment     string  `json:"commentary"`                             // kommentar | Kommentare zum Bild, nicht für die Öffentlichkeit gedacht

	// Schaffensphasen können zeitlich („die Werke zwischen 1918 und 1933“), räumlich („die Pariser Arbeiten“) oder thematisch („die blaue Periode“) gegliedert sein.
	// Diese Einteilungen erfolgen in der Regel nicht durch den Künstler selbst.
	// Es handelt sich dabei meist um Zuschreibungen durch andere, beispielsweise Biographen, Literaturwissenschaftler oder Kunsthistoriker, die rückblickend erfolgen.
	// Insofern können derartige Einteilungen eines Œuvres, eines Gesamtwerkes, in einzelne Abschnitte auch völlig unterschiedlich ausfallen.
	Schaffensphase string `db:"phase"       json:"phase" schema:"phase"` // phase | Natur   --- e.g. Period???

	Teile int `json:"teile"`

	// Kunststil bezeichnet einen Stil in der Kunst, d. h. die in charakteristischen Merkmalen einheitliche Gestaltung von Kunstwerken und Kulturerzeugnissen eines Zeitalters, eines Künstlers bzw. einer Künstlergruppe oder einer Kunstschule. Alternativ werden auch die Bezeichnungen Kunstrichtung oder Strömung verwandt.

}

func (d *Work) Clone() interface{} {
	var wa Work = *d
	return &wa
}

func (d *Work) AugmentSearchDoc(sd *search.Document, l search.Level) (interface{}, error) {

	doc := workIndexDoc{Document: *sd}
	// The archive's own status, DEL included. A deleted work is still a work
	// that existed, so it stays findable rather than being dropped from the
	// index: searching for it is the point of keeping it.
	doc.Artwork.Status = d.Status
	doc.Artwork.Year = d.Year
	doc.Artwork.Form = d.Form
	doc.Artwork.Genre = d.Genre
	doc.Artwork.Style = d.Style
	doc.Artwork.Medium = d.Medium
	doc.Artwork.Support = d.Support
	doc.Artwork.Height = d.Height
	doc.Artwork.Width = d.Width
	doc.Artwork.Depth = d.Depth
	return &doc, nil
}

type workIndexDoc struct {
	search.Document
	Artwork struct {
		Status  string `json:"status"` // as the old database had it: "", "DEL", …
		Year    int    `json:"year"`
		Form    string `json:"form"`     // gattung | drawing, painting, sculpture -> art form
		Genre   string `json:"genre"`    // in der bildenden Kunst ein thematisch-inhaltliches Gebiet, wie Historie, Landschaft, Porträt oder Stillleben
		Style   string `json:"style"`    // dt: Kunststil, en: art movement
		Medium  string `json:"medium"`   // technik | Oel, Aquarell, Pastell, Radierung, Buntstifte, Tinte, Siebdruck
		Support string `json:"support" ` // träger  | Leinwand, Papier, Holz,
		Height  int    `json:"height" `
		Width   int    `json:"width"`
		Depth   int    `json:"depth"`
	} `json:"artwork"`
}

// AdaptMapping declares how the artwork fields AugmentSearchDoc emits are
// indexed. Registered with search.RegisterMappingAdapter, which is the only
// mechanism the CMS still has: it builds one shared document mapping for every
// entry and hands it to each adapter in turn, so an adapter adds its own
// sub-document and leaves the rest alone.
//
// What is declared here is exactly what workIndexDoc carries and no more. The
// mapping this replaced also declared name, id, title, area, remark, commentary,
// phase and teile, none of which AugmentSearchDoc has ever filled in — they
// indexed nothing. Commentary in particular must stay out: the field it comes
// from is marked as not for the public.
func AdaptMapping(entrymap *mapping.DocumentMapping) {

	keyword := bleve.NewKeywordFieldMapping()
	integer := bleve.NewNumericFieldMapping()

	artwork := bleve.NewDocumentMapping()
	artwork.Dynamic = false

	artwork.AddFieldMappingsAt("status", keyword)
	artwork.AddFieldMappingsAt("year", integer)
	artwork.AddFieldMappingsAt("form", keyword)
	artwork.AddFieldMappingsAt("genre", keyword)
	artwork.AddFieldMappingsAt("style", keyword)
	artwork.AddFieldMappingsAt("medium", keyword)
	artwork.AddFieldMappingsAt("support", keyword)
	artwork.AddFieldMappingsAt("height", integer)
	artwork.AddFieldMappingsAt("width", integer)
	artwork.AddFieldMappingsAt("depth", integer)

	entrymap.AddSubDocumentMapping("artwork", artwork)
}

type Ausstellung struct { // Exhibition
	content.EntryContent `type:"Ausstellung"  json:"-" mimetype:"application/json"`
	ID                   int        `json:"id"`
	Code                 string     `json:"code"`
	Ort                  string     `json:"ort"`
	Jahr                 int        `json:"jahr"`
	Venue                string     `json:"venue"`
	Titel                string     `json:"titel"`
	Untertitel           string     `json:"untertitel"`
	Typ                  string     `json:"typ"`
	Von                  *time.Time `json:"von"`
	Bis                  *time.Time `json:"bis"`
	Kommentar            string     `json:"kommentar"`
	Fotos                []string   `json:"fotos"`
}

func (d *Ausstellung) Clone() interface{} {
	var wa Ausstellung = *d
	return &wa
}

func (d *Ausstellung) AugmentSearchDoc(doc *search.Document, lev search.Level) (interface{}, error) {
	return struct {
		search.Document
		Ausstellung `json:"exhibition"`
	}{Document: *doc, Ausstellung: *d}, nil
}

// AdaptMappingAusstellung declares how an exhibition is indexed. The whole
// struct goes under `exhibition`, which is what AugmentSearchDoc embeds.
//
// Two declarations are corrected rather than carried over: kommentar and fotos
// were declared numeric, though one is a string and the other a list of them,
// so neither was ever searchable; and the mapping named "title" where the field
// is spelled "titel", so the exhibition's own title indexed nothing.
func AdaptMappingAusstellung(entrymap *mapping.DocumentMapping) {

	text := bleve.NewTextFieldMapping()
	text.Analyzer = de.AnalyzerName
	keyword := bleve.NewKeywordFieldMapping()
	integer := bleve.NewNumericFieldMapping()
	datetime := bleve.NewDateTimeFieldMapping()

	exhibition := bleve.NewDocumentMapping()
	exhibition.Dynamic = false

	exhibition.AddFieldMappingsAt("id", integer)
	exhibition.AddFieldMappingsAt("code", keyword)
	exhibition.AddFieldMappingsAt("ort", keyword)
	exhibition.AddFieldMappingsAt("jahr", integer)
	exhibition.AddFieldMappingsAt("venue", text)
	exhibition.AddFieldMappingsAt("titel", text)
	exhibition.AddFieldMappingsAt("untertitel", text)
	exhibition.AddFieldMappingsAt("typ", keyword)
	exhibition.AddFieldMappingsAt("von", datetime)
	exhibition.AddFieldMappingsAt("bis", datetime)
	exhibition.AddFieldMappingsAt("kommentar", text)
	exhibition.AddFieldMappingsAt("fotos", keyword)

	entrymap.AddSubDocumentMapping("exhibition", exhibition)
}
