package hi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"path"
	"strings"

	"github.com/dhowden/tag"
	"github.com/uptrace/bunrouter"
)

type AuthCtxKey struct{}

// hier sind portierte Versionen der HAndler aus cli/hi/handler.go: MetaHandler, FileHandler, ThumbHandler, TagsHandler

func MetaHandler_Dep(w http.ResponseWriter, req bunrouter.Request) error {

	accesstoken := req.Context().Value(AuthCtxKey{})
	if accesstoken == nil {
		return errors.New("no accesstoken")
	}

	client := NewClient(accesstoken.(string), "/")

	path := "/" + req.Param("path") // strings.TrimPrefix(req.URL.Path, "/hi/meta") //req.Param("path")

	var meta *Meta
	var err error

	if path == "" || strings.HasSuffix(path, "/") {
		meta, err = client.GetDir(path)
	} else {
		meta, err = client.GetMeta(path)
	}
	if err != nil {
		fmt.Println("meta error:", err.Error())
		return err
	}

	return bunrouter.JSON(w, meta)
}

func hicookie(r *http.Request) string {
	for _, cookie := range r.Cookies() {
		if cookie.Name == "hitoken" {
			fmt.Println("hicookie", cookie)
			return cookie.Value
		}
	}
	return ""
}

func MetaHandlerMux(w http.ResponseWriter, r *http.Request) error {

	client := NewClient(hicookie(r), "/")

	path := "/" + r.PathValue("path")

	fmt.Println("meta path:", path)

	var meta *Meta
	var err error

	if path == "" || strings.HasSuffix(path, "/") {
		meta, err = client.GetDir(path)
		fmt.Println("meta dir:", meta)

	} else {
		meta, err = client.GetMeta(path)
		fmt.Println("meta:", meta)
	}
	if err != nil {
		fmt.Println("meta error:", err.Error())
		return err
	}

	// w.Header().Set("Content-Type", "application/json")

	return JSON(w, meta)
}

func FileHandler_Dep(w http.ResponseWriter, req bunrouter.Request) error {

	accesstoken := req.Context().Value(AuthCtxKey{})
	if accesstoken == nil {
		return errors.New("no accesstoken")
	}

	params := req.Request.URL.Query()
	params.Set("path", "/"+req.Param("path"))

	resp, err := rq(GET, "/file", params,
		headers(req.Request.Header),
		// header("X-Forwarded-Host", req.Header.Get("Host")),
		bearer(accesstoken.(string)),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	for key, val := range resp.Header {
		w.Header().Set(key, val[0])
	}
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Content-Disposition", "inline")

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)

	return err
}

func FileHandlerMux(w http.ResponseWriter, r *http.Request) error {

	accesstoken := hicookie(r)

	params := r.URL.Query()
	params.Set("path", "/"+r.PathValue("path"))

	resp, err := rq(GET, "/file", params,
		headers(r.Header),
		// header("X-Forwarded-Host", req.Header.Get("Host")),
		bearer(accesstoken),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	for key, val := range resp.Header {
		w.Header().Set(key, val[0])
	}
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Content-Disposition", "inline")

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)

	return err
}

// cleanPath
func cleanPath(filepath, home string) string {
	clean := path.Clean(filepath)
	if strings.HasPrefix(clean, "~") {
		clean = strings.Replace(clean, "~", home, 1)
	}

	if !strings.HasPrefix(clean, "/") {
		clean = "/" + clean
	}
	return clean
}
func JSON(w http.ResponseWriter, value interface{}) error {
	if value == nil {
		return nil
	}

	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	if err := enc.Encode(value); err != nil {
		return err
	}

	return nil
}
func ThumbHandler_Dep(w http.ResponseWriter, req bunrouter.Request) error {

	// prefix, accesstoken, _ := auth.GetPrefixAndToken(req.Request)
	accesstoken := req.Context().Value(AuthCtxKey{})
	if accesstoken == nil {
		return errors.New("no accesstoken")
	}
	prefix := ""

	params := req.URL.Query()
	path := req.Param("path")
	if path == "" {
		path = params.Get("path")
	}
	path = cleanPath(path, prefix)
	params.Set("path", path)

	fmt.Println("path:", path)

	resp, err := rq(GET, "/file/thumbnail", params, bearer(accesstoken.(string)))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	for key, val := range resp.Header {
		w.Header().Set(key, val[0])
	}
	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	return err

}

