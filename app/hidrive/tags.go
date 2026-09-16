package hidrive

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"net/url"
	stdpath "path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dhowden/tag"
	"github.com/ihleven/ihlvn/pkg/hi"
)

// What a music file says about itself.
//
// The shelf does not record any of this. A folder is called "1987 - Appetite
// For Destruction (lameV3A)" and the files inside it "01. Welcome To The
// Jungle.mp3", so a title can be read off a name — but the *artist* appears
// nowhere in the layout at all, and there is no level of folders to put it in.
// It is in the files, and this is the only way to it.
type Tags struct {
	// Name is the file, so a caller can line these up with a listing it already
	// has rather than trusting the order.
	Name string `json:"name"`
	// Read says whether the file had tags at all. Without it a track with an
	// empty title is indistinguishable from one this failed to read, and the
	// two want different things on screen.
	Read        bool   `json:"read"`
	Title       string `json:"title,omitempty"`
	Artist      string `json:"artist,omitempty"`
	AlbumArtist string `json:"album_artist,omitempty"`
	Album       string `json:"album,omitempty"`
	Genre       string `json:"genre,omitempty"`
	Year        int    `json:"year,omitempty"`
	Track       int    `json:"track,omitempty"`
	Tracks      int    `json:"tracks,omitempty"`
	Disc        int    `json:"disc,omitempty"`
	// Format is what carried them — ID3v2.3, ID3v1, VORBIS. Kept because it is
	// the first thing worth knowing when a file reads oddly.
	Format string `json:"format,omitempty"`
	// Seconds is how long the track runs, where the file says so exactly.
	//
	// Zero means it does not, and a caller then has only the estimate it can
	// make from the file's size — which is what the page fell back to before
	// this, and is out by whatever the encoder decided moment to moment.
	Seconds int `json:"seconds,omitempty"`
}

// headRead is how much of a file is fetched before its tags can be located.
//
// A guess, and the size of the guess is the whole design. An ID3v2 tag declares
// its own length in the first ten bytes, so the exact amount needed is knowable
// — but only after a request. Reading a little more than the header costs
// nothing and settles most files in one request instead of two: measured on
// this shelf, tags run from 325 bytes to 330 kilobytes, and the small ones are
// the common case by far. The large one is a file with the sleeve embedded in
// it, which is fetched in full only because it says it must be.
const headRead = 4096

// xingWindow is how far past the tag to read, so that the first audio frame
// comes along with it.
//
// That frame is where a variable-bitrate file albums how many frames it has,
// which is its length exactly — and the files here are variable-bitrate, being
// LAME rips. Reading it costs nothing: for an ordinary tag it is already inside
// headRead, and for a large one it is a slightly longer second request rather
// than another one.
const xingWindow = 2048

// tagsTTL is how long an album's tags are reused.
//
// They come from inside the files, so they change only when the files do, which
// on a shelf of finished albums is next to never. Long enough that opening the
// same album twice in a sitting costs one round trip rather than thirty, short
// enough that a corrected tag appears without a restart.
const tagsTTL = 30 * time.Minute

// maxTagReaders bounds how many files are read at once.
//
// One request per file, so an album of twenty is twenty requests; letting them
// all go at once is neither faster nor polite to a storage provider that is
// answering other requests for the same account.
const maxTagReaders = 8

type tagCache struct {
	ttl time.Duration

	mu      sync.Mutex
	entries map[string]tagEntry
}

type tagEntry struct {
	tags    []Tags
	expires time.Time
}

func newTagCache(ttl time.Duration) *tagCache {
	if ttl <= 0 {
		ttl = tagsTTL
	}

	return &tagCache{ttl: ttl, entries: map[string]tagEntry{}}
}

func (c *tagCache) get(key string) ([]Tags, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]
	if !ok || time.Now().After(entry.expires) {
		return nil, false
	}

	return entry.tags, true
}

func (c *tagCache) put(key string, tags []Tags) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = tagEntry{tags: tags, expires: time.Now().Add(c.ttl)}
}

