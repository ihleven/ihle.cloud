package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ihleven/ihlvn/app/art/importart"
	"github.com/ihleven/ihlvn/app/auth"
	"github.com/ihleven/ihlvn/app/cmsapi"
	"github.com/ihleven/ihlvn/app/db"
	"github.com/ihleven/ihlvn/app/films"
	"github.com/ihleven/ihlvn/app/geheimtipp"
	"github.com/ihleven/ihlvn/app/hidrive"
	"github.com/ihleven/ihlvn/pkg/blob"
	"github.com/ihleven/ihlvn/pkg/cmd"
	"github.com/ihleven/ihlvn/pkg/mail"
	"github.com/ihleven/ihlvn/pkg/spa"
	"github.com/moby/moby/pkg/pidfile"

	"github.com/interhome-group/cms/content"
	"github.com/interhome-group/cms/mgmt"
	"github.com/interhome-group/cms/mgmt/search"
	"github.com/interhome-group/cms/pkg/godoc"

	"github.com/alexflint/go-arg"
	"github.com/ihleven/ihlvn/app/art"
	"github.com/ihleven/ihlvn/app/familie"
	"github.com/ihleven/ihlvn/pkg/api"
	"github.com/ihleven/ihlvn/pkg/hi"
	"github.com/interhome-group/cms/pkg/errs"

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

	// The geheimtipp backend: a separate running service with its own database.
	// Deliberately its own /api/v1, not the /api the old frontend exposes —
	// that one is served by that frontend's node layer, which goes away with
	// it. Not a constant, or after the domain moves the proxy would call itself.
	GeheimtippAPI   string `arg:"--ght-api,   env:GHT_API"   default:"https://ihleven.de/api/v1" placeholder:"URL"`
	GeheimtippMedia string `arg:"--ght-media, env:GHT_MEDIA" default:"https://ihleven.de/media" placeholder:"URL"`
	// The key the pool signs its tokens with. Shared with it by necessity: the
	// pool verifies a signature and asks nobody anything, so signing with the
	// same key is what lets this app speak for someone it has signed in. Unset
	// proxies the pool anonymously rather than refusing to start.
	GeheimtippSecret string `arg:"--ght-secret, env:GHT_SECRET" placeholder:"KEY" help:"signing key for the geheimtipp pool's tokens. Unset disables minting"`
	// Whether a pool player with no account here can still sign in, which
	// creates one. Meant to be turned off once everyone has: see
	// geheimtipp_migration.
	GeheimtippAdopt bool `arg:"--ght-adopt, env:GHT_ADOPT" default:"true" help:"let a geheimtipp player sign in with their pool password and create an account"`

	// The HiDrive account media is delivered from. Delivery is not per-viewer:
	// a film lives in one place, and the same entry has to resolve to the same
	// bytes for everyone — including, for public media, for nobody at all.
	// Browsing is the other case, and takes its drive from the session.
	MediaAlias    string `arg:"--media-alias, env:MEDIA_ALIAS" placeholder:"ALIAS" help:"HiDrive alias media is delivered from. Unset disables the media route"`
	MediaRoot     string `arg:"--media-root,  env:MEDIA_ROOT"  placeholder:"PATH"  help:"directory media keys resolve against. Nothing outside it can be addressed"`
	DriveRoot     string `arg:"--drive-root,  env:DRIVE_ROOT"  placeholder:"PATH"  default:"/public" help:"directory the file browser falls back to for an account with no drive of its own. Deliberately not the media root, which is where film keys resolve"`
	MediathekRoot string `arg:"--mediathek-root, env:MEDIATHEK_ROOT" placeholder:"PATH" help:"directory the shared video library resolves against. Its own root, not the media or drive one, so a bug in either route cannot reach the other's files. Unset disables the mediathek routes"`
	DataDir       string `arg:"--data-dir,     env:DATA_DIR"         default:"data" placeholder:"DIR"`
	SearchLevel   string `arg:"--search-level, env:SEARCH_LEVEL"     default:"basic" help:"Search level: off,basic,fulltext,extended" placeholder:"LEVEL"`
	SearchDir     string `arg:"--search-dir,   env:SEARCH_DIR"       default:"bleve" help:"Dirname of on disk search index, relative to the data dir" placeholder:"DIR"`
	SearchIndex   string `arg:"--search-index, env:SEARCH_INDEX"     default:"in-mem" help:"Search index mode: in-mem,recycle,create. NOTE: recycle and create DELETE the on-disk index if it cannot be opened" placeholder:"MODE"`

	JWTIssuer      string        `arg:"--jwt-issuer,env:JWT_ISSUER"                         default:"ihle.cloud" placeholder:"ISSUER"`
	JWTDuration    int           `arg:"--jwt-duration,env:JWT_DURATION"        default:"36000" help:"Duration of JWT token in seconds"`
	CookieName     string        `arg:"env:COOKIE_NAME"                        default:"jwt"  help:"Name for auth cookie"`
	CookieSameSite http.SameSite `arg:"env:COOKIE_SAME_SITE"                   default:"2"    help:"SameSite attribute of auth cookie: Default (1) Lax (2), Strict (3), None (4)"`
	CookieSecure   *bool         `arg:"env:COOKIE_SECURE"                                     help:"mark auth cookies Secure. Unset derives it from PUBLIC_URL's scheme, which is almost always what you want"`
	PublicURL      string        `arg:"--public-url,env:PUBLIC_URL"            default:"http://localhost:8000" placeholder:"URL" help:"the URL a browser reaches this app at. Behind a proxy this is the public https URL, not the local port. Passkey, cookie and OAuth settings derive from it"`
	CorsOrigins    []string      `arg:"--cors-origin,env:CORS_ORIGINS"                        help:"origins allowed to call this app cross-site. Empty is correct when the app serves its own frontend"`
	// MinPasswordLength is where the "this is short" warning starts, not a rule:
	// nothing refuses a password for being under it. It lives here because the
	// terminal, the admin form and the sentence that form shows all have to agree
	// on the number.
	MinPasswordLength int `arg:"env:MIN_PASSWORD_LENGTH" default:"12" help:"length below which a password is called short"`

	DbConn         string   `arg:"env:DB_CONN"                            default:"postgres://localhost:5432/authdb" placeholder:"CONN"`
	PasskeyRPID    string   `arg:"--passkey-rpid,env:PASSKEY_RPID"        default:"" help:"WebAuthn relying party id. Unset derives it from PUBLIC_URL's host; set it to a parent domain to share credentials across subdomains" placeholder:"HOST"`
	PasskeyOrigins []string `arg:"--passkey-origin,env:PASSKEY_ORIGINS"   help:"origins a passkey ceremony may come from. Unset derives PUBLIC_URL" placeholder:"URL"`
	Pidfile        string   `arg:"--pid-file,   env:PIDFILE"              default:"" placeholder:"FILENAME"`
	SPAPath        string   `arg:"--spa-path,   env:SPA_PATH"             default:"ui/.output/public" placeholder:"DIR" help:"path to nuxt spa"`
	Debug          bool     `arg:"-d,--debug,env:DEBUG" help:"verbose diagnostics, including CORS decisions"`
	// 	Pretty       bool   `arg:"--pretty,env:LOG_PRETTY"                      help:"Enable pretty logging"`
	// 	Verbose      bool   `arg:"-v,--verbose,env"                             help:"Enable verbose mode"`
}

