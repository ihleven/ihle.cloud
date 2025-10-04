package main

import (
	"github.com/alexflint/go-arg"

	"bitbucket.org/hotelplan/webcc-pkg/web"
	"github.com/ihleven/ihle.cloud/hi"
	"github.com/ihleven/pkg/hidrive"
)

var (
	VERSION       string
	BUILDTIME     string
	CLIENT_ID     string
	CLIENT_SECRET string
)

type Command struct {
	Flags
}

type Flags struct {
	Port         int    `arg:"-p,--port,env"       default:"10815"          help:"Port numbe"`
	Debug        bool   `arg:"-d,--debug,env"      default:"false"          help:"Enable debug mode"`
	Pretty       bool   `arg:"--pretty,env:LOG_PRETTY"                      help:"Enable pretty logging"`
	Verbose      bool   `arg:"-v,--verbose,env"                             help:"Enable verbose mode"`
	FrontendPath string `arg:"--frontend-path,env" default:".output/public" help:"path to nuxt .output/public"`
}

func main() {

	var cmd Command

	arg.MustParse(&cmd)

	cmd.run()
}

var drive *hidrive.Drive

func (cmd *Command) run() {

	manager := hidrive.NewAuthManager(CLIENT_ID, CLIENT_SECRET)
	drive = hidrive.NewDrive(manager)

	srv := web.NewServer(false, web.Addr("", cmd.Port))
	// srv.Register("/", serveSPA(cmd.FrontendPath)) // serve prerendred nuxt app
	// srv.Register("/hidrive", handler)                    //
	// srv.Register("/serve", serve)                        //
	// srv.Register("/wolfgang-ihle", serveWolfgangIhle())  // used for catalogs on wolfgang-ihle.de
	// srv.Register("/media/videos", servePrefix("videos")) // used for serving local video on opalstack
	// srv.Register("/proxy", serveReverseProxy())          // goldene hochzeit
	// srv.Register("/thumbs", thumbs) //

	// neu
	t, _ := manager.GetAccessToken("wolfgang")
	hfs := hi.New(t.AccessToken)
	// srv.Register("/api/meta", hi.MetaHandler("", *t))
	srv.Register("/api/raw", hi.FileHandler("", *t)) // neu: hi.FileHandler
	// srv.Register("/api/thumbs", hi.ThumbHandler(t.AccessToken))
	// srv.Register("/api/hidrive", FileServer(hfs))
	// srv.Register("/hidrive-new", FileServer(hfs))
	// srv.Register("/api/home", FileServer((dirFS)("/Users/ih"))) // lokales filesystem
	srv.Register("/api/tag", hi.TagsHandler(t.AccessToken, hfs))

	srv.Run()
}
