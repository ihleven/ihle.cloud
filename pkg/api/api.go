package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/ihleven/ihlvn/pkg/hi"
	"github.com/interhome-group/cms/pkg/errs"
	jubcors "github.com/jub0bs/cors"
	"github.com/uptrace/bunrouter"
)

// https://ieftimov.com/posts/make-resilient-golang-net-http-servers-using-timeouts-deadlines-context-cancellation/
// https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/

func New(options ...func(*Api)) *Api {

	mux := http.NewServeMux()

	// mux.HandleFunc("/hi/auth", TokenAuthHandler)
	// mux.HandleFunc("/hi/auth/signin", auth.InfoMux)
	// mux.HandleFunc("/hi/auth/authcode", auth.AuthorizeCallbackMux)
	// mux.HandleFunc("/hi/auth/info", auth.InfoMux)

	// // go reverseProxy auf hi.GetURL
	// // mux.HandleFunc("GET /hi/media/proxy/{path...}", hi.ServeReverseProxyMux) // used for serving local video on opalstack

	// // rohdaten
	// mux.HandleFunc("GET /hi/media/{path...}", FileHandler)

	// // thumbnails
	// // mux.HandleFunc("GET /hi/media/thumbs/{path...}", ThumbHandler)

	// // metadaten zu dirs and files
	// mux.HandleFunc("GET /hi/meta/{path...}", errmw(MetaHandler))

	// // mux.HandleFunc("GET /hi/tags/{path...}", hi.TagsHandler)

	a := &Api{routr: mux}

	for _, option := range options {
		option(a)
	}

	return a
}

type Api struct {
	routr *http.ServeMux
}

func WithRoute(pattern string, h func(http.ResponseWriter, *http.Request) error) func(*Api) {
	return func(api *Api) {
		api.routr.HandleFunc(strings.TrimSpace(pattern), errmw(h))
	}
}

func Handle(pattern string, h func(http.ResponseWriter, *http.Request)) func(*Api) {
	return func(api *Api) {
		api.routr.HandleFunc(strings.TrimSpace(pattern), h)
	}
}

func Handler(pattern string, h http.Handler) func(*Api) {
	return func(api *Api) {
		api.routr.Handle(strings.TrimSpace(pattern), h)
	}
}

// ListenAndServe serves the API, granting cross-site access to origins.
//
// No origins is the normal case: this app serves its own frontend, so the
// browser is always talking to the origin it loaded the page from and CORS does
// not enter into it. A grant is only needed for a separately hosted client, and
// then it has to be named — credentialed CORS cannot use a wildcard, and a
// default of "some development port" would be a grant nobody asked for.
func (a *Api) ListenAndServe(port int, origins []string, debug bool) error {

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

func JSON(w http.ResponseWriter, value interface{}) error {
	if value == nil {
		return nil
	}

	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	if err := enc.Encode(value); err != nil {
		return err
	}

	return nil
}

func RespondJSON(w http.ResponseWriter, value interface{}) error {
	if value == nil {
		return nil
	}

	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	if err := enc.Encode(value); err != nil {
		return err
	}

	return nil
}

func errmw(next func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Call the next handler on the chain to get the error.
		err := next(w, r)

		statusCode := 200

		switch err := err.(type) {
		case nil:
			// no error
		case hi.Error: // already a HTTPError
			statusCode = err.HTTPStatus
			w.WriteHeader(err.HTTPStatus)
			_ = bunrouter.JSON(w, err)
		default:
			fmt.Println("errmw =>", errs.StatusOf(err), errs.Cause(err), err)

			statusCode = errs.StatusOf(err)
			if statusCode == 0 {
				statusCode = 500
			}
			// httpErr := (err)
			w.WriteHeader(statusCode)
			_ = bunrouter.JSON(w, err.Error())
		}
		// if true {

		var args []interface{}
		args = append(args,
			color.New(color.BgBlue, color.FgHiWhite).Sprintf("[cloud-api]"),
			time.Now().Format(" 15:04:05.000 "),
			color.New(color.BgBlue, color.FgHiWhite).Sprintf(" %-7s ", r.Method),
			r.URL.String(),
			statusColor(statusCode).Sprintf(" %d ", statusCode),
			// color.New(color.BgBlue, color.FgHiWhite).Sprintf("[%d bytes]", rw.Count()),
			fmt.Sprintf(" %10s ", time.Since(start).Round(time.Microsecond)),
		)
		if err != nil {
			args = append(args, fmt.Sprintf("ERROR %T", err), color.New(color.BgRed, color.FgHiWhite).Sprintf("[%s]", err.Error()))
		}
		fmt.Println(args...)
		fmt.Println()
		// }
	}
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

// SplitPath is like ShiftPath but without the leading slash in the result
func SplitPath(p string) (head, tail string) {

	p = strings.TrimLeft(p, "/")

	if i := strings.Index(p, "/"); i >= 0 {

		return p[0:i], p[i+1:]
	}
	return p, ""
}
