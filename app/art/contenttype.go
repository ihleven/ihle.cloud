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

func (d *workIndexDoc) BleveType() string {
	return "Work"
}
func (d Work) BleveMapping() *mapping.DocumentMapping {

	text := bleve.NewTextFieldMapping()
	text.Analyzer = de.AnalyzerName
	keyword := bleve.NewKeywordFieldMapping()
	integer := bleve.NewNumericFieldMapping()

	mapping := search.CreateEntryDocMapping()
	// mapping := bleve.NewDocumentMapping()
	mapping.Dynamic = false

	artwork := bleve.NewDocumentMapping()
	artwork.Dynamic = false
	artwork.AddFieldMappingsAt("name", text)
	artwork.AddFieldMappingsAt("id", keyword)
	artwork.AddFieldMappingsAt("status", keyword)
	artwork.AddFieldMappingsAt("title", text)
	artwork.AddFieldMappingsAt("year", integer)
	artwork.AddFieldMappingsAt("form", keyword)
	artwork.AddFieldMappingsAt("genre", keyword)
	artwork.AddFieldMappingsAt("style", keyword)
	artwork.AddFieldMappingsAt("medium", keyword)
	artwork.AddFieldMappingsAt("support", keyword)
	artwork.AddFieldMappingsAt("height", integer)
	artwork.AddFieldMappingsAt("width", integer)
	artwork.AddFieldMappingsAt("depth", integer)
	artwork.AddFieldMappingsAt("area", integer)
	artwork.AddFieldMappingsAt("remark", text)
	artwork.AddFieldMappingsAt("commentary", text)
	artwork.AddFieldMappingsAt("phase", keyword)
	artwork.AddFieldMappingsAt("teile", integer)

	mapping.AddSubDocumentMapping("artwork", artwork)

	return mapping
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

func (d Ausstellung) BleveType() string {
	return "exhibition"
}

func (d *Ausstellung) AugmentSearchDoc(doc *search.Document, lev search.Level) (interface{}, error) {
	return struct {
		search.Document
		// BleveType   string `json:"type"`
		Ausstellung `json:"exhibition"`
	}{Document: *doc, Ausstellung: *d}, nil
}

func (d Ausstellung) BleveMapping() *mapping.DocumentMapping {

	text := bleve.NewTextFieldMapping()
	text.Analyzer = de.AnalyzerName
	keyword := bleve.NewKeywordFieldMapping()
	integer := bleve.NewNumericFieldMapping()
	datetime := bleve.NewDateTimeFieldMapping()

	emapping := search.CreateEntryDocMapping()
	emapping.Dynamic = false

	mapping := bleve.NewDocumentMapping()
	mapping.Dynamic = false

	// mapping.AddFieldMappingsAt("name", textFieldMapping)
	// mapping.AddFieldMappingsAt("id", keywordFieldMapping)
	mapping.AddFieldMappingsAt("code", keyword)
	mapping.AddFieldMappingsAt("ort", keyword)
	mapping.AddFieldMappingsAt("jahr", integer)
	mapping.AddFieldMappingsAt("venue", text)
	mapping.AddFieldMappingsAt("title", text)
	mapping.AddFieldMappingsAt("untertitel", keyword)
	mapping.AddFieldMappingsAt("typ", keyword)
	mapping.AddFieldMappingsAt("von", datetime)
	mapping.AddFieldMappingsAt("bis", datetime)
	mapping.AddFieldMappingsAt("kommentar", integer)
	mapping.AddFieldMappingsAt("fotos", integer)

	emapping.AddSubDocumentMapping("exhibition", mapping)

	return emapping
}
