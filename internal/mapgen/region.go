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
// MAP:START/END marker pair.
var ErrNestedRegion = errors.New("mapgen: more than one MAP:START/END region found")

// Splice replaces the content between the MAP:START/END markers in existing
// with body. If the markers are absent, they are appended (wrapping body) to
// the end of existing. All bytes outside the markers are preserved exactly.
func Splice(existing, body string) (string, error) {
	if strings.Count(existing, mapStart) > 1 || strings.Count(existing, mapEnd) > 1 {
		return "", ErrNestedRegion
	}

	startIdx := strings.Index(existing, mapStart)
	endIdx := strings.Index(existing, mapEnd)

	if startIdx == -1 || endIdx == -1 {
		prefix := existing
		if prefix != "" && !strings.HasSuffix(prefix, "\n") {
			prefix += "\n"
		}
		return prefix + mapStart + "\n" + body + mapEnd + "\n", nil
	}
	if endIdx < startIdx {
		return "", ErrNestedRegion
	}

	before := existing[:startIdx+len(mapStart)]
	after := existing[endIdx:]
	return before + "\n" + body + after, nil
}
