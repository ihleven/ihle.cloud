package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/ihleven/ihle.cloud/app/db"
	"github.com/ihleven/ihle.cloud/backend/spa"
	"github.com/moby/moby/pkg/pidfile"

	"bitbucket.org/hotelplan/webcc-content/cms/content"
	"bitbucket.org/hotelplan/webcc-content/cms/mgmt/search"

	"github.com/alexflint/go-arg"
	"github.com/ihleven/ihle.cloud/app/art"
	"github.com/ihleven/ihle.cloud/app/cmsuc"
	"github.com/ihleven/ihle.cloud/app/familie"
	"github.com/ihleven/ihle.cloud/backend/api"
	"github.com/ihleven/ihle.cloud/backend/auth"
	"github.com/ihleven/ihle.cloud/backend/hi"

	_ "github.com/joho/godotenv/autoload"
)

type Flags struct {
	Port                int    `arg:"-p,--port,      env"                  default:"8000"  help:"Port number"`
	HidriveClientID     string `arg:"                env:CLIENT_ID"`
	HidriveClientSecret string `arg:"                env:CLIENT_SECRET"`
	RepoContent         string `arg:"                env:REPO_CONTENT" `
	DataDir             string `arg:"--data-dir,     env:DATA_DIR"         default:"data"`
	SearchLevel         string `arg:"--search-level, env:SEARCH_LEVEL"     default:"basic" help:"Search level: off,basic,fulltext,extended"`
	SearchDir           string `arg:"--search-dir,   env:SEARCH_DIR"       default:"bleve" help:"Dirname of on disk search index, leave empty for in mem index"`

	JWTIssuer      string        `arg:"env:JWT_ISSUER"                              default:"ihle.cloud"`
	JWTSecretKey   string        `arg:"env:JWT_SECRET_KEY" `
	JWTDuration    int           `arg:"env:JWT_DURATION"                            default:"3600" help:"Duration of JWT token in seconds"`
	CookieName     string        `arg:"env:COOKIE_NAME"                        default:"jwt" help:"Name for auth cookie"`
	CookieSameSite http.SameSite `arg:"env:COOKIE_SAME_SITE"                        default:"2" help:"SameSite attribute of auth cookie: Default (1) Lax (2), Strict (3), None (4)"`
	DbConn         string        `arg:"env:DB_CONN"                                 default:"postgres://localhost:5432/authdb"`
	Pidfile        string        `arg:"--pid-file,   env:PIDFILE"    `
	SPAPath        string        `arg:"--spa-path,   env:SPA_PATH"    default:"ui/.output/public"`
}

func (f Flags) cmsConfig() cmsuc.Config {
	conf := cmsuc.Config{
		RepoContent: flags.RepoContent,
		DataDir:     flags.DataDir,
		Search:      search.Config{Level: search.ParseLevel(flags.SearchLevel), DataDirPath: flags.DataDir, IndexName: flags.SearchDir},
	}
	return conf
}

// ihle-api
// cloud-ihleven
// ihleven-api
// ihleven.de/hi/media
// api.ihle.cloud/api

var flags Flags

// var tokenmap map[string]hi.Token = map[string]hi.Token{}
var route = api.WithRoute

