package srvr

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/interhome-group/cms/pkg/errs"
)

// carried is an error that answers for itself: it names a status and expects to
// be the body. A store's reply is the real one — this is its shape.
type carried struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"msg"`
}

func (e carried) Error() string   { return e.Message }
func (e carried) StatusCode() int { return e.Status }

func served(t *testing.T, h func(http.ResponseWriter, *http.Request) error) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()
	errmw(&Api{}, h)(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	return rec
}

// An error that carries a status is answered with it, and is itself the body.
//
// This is how a refusal from the storage provider reaches the browser intact —
// deliberately, so that what it said is not flattened into "something went
// wrong". Before, the branch was chosen by naming the storage client's own type
// here; now the error opts in by carrying a status, and this package knows no
// concrete error at all.
func TestAnErrorThatCarriesItsStatusIsTheAnswer(t *testing.T) {
	rec := served(t, func(http.ResponseWriter, *http.Request) error {
		return carried{Status: http.StatusForbidden, Code: "403", Message: "not a regular file"}
	})

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content type = %q, want application/json", ct)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("the body is not an object: %v: %s", err, rec.Body.String())
	}
	if body["msg"] != "not a regular file" {
		t.Errorf("msg = %q, want what the error said", body["msg"])
	}
}

// Every other error is answered with the status it carries and its message as
// the body, which is a JSON string rather than an object — the difference the
// branch above exists for.
func TestAnOrdinaryErrorIsAnsweredWithItsMessage(t *testing.T) {
	rec := served(t, func(http.ResponseWriter, *http.Request) error {
		return errs.New("nope", errs.HTTPStatus(http.StatusUnauthorized))
	})

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	var msg string
	if err := json.Unmarshal(rec.Body.Bytes(), &msg); err != nil {
		t.Fatalf("the body is not a string: %v: %s", err, rec.Body.String())
	}
	if !strings.Contains(msg, "nope") {
		t.Errorf("body = %q, want the message in it", msg)
	}
}

// An error saying nothing about status is a fault here rather than a refusal.
func TestAnErrorWithNoStatusIsFiveHundred(t *testing.T) {
	rec := served(t, func(http.ResponseWriter, *http.Request) error {
		return errs.New("something came apart")
	})

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}

// A handler that returns nil has already written, and nothing here touches what
// it wrote — the whole point of returning the error instead of writing it is
// that the two cases never overlap.
func TestASuccessfulHandlerIsLeftAlone(t *testing.T) {
	rec := served(t, func(w http.ResponseWriter, _ *http.Request) error {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("mine"))

		return nil
	})

	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %d, want the handler's own", rec.Code)
	}
	if rec.Body.String() != "mine" {
		t.Errorf("body = %q, want the handler's own", rec.Body.String())
	}
}

// The request log is off unless asked for. It used to be unconditional, which
// is pleasant in a terminal and noise in a log file.
func TestTheRequestLogIsQuietByDefault(t *testing.T) {
	if (&Api{}).logRequests {
		t.Error("a freshly built Api logs every request")
	}
}
