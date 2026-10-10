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
	"github.com/ihleven/ihlvn/app/gate"
	"github.com/ihleven/ihlvn/app/geheimtipp"
	"github.com/ihleven/ihlvn/app/hidrive"
	"github.com/ihleven/ihlvn/cli"
	"github.com/ihleven/ihlvn/pkg/blob"
	"github.com/ihleven/ihlvn/pkg/cmd"
	"github.com/ihleven/ihlvn/pkg/mail"
	"github.com/ihleven/ihlvn/pkg/pidfile"
	"github.com/ihleven/ihlvn/pkg/spa"

	"github.com/interhome-group/cms/content"
	"github.com/interhome-group/cms/mgmt"
	"github.com/interhome-group/cms/mgmt/search"
	"github.com/interhome-group/cms/pkg/godoc"

	"github.com/alexflint/go-arg"
	"github.com/ihleven/ihlvn/app/art"
	"github.com/ihleven/ihlvn/app/familie"
	"github.com/ihleven/ihlvn/pkg/hi"
	"github.com/ihleven/ihlvn/pkg/srvr"
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

	RetroRoot   string `arg:"--retro-root, env:RETRO_ROOT" placeholder:"PATH" help:"directory the magazine archive resolves against, on the same reasoning as the mediathek root: its own, so neither can reach the other's files. Unset disables the retro routes"`
	MusikRoot   string `arg:"--musik-root, env:MUSIK_ROOT" default:"/public/alben" placeholder:"PATH" help:"directory the music shelf resolves against, on the same reasoning as the mediathek root: its own, so neither can reach the other's files. Unset disables the musik routes"`
	DjvetRoot   string `arg:"--djvet-root, env:DJVET_ROOT" default:"/public/djvet" placeholder:"PATH" help:"directory the DJ archive resolves against, on the same reasoning as the mediathek root: its own, so neither can reach the other's files. Unset disables the djvet routes"`
	DataDir     string `arg:"--data-dir,     env:DATA_DIR"         default:"data" placeholder:"DIR"`
	SearchLevel string `arg:"--search-level, env:SEARCH_LEVEL"     default:"basic" help:"Search level: off,basic,fulltext,extended" placeholder:"LEVEL"`
	SearchDir   string `arg:"--search-dir,   env:SEARCH_DIR"       default:"bleve" help:"Dirname of on disk search index, relative to the data dir" placeholder:"DIR"`
	SearchIndex string `arg:"--search-index, env:SEARCH_INDEX"     default:"in-mem" help:"Search index mode: in-mem,recycle,create. NOTE: recycle and create DELETE the on-disk index if it cannot be opened" placeholder:"MODE"`

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
	*cli.AccountCmd      `arg:"subcommand:account"`
	*cli.HiTokenCmd      `arg:"subcommand:hitoken"`

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
	search.RegisterMappingAdapter(art.AdaptMapping)
	search.RegisterMappingAdapter(art.AdaptMappingAusstellung)
	// content.Register(ctype.Page{})

	var err error
	var root RootCmd
	var flags Flags

	p := arg.MustParse(&root, &flags)

	switch subcmd := p.Subcommand().(type) {
	case *importart.ImportCmd:
		// Only a writing import needs the content manager, and only then is a
		// configured repository worth insisting on: a dry run reads the old
		// database and prints, which nothing about the archive can affect.
		// Built here rather than in the command because that is where the flags
		// are.
		var cms *mgmt.Mngr
		if subcmd.Write {
			// Refused rather than left to the manager. An empty repository
			// still yields a usable Mngr, so the import runs to completion and
			// reports the entries written — into a repository that is nowhere,
			// with the push failing separately as a warning. Succeeding at
			// nothing is the worst of the three outcomes.
			if flags.RepoContent == "" {
				err = fmt.Errorf("--write needs a content repository: set REPO_CONTENT")
				break
			}
			cms, err = mgmt.NewContentManager(flags.RepoContent, withMngrOptions(flags))
		}
		if err == nil {
			err = subcmd.Run(cms)
		}
	case *mail.MailCmd:
		err = subcmd.Run()
	case *cli.AccountCmd:
		err = subcmd.Run(cli.Accounts{
			DbConn:            flags.DbConn,
			MinPasswordLength: flags.MinPasswordLength,
		})
	case *cli.HiTokenCmd:
		err = subcmd.Run(cli.Tokens{
			DbConn:       flags.DbConn,
			ClientID:     flags.HidriveClientID,
			ClientSecret: flags.HidriveClientSecret,
			PublicURL:    flags.PublicURL,
		})
	default:
		// Only the server owns the pidfile. Claiming it for a subcommand would
		// make every command refuse to run while the server is up, since it
		// would find the server's own pid there and take it for a duplicate.
		claimPidfile(flags.Pidfile)
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

// deps is everything the routes are declared against.
//
// A struct rather than a long parameter list because routes() is the whole of
// what reads it, and what routes() may reach is exactly what belongs in here —
// so adding a field is a decision rather than a habit.
type deps struct {
	flags Flags
	site  *site

	authsvc *auth.Service
	admin   *auth.AdminAPI
	cms     *mgmt.Mngr

	hitokens     hidrive.AccessTokens
	hidriveOAuth *hidrive.OAuth
	drives       *hidrive.API
	filmstore    blob.Blobstore
	filmthumbs   films.Thumbnailer

	// The familie and art APIs are named by what the routes ask of them, because
	// both constructors return an unexported type — this is the shape that
	// crosses the package boundary, and writing it down is what lets deps hold
	// them at all.
	fapi familieAPI
	aapi artAPI

	ghtAPI   handler
	ghtMedia handler

	godocs *godoc.Handler

	// The CMS's API as a handler, and the mux it was registered on. The mux is
	// kept because it is the only record of what the CMS mounted there — those
	// routes are declared by its code, not by the table below.
	contentAPI http.Handler
	contentMux *srvr.RecordingMux
}

// handler is the signature this app's routes are written in: an error comes
// back rather than being written, so the router's middleware renders it.
type handler = func(http.ResponseWriter, *http.Request) error

type familieAPI interface {
	PersonHandler(http.ResponseWriter, *http.Request) error
	ReiseHandler(http.ResponseWriter, *http.Request) error
}

type artAPI interface {
	ArtworkHandler(http.ResponseWriter, *http.Request) error
	MappingHandler(http.ResponseWriter, *http.Request) error
	Search(http.ResponseWriter, *http.Request) error
}

// build brings up everything the server needs, and hands back the way to put it
// down again.
//
// Separated from declaring the routes because they are different jobs and only
// one of them needs a database: what routes() produces can be inspected by a
// test, which is how the whole URL space can be checked for the collisions that
// Go's router only reports by panicking as the server starts.
//
// Failures are returned rather than fatal. This used to call log.Fatal six
// times, which is fine for a function only ever called from a running binary
// and useless for one a test calls.
func build(flags Flags) (*deps, func(), error) {

	pg, err := db.New(flags.DbConn)
	if err != nil {
		return nil, nil, fmt.Errorf("connecting to the database: %w", err)
	}
	// Given back rather than deferred here: what this function opens outlives
	// it, and the caller is what knows when serving has stopped.
	done := func() { pg.Close() }

	// Everything host-dependent comes from here, so a bad value fails at startup
	// rather than surfacing later as a ceremony that will not complete.
	site, err := newSite(flags)
	if err != nil {
		return nil, done, err
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

	authsvc, err := openAuth(context.Background(), pg, site, flags)
	if err != nil {
		return nil, done, fmt.Errorf("bringing up authentication: %w", err)
	}
	go authsvc.StartCleanup(context.Background(), time.Hour)

	cms, err := mgmt.NewContentManager(flags.RepoContent, withMngrOptions(flags))
	if err != nil {
		return nil, done, fmt.Errorf("opening the content repository: %w", err)
	}

	// The CMS's own reader rather than its repository: what used to be a field
	// holding the old storage layer is now how the manager serves every read.
	fapi := familie.NewApi(cms.Storage(), cms.Engine)

	// The art archive's own API. Its search is faceted over the artwork fields
	// (Gattung, Technik, Träger) in a way the generic entry search is not, which
	// is the whole reason it exists — see /werke.
	aapi := art.NewApi(cms.Storage(), cms.Engine)

	// Whoever is signed in here is who the pool is told about. The token is
	// minted per forwarded request and never given to the browser, so there is
	// one credential in play and one place it can be revoked.
	ghtID := gate.PoolIdentity{Svc: authsvc}
	ghtMinter := geheimtipp.NewMinter(flags.GeheimtippSecret)
	if ghtMinter == nil {
		slog.Warn("geheimtipp: no signing secret, the pool will be proxied anonymously")
	}

	ghtAPI, err := geheimtipp.Proxy(flags.GeheimtippAPI, "/ght", ghtID, ghtMinter)
	if err != nil {
		return nil, done, fmt.Errorf("proxying the pool: %w", err)
	}
	ghtMedia, err := geheimtipp.Proxy(flags.GeheimtippMedia, "/ght/media", ghtID, ghtMinter)
	if err != nil {
		return nil, done, fmt.Errorf("proxying the pool's media: %w", err)
	}

	// The enrollment link's lifetime matches the CLI's default: long enough to
	// hand over, short enough that a link left in a chat log stops working.
	admin := auth.NewAdminAPI(auth.NewAdmin(auth.NewStore(pg.Pool()), site.Origin, 15*time.Minute, flags.MinPasswordLength))

	// uiBase is where the JSON view points its cross-links; the same prefix the
	// routes below are mounted on, so a link in the rendered docs lands back
	// here rather than on the CMS this renderer came from.
	godocs := godoc.New(appSource, ".", "github.com/ihleven/ihlvn", "/godoc")

	contentHandler, contentMux := contentAPI(cms, authsvc, cmd.Version())

	return &deps{
		flags: flags, site: site,
		contentAPI: contentHandler, contentMux: contentMux,
		authsvc: authsvc, admin: admin, cms: cms,
		hitokens: hitokens, hidriveOAuth: hidriveOAuth, drives: drives,
		filmstore: filmstore, filmthumbs: filmthumbs,
		fapi: fapi, aapi: aapi,
		ghtAPI: ghtAPI, ghtMedia: ghtMedia,
		godocs: godocs,
	}, done, nil
}

// routes is the whole of this application's URL space.
//
// Given the dependencies and nothing else, so it can be called without a
// database behind it — which is what lets a test register every pattern and
// find out whether they can coexist, rather than finding out when the server
// refuses to start.
func routes(d *deps) []func(*srvr.Api) {
	var route = srvr.WithRoute

	// Named locally so the list below is the list that was here before, word for
	// word. A route declaration should read as an address and a handler, not as
	// an address and a field access.
	flags := d.flags
	authsvc, admin, cms := d.authsvc, d.admin, d.cms
	hitokens, hidriveOAuth, drives := d.hitokens, d.hidriveOAuth, d.drives
	filmstore, filmthumbs := d.filmstore, d.filmthumbs
	fapi, aapi := d.fapi, d.aapi
	ghtAPI, ghtMedia, godocs := d.ghtAPI, d.ghtMedia, d.godocs

	return []func(*srvr.Api){

		// The CMS's API, mounted whole rather than re-implemented. See contentapi.go.
		srvr.Handler(contentAPIPrefix, d.contentAPI),

		srvr.Handler("/", spa.Serve(flags.SPAPath)),

		// route(" POST /tokenauth                ", tokenauth), // soll token liefern für spezielle Funktionalität
		// Binding a HiDrive account is an operator action, not something a
		// visitor can start.
		route("      /apihle/auth/authorize    ", gate.Account(authsvc, hidriveOAuth.Authorize)),
		route("      "+oauthCallbackPath+"     ", gate.Account(authsvc, hidriveOAuth.Callback)),

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
		// Under an API prefix, not /godoc: the pages at /godoc belong to the
		// SPA, which fetches these and renders them itself. A handler mounted
		// there would win against the SPA's catch-all and serve its own HTML
		// instead of the app.
		//
		// The prefix is the CMS's rather than this app's, because the pages that
		// call these come from the CMS's godoc layer, and that layer asks where
		// the CMS API is. This app has several APIs and the layer has one key, so
		// the endpoints it fetches move to the prefix it knows. That they are
		// served from this app's own source, not the CMS's, is the point of them.
		//
		// Order matters. The package route matches anything, so it goes last, or
		// it would swallow the source route.
		route("  GET /content-api/v1/godoc                ", gate.Account(authsvc, godocs.Index)),
		route("  GET /content-api/v1/godoc/src/{file...}  ", gate.Account(authsvc, godocs.Source)),
		route("  GET /content-api/v1/godoc/{pkg...}       ", gate.Account(authsvc, godocs.Package)),

		// higrp.GET("/tags/*path", hi.TagsHandler)
		route("  GET /media/videos/{path...}   ", serveContentWithPrefix("videos")), // used for serving local video on opalstack

		// Reading content is open: the public site is rendered from it. Where an
		// account is present it is attached, so the entry's own ACL can decide.
		route("  GET /api/v1/entries/{path...} ", gate.OptionalAccount(authsvc, cmsapi.EntryDetails(cms))),
		// Writing content requires an account. The entry ACL cannot stand in for
		// this: every entry is mode-zero, and a zero mode allows everyone.
		route("  PUT /api/v1/entries/{path...} ", gate.Account(authsvc, cmsapi.EntryUpdate(cms))),

		route("  GET /api/v1/entry             ", gate.OptionalAccount(authsvc, cmsapi.EntryLookup(cms))), // lookup single entry with search params
		// route("  GET /api/v1/entries           ", cmsuc.EntriesLookup(cmsapi.CMS)), // lookup entries with search params

		// Two ways to reach a file, distinguished by how it is addressed.
		//
		// A film is addressed by its entry id, and the storage key is read off
		// the entry — so nothing about where the bytes live reaches the
		// browser, and the film is one resource with sub-resources. The id is a
		// single segment, which is what lets "/poster.jpg" hang off it.
		route("  GET  /api/v1/films/{id}", gate.Account(authsvc, films.Handler(cms, filmstore, slog.Default()))),
		route("  GET  /api/v1/films/{id}/poster.jpg", gate.Account(authsvc, films.Poster(cms, filmthumbs))),
		route("  GET  /api/v1/films/{id}/chapters.vtt", gate.Account(authsvc, films.ChapterTrack(cms))),

		// The file browser. Gated on the entitlement rather than on merely being
		// signed in, because the entitlement is what decides that someone may
		// browse at all — see app/drive.
		route("  GET  /api/v1/drive/meta/{path...}  ", gate.Hidrive(authsvc, drives.Meta)),
		route("  GET  /api/v1/drive/media/{path...} ", gate.Hidrive(authsvc, drives.Media)),
		route("  GET  /api/v1/drive/thumb           ", gate.Hidrive(authsvc, drives.Thumbnail)),
		// Same drive, different delivery: stream reads the bytes here instead of
		// handing out a pre-signed URL, which costs a round-trip less per request
		// and keeps the store credential on this side. Worth it for anything
		// seeked through; media stays for fetching a file once.
		route("  GET  /api/v1/drive/stream/{path...}", gate.Hidrive(authsvc, drives.Stream)),

		// The shelves: the mediathek, the magazines, the music and the DJ sets.
		// Declared as a list in shelves.go, where the root each is confined to
		// and the entitlement each demands can be read down one page.
		shelfRoutes(authsvc, hitokens, flags),

		route("  GET  /api/v1/personen/{person}", fapi.PersonHandler),
		route("  GET  /api/v1/reisen/{key}", fapi.ReiseHandler),

		// Ungated, as the familie routes above are. The archive is readable and
		// the UI gates the page, not the data.
		route("  GET  /api/v1/artworks/{id}", aapi.ArtworkHandler),
		route("  GET  /api/v1/art/search", aapi.Search),
		route("  GET  /api/v1/art/search/mapping", aapi.MappingHandler),
		route("  GET  /api/v1/search", gate.OptionalAccount(authsvc, cmsapi.SearchHandler(cms.Engine))),

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
		route("     GET /api/v1/admin/modules                       ", gate.Admin(authsvc, admin.Areas)),
		route("     GET /api/v1/admin/accounts                      ", gate.Admin(authsvc, admin.List)),
		route("    POST /api/v1/admin/accounts                      ", gate.Admin(authsvc, admin.Create)),
		route("     GET /api/v1/admin/accounts/{name}               ", gate.Admin(authsvc, admin.Get)),
		route("     PUT /api/v1/admin/accounts/{name}               ", gate.Admin(authsvc, admin.Update)),
		route("    POST /api/v1/admin/accounts/{name}/password      ", gate.Admin(authsvc, admin.SetPassword)),
		route("    POST /api/v1/admin/accounts/{name}/enroll        ", gate.Admin(authsvc, admin.IssueEnrollment)),
		route("     GET /api/v1/admin/accounts/{name}/passkeys      ", gate.Admin(authsvc, admin.Passkeys)),
		route("  DELETE /api/v1/admin/accounts/{name}/passkeys/{id} ", gate.Admin(authsvc, admin.DeletePasskey)),
		route("  DELETE /api/v1/admin/accounts/{name}/passkeys      ", gate.Admin(authsvc, admin.Revoke)),
		route("  DELETE /api/v1/admin/accounts/{name}/sessions      ", gate.Admin(authsvc, admin.DeleteSessions)),
	}
}

func (cmd *RootCmd) RunServer(flags Flags) error {
	d, done, err := build(flags)
	if done != nil {
		defer done()
	}
	if err != nil {
		return err
	}

	log.Printf("serving %s on port %d", d.site.Origin, cmd.Port)

	return srvr.New(routes(d)...).ListenAndServe(cmd.Port, flags.CorsOrigins, flags.Debug)
}

// openAuth brings up the authentication service against the account database.
//
// Migrations run here rather than as a separate deployment step, so the binary
// and the schema it expects cannot get out of step.
func openAuth(ctx context.Context, pg *db.DB, site *site, flags Flags) (*auth.Service, error) {
	if err := db.Migrate(ctx, pg.Pool()); err != nil {
		return nil, fmt.Errorf("migrating the account database: %w", err)
	}

	return auth.NewService(auth.NewStore(pg.Pool()), auth.Config{
		CookieName:    flags.CookieName,
		CookieSecure:  site.CookieSecure,
		SameSite:      flags.CookieSameSite,
		SessionTTL:    time.Duration(flags.JWTDuration) * time.Second,
		Issuer:        flags.JWTIssuer,
		Audience:      "famihlie",
		FailureLimit:  10,
		FailureWindow: 15 * time.Minute,

		// Whether an account with no drive of its own still has something to
		// browse, which is what decides if the file browser is offered at all.
		SharedDrive: flags.MediaAlias != "",

		// Whether a pool player with no account here can sign in with their
		// pool password, which creates one.
		AdoptGeheimtipp: flags.GeheimtippAdopt,

		// The same screening policy the administrator's path uses, so a
		// password someone picks for themselves is held to one standard.
		MinPasswordLength: flags.MinPasswordLength,

		RPID:      site.RPID,
		RPOrigins: site.PasskeyOrigins,
		// RPDisplayName is left to the passkey package, which falls back to the relying party
		// id — so the name a browser shows follows the domain automatically.
	}, slog.Default())
}

// serveContentWithPrefix serves files from a local directory — the videos kept
// beside the binary rather than on the media drive.
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

// claimPidfile takes the pidfile for this process and gives it up on the way
// out, so a stale file does not outlive the server that wrote it.
func claimPidfile(filename string) {
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