func handle_pidfile(p string) {
	if p == "" {
		return
	}
	err := pidfile.Write(flags.Pidfile, os.Getpid())
	if err != nil {
		fmt.Println(err)
		os.Exit(99)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func(f string) {
		sig := <-c
		fmt.Println("\nsignal:", sig)
		err := os.Remove(f)
		if err != nil {
			fmt.Println(err)
		}
		os.Exit(0)
	}(flags.Pidfile)
}

func main() {
	arg.MustParse(&flags)

	handle_pidfile(flags.Pidfile)

	db, err := db.New(flags.DbConn)
	if err != nil {
		log.Fatal("db.New:", err)
	}
	defer db.Close()

	content.Register(familie.Person{})
	content.Register(familie.Reise{})
	content.Register(art.Ausstellung{})
	content.Register(art.Work{})

	// conf := cms.Config{
	// 	RepoContent: flags.RepoContent,
	// 	DataDir:     flags.DataDir,
	// 	Search:      search.Config{Level: search.ParseLevel(flags.SearchLevel), DataDirPath: flags.DataDir, IndexName: flags.SearchDir},
	// }

	auth.New(flags.JWTIssuer, flags.JWTSecretKey, flags.JWTDuration, flags.CookieName, flags.CookieSameSite, db, db)

	cmsapi, err := cmsuc.NewCMSApi(flags.cmsConfig())
	if err != nil {
		log.Fatal(err)
	}

	fapi := familie.NewApi(cmsapi.Repo, cmsapi.Engine)

	srvr := api.New(

		api.Handler("/", spa.Serve(flags.SPAPath)),

		// route(" POST /tokenauth                ", tokenauth), // soll token liefern für spezielle Funktionalität
		route("      /apihle/auth/authorize    ", authorize),
		route("      /hi/auth/authcode         ", callback(db)),

		route("      /auth/signin        ", auth.Signin),
		route("      /auth/login         ", auth.Login),
		route("      /auth/token         ", auth.TokenAuthHandler),
		route("      /auth/logout        ", auth.Logout),
		route("      /auth/session       ", auth.Session),

		route("  GET /hi/meta/{path...}        ", hi.MetaHandlerMux),
		route("  GET /hi/media/{path...}       ", hi.FileHandlerMux),
		route("  GET /hi/media/thumbs          ", hi.ThumbHandlerMux),
		route("  GET /hi/media/proxy/{path...} ", hi.ServeReverseProxyMux),
		// higrp.GET("/tags/*path", hi.TagsHandler)
		route("  GET /media/videos/{path...}   ", serveContentWithPrefix("videos")), // used for serving local video on opalstack

		route("  GET /api/v1/entries/{path...} ", cmsapi.EntryDetails),
		route("  PUT /api/v1/entries/{path...} ", cmsapi.EntryUpdate),

		route("GET  /api/v1/personen/{person}", fapi.PersonHandler),
		route("GET  /api/v1/reisen/{key}", fapi.ReiseHandler),
		route("GET  /api/v1/search", cmsuc.SearchEntries(cmsapi.CMS)),
	)

	srvr.ListenAndServe(flags.Port)
}

func authorize(w http.ResponseWriter, r *http.Request) error {

	url := fmt.Sprintf("https://my.hidrive.com/client/authorize?client_id=%s=&response_type=code&scope=admin,rw&state=%s&redirect_uri=http://localhost:8000/hi/auth/authcode", flags.HidriveClientID, r.URL.Query().Get("state"))
	http.Redirect(w, r, url, http.StatusSeeOther)
	return nil
}

func callback(db *db.DB) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {

		code := r.URL.Query().Get("code")
		fmt.Println("code:", code, flags.HidriveClientID, flags.HidriveClientSecret)

		token, err := hi.RefreshTokenWithAuthCode(flags.HidriveClientID, flags.HidriveClientSecret, code)
		if err != nil {
			http.Error(w, "callback: "+err.Error(), 401)
			return err
		}
		fmt.Println("token:", token)
		// tokenmap[token.UserID] = *token
		err = db.StoreToken(token)
		if err != nil {
			return err
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "hitoken",
			Value:    token.AccessToken,
			Path:     "/",
			MaxAge:   token.ExpiresIn,
			HttpOnly: true,
			// Secure:   true,
			// SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, r.URL.Query().Get("state"), http.StatusSeeOther)
		return nil
	}
}

// func RequestToken(id, secret, code string) (*hi.Token, error) {
// 	fmt.Println("id, secret, code:", id, secret, code)

// 	params := url.Values{
// 		"client_id":     []string{id},
// 		"client_secret": []string{secret},
// 		"grant_type":    []string{"authorization_code"},
// 		"code":          []string{code},
// 	}

// 	request, err := http.NewRequest("POST", "https://my.hidrive.com/oauth2/token?"+params.Encode(), nil)
// 	if err != nil {
// 		return nil, hi.Error{HTTPStatus: 500, Message: "Couldn't create request: " + err.Error()}
// 	}

// 	resp, err := http.DefaultClient.Do(request)
// 	if err != nil {
// 		if os.IsTimeout(err) {
// 			// HTTP 504 Gateway Timeout
// 			return nil, hi.Error{HTTPStatus: 504, Message: "timeout exceeded: " + err.Error()}
// 		}

// 		// HTTP 502 Bad Gateway
// 		return nil, hi.Error{HTTPStatus: 502, Message: "HTTP client couldn't Do request: " + err.Error()}
// 	}

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		fmt.Println("read err:", err)
// 		return nil, hi.Error{Message: "read error: " + err.Error()}
// 	}

// 	fmt.Println(string(body))
// 	var t hi.Token
// 	err = json.Unmarshal(body, &t)
// 	if err != nil {
// 		fmt.Println("err:", err)
// 		return nil, hi.Error{Message: err.Error()}
// 	}
// 	fmt.Println("token:", t)
// 	return &t, nil
// }

/////////////// code aus altem main mit bunrouter ////////////////

// videos aus lokalen unterverzeichnis videos
func serveContentWithPrefix(prefix string) func(http.ResponseWriter, *http.Request) error {

	return func(w http.ResponseWriter, r *http.Request) error {

		filename := path.Join(prefix, r.PathValue("path"))

		fd, err := os.Open(filename)
		if err != nil {
			return err
		}
		stat, err := fd.Stat()
		if err != nil {
			return err
		}

		// w.Header().Set("Access-Control-Allow-Origin", "*")
		// w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")

		http.ServeContent(w, r, fd.Name(), stat.ModTime(), fd)

		return nil
	}
}

func testNewHiFS(accesstoken string) {

	fs := hi.NewFS(accesstoken)
	f, err := fs.Open("/public/blog/README.md")
	fmt.Println("file", f, err)
	stat, err := f.Stat()
	fmt.Println("stat", stat.Size(), err)
	var bytes []byte
	n, err := f.Read(bytes)
	fmt.Printf("%d: %s %s", n, bytes, err)

}
