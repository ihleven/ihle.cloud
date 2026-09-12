package films

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ihleven/ihlvn/pkg/stream"
	"github.com/interhome-group/cms/mgmt"
	"github.com/interhome-group/cms/pkg/errs"
)

// Chaptered is implemented by a type whose entry divides into named stretches.
type Chaptered interface {
	Chapters() []Scene
}

// Chapters implements Chaptered. Scenes segment the film, which is exactly what
// a chapter track is.
func (f *Film) Chapters() []Scene { return f.Scenes }

// ChapterTrack serves a film's scenes as WebVTT.
//
// Derived rather than stored: the entry holds the domain model, and this is one
// projection of it. Serving the standard format means the page can say
// <track kind="chapters" src="…"> and let the browser parse it — which is both
// less code and less wrong than building cues by hand from timestamps.
func ChapterTrack(mngr *mgmt.Mngr) func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		entry, err := lookup(mngr, r.Context(), r.PathValue("id"))
		if errors.Is(err, stream.ErrNoSuchItem) {
			return errs.New("no such film", errs.HTTPStatus(http.StatusNotFound))
		}
		if err != nil {
			return err
		}

		chaptered, ok := entry.Content.(Chaptered)
		if !ok {
			return errs.New("no such film", errs.HTTPStatus(http.StatusNotFound))
		}

		var duration float64
		if film, ok := entry.Content.(*Film); ok {
			duration = film.Duration
		}

		w.Header().Set("Content-Type", "text/vtt; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// Private because the entry's ACL decided this response; a shared cache
		// must not hand it to someone the check would have refused.
		w.Header().Set("Cache-Control", "private, max-age=3600")

		_, _ = w.Write([]byte(renderVTT(chaptered.Chapters(), duration)))
		return nil
	}
}

// renderVTT writes scenes as a WebVTT chapter track.
//
// The cue payload is the scene's title, because that is what a player shows in
// a chapter menu; the description belongs to the entry, not to the track.
//
// A scene's end is its own where it has one, otherwise the start of the next
// scene — scenes segment the film, so the next one beginning is this one
// ending. The last scene falls back to the film's duration, and is dropped if
// there is nothing to close it with: a cue with no end is not a cue.
func renderVTT(scenes []Scene, duration float64) string {
	var out strings.Builder
	out.WriteString("WEBVTT\n")

	for i, scene := range scenes {
		end := scene.End
		switch {
		case end > 0:
		case i+1 < len(scenes):
			end = scenes[i+1].Start
		default:
			end = duration
		}
		if end <= scene.Start {
			continue
		}

		fmt.Fprintf(&out, "\n%s --> %s\n%s\n", vttTime(scene.Start), vttTime(end), scene.Title)
	}

	return out.String()
}

// vttTime formats seconds as WebVTT's HH:MM:SS.mmm. The hours field is always
// written: it is optional in the format, and always present is one case rather
// than two.
func vttTime(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}
	ms := int64(seconds*1000 + 0.5)

	return fmt.Sprintf("%02d:%02d:%02d.%03d", ms/3600000, ms/60000%60, ms/1000%60, ms%1000)
}
