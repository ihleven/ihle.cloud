package authn

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func ceremonyService(t *testing.T) (*Service, *Store) {
	t.Helper()
	store, _ := testStore(t)
	svc, err := New(store, Config{ChallengeTTL: time.Minute}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return svc, store
}

func finishRequest(cookie, header string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/auth/passkey/login/finish", nil)
	if cookie != "" {
		r.AddCookie(&http.Cookie{Name: "ihlvn_ceremony", Value: cookie})
	}
	if header != "" {
		r.Header.Set(CeremonyHeader, header)
	}
	return r
}

// A browser can have two ceremonies open at once: one offered silently in the
// username field's autofill, one started by the button. The cookie can only
// name the later of the two, so whichever finished second used to be told no
// ceremony was in progress — which is what made passkey sign-in fail in Chrome
// while working in Safari.
func TestTwoCeremoniesCanBeInFlight(t *testing.T) {
	svc, store := ceremonyService(t)
	ctx := t.Context()

	first, err := store.SaveChallenge(ctx, Challenge{Purpose: PurposeLogin}, time.Minute)
	if err != nil {
		t.Fatalf("saving the first challenge: %v", err)
	}
	second, err := store.SaveChallenge(ctx, Challenge{Purpose: PurposeLogin}, time.Minute)
	if err != nil {
		t.Fatalf("saving the second challenge: %v", err)
	}

	// The cookie names the second, because it began last. The first completes
	// anyway, by naming itself.
	if _, err := svc.takeCeremony(finishRequest(second, first), PurposeLogin); err != nil {
		t.Fatalf("the earlier ceremony could not be completed: %v", err)
	}

	// And the later one still works afterwards.
	if _, err := svc.takeCeremony(finishRequest(second, second), PurposeLogin); err != nil {
		t.Fatalf("the later ceremony was destroyed by the earlier one: %v", err)
	}
}

// Without the header the cookie still decides, so a page that does not send one
// behaves as before.
func TestCookieIsUsedWhenNoHeaderIsSent(t *testing.T) {
	svc, store := ceremonyService(t)

	id, err := store.SaveChallenge(t.Context(), Challenge{Purpose: PurposeLogin}, time.Minute)
	if err != nil {
		t.Fatalf("saving: %v", err)
	}
	if _, err := svc.takeCeremony(finishRequest(id, ""), PurposeLogin); err != nil {
		t.Fatalf("the cookie alone did not identify the ceremony: %v", err)
	}
}

// A challenge is single-use: that, not the cookie, is what stops a replay.
func TestACeremonyCannotBeCompletedTwice(t *testing.T) {
	svc, store := ceremonyService(t)

	id, err := store.SaveChallenge(t.Context(), Challenge{Purpose: PurposeLogin}, time.Minute)
	if err != nil {
		t.Fatalf("saving: %v", err)
	}
	if _, err := svc.takeCeremony(finishRequest(id, id), PurposeLogin); err != nil {
		t.Fatalf("first completion: %v", err)
	}
	if _, err := svc.takeCeremony(finishRequest(id, id), PurposeLogin); err == nil {
		t.Fatal("a spent challenge was accepted a second time")
	}
}

// A challenge raised for one purpose must not complete another.
func TestACeremonyIsScopedToItsPurpose(t *testing.T) {
	svc, store := ceremonyService(t)

	id, err := store.SaveChallenge(t.Context(), Challenge{Purpose: PurposeRegister}, time.Minute)
	if err != nil {
		t.Fatalf("saving: %v", err)
	}
	if _, err := svc.takeCeremony(finishRequest(id, id), PurposeLogin); err == nil {
		t.Fatal("a registration challenge was accepted as a login")
	}
}

func TestNoCeremonyAtAll(t *testing.T) {
	svc, _ := ceremonyService(t)

	if _, err := svc.takeCeremony(finishRequest("", ""), PurposeLogin); err == nil {
		t.Fatal("a finish with neither cookie nor header was accepted")
	}
}
