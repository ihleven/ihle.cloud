package authn

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store is every database access the authentication layer makes. Nothing above
// it knows SQL.
type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

var (
	// ErrNoAccount covers an unknown or disabled account. A disabled account is
	// refused at the point of loading rather than at each call site, so it
	// cannot be used anywhere.
	ErrNoAccount = errors.New("authn: no such account")
	// ErrBadToken covers a token that is unknown, expired or already spent. The
	// three are deliberately indistinguishable to a caller.
	ErrBadToken  = errors.New("authn: invalid token")
	ErrNoSession = errors.New("authn: no such session")
	ErrExists    = errors.New("authn: already exists")
)

// Account is a person. Name is the login, and it is what the CMS sees as
// content.User.ID — so it is the string compared against an entry's owner.
// DisplayName and Email are what a git commit is signed with.
type Account struct {
	ID          int64
	Name        string
	DisplayName string
	Email       string
	Handle      []byte
	Disabled    bool
	HiDrive     HiDrive
	CMS         CMSProfile
	CreatedAt   time.Time

	// Never leaves the package: password.go verifies against it, and callers
	// change it through SetPasswordHash.
	passwordHash string
}

// HiDrive is the account's pointer into the hitoken table. Aliases are shared —
// the anonymous account uses one — so this is a reference, not ownership.
type HiDrive struct {
	Alias string `json:"alias"`
	Home  string `json:"home"`
	Root  string `json:"root"`
}

// CMSProfile is what content.User is built from. An account with no cmsauth row
// has empty slices here, which is a valid account holding no CMS rights.
type CMSProfile struct {
	Groups      []string
	Permissions []string
}

// HasPassword reports whether the account can authenticate with a password. The
// anonymous account cannot: it is a stand-in for "nobody signed in" and has
// never been a login.
func (a *Account) HasPassword() bool { return a.passwordHash != "" }

const accountColumns = `
	a.id, a.name, a.display_name, a.email, a.handle, a.disabled,
	coalesce(a.password_hash, ''), a.hidrive, a.created_at,
	coalesce(c.groups, '{}'), coalesce(c.permissions, '{}')`

// CreateAccount adds an account with no credential; it cannot sign in until a
// password is set or a passkey is enrolled.
func (s *Store) CreateAccount(ctx context.Context, name, displayName, email string) (*Account, error) {
	handle := make([]byte, 32)
	if _, err := rand.Read(handle); err != nil {
		return nil, fmt.Errorf("authn: generating an account handle: %w", err)
	}

	var id int64
	var createdAt time.Time
	err := s.pool.QueryRow(ctx, `
		insert into account (name, display_name, email, handle)
		values ($1, $2, $3, $4)
		returning id, created_at`, name, displayName, email, handle).Scan(&id, &createdAt)
	if isUniqueViolation(err) {
		return nil, ErrExists
	}
	if err != nil {
		return nil, fmt.Errorf("authn: creating account %s: %w", name, err)
	}
	return &Account{
		ID: id, Name: name, DisplayName: displayName, Email: email,
		Handle: handle, CreatedAt: createdAt,
	}, nil
}

func (s *Store) AccountByName(ctx context.Context, name string) (*Account, error) {
	return s.account(ctx, `a.name = $1`, name)
}

func (s *Store) AccountByID(ctx context.Context, id int64) (*Account, error) {
	return s.account(ctx, `a.id = $1`, id)
}

// AccountByHandle resolves the opaque handle an authenticator returns during a
// usernameless login.
func (s *Store) AccountByHandle(ctx context.Context, handle []byte) (*Account, error) {
	return s.account(ctx, `a.handle = $1`, handle)
}

// DisabledAccountByName loads an account whether or not it is disabled.
//
// Every other lookup refuses a disabled account, which is what makes disabling
// take effect everywhere at once. Administration is the one caller that must see
// through that: a disabled account appears in a listing, and something has to be
// able to load it in order to turn it back on.
func (s *Store) DisabledAccountByName(ctx context.Context, name string) (*Account, error) {
	return s.load(ctx, `a.name = $1`, name)
}

