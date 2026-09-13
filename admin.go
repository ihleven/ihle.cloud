package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"golang.org/x/term"

	"github.com/ihleven/ihlvn/app/db"
	"github.com/ihleven/ihlvn/pkg/authn"
)

// Account administration.
//
// There is no self-service registration: an operator creates an account and
// gives the person a password or a single-use enrollment link out of band. That
// keeps the public surface at the sign-in form and nothing else.

type AccountCmd struct {
	Action string   `arg:"positional" help:"list | add | passwd | permissions | groups | enroll | passkeys | revoke | disable | enable"`
	Args   []string `arg:"positional" help:"arguments for the action"`

	TTL     time.Duration `arg:"--ttl" default:"15m" help:"lifetime of an enrollment link"`
	BaseURL string        `arg:"--base-url,env:PUBLIC_URL" default:"http://localhost:8000" help:"base URL for enrollment links. Shares PUBLIC_URL with the server, so a link issued on the server points at the right host without being told twice"`
}

func (c *AccountCmd) Run(flags Flags) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pg, err := db.New(flags.DbConn)
	if err != nil {
		return fmt.Errorf("connecting to %s: %w", flags.DbConn, err)
	}
	defer pg.Close()

	// The commands share the server's migration step, so a command can never
	// talk to a schema the binary does not expect.
	if err := authn.Migrate(ctx, pg.Pool()); err != nil {
		return err
	}
	store := authn.NewStore(pg.Pool())

	switch c.Action {
	case "", "list":
		return listAccounts(ctx, store, os.Stdout)
	case "add":
		return addAccount(ctx, store, c.Args)
	case "passwd":
		return setPassword(ctx, store, c.Args)
	case "enroll":
		return issueEnrollment(ctx, store, c.Args, c.BaseURL, c.TTL)
	case "permissions":
		return setCMSField(ctx, store, c.Args, "permissions")
	case "groups":
		return setCMSField(ctx, store, c.Args, "groups")
	case "passkeys":
		return listPasskeys(ctx, store, os.Stdout, c.Args)
	case "revoke":
		return revokeCredentials(ctx, store, c.Args)
	case "disable":
		return setDisabled(ctx, store, c.Args, true)
	case "enable":
		return setDisabled(ctx, store, c.Args, false)
	default:
		return fmt.Errorf("unknown action %q; try list, add, passwd, permissions, groups, enroll, passkeys, revoke, disable or enable", c.Action)
	}
}

func firstArg(args []string, what string) (string, error) {
	if len(args) == 0 || args[0] == "" {
		return "", fmt.Errorf("expected %s", what)
	}
	return args[0], nil
}

func listAccounts(ctx context.Context, store *authn.Store, out io.Writer) error {
	accounts, err := store.ListAccounts(ctx)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tDISPLAY\tEMAIL\tSTATE\tCREDENTIALS\tGROUPS\tPERMISSIONS")
	for _, a := range accounts {
		keys, err := store.ListPasskeys(ctx, a.ID)
		if err != nil {
			return err
		}

		state := "active"
		if a.Disabled {
			state = "disabled"
		}
		var credentials []string
		if a.HasPassword() {
			credentials = append(credentials, "password")
		}
		if len(keys) > 0 {
			credentials = append(credentials, fmt.Sprintf("%d passkey(s)", len(keys)))
		}
		if len(credentials) == 0 {
			credentials = append(credentials, "none")
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%d\n",
			a.Name, a.DisplayName, a.Email, state,
			strings.Join(credentials, ", "),
			strings.Join(a.CMS.Groups, ","), len(a.CMS.Permissions))
	}
	if err := w.Flush(); err != nil {
		return err
	}

	// A permission the CMS does not define is dropped when the scope is built,
	// so it grants nothing while still looking like a grant in the database.
	for _, a := range accounts {
		if unknown := a.UnregisteredPermissions(); len(unknown) > 0 {
			fmt.Fprintf(out, "\n%s holds %d permission(s) this build does not define, which grant nothing:\n  %s\n",
				a.Name, len(unknown), strings.Join(unknown, "\n  "))
		}
	}
	return nil
}

