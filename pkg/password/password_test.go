package password

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("hashPassword: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("hash is not argon2id: %q", hash)
	}

	ok, rehash, err := Verify(hash, "correct horse battery staple")
	if err != nil || !ok {
		t.Fatalf("the right password did not verify: ok=%v err=%v", ok, err)
	}
	if rehash {
		t.Error("a hash written with the current parameters asked to be replaced")
	}

	ok, _, err = Verify(hash, "wrong")
	if err != nil {
		t.Fatalf("verifying a wrong password errored: %v", err)
	}
	if ok {
		t.Error("the wrong password verified")
	}
}

// Every salt is fresh, so the same password hashes differently each time.
func TestHashPasswordIsSalted(t *testing.T) {
	first, err := Hash("same")
	if err != nil {
		t.Fatalf("hashPassword: %v", err)
	}
	second, err := Hash("same")
	if err != nil {
		t.Fatalf("hashPassword: %v", err)
	}
	if first == second {
		t.Error("two hashes of the same password are identical; the salt is not random")
	}
}

// The hashes carried over from the old credentials column are bcrypt. They must
// keep verifying, and a successful sign-in is when they get upgraded.
func TestBcryptVerifiesAndAsksForRehash(t *testing.T) {
	raw, err := bcrypt.GenerateFromPassword([]byte("legacy"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("generating a bcrypt hash: %v", err)
	}

	ok, rehash, err := Verify(string(raw), "legacy")
	if err != nil || !ok {
		t.Fatalf("a bcrypt hash did not verify: ok=%v err=%v", ok, err)
	}
	if !rehash {
		t.Error("a bcrypt hash should be upgraded on the next successful sign-in")
	}

	ok, rehash, err = Verify(string(raw), "wrong")
	if err != nil {
		t.Fatalf("verifying a wrong password errored: %v", err)
	}
	if ok || rehash {
		t.Error("a failed bcrypt verification must not report success or ask for a rehash")
	}
}

// Parameters travel with the hash, so one written under weaker settings still
// verifies — and is flagged for replacement.
func TestWeakerArgonParametersAskForRehash(t *testing.T) {
	// Built directly rather than through hashPassword, which only ever writes
	// the current parameters.
	salt := []byte("sixteen-byte-slt")
	sum := argon2.IDKey([]byte("weak-password"), salt, 1, 1024, 1, argonKeyLen)
	hash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, 1024, 1, 1,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(sum))

	ok, rehash, err := Verify(hash, "weak-password")
	if err != nil || !ok {
		t.Fatalf("a weak-parameter hash did not verify: ok=%v err=%v", ok, err)
	}
	if !rehash {
		t.Error("a hash weaker than the current parameters should be replaced")
	}
}

func TestUnrecognisedHashIsAnError(t *testing.T) {
	for _, hash := range []string{"", "plaintext", "$md5$nope", "$argon2id$missing-fields"} {
		if _, _, err := Verify(hash, "x"); !errors.Is(err, errBadHash) {
			t.Errorf("Verify(%q): got %v, want errBadHash", hash, err)
		}
	}
}
