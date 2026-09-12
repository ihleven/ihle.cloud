package main

import "github.com/interhome-group/cms/modules"

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
	modules.Define("mediathek", "Die Mediathek.")
	modules.Define("musik", "Die Musik.")
	modules.Define("search", "Die Suche.")
}
