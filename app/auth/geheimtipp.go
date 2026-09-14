package auth

// Signing in a geheimtipp player who has no account here yet.
//
// Everything in this file is migration scaffolding and is meant to be deleted.
// When geheimtipp_migration is empty every player has an account, the fallback
// in authenticatePassword can be switched off, and this file and the table go
// together.

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// Account kinds. See the comment on account.type in 0003_geheimtipp.sql: this
// is what an account may do, not where it came from.
const (
	TypeFull       = "full"
	TypeGeheimtipp = "geheimtipp"
)

// Confined reports whether the account may reach nothing but the pool.
//
// It is asked at one place — requireAccount — so that a route added later is
// refused by default rather than reachable until someone remembers it.
func (a *Account) Confined() bool { return a.Type == TypeGeheimtipp }

// GeheimtippPerson is what creating an account for a pool player needs.
//
// It deliberately has no password field. The comparison happens inside
// VerifyGeheimtippPassword, so the pool's plaintext never reaches a struct that
// something could log or marshal.
type GeheimtippPerson struct {
	Login    string
	Vorname  string
	Nachname string
	Email    string
}

// DisplayName is the two name parts as one string, falling back to the login.
//
// The parts are stored separately because that is how the pool holds them, and
// joining is lossless where splitting is not. The fallback matters because
// display_name is the name a commit is authored with, and a blank one is worse
// than a terse one.
func (p *GeheimtippPerson) DisplayName() string {
	if name := strings.TrimSpace(p.Vorname + " " + p.Nachname); name != "" {
		return name
	}
	return p.Login
}

// VerifyGeheimtippPassword reports the player behind a login and password.
//
// The password is compared here rather than returned for comparison, so that
// the only place the pool's plaintext exists outside the database is one
// local variable in this function.
func (s *Store) VerifyGeheimtippPassword(ctx context.Context, login, plain string) (*GeheimtippPerson, error) {
	var p GeheimtippPerson
	var passwd string
	err := s.pool.QueryRow(ctx, `
		select login, passwd, vorname, nachname, email
		from geheimtipp_migration where login = $1`, login,
	).Scan(&p.Login, &passwd, &p.Vorname, &p.Nachname, &p.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoAccount
	}
	if err != nil {
		return nil, fmt.Errorf("auth: loading geheimtipp migration row: %w", err)
	}

	// Constant time because the length of the match would otherwise be
	// observable, and these are passwords people have reused for years.
	if subtle.ConstantTimeCompare([]byte(passwd), []byte(plain)) != 1 {
		return nil, ErrNoAccount
	}
	return &p, nil
}

// AdoptGeheimtipper creates a confined account for a pool player and consumes
// their migration row.
//
// One transaction, because a created account whose migration row survived would
// let the fallback run again, and a deleted row whose account was not created
// would lock the person out with no way back.
//
// The email is checked case-insensitively although the unique constraint is not:
// an address that differs only in case is the same person, and adopting them
// would mean a second account for someone who already has one. Adoption refuses
// rather than claiming it, because the pool does not verify addresses and
// claiming on one would be a way into an account.
func (s *Store) AdoptGeheimtipper(ctx context.Context, p *GeheimtippPerson, passwordHash string) (*Account, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("auth: adopting %s: %w", p.Login, err)
	}
	defer tx.Rollback(context.WithoutCancel(ctx))

	var taken bool
	if err := tx.QueryRow(ctx,
		`select exists (select 1 from account where lower(email) = lower($1))`,
		p.Email).Scan(&taken); err != nil {
		return nil, fmt.Errorf("auth: adopting %s: %w", p.Login, err)
	}
	if taken {
		return nil, ErrExists
	}

	handle := make([]byte, 32)
	if _, err := rand.Read(handle); err != nil {
		return nil, fmt.Errorf("auth: generating an account handle: %w", err)
	}

	displayName := p.DisplayName()
	var id int64
	var createdAt time.Time
	err = tx.QueryRow(ctx, `
		insert into account (name, display_name, email, handle, password_hash, type)
		values ($1, $2, $3, $4, $5, $6)
		returning id, created_at`,
		p.Login, displayName, p.Email, handle, passwordHash, TypeGeheimtipp,
	).Scan(&id, &createdAt)
	if isUniqueViolation(err) {
		return nil, ErrExists
	}
	if err != nil {
		return nil, fmt.Errorf("auth: adopting %s: %w", p.Login, err)
	}

	if _, err := tx.Exec(ctx, `delete from geheimtipp_migration where login = $1`, p.Login); err != nil {
		return nil, fmt.Errorf("auth: adopting %s: %w", p.Login, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("auth: adopting %s: %w", p.Login, err)
	}

	return &Account{
		ID: id, Name: p.Login, DisplayName: displayName, Email: p.Email,
		Handle: handle, Type: TypeGeheimtipp, CreatedAt: createdAt,
		passwordHash: passwordHash,
	}, nil
}

// PendingGeheimtipper counts the players who have not signed in yet.
//
// They are the people who would be locked out the moment the fallback is
// switched off, so the count is what says whether it is safe to switch it off.
func (s *Store) PendingGeheimtipper(ctx context.Context) (int, error) {
	var n int
	if err := s.pool.QueryRow(ctx, `select count(*) from geheimtipp_migration`).Scan(&n); err != nil {
		return 0, fmt.Errorf("auth: counting pending geheimtipper: %w", err)
	}
	return n, nil
}
