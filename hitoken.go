package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/ihleven/ihlvn/app/db"
	"github.com/ihleven/ihlvn/pkg/hiauth"
)

// HiDrive token inspection.
//
// A refresh token is the only credential the app cannot obtain for itself: when
// HiDrive rejects one, storage access stops until a person redoes the consent
// flow. This reports what is stored and, on request, whether HiDrive still
// accepts it — which the server's log otherwise only reveals at startup.

type HiTokenCmd struct {
	Check bool `arg:"--check" help:"ask HiDrive whether each stored token still works (one request per alias)"`
}

func (c *HiTokenCmd) Run(flags Flags) error {
	pg, err := db.New(flags.DbConn)
	if err != nil {
		return fmt.Errorf("connecting to %s: %w", flags.DbConn, err)
	}
	defer pg.Close()

	tokens, err := pg.LoadTokens()
	if err != nil {
		return fmt.Errorf("reading hitoken: %w", err)
	}
	if len(tokens) == 0 {
		return errors.New("no rows in hitoken: authorise an alias at /apihle/auth/authorize first")
	}

	return reportHiTokens(tokens, c.Check, flags, os.Stdout)
}

func reportHiTokens(tokens []map[string]string, check bool, flags Flags, out io.Writer) error {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ALIAS\tSCOPE\tLENGTH\tSTORED\tEXPIRES_AT\tSTATUS")

	var failed int

	for _, t := range tokens {
		refresh := t["refresh_token"]

		status := "not checked"
		if check {
			status = checkHiToken(flags, refresh)
			if !strings.HasPrefix(status, "ok") {
				failed++
			}
		}

		// A token that round-trips through an editor or a paste can arrive with
		// whitespace, which HiDrive rejects exactly as it rejects a dead one —
		// so the stored shape is worth seeing next to the verdict.
		stored := "clean"
		switch {
		case refresh == "":
			stored = "EMPTY"
		case strings.TrimSpace(refresh) != refresh:
			stored = "HAS WHITESPACE"
		}

		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n",
			t["alias"], t["scope"], len(refresh), stored, t["expires_at"], status)
	}

	if err := w.Flush(); err != nil {
		return err
	}

	if failed > 0 {
		fmt.Fprintf(out, "\n%d of %d rejected. Re-authorise at %s/apihle/auth/authorize as the same HiDrive\naccount, then restart the server — refreshers read this table only at startup.\n",
			failed, len(tokens), strings.TrimSuffix(flags.PublicURL, "/"))
	}

	return nil
}

// checkHiToken asks HiDrive whether one refresh token still works. The answer is
// the same exchange the server makes, so a token that passes here is one the
// server can use.
func checkHiToken(flags Flags, refresh string) string {
	mngr := hiauth.NewTokenChecker(flags.HidriveClientID, flags.HidriveClientSecret)

	token, err := mngr.RefreshToken(refresh)
	if err != nil {
		var oauthErr *hiauth.OAuthError
		if errors.As(err, &oauthErr) {
			if oauthErr.Permanent() {
				return "REJECTED: " + oauthErr.Desc
			}

			return "error (retryable): " + oauthErr.Desc
		}

		return "error: " + err.Error()
	}

	return fmt.Sprintf("ok (access token for %ds)", token.ExpiresIn)
}
