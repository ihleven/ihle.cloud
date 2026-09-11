package authn

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// Passwords are hashed with argon2id and stored in PHC string format, so the
// algorithm and its parameters travel with the hash. That is what lets the
// bcrypt hashes carried over from the old schema keep verifying, and lets a
// stronger parameter set be adopted later without a flag day: a hash that does
// not match the current parameters is re-hashed the next time its owner signs
// in successfully, which is the one moment the plaintext is available.

// Current parameters. OWASP's guidance is a memory-hard configuration in the
// tens of megabytes; logins are rare here, so the cost is paid by an attacker
// far more often than by a person.
const (
	argonTime    uint32 = 3
	argonMemory  uint32 = 64 * 1024 // KiB
	argonThreads uint8  = 2
	argonKeyLen  uint32 = 32
	argonSaltLen        = 16
)

var errBadHash = errors.New("authn: unrecognised password hash")

// HashPassword returns a PHC-format argon2id hash, for callers that set a
// password without going through a sign-in — the admin commands.
func HashPassword(plain string) (string, error) { return hashPassword(plain) }

// hashPassword returns a PHC-format argon2id hash.
func hashPassword(plain string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("authn: generating a salt: %w", err)
	}
	threads := argonThreads
	if n := runtime.NumCPU(); uint8(n) < threads {
		threads = uint8(n)
	}
	sum := argon2.IDKey([]byte(plain), salt, argonTime, argonMemory, threads, argonKeyLen)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(sum)), nil
}

// verifyPassword checks plain against a stored hash. It reports whether the
// hash should be replaced — because it is bcrypt, or because it is argon2id
// with parameters weaker than the current ones.
func verifyPassword(hash, plain string) (ok bool, rehash bool, err error) {
	switch {
	case strings.HasPrefix(hash, "$argon2id$"):
		ok, params, err := verifyArgon2id(hash, plain)
		if err != nil {
			return false, false, err
		}
		weaker := params.memory < argonMemory || params.time < argonTime
		return ok, ok && weaker, nil

	case strings.HasPrefix(hash, "$2a$"), strings.HasPrefix(hash, "$2b$"), strings.HasPrefix(hash, "$2y$"):
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, false, nil
		}
		if err != nil {
			return false, false, fmt.Errorf("authn: verifying password: %w", err)
		}
		// Every bcrypt hash is a leftover from the old schema, so a successful
		// sign-in is the moment to upgrade it.
		return true, true, nil

	default:
		return false, false, errBadHash
	}
}

type argonParams struct {
	memory  uint32
	time    uint32
	threads uint8
}

func verifyArgon2id(hash, plain string) (bool, argonParams, error) {
	// $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return false, argonParams{}, errBadHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return false, argonParams{}, errBadHash
	}
	var p argonParams
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.memory, &p.time, &p.threads); err != nil {
		return false, argonParams{}, errBadHash
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, argonParams{}, errBadHash
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, argonParams{}, errBadHash
	}

	got := argon2.IDKey([]byte(plain), salt, p.time, p.memory, p.threads, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1, p, nil
}

// ------------------------------------------------------------- throttling --

// throttle limits how fast a given key — an account name, or a client address —
// may fail. Both are limited: per-account so one person's password cannot be
// ground down, per-address so a single client cannot sweep many accounts.
//
// It is in-process, so it does not survive a restart and does not coordinate
// across replicas. For an application of this size that is the right trade:
// it raises the cost of an online guessing attack without a shared store.
type throttle struct {
	mu       sync.Mutex
	failures map[string]*failureCount
	limit    int
	window   time.Duration
}

type failureCount struct {
	n     int
	until time.Time
}

func newThrottle(limit int, window time.Duration) *throttle {
	return &throttle{failures: map[string]*failureCount{}, limit: limit, window: window}
}

// blocked reports whether a key has failed too often to be allowed another try.
func (t *throttle) blocked(keys ...string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	for _, k := range keys {
		f, ok := t.failures[k]
		if !ok {
			continue
		}
		if now.After(f.until) {
			delete(t.failures, k)
			continue
		}
		if f.n >= t.limit {
			return true
		}
	}
	return false
}

func (t *throttle) fail(keys ...string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	for _, k := range keys {
		f, ok := t.failures[k]
		if !ok || now.After(f.until) {
			f = &failureCount{}
			t.failures[k] = f
		}
		f.n++
		// Each failure extends the window, so a persistent attacker stays
		// blocked while someone who mistypes once is let through shortly.
		f.until = now.Add(t.window)
	}
}

func (t *throttle) succeed(keys ...string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, k := range keys {
		delete(t.failures, k)
	}
}

// sweep drops entries whose window has passed, so the map cannot grow without
// bound from failed attempts against random names.
func (t *throttle) sweep() {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	for k, f := range t.failures {
		if now.After(f.until) {
			delete(t.failures, k)
		}
	}
}
