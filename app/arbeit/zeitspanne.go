package arbeit

// IsContained reports whether a Zeitspanne with this number is already in the
// list.
//
// Identity is the number and not the times: a span that has been moved is the
// same span, which is what lets a save work out which of the stored ones the
// editor has dropped.
func IsContained(haystack []Zeitspanne, needle Zeitspanne) bool {
	for _, n := range haystack {
		if n.Nr == needle.Nr {
			return true
		}
	}

	return false
}
