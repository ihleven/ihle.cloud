package hidrive

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ihleven/ihlvn/pkg/hiauth"
)

// Whether the app can still reach the storage provider.
//
// A refresh token is the only credential the app cannot obtain for itself: when
// HiDrive rejects one, storage access stops until a person redoes the consent
// flow. Asking is worth doing on demand, because otherwise the answer only
// appears in the server log at startup — and by then a film has already failed
// to play for somebody.

// TokenStatus is what is known about one stored alias.
type TokenStatus struct {
	Alias     string
	Scope     string
	ExpiresAt string

	// Length and Stored describe the value as it sits in the database. A token
	// that round-trips through an editor or a paste can arrive with whitespace,
	// which HiDrive rejects exactly as it rejects a dead one — so the stored
	// shape is worth seeing next to the verdict.
	Length int
	Stored string

	// Checked says whether the provider was asked at all; Accepted and Detail
	// are its answer. Permanent distinguishes "this will never work again" from
	// "try later", because only the first needs a person.
	Checked   bool
	Accepted  bool
	Permanent bool
	Detail    string
}

// How a stored value looks before anyone asks the provider about it.
const (
	StoredClean      = "clean"
	StoredEmpty      = "EMPTY"
	StoredWhitespace = "HAS WHITESPACE"
)

// InspectTokens reports what is stored and, when check is set, whether HiDrive
// still accepts each one — which costs one request per alias.
//
// It returns facts. How to lay them out, and what to tell whoever is reading,
// belongs to the command that asked.
func InspectTokens(ctx context.Context, store *Store, clientID, clientSecret string, check bool) ([]TokenStatus, error) {
	rows, err := store.LoadTokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading hitoken: %w", err)
	}

	var checker *hiauth.TokenMngr
	if check {
		checker = hiauth.NewTokenChecker(clientID, clientSecret)
	}

	statuses := make([]TokenStatus, 0, len(rows))
	for _, row := range rows {
		refresh := row["refresh_token"]

		status := TokenStatus{
			Alias:     row["alias"],
			Scope:     row["scope"],
			ExpiresAt: row["expires_at"],
			Length:    len(refresh),
			Stored:    storedShape(refresh),
		}
		if checker != nil {
			status.Checked = true
			status.Accepted, status.Permanent, status.Detail = ask(checker, refresh)
		}
		statuses = append(statuses, status)
	}
	return statuses, nil
}

func storedShape(refresh string) string {
	switch {
	case refresh == "":
		return StoredEmpty
	case strings.TrimSpace(refresh) != refresh:
		return StoredWhitespace
	}
	return StoredClean
}

// ask makes the same exchange the server makes, so a token that passes here is
// one the server can use.
func ask(checker *hiauth.TokenMngr, refresh string) (accepted, permanent bool, detail string) {
	token, err := checker.RefreshToken(refresh)
	if err == nil {
		return true, false, fmt.Sprintf("access token for %ds", token.ExpiresIn)
	}

	var oauthErr *hiauth.OAuthError
	if errors.As(err, &oauthErr) {
		return false, oauthErr.Permanent(), oauthErr.Desc
	}
	return false, false, err.Error()
}
