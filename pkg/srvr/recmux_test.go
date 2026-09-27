package srvr

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

// It is a mux first: recording is what it does besides serving, not instead.
func TestARecordedRouteStillAnswers(t *testing.T) {
	m := NewRecordingMux()
	m.HandleFunc("GET /thing", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/thing", nil))

	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %d, want the handler's own", rec.Code)
	}
}

// Both ways in are recorded, because huma uses one and the table uses the other.
func TestBothRegistrationsAreRecorded(t *testing.T) {
	m := NewRecordingMux()
	m.Handle("GET /handler", http.NotFoundHandler())
	m.HandleFunc("POST /func", func(http.ResponseWriter, *http.Request) {})

	want := []string{"GET /handler", "POST /func"}
	if got := m.Patterns(); !slices.Equal(got, want) {
		t.Errorf("patterns = %v, want %v", got, want)
	}
}

// Tables are written to line up, and the alignment is not part of the address.
// Two tables indented differently have to compare equal, or a listing can only
// be read by whoever wrote it.
func TestAlignmentIsNotPartOfTheAddress(t *testing.T) {
	m := NewRecordingMux()
	m.HandleFunc("  GET     /spaced/{id}   ", func(http.ResponseWriter, *http.Request) {})

	if got := m.Patterns(); !slices.Equal(got, []string{"GET /spaced/{id}"}) {
		t.Errorf("patterns = %q, want the address without its padding", got)
	}

	// And it still matches, which is the part that would be expensive to get
	// wrong: the pattern handed to ServeMux is the canonical one.
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/spaced/7", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want the route to match", rec.Code)
	}
}

// What a mounted mux registered is part of what the application serves, and the
// mount point stays: something is registered there, and it stands in front of
// everything below it.
func TestMountedAddsWhatAnotherMuxRecorded(t *testing.T) {
	inner := NewRecordingMux()
	inner.HandleFunc("GET /content-api/v1/entries", func(http.ResponseWriter, *http.Request) {})

	outer := NewRecordingMux()
	outer.Handle("/content-api/v1/", inner)
	outer.HandleFunc("GET /api/v1/films/{id}", func(http.ResponseWriter, *http.Request) {})

	whole := outer.Mounted(inner)
	for _, want := range []string{
		"/content-api/v1/",
		"GET /content-api/v1/entries",
		"GET /api/v1/films/{id}",
	} {
		if !slices.Contains(whole, want) {
			t.Errorf("%q is served but not listed: %v", want, whole)
		}
	}
}

// A nil mux is nothing mounted, not a crash: the CMS API is absent from a
// deployment that does not build it.
func TestMountedToleratesNothingMounted(t *testing.T) {
	m := NewRecordingMux()
	m.HandleFunc("GET /x", func(http.ResponseWriter, *http.Request) {})

	if got := m.Mounted(nil); !slices.Equal(got, []string{"GET /x"}) {
		t.Errorf("Mounted(nil) = %v, want what this mux has", got)
	}
}
