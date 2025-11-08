package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/ihleven/ihlvn/app/art/importart"
	"github.com/ihleven/ihlvn/app/db"
	"github.com/ihleven/ihlvn/app/super8"
	"github.com/ihleven/ihlvn/pkg/cmd"
	"github.com/ihleven/ihlvn/pkg/mail"
	"github.com/ihleven/ihlvn/pkg/spa"
	"github.com/moby/moby/pkg/pidfile"

	"bitbucket.org/hotelplan/webcc-content/cms/content"
	"bitbucket.org/hotelplan/webcc-content/cms/mgmt/search"

	"github.com/alexflint/go-arg"
	"github.com/ihleven/ihlvn/app/art"
	"github.com/ihleven/ihlvn/app/cmsuc"
	"github.com/ihleven/ihlvn/app/familie"
	"github.com/ihleven/ihlvn/pkg/api"
	"github.com/ihleven/ihlvn/pkg/auth"
	"github.com/ihleven/ihlvn/pkg/hi"

	_ "github.com/joho/godotenv/autoload"
)

var (
	BUILD_DIR       string
	BUILD_TIME      string
	BUILD_OUTPUT    string
	GIT_DESCRIPTION string
)

type Flags struct {
	Port                int    `arg:"-p,--port,               env"                  default:"8000"  help:"Port number"`
	HidriveClientID     string `arg:"--hidrive-client-id,     env:CLIENT_ID"        placeholder:"ID"`
	HidriveClientSecret string `arg:"--hidrive-client-secret, env:CLIENT_SECRET"    placeholder:"SECRET"`
	RepoContent         string `arg:"                         env:REPO_CONTENT"     placeholder:"URL" `
	DataDir             string `arg:"--data-dir,     env:DATA_DIR"         default:"data" placeholder:"DIR"`
	SearchLevel         string `arg:"--search-level, env:SEARCH_LEVEL"     default:"basic" help:"Search level: off,basic,fulltext,extended" placeholder:"LEVEL"`
	SearchDir           string `arg:"--search-dir,   env:SEARCH_DIR"       default:"bleve" help:"Dirname of on disk search index, leave empty for in mem index" placeholder:"DIR"`

	JWTIssuer      string        `arg:"--jwt-issuer,env:JWT_ISSUER"                         default:"ihle.cloud" placeholder:"ISSUER"`
	JWTSecretKey   string        `arg:"--jwt-secret,env:JWT_SECRET_KEY"                                          placeholder:"KEY"`
	JWTDuration    int           `arg:"--jwt-duration,env:JWT_DURATION"        default:"36000" help:"Duration of JWT token in seconds"`
	CookieName     string        `arg:"env:COOKIE_NAME"                        default:"jwt"  help:"Name for auth cookie"`
	CookieSameSite http.SameSite `arg:"env:COOKIE_SAME_SITE"                   default:"2"    help:"SameSite attribute of auth cookie: Default (1) Lax (2), Strict (3), None (4)"`
	DbConn         string        `arg:"env:DB_CONN"                            default:"postgres://localhost:5432/authdb" placeholder:"CONN"`
	Pidfile        string        `arg:"--pid-file,   env:PIDFILE"              default:"" placeholder:"FILENAME"`
	SPAPath        string        `arg:"--spa-path,   env:SPA_PATH"             default:"ui/.output/public" placeholder:"DIR" help:"path to nuxt spa"`
	// 	Debug        bool   `arg:"-d,--debug,env"      default:"false"          help:"Enable debug mode"`
	// 	Pretty       bool   `arg:"--pretty,env:LOG_PRETTY"                      help:"Enable pretty logging"`
	// 	Verbose      bool   `arg:"-v,--verbose,env"                             help:"Enable verbose mode"`

}

type RootCmd struct {
	// *ServerCmd `arg:"subcommand:server"`
	*importart.ImportCmd `arg:"subcommand:import"`
	*mail.MailCmd        `arg:"subcommand:mail"`

	// root cmd flags
	Port  int  `arg:"-p,--port,env:PORT" default:"8000"   help:"Port numbe"` // default:"10815"
	Clone bool `arg:"--clone"`
}

func (RootCmd) Version() string {
	return cmd.Info.Version.String()
}

func main() {

	// set cmd.Info
	cmd.SetLdflags(BUILD_DIR, BUILD_TIME, BUILD_OUTPUT, GIT_DESCRIPTION)

	content.Register(super8.Super8{})
	// yaml.RegisterCustomUnmarshaler[content.Entry](content.UnmarshalYAMLEntry)
	content.Register(familie.Person{})
	content.Register(familie.Reise{})
	search.RegisterMappingAdapter(familie.AdaptMapping)
	search.RegisterMappingAdapter(familie.AdaptMappingReise)
	content.Register(art.Ausstellung{})
	content.Register(art.Work{})
	// content.Register(ctype.Page{})

	var err error
	var root RootCmd
	var flags Flags

	p := arg.MustParse(&root, &flags)

	handle_pidfile(flags.Pidfile)

	switch subcmd := p.Subcommand().(type) {
	case *importart.ImportCmd:
		err = subcmd.Run()
	case *mail.MailCmd:
		err = subcmd.Run()
	default:
		err = root.RunServer(flags)
	}

	if err != nil {
		fmt.Println("FEHLER: ", err)
		os.Exit(1)
	}
}

func (flags Flags) cmsConfig() cmsuc.Config {
	return cmsuc.Config{
		RepoContent: flags.RepoContent,
		DataDir:     flags.DataDir,
		Search:      search.Config{Level: search.ParseLevel(flags.SearchLevel), DataDirPath: flags.DataDir, IndexName: flags.SearchDir},
	}
}

