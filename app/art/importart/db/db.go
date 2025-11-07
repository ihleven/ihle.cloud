package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	// _ "github.com/lib/pq" // PostgreSQL driver
)

func New(dbpool *pgxpool.Pool) *kunstDB {
	return &kunstDB{ctx: context.Background(), pool: dbpool}
}

type kunstDB struct {
	ctx  context.Context
	pool *pgxpool.Pool
}

type KunstDB interface {
	// SaveTask(title, description string) error
	GetAusstellungen() ([]Ausstellung, error)
	LoadBilder(where map[string]interface{}, serienbilder bool, deleted bool, orderBy string) ([]Bild2, error)
}

func (r *kunstDB) Select(dst interface{}, query string, args ...interface{}) error {

	err := pgxscan.Select(r.ctx, r.pool, dst, query, args...)
	if err != nil {
		// if errors.As(err, &pgx.ErrNoRows) {
		// 	return errors.NewWithCode(errors.NotFound, "Not found: %s", query)
		// }
		return errors.Wrap(err, "Error")
	}
	return nil
}

// type SQLTaskStore struct {
// 	DB *sql.DB
// }

// func (s *SQLTaskStore) InsertAusstellung(a Ausstellung) error {
// 	_, err := s.DB.Exec("INSERT INTO tasks (title, description) VALUES ($1, $2)", title, description)
// 	return err
// }

type Status string

const (
	DEL Status = "DEL"
)

type Schaffensphase string

// NATUR                         FIGUR
// ABSTR AKTION                  ABSRKT
// DESUB GENS de-substantiation  ENTGEGN
// MONO CH ROM                   MONOCHRM
const (
	// Früh                 Schaffensphase = "Frühwerk"
	FIGUR    Schaffensphase = "I"
	ABSRKT   Schaffensphase = "A"
	ENTGEGN  Schaffensphase = "E"
	MONOCHRM Schaffensphase = "M"
)

func (p Schaffensphase) IsValid() (Schaffensphase, bool) {
	switch p {
	case "F", "I", "FIGUR":
		return FIGUR, true
	case "A", "II", "AB", "ABSTRAKTION":
		return ABSRKT, true
	case "E", "ENT", "ENTGEGN":
		return ENTGEGN, true
	case "M", "MONO", "MONOCHROM":
		return MONOCHRM, true
	}
	return "", false
}

type Katalog struct {
	ID         int `db:"id"           json:"id"`
	Code       string
	Titel      string
	Untertitel string
	Jahr       int
	Datum      *time.Time
	Kommentar  string `db:"kommentar"    json:"kommentar"`

	Modified *time.Time `db:"modified" json:"modified"`

	// IndexFoto *hidrive.Meta  `db:"-"        json:"foto"`
	// Fotos     []hidrive.Meta `db:"-"        json:"fotos"`
	Dokumente interface{} `db:"-"        json:"dokumente"`
}

type Serie struct {
	ID             int     `db:"id"         json:"id"`
	Slug           string  `db:"slug"       json:"slug,omitempty"  `
	Jahr           int     `db:"jahr"       json:"jahr,omitempty"  `
	Titel          string  `db:"titel"      json:"titel"           `
	Untertitel     string  `db:"untertitel" json:"untertitel"      `
	Anzahl         int     `db:"anzahl"    json:"anzahl,omitempty"`
	JahrBis        int     `db:"jahrbis"    json:"jahrbis,omitempty"`
	Technik        string  `db:"technik"    json:"technik"         `                 // Oel, Aquarell, Pastell, Radierung, Buntstifte, Tinte, Siebdruck
	Träger         string  `db:"traeger" schema:"traeger"   json:"traeger"         ` // Leinwand, Papier, Holz,
	Höhe           int     `db:"hoehe"      schema:"hoehe" json:"hoehe"           `  //
	Breite         int     `db:"breite"     json:"breite"`                           //
	Tiefe          int     `db:"tiefe"      json:"tiefe"`                            //
	Anmerkungen    string  `db:"anmerkungen" json:"anmerkungen"`                     // Anmerkungen des Künstlers
	Kommentar      string  `db:"kommentar"  json:"kommentar"       `
	Schaffensphase string  `db:"phase"       json:"phase" schema:"phase"` // Natur
	Bilder         []Bild2 `db:"-"          json:"bilder"           schema:"-"`
	Fotos          []Foto  `db:"-"          json:"fotos"            schema:"-"`
}

