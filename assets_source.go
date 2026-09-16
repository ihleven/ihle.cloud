package main

import "embed"

// appSource embeds this application's own Go source so the /godoc routes can
// render its documentation at runtime — a deployed binary has no source tree on
// disk to read.
//
// Directories are embedded whole; the parser reads only non-test .go files and
// ignores everything else, so the handful of SQL migrations and a README that
// ride along cost a few kilobytes and are never served. Files beginning with a
// dot are excluded by embed itself, which is what keeps stray editor and
// finder droppings out.
//
// Kept in its own file with no build tag, so it is present in every build.
//
//go:embed *.go app pkg cli
var appSource embed.FS
