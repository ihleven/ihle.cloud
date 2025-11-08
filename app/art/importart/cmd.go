package importart

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/ihleven/ihlvn/app/art"
	"github.com/ihleven/ihlvn/app/art/importart/db"

	"bitbucket.org/hotelplan/webcc-content/cms/content"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ImportCmd struct {
	DBToImportConnString string `arg:"--db" default:"postgresql://localhost/wi20241020"`
	Bilder               bool   `arg:"--bilder"`
	Ausstellungen        bool   `arg:"--ausstellungen"`
}

func (cmd *ImportCmd) Run() error {
	fmt.Println("import", cmd)
	// cms, err := art.NewCMS(flags.Internal, flags.Repository)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// account := auth.Account{}

	dbpool, err := pgxpool.New(context.Background(), cmd.DBToImportConnString)
	if err != nil {
		return err
	}
	defer dbpool.Close()

	DB := db.New(dbpool)

	if cmd.Ausstellungen {
		var ausstellungEntries []content.Entry
		ausstellungen, err := DB.GetAusstellungen()
		if err != nil {
			return err
		}
		for i, a := range ausstellungen {
			entry := content.Entry{
				Meta: content.Meta{
					Path:     "ausstellungen/" + a.Slug + ".json",
					Repo:     "content",
					ID:       strconv.Itoa(a.ID),
					Slug:     a.Slug,
					FullSlug: "ausstellungen/" + a.Slug,
					Type:     "Ausstellung",
					MIME:     "application/json",
				},
				Content: &art.Ausstellung{
					ID:         a.ID,
					Code:       a.Slug,
					Ort:        a.Ort,
					Jahr:       a.Jahr,
					Venue:      a.Venue,
					Titel:      a.Titel,
					Untertitel: a.Untertitel,
					Typ:        a.Typ,
					Von:        a.Von,
					Bis:        a.Bis,
					Kommentar:  a.Kommentar,
					Fotos:      a.Fotos,
				},
			}
			ausstellungEntries = append(ausstellungEntries, entry)
			// e, err := cms.CreateEntry(&entry, account, "")
			fmt.Println("ausstellung: ", i, entry)
		}

		// es, err := cms.SaveEntries(true, account, "initial import", ausstellungEntries...)
		// fmt.Println(es, err)
	}

	if cmd.Bilder {
		bilder, err := DB.LoadBilder(true)
		if err != nil {
			return err
		}

		var was []art.Work
		var entries []content.Entry
		for _, b := range bilder {
			wa := art.Work{
				ID:          b.ID,
				Dir:         b.Directory,
				Foto:        b.Foto,
				Serie:       fmt.Sprintf("%v", b.SerieID),
				SerieNummer: b.SerieNr,
				Status:      b.Status,

				Title:          b.Titel,
				Description:    "",
				Year:           b.Jahr,
				Form:           b.Gattung,
				Genre:          "",
				Style:          "",
				Medium:         b.Technik,
				Support:        b.Träger,
				Height:         b.Höhe,
				Width:          b.Breite,
				Depth:          b.Tiefe,
				Area:           float64(b.Breite*b.Höhe) / 10000,
				Remark:         b.Anmerkungen,
				Comment:        b.Kommentar,
				Schaffensphase: b.Phase,
				Teile:          b.Teile,
			}
			entry := content.Entry{
				Meta: content.Meta{
					Path:       fmt.Sprintf("artworks/%d.json", b.ID),
					Locale:     "de",
					Version:    string(content.DRAFT),
					Status:     content.ACTIVE,
					Slug:       strings.TrimPrefix(b.Directory, "bilder/"),
					FullSlug:   b.Directory,
					Access:     content.Access{},
					ID:         strconv.Itoa(b.ID),
					Collection: "artworks", Space: "wi",
					Repo: "content",
					Name: "Bild " + b.Titel, Notes: b.Kommentar, Tags: nil,
					MIME: "application/json",
					Type: "Work",
				},
				Content: &wa,
			}
			if b.Status == "DEL" {
				entry.Status = content.INACTIVE
			}
			// e, err := cms.CreateEntry(&entry, account, "")
			// fmt.Println("entry:", e, err)
			was = append(was, wa)
			entries = append(entries, entry)

		}
		// es, err := cms.SaveEntries(true, account, "initial import", entries...)
		// fmt.Println(es, err)
	}
	return nil
}
