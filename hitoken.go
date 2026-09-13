package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ihleven/ihlvn/app/db"
	"github.com/ihleven/ihlvn/app/hidrive"
)

// Reporting on the storage connection.
//
// What is stored, and whether the provider still accepts it, is app/hidrive's to
// answer. What is here is a table and a sentence about what to do next.

type HiTokenCmd struct {
	Check bool `arg:"--check" help:"ask HiDrive whether each stored token still works (one request per alias)"`
}

func (c *HiTokenCmd) Run(flags Flags) error {
	// The check talks to HiDrive once per alias, so this is generous rather
	// than instant.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pg, err := db.New(flags.DbConn)
	if err != nil {
		return fmt.Errorf("connecting to %s: %w", flags.DbConn, err)
	}
	defer pg.Close()

	// As the account commands do: a command can never talk to a schema the
	// binary does not expect, and on a fresh database the alternative is a bare
	// "relation hitoken does not exist".
	if err := db.Migrate(ctx, pg.Pool()); err != nil {
		return err
	}

	statuses, err := hidrive.InspectTokens(ctx, hidrive.NewStore(pg.Pool()),
		flags.HidriveClientID, flags.HidriveClientSecret, c.Check)
	if err != nil {
		return err
	}
	if len(statuses) == 0 {
		return errors.New("no rows in hitoken: authorise an alias at /apihle/auth/authorize first")
	}
	return reportHiTokens(statuses, flags.PublicURL, os.Stdout)
}

func reportHiTokens(statuses []hidrive.TokenStatus, publicURL string, out io.Writer) error {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ALIAS\tSCOPE\tLENGTH\tSTORED\tEXPIRES_AT\tSTATUS")

	var failed int
	for _, s := range statuses {
		if s.Checked && !s.Accepted {
			failed++
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n",
			s.Alias, s.Scope, s.Length, s.Stored, s.ExpiresAt, verdict(s))
	}
	if err := w.Flush(); err != nil {
		return err
	}

	if failed > 0 {
		fmt.Fprintf(out, "\n%d of %d rejected. Re-authorise at %s/apihle/auth/authorize as the same HiDrive\naccount, then restart the server — refreshers read this table only at startup.\n",
			failed, len(statuses), strings.TrimSuffix(publicURL, "/"))
	}
	return nil
}

// verdict puts one status into a column. Permanent and retryable read
// differently on purpose: only the first is a job for a person.
func verdict(s hidrive.TokenStatus) string {
	switch {
	case !s.Checked:
		return "not checked"
	case s.Accepted:
		return "ok (" + s.Detail + ")"
	case s.Permanent:
		return "REJECTED: " + s.Detail
	default:
		return "error (retryable): " + s.Detail
	}
}
