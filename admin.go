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

	"github.com/ihleven/ihlvn/app/auth"
	"github.com/ihleven/ihlvn/app/db"
	"github.com/ihleven/ihlvn/pkg/password"
)

// Account administration from a terminal.
//
// The rules live in auth.Admin, which the admin section calls too; what is here
// is a terminal's half of the conversation — reading arguments, laying out
// tables, and asking questions that only make sense when someone is watching.
//
// There is no self-service registration: an operator creates an account and
// hands over a single-use enrollment link out of band. That keeps the public
// surface at the sign-in form and nothing else.
//
// Nothing here puts an account in the context, so the lock-out guard does not
// apply: the terminal is deliberately the way back in when the browser has
// locked someone out.

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
	if err := db.Migrate(ctx, pg.Pool()); err != nil {
		return err
	}
	admin := auth.NewAdmin(auth.NewStore(pg.Pool()), c.BaseURL, c.TTL, flags.MinPasswordLength)

	switch c.Action {
	case "", "list":
		return listAccounts(ctx, admin, os.Stdout)
	case "add":
		return addAccount(ctx, admin, c.Args)
	case "passwd":
		return setPassword(ctx, admin, c.Args)
	case "enroll":
		return issueEnrollment(ctx, admin, c.Args)
	case "permissions":
		return setCMSField(ctx, admin, c.Args, "permissions")
	case "groups":
		return setCMSField(ctx, admin, c.Args, "groups")
	case "passkeys":
		return listPasskeys(ctx, admin, os.Stdout, c.Args)
	case "revoke":
		return revokeCredentials(ctx, admin, c.Args)
	case "disable":
		return setDisabled(ctx, admin, c.Args, true)
	case "enable":
		return setDisabled(ctx, admin, c.Args, false)
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

// named puts a failure into a terminal's words. Only the "no such account" case
// needs it: everything else already reads as a sentence, because auth.Admin
// writes its refusals for a person rather than for a status code.
func named(name string, err error) error {
	if errors.Is(err, auth.ErrNoAccount) {
		return fmt.Errorf("no account named %q", name)
	}
	return err
}

func listAccounts(ctx context.Context, admin *auth.Admin, out io.Writer) error {
	accounts, err := admin.List(ctx)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tDISPLAY\tEMAIL\tSTATE\tCREDENTIALS\tGROUPS\tPERMISSIONS")
	for _, a := range accounts {
		state := "active"
		if a.Disabled {
			state = "disabled"
		}
		var credentials []string
		if a.HasPassword {
			credentials = append(credentials, "password")
		}
		if a.Passkeys > 0 {
			credentials = append(credentials, fmt.Sprintf("%d passkey(s)", a.Passkeys))
		}
		if len(credentials) == 0 {
			credentials = append(credentials, "none")
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%d\n",
			a.Name, a.DisplayName, a.Email, state,
			strings.Join(credentials, ", "),
			strings.Join(a.Groups, ","), len(a.Permissions))
	}
	if err := w.Flush(); err != nil {
		return err
	}

	// A permission the CMS does not define is dropped when the scope is built,
	// so it grants nothing while still looking like a grant in the database.
	for _, a := range accounts {
		if len(a.Unregistered) > 0 {
			fmt.Fprintf(out, "\n%s holds %d permission(s) this build does not define, which grant nothing:\n  %s\n",
				a.Name, len(a.Unregistered), strings.Join(a.Unregistered, "\n  "))
		}
	}
	return nil
}

func addAccount(ctx context.Context, admin *auth.Admin, args []string) error {
	name, err := firstArg(args, "an account name")
	if err != nil {
		return err
	}
	if len(args) < 2 {
		return errors.New("expected an email address")
	}
	display := name
	if len(args) > 2 {
		display = strings.Join(args[2:], " ")
	}

	account, err := admin.Create(ctx, auth.NewAccount{
		Name: name, Email: args[1], DisplayName: display,
	})
	if err != nil {
		return err
	}
	// Both steps, in order: a device cannot be registered without a password to
	// prove, so the link is useless until there is one.
	fmt.Printf("created %s (%s)\nNext: ihlvn account passwd %s, then ihlvn account enroll %s\n",
		account.Name, account.Email, account.Name, account.Name)
	return nil
}

func setPassword(ctx context.Context, admin *auth.Admin, args []string) error {
	name, err := firstArg(args, "an account name")
	if err != nil {
		return err
	}

	plain, err := readPasswordTwice()
	if err != nil {
		return err
	}

	// The command's deadline has been running while someone typed, so the work
	// that follows gets its own. Otherwise a slow typist times out.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()

	// Asked for without confirmation first: the screening gets a chance to
	// object, and nothing is written while it does.
	result, err := admin.SetPassword(ctx, name, plain, false)
	if err != nil {
		return named(name, err)
	}
	if !result.Set {
		for _, warning := range warnings(result.Advice) {
			fmt.Printf("\n%s\n", warning)
		}
		if err := confirm("Use it anyway?"); err != nil {
			return err
		}
		if _, err := admin.SetPassword(ctx, name, plain, true); err != nil {
			return named(name, err)
		}
	}
	fmt.Printf("password set for %s\n", name)
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
	return string(first), nil
}

// warnings puts the advice into words, one concern per line, in the order they
// matter: a leak outranks a length, because it is evidence rather than a
// heuristic.
//
// The wording lives here rather than with the policy because it is this
// command's half of a conversation — the admin form says the same things in
// German, to someone who has not asked to be lectured about entropy.
func warnings(a password.Advice) []string {
	var out []string
	if a.Breaches > 0 {
		out = append(out, fmt.Sprintf(
			"This password appears %s in known breaches. "+
				"Anything that has leaked is in the lists attackers try first, however long it is.",
			times(a.Breaches)))
	}
	if a.Unchecked != "" {
		out = append(out, "Could not check this password against known breaches: "+a.Unchecked)
	}
	if a.TooShort {
		out = append(out, fmt.Sprintf(
			"This password is %d characters; %d or more is the usual advice. "+
				"Short passwords are the ones that fall first if the database is ever leaked.",
			a.Length, a.MinLength))
	}
	return out
}

// times reads a count as English, so a warning can be read aloud.
func times(n int) string {
	if n == 1 {
		return "once"
	}
	return fmt.Sprintf("%d times", n)
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

func issueEnrollment(ctx context.Context, admin *auth.Admin, args []string) error {
	name, err := firstArg(args, "an account name")
	if err != nil {
		return err
	}
	link, err := admin.IssueEnrollment(ctx, name)
	if err != nil {
		return named(name, err)
	}

	// Printed rather than logged, so the link never lands in a request log. The
	// password is asked for as well when the link is redeemed, and is not in this
	// output: the two have to travel separately or the pair of them is one
	// factor.
	fmt.Printf("Single-use enrollment link for %s, valid until %s:\n\n  %s\n\n"+
		"They will be asked for their password to complete it.\n",
		name, link.ExpiresAt.Format(time.TimeOnly), link.URL)
	return nil
}

// setCMSField shows or replaces one half of an account's CMS profile — the
// groups an entry's ACL is matched against, or the permission names a scope is
// built from.
//
// Replacing rather than adding keeps it obvious what an account ends up with:
// the argument list is the result, not a delta.
func setCMSField(ctx context.Context, admin *auth.Admin, args []string, field string) error {
	name, err := firstArg(args, "an account name")
	if err != nil {
		return err
	}
	account, err := admin.Get(ctx, name)
	if err != nil {
		return named(name, err)
	}

	values := args[1:]
	if len(values) == 0 {
		current := account.Permissions
		if field == "groups" {
			current = account.Groups
		}
		if len(current) == 0 {
			fmt.Printf("%s has no %s\n", account.Name, field)
			return nil
		}
		fmt.Printf("%s %s:\n  %s\n", account.Name, field, strings.Join(current, "\n  "))
		return nil
	}

	// Everything else about the account is carried over unchanged: the edit
	// replaces an account wholesale, and this command is only about one field.
	edit := auth.AccountEdit{
		DisplayName: account.DisplayName,
		Email:       account.Email,
		Groups:      account.Groups,
		Permissions: account.Permissions,
		Disabled:    account.Disabled,
	}
	if field == "groups" {
		edit.Groups = values
	} else {
		edit.Permissions = values
	}

	updated, err := admin.Update(ctx, name, edit)
	if err != nil {
		return named(name, err)
	}
	fmt.Printf("%s %s: %s\n", updated.Name, field, strings.Join(values, ", "))

	// A name this build does not define is dropped when the scope is built, so
	// it would look granted and grant nothing. Only for permissions: groups are
	// matched against an entry's ACL and are not registered anywhere, so there is
	// no such thing as an unknown one.
	if field == "permissions" && len(updated.Unregistered) > 0 {
		fmt.Printf("\nWarning: %d of these are not defined in this build and will grant nothing:\n  %s\n",
			len(updated.Unregistered), strings.Join(updated.Unregistered, "\n  "))
	}
	return nil
}

func listPasskeys(ctx context.Context, admin *auth.Admin, out io.Writer, args []string) error {
	name, err := firstArg(args, "an account name")
	if err != nil {
		return err
	}
	keys, err := admin.Passkeys(ctx, name)
	if err != nil {
		return named(name, err)
	}
	if len(keys) == 0 {
		fmt.Fprintf(out, "%s has no passkeys\n", name)
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
		if k.Syncable {
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

func revokeCredentials(ctx context.Context, admin *auth.Admin, args []string) error {
	name, err := firstArg(args, "an account name")
	if err != nil {
		return err
	}
	revoked, err := admin.Revoke(ctx, name)
	if err != nil {
		return named(name, err)
	}
	fmt.Printf("removed %d passkey(s) and %d session(s) for %s\n",
		revoked.Passkeys, revoked.Sessions, name)
	return nil
}

func setDisabled(ctx context.Context, admin *auth.Admin, args []string, disabled bool) error {
	name, err := firstArg(args, "an account name")
	if err != nil {
		return err
	}
	account, err := admin.Get(ctx, name)
	if err != nil {
		return named(name, err)
	}

	edit := auth.AccountEdit{
		DisplayName: account.DisplayName,
		Email:       account.Email,
		Groups:      account.Groups,
		Permissions: account.Permissions,
		Disabled:    disabled,
	}
	if _, err := admin.Update(ctx, name, edit); err != nil {
		return named(name, err)
	}

	if disabled {
		// Updating drops the account's sessions as it disables it; loading
		// already refuses a disabled account, so this only makes that explicit.
		fmt.Printf("%s disabled, its sessions ended\n", name)
		return nil
	}
	fmt.Printf("%s enabled\n", name)
	return nil
}
