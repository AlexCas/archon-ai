package mapgen

import (
	"errors"
	"strings"
)

const (
	mapStart = "<!-- MAP:START -->"
	mapEnd   = "<!-- MAP:END -->"
)

// ErrNestedRegion is returned when existing contains more than one
// MAP:START/END marker pair, or the markers are present but out of order.
var ErrNestedRegion = errors.New("mapgen: more than one MAP:START/END region found")

// ErrPartialMarker is returned when existing contains exactly one of the two
// MAP:START/MAP:END markers (e.g. a user hand-editing the authored preamble
// accidentally deleted the other one). Appending a fresh pair in this case
// would leave two MAP:START (or MAP:END) markers behind, which is a
// corruption ErrNestedRegion would then report forever — so this case is
// signaled distinctly instead of silently appending.
var ErrPartialMarker = errors.New("mapgen: exactly one of MAP:START/MAP:END markers found")

// markerSpan locates the single MAP:START/END marker pair in existing.
// startIdx == -1 (err == nil) means both markers are absent — safe to
// append. Returns ErrPartialMarker if exactly one marker is present, and
// ErrNestedRegion if either marker repeats or the pair is out of order.
func markerSpan(existing string) (startIdx, endIdx int, err error) {
	startCount := strings.Count(existing, mapStart)
	endCount := strings.Count(existing, mapEnd)

	if startCount > 1 || endCount > 1 {
		return -1, -1, ErrNestedRegion
	}
	if startCount != endCount {
		return -1, -1, ErrPartialMarker
	}
	if startCount == 0 {
		return -1, -1, nil
	}

	startIdx = strings.Index(existing, mapStart)
	endIdx = strings.Index(existing, mapEnd)
	if endIdx < startIdx {
		return -1, -1, ErrNestedRegion
	}
	return startIdx, endIdx, nil
}

// Splice replaces the content between the MAP:START/END markers in existing
// with body. If the markers are absent, they are appended (wrapping body) to
// the end of existing. All bytes outside the markers are preserved exactly.
func Splice(existing, body string) (string, error) {
	startIdx, endIdx, err := markerSpan(existing)
	if err != nil {
		return "", err
	}

	if startIdx == -1 {
		prefix := existing
		if prefix != "" && !strings.HasSuffix(prefix, "\n") {
			prefix += "\n"
		}
		return prefix + mapStart + "\n" + body + mapEnd + "\n", nil
	}

	before := existing[:startIdx+len(mapStart)]
	after := existing[endIdx:]
	return before + "\n" + body + after, nil
}
