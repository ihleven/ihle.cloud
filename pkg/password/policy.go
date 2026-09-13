package password

import (
	"context"
	"unicode/utf8"
)

// The policy is advice, not enforcement: no composition rules, no rotation, no
// reuse check, because those push people towards predictable passwords rather
// than away from them — and the checks that remain say what they found and then
// let whoever is choosing decide.
//
// The minimum is passed in rather than fixed here. It is a judgement about how
// much to nag, which the application makes: the number has to be the same in the
// terminal, in the admin form and in the sentence the form shows, and the only
// way to guarantee that is to have one place own it.
//
// MaxLength is the one hard bound, and it is a technical one — it keeps a
// pathological input from reaching the hasher. It is well above anything someone
// types.
const MaxLength = 64

// Advice is what could be found out about a proposed password.
//
// It is measured and returned rather than acted on, so that the terminal and the
// browser weigh the same facts and differ only in how they ask. Nothing here
// refuses a password: the person choosing it knows things the check does not.
type Advice struct {
	Length   int  `json:"length"`
	TooShort bool `json:"too_short"`

	// MinLength is the threshold TooShort was measured against. Reported so that
	// whatever tells someone their password is short can name the number without
	// keeping a copy of it.
	MinLength int `json:"min_length"`

	// Breaches is how many times the password appears in known breaches, and
	// Unchecked says why that could not be established — the two are exclusive,
	// and a zero count with no reason means it was checked and not found.
	Breaches  int    `json:"breaches"`
	Unchecked string `json:"unchecked,omitempty"`
}

// Concerning reports whether there is anything worth putting to the person.
func (a Advice) Concerning() bool {
	return a.TooShort || a.Breaches > 0 || a.Unchecked != ""
}

// TooLong reports the one hard bound. Counted in characters rather than bytes,
// because len() on the raw input would let eight emoji pass as "32", which is
// not what a length means to the person typing.
func TooLong(plain string) bool {
	return utf8.RuneCountInString(plain) > MaxLength
}

// Check measures a password against the policy.
//
// A breach lookup is a network call, so its failure is reported as something
// unknown rather than as a verdict: not being able to ask is not the same as
// asking and being told the password is clean.
func Check(ctx context.Context, checker *BreachChecker, plain string, minLength int) Advice {
	advice := Advice{Length: utf8.RuneCountInString(plain), MinLength: minLength}
	advice.TooShort = advice.Length < minLength

	if checker == nil {
		checker = &BreachChecker{}
	}
	count, err := checker.Count(ctx, plain)
	if err != nil {
		advice.Unchecked = err.Error()
		return advice
	}
	advice.Breaches = count
	return advice
}
