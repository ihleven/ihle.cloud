package hidrive

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// id3v2 builds a tag of the shape the shelf actually carries: version 2.3, a
// frame per field, and a length in the header that says how far it reaches.
//
// pad makes the tag larger than it needs to be, which is how a file with its
// sleeve embedded in it behaves — the case the two-step read exists for.
func id3v2(title, artist, album string, pad int) []byte {
	var frames []byte
	add := func(id, text string) {
		body := append([]byte{0}, []byte(text)...) // 0: ISO-8859-1
		size := len(body)
		frames = append(frames, []byte(id)...)
		frames = append(frames, byte(size>>24), byte(size>>16), byte(size>>8), byte(size))
		frames = append(frames, 0, 0)
		frames = append(frames, body...)
	}
	add("TIT2", title)
	add("TPE1", artist)
	add("TALB", album)
	frames = append(frames, make([]byte, pad)...)

	size := len(frames)
	head := []byte{'I', 'D', '3', 3, 0, 0,
		byte(size >> 21 & 0x7f), byte(size >> 14 & 0x7f), byte(size >> 7 & 0x7f), byte(size & 0x7f)}

	return append(head, frames...)
}

// A file: its tag, then enough "audio" after it that a read of the tag is
// visibly a read of part of the file and not all of it.
func track(tag []byte) []byte {
	return append(append([]byte{}, tag...), make([]byte, 50_000)...)
}

// shelf stands in for the storage: it lists one directory and serves ranges out
// of the files in it, counting what was asked for — which is the thing these
// tests are actually about.
type shelf struct {
	files map[string][]byte

	mu     sync.Mutex
	reads  int
	ranges []string
}

func (s *shelf) server(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/dir", func(w http.ResponseWriter, r *http.Request) {
		var members []string
		for name := range s.files {
			members = append(members, fmt.Sprintf(`{"name":%q,"type":"file","category":"file","size":%d}`,
				name, len(s.files[name])))
		}
		fmt.Fprintf(w, `{"name":"album","members":[%s]}`, strings.Join(members, ","))
	})
	mux.HandleFunc("/meta", func(w http.ResponseWriter, r *http.Request) {
		name := lastPathSegment(r.URL.Query().Get("path"))
		fmt.Fprintf(w, `{"name":%q,"type":"file","size":%d}`, name, len(s.files[name]))
	})
	mux.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
		name := lastPathSegment(r.URL.Query().Get("path"))
		content, ok := s.files[name]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		spec := r.Header.Get("Range")
		s.mu.Lock()
		s.reads++
		s.ranges = append(s.ranges, name+" "+spec)
		s.mu.Unlock()

		from, to := 0, len(content)-1
		if strings.HasPrefix(spec, "bytes=") {
			parts := strings.SplitN(strings.TrimPrefix(spec, "bytes="), "-", 2)
			from, _ = strconv.Atoi(parts[0])
			if len(parts) > 1 && parts[1] != "" {
				to, _ = strconv.Atoi(parts[1])
			}
		}
		if from > len(content) {
			from = len(content)
		}
		if to >= len(content) {
			to = len(content) - 1
		}
		_, _ = w.Write(content[from : to+1])
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return server
}

func lastPathSegment(p string) string {
	parts := strings.Split(strings.TrimSuffix(p, "/"), "/")

	return parts[len(parts)-1]
}