func (s *Store) account(ctx context.Context, where string, arg any) (*Account, error) {
	a, err := s.load(ctx, where, arg)
	if err != nil {
		return nil, err
	}
	// A disabled account must not be usable anywhere, so it is refused here
	// rather than at each call site. Every live session dies with it.
	if a.Disabled {
		return nil, ErrNoAccount
	}
	return a, nil
}

func (s *Store) load(ctx context.Context, where string, arg any) (*Account, error) {
	row := s.pool.QueryRow(ctx, `select`+accountColumns+`
		from account a left join cmsauth c on c.account_id = a.id
		where `+where, arg)

	a, err := scanAccount(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoAccount
	}
	if err != nil {
		return nil, fmt.Errorf("authn: loading account: %w", err)
	}
	return a, nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanAccount(row scannable) (*Account, error) {
	var a Account
	var hidrive []byte
	if err := row.Scan(&a.ID, &a.Name, &a.DisplayName, &a.Email, &a.Handle, &a.Disabled,
		&a.passwordHash, &hidrive, &a.CreatedAt,
		&a.CMS.Groups, &a.CMS.Permissions); err != nil {
		return nil, err
	}
	if len(hidrive) > 0 {
		if err := json.Unmarshal(hidrive, &a.HiDrive); err != nil {
			return nil, fmt.Errorf("decoding hidrive settings for %s: %w", a.Name, err)
		}
	}
	return &a, nil
}

// ListAccounts returns every account, including disabled ones, for the admin
// commands. It is the one reader that does not hide a disabled account.
func (s *Store) ListAccounts(ctx context.Context) ([]*Account, error) {
	rows, err := s.pool.Query(ctx, `select`+accountColumns+`
		from account a left join cmsauth c on c.account_id = a.id
		order by a.name`)
	if err != nil {
		return nil, fmt.Errorf("authn: listing accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*Account
	for rows.Next() {
		a, err := scanAccount(rows)
		if err != nil {
			return nil, fmt.Errorf("authn: listing accounts: %w", err)
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

// SetDisabled blocks or unblocks an account. Callers also drop its sessions:
// loading refuses a disabled account, but dropping the rows makes that explicit.
// It returns the account's id, because a caller that has just disabled an
// account can no longer look it up — loading refuses a disabled account.
func (s *Store) SetDisabled(ctx context.Context, name string, disabled bool) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx,
		`update account set disabled = $2 where name = $1 returning id`, name, disabled).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNoAccount
	}
	if err != nil {
		return 0, fmt.Errorf("authn: setting disabled on %s: %w", name, err)
	}
	return id, nil
}

// SetIdentity changes how an account is described: the name shown and the
// address a commit is signed with.
//
// The login name is deliberately not settable. It is what entries are owned by
// (ContentUser uses it as the id), so renaming an account would silently detach
// it from everything it has written.
func (s *Store) SetIdentity(ctx context.Context, id int64, displayName, email string) error {
	tag, err := s.pool.Exec(ctx,
		`update account set display_name = $2, email = $3 where id = $1`, id, displayName, email)
	if err != nil {
		return fmt.Errorf("authn: setting identity for account %d: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNoAccount
	}
	return nil
}

// SetPasswordHash stores an already-hashed password. Hashing is password.go's
// job; the store never sees a plaintext.
func (s *Store) SetPasswordHash(ctx context.Context, id int64, hash string) error {
	tag, err := s.pool.Exec(ctx, `update account set password_hash = $2 where id = $1`, id, hash)
	if err != nil {
		return fmt.Errorf("authn: setting password for account %d: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNoAccount
	}
	return nil
}

// SetCMSProfile replaces the groups and permissions content.User is built from.
func (s *Store) SetCMSProfile(ctx context.Context, id int64, p CMSProfile) error {
	// A nil slice would be encoded as NULL, and both columns are not-null: an
	// account with no groups holds an empty array, not an absent one.
	_, err := s.pool.Exec(ctx, `
		insert into cmsauth (account_id, groups, permissions) values ($1, $2, $3)
		on conflict (account_id) do update set groups = excluded.groups, permissions = excluded.permissions`,
		id, orEmpty(p.Groups), orEmpty(p.Permissions))
	if err != nil {
		return fmt.Errorf("authn: setting cms profile for account %d: %w", id, err)
	}
	return nil
}

// ---------------------------------------------------------------- passkeys --

// StoredPasskey is a credential with the metadata used to list and revoke it.
type StoredPasskey struct {
	ID             []byte
	Name           string
	BackupEligible bool
	CreatedAt      time.Time
	LastUsedAt     *time.Time
	Credential     webauthn.Credential
}

// Passkeys returns an account's credentials. They are loaded on demand rather
// than with every account, because only the ceremonies need them.
func (s *Store) Passkeys(ctx context.Context, accountID int64) ([]webauthn.Credential, error) {
	rows, err := s.pool.Query(ctx,
		`select credential from passkey where account_id = $1 order by created_at`, accountID)
	if err != nil {
		return nil, fmt.Errorf("authn: loading passkeys: %w", err)
	}
	defer rows.Close()

	var creds []webauthn.Credential
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, fmt.Errorf("authn: loading passkeys: %w", err)
		}
		var c webauthn.Credential
		if err := json.Unmarshal(raw, &c); err != nil {
			return nil, fmt.Errorf("authn: decoding passkey: %w", err)
		}
		creds = append(creds, c)
	}
	return creds, rows.Err()
}

func (s *Store) ListPasskeys(ctx context.Context, accountID int64) ([]StoredPasskey, error) {
	rows, err := s.pool.Query(ctx, `
		select id, coalesce(name, ''), backup_eligible, created_at, last_used_at, credential
		  from passkey where account_id = $1 order by created_at`, accountID)
	if err != nil {
		return nil, fmt.Errorf("authn: listing passkeys: %w", err)
	}
	defer rows.Close()

	var keys []StoredPasskey
	for rows.Next() {
		var k StoredPasskey
		var raw []byte
		if err := rows.Scan(&k.ID, &k.Name, &k.BackupEligible, &k.CreatedAt, &k.LastUsedAt, &raw); err != nil {
			return nil, fmt.Errorf("authn: listing passkeys: %w", err)
		}
		if err := json.Unmarshal(raw, &k.Credential); err != nil {
			return nil, fmt.Errorf("authn: decoding passkey: %w", err)
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// AddPasskey stores a newly registered credential. Re-registering the same
// authenticator is a no-op rather than an error.
func (s *Store) AddPasskey(ctx context.Context, accountID int64, cred *webauthn.Credential, name string) error {
	raw, err := json.Marshal(cred)
	if err != nil {
		return fmt.Errorf("authn: encoding passkey: %w", err)
	}
	_, err = s.pool.Exec(ctx, `
		insert into passkey (id, account_id, credential, name, backup_eligible)
		values ($1, $2, $3, $4, $5)
		on conflict (id) do nothing`,
		cred.ID, accountID, raw, name, cred.Flags.BackupEligible)
	if err != nil {
		return fmt.Errorf("authn: storing passkey: %w", err)
	}
	return nil
}

// UpdatePasskey writes back a credential whose sign counter or flags advanced
// during a login.
func (s *Store) UpdatePasskey(ctx context.Context, cred *webauthn.Credential) error {
	raw, err := json.Marshal(cred)
	if err != nil {
		return fmt.Errorf("authn: encoding passkey: %w", err)
	}
	_, err = s.pool.Exec(ctx,
		`update passkey set credential = $2, last_used_at = now() where id = $1`, cred.ID, raw)
	if err != nil {
		return fmt.Errorf("authn: updating passkey: %w", err)
	}
	return nil
}

// DeletePasskey removes one credential, so losing a single device does not cost
// an account the rest of its keys.
func (s *Store) DeletePasskey(ctx context.Context, accountID int64, id []byte) error {
	tag, err := s.pool.Exec(ctx, `delete from passkey where account_id = $1 and id = $2`, accountID, id)
	if err != nil {
		return fmt.Errorf("authn: deleting passkey: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNoAccount
	}
	return nil
}

func (s *Store) DeletePasskeys(ctx context.Context, accountID int64) (int64, error) {
	tag, err := s.pool.Exec(ctx, `delete from passkey where account_id = $1`, accountID)
	if err != nil {
		return 0, fmt.Errorf("authn: deleting passkeys: %w", err)
	}
	return tag.RowsAffected(), nil
}

// ------------------------------------------------------- enrollment tokens --

// CreateEnrollToken issues a single-use link token. Only its hash is stored, so
// a database leak yields no usable links; the plaintext is returned once.
func (s *Store) CreateEnrollToken(ctx context.Context, accountID int64, ttl time.Duration) (string, error) {
	token, err := randomToken()
	if err != nil {
		return "", err
	}
	sum := hashToken(token)
	if _, err := s.pool.Exec(ctx,
		`insert into enroll_token (token_hash, account_id, expires_at) values ($1, $2, $3)`,
		sum[:], accountID, time.Now().Add(ttl)); err != nil {
		return "", fmt.Errorf("authn: creating enrollment token: %w", err)
	}
	return token, nil
}

// EnrollTokenAccount validates a token without spending it: opening the link —
// a mail client prefetching it, say — must not burn it.
func (s *Store) EnrollTokenAccount(ctx context.Context, token string) (*Account, error) {
	sum := hashToken(token)
	var accountID int64
	err := s.pool.QueryRow(ctx, `
		select account_id from enroll_token
		 where token_hash = $1 and used_at is null and expires_at > now()`, sum[:]).Scan(&accountID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrBadToken
	}
	if err != nil {
		return nil, fmt.Errorf("authn: reading enrollment token: %w", err)
	}
	return s.AccountByID(ctx, accountID)
}

// ConsumeEnrollToken spends the token. The update is the check, so two
// concurrent submissions cannot both succeed.
func (s *Store) ConsumeEnrollToken(ctx context.Context, token string) (*Account, error) {
	sum := hashToken(token)
	var accountID int64
	err := s.pool.QueryRow(ctx, `
		update enroll_token set used_at = now()
		 where token_hash = $1 and used_at is null and expires_at > now()
		returning account_id`, sum[:]).Scan(&accountID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrBadToken
	}
	if err != nil {
		return nil, fmt.Errorf("authn: consuming enrollment token: %w", err)
	}
	return s.AccountByID(ctx, accountID)
}

// ---------------------------------------------------------------- sessions --

func (s *Store) CreateSession(ctx context.Context, accountID int64, ttl time.Duration) (string, error) {
	token, err := randomToken()
	if err != nil {
		return "", err
	}
	sum := hashToken(token)
	if _, err := s.pool.Exec(ctx,
		`insert into session (token_hash, account_id, expires_at) values ($1, $2, $3)`,
		sum[:], accountID, time.Now().Add(ttl)); err != nil {
		return "", fmt.Errorf("authn: creating session: %w", err)
	}
	return token, nil
}

// SessionAccount resolves a session cookie to its account.
//
// The expiry slides so an active viewer is not signed out mid-film, but only
// once last_seen is older than slideAfter — otherwise every range request of a
// video would cost a write.
func (s *Store) SessionAccount(ctx context.Context, token string, ttl, slideAfter time.Duration) (*Account, time.Time, error) {
	sum := hashToken(token)
	var (
		accountID int64
		expires   time.Time
	)
	// The expiry returned is the one read before any slide, so it can only
	// understate how long the session has left — the safe direction for a
	// caller counting down towards it.
	err := s.pool.QueryRow(ctx, `
		with live as (
		    select token_hash, account_id, last_seen, expires_at
		      from session
		     where token_hash = $1 and expires_at > now()
		), slid as (
		    update session set last_seen = now(), expires_at = $2
		     where token_hash = (select token_hash from live where last_seen < now() - make_interval(secs => $3))
		)
		select account_id, expires_at from live`,
		sum[:], time.Now().Add(ttl), slideAfter.Seconds()).Scan(&accountID, &expires)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, time.Time{}, ErrNoSession
	}
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("authn: reading session: %w", err)
	}
	account, err := s.AccountByID(ctx, accountID)
	if err != nil {
		return nil, time.Time{}, err
	}
	return account, expires, nil
}

func (s *Store) DeleteSession(ctx context.Context, token string) error {
	sum := hashToken(token)
	if _, err := s.pool.Exec(ctx, `delete from session where token_hash = $1`, sum[:]); err != nil {
		return fmt.Errorf("authn: deleting session: %w", err)
	}
	return nil
}

func (s *Store) DeleteSessions(ctx context.Context, accountID int64) (int64, error) {
	tag, err := s.pool.Exec(ctx, `delete from session where account_id = $1`, accountID)
	if err != nil {
		return 0, fmt.Errorf("authn: deleting sessions: %w", err)
	}
	return tag.RowsAffected(), nil
}

// -------------------------------------------------------------- challenges --

// Challenge is the server-side half of a WebAuthn ceremony.
type Challenge struct {
	Purpose string
	// Nil for a usernameless login, where the account is not known until the
	// authenticator answers.
	AccountID *int64
	Data      *webauthn.SessionData
	// Reauthed records that the account re-entered its password to authorise a
	// registration. Finishing refuses a challenge without it.
	Reauthed bool
}

const (
	PurposeRegister = "register"
	PurposeLogin    = "login"
)

// SaveChallenge stores ceremony state and returns the id handed to the client.
// Only the hash of that id is stored, so a leaked row cannot complete a ceremony.
func (s *Store) SaveChallenge(ctx context.Context, c Challenge, ttl time.Duration) (string, error) {
	id, err := randomToken()
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(c.Data)
	if err != nil {
		return "", fmt.Errorf("authn: encoding challenge: %w", err)
	}
	var reauthAt *time.Time
	if c.Reauthed {
		now := time.Now()
		reauthAt = &now
	}
	sum := hashToken(id)
	if _, err := s.pool.Exec(ctx, `
		insert into webauthn_challenge (id, account_id, purpose, data, reauth_at, expires_at)
		values ($1, $2, $3, $4, $5, $6)`,
		sum[:], c.AccountID, c.Purpose, data, reauthAt, time.Now().Add(ttl)); err != nil {
		return "", fmt.Errorf("authn: saving challenge: %w", err)
	}
	return id, nil
}

// TakeChallenge consumes a challenge. The delete is the check, so a challenge
// cannot be replayed, and the purpose is matched so a login challenge cannot be
// crossed into a registration.
func (s *Store) TakeChallenge(ctx context.Context, purpose, id string) (*Challenge, error) {
	sum := hashToken(id)
	var (
		accountID *int64
		data      []byte
		reauthAt  *time.Time
	)
	err := s.pool.QueryRow(ctx, `
		delete from webauthn_challenge
		 where id = $1 and purpose = $2 and expires_at > now()
		returning account_id, data, reauth_at`, sum[:], purpose).Scan(&accountID, &data, &reauthAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrBadToken
	}
	if err != nil {
		return nil, fmt.Errorf("authn: taking challenge: %w", err)
	}
	var sd webauthn.SessionData
	if err := json.Unmarshal(data, &sd); err != nil {
		return nil, fmt.Errorf("authn: decoding challenge: %w", err)
	}
	return &Challenge{
		Purpose: purpose, AccountID: accountID, Data: &sd, Reauthed: reauthAt != nil,
	}, nil
}

// ------------------------------------------------------------ housekeeping --

// Cleanup drops rows that have expired. Correctness does not depend on it —
// every query filters on expiry — it only keeps the tables from growing.
func (s *Store) Cleanup(ctx context.Context) error {
	for _, stmt := range []string{
		`delete from session where expires_at < now()`,
		`delete from webauthn_challenge where expires_at < now()`,
		`delete from enroll_token where expires_at < now() or used_at is not null`,
	} {
		if _, err := s.pool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("authn: cleanup: %w", err)
		}
	}
	return nil
}

// randomToken returns a URL-safe secret with 256 bits of entropy.
func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("authn: generating a token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// hashToken is a plain SHA-256: the input already has full entropy, so there is
// nothing for a work factor to defend against.
func hashToken(token string) [32]byte {
	return sha256.Sum256([]byte(token))
}

func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// isUniqueViolation asks structurally rather than importing pgconn for one code.
func isUniqueViolation(err error) bool {
	var sqlErr interface{ SQLState() string }
	return errors.As(err, &sqlErr) && sqlErr.SQLState() == "23505"
}
