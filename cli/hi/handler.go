package hi

import (
	"fmt"
	"io"
	"net/http"
	gopath "path"

	"strings"

	"bitbucket.org/hotelplan/webcc-pkg/web"
	"github.com/dhowden/tag"
	"github.com/ihleven/pkg/errors"
	"github.com/ihleven/pkg/hidrive"
)

func MetaHandler(prefix string, t hidrive.Token) web.HandlerFunc {

	return func(w *web.ResponseWriter, r *http.Request) error {

		meta, err := (&hdclient{"", t.AccessToken}).GetDir(r.URL.Path)
		if err != nil {
			if e, ok := (errors.Cause(err)).(*Error); ok {
				fmt.Println(e.Code_, e.Message)
				if e.Code_ == 403 && strings.HasSuffix(e.Message, "not a directory") {
					meta, err = (&hdclient{"", t.AccessToken}).GetMeta(r.URL.Path)
					return w.RespondJSON(meta)
				}
			}

			return errors.Wrap(err, "Couldn't execute request")
		}
		return w.RespondJSON(meta)
	}
}

func FileHandler(prefix string, t hidrive.Token) web.HandlerFunc {

	return func(w *web.ResponseWriter, r *http.Request) error {

		req := newRequest(GET, "/file",
			HiHeaders(r.Header),
			bearer(t.AccessToken),
			path(prefix+r.URL.Path),
		)

		resp, err := req.Exec(nil)
		if err != nil {
			return errors.Wrap(err, "Couldn't execute request")
		}
		defer resp.Body.Close()

		for key, val := range resp.Header {
			w.Header().Set(key, val[0])
		}
		w.Header().Set("Content-Disposition", "inline")

		w.WriteHeader(resp.StatusCode)

		_, err = io.Copy(w, resp.Body)

		return err
	}
}

func ThumbHandler(token string) web.HandlerFunc {

	return func(w *web.ResponseWriter, r *http.Request) error {

		req := newRequest(GET, "/file/thumbnail?"+r.URL.Query().Encode(),
			bearer(token),
		)

		resp, err := req.Exec(&client)
		if err != nil {
			return errors.Wrap(err, "Couldn't execute request")
		}

		defer resp.Body.Close()

		for key, val := range resp.Header {
			w.Header().Set(key, val[0])
		}
		w.WriteHeader(resp.StatusCode)

		_, err = io.Copy(w, resp.Body)
		return err
	}
}

type ID3Tags struct {
	Format      string                 `json:"_format,omitempty"`
	FileType    string                 `json:"_filetype,omitempty"`
	Title       string                 `json:"title,omitempty"`
	Album       string                 `json:"album,omitempty"`
	Artist      string                 `json:"artist,omitempty"`
	AlbumArtist string                 `json:"albumArtist,omitempty"`
	Composer    string                 `json:"composer,omitempty"`
	Genre       string                 `json:"genre,omitempty"`
	Year        int                    `json:"year,omitempty"`
	Track       []int                  `json:"track,omitempty"`
	Disc        []int                  `json:"disc,omitempty"`
	Lyrics      string                 `json:"lyrics,omitempty"`
	Comment     string                 `json:"comment,omitempty"`
	Raw         map[string]interface{} `json:"raw"`     // NB: raw tag names are not consistent across formats.
	Picture     *tag.Picture           `json:"artwork"` // Artwork
}

type AlbumTags struct {
	Tags map[string]ID3Tags
}

func TagsHandler(token string, hfs *Drive) web.HandlerFunc {

	return func(w *web.ResponseWriter, r *http.Request) error {

		albumtags := AlbumTags{Tags: make(map[string]ID3Tags)}

		entries, err := hfs.ReadDir(r.URL.Path)
		if err != nil {
			return err
		}

		for i := range entries {
			e := entries[i]
			if !strings.HasSuffix(e.Name(), ".mp3") || i > 3 {
				continue
			}
			file, err := hfs.Open(gopath.Join(r.URL.Path, e.Name()))
			if err != nil {
				return err
			}

			if seeker, ok := file.(io.ReadSeeker); ok {
				metadata, err := tag.ReadFrom(seeker)
				if err != nil {
					return err
				}
				n, t := metadata.Track()
				d, nd := metadata.Disc()
				tags := ID3Tags{
					Format:      string(metadata.Format()),
					FileType:    string(metadata.FileType()),
					Title:       metadata.Title(),
					Album:       metadata.Album(),
					Artist:      metadata.Artist(),
					AlbumArtist: metadata.AlbumArtist(),
					Composer:    metadata.Composer(),
					Genre:       metadata.Genre(),
					Year:        metadata.Year(),
					Track:       []int{n, t},
					Disc:        []int{d, nd},
					Lyrics:      metadata.Lyrics(),
					Comment:     metadata.Comment(),
					Raw:         metadata.Raw(),
					Picture:     metadata.Picture(),
				}
				albumtags.Tags[e.Name()] = tags
			}
		}

		return w.RespondJSON(albumtags.Tags)
	}
}