func askTags(t *testing.T, lib *Library, dir string) ([]Tags, string) {
	t.Helper()

	r := httptest.NewRequest(http.MethodGet, "/api/v1/musik/tags/x", nil)
	r.SetPathValue("path", dir)
	w := httptest.NewRecorder()
	if err := lib.Tags(w, r); err != nil {
		t.Fatalf("reading tags: %v", err)
	}

	var body struct {
		Tracks []Tags `json:"tracks"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("reading the answer: %v", err)
	}

	return body.Tracks, w.Header().Get("X-Cache")
}

// The whole point of the feature: the artist is in the file and nowhere else.
func TestTagsReadWhatTheFolderNamesCannotSay(t *testing.T) {
	s := &shelf{files: map[string][]byte{
		"01. Poison.mp3": track(id3v2("Poison", "Alice Cooper", "Trash", 0)),
	}}
	lib := libraryOf(s.server(t), "/public/alben")

	tracks, _ := askTags(t, lib, "Trash")
	if len(tracks) != 1 {
		t.Fatalf("got %d tracks, want 1", len(tracks))
	}
	if !tracks[0].Read {
		t.Fatal("reported the file as untagged")
	}
	if tracks[0].Artist != "Alice Cooper" {
		t.Errorf("artist %q, want Alice Cooper", tracks[0].Artist)
	}
	if tracks[0].Title != "Poison" || tracks[0].Album != "Trash" {
		t.Errorf("read %q / %q", tracks[0].Title, tracks[0].Album)
	}
}

// Why headRead is four kilobytes rather than the ten bytes that would do.
//
// An ID3v2 tag declares its own length, so ten bytes is enough to learn how
// much to ask for — and then a second request has to ask for it. Reading a
// little more settles the ordinary file in one request, and on this shelf the
// ordinary file's tag is about 350 bytes.
func TestASmallTagIsReadInOneRequest(t *testing.T) {
	s := &shelf{files: map[string][]byte{
		"01. Poison.mp3": track(id3v2("Poison", "Alice Cooper", "Trash", 0)),
	}}
	lib := libraryOf(s.server(t), "/public/alben")

	if _, _ = askTags(t, lib, "Trash"); s.reads != 1 {
		t.Errorf("read the file %d times: %v", s.reads, s.ranges)
	}
}

// And why it is not simply larger: a tag can be far bigger than any guess worth
// making — a third of a megabyte on this shelf, where the sleeve is inside the
// file. It says so in its header, so the second request asks for exactly that
// and the tag is still read.
func TestALargeTagIsFetchedInFullAndStillRead(t *testing.T) {
	big := id3v2("Civil War", "Guns N' Roses", "Use Your Illusion II", 40_000)
	if len(big) <= headRead {
		t.Fatalf("the fixture is not larger than one read: %d", len(big))
	}

	s := &shelf{files: map[string][]byte{"01. Civil War.mp3": track(big)}}
	lib := libraryOf(s.server(t), "/public/alben")

	tracks, _ := askTags(t, lib, "Use Your Illusion II")
	if len(tracks) != 1 || !tracks[0].Read {
		t.Fatalf("a large tag was not read: %+v", tracks)
	}
	if tracks[0].Title != "Civil War" {
		t.Errorf("title %q", tracks[0].Title)
	}
	if s.reads != 2 {
		t.Errorf("read the file %d times, want 2 — the header, then the tag: %v", s.reads, s.ranges)
	}
	// The tag and a little past it — the little being where a variable-bitrate
	// file states its length, which is worth having for nothing. What matters
	// is that it is not the whole file: reading fifty kilobytes of audio to
	// learn a title is exactly what the declared length exists to avoid.
	if asked := s.ranges[1]; !strings.Contains(asked, "bytes=0-"+strconv.Itoa(len(big)+xingWindow-1)) {
		t.Errorf("second read asked for %q, want the tag and the frame after it", asked)
	}
}

// One odd file among twenty should cost that file and not the album. These are
// somebody's rips going back decades; some of them will be strange.
func TestAnUnreadableFileIsReportedRatherThanFailingTheAlbum(t *testing.T) {
	s := &shelf{files: map[string][]byte{
		"01. Poison.mp3":  track(id3v2("Poison", "Alice Cooper", "Trash", 0)),
		"02. Nothing.mp3": make([]byte, 5_000),
	}}
	lib := libraryOf(s.server(t), "/public/alben")

	tracks, _ := askTags(t, lib, "Trash")
	if len(tracks) != 2 {
		t.Fatalf("got %d tracks, want both", len(tracks))
	}

	byName := map[string]Tags{}
	for _, track := range tracks {
		byName[track.Name] = track
	}
	if !byName["01. Poison.mp3"].Read {
		t.Error("the readable file was not read")
	}
	if byName["02. Nothing.mp3"].Read {
		t.Error("a file with no tags was reported as having them")
	}
}

// Only what can be played. The shelf keeps a rip log beside every album, and
// asking the storage for parts of those is work for nothing.
func TestTagsIgnoreWhatIsNotAudio(t *testing.T) {
	s := &shelf{files: map[string][]byte{
		"01. Poison.mp3":           track(id3v2("Poison", "Alice Cooper", "Trash", 0)),
		"Alice Cooper - Trash.log": []byte("EAC extraction logfile"),
		"cover.jpeg":               []byte("\xff\xd8\xff"),
	}}
	lib := libraryOf(s.server(t), "/public/alben")

	tracks, _ := askTags(t, lib, "Trash")
	if len(tracks) != 1 {
		t.Fatalf("got %d tracks, want only the audio: %+v", len(tracks), tracks)
	}
	if s.reads != 1 {
		t.Errorf("fetched %d files: %v", s.reads, s.ranges)
	}
}

// Opening an album twice in a sitting is the ordinary way to use the page, and
// it is a request per file each time if nothing remembers.
func TestAnAlbumIsReadOnceAndThenRemembered(t *testing.T) {
	s := &shelf{files: map[string][]byte{
		"01. Poison.mp3": track(id3v2("Poison", "Alice Cooper", "Trash", 0)),
	}}
	lib := libraryOf(s.server(t), "/public/alben")

	first, cache := askTags(t, lib, "Trash")
	if cache != "miss" {
		t.Errorf("first asking reported %q", cache)
	}
	reads := s.reads

	second, cache := askTags(t, lib, "Trash")
	if cache != "hit" {
		t.Errorf("second asking reported %q", cache)
	}
	if s.reads != reads {
		t.Errorf("read the storage again: %d then %d", reads, s.reads)
	}
	if len(second) != len(first) || second[0].Artist != first[0].Artist {
		t.Error("the remembered answer differs from the read one")
	}
}

// The length ID3v2 puts in its header carries seven bits to the byte, so that
// no byte of it can be mistaken for the start of the audio. Reading it as an
// ordinary integer is off by a factor that grows with the size.
func TestSynchsafeDecodesSevenBitsPerByte(t *testing.T) {
	for _, c := range []struct {
		bytes []byte
		want  int
	}{
		{[]byte{0, 0, 0, 0}, 0},
		{[]byte{0, 0, 0, 0x7f}, 127},
		{[]byte{0, 0, 1, 0}, 128},
		{[]byte{0, 0, 2, 0x21}, 289},
		{[]byte{0x7f, 0x7f, 0x7f, 0x7f}, 268435455},
	} {
		if got := synchsafe(c.bytes); got != c.want {
			t.Errorf("synchsafe(%v) = %d, want %d", c.bytes, got, c.want)
		}
	}
}

func TestTagsSayWhenNoShelfIsConfigured(t *testing.T) {
	var absent *Library

	if err := absent.Tags(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/x", nil)); err == nil {
		t.Fatal("an unconfigured library read tags")
	}
}

// xingFrame builds the first audio frame of a variable-bitrate file: a frame
// header saying MPEG-1 layer III at 44.1 kHz, then the Xing header counting the
// frames in the file.
func xingFrame(frames int) []byte {
	frame := []byte{0xff, 0xfb, 0x90, 0x00}
	frame = append(frame, make([]byte, 32)...) // where the header actually sits

	frame = append(frame, []byte("Xing")...)
	frame = append(frame, 0, 0, 0, 1) // flags: a frame count follows
	frame = append(frame,
		byte(frames>>24), byte(frames>>16), byte(frames>>8), byte(frames))

	return append(frame, make([]byte, 400)...)
}

// tlen adds the frame that states a track's length outright, in milliseconds.
func tlen(ms string) []byte {
	body := append([]byte{0}, []byte(ms)...)
	size := len(body)
	frame := append([]byte("TLEN"), byte(size>>24), byte(size>>16), byte(size>>8), byte(size), 0, 0)

	return append(frame, body...)
}

// withFrames rebuilds a tag with extra frames inside it, so that the declared
// length still matches what is there.
func taggedFile(extra []byte, pad int, audio []byte) []byte {
	var frames []byte
	add := func(id, text string) {
		body := append([]byte{0}, []byte(text)...)
		size := len(body)
		frames = append(frames, []byte(id)...)
		frames = append(frames, byte(size>>24), byte(size>>16), byte(size>>8), byte(size))
		frames = append(frames, 0, 0)
		frames = append(frames, body...)
	}
	add("TIT2", "Poison")
	add("TPE1", "Alice Cooper")
	frames = append(frames, extra...)
	frames = append(frames, make([]byte, pad)...)

	size := len(frames)
	head := []byte{'I', 'D', '3', 3, 0, 0,
		byte(size >> 21 & 0x7f), byte(size >> 14 & 0x7f), byte(size >> 7 & 0x7f), byte(size & 0x7f)}

	return append(append(head, frames...), audio...)
}

// The length a tagger wrote down, which is exact where it is there at all.
func TestDurationIsReadFromTheTagWhereItIsWritten(t *testing.T) {
	s := &shelf{files: map[string][]byte{
		"01. Poison.mp3": taggedFile(tlen("269666"), 0, make([]byte, 5_000)),
	}}
	lib := libraryOf(s.server(t), "/public/alben")

	tracks, _ := askTags(t, lib, "Trash")
	if len(tracks) != 1 || tracks[0].Seconds != 269 {
		t.Fatalf("read %d seconds, want 269: %+v", tracks[0].Seconds, tracks)
	}
}

// And where it is not, the file still says how long it is — a variable-bitrate
// file counts its own frames, and a frame is a fixed number of samples. This is
// the case that covers half this shelf, whose larger tags carry no TLEN.
func TestDurationIsCountedFromTheFramesWhereTheTagIsSilent(t *testing.T) {
	// 10325 frames at 44.1 kHz is four minutes twenty-nine.
	s := &shelf{files: map[string][]byte{
		"01. Poison.mp3": taggedFile(nil, 0, xingFrame(10325)),
	}}
	lib := libraryOf(s.server(t), "/public/alben")

	tracks, _ := askTags(t, lib, "Trash")
	if len(tracks) != 1 || tracks[0].Seconds != 269 {
		t.Fatalf("counted %d seconds, want 269: %+v", tracks[0].Seconds, tracks)
	}
}

// The interaction that makes it work at all: the frame is *past* the tag, so a
// file whose tag is larger than one read would hide it unless the second read
// deliberately reaches beyond the tag's end. Both albums on this shelf without
// a TLEN are exactly that shape.
func TestDurationSurvivesATagLargerThanOneRead(t *testing.T) {
	file := taggedFile(nil, headRead*2, xingFrame(10325))
	s := &shelf{files: map[string][]byte{"01. Civil War.mp3": file}}
	lib := libraryOf(s.server(t), "/public/alben")

	tracks, _ := askTags(t, lib, "Use Your Illusion II")
	if len(tracks) != 1 || tracks[0].Seconds != 269 {
		t.Fatalf("counted %d seconds behind a large tag, want 269: %+v", tracks[0].Seconds, tracks)
	}
}

// A file that says nothing reports nothing, rather than a number worked out
// from its size — the page can make that guess itself and labels it as one.
func TestDurationIsAbsentRatherThanGuessed(t *testing.T) {
	s := &shelf{files: map[string][]byte{
		"01. Poison.mp3": taggedFile(nil, 0, make([]byte, 5_000)),
	}}
	lib := libraryOf(s.server(t), "/public/alben")

	tracks, _ := askTags(t, lib, "Trash")
	if len(tracks) != 1 {
		t.Fatalf("got %d tracks", len(tracks))
	}
	if tracks[0].Seconds != 0 {
		t.Errorf("invented a duration of %d seconds", tracks[0].Seconds)
	}
	if !tracks[0].Read {
		t.Error("a file with no duration was reported as having no tags either")
	}
}
