// Package srvr is this application's HTTP spine: the router, the server, and
// the one convention every handler here is written in.
//
// Named as its siblings in the other services are — the CMS and mytool-api both
// call this role srvr — rather than "api", which it shared with a package that
// does the opposite job: the CMS's pkg/api describes an API, and this one
// serves one. Two of them in one import block needed an alias to compile, which
// is a fair sign the name was wrong.
//
// That convention is the signature — func(w, r) error. A handler returns what
// went wrong instead of writing it, and errmw below turns the error into a
// status, a body and a log line. It is why nothing in this app calls
// http.Error: refusing a request is `return errs.New(..., HTTPStatus(403))`,
// and where that status comes from is settled in exactly one place.
//
// An error says its own status by carrying a StatusCode or HTTPStatus method,
// which is what errs.StatusOf looks for. Nothing here knows any concrete error
// type; an error that wants to be the response body says so by implementing
// coded.
package srvr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fatih/color"
	"github.com/interhome-group/cms/pkg/errs"
	jubcors "github.com/jub0bs/cors"
)

// https://ieftimov.com/posts/make-resilient-golang-net-http-servers-using-timeouts-deadlines-context-cancellation/
// https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/

func New(options ...func(*Api)) *Api {

	a := &Api{routr: NewRecordingMux()}

	for _, option := range options {
		option(a)
	}

	return a
}

type Api struct {
	routr *RecordingMux
	// logRequests writes a line per request. Set by ListenAndServe from the
	// same flag that puts CORS into debug, because both answer "I am watching
	// this run" rather than "this is how it should behave".
	logRequests bool
}

func WithRoute(pattern string, h func(http.ResponseWriter, *http.Request) error) func(*Api) {
	return func(api *Api) {
		api.routr.HandleFunc(pattern, errmw(api, h))
	}
}

func Handler(pattern string, h http.Handler) func(*Api) {
	return func(api *Api) {
		api.routr.Handle(pattern, h)
	}
}

// ListenAndServe serves the API, granting cross-site access to origins.
//
// No origins is the normal case: this app serves its own frontend, so the
// browser is always talking to the origin it loaded the page from and CORS does
// not enter into it. A grant is only needed for a separately hosted client, and
// then it has to be named — credentialed CORS cannot use a wildcard, and a
// default of "some development port" would be a grant nobody asked for.
// Patterns is everything registered on this Api, in registration order.
//
// Where a mux has been mounted under one of them — the CMS's API is — ask that
// mux for its own; see RecordingMux.Mounted.
func (a *Api) Patterns() []string { return a.routr.Patterns() }

func (a *Api) ListenAndServe(port int, origins []string, debug bool) error {

	a.logRequests = debug

	handler := http.Handler(a.routr)

	if len(origins) > 0 {
		corsMw, err := jubcors.NewMiddleware(jubcors.Config{
			Origins:        origins,
			Methods:        []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
			RequestHeaders: []string{"Authorization", "Content-Type"},
			Credentialed:   true,
		})
		if err != nil {
			return fmt.Errorf("configuring CORS for %v: %w", origins, err)
		}
		corsMw.SetDebug(debug)
		handler = corsMw.Wrap(a.routr)
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", port), handler)
}

// coded is an error that carries its own status and wants to be the body.
//
// An interface rather than a named type, so this package does not have to know
// what any particular error is. It used to name one — the HiDrive client's —
// which put a storage client's type into the HTTP spine and, less obviously,
// forwarded that store's statuses to the browser verbatim: a 401 from HiDrive
// arrived looking exactly like "you are not signed in here".
//
// That forwarding is deliberate and stays. What changes is that the error opts
// into it by implementing this, rather than this package reaching for it.
type coded interface {
	error
	StatusCode() int
}

// errmw turns a returned error into a response.
//
// The whole of what makes func(w, r) error work: a status from the error, a
// body, and a line in the log. A handler that has already written is unaffected
// — it returns nil and nothing here touches the response.
func errmw(a *Api, next func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		err := next(w, r)

		statusCode := 200

		switch err := err.(type) {
		case nil:
			// Nothing went wrong, and the handler has written whatever it wrote.
		case coded:
			// The error is the answer: its status, and itself as the body.
			statusCode = err.StatusCode()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			writeJSON(w, err)
		default:
			statusCode = errs.StatusOf(err)
			if statusCode == 0 {
				statusCode = 500
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			writeJSON(w, err.Error())
		}

		if a.logRequests {
			logRequest(r, statusCode, time.Since(start), err)
		}
	}
}

// writeJSON is the response body, and the only thing this package encodes.
//
// The header is not set here: by the time this is called the status has been
// written, which commits the headers — the content type is settled before that,
// where the status is decided.
func writeJSON(w http.ResponseWriter, value any) {
	_ = json.NewEncoder(w).Encode(value)
}

// logRequest writes the one line a request leaves behind.
//
// Coloured, which is worth it in a terminal and noise in a file — hence off
// unless asked for. It used to be unconditional, along with a bare Println of
// the error that predated it.
func logRequest(r *http.Request, status int, took time.Duration, err error) {
	args := []any{
		color.New(color.BgBlue, color.FgHiWhite).Sprintf("[cloud-api]"),
		time.Now().Format(" 15:04:05.000 "),
		color.New(color.BgBlue, color.FgHiWhite).Sprintf(" %-7s ", r.Method),
		r.URL.String(),
		statusColor(status).Sprintf(" %d ", status),
		fmt.Sprintf(" %10s ", took.Round(time.Microsecond)),
	}
	if err != nil {
		args = append(args,
			fmt.Sprintf("ERROR %T", err),
			color.New(color.BgRed, color.FgHiWhite).Sprintf("[%s]", err.Error()))
	}
	fmt.Println(args...)
	fmt.Println()
}

func statusColor(code int) *color.Color {
	switch {
	case code >= 200 && code < 300:
		return color.New(color.BgGreen, color.FgHiWhite)
	case code >= 300 && code < 400:
		return color.New(color.BgWhite, color.FgHiBlack)
	case code >= 400 && code < 500:
		return color.New(color.BgYellow, color.FgHiBlack)
	default:
		return color.New(color.BgRed, color.FgHiWhite)
	}
}
