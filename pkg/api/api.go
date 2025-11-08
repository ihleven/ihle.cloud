package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"bitbucket.org/hotelplan/webcc-content/cms/pkg/errors"
	"github.com/fatih/color"
	"github.com/ihleven/ihlvn/pkg/auth"
	"github.com/ihleven/ihlvn/pkg/hi"
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

func (a *Api) ListenAndServe(port int, origins []string) error {

	if len(origins) == 0 {
		origins = []string{"http://localhost:3000"}
	}

	corsMw, err := jubcors.NewMiddleware(jubcors.Config{
		Origins:        origins,
		Methods:        []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		RequestHeaders: []string{"Authorization", "Content-Type"},
		Credentialed:   true,
	})
	if err != nil {
		log.Fatal(err)
	}
	corsMw.SetDebug(true) // optional: turn debug mode on

	// handler := cors.New(cors.Options{
	// 	AllowedOrigins:   []string{"http://localhost:3000"},
	// 	AllowCredentials: true,
	// 	Debug:            true, // Enable Debugging for testing, consider disabling in production
	// }).Handler(a.routr)

	return http.ListenAndServe(fmt.Sprintf(":%d", port), corsMw.Wrap(a.routr))
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
			fmt.Println("errmw =>", errors.Code(err), errors.Cause(err), err)

			statusCode = errors.Code(err)
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
func GetAccessFromClaims(claims *auth.Claims) (*auth.Account, string, error) {
	fmt.Println("GetAccessFromClaims")
	account := auth.AuthenticatorPKG.Account(claims.Subject)
	fmt.Println("GetAccessFromClaims", account)
	token, err := auth.AuthenticatorPKG.GetToken(account)
	fmt.Println("GetAccessFromClaims", token, err)

	fmt.Println("GetAccessFromClaims")
	return account, token, err
}
func getPermissionForPath(path string) string {
	fmt.Println("getPermissionForPath", path)
	if strings.HasPrefix(path, "public/djvet") {
		return "djvet"
	} else if strings.HasPrefix(path, "public/mediathek") {
		return "mediathek"
	}
	return "other"
}
