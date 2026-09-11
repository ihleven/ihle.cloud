package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/ihleven/ihlvn/app/art/importart"
	"github.com/ihleven/ihlvn/app/cmsapi"
	"github.com/ihleven/ihlvn/app/db"
	"github.com/ihleven/ihlvn/app/super8"
	"github.com/ihleven/ihlvn/pkg/cmd"
	"github.com/ihleven/ihlvn/pkg/mail"
	"github.com/ihleven/ihlvn/pkg/spa"
	"github.com/moby/moby/pkg/pidfile"

	"github.com/interhome-group/cms/content"
	"github.com/interhome-group/cms/mgmt"
	"github.com/interhome-group/cms/mgmt/search"
	"github.com/interhome-group/cms/pkg/errs"

	"github.com/alexflint/go-arg"
	"github.com/ihleven/ihlvn/app/art"
	"github.com/ihleven/ihlvn/app/familie"
	"github.com/ihleven/ihlvn/pkg/api"
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
	HidriveClientID     string `arg:"--hidrive-client-id,     env:CLIENT_ID"        placeholder:"ID"`
	HidriveClientSecret string `arg:"--hidrive-client-secret, env:CLIENT_SECRET"    placeholder:"SECRET"`
	RepoContent         string `arg:"                         env:REPO_CONTENT"     placeholder:"URL" `
	DataDir             string `arg:"--data-dir,     env:DATA_DIR"         default:"data" placeholder:"DIR"`
	SearchLevel         string `arg:"--search-level, env:SEARCH_LEVEL"     default:"basic" help:"Search level: off,basic,fulltext,extended" placeholder:"LEVEL"`
	SearchDir           string `arg:"--search-dir,   env:SEARCH_DIR"       default:"bleve" help:"Dirname of on disk search index, relative to the data dir" placeholder:"DIR"`
	SearchIndex         string `arg:"--search-index, env:SEARCH_INDEX"     default:"in-mem" help:"Search index mode: in-mem,recycle,create. NOTE: recycle and create DELETE the on-disk index if it cannot be opened" placeholder:"MODE"`

	JWTIssuer      string        `arg:"--jwt-issuer,env:JWT_ISSUER"                         default:"ihle.cloud" placeholder:"ISSUER"`
	JWTSecretKey   string        `arg:"--jwt-secret,env:JWT_SECRET_KEY"                                          placeholder:"KEY"`
	JWTDuration    int           `arg:"--jwt-duration,env:JWT_DURATION"        default:"36000" help:"Duration of JWT token in seconds"`
	CookieName     string        `arg:"env:COOKIE_NAME"                        default:"jwt"  help:"Name for auth cookie"`
	CookieSameSite http.SameSite `arg:"env:COOKIE_SAME_SITE"                   default:"2"    help:"SameSite attribute of auth cookie: Default (1) Lax (2), Strict (3), None (4)"`
	CookieSecure   *bool         `arg:"env:COOKIE_SECURE"                                     help:"mark auth cookies Secure. Unset derives it from PUBLIC_URL's scheme, which is almost always what you want"`
	PublicURL      string        `arg:"--public-url,env:PUBLIC_URL"            default:"http://localhost:8000" placeholder:"URL" help:"the URL a browser reaches this app at. Behind a proxy this is the public https URL, not the local port. Passkey, cookie and OAuth settings derive from it"`
	CorsOrigins    []string      `arg:"--cors-origin,env:CORS_ORIGINS"                        help:"origins allowed to call this app cross-site. Empty is correct when the app serves its own frontend"`
	DbConn         string        `arg:"env:DB_CONN"                            default:"postgres://localhost:5432/authdb" placeholder:"CONN"`
	PasskeyRPID    string        `arg:"--passkey-rpid,env:PASSKEY_RPID"        default:"" help:"WebAuthn relying party id. Unset derives it from PUBLIC_URL's host; set it to a parent domain to share credentials across subdomains" placeholder:"HOST"`
	PasskeyOrigins []string      `arg:"--passkey-origin,env:PASSKEY_ORIGINS"   help:"origins a passkey ceremony may come from. Unset derives PUBLIC_URL" placeholder:"URL"`
	Pidfile        string        `arg:"--pid-file,   env:PIDFILE"              default:"" placeholder:"FILENAME"`
	SPAPath        string        `arg:"--spa-path,   env:SPA_PATH"             default:"ui/.output/public" placeholder:"DIR" help:"path to nuxt spa"`
	Debug          bool          `arg:"-d,--debug,env:DEBUG" help:"verbose diagnostics, including CORS decisions"`
	// 	Pretty       bool   `arg:"--pretty,env:LOG_PRETTY"                      help:"Enable pretty logging"`
	// 	Verbose      bool   `arg:"-v,--verbose,env"                             help:"Enable verbose mode"`
}