// Tags answers with what the files of one album say about themselves.
//
// A directory rather than a file, because the caller wants an album and asking
// per track would be one request per track from the browser as well as from
// here. Reading them is not cheap — a request per file to the storage — so an
// album is remembered once read.
//
// A file whose tags cannot be read is reported with Read false rather than
// failing the album: one odd file among twenty should cost that file, not the
// album.
func (l *Library) Tags(w http.ResponseWriter, r *http.Request) error {
	if err := l.configured(); err != nil {
		return err
	}

	dir := path(r)
	if cached, ok := l.tags.get(dir); ok {
		// Named so the difference is visible from outside: the second opening
		// of an album is meant to be free, and a header is how that is checked
		// without instrumenting the client.
		w.Header().Set("X-Cache", "hit")
		w.Header().Set("Cache-Control", "private, max-age=600")

		return writeJSON(w, struct {
			Tracks []Tags `json:"tracks"`
		}{cached})
	}

	listing, err := l.drive.Dir(r.Context(), dir)
	if err != nil {
		return err
	}

	var names []string
	for _, member := range listing.Members {
		name := driveName(member)
		if member.Category != "directory" && isAudioFile(name) {
			names = append(names, name)
		}
	}

	tags := l.readAll(r.Context(), dir, names)
	l.tags.put(dir, tags)

	w.Header().Set("X-Cache", "miss")
	w.Header().Set("Cache-Control", "private, max-age=600")

	return writeJSON(w, struct {
		Tracks []Tags `json:"tracks"`
	}{tags})
}

// readAll reads every file's tags, a few at a time, preserving the listing's
// order so the answer lines up with what the caller asked about.
func (l *Library) readAll(ctx context.Context, dir string, names []string) []Tags {
	tags := make([]Tags, len(names))

	var wg sync.WaitGroup
	gate := make(chan struct{}, maxTagReaders)

	for i, name := range names {
		wg.Add(1)
		go func(i int, name string) {
			defer wg.Done()
			gate <- struct{}{}
			defer func() { <-gate }()

			tags[i] = l.read(ctx, stdpath.Join(dir, name), name)
		}(i, name)
	}
	wg.Wait()

	return tags
}

// read fetches as much of one file as its tags occupy, and no more.
//
// Two shapes have to be allowed for and they sit at opposite ends of the file.
// An ID3v2 tag is at the front and declares its own length, so a short read
// finds out how much to ask for and a second request is only needed when the
// tag is genuinely large. An ID3v1 tag is the last 128 bytes, and is what a
// file carries when it has nothing better.
//
// Everything here treats an unreadable file as untagged rather than as a
// failure: these are somebody's rips going back decades and some of them will
// be odd.
func (l *Library) read(ctx context.Context, p, name string) Tags {
	head, err := l.readRange(ctx, p, 0, headRead)
	if err != nil {
		return Tags{Name: name}
	}

	// An ID3v2 tag says how long it is, in a form that cannot be mistaken for
	// the audio that follows: four bytes of seven bits each. If it does not fit
	// in what was read, ask for exactly the rest.
	body, tagLen := head, 0
	if len(head) >= 10 && string(head[:3]) == "ID3" {
		tagLen = 10 + synchsafe(head[6:10])
		if tagLen+xingWindow > len(head) {
			if whole, err := l.readRange(ctx, p, 0, int64(tagLen+xingWindow)); err == nil {
				body = whole
			}
		}
	}

	if md, err := tag.ReadFrom(bytes.NewReader(body)); err == nil {
		tags := tagsOf(name, md)
		tags.Seconds = seconds(md, body, tagLen)

		return tags
	}

	// Nothing at the front: the other place a tag lives is the very end.
	if size := l.size(ctx, p); size > 128 {
		if tail, err := l.readRange(ctx, p, size-128, 128); err == nil {
			if md, err := tag.ReadID3v1Tags(bytes.NewReader(tail)); err == nil {
				return tagsOf(name, md)
			}
		}
	}

	return Tags{Name: name}
}

func (l *Library) readRange(ctx context.Context, p string, off, n int64) ([]byte, error) {
	body, err := l.drive.File(ctx, p, off, n)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	return io.ReadAll(io.LimitReader(body, n))
}