func ThumbHandlerMux(w http.ResponseWriter, r *http.Request) error {

	// prefix, accesstoken, _ := auth.GetPrefixAndToken(req.Request)
	accesstoken := hicookie(r)
	prefix := ""

	params := r.URL.Query()
	path := r.PathValue("path")
	if path == "" {
		path = params.Get("path")
	}
	path = cleanPath(path, prefix)
	params.Set("path", path)

	fmt.Println("path:", path)

	resp, err := rq(GET, "/file/thumbnail", params, bearer(accesstoken))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	for key, val := range resp.Header {
		w.Header().Set(key, val[0])
	}
	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	return err

}

func ServeReverseProxy_Dep(w http.ResponseWriter, req bunrouter.Request) error {
	fmt.Println("ServeReverseProxy")
	// prefix, accesstoken, _ := auth.GetPrefixAndToken(req.Request)
	accesstoken := req.Context().Value(AuthCtxKey{})
	if accesstoken == nil {
		return errors.New("no accesstoken")
	}
	prefix := "/"

	path := path.Clean("/" + req.Param("path"))
	if p := req.URL.Query().Get("path"); p != "" {
		path = p
	}
	fmt.Println("ServeReverseProxy", path, prefix, "access", accesstoken)

	url, err := NewClient(accesstoken.(string), prefix).GetURL(path)
	if err != nil {
		return err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host
	// w.Header().Set("access-control-allow-origin", "*")
	proxy.ServeHTTP(w, req.Request)

	return nil
}

func ServeReverseProxyMux(w http.ResponseWriter, r *http.Request) error {
	fmt.Println("ServeReverseProxy")
	// prefix, accesstoken, _ := auth.GetPrefixAndToken(req.Request)
	accesstoken := hicookie(r)
	prefix := "/"

	path := path.Clean("/" + r.PathValue("path"))
	if p := r.URL.Query().Get("path"); p != "" {
		path = p
	}
	fmt.Println("ServeReverseProxy", path, prefix, "access", accesstoken)

	url, err := NewClient(accesstoken, prefix).GetURL(path)
	if err != nil {
		return err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
	r.URL.Host = url.Host
	r.URL.Scheme = url.Scheme
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = url.Host
	// w.Header().Set("access-control-allow-origin", "*")
	proxy.ServeHTTP(w, r)

	return nil
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

func TagsHandler(w http.ResponseWriter, req bunrouter.Request) error {
	accesstoken := req.Context().Value(AuthCtxKey{})
	if accesstoken == nil {
		return errors.New("no accesstoken")
	}

	client := NewClient(accesstoken.(string), "/")

	meta, err := client.GetDir(strings.TrimPrefix(req.URL.Path, "/hi/tags"))
	if err != nil {
		return err
	}
	fs := NewFS(accesstoken.(string))
	///////////
	albumtags := AlbumTags{Tags: make(map[string]ID3Tags)}

	// entries, err := hfs.ReadDir(r.URL.Path)
	// if err != nil {
	// 	return err
	// }

	for i := range meta.Members {

		e := meta.Members[i]
		fmt.Println("==========", path.Join(meta.Path, e.Name()))
		if !strings.HasSuffix(e.Name(), ".mp3") {
			continue
		}
		file, err := fs.Open(path.Join(meta.Path, e.Name()))
		if err != nil {
			fmt.Println("open error", err)
			return err
		}
		defer file.Close()

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

	return bunrouter.JSON(w, albumtags.Tags)
}
