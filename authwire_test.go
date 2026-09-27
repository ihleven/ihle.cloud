package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ihleven/ihlvn/app/auth"
	"github.com/ihleven/ihlvn/app/cmsauth"
	"github.com/interhome-group/cms/pkg/errs"
)

// signedInAs answers every request with the given account, or with nobody when
// it is nil. Standing in for the session service keeps these tests on the
// entitlement decision rather than on how a cookie is read.
type signedInAs struct{ account *auth.Account }

func (s signedInAs) Authenticate(*http.Request) (*auth.Account, bool) {
	return s.account, s.account != nil
}

// The admin gate is the only thing standing between a signed-in account and
// endpoints that grant rights and set passwords. The navigation hides the
// section from accounts that may not use it, but that runs in the browser and
// proves nothing, so what is checked here is the server's refusal.
func TestRequireAdmin(t *testing.T) {
	tests := []struct {
		name        string
		permissions []string
		wantStatus  int
		wantReached bool
	}{{
		name: "an account holding the admin entitlement", permissions: []string{"module.admin"},
		wantStatus: http.StatusOK, wantReached: true,
	}, {
		// "*" is a grant-all the scope expands to every registered permission,
		// so it satisfies the gate without naming it.
		name: "a superuser", permissions: []string{"*"},
		wantStatus: http.StatusOK, wantReached: true,
	}, {
		name: "an account entitled to something else", permissions: []string{"module.filme"},
		wantStatus: http.StatusForbidden,
	}, {
		name: "an account entitled to nothing", permissions: nil,
		wantStatus: http.StatusForbidden,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := &auth.Account{
				Name: "someone",
				CMS:  auth.CMSProfile{Permissions: tt.permissions},
			}

			reached := false
			h := requireAdmin(signedInAs{account}, func(w http.ResponseWriter, _ *http.Request) error {
				reached = true
				w.WriteHeader(http.StatusOK)
				return nil
			})

			w := httptest.NewRecorder()
			err := h(w, httptest.NewRequest(http.MethodGet, "/api/v1/admin/accounts", nil))

			if reached != tt.wantReached {
				t.Errorf("handler reached = %v, want %v", reached, tt.wantReached)
			}
			if tt.wantStatus == http.StatusOK {
				if err != nil {
					t.Fatalf("an entitled account was refused: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("the request was allowed through")
			}
			if status := errs.StatusOf(err); status != tt.wantStatus {
				t.Errorf("status = %d, want %d", status, tt.wantStatus)
			}
		})
	}
}

// Not being signed in is a different answer from not being allowed: one is
// fixable by signing in, the other is not, and a client that cannot tell them
// apart will offer the wrong remedy.
func TestRequireAdminRefusesAnonymously(t *testing.T) {
	h := requireAdmin(signedInAs{nil}, func(http.ResponseWriter, *http.Request) error {
		t.Error("the handler ran for a request with no account")
		return nil
	})

	err := h(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/v1/admin/accounts", nil))
	if err == nil {
		t.Fatal("an anonymous request was allowed through")
	}
	if status := errs.StatusOf(err); status != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", status)
	}
}

// The file browser is gated the same way, and for the same reason: the
// entitlement is the access decision, not a menu entry. An account with an
// alias configured for some other purpose — delivering films, say — must not
// reach the browser through it.
func TestRequireHidrive(t *testing.T) {
	tests := []struct {
		name        string
		permissions []string
		wantStatus  int
		wantReached bool
	}{{
		name: "an account holding the hidrive entitlement", permissions: []string{"module.hidrive"},
		wantStatus: http.StatusOK, wantReached: true,
	}, {
		name: "a superuser", permissions: []string{"*"},
		wantStatus: http.StatusOK, wantReached: true,
	}, {
		name:        "an account entitled to the media library but not the files",
		permissions: []string{"module.mediathek"}, wantStatus: http.StatusForbidden,
	}, {
		name: "an account entitled to nothing", permissions: nil,
		wantStatus: http.StatusForbidden,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := &auth.Account{
				Name: "someone",
				CMS:  auth.CMSProfile{Permissions: tt.permissions},
			}
			// A drive configured on the account is deliberately present in every
			// case: having one is not what decides this.
			account.HiDrive.Alias = "matt.ihle"

			reached := false
			h := requireHidrive(signedInAs{account}, func(w http.ResponseWriter, _ *http.Request) error {
				reached = true
				w.WriteHeader(http.StatusOK)
				return nil
			})

			err := h(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/v1/drive/meta/", nil))

			if reached != tt.wantReached {
				t.Errorf("handler reached = %v, want %v", reached, tt.wantReached)
			}
			if tt.wantStatus == http.StatusOK {
				if err != nil {
					t.Fatalf("an entitled account was refused: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("the request was allowed through")
			}
			if status := errs.StatusOf(err); status != tt.wantStatus {
				t.Errorf("status = %d, want %d", status, tt.wantStatus)
			}
		})
	}
}

// The preview's gate, which is a capability and not an area.
//
// It was first written as the Content area entitlement — "may you be shown the
// editor" — standing in for "may you read a file while writing content". Those
// are different questions with different holders, and the cases below are the
// two the substitution got wrong: somebody who edits content but has no reason
// to be offered the area, and somebody offered the area who holds nothing else.
func TestRequireMediaRead(t *testing.T) {
	tests := []struct {
		name        string
		permissions []string
		wantStatus  int
		wantReached bool
	}{{
		name: "an account holding the right", permissions: []string{"media.read"},
		wantStatus: http.StatusOK, wantReached: true,
	}, {
		// "*" expands to every registered permission when the scope is built.
		name: "a superuser", permissions: []string{"*"},
		wantStatus: http.StatusOK, wantReached: true,
	}, {
		// What the old gate accepted and this one does not: being shown the
		// editor is not the same as being allowed to read the store.
		name: "an account entitled to the editor area only", permissions: []string{"module.content"},
		wantStatus: http.StatusForbidden,
	}, {
		// And what the old gate refused: authoring without the area in the menu.
		name: "an account that may write content", permissions: []string{"entry.create"},
		wantStatus: http.StatusForbidden,
	}, {
		name: "an account holding nothing", permissions: nil,
		wantStatus: http.StatusForbidden,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := &auth.Account{
				Name: "someone",
				CMS:  auth.CMSProfile{Permissions: tt.permissions},
			}

			reached := false
			h := requirePermission(signedInAs{account}, cmsauth.MediaRead, "nope",
				func(w http.ResponseWriter, _ *http.Request) error {
					reached = true
					w.WriteHeader(http.StatusOK)

					return nil
				})

			w := httptest.NewRecorder()
			err := h(w, httptest.NewRequest(http.MethodGet, "/api/v1/media/stream/a.mp4", nil))
			_ = w

			if reached != tt.wantReached {
				t.Errorf("handler reached = %v, want %v", reached, tt.wantReached)
			}
			if tt.wantStatus == http.StatusOK {
				if err != nil {
					t.Fatalf("an account holding the right was refused: %v", err)
				}

				return
			}
			if err == nil {
				t.Fatal("the request was allowed through")
			}
			if status := errs.StatusOf(err); status != tt.wantStatus {
				t.Errorf("status = %d, want %d", status, tt.wantStatus)
			}
		})
	}
}

// Nobody signed in is a different refusal from somebody without the right, and
// the difference is the one that sent me chasing the wrong thing for an hour:
// 401 means the server could not tell who is asking, 403 means it could.
func TestRequireMediaReadRefusesAnonymouslyWith401(t *testing.T) {
	h := requirePermission(signedInAs{}, cmsauth.MediaRead, "nope",
		func(http.ResponseWriter, *http.Request) error { return nil })

	err := h(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/v1/media/stream/a.mp4", nil))
	if status := errs.StatusOf(err); status != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", status, http.StatusUnauthorized)
	}
}
