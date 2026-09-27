package main

import (
	"log/slog"
	"net/http"

	"github.com/ihleven/ihlvn/app/cmsauth"
	"github.com/ihleven/ihlvn/app/gate"
	"github.com/ihleven/ihlvn/app/hidrive"
	"github.com/ihleven/ihlvn/pkg/hi"
	"github.com/ihleven/ihlvn/pkg/srvr"
	"github.com/interhome-group/cms/content"
)

// The shelves: fixed trees of media, each offered to whoever is entitled to it.
//
// Four areas that are the same thing four times — one drive, one root, one
// entitlement, a handful of read operations — so they are written as a list
// rather than as four copies of the wiring. What is per-area stays per-area and
// becomes a column, which is the point: the root each is confined to and the
// right each demands are the two facts worth being able to read down a page,
// and until this they were spread over a construction block, a guard function
// in another file, and a run of route lines.
//
// Each root is its own directory, and separate on purpose: a mistake in the
// path handling of one set of routes cannot reach what another serves. That
// reasoning belonged to each area individually before; it applies to all of
// them at once and is stated here once.
//
// This is for shelves and nothing else. The file browser resolves its drive per
// request from the account rather than from a fixed root, and a film is reached
// through a stored entry and protected by that entry's own access list — so
// neither is a row here, and their absence says what they are.
type shelf struct {
	// prefix is where the area is served, without a trailing slash.
	prefix string
	// root is the directory on the media drive that is the whole of this area.
	// Empty means the deployment has not configured it, which is answered per
	// request rather than by the route being absent.
	root string
	// right is what admits somebody, and refusal is what they are told when
	// they lack it.
	//
	// Both kinds of right appear in this column and the value says which: an
	// area (module.<id>) is what a visitor may be shown, a capability is what
	// somebody may do. The distinction is real — see cmsauth — and here it is
	// carried by the key rather than by which wrapper a call site chose, which
	// is the clearer of the two when they sit in a list side by side.
	right   content.PermissionKey
	refusal string
	// ops are the operations this area offers, named as they appear in the URL.
	// They differ per area because the areas differ: a magazine has no stream
	// and an album has no cover to scan.
	ops []string
}

// libraryOps maps the name an operation has in a URL to the method that serves
// it. Method expressions rather than closures, so the table stays a table.
var libraryOps = map[string]func(*hidrive.Library, http.ResponseWriter, *http.Request) error{
	"meta":   (*hidrive.Library).Meta,
	"stream": (*hidrive.Library).Stream,
	"thumb":  (*hidrive.Library).Thumbnail,
	"media":  (*hidrive.Library).Media,
	"search": (*hidrive.Library).Search,
	"tags":   (*hidrive.Library).Tags,
}

// shelves is the whole of what this app offers as a fixed tree of media.
func shelves(flags Flags) []shelf {
	return []shelf{{
		// The shared video library. Gated on the mediathek entitlement, not the
		// hidrive one: being allowed to watch the shelf has nothing to do with
		// being allowed to browse your own files, and since the library resolves
		// against a fixed root there is no account-owned tree for the hidrive
		// rule to be about.
		prefix: "/api/v1/mediathek", root: flags.MediathekRoot,
		right: cmsauth.Mediathek, refusal: "the mediathek requires the mediathek entitlement",
		ops: []string{"meta", "stream"},
	}, {
		// The magazine archive. Thumbnails as well as bytes, because an issue is
		// a hundred-megabyte PDF and what a listing wants is its cover.
		// No streaming route here, unlike the mediathek: the archive holds
		// documents, and a reader asks for several pieces of one at once, which
		// only the store can answer. Serving them the cheap way would answer
		// such a request with the whole ninety-megabyte scan.
		prefix: "/api/v1/retro", root: flags.RetroRoot,
		right: cmsauth.Retro, refusal: "the archive requires the retro entitlement",
		ops: []string{"meta", "thumb", "media"},
	}, {
		// The music shelf reads its listings two ways on purpose: meta walks a
		// level at a time, search reaches the whole shelf in one call. Which is
		// the better bargain depends on how the shelf is arranged, so the page
		// can switch between them and be measured rather than argued about.
		prefix: "/api/v1/musik", root: flags.MusikRoot,
		right: cmsauth.Musik, refusal: "the music shelf requires the musik entitlement",
		ops: []string{"meta", "search", "tags", "thumb", "stream"},
	}, {
		// The DJ archive. The same five as the music shelf, because it is the
		// same kind of thing read the same way: a tree of folders holding audio,
		// pictures beside it, and tags inside the files.
		prefix: "/api/v1/djvet", root: flags.DjvetRoot,
		right: cmsauth.Djvet, refusal: "the DJ archive requires the djvet entitlement",
		ops: []string{"meta", "search", "tags", "thumb", "stream"},
	}, {
		// The media store itself, which is where a film's bytes live. Unlike the
		// four above it is not somewhere a visitor goes: it is read while content
		// is being written, by the editor's preview, which needs the file before
		// any entry names it. So it is admitted by a capability rather than by an
		// area — the one place in this list where the column holds the other kind
		// of right.
		//
		// Streaming only. There is nothing to browse here: the person asking
		// already knows the path, because they are the one who typed it.
		prefix: "/api/v1/media", root: flags.MediaRoot,
		right: cmsauth.MediaRead, refusal: "reading the media store requires the media.read permission",
		ops: []string{"stream"},
	}}
}

// library builds the drive this shelf is, or nothing where the deployment has
// not configured one.
//
// A shelf with no alias or no root is left nil rather than left out: the
// handlers answer "not configured" for themselves, which keeps a route that
// exists in the code visible at runtime instead of silently absent — easier to
// diagnose than a 404 that means two different things.
func (s shelf) library(tokens hi.TokenSource, alias string) *hidrive.Library {
	if alias == "" || s.root == "" {
		return nil
	}

	return hidrive.NewLibrary(hi.NewDrive(tokens, hi.DriveConfig{Alias: alias, Root: s.root}, nil), slog.Default())
}

// pattern is where one of a shelf's operations is served.
//
// One place builds the address, so what a test reads is what the router is
// given rather than a second spelling of it that has to be kept in step.
func (s shelf) pattern(op string) string {
	return "GET " + s.prefix + "/" + op + "/{path...}"
}

// routes declares one shelf: every operation it offers, behind its entitlement.
func (s shelf) routes(svc gate.Authenticator, lib *hidrive.Library) []func(*srvr.Api) {
	opts := make([]func(*srvr.Api), 0, len(s.ops))
	for _, name := range s.ops {
		op, ok := libraryOps[name]
		if !ok {
			// A row asking for an operation that does not exist is a mistake in
			// this file, not a condition to handle at runtime.
			panic("unknown shelf operation: " + name)
		}
		handler := func(w http.ResponseWriter, r *http.Request) error { return op(lib, w, r) }
		opts = append(opts, srvr.WithRoute(s.pattern(name), gate.Permission(svc, s.right, s.refusal, handler)))
	}

	return opts
}

// shelfRoutes declares every shelf, as one option.
func shelfRoutes(svc gate.Authenticator, tokens hi.TokenSource, flags Flags) func(*srvr.Api) {
	var opts []func(*srvr.Api)
	for _, s := range shelves(flags) {
		opts = append(opts, s.routes(svc, s.library(tokens, flags.MediaAlias))...)
	}

	return func(a *srvr.Api) {
		for _, opt := range opts {
			opt(a)
		}
	}
}