type RootCmd struct {
	// *ServerCmd `arg:"subcommand:server"`
	*importart.ImportCmd `arg:"subcommand:import"`
	*mail.MailCmd        `arg:"subcommand:mail"`
	*AccountCmd          `arg:"subcommand:account"`
	*HiTokenCmd          `arg:"subcommand:hitoken"`

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

	content.Register(films.Film{})
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
	case *HiTokenCmd:
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

	hidriveTokens := hidrive.NewStore(pg.Pool())
	hitokens := hidrive.NewAccessTokens(hidriveTokens)
	hidriveOAuth := hidrive.NewOAuth(flags.HidriveClientID, flags.HidriveClientSecret,
		site.OAuthRedirect, hidriveTokens)

	// The drive media is delivered from. Stat results are cached because every
	// ranged request needs the object's size and validators before it can
	// answer, and each of those is a round-trip to HiDrive otherwise.
	var (
		filmstore  blob.Blobstore
		filmthumbs films.Thumbnailer
	)
	if flags.MediaAlias != "" {
		drive := hi.NewDrive(hitokens, hi.DriveConfig{Alias: flags.MediaAlias, Root: flags.MediaRoot}, nil)
		filmstore, filmthumbs = blob.NewStatCache(drive.Blobs(), 0), drive
	}

	// Browsing the storage. The deployment's alias stands in for an entitled
	// account that names none of its own, so the entitlement alone is enough to
	// have something to look at — rooted at DriveRoot rather than at the media
	// root, so the fallback is the shared material and not the film store.
	drives := hidrive.NewAPI(hitokens, hidrive.Shared{Alias: flags.MediaAlias, Root: flags.DriveRoot})

	// The shared video library: one alias, one root, decided here and not per
	// request. Everyone entitled to it sees the same shelf, which is what makes
	// the entitlement the only gate and the responses cacheable by URL. Its own
	// root rather than the media or drive one, on the same reasoning that keeps
	// those two apart: the root is the containment.
	var mediathek *hidrive.Library
	if flags.MediaAlias != "" && flags.MediathekRoot != "" {
		mediathek = hidrive.NewLibrary(
			hi.NewDrive(hitokens, hi.DriveConfig{Alias: flags.MediaAlias, Root: flags.MediathekRoot}, nil),
			slog.Default(),
		)
	}

	authsvc, err := openAuth(context.Background(), pg, site, flags)
	if err != nil {
		log.Fatal("openAuth: ", err)
	}
	go authsvc.StartCleanup(context.Background(), time.Hour)

	cms, err := mgmt.NewContentManager(flags.RepoContent, withMngrOptions(flags))
	if err != nil {
		log.Fatal(err)
	}

	fapi := familie.NewApi(cms.Repo, cms.Engine)

	// Whoever is signed in here is who the pool is told about. The token is
	// minted per forwarded request and never given to the browser, so there is
	// one credential in play and one place it can be revoked.
	ghtID := poolIdentity{svc: authsvc}
	ghtMinter := geheimtipp.NewMinter(flags.GeheimtippSecret)
	if ghtMinter == nil {
		slog.Warn("geheimtipp: no signing secret, the pool will be proxied anonymously")
	}

	ghtAPI, err := geheimtipp.Proxy(flags.GeheimtippAPI, "/ght", ghtID, ghtMinter)
	if err != nil {
		log.Fatal(err)
	}
	ghtMedia, err := geheimtipp.Proxy(flags.GeheimtippMedia, "/ght/media", ghtID, ghtMinter)
	if err != nil {
		log.Fatal(err)
	}

	// The enrollment link's lifetime matches the CLI's default: long enough to
	// hand over, short enough that a link left in a chat log stops working.
	admin := auth.NewAdminAPI(auth.NewAdmin(auth.NewStore(pg.Pool()), site.Origin, 15*time.Minute, flags.MinPasswordLength))

	// uiBase is where the JSON view points its cross-links; the same prefix the
	// routes below are mounted on, so a link in the rendered docs lands back
	// here rather than on the CMS this renderer came from.
	godocs := godoc.New(appSource, ".", "github.com/ihleven/ihlvn", "/godoc")

	var route = api.WithRoute

	srvr := api.New(

		api.Handler("/", spa.Serve(flags.SPAPath)),

		// route(" POST /tokenauth                ", tokenauth), // soll token liefern für spezielle Funktionalität
		// Binding a HiDrive account is an operator action, not something a
		// visitor can start.
		route("      /apihle/auth/authorize    ", requireAccount(authsvc, hidriveOAuth.Authorize)),
		route("      "+oauthCallbackPath+"     ", requireAccount(authsvc, hidriveOAuth.Callback)),

		route("      /auth/login         ", authsvc.Login),
		route("      /auth/logout        ", authsvc.Logout),
		route("      /auth/session       ", authsvc.Session),
		// Not behind requireAccount: it authenticates itself, which is what
		// lets an account confined to the Tipprunde reach it. Being able to
		// replace the password they arrived with is the point.
		route(" POST /auth/password      ", authsvc.ChangePassword),

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

		// This application's own documentation, rendered from the source embedded
		// in the binary. The renderer is the CMS's, reused rather than copied —
		// it serves plain func(w, r) error handlers and takes no view on who may
		// read them, so the gate is ours. Signed in only: the source is not
		// secret, but it is not the public site either.
		//
		// Under the API prefix, not /godoc: the pages at /godoc belong to the
		// SPA, which fetches these and renders them itself. A handler mounted
		// there would win against the SPA's catch-all and serve its own HTML
		// instead of the app.
		//
		// Order matters. The package route matches anything, so it goes last, or
		// it would swallow the source route.
		route("  GET /api/v1/godoc                ", requireAccount(authsvc, godocs.Index)),
		route("  GET /api/v1/godoc/src/{file...}  ", requireAccount(authsvc, godocs.Source)),
		route("  GET /api/v1/godoc/{pkg...}       ", requireAccount(authsvc, godocs.Package)),

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

		// Two ways to reach a file, distinguished by how it is addressed.
		//
		// A film is addressed by its entry id, and the storage key is read off
		// the entry — so nothing about where the bytes live reaches the
		// browser, and the film is one resource with sub-resources. The id is a
		// single segment, which is what lets "/poster.jpg" hang off it.
		route("  GET  /api/v1/films/{id}", requireAccount(authsvc, films.Handler(cms, filmstore, slog.Default()))),
		route("  GET  /api/v1/films/{id}/poster.jpg", requireAccount(authsvc, films.Poster(cms, filmthumbs))),
		route("  GET  /api/v1/films/{id}/chapters.vtt", requireAccount(authsvc, films.ChapterTrack(cms))),

		// The file browser. Gated on the entitlement rather than on merely being
		// signed in, because the entitlement is what decides that someone may
		// browse at all — see app/drive.
		route("  GET  /api/v1/drive/meta/{path...}  ", requireHidrive(authsvc, drives.Meta)),
		route("  GET  /api/v1/drive/media/{path...} ", requireHidrive(authsvc, drives.Media)),
		route("  GET  /api/v1/drive/thumb           ", requireHidrive(authsvc, drives.Thumbnail)),
		// Same drive, different delivery: stream reads the bytes here instead of
		// handing out a pre-signed URL, which costs a round-trip less per request
		// and keeps the store credential on this side. Worth it for anything
		// seeked through; media stays for fetching a file once.
		route("  GET  /api/v1/drive/stream/{path...}", requireHidrive(authsvc, drives.Stream)),

		// The shared video library. Gated on the mediathek entitlement, not the
		// hidrive one: being allowed to watch the shelf has nothing to do with
		// being allowed to browse your own files, and since the library resolves
		// against a fixed root there is no account-owned tree for the hidrive
		// rule to be about.
		route("  GET  /api/v1/mediathek/meta/{path...}  ", requireMediathek(authsvc, mediathek.Meta)),
		route("  GET  /api/v1/mediathek/stream/{path...}", requireMediathek(authsvc, mediathek.Stream)),
		route("  GET  /api/v1/personen/{person}", fapi.PersonHandler),
		route("  GET  /api/v1/reisen/{key}", fapi.ReiseHandler),
		route("  GET  /api/v1/search", optionalAccount(authsvc, cmsapi.SearchHandler(cms.Engine))),

		// The geheimtipp pool's own backend, under this origin so a browser may
		// reach it. Not gated on an account here: the pool has its own sign-in
		// against its own users, and this app's identity has no standing there.
		// See app/geheimtipp for why this exists and when it goes.
		//
		// Sign-in is ours rather than the proxy's because the upstream returns
		// the token in the body and leaves the cookie to its caller.
		route("      /ght/media/{path...} ", ghtMedia),
		route("      /ght/{path...}       ", ghtAPI),

		// Account administration. Everything here is gated on the admin
		// entitlement rather than on being signed in: these endpoints can grant
		// rights and set passwords, so hiding the section in the navigation —
		// which is presentation — is not the protection.
		route("     GET /api/v1/admin/modules                       ", requireAdmin(authsvc, admin.Areas)),
		route("     GET /api/v1/admin/accounts                      ", requireAdmin(authsvc, admin.List)),
		route("    POST /api/v1/admin/accounts                      ", requireAdmin(authsvc, admin.Create)),
		route("     GET /api/v1/admin/accounts/{name}               ", requireAdmin(authsvc, admin.Get)),
		route("     PUT /api/v1/admin/accounts/{name}               ", requireAdmin(authsvc, admin.Update)),
		route("    POST /api/v1/admin/accounts/{name}/password      ", requireAdmin(authsvc, admin.SetPassword)),
		route("    POST /api/v1/admin/accounts/{name}/enroll        ", requireAdmin(authsvc, admin.IssueEnrollment)),
		route("     GET /api/v1/admin/accounts/{name}/passkeys      ", requireAdmin(authsvc, admin.Passkeys)),
		route("  DELETE /api/v1/admin/accounts/{name}/passkeys/{id} ", requireAdmin(authsvc, admin.DeletePasskey)),
		route("  DELETE /api/v1/admin/accounts/{name}/passkeys      ", requireAdmin(authsvc, admin.Revoke)),
		route("  DELETE /api/v1/admin/accounts/{name}/sessions      ", requireAdmin(authsvc, admin.DeleteSessions)),
	)

	log.Printf("serving %s on port %d", site.Origin, cmd.Port)

	return srvr.ListenAndServe(cmd.Port, flags.CorsOrigins, flags.Debug)
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

		// OpenInRoot, not Join+Open: the path comes from the URL, and joining a
		// cleaned prefix with an uncleaned remainder is not containment.
		// ServeMux redirects a literal "..", which is what made this look safe,
		// but a percent-encoded one arrives decoded in PathValue — so
		// "..%2f.env" read the secrets file, unauthenticated. OpenInRoot refuses
		// anything resolving outside the directory, symlinks included.
		fd, err := os.OpenInRoot(prefix, r.PathValue("path"))
		if err != nil {
			// One answer for "no such file" and "not yours to ask for": the
			// difference is only useful to whoever is probing.
			return errs.New("not found", errs.HTTPStatus(http.StatusNotFound))
		}
		defer fd.Close()
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
