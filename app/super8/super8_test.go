package super8

import "testing"

// Content written before the URLs were relative still carries an absolute host,
// so both forms have to reduce to the same path.
func TestMediaPath(t *testing.T) {
	tests := map[string]string{
		"http://localhost:8000/hi/media/proxy/videos/a.mp4": "videos/a.mp4",
		"https://tschabrun.de/hi/media/proxy/videos/a.mp4":  "videos/a.mp4",
		"/hi/media/proxy/videos/a.mp4":                      "videos/a.mp4",
		"videos/a.mp4":                                      "videos/a.mp4",
		"/videos/a.mp4":                                     "videos/a.mp4",
	}
	for source, want := range tests {
		if got := mediaPath(source); got != want {
			t.Errorf("mediaPath(%q) = %q, want %q", source, got, want)
		}
	}
}
