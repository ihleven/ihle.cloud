package srvr

import (
	"net/http"
	"slices"
	"strings"
)

// RecordingMux is a ServeMux that remembers what was registered on it.
//
// The route table is where this application declares its endpoints, and it is
// not the only place endpoints appear: the CMS's API is mounted whole, and
// huma registers its own operations, its OpenAPI document and its
// documentation UI directly on whatever mux it is handed. Those are real,
// reachable addresses. A listing that showed the table and called itself the
// service's surface would be quietly wrong, and a check that the addresses can
// coexist would be checking only some of them.
//
// Recording every registration makes the answer true by construction rather
// than by somebody remembering to write each route down twice.
//
// It satisfies the two methods huma's net/http adapter asks for, which is why
// the CMS's router can mount on it.
//
// Borrowed from api-gateway, where the same problem produced the same answer.
type RecordingMux struct {
	mux      *http.ServeMux
	patterns []string
}

// NewRecordingMux returns an empty mux.
func NewRecordingMux() *RecordingMux {
	return &RecordingMux{mux: http.NewServeMux()}
}

// canonical is a pattern with its whitespace settled.
//
// Route tables are written to line up — "GET     /entries" here, "  GET  /films"
// there — and ServeMux reads the method and the path across any run of spaces,
// so two spellings of one address both work and neither compares equal to the
// other. Collapsing them means what is recorded is what a reader would write,
// and that two tables can be compared without either knowing how the other was
// indented.
func canonical(pattern string) string {
	return strings.Join(strings.Fields(pattern), " ")
}

// Handle registers a handler and remembers the pattern.
func (m *RecordingMux) Handle(pattern string, h http.Handler) {
	pattern = canonical(pattern)
	m.patterns = append(m.patterns, pattern)
	m.mux.Handle(pattern, h)
}

// HandleFunc is the same, and is the method huma's adapter uses.
func (m *RecordingMux) HandleFunc(pattern string, h func(http.ResponseWriter, *http.Request)) {
	pattern = canonical(pattern)
	m.patterns = append(m.patterns, pattern)
	m.mux.HandleFunc(pattern, h)
}

func (m *RecordingMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}

// Patterns returns every pattern registered on this mux, in the order they
// were registered.
func (m *RecordingMux) Patterns() []string { return slices.Clone(m.patterns) }

// Mounted returns everything this application serves: what was registered here,
// and what was registered on a mux mounted under one of these patterns.
//
// A mounted mux is a subtree — the CMS's API is mounted at "/content-api/v1/"
// — so its own patterns are absolute and are simply added. The prefix itself
// stays in the list: something is registered there, and it is the thing that
// stands in front of everything below it.
func (m *RecordingMux) Mounted(others ...*RecordingMux) []string {
	all := m.Patterns()
	for _, other := range others {
		if other == nil {
			continue
		}
		for _, p := range other.Patterns() {
			if !slices.Contains(all, p) {
				all = append(all, p)
			}
		}
	}
	slices.Sort(all)

	return all
}
