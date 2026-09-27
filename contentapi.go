package main

import (
	"net/http"

	"github.com/interhome-group/cms/httpapi"
	"github.com/interhome-group/cms/mgmt"
	"github.com/interhome-group/cms/mgmt/controller"
	contentapi "github.com/interhome-group/cms/pkg/api"
)

// contentAPIPrefix is where the CMS's API is served, and it is the CMS's own
// name for it rather than a choice made here: the editing UI is mounted as a
// Nuxt layer and asks for this by default, so an app that serves it here
// configures nothing. It is a subtree pattern — ServeMux hands a subtree
// handler the unmodified path, which is what lets the absolute patterns
// registered inside still match.
const contentAPIPrefix = "/content-api/v1/"

// contentAPI is the CMS's own HTTP API, mounted whole.
//
// Mounted rather than re-implemented. This app used to hand-roll the two or
// three endpoints it happened to need, which is why creating an entry was
// impossible — there was no endpoint to create one — and why asking what content
// types exist had nowhere to go. Everything it serves is the CMS's code, and the
// only thing added here is who is asking.
//
// On a mux of its own because huma registers directly on the mux it is given:
// there is no wrapping a subset of an existing one afterwards, and every one of
// these endpoints wants the same identity middleware.
//
// Accounts, assets and pages are deliberately not mounted: this app administers
// its own accounts, has no asset service behind it, and does not register the
// CMS's page content types.
func contentAPI(cms *mgmt.Mngr, svc authenticator, version string) http.Handler {

	mux := http.NewServeMux()
	router := contentapi.New(mux, "ihlvn content API", version, "/content-api/v1")

	httpapi.RegisterEntry(router, &controller.EntryController{Mngr: cms})
	httpapi.RegisterDrafts(router, &controller.DraftController{Mngr: cms})
	httpapi.RegisterSearch(router, &controller.SearchController{Mngr: cms})
	httpapi.RegisterType(router, &controller.TypeController{Mngr: cms})
	httpapi.RegisterInfo(router, &controller.InfoController{Mngr: cms})
	router.FixGreedyPaths()

	return cmsIdentity(svc, mux)
}
