package cmsauth

import "github.com/interhome-group/cms/content"

// MediaRead is the right to read an object out of the media store by naming it.
//
// Not an area and deliberately not one. An area entitlement answers "may this
// visitor reach this part of the site", which is a question about consuming what
// has been published; this answers "may this person look at a file in the store
// by its path", which is a question asked while content is being written. The
// two have different holders: somebody may be entitled to watch the films
// without ever writing one, and somebody may write them without the films area
// appearing in their menu at all.
//
// What needs it is the editor's preview. A film entry names a file, and until
// the entry has been saved there is nothing to address it by except the path
// somebody is typing — so the usual protection, the entry's own access list,
// does not exist yet and something has to stand in its place.
//
// The nearest thing in the CMS is assets.read, which governs its own stored
// files rather than being derived from the right to write content, and which is
// prefix-capable so a grant can be narrowed to a folder. This follows both: it
// is a right of its own, and "media.read:/public/Super 8" is expressible without
// anything here changing.
//
// A right rather than a reachability argument. Somebody who may write a film
// entry can already name any key in it and play it through the delivery route,
// so this hands out no reach that was not there before — but that says the door
// is not dangerous, not who should hold the key, and the check has to be the
// second question.
var MediaRead = content.Define("media", "read", true, false,
	"This permission allows reading a file from the media store by its path.",
	"It is what the film editor's preview needs, where no entry exists yet to carry an access list.",
	"It can be used with a prefix to allow reading only within a folder, e.g. media.read:/public",
)
