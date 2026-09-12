// Package stream serves media bytes to a browser: HTTP range requests,
// conditional requests and HEAD, on top of a blob.Blobstore.
//
// Range handling is the reason a video player can seek. A player asks for the
// last few bytes first (to read the container index), then jumps around; each
// jump is a fresh ranged request. Getting the arithmetic wrong shows up as a
// video that plays but cannot be scrubbed, so the rules live in one small,
// separately tested function.
package stream

import (
	"errors"
	"strconv"
	"strings"
)

// ErrUnsatisfiable reports a syntactically valid range that cannot be served,
// which must be answered with 416 and a Content-Range of "bytes */size".
var ErrUnsatisfiable = errors.New("stream: range not satisfiable")

// Range is a resolved byte range: an absolute offset and a length, both already
// clamped to the object size.
type Range struct {
	Off int64
	Len int64
}

// Last returns the inclusive last byte position, for Content-Range.
func (r Range) Last() int64 { return r.Off + r.Len - 1 }

// ParseRange resolves a Range header against a known object size.
//
// It returns ok=false when there is no usable range and the caller should serve
// the whole object with 200: an absent header, a malformed one (RFC 9110 says an
// unparsable Range must be ignored rather than rejected), or a multi-range
// request, which browsers do not use for media and HiDrive does not support.
//
// It returns ErrUnsatisfiable for a well-formed range that lies outside the
// object, which is the one case that must become 416.
func ParseRange(header string, size int64) (Range, bool, error) {
	const prefix = "bytes="

	spec := strings.TrimSpace(header)
	if spec == "" || !strings.HasPrefix(spec, prefix) {
		return Range{}, false, nil
	}
	spec = strings.TrimSpace(strings.TrimPrefix(spec, prefix))
	// A comma means several ranges; serving the whole object is a valid and much
	// simpler answer than assembling a multipart/byteranges body.
	if spec == "" || strings.Contains(spec, ",") {
		return Range{}, false, nil
	}

	dash := strings.IndexByte(spec, '-')
	if dash < 0 {
		return Range{}, false, nil
	}
	startStr := strings.TrimSpace(spec[:dash])
	endStr := strings.TrimSpace(spec[dash+1:])

	// Suffix form "bytes=-N": the last N bytes. Players use this to read a
	// trailing index, so it has to work even when N exceeds the object.
	if startStr == "" {
		n, err := strconv.ParseInt(endStr, 10, 64)
		if err != nil || n < 0 {
			return Range{}, false, nil
		}
		if n == 0 || size == 0 {
			return Range{}, false, ErrUnsatisfiable
		}
		if n > size {
			n = size
		}
		return Range{Off: size - n, Len: n}, true, nil
	}

	start, err := strconv.ParseInt(startStr, 10, 64)
	if err != nil || start < 0 {
		return Range{}, false, nil
	}
	// Also covers size == 0, where every offset is out of bounds.
	if start >= size {
		return Range{}, false, ErrUnsatisfiable
	}

	// Open form "bytes=N-": from N to the end.
	if endStr == "" {
		return Range{Off: start, Len: size - start}, true, nil
	}

	end, err := strconv.ParseInt(endStr, 10, 64)
	if err != nil || end < 0 {
		return Range{}, false, nil
	}
	if end < start {
		return Range{}, false, ErrUnsatisfiable
	}
	// A request past the end is clamped rather than refused: players routinely
	// ask for more than is there when probing.
	if end >= size {
		end = size - 1
	}
	return Range{Off: start, Len: end - start + 1}, true, nil
}