func (cmd *RootCmd) RunServer(flags Flags) error {

	pg, err := db.New(flags.DbConn)
	if err != nil {
		log.Fatal("db.New:", err)
	}
	defer pg.Close()

	auth.New(flags.JWTIssuer, flags.JWTSecretKey, flags.JWTDuration, flags.CookieName, flags.CookieSameSite, pg, pg)

	cmsapi, err := cmsuc.NewCMSApi(flags.cmsConfig())
	if err != nil {
		log.Fatal(err)
	}

	fapi := familie.NewApi(cmsapi.Repo, cmsapi.Engine)

	var route = api.WithRoute

	srvr := api.New(

		api.Handler("/", spa.Serve(flags.SPAPath)),

		// route(" POST /tokenauth                ", tokenauth), // soll token liefern für spezielle Funktionalität
		route("      /apihle/auth/authorize    ", flags.authorize),
		route("      /hi/auth/authcode         ", flags.callback(pg)),

		route("      /auth/signin        ", auth.Signin),
		route("      /auth/login         ", auth.Login),
		route("      /auth/token         ", auth.TokenAuthHandler),
		route("      /auth/logout        ", auth.Logout),
		route("      /auth/session       ", auth.Session),

		route("      /api/auth/login         ", auth.Login),
		route("      /api/auth/logout        ", auth.Logout),
		route("      /api/auth/session       ", auth.Session),

		route("  GET /hi/meta/{path...}        ", hi.MetaHandlerMux),
		route("  GET /hi/media/{path...}       ", hi.FileHandlerMux),
		route("  GET /hi/media/thumbs          ", hi.ThumbHandlerMux),
		route("  GET /hi/media/proxy/{path...} ", hi.ServeReverseProxyMux),
		// higrp.GET("/tags/*path", hi.TagsHandler)
		route("  GET /media/videos/{path...}   ", serveContentWithPrefix("videos")), // used for serving local video on opalstack

		route("  GET /api/v1/entries/{path...} ", cmsapi.EntryDetails),
		route("  PUT /api/v1/entries/{path...} ", cmsapi.EntryUpdate),

		route("  GET /api/v1/entry             ", cmsuc.EntryLookup(cmsapi.CMS)), // lookup single entry with search params
		// route("  GET /api/v1/entries           ", cmsuc.EntriesLookup(cmsapi.CMS)), // lookup entries with search params

		route("  GET  /api/v1/super8/{path...}", super8.ServeHiVideo),
		route("  GET  /api/v1/personen/{person}", fapi.PersonHandler),
		route("  GET  /api/v1/reisen/{key}", fapi.ReiseHandler),
		route("  GET  /api/v1/search", cmsuc.SearchEntries(cmsapi.CMS)),
	)

	return srvr.ListenAndServe(cmd.Port, nil)
}

func (flags Flags) authorize(w http.ResponseWriter, r *http.Request) error {

	url := fmt.Sprintf("https://my.hidrive.com/client/authorize?client_id=%s=&response_type=code&scope=admin,rw&state=%s&redirect_uri=http://localhost:8000/hi/auth/authcode", flags.HidriveClientID, r.URL.Query().Get("state"))
	http.Redirect(w, r, url, http.StatusSeeOther)
	return nil
}

func (flags Flags) callback(db *db.DB) func(w http.ResponseWriter, r *http.Request) error {
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
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
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

func handle_pidfile(filename string) {
	if filename == "" {
		return
	}
	err := pidfile.Write(filename, os.Getpid())
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
	}(filename)
}

// package cli // main

// import (
// 	"bitbucket.org/hotelplan/webcc-pkg/web"
// 	"github.com/ihleven/ihle.cloud/hi"
// 	"github.com/ihleven/pkg/hidrive"
// )

// // github.com/ihleven/pkg/hidrive
// var drive *hidrive.Drive

// func (cmd *Command) run() {

// 	manager := hidrive.NewAuthManager(CLIENT_ID, CLIENT_SECRET)
// 	drive = hidrive.NewDrive(manager)

// 	// srv := web.NewServer(false, web.Addr("", cmd.Port))
// 	// srv.Register("/", serveSPA(cmd.FrontendPath)) // serve prerendred nuxt app
// 	// srv.Register("/hidrive", handler)                    //
// 	// srv.Register("/serve", serve)                        //
// 	// srv.Register("/wolfgang-ihle", serveWolfgangIhle())  // used for catalogs on wolfgang-ihle.de
// 	// srv.Register("/media/videos", servePrefix("videos")) // used for serving local video on opalstack
// 	// srv.Register("/proxy", serveReverseProxy())          // goldene hochzeit
// 	// srv.Register("/thumbs", thumbs) //

// 	// neu
// 	t, _ := manager.GetAccessToken("wolfgang")
// 	hfs := hi.New(t.AccessToken)
// 	// srv.Register("/api/meta", hi.MetaHandler("", *t))
// 	srv.Register("/api/raw", hi.FileHandler("", *t)) // neu: hi.FileHandler
// 	// srv.Register("/api/thumbs", hi.ThumbHandler(t.AccessToken))
// 	// srv.Register("/api/hidrive", FileServer(hfs))
// 	// srv.Register("/hidrive-new", FileServer(hfs))
// 	// srv.Register("/api/home", FileServer((dirFS)("/Users/ih"))) // lokales filesystem
// 	srv.Register("/api/tag", hi.TagsHandler(t.AccessToken, hfs))

// 	srv.Run()
// }
