package authn

import (
	"context"
	"fmt"
	"unicode/utf8"
)

// The password policy is advice, not enforcement: no composition rules, no
// rotation, no reuse check, because those push people towards predictable
// passwords rather than away from them — and the two checks that remain say so
// and then let whoever is choosing decide.
//
// MinPasswordLength is where a warning starts, not a rule. Length is a weak
// signal anyway: a twelve-character password that has leaked is worse than a
// short one nobody has ever used, which is what the breach screening is for.
//
// MaxPasswordLength is the one hard bound, and it is a technical one — it keeps
// a pathological input from being submitted. It is well above anything someone
// types.
const (
	MinPasswordLength = 12
	MaxPasswordLength = 64
)

// PasswordAdvice is what could be found out about a proposed password.
//
// It is measured and returned rather than acted on, so that the terminal and the
// browser weigh the same facts and differ only in how they ask. Nothing here
// refuses a password: the person choosing it knows things the check does not.
type PasswordAdvice struct {
	Length   int  `json:"length"`
	TooShort bool `json:"too_short"`

	// Breaches is how many times the password appears in known breaches, and
	// Unchecked says why that could not be established — the two are exclusive,
	// and a zero count with no reason means it was checked and not found.
	Breaches  int    `json:"breaches"`
	Unchecked string `json:"unchecked,omitempty"`
}

// Concerning reports whether there is anything worth putting to the person.
func (a PasswordAdvice) Concerning() bool {
	return a.TooShort || a.Breaches > 0 || a.Unchecked != ""
}

// Warnings renders the advice, one concern per line, in the order they matter:
// a leak outranks a length, because it is evidence rather than a heuristic.
func (a PasswordAdvice) Warnings() []string {
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
			a.Length, MinPasswordLength))
	}
	return out
}

// TooLong reports the one hard bound. Counted in characters rather than bytes,
// because len() on the raw input would let eight emoji pass as "32", which is
// not what a length means to the person typing.
func TooLong(password string) bool {
	return utf8.RuneCountInString(password) > MaxPasswordLength
}

// CheckPassword measures a password against the policy.
//
// A breach lookup is a network call, so its failure is reported as something
// unknown rather than as a verdict: not being able to ask is not the same as
// asking and being told the password is clean.
func CheckPassword(ctx context.Context, checker *BreachChecker, password string) PasswordAdvice {
	advice := PasswordAdvice{Length: utf8.RuneCountInString(password)}
	advice.TooShort = advice.Length < MinPasswordLength

	if checker == nil {
		checker = &BreachChecker{}
	}
	count, err := checker.Count(ctx, password)
	if err != nil {
		advice.Unchecked = err.Error()
		return advice
	}
	advice.Breaches = count
	return advice
}

// times reads a count as English, so a warning can be read aloud.
func times(n int) string {
	if n == 1 {
		return "once"
	}
	return fmt.Sprintf("%d times", n)
}