func addAccount(ctx context.Context, store *authn.Store, args []string) error {
	name, err := firstArg(args, "an account name")
	if err != nil {
		return err
	}
	if len(args) < 2 {
		return errors.New("expected an email address")
	}
	email := args[1]
	display := name
	if len(args) > 2 {
		display = strings.Join(args[2:], " ")
	}

	account, err := store.CreateAccount(ctx, name, display, email)
	if errors.Is(err, authn.ErrExists) {
		return fmt.Errorf("an account named %q already exists", name)
	}
	if err != nil {
		return err
	}
	fmt.Printf("created %s (%s)\nSet a password with: ihlvn account passwd %s\n",
		account.Name, account.Email, account.Name)
	return nil
}

func setPassword(ctx context.Context, store *authn.Store, args []string) error {
	name, err := firstArg(args, "an account name")
	if err != nil {
		return err
	}
	account, err := store.AccountByName(ctx, name)
	if errors.Is(err, authn.ErrNoAccount) {
		return fmt.Errorf("no active account named %q", name)
	}
	if err != nil {
		return err
	}

	password, err := readPasswordTwice()
	if err != nil {
		return err
	}

	// The command's deadline has been running while someone typed, so the work
	// that follows gets its own. Otherwise a slow typist times out.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()

	if err := confirmWeakPassword(ctx, password); err != nil {
		return err
	}
	hash, err := authn.HashPassword(password)
	if err != nil {
		return err
	}
	if err := store.SetPasswordHash(ctx, account.ID, hash); err != nil {
		return err
	}
	fmt.Printf("password set for %s\n", account.Name)
	return nil
}

// readPasswordTwice reads without echoing and confirms, so a typo does not lock
// someone out of an account they cannot reset themselves.
func readPasswordTwice() (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", errors.New("a password must be typed into a terminal, not piped")
	}

	fmt.Print("New password: ")
	first, err := term.ReadPassword(fd)
	fmt.Println()
	if err != nil {
		return "", err
	}
	fmt.Print("Repeat: ")
	second, err := term.ReadPassword(fd)
	fmt.Println()
	if err != nil {
		return "", err
	}

	if string(first) != string(second) {
		return "", errors.New("the two entries do not match")
	}
	// Counted in characters, not bytes: len() on the raw input would let eight
	// emoji pass as "32", which is not what a length means to the person typing.
	// Only the ceiling is refused here; being short is a warning, raised with the
	// other advice once the password is known to be what was intended.
	if authn.TooLong(string(first)) {
		return "", fmt.Errorf("use at most %d characters", authn.MaxPasswordLength)
	}
	return string(first), nil
}

// confirmWeakPassword says what is wrong with a password and asks whether to use
// it regardless.
//
// Everything here is advice rather than a verdict. Refusing a short password
// would substitute a number for the operator's judgement, and refusing a
// breached one would let a third party's corpus decide what this system accepts
// while leaving no way to override it. A screening failure — no network, service
// down — is reported and blocks nothing, or setting a password would depend on
// being online.
//
// What counts as weak is authn's to decide, so that the terminal and the admin
// UI hold a password to the same standard; only the asking belongs here.
func confirmWeakPassword(ctx context.Context, password string) error {
	advice := authn.CheckPassword(ctx, nil, password)
	if !advice.Concerning() {
		return nil
	}
	for _, warning := range advice.Warnings() {
		fmt.Printf("\n%s\n", warning)
	}
	return confirm("Use it anyway?")
}

// confirm asks a yes/no question, defaulting to no.
func confirm(question string) error {
	fmt.Printf("%s [y/N] ", question)
	answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return err
	}
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return nil
	default:
		return errors.New("cancelled")
	}
}

func issueEnrollment(ctx context.Context, store *authn.Store, args []string, baseURL string, ttl time.Duration) error {
	name, err := firstArg(args, "an account name")
	if err != nil {
		return err
	}
	account, err := store.AccountByName(ctx, name)
	if errors.Is(err, authn.ErrNoAccount) {
		return fmt.Errorf("no active account named %q", name)
	}
	if err != nil {
		return err
	}
	token, err := store.CreateEnrollToken(ctx, account.ID, ttl)
	if err != nil {
		return err
	}
	// Printed rather than logged, so the link never lands in a request log.
	// An account with no password yet chooses one as it enrols, so the link is
	// the whole of what it needs — which is what lets an account be created and
	// handed over without a password ever being conveyed.
	asked := "They will be asked for their password to complete it."
	if !account.HasPassword() {
		asked = "They will choose a password as they complete it."
	}
	fmt.Printf("Single-use enrollment link for %s, valid for %s:\n\n  %s\n\n%s\n",
		account.Name, ttl, authn.EnrollURL(baseURL, token), asked)
	return nil
}

