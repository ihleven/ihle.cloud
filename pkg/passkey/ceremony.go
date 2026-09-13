package passkey

import (
	"fmt"
	"io"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// Subject is what WebAuthn needs to know about whoever is enrolling or signing
// in. Four questions, none of which require knowing what an account is — which
// is the whole of this package's dependency on the application.
type Subject interface {
	// Handle is the opaque, stable id an authenticator stores and hands back
	// during a usernameless sign-in. It must not be anything derived from a
	// person: it is written to their device and is not a secret.
	Handle() []byte

	// Name identifies the subject to its owner; DisplayName is how an
	// authenticator labels the key when asking.
	Name() string
	DisplayName() string

	// Credentials are the passkeys already registered, which is what lets a
	// registration exclude a device that is already enrolled.
	Credentials() []webauthn.Credential
}

// Resolver answers which subject an authenticator says it holds a credential
// for. A usernameless sign-in offers no account name, so this is the only way
// from an assertion to whoever made it.
type Resolver func(rawID, handle []byte) (Subject, error)

// Session is what has to survive between the two halves of a ceremony: the
// challenge, and what was asked of it. The application keeps it — this package
// does not know where, and does not need to.
type Session = webauthn.SessionData

// BeginRegistration builds the options for adding a credential to a subject.
//
// Every credential is required to be discoverable and user-verifying: resident
// so signing in needs no account name, and verified so a stolen key alone is not
// enough — the device asks for a biometric or a PIN. No attachment preference,
// because a platform authenticator is as acceptable as a security key.
func (c *Ceremonies) BeginRegistration(s Subject) (*protocol.CredentialCreation, *Session, error) {
	// Offering the keys already enrolled lets the browser say "this one is
	// registered" instead of silently creating a duplicate.
	held := s.Credentials()
	exclude := make([]protocol.CredentialDescriptor, 0, len(held))
	for _, cred := range held {
		exclude = append(exclude, cred.Descriptor())
	}

	creation, session, err := c.wa.BeginRegistration(user{s},
		webauthn.WithExclusions(exclude),
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			ResidentKey:        protocol.ResidentKeyRequirementRequired,
			RequireResidentKey: protocol.ResidentKeyRequired(),
			UserVerification:   protocol.VerificationRequired,
		}))
	if err != nil {
		return nil, nil, fmt.Errorf("passkey: beginning registration: %w", err)
	}
	return creation, session, nil
}

// FinishRegistration verifies what the authenticator created and returns the
// credential to store.
//
// The response is read from a reader rather than a request, so that what this
// package verifies is the bytes, not a transport.
func (c *Ceremonies) FinishRegistration(s Subject, session Session, response io.Reader) (*webauthn.Credential, error) {
	parsed, err := protocol.ParseCredentialCreationResponseBody(response)
	if err != nil {
		return nil, fmt.Errorf("passkey: reading the registration response: %w", err)
	}
	credential, err := c.wa.CreateCredential(user{s}, session, parsed)
	if err != nil {
		return nil, fmt.Errorf("passkey: verifying the registration: %w", err)
	}
	return credential, nil
}

// BeginLogin builds the options for a usernameless sign-in: the authenticator
// says which credential it holds, so there is nobody to name up front.
func (c *Ceremonies) BeginLogin() (*protocol.CredentialAssertion, *Session, error) {
	assertion, session, err := c.wa.BeginDiscoverableLogin(
		webauthn.WithUserVerification(protocol.VerificationRequired))
	if err != nil {
		return nil, nil, fmt.Errorf("passkey: beginning login: %w", err)
	}
	return assertion, session, nil
}

// FinishLogin verifies an assertion and reports whose it was.
//
// The subject comes back because resolve found it on the way through: the
// library reports only the credential, and the caller needs to know who signed
// in, not merely that someone did.
func (c *Ceremonies) FinishLogin(session Session, response io.Reader, resolve Resolver) (Subject, *webauthn.Credential, error) {
	parsed, err := protocol.ParseCredentialRequestResponseBody(response)
	if err != nil {
		return nil, nil, fmt.Errorf("passkey: reading the login response: %w", err)
	}

	var found Subject
	handler := func(rawID, handle []byte) (webauthn.User, error) {
		s, err := resolve(rawID, handle)
		if err != nil {
			return nil, err
		}
		found = s
		return user{s}, nil
	}

	credential, err := c.wa.ValidateDiscoverableLogin(handler, session, parsed)
	if err != nil {
		return nil, nil, fmt.Errorf("passkey: verifying the login: %w", err)
	}
	if found == nil {
		// The library resolved a credential without ever asking the handler,
		// which should be impossible; refusing is safer than guessing.
		return nil, nil, fmt.Errorf("passkey: the assertion verified but named nobody")
	}
	return found, credential, nil
}

// user adapts a Subject to the library's interface, which is the one place the
// two vocabularies meet.
type user struct{ s Subject }

func (u user) WebAuthnID() []byte                         { return u.s.Handle() }
func (u user) WebAuthnName() string                       { return u.s.Name() }
func (u user) WebAuthnDisplayName() string                { return u.s.DisplayName() }
func (u user) WebAuthnCredentials() []webauthn.Credential { return u.s.Credentials() }
