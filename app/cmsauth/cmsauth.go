// Package cmsauth is what an account is allowed to see, in the CMS's terms.
//
// It holds the feature areas this app has and the rules for reading an account's
// permissions as access to them. Authentication — who someone is and how they
// proved it — is app/auth; this is the other question, asked afterwards.
//
// Nothing here knows what an account is. Everything takes a
// content.UserProvider, which is the CMS's own seam and which an account
// satisfies, so this package depends on the CMS's vocabulary and on nothing of
// ours. That is also what keeps app/auth free to use it without a cycle.
package cmsauth

import (
	"github.com/interhome-group/cms/content"
	"github.com/interhome-group/cms/modules"
)

// The feature areas of this app.
//
// An area is not a stored list on an account: it is a derived entitlement. Each
// one registers a module.<id> permission, and an account is entitled to the area
// exactly when its scope grants that permission — so "*" entitles everything and
// never needs updating when an area is added.
//
// Registering them here rather than in each feature package keeps the list in
// one readable place; the areas are UI surfaces, and several of them have no Go
// package of their own.
func init() {
	modules.Define("filme", "Die Super-8-Filme.")
	modules.Define("content", "Der Inhalts-Editor.")
	modules.Define("familie", "Stammbaum und Personen.")
	modules.Define("kalender", "Der Kalender.")
	modules.Define("search", "Die Suche.")
	modules.Define("art", "Das Kunstarchiv.")
}

// Hidrive gates browsing the family's storage.
//
// Held as a key rather than only defined, for the same reason as Admin: it is
// checked on the server, not merely offered in a menu. It decides *whether*
// someone may browse; which drive they land in is the account's HiDrive alias,
// or the deployment's when the account names none. Keeping the two apart is
// what stops an alias configured for some other purpose from quietly handing
// out a file browser.
var Hidrive = modules.Define(HidriveArea, "Die Dateien auf HiDrive.")

// Mediathek gates the shared video library.
//
// Held as a key rather than only defined, because it is checked on the server.
// Unlike Hidrive it decides nothing about *whose* tree is served: the library is
// one fixed drive, the same for everybody, which is what makes this entitlement
// the only thing standing in front of it — and what lets the responses be
// cached by URL, since the URL identifies the file for every account alike.
var Mediathek = modules.Define(MediathekArea, "Die Mediathek.")

// MediathekArea is the area's id, as the session reports it.
const MediathekArea = "mediathek"

// Retro gates the magazine archive. One fixed shelf like the Mediathek, and a
// separate entitlement for the same reason: reading old computer magazines and
// watching the family's videos are different permissions, and neither follows
// from the other.
var Retro = modules.Define(RetroArea, "Das Zeitschriften-Archiv.")

// RetroArea is the area's id, as the session reports it.
const RetroArea = "retro"

// HidriveArea is the area's id, as the session reports it. Named because the
// session withholds this one area from an account that has no storage to
// browse, and comparing against a literal there would be a second place to keep
// the spelling right.
const HidriveArea = "hidrive"

// Musik gates the music shelf. Its own entitlement for the reason the
// mediathek and the archive have theirs: one fixed shelf, checked on the
// server, and listening to the family's albums does not follow from watching
// its films.
var Musik = modules.Define(MusikArea, "Die Musik.")

// MusikArea is the area's id, as the session reports it.
const MusikArea = "musik"

// Djvet gates the DJ archive. Its own entitlement rather than the music one:
// the two are different collections with different audiences, and a shelf of
// the family's records says nothing about who should hear its sets.
var Djvet = modules.Define(DjvetArea, "Die DJ-Sets.")

// DjvetArea is the area's id, as the session reports it.
const DjvetArea = "djvet"

// Geheimtipp says that an account plays in the pool.
//
// The pool now shares this app's sign-in: an account here is how someone gets
// in, and the proxy mints the pool's own credential from the session. This is
// what decides whether it does — so the entitlement admits, and no longer
// merely offers, which is why it is held as a key like Hidrive and Admin
// rather than only defined.
//
// It is still absent from the route guard, and deliberately so: the proxy has
// to stay open, because the pool has pages a visitor with no account may read
// and gating the route would take those away. What the key gates is whose name
// goes into the token, not who may reach the pool at all.
var Geheimtipp = modules.Define(GeheimtippArea, "Die Tipprunde.")

// GeheimtippArea is the pool's area id, as the session reports it. Named for
// the same reason as HidriveArea: a confined account is reported as entitled to
// this area without its permissions being consulted, and a literal there would
// be a second place to keep the spelling right.
const GeheimtippArea = "geheimtipp"

// Admin gates account administration.
//
// It is the same module.<id> permission as any other area, so one grant both
// entitles the navigation and admits the API — and "*" satisfies it without
// being named, because a scope expands the wildcard to every registered
// permission when it is built. Held as a key because, unlike the other areas,
// this one is enforced on the server and not only offered in a menu.
var Admin = modules.Define("admin", "Konten und Rechte.")

// Entitled returns the areas a principal may see, derived from its scope rather
// than stored on it. A principal granting nothing is entitled to nothing, which
// is what keeps an anonymous visitor out of every area.
func Entitled(p content.UserProvider) []string {
	return modules.Entitled(p.ContentUser().Scope)
}

// May reports whether a principal holds a permission.
func May(p content.UserProvider, key content.PermissionKey) bool {
	return !p.ContentUser().Scope.Denies(key)
}

// wildcard is content.Scope's grant-all, matched there as the entire permission
// string.
const wildcard = "*"

// Unregistered returns the permission strings that no package in this binary has
// defined.
//
// The CMS drops them silently when building a scope, so an account can appear to
// hold rights that grant nothing. Reporting them is the only way that
// discrepancy is ever visible.
//
// It takes the stored strings rather than a principal because a scope has
// already thrown them away — which is exactly the problem being reported.
func Unregistered(permissions []string) []string {
	known := map[string]struct{}{}
	for _, p := range content.RegisteredPermissions() {
		known[p] = struct{}{}
	}

	var unknown []string
	for _, p := range permissions {
		// "*" is not a permission name but a grant-all the scope expands to every
		// registered permission when it is built. It is the whole string or
		// nothing: a constrained form like "*:de" is not expanded, and so really
		// does grant nothing.
		if p == wildcard {
			continue
		}

		// A permission may carry locale and path constraints after a colon; only
		// the name is registered.
		name := p
		for i := range len(p) {
			if p[i] == ':' {
				name = p[:i]
				break
			}
		}
		if _, ok := known[name]; !ok {
			unknown = append(unknown, p)
		}
	}
	return unknown
}
