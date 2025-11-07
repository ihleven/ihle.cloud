package db

import (
	"context"
	"time"
)

type Ausstellung struct { // Exhibition

	ID         int        `db:"id"           json:"id"`
	Slug       string     `db:"slug"         json:"slug"`
	Ort        string     `db:"ort"          json:"ort"`
	Jahr       int        `db:"jahr"         json:"jahr"`
	Venue      string     `db:"venue"        json:"venue"`
	Titel      string     `db:"titel"        json:"titel"`
	Untertitel string     `db:"untertitel"   json:"untertitel"`
	Typ        string     `db:"typ"          json:"typ"` // einzel, sammel, dauerleihgabe, ....
	Von        *time.Time `db:"von"          json:"von"`
	Bis        *time.Time `db:"bis"          json:"bis"`
	Kommentar  string     `db:"kommentar"    json:"kommentar"`
	// NumBilder int        `db:"num_bilder"          json:"num_bilder"  schema:"num_bilder"`
	Bilder    []Bild2     `db:"-"    json:"bilder"`
	FotosStr  string      `db:"fotos"    json:"-"`
	Fotos     []string    `db:"-	"    json:"fotos"`
	Dokumente interface{} `db:"-"    json:"dokumente"`
}

func (db *kunstDB) GetAusstellungen() ([]Ausstellung, error) {

	rows, err := db.pool.Query(
		context.Background(),
		"SELECT id,slug,titel,untertitel,typ,jahr,von,bis,ort,venue,kommentar,fotos FROM ausstellung ORDER BY jahr, von, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ausstellungen []Ausstellung

	for rows.Next() {

		var a Ausstellung

		err := rows.Scan(&a.ID, &a.Slug, &a.Titel, &a.Untertitel, &a.Typ, &a.Jahr, &a.Von, &a.Bis, &a.Ort, &a.Venue, &a.Kommentar, &a.FotosStr)
		if err != nil {
			return nil, err
		}

		ausstellungen = append(ausstellungen, a)
	}

	return ausstellungen, nil
}