type RootCmd struct {
	// *ServerCmd `arg:"subcommand:server"`
	*importart.ImportCmd `arg:"subcommand:import"`
	*mail.MailCmd        `arg:"subcommand:mail"`
	*AccountCmd          `arg:"subcommand:account"`

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

	switch subcmd := p.Subcommand().(type) {
	case *importart.ImportCmd:
		err = subcmd.Run()
	case *mail.MailCmd:
		err = subcmd.Run()
	case *AccountCmd:
		err = subcmd.Run(flags)
	default:
		// Only the server owns the pidfile. Claiming it for a subcommand would
		// make every command refuse to run while the server is up, since it
		// would find the server's own pid there and take it for a duplicate.
		handle_pidfile(flags.Pidfile)
		err = root.RunServer(flags)
	}

	// A command that ran and failed is not a usage mistake, so it reports the
	// reason on its own. p.Fail is for arguments that could not be parsed, where
	// the usage line is the useful part.
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
}

func withMngrOptions(flags Flags) func(*mgmt.Config) {
	return func(conf *mgmt.Config) {
		conf.DataDir = flags.DataDir
		conf.RepoClone = false
		// The search config now carries one path and a mode instead of a
		// directory plus an index name.
		conf.Search.IndexPath = filepath.Join(flags.DataDir, flags.SearchDir)
		conf.Search.IndexMode = flags.SearchIndex
		conf.Search.Level = search.ParseLevel(flags.SearchLevel)
	}
}

func (cmd *RootCmd) RunServer(flags Flags) error {

	pg, err := db.New(flags.DbConn)
	if err != nil {
		log.Fatal("db.New:", err)
	}
	defer pg.Close()

	// Everything host-dependent comes from here, so a bad value fails at startup
	// rather than surfacing later as a ceremony that will not complete.
	site, err := newSite(flags)
	if err != nil {
		log.Fatal(err)
	}

	hitokens := hiDriveTokens{hi.NewTokenMngr(pg)}

	authsvc, err := openAuth(context.Background(), pg, site, flags, hitokens)
	if err != nil {
		log.Fatal("openAuth: ", err)
	}
	go authsvc.StartCleanup(context.Background(), time.Hour)

	cms, err := mgmt.NewContentManager(flags.RepoContent, withMngrOptions(flags))
	if err != nil {
		log.Fatal(err)
	}

	fapi := familie.NewApi(cms.Repo, cms.Engine)

	var route = api.WithRoute

	srvr := api.New(

		api.Handler("/", spa.Serve(flags.SPAPath)),

		// route(" POST /tokenauth                ", tokenauth), // soll token liefern für spezielle Funktionalität
		// Binding a HiDrive account is an operator action, not something a
		// visitor can start.
		route("      /apihle/auth/authorize    ", requireAccount(authsvc, site.authorize(flags))),
		route("      "+oauthCallbackPath+"     ", requireAccount(authsvc, site.callback(flags, pg))),

		route("      /auth/login         ", authsvc.Login),
		route("      /auth/logout        ", authsvc.Logout),
		route("      /auth/session       ", authsvc.Session),

		// Passkey ceremonies. Registration is authorised by a session or an
		// enrollment link, and in either case by the account's password.
		route(" POST /auth/passkey/login/begin      ", authsvc.LoginBegin),
		route(" POST /auth/passkey/login/finish     ", authsvc.LoginFinish),
		route(" POST /auth/passkey/register/begin   ", authsvc.RegisterBegin),
		route(" POST /auth/passkey/register/finish  ", authsvc.RegisterFinish),
		route(" DELETE /auth/passkey/{id}           ", authsvc.DeletePasskey),
		route("  GET /auth/enroll                   ", authsvc.Enroll),
		route("  GET /auth/passkey                  ", authsvc.Passkeys),

		route("      /api/auth/login         ", authsvc.Login),
		route("      /api/auth/logout        ", authsvc.Logout),
		route("      /api/auth/session       ", authsvc.Session),

		route("  GET /hi/meta/{path...}        ", hi.MetaHandlerMux),
		route("  GET /hi/media/{path...}       ", hi.FileHandlerMux),
		route("  GET /hi/media/thumbs          ", hi.ThumbHandlerMux),
		route("  GET /hi/media/proxy/{path...} ", hi.ServeReverseProxyMux),
		// higrp.GET("/tags/*path", hi.TagsHandler)
		route("  GET /media/videos/{path...}   ", serveContentWithPrefix("videos")), // used for serving local video on opalstack

		// Reading content is open: the public site is rendered from it. Where an
		// account is present it is attached, so the entry's own ACL can decide.
		route("  GET /api/v1/entries/{path...} ", optionalAccount(authsvc, cmsapi.EntryDetails(cms))),
		// Writing content requires an account. The entry ACL cannot stand in for
		// this: every entry is mode-zero, and a zero mode allows everyone.
		route("  PUT /api/v1/entries/{path...} ", requireAccount(authsvc, cmsapi.EntryUpdate(cms))),

		route("  GET /api/v1/entry             ", optionalAccount(authsvc, cmsapi.EntryLookup(cms))), // lookup single entry with search params
		// route("  GET /api/v1/entries           ", cmsuc.EntriesLookup(cmsapi.CMS)), // lookup entries with search params

		// Family video: the account is needed both to allow the request and to
		// pick which stored HiDrive credential serves it.
		route("  GET  /api/v1/super8/{path...}", requireAccount(authsvc, super8.ServeHiVideo(hitokens))),
		route("  GET  /api/v1/personen/{person}", fapi.PersonHandler),
		route("  GET  /api/v1/reisen/{key}", fapi.ReiseHandler),
		route("  GET  /api/v1/search", optionalAccount(authsvc, cmsapi.SearchHandler(cms.Engine))),
	)

	log.Printf("serving %s on port %d", site.Origin, cmd.Port)

	return srvr.ListenAndServe(cmd.Port, flags.CorsOrigins, flags.Debug)
}

