package db

import (
	"time"
)

type Bild struct {
	ID          int       `db:"id"`
	Directory   string    `db:"dir"`
	Jahr        int       `db:"jahr"`
	Phase       string    `db:"phase"`
	Titel       string    `db:"titel"`
	Gattung     string    `db:"gattung"`  //char(1) NOT NULL DEFAULT 'M',
	Foto        int       `db:"foto_id"`  // integer NOT NULL DEFAULT 0,
	SerieID     *int      `db:"serie_id"` //serie_id    integer REFERENCES serie(id) ON UPDATE CASCADE ON DELETE RESTRICT,
	SerieNr     int       `db:"serie_nr"` //serie_nr    integer NOT NULL DEFAULT 0,
	Teile       int       `db:"teile"`
	Technik     string    `db:"technik"` // Oel, Aquarell, Pastell, Radierung, Buntstifte, Tinte, Siebdruck
	Träger      string    `db:"traeger"` // Leinwand, Papier, Holz,
	Höhe        int       `db:"hoehe"`
	Breite      int       `db:"breite"`
	Tiefe       int       `db:"tiefe"`
	Fläche      float64   `db:"flaeche"`
	Anmerkungen string    `db:"anmerkungen"`
	Kommentar   string    `db:"kommentar"` // Kommentare zum Bild, nicht für die Öffentlichkeit gedacht
	Modified    time.Time `db:"modified"`
	Status      string    `db:"status"`
}

func (r *kunstDB) LoadBilder(deleted bool) ([]Bild, error) {

	stmt := "SELECT * FROM bild"
	if !deleted {
		stmt += " WHERE status!='DEL' "
	}
	stmt += " ORDER BY id DESC"

	var bilder []Bild
	err := r.Select(&bilder, stmt)
	if err != nil {
		return nil, err
	}

	return bilder, nil
}