// setCMSField shows or replaces one half of an account's CMS profile — the
// groups an entry's ACL is matched against, or the permission names a scope is
// built from.
//
// Replacing rather than adding keeps it obvious what an account ends up with:
// the argument list is the result, not a delta.
func setCMSField(ctx context.Context, store *authn.Store, args []string, field string) error {
	name, err := firstArg(args, "an account name")
	if err != nil {
		return err
	}
	account, err := store.AccountByName(ctx, name)
	if errors.Is(err, authn.ErrNoAccount) {
		return fmt.Errorf("no active account named %q", name)
	}
	if err != nil {
		return err
	}

	profile := account.CMS
	values := args[1:]
	if len(values) == 0 {
		current := profile.Permissions
		if field == "groups" {
			current = profile.Groups
		}
		if len(current) == 0 {
			fmt.Printf("%s has no %s\n", account.Name, field)
			return nil
		}
		fmt.Printf("%s %s:\n  %s\n", account.Name, field, strings.Join(current, "\n  "))
		return nil
	}

	if field == "groups" {
		profile.Groups = values
	} else {
		profile.Permissions = values
	}
	if err := store.SetCMSProfile(ctx, account.ID, profile); err != nil {
		return err
	}
	fmt.Printf("%s %s: %s\n", account.Name, field, strings.Join(values, ", "))

	// A name this build does not define is dropped when the scope is built, so
	// it would look granted and grant nothing.
	account.CMS = profile
	if unknown := account.UnregisteredPermissions(); len(unknown) > 0 {
		fmt.Printf("\nWarning: %d of these are not defined in this build and will grant nothing:\n  %s\n",
			len(unknown), strings.Join(unknown, "\n  "))
	}
	return nil
}

func listPasskeys(ctx context.Context, store *authn.Store, out io.Writer, args []string) error {
	name, err := firstArg(args, "an account name")
	if err != nil {
		return err
	}
	account, err := store.AccountByName(ctx, name)
	if errors.Is(err, authn.ErrNoAccount) {
		return fmt.Errorf("no active account named %q", name)
	}
	if err != nil {
		return err
	}

	keys, err := store.ListPasskeys(ctx, account.ID)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		fmt.Fprintf(out, "%s has no passkeys\n", account.Name)
		return nil
	}

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tADDED\tLAST USED\tSYNCABLE")
	for _, k := range keys {
		last := "never"
		if k.LastUsedAt != nil {
			last = k.LastUsedAt.Format(time.DateOnly)
		}
		// A syncable key predates the device-bound rule; registration refuses
		// them now, so any that show here are worth replacing.
		syncable := "no"
		if k.BackupEligible {
			syncable = "YES — not device-bound"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			orDash(k.Name), k.CreatedAt.Format(time.DateOnly), last, syncable)
	}
	return w.Flush()
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func revokeCredentials(ctx context.Context, store *authn.Store, args []string) error {
	name, err := firstArg(args, "an account name")
	if err != nil {
		return err
	}
	account, err := store.AccountByName(ctx, name)
	if errors.Is(err, authn.ErrNoAccount) {
		return fmt.Errorf("no active account named %q", name)
	}
	if err != nil {
		return err
	}

	keys, err := store.DeletePasskeys(ctx, account.ID)
	if err != nil {
		return err
	}
	sessions, err := store.DeleteSessions(ctx, account.ID)
	if err != nil {
		return err
	}
	fmt.Printf("removed %d passkey(s) and %d session(s) for %s\n", keys, sessions, account.Name)
	return nil
}

func setDisabled(ctx context.Context, store *authn.Store, args []string, disabled bool) error {
	name, err := firstArg(args, "an account name")
	if err != nil {
		return err
	}
	id, err := store.SetDisabled(ctx, name, disabled)
	if err != nil {
		if errors.Is(err, authn.ErrNoAccount) {
			return fmt.Errorf("no account named %q", name)
		}
		return err
	}

	if !disabled {
		fmt.Printf("%s enabled\n", name)
		return nil
	}
	// Loading already refuses a disabled account, so its sessions are inert;
	// dropping the rows makes that explicit rather than implied.
	sessions, err := store.DeleteSessions(ctx, id)
	if err != nil {
		return err
	}
	fmt.Printf("%s disabled, %d session(s) ended\n", name, sessions)
	return nil
}