type Bild2 struct {
	ID          int     `db:"id"          json:"id"` //
	Directory   string  `db:"dir"        json:"dir"`
	Jahr        int     `db:"jahr"        json:"jahr"`  //
	Titel       string  `db:"titel"       json:"titel"` //
	Gattung     string  `     json:"gattung"`
	SerieID     *int    `db:"serie_id"       json:"serie_id"`
	Serie       *Serie  `db:"-"       json:"serie"`
	SerieNr     int     `db:"serie_nr"    json:"serie_nr"`
	Technik     string  `db:"technik"     json:"technik"`                    // Oel, Aquarell, Pastell, Radierung, Buntstifte, Tinte, Siebdruck
	Bildträger  string  `db:"traeger"     json:"traeger" schema:"traeger"`   // Leinwand, Papier, Holz,
	Höhe        int     `db:"hoehe"       json:"hoehe"       schema:"hoehe"` //
	Breite      int     `db:"breite"      json:"breite"`                     //
	Tiefe       int     `db:"tiefe"       json:"tiefe"`                      //
	Fläche      float64 `db:"flaeche"     json:"flaeche"`                    // Bildfläche in qm
	Anmerkungen string  `db:"anmerkungen" json:"anmerkungen"`                // Anmerkungen des Künstlers
	Kommentar   string  `db:"kommentar"   json:"kommentar"`                  // Kommentare zum Bild, nicht für die Öffentlichkeit gedacht
	// Überordnung    string  `db:"ordnung"     json:"ueberordnung"`               //
	Schaffensphase string `db:"phase"       json:"phase" schema:"phase"` // Natur
	Fotos          []Foto `db:"-"           json:"fotos" `
	IndexFotoID    int    `db:"foto_id"     json:"foto_id" schema:"foto_id"` //
	IndexFoto      *Foto  `db:"foto"     json:"foto" schema:"-"`             //
	// Systematik     string `db:"-"    json:"sytematik"`    //
	// Ordnung        string `db:"-"    json:"ordnung"`      //
	// Hauptfoto   *Foto
	// Diptychon / Triptychon
	// AusstellungID int
	// KatalogID     int
	Teile    int        `db:"teile"       json:"teile"`
	Modified *time.Time `db:"modified"    json:"modified"`
	// SeriePtr *Serie     `db:"-"       json:"Serie"`
	Status Status `db:"status"       json:"status"`
}

type Foto struct {
	ID        int       `db:"id"        json:"id"`
	BildID    int       `db:"bild_id"   json:"-"`
	Bild      *Bild2    `db:"-"         json:"-"`
	Index     int       `db:"index"     json:"nr"`
	Name      string    `db:"name"      json:"name"`
	Size      int       `db:"size"      json:"size"`
	Uploaded  time.Time `db:"uploaded"  json:"uploaded"`
	Path      string    `db:"path"      json:"path"`
	Format    string    `db:"format"    json:"format"`
	Width     int       `db:"width"     json:"width"`
	Height    int       `db:"height"    json:"height"`
	Taken     time.Time `db:"taken"     json:"taken"`
	Caption   string    `db:"caption"   json:"caption"`
	Kommentar string    `db:"kommentar" json:"kommentar"`
	Labels    []string  `db:"-"         json:"labels"`

	// AusstellungID *int `db:"aust_id"         json:"aust_id"`
	// KatalogID     int

	Serie  *string
	Status Status `db:"status"       json:"status"`
}

func (r *kunstDB) LoadBilderD(where map[string]interface{}, serienbilder bool, deleted bool, orderBy string) ([]Bild2, error) {
	var fields []string
	var values []interface{}

	stmt := "SELECT * FROM bild"
	if serienbilder {
		stmt += " WHERE (serie_id is null OR serie_id is not null)"
	} else {
		stmt += " WHERE serie_id is null "
	}
	if !deleted {
		stmt += " AND status!='DEL' "
	}

	i := 1
	for k, v := range where {
		fields = append(fields, fmt.Sprintf(" %s=$%d ", k, i))
		values = append(values, v)
		i++
	}

	if len(fields) > 0 {
		stmt += " AND "
		stmt += strings.Join(fields, "AND")
	}

	switch orderBy {
	case "updated":
		stmt += " ORDER BY modified DESC"
	case "jahr":
		stmt += " ORDER BY jahr DESC, id DESC"
	case "name":
		stmt += " ORDER BY titel ASC"
	default:
		stmt += " ORDER BY id DESC"
	}

	var bilder []Bild2
	err := r.Select(&bilder, stmt, values...)
	if err != nil {
		return bilder, err
	}
	indexByID := make(map[int]int)
	for i, bild := range bilder {
		indexByID[bild.ID] = i
	}
	var fotos []Foto
	err = r.Select(&fotos, "SELECT * FROM foto WHERE id IN (SELECT foto_id FROM bild)")
	if err != nil {
		fmt.Println("error in foto", err)
		return bilder, err
	}

	for i, foto := range fotos {
		if index, ok := indexByID[foto.BildID]; ok {
			bilder[index].IndexFoto = &fotos[i]
		}
	}

	var serien []Serie
	err = r.Select(&serien, "SELECT * FROM serie WHERE id IN (SELECT serie_id FROM bild)")
	if err != nil {
		return bilder, err
	}

	serienByID := make(map[int]*Serie)
	for i, serie := range serien {
		serienByID[serie.ID] = &serien[i]
	}

	for i, bild := range bilder {
		if bild.SerieID != nil && serienByID[*bild.SerieID] != nil {
			bilder[i].Serie = serienByID[*bild.SerieID]
		}
	}

	return bilder, nil
}