// authorize starts the HiDrive consent flow, which is how a refresh token is
// obtained for an alias in the first place — the session system authenticates
// people, this authorises storage access.
//
// The redirect_uri is built from the public URL and has to match the
// registration held by HiDrive exactly, so changing the domain means updating it
// there too.
func (site *site) authorize(flags Flags) func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		// state comes back untouched and is where the browser is sent
		// afterwards, so it is a path on this site and is validated as one.
		next := r.URL.Query().Get("state")
		if next == "" {
			next = "/"
		}
		if !isLocalPath(next) {
			return errs.New("state must be a path on this site", errs.HTTPStatus(http.StatusBadRequest))
		}

		params := url.Values{
			"client_id":     {flags.HidriveClientID},
			"response_type": {"code"},
			"scope":         {"admin,rw"},
			"state":         {next},
			"redirect_uri":  {site.OAuthRedirect},
		}
		http.Redirect(w, r, "https://my.hidrive.com/client/authorize?"+params.Encode(), http.StatusSeeOther)
		return nil
	}
}

// isLocalPath keeps a redirect on this site. Without it the state parameter is
// an open redirect: whatever it holds is where the browser ends up.
func isLocalPath(p string) bool {
	return strings.HasPrefix(p, "/") && !strings.HasPrefix(p, "//")
}

func (site *site) callback(flags Flags, db *db.DB) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {

		token, err := hi.RefreshTokenWithAuthCode(
			flags.HidriveClientID, flags.HidriveClientSecret, r.URL.Query().Get("code"))
		if err != nil {
			return errs.Wrap(err, "exchanging the authorization code", errs.HTTPStatus(http.StatusUnauthorized))
		}
		if err := db.StoreToken(token); err != nil {
			return err
		}

		// SameSite=None is only accepted alongside Secure, so over plain http the
		// pairing has to fall back or the browser discards the cookie outright.
		sameSite := http.SameSiteNoneMode
		if !site.CookieSecure {
			sameSite = http.SameSiteLaxMode
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "hitoken",
			Value:    token.AccessToken,
			Path:     "/",
			MaxAge:   token.ExpiresIn,
			HttpOnly: true,
			Secure:   site.CookieSecure,
			SameSite: sameSite,
		})

		next := r.URL.Query().Get("state")
		if !isLocalPath(next) {
			next = "/"
		}
		http.Redirect(w, r, next, http.StatusSeeOther)
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

// auth0
