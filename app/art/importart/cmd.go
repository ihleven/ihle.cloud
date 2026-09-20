package importart

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ihleven/ihlvn/app/art"
	"github.com/ihleven/ihlvn/app/art/importart/db"

	"github.com/interhome-group/cms/content"
	"github.com/interhome-group/cms/mgmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ImportCmd struct {
	DBToImportConnString string `arg:"--db" default:"postgresql://localhost/wi20241020"`
	Bilder               bool   `arg:"--bilder"`
	Ausstellungen        bool   `arg:"--ausstellungen"`

	// Nothing is written without this. The archive it would write to already
	// holds the result of the run on 2024-12-10 — 308 artworks and 123
	// exhibitions — so the default is to build the entries and show them, and
	// rewriting the lot has to be asked for.
	Write bool `arg:"--write" help:"actually write the entries; otherwise print them"`

	// Whether works the old database marks DEL come across too.
	//
	// Default false, because that is what the archive holds: the database has
	// 308 live rows and 121 marked DEL, and the 2024-12-10 import wrote exactly
	// the 308. So the default reproduces what is there, and taking the deleted
	// ones is a thing you ask for. They arrive as INACTIVE entries carrying
	// their DEL status, which is indexed — see art.AdaptMapping — so they can
	// be searched for rather than merely stored.
	Deleted bool `arg:"--deleted" help:"also import works marked DEL in the source database"`

	// Who the commit is attributed to. There is no session here, so it cannot be
	// taken from one.
	As    string `arg:"--as" default:"import" help:"user the import is committed as"`
	Email string `arg:"--email" default:"import@ihleven.de" help:"email for the commit signature"`
}

// Run reads the old Kunst database and turns it into CMS entries.
//
// The manager is passed in rather than built here: it is configured from the
// same flags the server uses, and those live in package main.
//
// The import is one changeset staged and published in a single step, not an
// entry at a time. A half-finished import is worse than none: the two branches
// below are one archive, and a per-entry write would leave the repository
// holding part of it with no record of where it stopped.
func (cmd *ImportCmd) Run(cms *mgmt.Mngr) error {
	fmt.Println("import", cmd)

	// Grant-all, because an import writes across the whole archive and answers
	// to no ACL: "*" expands to every registered permission at scope-build time
	// (content.Scope.Add). The name is what the commit is authored with.
	usr := content.NewUser(cmd.As,
		content.Signature{Name: cmd.As, Email: cmd.Email},
		nil, []string{"*"})

	changeset := content.Changeset{}

	dbpool, err := pgxpool.New(context.Background(), cmd.DBToImportConnString)
	if err != nil {
		return err
	}
	defer dbpool.Close()

	DB := db.New(dbpool)

	if cmd.Ausstellungen {
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
			changeset[entry.Path] = &entry
			fmt.Println("ausstellung: ", i, entry.Path)
		}
	}

	if cmd.Bilder {
		bilder, err := DB.LoadBilder(cmd.Deleted)
		if err != nil {
			return err
		}

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
			changeset[entry.Path] = &entry
			fmt.Println("bild: ", entry.Path)
		}
	}

	if len(changeset) == 0 {
		fmt.Println("nothing selected — pass --bilder and/or --ausstellungen")
		return nil
	}

	if !cmd.Write {
		fmt.Printf("%d entries built, nothing written (pass --write)\n", len(changeset))
		return nil
	}

	if cms == nil {
		return fmt.Errorf("--write needs a content repository: set REPO_CONTENT")
	}

	// Staged and then published, which is how every other write in this app
	// reaches the branch (see app/cmsapi/handlers.go). Publishing is a separate
	// act: staging records the work on a ref of its own, and only the publish
	// moves the branch other readers see.
	now := time.Now()
	msg := fmt.Sprintf("import %d entries from %s", len(changeset), cmd.DBToImportConnString)

	draft, err := cms.StageAtHead(changeset, usr, msg, now)
	if err != nil {
		return err
	}

	_, warnings, err := cms.PublishDraft(context.Background(), usr, draft.Ref, now)
	if err != nil {
		return err
	}
	for _, w := range warnings {
		fmt.Println("warning:", w)
	}
	fmt.Printf("wrote %d entries\n", len(changeset))

	return nil
}
