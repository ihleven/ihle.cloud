package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/interhome-group/cms/pkg/errs"
)

// What the transport is responsible for, and only that: reading the request,
// writing the answer, and giving a failure its HTTP meaning. The rules those
// failures come from are tested against Admin directly, without a request in
// sight.

// asStatus is the whole translation layer, so it is worth pinning kind by kind.
// A wrong mapping here is not a crash — it is a client told "try again" when it
// should have been told "that is not allowed".
func TestFailuresGetTheirHTTPMeaning(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"an unknown account", ErrNoAccount, http.StatusNotFound},
		{"a request that cannot be carried out", invalid("an account needs an email address"), http.StatusBadRequest},
		{"a change that would lock its author out", conflict("cannot disable your own account"), http.StatusConflict},
		{"an account that already exists", ErrExists, http.StatusConflict},
		{"anything else", errors.New("the database fell over"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errs.StatusOf(asStatus(tt.err)); got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
		})
	}
}

// The sentence a person reads survives the translation. A kind decides the
// number; the message is still the one Admin wrote.
func TestTheMessageSurvives(t *testing.T) {
	err := asStatus(conflict("an administrator cannot disable their own account"))

	if !strings.Contains(err.Error(), "cannot disable their own account") {
		t.Errorf("Error() = %q, want Admin's own sentence", err.Error())
	}
	if strings.Contains(err.Error(), "auth: conflict") {
		t.Errorf("Error() = %q, want the kind kept out of what a client reads", err.Error())
	}
}

// A body that is not JSON is the transport's own failure, answered before Admin
// is ever called.
func TestABrokenBodyIsRefusedBeforeAnythingHappens(t *testing.T) {
	api := NewAdminAPI(nil) // never reached, which is the point

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{not json"))
	err := api.Create(httptest.NewRecorder(), r)

	if err == nil {
		t.Fatal("a malformed body was accepted")
	}
	if got := errs.StatusOf(err); got != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", got)
	}
}

// The one thing only the transport can get wrong about a success: the code.
// Creating answers 201, because something now exists that did not before.
func TestCreateAnswers201(t *testing.T) {
	admin, store := testAdmin(t)
	caller := account(t, store, "admin", "*")

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"wolfgang","email":"w@example.test"}`))
	r = r.WithContext(WithAccount(r.Context(), caller))

	if err := NewAdminAPI(admin).Create(w, r); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", w.Code)
	}

	var created AccountInfo
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if created.Name != "wolfgang" {
		t.Errorf("body = %+v, want the created account", created)
	}
}

// The name comes from the path, which is the transport's other job.
func TestTheNameIsReadFromThePath(t *testing.T) {
	admin, store := testAdmin(t)
	caller := account(t, store, "admin", "*")
	account(t, store, "wolfgang")

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(WithAccount(r.Context(), caller))
	r.SetPathValue("name", "wolfgang")

	if err := NewAdminAPI(admin).Get(w, r); err != nil {
		t.Fatalf("Get: %v", err)
	}

	var got AccountInfo
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if got.Name != "wolfgang" {
		t.Errorf("answered with %q, want the account named in the path", got.Name)
	}
}