func (l *Library) size(ctx context.Context, p string) int64 {
	meta, err := l.drive.Meta(ctx, p)
	if err != nil {
		return 0
	}

	return int64(meta.Size_)
}

func tagsOf(name string, md tag.Metadata) Tags {
	track, tracks := md.Track()
	disc, _ := md.Disc()

	return Tags{
		Name:        name,
		Read:        true,
		Title:       md.Title(),
		Artist:      md.Artist(),
		AlbumArtist: md.AlbumArtist(),
		Album:       md.Album(),
		Genre:       md.Genre(),
		Year:        md.Year(),
		Track:       track,
		Tracks:      tracks,
		Disc:        disc,
		Format:      string(md.Format()),
	}
}

// seconds is how long a track runs, from whichever of two places says so.
//
// TLEN is a tag frame holding the length in milliseconds. It is exact where it
// is written and simply absent where it is not — on this shelf, present in the
// files with small tags and missing from the two albums with their sleeves
// embedded.
//
// Failing that, the first audio frame of a variable-bitrate file carries a
// Xing or Info header counting the frames in the file, and a frame is a fixed
// number of samples, so the count and the sample rate give the length exactly.
// Both agreed with each other and with what a browser reports, where both were
// present.
func seconds(md tag.Metadata, body []byte, tagLen int) int {
	if raw, ok := md.Raw()["TLEN"]; ok {
		if ms, err := strconv.Atoi(strings.TrimSpace(fmt.Sprint(raw))); err == nil && ms > 0 {
			return ms / 1000
		}
	}

	if tagLen <= 0 || tagLen >= len(body) {
		return 0
	}
	audio := body[tagLen:]

	frames := 0
	for _, marker := range []string{"Xing", "Info"} {
		if at := bytes.Index(window(audio), []byte(marker)); at >= 0 {
			frames = vbrFrames(audio[at:])
			break
		}
	}
	rate := sampleRate(audio)
	if frames <= 0 || rate <= 0 {
		return 0
	}

	// 1152 samples to the frame, which is what an MPEG-1 layer III frame holds.
	return frames * 1152 / rate
}

// window bounds the search for a header that is always near the front, so that
// a stray "Xing" in the audio cannot be mistaken for one.
func window(audio []byte) []byte {
	if len(audio) > 1200 {
		return audio[:1200]
	}

	return audio
}

// vbrFrames reads the frame count out of a Xing or Info header, which is only
// there when the header says it is.
func vbrFrames(b []byte) int {
	if len(b) < 12 {
		return 0
	}
	if binary.BigEndian.Uint32(b[4:8])&1 == 0 {
		return 0
	}

	return int(binary.BigEndian.Uint32(b[8:12]))
}

// sampleRate reads it from the first frame header — the sync word, then the
// version and the rate index it carries.
func sampleRate(b []byte) int {
	rates := map[byte][]int{0: {11025, 12000, 8000}, 2: {22050, 24000, 16000}, 3: {44100, 48000, 32000}}

	for i := 0; i+4 < len(b) && i < 2048; i++ {
		if b[i] != 0xff || b[i+1]&0xe0 != 0xe0 {
			continue
		}
		if set, ok := rates[(b[i+1]>>3)&3]; ok {
			if index := (b[i+2] >> 2) & 3; int(index) < len(set) {
				return set[index]
			}
		}
	}

	return 0
}

// synchsafe decodes the length ID3v2 puts in its header: four bytes carrying
// seven bits each, so that no byte of the length can be mistaken for the start
// of the audio.
func synchsafe(b []byte) int {
	return int(b[0])<<21 | int(b[1])<<14 | int(b[2])<<7 | int(b[3])
}

func isAudioFile(name string) bool {
	switch strings.ToLower(strings.TrimPrefix(stdpath.Ext(name), ".")) {
	case "mp3", "m4a", "flac", "ogg", "opus", "wav", "aac", "wma":
		return true
	default:
		return false
	}
}

// driveName is an entry's name as text: the provider percent-encodes them.
func driveName(m hi.Meta) string {
	if decoded, err := url.QueryUnescape(m.NameURLEncoded); err == nil {
		return decoded
	}

	return m.NameURLEncoded
}
